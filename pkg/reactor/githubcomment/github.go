package githubcomment

import (
	"context"
	"fmt"

	"strconv"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/github"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/reactor"
	"github.com/kcloutie/event-reactor/pkg/template"

	"go.uber.org/zap"
)

var _ reactor.ReactorInterface = (*Reactor)(nil)

type Reactor struct {
	Log           *zap.Logger
	reactorName   string
	reactorConfig config.ReactorConfig
}

type ReactorConfig struct {
	// Terraform plan task name, set the value when using terraform and additional data will be displayed in the github comment
	PlanTaskName string `json:"planTaskName,omitempty"`

	// True or false whether existing pipeline comments on all PR commits will be removed.
	// A pull request can contain multiple commits. Depending on how the pull request is pushed up, a pipeline may execute on each commit
	// and in turn each commit will contain a comment. When this property is true, every comment on every commit related to the pull request
	// will be removed. The default should be set to false in order to keep the pipelineRun history of each pipeline execution
	RemoveExistingCommentsFromAllPullRequestCommits bool `json:"removeExistingCommentsFromAllPullRequestCommits,omitempty"`

	// True or false whether existing pipeline comments on PR should be removed.
	// Each time the pipeline executes it will write a new issue comment to the pull request of the results of the pipelineRun
	// When this is set to true, all existing issue comments created by the pipeline will be removed. This is to make it easier
	// for people reviewing to find the results of the latest pipelineRun.
	// The default for this should be set to true so that the pull request only contains the latest pipelineRun result comment.
	// NOTE: The same comment is written to the latest commit and can be viewed by looking at the commit
	RemoveExistingPullRequestComments bool `json:"removeExistingPullRequestComments,omitempty"`

	// True or false to remove duplicate pipeline comments on a single (latest) commit
	// Each time the pipeline executes, this operator will write a comment on the latest commit. When the pipeline executes for a second time
	// on the same commit, a second comment will be written to that commit.
	// When this is set to true, any existing comments (created by this operator) on the latest commit will be removed keeping only the latest comment
	RemoveDuplicateCommitComments bool `json:"removeDuplicateCommitComments,omitempty"`

	GithubConfig *github.GitHubConfiguration

	Heading string
	Body    string
}

func New() *Reactor {
	return &Reactor{
		reactorName: "github/comment",
	}
}

func (v *Reactor) SetLogger(logger *zap.Logger) {
	v.Log = logger
}
func (v *Reactor) GetName() string {
	return v.reactorName
}

func (v *Reactor) GetConfigExample() string {
	return `
  reactorConfigs:
  - name: test_github
    celExpressionFilter: attributes.test == 'github'
    type: github/comment
    properties:
      heading:
        value: Testing GitHub Comment
      body:
        fromFile: test/testdata/github-comment.md
      token:
        fromEnv: GIT_TOKEN
      org:
        payloadValue:
          propertyPaths:
          - data.githubOrg
      repo:
        payloadValue:
          propertyPaths:
          - data.githubRepo
      commitSha:
        payloadValue:
          propertyPaths:
          - data.githubHeadSha
      prNumber:
        payloadValue:
          propertyPaths:
          - data.githubPrNum
      enterpriseUrl:
        value: https://github.someplace.com
`
}

func (v *Reactor) GetDescription() string {
	return "This reactor writes comments on commits and pull requests. It requires a GitHub token with appropriate permissions to interact with the specified repository. Key inputs include the organization, repository, commit SHA, and pull request number. One of the standout features of this reactor is its support for Go templating, which can be used to customize the heading and body of the comments. The heading also plays a crucial role in identifying previous comments for deletion. Moreover, the reactor offers a suite of configuration options for enhanced control. These include the ability to purge existing comments from all commits associated with a pull request, remove comments from the pull request itself, and eliminate duplicate commit comments."
}

func (v *Reactor) SetReactor(reactor config.ReactorConfig) {
	v.reactorConfig = reactor
}

func (v *Reactor) ProcessEvent(ctx context.Context, data *message.EventData) error {
	v.Log = v.Log.With(zap.String("reactor", v.reactorName))
	_, err := reactor.HasRequiredProperties(v.reactorConfig.Properties, v.GetRequiredPropertyNames())
	if err != nil {
		return err
	}

	reactorConfig, err := v.GetReactorConfig(ctx, data, v.Log)
	if err != nil {
		return err
	}

	v.Log = v.Log.With(zap.String("org", reactorConfig.GithubConfig.Org), zap.String("repo", reactorConfig.GithubConfig.Repo), zap.String("commitSha", reactorConfig.GithubConfig.CommitSha), zap.Int("pr", reactorConfig.GithubConfig.PrNumber), zap.String("enterpriseUrl", reactorConfig.GithubConfig.EnterpriseUrl))

	if reactorConfig.RemoveExistingCommentsFromAllPullRequestCommits {
		v.Log.Info("Cleaning up existing commit comments")
		reactorConfig.GithubConfig.CleanExistingCommentsOnAllPullRequestCommits(reactorConfig.Heading)
		// v.log.Info("Finished cleaning up existing comments on all commits of the pull request")
	} else {
		v.Log.Info("RemoveExistingCommentsFromAllPullRequestCommits was set to false, skipping the deletion of existing comments")
	}

	if reactorConfig.GithubConfig.PrNumber > 0 {
		if reactorConfig.RemoveExistingPullRequestComments {
			v.Log.Info("Cleaning up existing pull request comments")
			reactorConfig.GithubConfig.CleanExistingCommentsOnPullRequest(reactorConfig.Heading)
			// v.log.Info("Finished cleaning up existing comments on the pull request")
		} else {
			v.Log.Info("RemoveExistingPullRequestComments was set to false, skipping the deletion of existing comments")
		}
	} else {
		v.Log.Info("Pull request number was not greater than 0, skipping the deletion of existing comments")
	}

	// Would normally generate the comment body here, but the body is not generated using templates
	v.Log.Info("Creating commit comment")
	newComment, err := reactorConfig.GithubConfig.WriteCommitComment(reactorConfig.Body, reactorConfig.Heading, reactorConfig.RemoveDuplicateCommitComments)
	if err != nil {
		// v.log.Error("failed to write the github commit comment", zap.Error(err))
		return fmt.Errorf("unable to write github commit comment. Error: %v", err)
	}

	v.Log = v.Log.With(zap.String("commitCommentUrl", newComment.GetHTMLURL()))
	v.Log.Info("github commit comment has been created")

	if reactorConfig.GithubConfig.PrNumber > 0 {
		v.Log.Info("Creating pull request comment")
		newComment, err := reactorConfig.GithubConfig.WritePullRequestComment(reactorConfig.Body)
		if err != nil {
			// return githubToken, fmt.Errorf("unable to write github pull request comment. Error: %v", err)
			return fmt.Errorf("unable to write github pull request comment. Error: %v", err)
		}

		v.Log = v.Log.With(zap.String("PrCommentUrl", newComment.GetHTMLURL()))
		v.Log.Info("github pull request comment has been created")
	} else {
		v.Log.Info("Pull request number was not greater than 0, skipping the creation of the pull request comment")
	}

	return nil
}

func (v *Reactor) GetReactorConfig(ctx context.Context, data *message.EventData, log *zap.Logger) (*ReactorConfig, error) {

	templateConfig := template.NewRenderTemplateOptions()
	reactor.SetGoTemplateOptionValues(ctx, v.Log, &templateConfig, v.reactorConfig.Properties)

	heading, err := v.reactorConfig.Properties["heading"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	renderedHeading, err := template.RenderTemplateValues(ctx, heading, fmt.Sprintf("%s_%s/heading", data.ID, v.reactorName), data.AsMap(), []string{}, templateConfig)
	if err != nil {
		return nil, err
	}

	body, err := v.reactorConfig.Properties["body"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	renderedBody, err := template.RenderTemplateValues(ctx, body, fmt.Sprintf("%s_%s/body", data.ID, v.reactorName), data.AsMap(), []string{}, templateConfig)
	if err != nil {
		return nil, err
	}
	if body == "" {
		return nil, fmt.Errorf("the body property was not supplied or was empty")
	}

	// ==========================================================
	// Get general reactor configuration
	// ==========================================================
	planTaskName, err := v.reactorConfig.Properties["planTaskName"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		planTaskName = ""
	}

	removeExistingCommentsFromAllPullRequestCommits := false
	removeExistingCommentsFromAllPullRequestCommitsStr, err := v.reactorConfig.Properties["removeExistingCommentsFromAllPullRequestCommits"].GetStringValue(ctx, v.Log, data)
	if err == nil && removeExistingCommentsFromAllPullRequestCommitsStr != "" {
		removeExistingCommentsFromAllPullRequestCommits, err = strconv.ParseBool(removeExistingCommentsFromAllPullRequestCommitsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied removeExistingCommentsFromAllPullRequestCommits '%v' to a boolean. Error: %v", removeExistingCommentsFromAllPullRequestCommitsStr, err)
		}
	}

	removeExistingPullRequestComments := true
	removeExistingPullRequestCommentsStr, err := v.reactorConfig.Properties["removeExistingPullRequestComments"].GetStringValue(ctx, v.Log, data)
	if err == nil && removeExistingPullRequestCommentsStr != "" {
		removeExistingPullRequestComments, err = strconv.ParseBool(removeExistingPullRequestCommentsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied removeExistingPullRequestComments '%v' to a boolean. Error: %v", removeExistingPullRequestCommentsStr, err)
		}
	}

	removeDuplicateCommitComments := true
	removeDuplicateCommitCommentsStr, err := v.reactorConfig.Properties["removeDuplicateCommitComments"].GetStringValue(ctx, v.Log, data)
	if err == nil && removeDuplicateCommitCommentsStr != "" {
		removeDuplicateCommitComments, err = strconv.ParseBool(removeDuplicateCommitCommentsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied removeDuplicateCommitComments '%v' to a boolean. Error: %v", removeDuplicateCommitCommentsStr, err)
		}
	}

	// ==========================================================
	// Get github configuration
	// ==========================================================
	token, err := v.reactorConfig.Properties["token"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("the github token property was not supplied or was empty")
	}

	org, err := v.reactorConfig.Properties["org"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if org == "" {
		return nil, fmt.Errorf("the github org property was not supplied or was empty")
	}

	repo, err := v.reactorConfig.Properties["repo"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if repo == "" {
		return nil, fmt.Errorf("the github repo property was not supplied or was empty")
	}

	commitSha, err := v.reactorConfig.Properties["commitSha"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if commitSha == "" {
		return nil, fmt.Errorf("the github commitSha property was not supplied or was empty")
	}

	prNumberStr, err := v.reactorConfig.Properties["prNumber"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if prNumberStr == "" {
		return nil, fmt.Errorf("the github prNumber property was not supplied or was empty")
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {

		return nil, fmt.Errorf("failed to convert the supplied pr number '%v' to an integer. Error: %v", prNumberStr, err)
	}

	isEnterprise := false
	enterpriseUrl, err := v.reactorConfig.Properties["enterpriseUrl"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		enterpriseUrl = github.DefaultBaseURL
	} else {
		isEnterprise = true
	}

	config := ReactorConfig{
		PlanTaskName: planTaskName,
		RemoveExistingCommentsFromAllPullRequestCommits: removeExistingCommentsFromAllPullRequestCommits,
		RemoveExistingPullRequestComments:               removeExistingPullRequestComments,
		RemoveDuplicateCommitComments:                   removeDuplicateCommitComments,
		GithubConfig:                                    github.New(ctx, log, org, repo, commitSha, token, prNumber, enterpriseUrl, isEnterprise),
		Heading:                                         string(renderedHeading),
		Body:                                            string(renderedBody),
	}
	return &config, nil
}

func (v *Reactor) GetHelp() string {
	return reactor.GetReactorHelp(v)
}

func (v *Reactor) GetProperties() []config.ReactorConfigProperty {
	return []config.ReactorConfigProperty{
		{
			Name:        "heading",
			Description: "The heading of the comment. This field supports go templating. The heading is also used to find previous comments to remove",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "body",
			Description: "The body of the comment. This field supports go templating",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "token",
			Description: "The github token to use for authentication. This token should have the necessary permissions to write comments to the repository",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "org",
			Description: "The github organization",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "repo",
			Description: "The github repository",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "commitSha",
			Description: "The commit sha where the comment will be written",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "prNumber",
			Description: "The pull request number where the comment will be written. If the value is less than 0, the comment will not be written to the pull request",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
	}
}

func (v *Reactor) GetRequiredPropertyNames() []string {
	return reactor.GetRequiredPropertyNames(v)
}
