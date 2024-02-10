package http

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gregjones/httpcache/diskcache"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/kcloutie/event-reactor/pkg/logger"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
	"go.uber.org/zap"
)

var (
	// A regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// A regular expression to match the error returned by net/http when the
	// scheme specified in the URL is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	// A regular expression to match the error returned by net/http when the
	// TLS certificate is not trusted. This error isn't typed
	// specifically so we resort to matching on the error string.
	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)

	cache *diskcache.Cache = nil
	// transport *httpcache.Transport = nil

	CacheDirectory       = path.Join(os.TempDir(), fmt.Sprintf("%s-cache", settings.CliBinaryName))
	MaxCacheAgeInSeconds = 300
)

func NewHttpRetryClient(log *zap.Logger, maxRetries int) *retryablehttp.Client {
	if cache == nil {
		cache = diskcache.New(CacheDirectory)
	}
	// transport := NewTransport(cache)
	// transport := NewTransportLocalCacheTime(cache)
	retryClient := retryablehttp.NewClient()
	// retryClient.HTTPClient.Transport = transport
	lgr := logger.NewLeveledLogger(log)
	retryClient.Logger = retryablehttp.LeveledLogger(&lgr)
	retryClient.CheckRetry = HttpErrorPropagatedRetryPolicy
	retryClient.RetryMax = maxRetries
	return retryClient
}

// ErrorPropagatedRetryPolicy is the same as DefaultRetryPolicy, except it
// propagates errors back instead of returning nil. This allows you to inspect
// why it decided to retry or not.
func HttpErrorPropagatedRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	return baseRetryPolicy(resp, err)
}

func baseRetryPolicy(resp *http.Response, err error) (bool, error) {

	if err != nil {
		if v, ok := err.(*url.Error); ok {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false, v
			}

			// Don't retry if the error was due to an invalid protocol scheme.
			if schemeErrorRe.MatchString(v.Error()) {
				return false, v
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if notTrustedErrorRe.MatchString(v.Error()) {
				return false, v
			}
			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
				return false, v
			}
		}

		// The error is likely recoverable so retry.
		return true, nil
	}

	// 429 Too Many Requests is recoverable. Sometimes the server puts
	// a Retry-After response header to indicate when the server is
	// available to start processing request from client.
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}

	// Check the response code. We retry on 500-range responses to allow
	// the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side. This will catch
	// invalid response codes as well, like 0 and 999.
	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		body := []byte{}
		if resp.Body != nil {
			defer resp.Body.Close()
			body, _ = io.ReadAll(resp.Body)
		}
		shouldRetry := true
		strBody := string(body)
		if strings.Contains(strBody, "AN ERROR HAS OCCURRED!!") {
			shouldRetry = false
		}
		if len(body) > 0 {
			return shouldRetry, fmt.Errorf("unexpected HTTP status %s. Error: %s", resp.Status, string(body))
		} else {
			return shouldRetry, fmt.Errorf("unexpected HTTP status %s", resp.Status)
		}
	}

	return false, nil
}

func CreateHttpErrorMessage(url string, body string, statusCode int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Http request failed with status code %v\n\n", statusCode))
	sb.WriteString(fmt.Sprintf("  URL: %s\n\n", url))
	sb.WriteString(fmt.Sprintf("%s\n", body))
	return sb.String()
}
