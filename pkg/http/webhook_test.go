package http

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestWebhookConfig_SendWebhook(t *testing.T) {
	body := "test body"
	secret := []byte("secret")
	rawSignature := SignBody(secret, []byte(body))
	signature := hex.EncodeToString(rawSignature)
	additionalHeaders := map[string]string{
		"X-Test-Header": "test",
	}
	bearerToken := "test-token"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/signature" {
			if req.Header.Get("X-Event-Reactor-Signature") != signature {
				t.Errorf("WebhookConfig.SendWebhook() signature = %v, want %v", req.Header.Get("X-Event-Reactor-Signature"), signature)
				return
			}
			rw.WriteHeader(http.StatusOK)
			return
		}
		if req.URL.Path == "/headers" {
			for key, value := range additionalHeaders {
				hVal, exists := req.Header[key]
				if !exists {
					t.Errorf("WebhookConfig.SendWebhook() header %s does not exist", key)
					return
				}
				if hVal[0] != value {
					t.Errorf("WebhookConfig.SendWebhook() header %s = %v, want %v", key, hVal, value)
					return
				}
			}

			rw.WriteHeader(http.StatusOK)
			return
		}
		if req.URL.Path == "/token" {
			if req.Header.Get("Authorization") != "Bearer "+bearerToken {
				t.Errorf("WebhookConfig.SendWebhook() Authorization = %v, want %v", req.Header.Get("Authorization"), "Bearer "+bearerToken)
				return
			}
			rw.WriteHeader(http.StatusOK)
			return
		}
		if req.URL.Path == "/fail" {

			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name    string
		c       *WebhookConfig
		wantErr bool
	}{
		{
			name: "signature",
			c: &WebhookConfig{
				Log:             zaptest.NewLogger(t),
				Url:             server.URL + "/signature",
				HookSecret:      string(secret),
				Body:            body,
				SignatureHeader: "X-Event-Reactor-Signature",
			},
		},
		{
			name: "additional headers",
			c: &WebhookConfig{
				Log:               zaptest.NewLogger(t),
				Url:               server.URL + "/headers",
				HookSecret:        string(secret),
				Body:              body,
				SignatureHeader:   "X-Event-Reactor-Signature",
				AdditionalHeaders: additionalHeaders,
			},
		},
		{
			name: "bearer token",
			c: &WebhookConfig{
				Log:   zaptest.NewLogger(t),
				Url:   server.URL + "/token",
				Body:  body,
				Token: bearerToken,
			},
		},
		{
			name: "500 error",
			c: &WebhookConfig{
				Log:  zaptest.NewLogger(t),
				Url:  server.URL + "/fail",
				Body: body,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.SendWebhook(); (err != nil) != tt.wantErr {
				t.Errorf("WebhookConfig.SendWebhook() error = %v", err)
			}
		})
	}
}
