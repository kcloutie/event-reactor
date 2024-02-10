package webex

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap/zaptest"
)

func TestProcessEvent(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		rw.Write([]byte(`OK`))
	}))
	// Close the server when test finishes
	defer server.Close()

	tests := []struct {
		name        string
		reactorName string
		properties  map[string]config.PropertyAndValue
		wantErr     bool
	}{
		{
			name:        "Test with card",
			reactorName: "TestReactor",
			properties: map[string]config.PropertyAndValue{
				"apiUrl": {
					Value: server.URL,
				},
				"card": {
					Value: `{"type":"AdaptiveCard","body":[]}`,
				},
				"spaceId": {
					Value: "spaceid",
				},
				"token": {
					Value: "token",
				},
			},
			wantErr: false,
		},
		{
			name:        "Test with message",
			reactorName: "TestReactor",
			properties: map[string]config.PropertyAndValue{
				"apiUrl": {
					Value: server.URL,
				},
				"message": {
					Value: `test123`,
				},
				"spaceId": {
					Value: "spaceid",
				},
				"token": {
					Value: "token",
				},
			},
			wantErr: false,
		},
		{
			name:        "Test with missing properties",
			reactorName: "TestReactor",
			properties: map[string]config.PropertyAndValue{
				"message": {
					Value: `test123`,
				},
			},
			wantErr: true,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Reactor{
				reactorName: tt.reactorName,
				reactorConfig: config.ReactorConfig{
					Properties: tt.properties,
				},
				Log: zaptest.NewLogger(t),
			}

			data := &message.EventData{}
			ctx := context.Background()

			err := v.ProcessEvent(ctx, data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
