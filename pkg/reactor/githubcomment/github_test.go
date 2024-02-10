package githubcomment

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/github"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap/zaptest"
)

func TestReactor_GetReactorConfig(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		props     map[string]config.PropertyAndValue
		eventData *message.EventData
		want      *ReactorConfig
		wantErr   bool
	}{
		{
			name: "All properties are valid",
			eventData: &message.EventData{
				Data: map[string]interface{}{
					"token":         "test token",
					"org":           "test org",
					"repo":          "test repo",
					"commitSha":     "test commitSha",
					"prNumber":      "123",
					"enterpriseUrl": "test enterpriseUrl",
				},
			},
			props: map[string]config.PropertyAndValue{
				"planTaskName": {Value: toPtrString("test planTaskName")},
				"removeExistingCommentsFromAllPullRequestCommits": {Value: toPtrString("true")},
				"removeExistingPullRequestComments":               {Value: toPtrString("false")},
				"removeDuplicateCommitComments":                   {Value: toPtrString("true")},
				"token":                                           {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.token"}}},
				"org":                                             {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.org"}}},
				"repo":                                            {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.repo"}}},
				"commitSha":                                       {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.commitSha"}}},
				"prNumber":                                        {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.prNumber"}}},
				"enterpriseUrl":                                   {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.enterpriseUrl"}}},
				"body":                                            {Value: toPtrString("Body of commit {{ .data.org }}/{{ .data.repo }}/{{ .data.commitSha }}")},
			},
			want: &ReactorConfig{
				PlanTaskName: "test planTaskName",
				RemoveExistingCommentsFromAllPullRequestCommits: true,
				RemoveExistingPullRequestComments:               false,
				RemoveDuplicateCommitComments:                   true,
				GithubConfig:                                    github.New(context.Background(), zaptest.NewLogger(t), "test org", "test repo", "test commitSha", "test token", 123, "test enterpriseUrl", false),
			},
			wantErr: false,
		},
		{
			name: "Invalid boolean value",
			eventData: &message.EventData{
				Data: map[string]interface{}{
					"token":         "test token",
					"org":           "test org",
					"repo":          "test repo",
					"commitSha":     "test commitSha",
					"prNumber":      "123",
					"enterpriseUrl": "test enterpriseUrl",
				},
			},
			props: map[string]config.PropertyAndValue{
				"planTaskName": {Value: toPtrString("test planTaskName")},
				"removeExistingCommentsFromAllPullRequestCommits": {Value: toPtrString("invalid")},
				"removeExistingPullRequestComments":               {Value: toPtrString("true")},
				"removeDuplicateCommitComments":                   {Value: toPtrString("true")},
				"token":                                           {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.token"}}},
				"org":                                             {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.org"}}},
				"repo":                                            {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.repo"}}},
				"commitSha":                                       {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.commitSha"}}},
				"prNumber":                                        {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.prNumber"}}},
				"enterpriseUrl":                                   {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.enterpriseUrl"}}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "removeExistingPullRequestComments not supplied",
			eventData: &message.EventData{
				Data: map[string]interface{}{
					"token":         "test token",
					"org":           "test org",
					"repo":          "test repo",
					"commitSha":     "test commitSha",
					"prNumber":      "123",
					"enterpriseUrl": "test enterpriseUrl",
				},
			},
			props: map[string]config.PropertyAndValue{
				"planTaskName": {Value: toPtrString("test planTaskName")},
				"removeExistingCommentsFromAllPullRequestCommits": {Value: toPtrString("true")},
				"removeDuplicateCommitComments":                   {Value: toPtrString("true")},
				"token":                                           {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.token"}}},
				"org":                                             {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.org"}}},
				"repo":                                            {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.repo"}}},
				"commitSha":                                       {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.commitSha"}}},
				"prNumber":                                        {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.prNumber"}}},
				"enterpriseUrl":                                   {PayloadValue: &config.PayloadValueRef{PropertyPaths: []string{"data.enterpriseUrl"}}},
				"body":                                            {Value: toPtrString("Body of commit {{ .data.org }}/{{ .data.repo }}/{{ .data.commitSha }}")},
			},
			want: &ReactorConfig{
				PlanTaskName: "test planTaskName",
				RemoveExistingCommentsFromAllPullRequestCommits: true,
				RemoveExistingPullRequestComments:               true,
				RemoveDuplicateCommitComments:                   true,
				GithubConfig:                                    github.New(context.Background(), zaptest.NewLogger(t), "test org", "test repo", "test commitSha", "test token", 123, "test enterpriseUrl", false),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: tt.props,
				},
			}
			got, err := p.GetReactorConfig(context.Background(), tt.eventData, testLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reactor.GetReactorConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got.PlanTaskName, tt.want.PlanTaskName) {
				t.Errorf("Reactor.GetReactorConfig() PlanTaskName = %v, want %v", got.PlanTaskName, tt.want.PlanTaskName)
			}
			if !reflect.DeepEqual(got.RemoveExistingCommentsFromAllPullRequestCommits, tt.want.RemoveExistingCommentsFromAllPullRequestCommits) {
				t.Errorf("Reactor.GetReactorConfig() RemoveExistingCommentsFromAllPullRequestCommits = %v, want %v", got.RemoveExistingCommentsFromAllPullRequestCommits, tt.want.RemoveExistingCommentsFromAllPullRequestCommits)
			}
			if !reflect.DeepEqual(got.RemoveExistingPullRequestComments, tt.want.RemoveExistingPullRequestComments) {
				t.Errorf("Reactor.GetReactorConfig() RemoveExistingPullRequestComments = %v, want %v", got.RemoveExistingPullRequestComments, tt.want.RemoveExistingPullRequestComments)
			}
			if !reflect.DeepEqual(got.RemoveDuplicateCommitComments, tt.want.RemoveDuplicateCommitComments) {
				t.Errorf("Reactor.GetReactorConfig() RemoveDuplicateCommitComments = %v, want %v", got.RemoveDuplicateCommitComments, tt.want.RemoveDuplicateCommitComments)
			}

			if !reflect.DeepEqual(got.GithubConfig.EnterpriseUrl, tt.want.GithubConfig.EnterpriseUrl) {
				t.Errorf("Reactor.GetGithubConfig() EnterpriseUrl = %v, want %v", got.GithubConfig.EnterpriseUrl, tt.want.GithubConfig.EnterpriseUrl)
			}
			if !reflect.DeepEqual(got.GithubConfig.CommitSha, tt.want.GithubConfig.CommitSha) {
				t.Errorf("Reactor.GetGithubConfig() CommitSha = %v, want %v", got.GithubConfig.CommitSha, tt.want.GithubConfig.CommitSha)
			}
			if !reflect.DeepEqual(got.GithubConfig.Org, tt.want.GithubConfig.Org) {
				t.Errorf("Reactor.GetGithubConfig() Org = %v, want %v", got.GithubConfig.Org, tt.want.GithubConfig.Org)
			}
			if !reflect.DeepEqual(got.GithubConfig.PrNumber, tt.want.GithubConfig.PrNumber) {
				t.Errorf("Reactor.GetGithubConfig() PrNumber = %v, want %v", got.GithubConfig.PrNumber, tt.want.GithubConfig.PrNumber)
			}
			if !reflect.DeepEqual(got.GithubConfig.Repo, tt.want.GithubConfig.Repo) {
				t.Errorf("Reactor.GetGithubConfig() Repo = %v, want %v", got.GithubConfig.Repo, tt.want.GithubConfig.Repo)
			}

			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Reactor.GetReactorConfig() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func toPtrString(val string) *string {
	return &val
}

func TestReactor_ExecuteReactor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		pp := req.URL.Path
		cc := req.Method
		fmt.Println(pp)
		fmt.Println(cc)

		if req.Header["Authorization"][0] != fmt.Sprintf("Bearer %v", "token") {
			t.Fatalf("Invalid token: '%v'", req.Header["Authorization"][0])
		}

		if req.URL.Path == "/api/v3/repos/org/repo/commits/sha/comments" {
			if req.Method == "GET" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
			if req.Method == "POST" {
				rw.Write([]byte(`{"id":1,"html_url":"url","body":"heading\nbody"}`))
				return
			}
		}
		if req.URL.Path == "/api/v3/repos/org/repo/comments/1" {
			if req.Method == "DELETE" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
		}

		if req.URL.Path == "/api/v3/repos/org/repo/issues/1/comments" {
			if req.Method == "GET" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
			if req.Method == "POST" {
				rw.Write([]byte(`{"id":1,"html_url":"url","body":"heading\nbody"}`))
				return
			}
		}

		if req.URL.Path == "/api/v3/repos/org/repo/issues/comments/1" {
			if req.Method == "DELETE" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
		}

		if req.URL.Path == "/api/v3/repos/org/repo/pulls/1/commits" {
			if req.Method == "GET" {
				rw.Write([]byte(`[{"sha":"1","html_url":"url"}]`))
				return
			}
		}

		if req.URL.Path == "/api/v3/repos/org/repo/commits/1/comments" {
			if req.Method == "GET" {
				rw.Write([]byte(`[{"id":1,"html_url":"url","body":"heading\nbody"}]`))
				return
			}
		}
		fmt.Printf("Unknown path: %v, %s\n", req.URL.Path, req.Method)
	}))
	defer server.Close()
	type args struct {
		data          *message.EventData
		reactorConfig config.ReactorConfig
	}
	tests := []struct {
		name    string
		v       *Reactor
		args    args
		wantErr bool
	}{
		{
			name: "Valid properties non PR",
			v:    New(),
			args: args{
				data: &message.EventData{
					Data: map[string]interface{}{},
				},
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"token":         {Value: toPtrString("token")},
						"org":           {Value: toPtrString("org")},
						"repo":          {Value: toPtrString("repo")},
						"commitSha":     {Value: toPtrString("sha")},
						"prNumber":      {Value: toPtrString("-1")},
						"enterpriseUrl": {Value: toPtrString(server.URL)},
						"heading":       {Value: toPtrString("heading")},
						"body":          {Value: toPtrString("body")},
					},
				},
			},
		},
		{
			name: "Valid properties PR",
			v:    New(),
			args: args{
				data: &message.EventData{
					Data: map[string]interface{}{},
				},
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"token":         {Value: toPtrString("token")},
						"org":           {Value: toPtrString("org")},
						"repo":          {Value: toPtrString("repo")},
						"commitSha":     {Value: toPtrString("sha")},
						"prNumber":      {Value: toPtrString("1")},
						"enterpriseUrl": {Value: toPtrString(server.URL)},
						"heading":       {Value: toPtrString("heading")},
						"body":          {Value: toPtrString("body")},
					},
				},
			},
		},
		{
			name: "Valid properties PR, removeExistingCommentsFromAllPullRequestCommits true",
			v:    New(),
			args: args{
				data: &message.EventData{
					Data: map[string]interface{}{},
				},
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"token":         {Value: toPtrString("token")},
						"org":           {Value: toPtrString("org")},
						"repo":          {Value: toPtrString("repo")},
						"commitSha":     {Value: toPtrString("sha")},
						"prNumber":      {Value: toPtrString("1")},
						"enterpriseUrl": {Value: toPtrString(server.URL)},
						"heading":       {Value: toPtrString("heading")},
						"body":          {Value: toPtrString("body")},
						"removeExistingCommentsFromAllPullRequestCommits": {Value: toPtrString("true")},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger := zaptest.NewLogger(t)
			ctx := context.Background()
			tt.v.SetLogger(testLogger)
			tt.v.SetReactor(tt.args.reactorConfig)

			err := tt.v.ProcessEvent(ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reactor.ExecuteReactor() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
