package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
	"github.com/stretchr/testify/assert"
)

func TestPubSubListener(t *testing.T) {
	servConf := config.ServerConfiguration{
		LoadTestReactor: true,
		ReactorConfigs: []config.ReactorConfig{
			{
				Name: "testReactor",
				Type: "testReactor",
				Properties: map[string]config.PropertyAndValue{
					"test": {
						Value: "test",
					},
					"message": {
						Value: "test",
					},
				},
			},
		},
	}

	ctx := config.WithCtx(context.Background(), &servConf)
	router := CreateRouter(ctx, 1)

	badBody := strings.NewReader(`{"test":"test"}`)
	goodPayload, err := os.ReadFile("testdata/pubsubPayload.json")
	if err != nil {
		t.Fatal(err)
	}
	goodPayloadBytes := strings.NewReader(string(goodPayload))
	type args struct {
		method string
		url    string
		body   io.Reader
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantBody string
	}{
		{
			name: "no body",
			args: args{
				method: "POST",
				url:    "/api/v1/" + settings.PubSubEndpoint,
				body:   nil,
			},
			wantCode: 400,
			wantBody: "[{\"type\":\"pub/sub-get-request-body\",\"title\":\"pub/sub Get Request Body\",\"status\":400,\"detail\":\"request body was empty, request cannot be processed\",\"instance\":\"pubsub\"}]",
		},

		{
			name: "bad body",
			args: args{
				method: "POST",
				url:    "/api/v1/" + settings.PubSubEndpoint,
				body:   badBody,
			},
			wantCode: 400,
			wantBody: "[{\"type\":\"convert-pubsub-message\",\"title\":\"Convert Pub/Sub Message\",\"status\":400,\"detail\":\"message property not found in the pub/sub message\",\"instance\":\"pubsub\"}]",
		},
		{
			name: "good payload, test reactor",
			args: args{
				method: "POST",
				url:    "/api/v1/" + settings.PubSubEndpoint,
				body:   goodPayloadBytes,
			},
			wantCode: 200,
			wantBody: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.args.method, tt.args.url, tt.args.body)
			router.ServeHTTP(w, req)
			ff := w.Body.String()
			fmt.Println(ff)
			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
