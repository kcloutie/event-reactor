package github

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v57/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	DefaultBaseURL       = "https://api.github.com/"
	DefaultUploadBaseURL = "https://uploads.github.com/"
)

type GitHubConfiguration struct {
	context       context.Context
	logger        *zap.Logger
	accessToken   string
	Org           string
	Repo          string
	CommitSha     string
	EnterpriseUrl string
	PrNumber      int
	IsEnterprise  bool
}

func New(context context.Context, log *zap.Logger, org, repo, commitSha, accessToken string, prNumber int, enterpriseUrl string, isEnterprise bool) *GitHubConfiguration {
	logger := log.With(zap.String("reactor", "github/comment"), zap.String("org", org), zap.String("repo", repo), zap.String("commitSha", commitSha))
	return &GitHubConfiguration{
		context:       context,
		logger:        logger,
		accessToken:   accessToken,
		Org:           org,
		Repo:          repo,
		CommitSha:     commitSha,
		PrNumber:      prNumber,
		EnterpriseUrl: enterpriseUrl,
		IsEnterprise:  isEnterprise,
	}
}

func (c *GitHubConfiguration) WriteCommitComment(body string, commentHeading string, removeDuplicateCommitComment bool) (*github.RepositoryComment, error) {
	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	comment := &github.RepositoryComment{
		Body: &body,
	}

	if removeDuplicateCommitComment {
		c.logger.Info("RemoveDuplicateCommitComments was true, removing existing comments on the commit")
		r, _ := regexp.Compile(fmt.Sprintf("^%v", strings.Trim(commentHeading, " ")))
		err = c.removeCommitComments(client, r, c.CommitSha)
		if err != nil {
			c.logger.Error(fmt.Sprintf("failed to cleanup existing comments on commit '%v'", c.CommitSha), zap.Error(err))
		}
	} else {
		c.logger.Info("RemoveDuplicateCommitComments was false, skipping the removal of existing comments on the commit")
	}

	newComment, resp, err := client.Repositories.CreateComment(c.context, c.Org, c.Repo, c.CommitSha, comment)
	body, err = c.checkHttpResponse(newComment, resp, err)
	return newComment, err

}

func (c *GitHubConfiguration) NewClient() (*github.Client, error) {
	var client *github.Client
	var httpClient *http.Client
	var err error

	if c.accessToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: c.accessToken},
		)
		httpClient = oauth2.NewClient(c.context, ts)
	}
	client = github.NewClient(httpClient)

	if c.IsEnterprise {
		client, err = client.WithEnterpriseURLs(c.EnterpriseUrl, c.EnterpriseUrl)
		if err != nil {
			return nil, err
		}
	}

	return client, err
}

func (c *GitHubConfiguration) WritePullRequestComment(body string) (*github.IssueComment, error) {
	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	comment := &github.IssueComment{
		Body: &body,
	}

	newComment, resp, err := client.Issues.CreateComment(c.context, c.Org, c.Repo, c.PrNumber, comment)
	body, err = c.checkHttpResponse(newComment, resp, err)

	return newComment, err

}

func (c *GitHubConfiguration) checkHttpResponse(_ interface{}, resp *github.Response, err error) (string, error) {
	respBodyString := ""
	if err != nil {
		if resp != nil && resp.Body != nil {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			respBodyString = buf.String()
		} else {
			respBodyString = err.Error()
		}

		return respBodyString, fmt.Errorf("github api error. Response Body: %v", respBodyString)
	}

	if resp.StatusCode <= 199 || resp.StatusCode >= 400 {
		return respBodyString, fmt.Errorf("github API call failed. Status Code: %v. Response Body: %v", resp.StatusCode, respBodyString)
	}

	return respBodyString, nil
}

func (c *GitHubConfiguration) CleanExistingCommentsOnAllPullRequestCommits(commentHeading string) error {
	client, err := c.NewClient()
	if err != nil {
		return err
	}

	r, _ := regexp.Compile(fmt.Sprintf("^%v", strings.Trim(commentHeading, " ")))

	if c.PrNumber == -1 || c.PrNumber == 0 {
		comments, _, err := client.Repositories.ListCommitComments(c.context, c.Org, c.Repo, c.CommitSha, nil)

		if err != nil {
			return fmt.Errorf("an error occurred attempting to list the commits on comment '%v'. Error: %v", c.CommitSha, err)
		}
		for _, comment := range comments {
			c.logger.Info("found commit comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL))
			c.removeCommitComment(*comment.Body, *comment.ID, r, client)
		}

	} else {
		commits, _, err := client.PullRequests.ListCommits(c.context, c.Org, c.Repo, c.PrNumber, nil)

		if err != nil {
			return fmt.Errorf("an error occurred attempting to list commits from pull request '%v'. Error: %v", c.PrNumber, err)
		}
		for _, commit := range commits {
			sha := *commit.SHA
			err := c.removeCommitComments(client, r, sha)
			if err != nil {
				return fmt.Errorf("an error occurred attempting to remove comments on commit '%v'. Error: %v", sha, err)
			}
		}
	}
	return nil
}

func (c *GitHubConfiguration) CleanExistingCommentsOnPullRequest(commentHeading string) error {
	client, err := c.NewClient()
	if err != nil {
		return err
	}

	r, _ := regexp.Compile(fmt.Sprintf("^%v", strings.Trim(commentHeading, " ")))

	comments, _, _ := client.Issues.ListComments(c.context, c.Org, c.Repo, c.PrNumber, nil)

	if err != nil {
		return fmt.Errorf("an error occurred attempting to list the commits on pr '%v'. Error: %v", c.PrNumber, err)
	}
	for _, comment := range comments {
		c.logger.Info("found pull request comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL))
		body := *comment.Body
		if r.MatchString(body) {
			_, err := client.Issues.DeleteComment(c.context, c.Org, c.Repo, *comment.ID)

			if err != nil {
				c.logger.Error("failed to remove github pull request comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL), zap.Error(err))
			} else {
				c.logger.Info("removed github pull request comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL))
			}
		}
	}
	return nil
}

func (c *GitHubConfiguration) CleanExistingCommentsOnCommit(commentHeading string) error {
	client, err := c.NewClient()
	if err != nil {
		return err
	}

	r, _ := regexp.Compile(fmt.Sprintf("^%v", strings.Trim(commentHeading, " ")))

	err = c.removeCommitComments(client, r, c.CommitSha)
	if err != nil {
		return fmt.Errorf("an error occurred attempting to remove comments on commit '%v'. Error: %v", c.CommitSha, err)
	}

	return nil
}

func (c *GitHubConfiguration) removeCommitComments(client *github.Client, r *regexp.Regexp, sha string) error {
	comments, _, err := client.Repositories.ListCommitComments(c.context, c.Org, c.Repo, sha, nil)

	if err != nil {
		return fmt.Errorf("an error occurred attempting to list the comments on commit '%v'. Error: %v", c.CommitSha, err)
	}
	for _, comment := range comments {
		c.logger.Info("found commit comment", zap.String("commentId", fmt.Sprintf("%v", comment.ID)), zap.String("commentHtmlUrl", *comment.HTMLURL))
		c.removeCommitComment(*comment.Body, *comment.ID, r, client)
	}
	return nil
}

func (c *GitHubConfiguration) removeCommitComment(body string, commentId int64, r *regexp.Regexp, client *github.Client) {
	// body := *comment.Body
	if r.MatchString(body) {
		resp, err := client.Repositories.DeleteComment(c.context, c.Org, c.Repo, commentId)
		fmt.Println(resp)
		if err != nil {
			c.logger.Error("failed to remove github commit comment", zap.String("commentId", fmt.Sprintf("%v", commentId)), zap.Error(err))
		} else {
			c.logger.Info("removed github commit comment", zap.String("commentId", fmt.Sprintf("%v", commentId)))
		}
	}
}

func (c *GitHubConfiguration) WriteCommitCheckStatus(state string, context string, targetUrl string, description string) (*github.RepoStatus, error) {
	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	commitStatus := &github.RepoStatus{
		State:       &state,
		Context:     &context,
		TargetURL:   &targetUrl,
		Description: &description,
	}

	newCommitStatus, resp, err := client.Repositories.CreateStatus(c.context, c.Org, c.Repo, c.CommitSha, commitStatus)

	_, err = c.checkHttpResponse(newCommitStatus, resp, err)
	return newCommitStatus, err

}

func (c *GitHubConfiguration) GetCommitCheckStatus(state string, context string, targetUrl string, description string) ([]*github.RepoStatus, error) {
	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	statuses, resp, err := client.Repositories.ListStatuses(c.context, c.Org, c.Repo, c.CommitSha, &github.ListOptions{})
	_, err = c.checkHttpResponse(statuses, resp, err)
	return statuses, err

}

func (c *GitHubConfiguration) WriteCommitCheckStatusIfFailedDoesNotExist(state string, context string, targetUrl string, description string) (*github.RepoStatus, error) {
	status, err := c.GetCommitCheckStatus(state, context, targetUrl, description)

	if err != nil {
		return nil, err
	}
	if state == "success" {
		for _, val := range status {
			state := *val.State
			if *val.Context == context && (state == "failure" || state == "error") {
				return nil, nil
			}
		}
	}

	return c.WriteCommitCheckStatus(state, context, targetUrl, description)
}

func (c *GitHubConfiguration) CreateDeploymentAndStatus(environment string, envIsProd bool, autoMerge bool, description string, requiredContexts []string, state string, autoInactive bool, logsUrl *string, environmentUrl *string) (*github.Deployment, *github.DeploymentStatus, error) {

	deployment, err := c.CreateDeployment(environment, envIsProd, autoMerge, description, requiredContexts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create the deployment object for environment '%v' in the '%v/%v' repository on commit %v. - %v", environment, c.Org, c.Repo, c.CommitSha, err)
	}

	deploymentStatus, err := c.CreateDeploymentStatus(*deployment.ID, environment, state, autoInactive, description, logsUrl, environmentUrl)
	if err != nil {
		deploymentIdString := strconv.FormatInt(*deployment.ID, 10)
		return deployment, nil, fmt.Errorf("failed to create the deployment status on deployment ID %v for environment '%v/%v' in the '%v' repository on commit %v. - %v", deploymentIdString, environment, c.Org, c.Repo, c.CommitSha, err)
	}

	return deployment, deploymentStatus, nil

}

func (c *GitHubConfiguration) CreateDeployment(environment string, envIsProd bool, autoMerge bool, description string, requiredContexts []string) (*github.Deployment, error) {
	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	isProd := false
	isTrans := true

	if envIsProd {
		isProd = true
		isTrans = false
	}

	req := &github.DeploymentRequest{
		Ref:                   &c.CommitSha,
		RequiredContexts:      &requiredContexts,
		Environment:           &environment,
		Description:           &description,
		TransientEnvironment:  &isTrans,
		ProductionEnvironment: &isProd,
		AutoMerge:             &autoMerge,
	}

	deployment, resp, err := client.Repositories.CreateDeployment(c.context, c.Org, c.Repo, req)

	_, err = c.checkHttpResponse(deployment, resp, err)
	return deployment, err

}

func (c *GitHubConfiguration) CreateDeploymentStatus(deploymentId int64, environment string, state string, autoInactive bool, description string, logsUrl *string, environmentUrl *string) (*github.DeploymentStatus, error) {
	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	req := &github.DeploymentStatusRequest{
		State:          &state,
		Environment:    &environment,
		Description:    &description,
		AutoInactive:   &autoInactive,
		LogURL:         logsUrl,
		EnvironmentURL: environmentUrl,
	}

	deployment, resp, err := client.Repositories.CreateDeploymentStatus(c.context, c.Org, c.Repo, deploymentId, req)

	_, err = c.checkHttpResponse(deployment, resp, err)
	return deployment, err

}
