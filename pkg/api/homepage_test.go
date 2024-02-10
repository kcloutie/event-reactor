package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/params/version"
	"github.com/stretchr/testify/assert"
)

func TestHome(t *testing.T) {
	ctx := context.Background()
	router := CreateRouter(ctx, 1)

	expectedHomePageBytes, err := os.ReadFile("testdata/homepage_expected.html")
	if err != nil {
		t.Errorf("TestHome() unable to reade the expected home page file - %v", err)
		return
	}

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
				url:    "/",
				body:   nil,
			},
			wantCode: 200,
			wantBody: string(expectedHomePageBytes),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version.BuildVersion = "v1.1.1"
			version.Commit = "commit1"
			version.BuildTime = "buildtime"
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.args.method, tt.args.url, tt.args.body)
			router.ServeHTTP(w, req)

			// os.WriteFile("got.html", w.Body.Bytes(), 0644)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
