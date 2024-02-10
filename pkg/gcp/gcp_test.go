package gcp

import (
	"context"
	"fmt"
	"hash/crc32"
	"testing"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func Test_createKey(t *testing.T) {
	type args struct {
		project string
		name    string
		version string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with valid inputs",
			args: args{project: "test-project", name: "test-name", version: "test-version"},
			want: "projects/test-project/secrets/test-name/versions/test-version",
		},
		{
			name: "Test with empty inputs",
			args: args{project: "", name: "", version: ""},
			want: "projects//secrets//versions/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createKey(tt.args.project, tt.args.name, tt.args.version); got != tt.want {
				t.Errorf("createKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheExpired(t *testing.T) {
	tests := []struct {
		name  string
		cache secretValueCache
		want  bool
	}{
		{
			name: "Cache value is empty",
			cache: secretValueCache{
				CachedValue: "",
				CachedTime:  time.Now(),
				TimeToLive:  1 * time.Hour,
			},
			want: true,
		},
		{
			name: "Current time is after cache expiration time",
			cache: secretValueCache{
				CachedValue: "test",
				CachedTime:  time.Now().Add(-2 * time.Hour),
				TimeToLive:  1 * time.Hour,
			},
			want: true,
		},
		{
			name: "Current time is before cache expiration time",
			cache: secretValueCache{
				CachedValue: "test",
				CachedTime:  time.Now(),
				TimeToLive:  1 * time.Hour,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cacheExpired(tt.cache); got != tt.want {
				t.Errorf("cacheExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckCache(t *testing.T) {
	type args struct {
		project string
		name    string
		version string
	}
	tests := []struct {
		name             string
		args             args
		setupCache       func()
		wantSecretVal    string
		wantRefreshCache bool
	}{
		{
			name: "Cache does not exist",
			args: args{project: "test-project", name: "test-name", version: "test-cache-version"},
			setupCache: func() {
				secretCache = make(map[string]secretValueCache)
			},
			wantSecretVal:    "",
			wantRefreshCache: true,
		},
		{
			name: "Cache exists and is expired",
			args: args{project: "test-project", name: "test-name", version: "test-cache-version"},
			setupCache: func() {
				secretCache = map[string]secretValueCache{
					"projects/test-project/secrets/test-name/versions/test-cache-version": {
						CachedValue: "test",
						CachedTime:  time.Now().Add(-2 * time.Hour),
						TimeToLive:  1 * time.Hour,
					},
				}
			},
			wantSecretVal:    "test",
			wantRefreshCache: true,
		},
		{
			name: "Cache exists and is not expired",
			args: args{project: "test-project", name: "test-name", version: "test-cache-version"},
			setupCache: func() {
				secretCache = map[string]secretValueCache{
					"projects/test-project/secrets/test-name/versions/test-cache-version": {
						CachedValue: "test",
						CachedTime:  time.Now(),
						TimeToLive:  1 * time.Hour,
					},
				}
			},
			wantSecretVal:    "test",
			wantRefreshCache: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			gotSecretVal, gotRefreshCache := checkCache(tt.args.project, tt.args.name, tt.args.version)
			if gotSecretVal != tt.wantSecretVal {
				t.Errorf("checkCache() gotSecretVal = %v, want %v", gotSecretVal, tt.wantSecretVal)
			}
			if gotRefreshCache != tt.wantRefreshCache {
				t.Errorf("checkCache() gotRefreshCache = %v, want %v", gotRefreshCache, tt.wantRefreshCache)
			}
		})
	}
}

func TestFromCtx(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		setupGlobal func()
		wantNil     bool
	}{
		{
			name: "Context contains secretmanager.Client",
			ctx:  context.WithValue(context.Background(), ctxSecManClientKey{}, &secretmanager.Client{}),
			setupGlobal: func() {
				secretManagerClient = nil
			},
			wantNil: false,
		},
		{
			name: "Context does not contain secretmanager.Client, but secretManagerClient is not nil",
			ctx:  context.Background(),
			setupGlobal: func() {
				secretManagerClient = &secretmanager.Client{}
			},
			wantNil: false,
		},
		{
			name: "Neither context contains secretmanager.Client nor secretManagerClient is not nil",
			ctx:  context.Background(),
			setupGlobal: func() {
				secretManagerClient = nil
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupGlobal()
			got := FromCtx(tt.ctx)
			if (got == nil) != tt.wantNil {
				t.Errorf("FromCtx() = %v, want nil = %v", got, tt.wantNil)
			}
		})
	}
}

func TestWithCtx(t *testing.T) {
	client := secretmanager.Client{}
	clientKey := ctxSecManClientKey{}
	tests := []struct {
		name string
		ctx  context.Context
		l    *secretmanager.Client
		want *secretmanager.Client
	}{
		{
			name: "Context already contains the same secretmanager.Client",
			ctx:  context.WithValue(context.Background(), clientKey, &client),
			l:    &client,
			want: &client,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCtx := WithCtx(tt.ctx, tt.l)
			got := gotCtx.Value(ctxSecManClientKey{}).(*secretmanager.Client)
			if got != tt.want {
				t.Errorf("WithCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSecret(t *testing.T) {

	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum([]byte("test"), crc32c))

	type args struct {
		project string
		name    string
		version string
	}
	tests := []struct {
		name        string
		args        args
		secretValue string
		checkSum    int64
		want        string
		wantErr     bool
	}{
		{
			name:        "secret exists and valid",
			args:        args{project: "test-project", name: "test-name", version: "test-version"},
			secretValue: "test",
			want:        "test",
			checkSum:    checksum,
			wantErr:     false,
		},
		{
			name:        "secret exists invalid checksum",
			args:        args{project: "test-project", name: "test-name", version: "test-version2"},
			secretValue: "test",
			want:        "",
			checkSum:    123456789,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			testServer, client := NewFakeServerAndClient(ctx, t)
			secretName := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", tt.args.project, tt.args.name, tt.args.version)

			testServer.Responses[secretName] = FakeSecretManagerServerResponse{
				Response: &secretmanagerpb.AccessSecretVersionResponse{
					Name: secretName,
					Payload: &secretmanagerpb.SecretPayload{
						Data:       []byte(tt.want),
						DataCrc32C: &tt.checkSum,
					},
				},
				Err: nil,
			}

			ctx = WithCtx(ctx, client)

			got, err := GetSecret(ctx, client, tt.args.project, tt.args.name, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}
