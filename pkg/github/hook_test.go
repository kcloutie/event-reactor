package github

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	payload := []byte(`{"key": "value"}`)
	tests := []struct {
		name    string
		secret  []byte
		req     *http.Request
		wantErr error
	}{
		{
			name:   "Test with valid signature",
			secret: []byte("test-secret"),
			req: func() *http.Request {
				req := httptest.NewRequest("POST", "http://example.com", bytes.NewBuffer(payload))
				req.Header.Set("x-hub-signature-256", "sha256=ba198e310b91b9724f73800379ac715ec3253bc6d2f916669b6508e53e0aa07b")
				return req
			}(),
			wantErr: nil,
		},
		{
			name:   "Test with invalid signature",
			secret: []byte("test-secret"),
			req: func() *http.Request {
				req := httptest.NewRequest("POST", "http://example.com", bytes.NewBuffer(payload))
				req.Header.Set("x-hub-signature-256", "sha256=ba198e310b91b9724f73800379ac715ec3253bc6d2f916669b6508e53e0aa08b")
				return req
			}(),
			wantErr: errors.New("invalid signature"),
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.secret, tt.req)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
