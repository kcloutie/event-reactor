package http

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCacheKey(t *testing.T) {
	tests := []struct {
		name   string
		method string
		url    string
		want   string
	}{
		{
			name:   "Test with GET request",
			method: http.MethodGet,
			url:    "http://test.com",
			want:   "http://test.com",
		},
		{
			name:   "Test with POST request",
			method: http.MethodPost,
			url:    "http://test.com",
			want:   "POST http://test.com",
		},
		{
			name:   "Test with PUT request",
			method: http.MethodPut,
			url:    "http://test.com",
			want:   "PUT http://test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			if got := cacheKey(req); got != tt.want {
				t.Errorf("cacheKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCachedResponse(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		url      string
		cacheVal []byte
		wantErr  bool
	}{
		{
			name:     "Test with GET request and valid cache",
			method:   http.MethodGet,
			url:      "http://test.com",
			cacheVal: []byte("HTTP/1.1 200 OK\r\n\r\n"),
			wantErr:  false,
		},
		{
			name:     "Test with POST request and valid cache",
			method:   http.MethodPost,
			url:      "http://test.com",
			cacheVal: []byte("HTTP/1.1 200 OK\r\n\r\n"),
			wantErr:  false,
		},
		{
			name:     "Test with PUT request and valid cache",
			method:   http.MethodPut,
			url:      "http://test.com",
			cacheVal: []byte("HTTP/1.1 200 OK\r\n\r\n"),
			wantErr:  false,
		},
		{
			name:     "Test with GET request and invalid cache",
			method:   http.MethodGet,
			url:      "http://test.com",
			cacheVal: []byte("INVALID CACHE"),
			wantErr:  true,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			c := NewMemoryCache()
			c.Set(cacheKey(req), tt.cacheVal)

			_, err := CachedResponse(c, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CachedResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryCache_Get(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		cacheVal []byte
		wantOk   bool
	}{
		{
			name:     "Test with existing key",
			key:      "existingKey",
			cacheVal: []byte("value"),
			wantOk:   true,
		},
		{
			name:     "Test with non-existing key",
			key:      "nonExistingKey",
			cacheVal: nil,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMemoryCache()
			if tt.cacheVal != nil {
				c.Set(tt.key, tt.cacheVal)
			}

			gotVal, gotOk := c.Get(tt.key)
			if gotOk != tt.wantOk {
				t.Errorf("MemoryCache.Get() ok = %v, want %v", gotOk, tt.wantOk)
			}

			if gotOk && !bytes.Equal(gotVal, tt.cacheVal) {
				t.Errorf("MemoryCache.Get() val = %v, want %v", gotVal, tt.cacheVal)
			}
		})
	}
}

func TestMemoryCache_Set(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  []byte
	}{
		{
			name: "Test with key1",
			key:  "key1",
			val:  []byte("value1"),
		},
		{
			name: "Test with key2",
			key:  "key2",
			val:  []byte("value2"),
		},
		{
			name: "Test with key3",
			key:  "key3",
			val:  []byte("value3"),
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMemoryCache()
			c.Set(tt.key, tt.val)

			gotVal, ok := c.Get(tt.key)
			if !ok {
				t.Errorf("MemoryCache.Set() key = %v not found", tt.key)
			}

			if !bytes.Equal(gotVal, tt.val) {
				t.Errorf("MemoryCache.Set() val = %v, want %v", gotVal, tt.val)
			}
		})
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  []byte
	}{
		{
			name: "Test with key1",
			key:  "key1",
			val:  []byte("value1"),
		},
		{
			name: "Test with key2",
			key:  "key2",
			val:  []byte("value2"),
		},
		{
			name: "Test with key3",
			key:  "key3",
			val:  []byte("value3"),
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMemoryCache()
			c.Set(tt.key, tt.val)

			c.Delete(tt.key)

			_, ok := c.Get(tt.key)
			if ok {
				t.Errorf("MemoryCache.Delete() key = %v found, but it should have been deleted", tt.key)
			}
		})
	}
}

func TestVaryMatches(t *testing.T) {
	tests := []struct {
		name       string
		cachedResp *http.Response
		req        *http.Request
		want       bool
	}{
		{
			name: "Test with matching headers",
			cachedResp: &http.Response{
				Header: http.Header{
					"Vary":                 []string{"Test-Header"},
					"X-Varied-Test-Header": []string{"test value"},
				},
			},
			req: &http.Request{
				Header: http.Header{
					"Test-Header": []string{"test value"},
				},
			},
			want: true,
		},
		{
			name: "Test with non-matching headers",
			cachedResp: &http.Response{
				Header: http.Header{
					"Vary":                 []string{"Test-Header"},
					"X-Varied-Test-Header": []string{"test value"},
				},
			},
			req: &http.Request{
				Header: http.Header{
					"Test-Header": []string{"different value"},
				},
			},
			want: false,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := varyMatches(tt.cachedResp, tt.req); got != tt.want {
				t.Errorf("varyMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetCacheExpireHeader(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		want    bool
	}{
		{
			name: "Test with CacheExpireHeaderKey not present",
			headers: http.Header{
				"Other-Header": []string{"value"},
			},
			want: true,
		},
		{
			name: "Test with CacheExpireHeaderKey present",
			headers: http.Header{
				CacheExpireHeaderKey: []string{"value"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetCacheExpireHeader(tt.headers)
			_, exists := tt.headers[CacheExpireHeaderKey]

			if exists != tt.want {
				t.Errorf("SetCacheExpireHeader() = %v, want %v", exists, tt.want)
			}
		})
	}
}

func TestDate(t *testing.T) {
	tests := []struct {
		name        string
		respHeaders http.Header
		wantDate    time.Time
		wantErr     string
	}{
		{
			name: "Test with valid date header",
			respHeaders: http.Header{
				http.CanonicalHeaderKey("date"): []string{"Sun, 28 Feb 2016 08:49:37 GMT"},
			},
			wantDate: time.Date(2016, time.February, 28, 8, 49, 37, 0, time.UTC),
			wantErr:  "",
		},
		{
			name: "Test with invalid date header",
			respHeaders: http.Header{
				http.CanonicalHeaderKey("date"): []string{"invalid date"},
			},
			wantDate: time.Time{},
			wantErr:  "invalid date",
		},
		{
			name:        "Test with no date header",
			respHeaders: http.Header{},
			wantDate:    time.Time{},
			wantErr:     "no Date header",
		},
		{
			name: "Test with blank date header",
			respHeaders: http.Header{
				http.CanonicalHeaderKey("date"): []string{""},
			},
			wantDate: time.Time{},
			wantErr:  "no Date header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, gotErr := Date(tt.respHeaders)
			if !gotDate.Equal(tt.wantDate) {
				t.Errorf("Date() date = %v, want %v", gotDate, tt.wantDate)
			}

			if gotErr == nil && tt.wantErr == "" {
				return
			}
			if !strings.Contains(gotErr.Error(), tt.wantErr) {
				t.Errorf("Date() error = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestGetFreshness(t *testing.T) {
	tests := []struct {
		name               string
		respHeaders        http.Header
		reqHeaders         http.Header
		useLocalCacheTimes bool
		want               int
	}{
		{
			name: "Test with no-cache directive in request",
			respHeaders: http.Header{
				"Cache-Control": []string{"public, max-age=200"},
			},
			reqHeaders: http.Header{
				"Cache-Control": []string{"no-cache"},
			},
			useLocalCacheTimes: false,
			want:               transparent,
		},
		{
			name: "Test with no-cache directive in response",
			respHeaders: http.Header{
				"Cache-Control": []string{"no-cache"},
			},
			reqHeaders: http.Header{
				"Cache-Control": []string{"public, max-age=200"},
			},
			useLocalCacheTimes: false,
			want:               stale,
		},
		{
			name: "Test with only-if-cached directive in request",
			respHeaders: http.Header{
				"Cache-Control": []string{"public, max-age=200"},
			},
			reqHeaders: http.Header{
				"Cache-Control": []string{"only-if-cached"},
			},
			useLocalCacheTimes: false,
			want:               fresh,
		},
		{
			name: "Test with max-age directive in request",
			respHeaders: http.Header{
				"Cache-Control": []string{"public, max-age=200"},
				"Date":          []string{time.Now().Add(-100 * time.Second).Format(time.RFC1123)},
			},
			reqHeaders: http.Header{
				"Cache-Control": []string{"max-age=150"},
			},
			useLocalCacheTimes: false,
			want:               fresh,
		},
		// {
		// 	name: "Test with max-age directive in response",
		// 	respHeaders: http.Header{
		// 		"Cache-Control": []string{"public, max-age=50"},
		// 		"Date":          []string{time.Now().Add(-100 * time.Second).Format(time.RFC1123)},
		// 	},
		// 	reqHeaders: http.Header{
		// 		"Cache-Control": []string{"public, max-age=200"},
		// 	},
		// 	useLocalCacheTimes: false,
		// 	want:               stale,
		// },
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFreshness(tt.respHeaders, tt.reqHeaders, tt.useLocalCacheTimes); got != tt.want {
				t.Errorf("getFreshness() = %v, want %v", got, tt.want)
			}
		})
	}
}
