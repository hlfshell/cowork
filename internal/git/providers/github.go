package gitprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/hlfshell/cowork/internal/git"
)

// GitHubProvider implements the GitProvider interface for GitHub
type GitHubProvider struct {
	client   *github.Client
	baseURL  string
	token    string
	username string
	password string
}

// NewGitHubProvider creates a new GitHub provider instance
func NewGitHubProvider(token, baseURL string) (*GitHubProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	client := github.NewClient(httpClient).WithAuthToken(token)

	// Set custom base URL for GitHub Enterprise
	if baseURL != "https://api.github.com" {
		parsedURL, err := url.Parse(strings.TrimSuffix(baseURL, "/") + "/")
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		client.BaseURL = parsedURL
	}

	return &GitHubProvider{
		client:  client,
		baseURL: baseURL,
		token:   token,
	}, nil
}

// GetProviderType returns the GitHub provider type
func (gp *GitHubProvider) GetProviderType() git.ProviderType {
	return git.ProviderGitHub
}

// TestAuth verifies GitHub authentication by making a minimal API call
func (gp *GitHubProvider) TestAuth(ctx context.Context) error {
	_, _, err := gp.client.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("GitHub authentication failed: %w", err)
	}
	return nil
}

// GetRepositoryInfo retrieves basic information about a GitHub repository
func (gp *GitHubProvider) GetRepositoryInfo(ctx context.Context, owner, repo string) (*git.Repository, error) {
	githubRepo, _, err := gp.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	return &git.Repository{
		Owner:         owner,
		Name:          repo,
		FullName:      githubRepo.GetFullName(),
		Description:   githubRepo.GetDescription(),
		URL:           githubRepo.GetURL(),
		Private:       githubRepo.GetPrivate(),
		DefaultBranch: githubRepo.GetDefaultBranch(),
		CreatedAt:     githubRepo.GetCreatedAt().Time,
		UpdatedAt:     githubRepo.GetUpdatedAt().Time,
	}, nil
}

// GetIssues retrieves issues from a GitHub repository with optional filtering
func (gp *GitHubProvider) GetIssues(ctx context.Context, owner, repo string, options *git.IssueListOptions) ([]*git.Issue, error) {
	githubOpts := &github.IssueListByRepoOptions{
		State:       "all",
		Sort:        "created",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// Apply filters
	if options != nil {
		if options.State != "" {
			githubOpts.State = options.State
		}
		if options.Assignee != "" {
			githubOpts.Assignee = options.Assignee
		}
		if options.Creator != "" {
			githubOpts.Creator = options.Creator
		}
		if options.Mentioned != "" {
			githubOpts.Mentioned = options.Mentioned
		}
		if options.Labels != "" {
			githubOpts.Labels = strings.Split(options.Labels, ",")
		}

		if !options.Since.IsZero() {
			githubOpts.Since = options.Since
		}
		if options.Sort != "" {
			githubOpts.Sort = options.Sort
		}
		if options.Direction != "" {
			githubOpts.Direction = options.Direction
		}
		if options.Page > 0 {
			githubOpts.Page = options.Page
		}
		if options.PerPage > 0 {
			githubOpts.PerPage = options.PerPage
		}
	}

	githubIssues, _, err := gp.client.Issues.ListByRepo(ctx, owner, repo, githubOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	issues := make([]*git.Issue, len(githubIssues))
	for i, githubIssue := range githubIssues {
		issues[i] = convertGitHubIssue(githubIssue)
	}

	return issues, nil
}

// GetIssue retrieves a specific GitHub issue by number
func (gp *GitHubProvider) GetIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.Issue, error) {
	githubIssue, _, err := gp.client.Issues.Get(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %d: %w", issueNumber, err)
	}

	return convertGitHubIssue(githubIssue), nil
}

// CreateIssue creates a new issue in a GitHub repository
func (gp *GitHubProvider) CreateIssue(ctx context.Context, owner, repo string, issue *git.CreateIssueRequest) (*git.Issue, error) {
	githubIssue := &github.IssueRequest{
		Title:     &issue.Title,
		Body:      &issue.Body,
		Assignees: &issue.Assignees,
		Labels:    &issue.Labels,
	}

	createdIssue, _, err := gp.client.Issues.Create(ctx, owner, repo, githubIssue)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return convertGitHubIssue(createdIssue), nil
}

// UpdateIssue updates an existing GitHub issue
func (gp *GitHubProvider) UpdateIssue(ctx context.Context, owner, repo string, issueNumber int, updates *git.UpdateIssueRequest) (*git.Issue, error) {
	githubIssue := &github.IssueRequest{}

	if updates.Title != nil {
		githubIssue.Title = updates.Title
	}
	if updates.Body != nil {
		githubIssue.Body = updates.Body
	}
	if updates.State != nil {
		githubIssue.State = updates.State
	}
	if updates.Assignees != nil {
		githubIssue.Assignees = updates.Assignees
	}
	if updates.Labels != nil {
		githubIssue.Labels = updates.Labels
	}

	updatedIssue, _, err := gp.client.Issues.Edit(ctx, owner, repo, issueNumber, githubIssue)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue %d: %w", issueNumber, err)
	}

	return convertGitHubIssue(updatedIssue), nil
}

// GetPullRequests retrieves pull requests from a GitHub repository
func (gp *GitHubProvider) GetPullRequests(ctx context.Context, owner, repo string, options *git.PullRequestListOptions) ([]*git.PullRequest, error) {
	githubOpts := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "created",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// Apply filters
	if options != nil {
		if options.State != "" {
			githubOpts.State = options.State
		}
		if options.Head != "" {
			githubOpts.Head = options.Head
		}
		if options.Base != "" {
			githubOpts.Base = options.Base
		}
		if options.Sort != "" {
			githubOpts.Sort = options.Sort
		}
		if options.Direction != "" {
			githubOpts.Direction = options.Direction
		}
		if options.Page > 0 {
			githubOpts.Page = options.Page
		}
		if options.PerPage > 0 {
			githubOpts.PerPage = options.PerPage
		}
	}

	githubPRs, _, err := gp.client.PullRequests.List(ctx, owner, repo, githubOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	pullRequests := make([]*git.PullRequest, len(githubPRs))
	for i, githubPR := range githubPRs {
		pullRequests[i] = convertGitHubPullRequest(githubPR)
	}

	return pullRequests, nil
}

// GetPullRequest retrieves a specific GitHub pull request by number
func (gp *GitHubProvider) GetPullRequest(ctx context.Context, owner, repo string, prNumber int) (*git.PullRequest, error) {
	githubPR, _, err := gp.client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request %d: %w", prNumber, err)
	}

	return convertGitHubPullRequest(githubPR), nil
}

// GetPullRequestByBranch retrieves a pull request by source branch name
func (gp *GitHubProvider) GetPullRequestByBranch(ctx context.Context, owner, repo string, branchName string) (*git.PullRequest, error) {
	// Search for PRs with the specified head branch
	githubOpts := &github.PullRequestListOptions{
		Head:        fmt.Sprintf("%s:%s", owner, branchName),
		ListOptions: github.ListOptions{PerPage: 1},
	}

	githubPRs, _, err := gp.client.PullRequests.List(ctx, owner, repo, githubOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to find pull request for branch %s: %w", branchName, err)
	}

	if len(githubPRs) == 0 {
		return nil, fmt.Errorf("no pull request found for branch %s", branchName)
	}

	return convertGitHubPullRequest(githubPRs[0]), nil
}

// GetPullRequestByIssue retrieves a pull request that closes a specific issue
func (gp *GitHubProvider) GetPullRequestByIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.PullRequest, error) {
	// First get the issue to see if it's a pull request
	githubIssue, _, err := gp.client.Issues.Get(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %d: %w", issueNumber, err)
	}

	// Check if the issue is actually a pull request
	if githubIssue.PullRequestLinks == nil {
		return nil, fmt.Errorf("issue %d is not a pull request", issueNumber)
	}

	// Get the pull request
	githubPR, _, err := gp.client.PullRequests.Get(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request for issue %d: %w", issueNumber, err)
	}

	return convertGitHubPullRequest(githubPR), nil
}

// CreatePullRequest creates a new pull request in a GitHub repository
func (gp *GitHubProvider) CreatePullRequest(ctx context.Context, owner, repo string, pr *git.CreatePullRequestRequest) (*git.PullRequest, error) {
	githubPR := &github.NewPullRequest{
		Title: &pr.Title,
		Body:  &pr.Body,
		Head:  &pr.Head,
		Base:  &pr.Base,
		Draft: &pr.Draft,
	}

	createdPR, _, err := gp.client.PullRequests.Create(ctx, owner, repo, githubPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return convertGitHubPullRequest(createdPR), nil
}

// UpdatePullRequest updates an existing GitHub pull request
func (gp *GitHubProvider) UpdatePullRequest(ctx context.Context, owner, repo string, prNumber int, updates *git.UpdatePullRequestRequest) (*git.PullRequest, error) {
	githubPR := &github.PullRequest{}

	if updates.Title != nil {
		githubPR.Title = updates.Title
	}
	if updates.Body != nil {
		githubPR.Body = updates.Body
	}
	if updates.State != nil {
		githubPR.State = updates.State
	}
	if updates.Base != nil {
		githubPR.Base = &github.PullRequestBranch{Ref: updates.Base}
	}
	if updates.Draft != nil {
		githubPR.Draft = updates.Draft
	}

	updatedPR, _, err := gp.client.PullRequests.Edit(ctx, owner, repo, prNumber, githubPR)
	if err != nil {
		return nil, fmt.Errorf("failed to update pull request %d: %w", prNumber, err)
	}

	return convertGitHubPullRequest(updatedPR), nil
}

// GetPullRequestReviews retrieves reviews for a GitHub pull request
func (gp *GitHubProvider) GetPullRequestReviews(ctx context.Context, owner, repo string, prNumber int) ([]*git.Review, error) {
	githubReviews, _, err := gp.client.PullRequests.ListReviews(ctx, owner, repo, prNumber, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request reviews: %w", err)
	}

	reviews := make([]*git.Review, len(githubReviews))
	for i, githubReview := range githubReviews {
		reviews[i] = convertGitHubReview(githubReview)
	}

	return reviews, nil
}

// GetPullRequestComments retrieves comments for a GitHub pull request
func (gp *GitHubProvider) GetPullRequestComments(ctx context.Context, owner, repo string, prNumber int) ([]*git.Comment, error) {
	githubComments, _, err := gp.client.PullRequests.ListComments(ctx, owner, repo, prNumber, &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request comments: %w", err)
	}

	comments := make([]*git.Comment, len(githubComments))
	for i, githubComment := range githubComments {
		comments[i] = convertGitHubComment(githubComment)
	}

	return comments, nil
}

// GetIssueComments retrieves comments for a GitHub issue
func (gp *GitHubProvider) GetIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*git.Comment, error) {
	githubComments, _, err := gp.client.Issues.ListComments(ctx, owner, repo, issueNumber, &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issue comments: %w", err)
	}

	comments := make([]*git.Comment, len(githubComments))
	for i, githubComment := range githubComments {
		comments[i] = convertGitHubIssueComment(githubComment)
	}

	return comments, nil
}

// CreateComment creates a comment on a GitHub issue or pull request
func (gp *GitHubProvider) CreateComment(ctx context.Context, owner, repo string, issueNumber int, comment *git.CreateCommentRequest) (*git.Comment, error) {
	githubComment := &github.IssueComment{
		Body: &comment.Body,
	}

	createdComment, _, err := gp.client.Issues.CreateComment(ctx, owner, repo, issueNumber, githubComment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return convertGitHubIssueComment(createdComment), nil
}

// GetLabels retrieves available labels for a GitHub repository
func (gp *GitHubProvider) GetLabels(ctx context.Context, owner, repo string) ([]*git.Label, error) {
	githubLabels, _, err := gp.client.Issues.ListLabels(ctx, owner, repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	labels := make([]*git.Label, len(githubLabels))
	for i, githubLabel := range githubLabels {
		labels[i] = convertGitHubLabel(githubLabel)
	}

	return labels, nil
}

// Helper functions to convert GitHub types to our generic types

func convertGitHubIssue(githubIssue *github.Issue) *git.Issue {
	issue := &git.Issue{
		Number:    githubIssue.GetNumber(),
		Title:     githubIssue.GetTitle(),
		Body:      githubIssue.GetBody(),
		State:     githubIssue.GetState(),
		Author:    convertGitHubUser(githubIssue.GetUser()),
		Assignees: convertGitHubUsers(githubIssue.Assignees),
		Labels:    convertGitHubLabels(githubIssue.Labels),
		CreatedAt: githubIssue.GetCreatedAt().Time,
		UpdatedAt: githubIssue.GetUpdatedAt().Time,
		Comments:  git.Comment{ID: githubIssue.GetComments()},
		URL:       githubIssue.GetURL(),
	}

	if githubIssue.ClosedAt != nil {
		issue.ClosedAt = &githubIssue.ClosedAt.Time
	}

	return issue
}

func convertGitHubPullRequest(githubPR *github.PullRequest) *git.PullRequest {
	pr := &git.PullRequest{
		Number:             githubPR.GetNumber(),
		Title:              githubPR.GetTitle(),
		Body:               githubPR.GetBody(),
		State:              githubPR.GetState(),
		Merged:             githubPR.GetMerged(),
		Author:             convertGitHubUser(githubPR.GetUser()),
		Assignees:          convertGitHubUsers(githubPR.Assignees),
		Labels:             convertGitHubLabels(githubPR.Labels),
		CreatedAt:          githubPR.GetCreatedAt().Time,
		UpdatedAt:          githubPR.GetUpdatedAt().Time,
		Comments:           git.Comment{ID: githubPR.GetComments()},
		URL:                githubPR.GetURL(),
		Draft:              githubPR.GetDraft(),
		RequestedReviewers: convertGitHubUsers(githubPR.RequestedReviewers),
	}

	if githubPR.ClosedAt != nil {
		pr.ClosedAt = &githubPR.ClosedAt.Time
	}

	if githubPR.MergedAt != nil {
		pr.MergedAt = &githubPR.MergedAt.Time
	}

	if githubPR.Head != nil {
		pr.Head = convertGitHubBranchInfo(githubPR.Head)
	}

	if githubPR.Base != nil {
		pr.Base = convertGitHubBranchInfo(githubPR.Base)
	}

	if githubPR.Mergeable != nil {
		pr.Mergeable = githubPR.Mergeable
	}

	return pr
}

func convertGitHubUser(githubUser *github.User) *git.User {
	if githubUser == nil {
		return nil
	}

	return &git.User{
		ID:    int(githubUser.GetID()),
		Login: githubUser.GetLogin(),
		Name:  githubUser.GetName(),
		Email: githubUser.GetEmail(),
		Type:  githubUser.GetType(),
	}
}

func convertGitHubUsers(githubUsers []*github.User) []*git.User {
	if githubUsers == nil {
		return nil
	}

	users := make([]*git.User, len(githubUsers))
	for i, githubUser := range githubUsers {
		users[i] = convertGitHubUser(githubUser)
	}

	return users
}

func convertGitHubLabel(githubLabel *github.Label) *git.Label {
	if githubLabel == nil {
		return nil
	}

	return &git.Label{
		ID:          int(githubLabel.GetID()),
		Name:        githubLabel.GetName(),
		Description: githubLabel.GetDescription(),
		URL:         githubLabel.GetURL(),
	}
}

func convertGitHubLabels(githubLabels []*github.Label) []*git.Label {
	if githubLabels == nil {
		return nil
	}

	labels := make([]*git.Label, len(githubLabels))
	for i, githubLabel := range githubLabels {
		labels[i] = convertGitHubLabel(githubLabel)
	}

	return labels
}

func convertGitHubReview(githubReview *github.PullRequestReview) *git.Review {
	if githubReview == nil {
		return nil
	}

	return &git.Review{
		ID:          int(githubReview.GetID()),
		User:        convertGitHubUser(githubReview.GetUser()),
		Body:        githubReview.GetBody(),
		State:       githubReview.GetState(),
		SubmittedAt: githubReview.GetSubmittedAt().Time,
		CommitID:    githubReview.GetCommitID(),
		URL:         "",
	}
}

func convertGitHubComment(githubComment *github.PullRequestComment) *git.Comment {
	if githubComment == nil {
		return nil
	}

	return &git.Comment{
		ID:        int(githubComment.GetID()),
		User:      convertGitHubUser(githubComment.GetUser()),
		Body:      githubComment.GetBody(),
		CreatedAt: githubComment.GetCreatedAt().Time,
		UpdatedAt: githubComment.GetUpdatedAt().Time,
		URL:       githubComment.GetURL(),
	}
}

func convertGitHubIssueComment(githubComment *github.IssueComment) *git.Comment {
	if githubComment == nil {
		return nil
	}

	return &git.Comment{
		ID:        int(githubComment.GetID()),
		User:      convertGitHubUser(githubComment.GetUser()),
		Body:      githubComment.GetBody(),
		CreatedAt: githubComment.GetCreatedAt().Time,
		UpdatedAt: githubComment.GetUpdatedAt().Time,
		URL:       githubComment.GetURL(),
	}
}

func convertGitHubBranchInfo(githubBranch *github.PullRequestBranch) *git.Branch {
	if githubBranch == nil {
		return nil
	}

	return &git.Branch{
		Ref:  githubBranch.GetRef(),
		SHA:  githubBranch.GetSHA(),
		Repo: convertGitHubRepositoryInfo(githubBranch.GetRepo()),
		User: convertGitHubUser(githubBranch.GetUser()),
	}
}

func convertGitHubRepositoryInfo(githubRepo *github.Repository) *git.Repository {
	if githubRepo == nil {
		return nil
	}

	return &git.Repository{
		Owner:         githubRepo.GetOwner().GetLogin(),
		Name:          githubRepo.GetName(),
		FullName:      githubRepo.GetFullName(),
		Description:   githubRepo.GetDescription(),
		URL:           githubRepo.GetURL(),
		Private:       githubRepo.GetPrivate(),
		DefaultBranch: githubRepo.GetDefaultBranch(),
		CreatedAt:     githubRepo.GetCreatedAt().Time,
		UpdatedAt:     githubRepo.GetUpdatedAt().Time,
	}
}
