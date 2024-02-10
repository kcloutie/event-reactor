package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	ctx := context.Background()
	router := CreateRouter(ctx, 1)

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
			name: "basic",
			args: args{
				method: "GET",
				url:    "/healthz",
				body:   nil,
			},
			wantCode: 200,
			wantBody: `{"status":"ok"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.args.method, tt.args.url, tt.args.body)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
