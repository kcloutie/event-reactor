package gcp

import (
	"context"
	"fmt"
	"hash/crc32"

	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/kcloutie/event-reactor/pkg/logger"
)

var (
	secretCache                     = map[string]secretValueCache{}
	GcpSecretsCacheTTLInMinutes int = 5
)

type secretValueCache struct {
	TimeToLive  time.Duration
	CachedTime  time.Time
	CachedValue string
}

func createKey(project string, name string, version string) string {
	return fmt.Sprintf("projects/%s/secrets/%v/versions/%v", project, name, version)
}

func cacheExpired(cache secretValueCache) bool {
	now := time.Now()
	expiredOn := cache.CachedTime.Add(cache.TimeToLive)
	if cache.CachedValue == "" || now.After(expiredOn) {
		return true
	}
	return false
}

func checkCache(project string, name string, version string) (string, bool) {
	key := createKey(project, name, version)
	refreshCache := false
	val, exists := secretCache[key]
	secretVal := ""
	if exists {
		refreshCache = cacheExpired(val)
		secretVal = val.CachedValue
	} else {
		refreshCache = true
	}
	return secretVal, refreshCache
}

func GetSecret(ctx context.Context, client *secretmanager.Client, project, name, version string) (string, error) {
	log := logger.FromCtx(ctx)
	secretVal, refreshCache := checkCache(project, name, version)
	if !refreshCache {
		log.Info("Getting GCP secret from the cache")
		return secretVal, nil
	}
	log.Info("GCP secret did not exist in cache or was expired...getting the value and caching it")

	if client == nil {
		var err error
		client, err = secretmanager.NewClient(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to setup secret manager client: %v", err)
		}
		defer client.Close()
	}

	fullName := createKey(project, name, version)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fullName,
	}
	result, err := client.AccessSecretVersion(ctx, req)

	if err != nil {
		return "", fmt.Errorf("failed to get secret '%v': %v", req.Name, err)
	}

	if result.Payload.DataCrc32C != nil {
		crc32c := crc32.MakeTable(crc32.Castagnoli)
		checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
		if checksum != *result.Payload.DataCrc32C {
			return "", fmt.Errorf("data corruption detected on the value of secret version '%v': %v", req.Name, err)
		}
	}

	secretCache[fullName] = secretValueCache{
		TimeToLive:  time.Duration(time.Duration(GcpSecretsCacheTTLInMinutes) * time.Minute),
		CachedValue: string(result.Payload.Data),
		CachedTime:  time.Now(),
	}
	return secretCache[fullName].CachedValue, nil
}

var secretManagerClient *secretmanager.Client

type ctxSecManClientKey struct{}

func FromCtx(ctx context.Context) *secretmanager.Client {
	if l, ok := ctx.Value(ctxSecManClientKey{}).(*secretmanager.Client); ok {
		return l
	} else if l := secretManagerClient; l != nil {
		return l
	}
	return nil
}

func WithCtx(ctx context.Context, l *secretmanager.Client) context.Context {
	if lp, ok := ctx.Value(ctxSecManClientKey{}).(*secretmanager.Client); ok {
		if lp == l {
			return ctx
		}
	}
	return context.WithValue(ctx, ctxSecManClientKey{}, l)
}
