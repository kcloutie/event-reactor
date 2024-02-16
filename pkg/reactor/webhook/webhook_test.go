package webhook

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap"
)

func TestReactor_ProcessEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		authHeader, exists := req.Header["Authorization"]
		if !exists {
			t.Fatalf("Authorization header does not exist")
		}

		if authHeader[0] != fmt.Sprintf("Bearer %v", "token") {
			t.Fatalf("Invalid token: '%v'", authHeader[0])
		}

		if req.URL.Path == "/valid" {
			rw.WriteHeader(http.StatusOK)
			return
		}

		fmt.Printf("Unknown path: %v, %s\n", req.URL.Path, req.Method)
	}))
	defer server.Close()
	tests := []struct {
		name    string
		v       *Reactor
		data    *message.EventData
		wantErr bool
	}{
		{
			name: "Test with valid parameters",
			v: &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"url": {
							Value: fmt.Sprintf(server.URL + "/valid"),
						},
						"webhookSecret": {
							Value: "test123",
						},
						"bodyTemplate": {
							Value: `Test123 {{ .data.child1.child2.test_child }}`,
						},
						"token": {
							Value: "token",
						},
						"maxRetries": {
							Value: "5",
						},
						"additionalHeaders": {
							Value: map[string]string{
								"header1": "value1",
							},
						},
					},
				},
				Log: zap.NewNop(),
			},
			data: &message.EventData{
				Data: map[string]interface{}{
					"test": "data test value",
					"child1": map[string]interface{}{
						"child2": map[string]interface{}{
							"test_child": "data test child value",
						},
					},
				},
				Attributes: map[string]string{},
				ID:         "test-id",
			},
			wantErr: false,
		},
		{
			name: "Test with missing required properties",
			v: &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"jsonDepth": {
							Value: "4",
						},
						"parameters": {
							Value: map[string]interface{}{
								"param1": "value1",
								"param2": "value2",
							},
						},
					},
				},
				Log: zap.NewNop(),
			},
			data: &message.EventData{
				Data: map[string]interface{}{
					"test": "data test value",
					"child1": map[string]interface{}{
						"child2": map[string]interface{}{
							"test_child": "data test child value",
						},
					},
				},
				Attributes: map[string]string{},
				ID:         "test-id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.v.ProcessEvent(context.Background(), tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reactor.ProcessEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
