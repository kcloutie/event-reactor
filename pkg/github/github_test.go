package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestRemoveCommitComment(t *testing.T) {
	httpCalls := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/v3/repos/org/repo/comments/1" && req.Method == "DELETE" {
			httpCalls["/api/v3/repos/org/repo/comments/1"] = req.Method
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		if req.URL.Path == "/api/v3/repos/org/repo/comments/2" && req.Method == "DELETE" {
			httpCalls["/api/v3/repos/org/repo/comments/2"] = req.Method
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		t.Fatalf("Unexpected path: %v, method: %s", req.URL.Path, req.Method)
	}))
	defer server.Close()

	tests := []struct {
		name          string
		body          string
		commentId     int64
		regex         *regexp.Regexp
		wantHttpCalls map[string]string
	}{
		{
			name:          "Body matches regex, comment deleted successfully",
			body:          "test",
			commentId:     1,
			regex:         regexp.MustCompile("test"),
			wantHttpCalls: map[string]string{"/api/v3/repos/org/repo/comments/1": "DELETE"},
		},
		{
			name:          "Body matches regex, error deleting comment",
			body:          "test",
			commentId:     2,
			regex:         regexp.MustCompile("test"),
			wantHttpCalls: map[string]string{"/api/v3/repos/org/repo/comments/2": "DELETE"},
		},
		{
			name:          "Body does not match regex",
			body:          "not match",
			commentId:     1,
			regex:         regexp.MustCompile("test"),
			wantHttpCalls: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpCalls = map[string]string{}
			testLogger := zaptest.NewLogger(t)
			ctx := context.Background()
			c := New(ctx, testLogger, "org", "repo", "sha", "token", 1, server.URL, true)
			client, _ := c.NewClient()
			c.removeCommitComment(tt.body, tt.commentId, tt.regex, client)
			calls := []string{}
			for k, v := range httpCalls {
				calls = append(calls, fmt.Sprintf("%s:%s", k, v))
			}

			for k, v := range tt.wantHttpCalls {
				if httpCalls[k] != v {
					t.Errorf("Expected %s call to %s, got %s", v, k, strings.Join(calls, ", "))
				}
			}
			// if httpCalls["/api/v3/repos/org/repo/comments/1"] != "DELETE" {
			// 	calls := []string{}
			// 	for k, v := range httpCalls {
			// 		calls = append(calls, fmt.Sprintf("%s:%s", k, v))
			// 	}
			// 	t.Errorf("Expected DELETE call to /api/v3/repos/org/repo/comments/1, got %s", strings.Join(calls, ", "))
			// }

		})
	}
}

func TestRemoveCommitComments(t *testing.T) {
	httpCalls := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/v3/repos/org/repo/commits/sha/comments" && req.Method == "GET" {
			httpCalls["/api/v3/repos/org/repo/commits/sha/comments"] = req.Method
			rw.Write([]byte(`[{"id":1,"html_url":"url","body":"test"},{"id":2,"html_url":"url","body":"not match"}]`))
			return
		}
		if req.URL.Path == "/api/v3/repos/org/repo/comments/1" && req.Method == "DELETE" {
			httpCalls["/api/v3/repos/org/repo/comments/1"] = req.Method
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		t.Fatalf("Unexpected path: %v, method: %s", req.URL.Path, req.Method)
	}))
	defer server.Close()

	tests := []struct {
		name          string
		regex         *regexp.Regexp
		wantHttpCalls map[string]string
	}{
		{
			name:  "Body matches regex, comment deleted successfully",
			regex: regexp.MustCompile("test"),
			wantHttpCalls: map[string]string{
				"/api/v3/repos/org/repo/comments/1":           "DELETE",
				"/api/v3/repos/org/repo/commits/sha/comments": "GET",
			},
		},
		{
			name:  "Body does not match regex",
			regex: regexp.MustCompile("test"),
			wantHttpCalls: map[string]string{
				"/api/v3/repos/org/repo/commits/sha/comments": "GET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpCalls = map[string]string{}
			testLogger := zaptest.NewLogger(t)
			ctx := context.Background()
			// c := &github.GitHubConfiguration{
			// 		context: ctx,
			// 		logger:  testLogger,
			// 		Org:     "org",
			// 		Repo:    "repo",
			// 		CommitSha: "sha",
			// }
			// client := github.NewClient(&http.Client{}, server.URL)
			c := New(ctx, testLogger, "org", "repo", "sha", "token", 1, server.URL, true)
			client, _ := c.NewClient()
			err := c.removeCommitComments(client, tt.regex, "sha")
			if err != nil {
				t.Errorf("removeCommitComments() error = %v", err)
			}

			calls := []string{}
			for k, v := range httpCalls {
				calls = append(calls, fmt.Sprintf("%s:%s", k, v))
			}

			for k, v := range tt.wantHttpCalls {
				if httpCalls[k] != v {
					t.Errorf("Expected %s call to %s, got %s", v, k, strings.Join(calls, ", "))
				}
			}

		})
	}
}

func TestCleanExistingCommentsOnCommit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/v3/repos/org/repo/commits/sha/comments" && req.Method == "GET" {
			rw.Write([]byte(`[{"id":1,"html_url":"url","body":"test"}]`))
			return
		}
		if req.URL.Path == "/api/v3/repos/org/repo/commits/error/comments" && req.Method == "GET" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		if req.URL.Path == "/api/v3/repos/org/repo/comments/1" && req.Method == "DELETE" {
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		t.Fatalf("Unexpected path: %v, method: %s", req.URL.Path, req.Method)
	}))
	defer server.Close()

	tests := []struct {
		name    string
		comment string
		wantErr bool
		sha     string
	}{
		{
			name:    "removeCommitComments returns error",
			comment: "unknown",
			wantErr: true,
			sha:     "error",
		},
		{
			name:    "removeCommitComments does not return error",
			comment: "test",
			wantErr: false,
			sha:     "sha",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger := zaptest.NewLogger(t)
			ctx := context.Background()
			c := New(ctx, testLogger, "org", "repo", tt.sha, "token", 1, server.URL, true)
			err := c.CleanExistingCommentsOnCommit(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CleanExistingCommentsOnCommit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
