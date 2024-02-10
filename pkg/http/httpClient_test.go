package http

import (
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestCreateHttpErrorMessage(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		body       string
		statusCode int
		want       string
	}{
		{
			name:       "Test with valid inputs",
			url:        "http://test.com",
			body:       "Test body",
			statusCode: 404,
			want:       fmt.Sprintf("Http request failed with status code %v\n\n  URL: %s\n\n%s\n", 404, "http://test.com", "Test body"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateHttpErrorMessage(tt.url, tt.body, tt.statusCode); got != tt.want {
				t.Errorf("CreateHttpErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseRetryPolicy(t *testing.T) {
	tests := []struct {
		name string
		resp *http.Response
		err  error
		want bool
	}{
		{
			name: "Test with too many redirects error",
			resp: nil,
			err:  &url.Error{Err: fmt.Errorf("stopped after 10 redirects")},
			want: false,
		},
		{
			name: "Test with invalid protocol scheme error",
			resp: nil,
			err:  &url.Error{Err: fmt.Errorf("unsupported protocol scheme")},
			want: false,
		},
		{
			name: "Test with TLS cert verification failure error",
			resp: nil,
			err:  &url.Error{Err: x509.UnknownAuthorityError{}},
			want: false,
		},
		{
			name: "Test with recoverable error",
			resp: nil,
			err:  fmt.Errorf("recoverable error"),
			want: true,
		},
		{
			name: "Test with 429 Too Many Requests response",
			resp: &http.Response{StatusCode: http.StatusTooManyRequests},
			err:  nil,
			want: true,
		},
		{
			name: "Test with 500 Internal Server Error response",
			resp: &http.Response{StatusCode: http.StatusInternalServerError},
			err:  nil,
			want: true,
		},
		{
			name: "Test with 501 Not Implemented response",
			resp: &http.Response{StatusCode: http.StatusNotImplemented},
			err:  nil,
			want: false,
		},
		{
			name: "Test with 0 status code response",
			resp: &http.Response{StatusCode: 0},
			err:  nil,
			want: true,
		},
		{
			name: "Test with non-retryable error in response body",
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("AN ERROR HAS OCCURRED!!")),
			},
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := baseRetryPolicy(tt.resp, tt.err)
			if got != tt.want {
				t.Errorf("baseRetryPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHttpRetryClient(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
	}{
		{
			name:       "Test with zero retries",
			maxRetries: 0,
		},
		{
			name:       "Test with one retry",
			maxRetries: 1,
		},
		{
			name:       "Test with multiple retries",
			maxRetries: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := zaptest.NewLogger(t)
			got := NewHttpRetryClient(log, tt.maxRetries)

			if got.RetryMax != tt.maxRetries {
				t.Errorf("NewHttpRetryClient() RetryMax = %v, want %v", got.RetryMax, tt.maxRetries)
			}

			if got.CheckRetry == nil {
				t.Errorf("NewHttpRetryClient() CheckRetry is nil")
			}

			if got.Logger == nil {
				t.Errorf("NewHttpRetryClient() Logger is nil")
			}

			if got.HTTPClient.Transport == nil {
				t.Errorf("NewHttpRetryClient() HTTPClient.Transport is nil")
			}
		})
	}
}
