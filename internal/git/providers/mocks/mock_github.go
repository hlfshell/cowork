package gitprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/hlfshell/cowork/internal/git"
)

// MockGitHubProvider implements the GitProvider interface with GitHub-specific behavior
// This mock provides realistic responses based on GitHub API documentation
type MockGitHubProvider struct {
	shouldFail  bool
	failMethod  string
	rateLimited bool
}

// NewMockGitHubProvider creates a new GitHub mock provider for testing
func NewMockGitHubProvider() *MockGitHubProvider {
	return &MockGitHubProvider{
		shouldFail:  false,
		rateLimited: false,
	}
}

// NewMockGitHubProviderWithFailure creates a GitHub mock provider that fails on specific methods
func NewMockGitHubProviderWithFailure(failMethod string) *MockGitHubProvider {
	return &MockGitHubProvider{
		shouldFail:  true,
		failMethod:  failMethod,
		rateLimited: false,
	}
}

// NewMockGitHubProviderWithRateLimit creates a GitHub mock provider that simulates rate limiting
func NewMockGitHubProviderWithRateLimit() *MockGitHubProvider {
	return &MockGitHubProvider{
		shouldFail:  false,
		rateLimited: true,
	}
}

// GetProviderType returns the type of this provider
func (mgp *MockGitHubProvider) GetProviderType() git.ProviderType {
	return git.ProviderGitHub
}

// TestAuth verifies if the provided authentication is valid
func (mgp *MockGitHubProvider) TestAuth(ctx context.Context) error {
	if mgp.shouldFail && mgp.failMethod == "TestAuth" {
		return fmt.Errorf("GitHub authentication failed: 401 Unauthorized")
	}
	if mgp.rateLimited {
		return fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}
	return nil
}

// GetRepositoryInfo retrieves basic information about a GitHub repository
func (mgp *MockGitHubProvider) GetRepositoryInfo(ctx context.Context, owner, repo string) (*git.Repository, error) {
	if mgp.shouldFail && mgp.failMethod == "GetRepositoryInfo" {
		return nil, fmt.Errorf("GitHub repository not found: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	// GitHub-specific repository data structure
	return &git.Repository{
		Owner:         owner,
		Name:          repo,
		FullName:      fmt.Sprintf("%s/%s", owner, repo),
		Description:   "A mock GitHub repository for testing purposes",
		URL:           fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo),
		Private:       false,
		DefaultBranch: "main", // GitHub's default branch is typically "main"
		CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

// GetIssues retrieves issues from a GitHub repository with optional filtering
func (mgp *MockGitHubProvider) GetIssues(ctx context.Context, owner, repo string, options *git.IssueListOptions) ([]*git.Issue, error) {
	if mgp.shouldFail && mgp.failMethod == "GetIssues" {
		return nil, fmt.Errorf("GitHub issues retrieval failed: 500 Internal Server Error")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	// GitHub-specific issues data structure
	issues := []*git.Issue{
		{
			Number:    1,
			Title:     "GitHub Mock Issue 1",
			Body:      "This is a mock GitHub issue for testing\n\n## Description\nThis issue demonstrates GitHub's markdown support.",
			State:     "open",
			Author:    mgp.mockGitHubUser(1, "github-user1"),
			Assignees: []*git.User{mgp.mockGitHubUser(2, "github-assignee1")},
			Labels:    []*git.Label{mgp.mockGitHubLabel(1, "bug"), mgp.mockGitHubLabel(2, "enhancement")},
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Now().UTC(),
			Comments:  git.Comment{ID: 5},
			URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/1", owner, repo),
		},
		{
			Number:    2,
			Title:     "GitHub Mock Issue 2",
			Body:      "Another mock GitHub issue for testing",
			State:     "closed", // GitHub uses "closed" for closed issues
			Author:    mgp.mockGitHubUser(3, "github-user2"),
			Assignees: []*git.User{},
			Labels:    []*git.Label{mgp.mockGitHubLabel(3, "documentation")},
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			ClosedAt:  &time.Time{},
			Comments:  git.Comment{ID: 2},
			URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/2", owner, repo),
		},
	}

	return issues, nil
}

// GetIssue retrieves a specific GitHub issue by ID
func (mgp *MockGitHubProvider) GetIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.Issue, error) {
	if mgp.shouldFail && mgp.failMethod == "GetIssue" {
		return nil, fmt.Errorf("GitHub issue not found: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	return &git.Issue{
		Number:    issueNumber,
		Title:     fmt.Sprintf("GitHub Mock Issue %d", issueNumber),
		Body:      fmt.Sprintf("This is mock GitHub issue %d for testing", issueNumber),
		State:     "open",
		Author:    mgp.mockGitHubUser(1, "github-user"),
		Assignees: []*git.User{mgp.mockGitHubUser(2, "github-assignee")},
		Labels:    []*git.Label{mgp.mockGitHubLabel(1, "bug")},
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 3},
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, issueNumber),
	}, nil
}

// CreateIssue creates a new GitHub issue
func (mgp *MockGitHubProvider) CreateIssue(ctx context.Context, owner, repo string, issue *git.CreateIssueRequest) (*git.Issue, error) {
	if mgp.shouldFail && mgp.failMethod == "CreateIssue" {
		return nil, fmt.Errorf("GitHub issue creation failed: 422 Unprocessable Entity")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	return &git.Issue{
		Number:    999,
		Title:     issue.Title,
		Body:      issue.Body,
		State:     "open",
		Author:    mgp.mockGitHubUser(1, "github-currentuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/999", owner, repo),
	}, nil
}

// UpdateIssue updates an existing GitHub issue
func (mgp *MockGitHubProvider) UpdateIssue(ctx context.Context, owner, repo string, issueNumber int, updates *git.UpdateIssueRequest) (*git.Issue, error) {
	if mgp.shouldFail && mgp.failMethod == "UpdateIssue" {
		return nil, fmt.Errorf("GitHub issue update failed: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	title := fmt.Sprintf("Updated GitHub Mock Issue %d", issueNumber)
	if updates.Title != nil {
		title = *updates.Title
	}

	return &git.Issue{
		Number:    issueNumber,
		Title:     title,
		Body:      "Updated GitHub mock issue body",
		State:     "open",
		Author:    mgp.mockGitHubUser(1, "github-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, issueNumber),
	}, nil
}

// GetPullRequests retrieves pull requests from a GitHub repository
func (mgp *MockGitHubProvider) GetPullRequests(ctx context.Context, owner, repo string, options *git.PullRequestListOptions) ([]*git.PullRequest, error) {
	if mgp.shouldFail && mgp.failMethod == "GetPullRequests" {
		return nil, fmt.Errorf("GitHub pull requests retrieval failed: 500 Internal Server Error")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	// GitHub-specific pull requests data structure
	prs := []*git.PullRequest{
		{
			Number:    1,
			Title:     "GitHub Mock Pull Request 1",
			Body:      "This is a mock GitHub pull request for testing\n\n## Changes\n- Added new feature\n- Updated documentation",
			State:     "open",
			Merged:    false,
			Author:    mgp.mockGitHubUser(1, "github-user1"),
			Assignees: []*git.User{mgp.mockGitHubUser(2, "github-assignee1")},
			Labels:    []*git.Label{mgp.mockGitHubLabel(1, "feature")},
			Head:      mgp.mockGitHubBranch("feature-branch", "abc123"),
			Base:      mgp.mockGitHubBranch("main", "def456"),
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Now().UTC(),
			Comments:  git.Comment{ID: 3},
			URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/1", owner, repo),
			Draft:     false,
			Mergeable: &[]bool{true}[0],
		},
		{
			Number:    2,
			Title:     "GitHub Mock Pull Request 2",
			Body:      "Another mock GitHub pull request for testing",
			State:     "closed", // GitHub uses "closed" for merged/closed PRs
			Merged:    true,
			MergedAt:  &time.Time{},
			Author:    mgp.mockGitHubUser(3, "github-user2"),
			Assignees: []*git.User{},
			Labels:    []*git.Label{mgp.mockGitHubLabel(2, "bugfix")},
			Head:      mgp.mockGitHubBranch("bugfix-branch", "ghi789"),
			Base:      mgp.mockGitHubBranch("main", "def456"),
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			ClosedAt:  &time.Time{},
			Comments:  git.Comment{ID: 1},
			URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/2", owner, repo),
			Draft:     false,
			Mergeable: &[]bool{true}[0],
		},
	}

	return prs, nil
}

// GetPullRequest retrieves a specific GitHub pull request by number
func (mgp *MockGitHubProvider) GetPullRequest(ctx context.Context, owner, repo string, prNumber int) (*git.PullRequest, error) {
	if mgp.shouldFail && mgp.failMethod == "GetPullRequest" {
		return nil, fmt.Errorf("GitHub pull request not found: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	return &git.PullRequest{
		Number:    prNumber,
		Title:     fmt.Sprintf("GitHub Mock Pull Request %d", prNumber),
		Body:      fmt.Sprintf("This is mock GitHub pull request %d for testing", prNumber),
		State:     "open",
		Merged:    false,
		Author:    mgp.mockGitHubUser(1, "github-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{mgp.mockGitHubLabel(1, "feature")},
		Head:      mgp.mockGitHubBranch("feature-branch", "abc123"),
		Base:      mgp.mockGitHubBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 2},
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestByBranch retrieves a GitHub pull request by source branch name
func (mgp *MockGitHubProvider) GetPullRequestByBranch(ctx context.Context, owner, repo string, branchName string) (*git.PullRequest, error) {
	if mgp.shouldFail && mgp.failMethod == "GetPullRequestByBranch" {
		return nil, fmt.Errorf("GitHub pull request not found for branch: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	return &git.PullRequest{
		Number:    123,
		Title:     fmt.Sprintf("GitHub Mock PR for branch %s", branchName),
		Body:      fmt.Sprintf("This is a mock GitHub pull request for branch %s", branchName),
		State:     "open",
		Merged:    false,
		Author:    mgp.mockGitHubUser(1, "github-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mgp.mockGitHubBranch(branchName, "abc123"),
		Base:      mgp.mockGitHubBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/123", owner, repo),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestByIssue retrieves a GitHub pull request that closes a specific issue
func (mgp *MockGitHubProvider) GetPullRequestByIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.PullRequest, error) {
	if mgp.shouldFail && mgp.failMethod == "GetPullRequestByIssue" {
		return nil, fmt.Errorf("GitHub pull request not found for issue: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	return &git.PullRequest{
		Number:    456,
		Title:     fmt.Sprintf("GitHub Mock PR closing issue %d", issueNumber),
		Body:      fmt.Sprintf("This GitHub pull request closes issue #%d", issueNumber),
		State:     "open",
		Merged:    false,
		Author:    mgp.mockGitHubUser(1, "github-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mgp.mockGitHubBranch("fix-issue", "abc123"),
		Base:      mgp.mockGitHubBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/456", owner, repo),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// CreatePullRequest creates a new GitHub pull request
func (mgp *MockGitHubProvider) CreatePullRequest(ctx context.Context, owner, repo string, pr *git.CreatePullRequestRequest) (*git.PullRequest, error) {
	if mgp.shouldFail && mgp.failMethod == "CreatePullRequest" {
		return nil, fmt.Errorf("GitHub pull request creation failed: 422 Unprocessable Entity")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	return &git.PullRequest{
		Number:    999,
		Title:     pr.Title,
		Body:      pr.Body,
		State:     "open",
		Merged:    false,
		Author:    mgp.mockGitHubUser(1, "github-currentuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mgp.mockGitHubBranch(pr.Head, "abc123"),
		Base:      mgp.mockGitHubBranch(pr.Base, "def456"),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/999", owner, repo),
		Draft:     pr.Draft,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// UpdatePullRequest updates an existing GitHub pull request
func (mgp *MockGitHubProvider) UpdatePullRequest(ctx context.Context, owner, repo string, prNumber int, updates *git.UpdatePullRequestRequest) (*git.PullRequest, error) {
	if mgp.shouldFail && mgp.failMethod == "UpdatePullRequest" {
		return nil, fmt.Errorf("GitHub pull request update failed: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	title := fmt.Sprintf("Updated GitHub Mock PR %d", prNumber)
	if updates.Title != nil {
		title = *updates.Title
	}

	return &git.PullRequest{
		Number:    prNumber,
		Title:     title,
		Body:      "Updated GitHub mock pull request body",
		State:     "open",
		Merged:    false,
		Author:    mgp.mockGitHubUser(1, "github-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mgp.mockGitHubBranch("feature-branch", "abc123"),
		Base:      mgp.mockGitHubBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestReviews retrieves reviews for a GitHub pull request
func (mgp *MockGitHubProvider) GetPullRequestReviews(ctx context.Context, owner, repo string, prNumber int) ([]*git.Review, error) {
	if mgp.shouldFail && mgp.failMethod == "GetPullRequestReviews" {
		return nil, fmt.Errorf("GitHub pull request reviews retrieval failed: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	reviews := []*git.Review{
		{
			ID:          1,
			User:        mgp.mockGitHubUser(1, "github-reviewer1"),
			Body:        "This looks good! 👍",
			State:       "approved",
			SubmittedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			CommitID:    "abc123",
			URL:         fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews/1", owner, repo, prNumber),
		},
		{
			ID:          2,
			User:        mgp.mockGitHubUser(2, "github-reviewer2"),
			Body:        "Please fix the formatting",
			State:       "changes_requested",
			SubmittedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			CommitID:    "def456",
			URL:         fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/reviews/2", owner, repo, prNumber),
		},
	}

	return reviews, nil
}

// GetPullRequestComments retrieves comments for a GitHub pull request
func (mgp *MockGitHubProvider) GetPullRequestComments(ctx context.Context, owner, repo string, prNumber int) ([]*git.Comment, error) {
	if mgp.shouldFail && mgp.failMethod == "GetPullRequestComments" {
		return nil, fmt.Errorf("GitHub pull request comments retrieval failed: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	comments := []*git.Comment{
		{
			ID:        1,
			User:      mgp.mockGitHubUser(1, "github-commenter1"),
			Body:      "Great work on this PR! 🎉",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/comments/1", owner, repo, prNumber),
		},
		{
			ID:        2,
			User:      mgp.mockGitHubUser(2, "github-commenter2"),
			Body:      "Can you add more tests?",
			CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/comments/2", owner, repo, prNumber),
		},
	}

	return comments, nil
}

// GetIssueComments retrieves comments for a GitHub issue
func (mgp *MockGitHubProvider) GetIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*git.Comment, error) {
	if mgp.shouldFail && mgp.failMethod == "GetIssueComments" {
		return nil, fmt.Errorf("GitHub issue comments retrieval failed: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	comments := []*git.Comment{
		{
			ID:        1,
			User:      mgp.mockGitHubUser(1, "github-commenter1"),
			Body:      "This issue is important",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments/1", owner, repo, issueNumber),
		},
		{
			ID:        2,
			User:      mgp.mockGitHubUser(2, "github-commenter2"),
			Body:      "I can help with this",
			CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments/2", owner, repo, issueNumber),
		},
	}

	return comments, nil
}

// CreateComment creates a comment on a GitHub issue or pull request
func (mgp *MockGitHubProvider) CreateComment(ctx context.Context, owner, repo string, issueNumber int, comment *git.CreateCommentRequest) (*git.Comment, error) {
	if mgp.shouldFail && mgp.failMethod == "CreateComment" {
		return nil, fmt.Errorf("GitHub comment creation failed: 422 Unprocessable Entity")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	return &git.Comment{
		ID:        999,
		User:      mgp.mockGitHubUser(1, "github-currentuser"),
		Body:      comment.Body,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		URL:       fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments/999", owner, repo, issueNumber),
	}, nil
}

// GetLabels retrieves available labels for a GitHub repository
func (mgp *MockGitHubProvider) GetLabels(ctx context.Context, owner, repo string) ([]*git.Label, error) {
	if mgp.shouldFail && mgp.failMethod == "GetLabels" {
		return nil, fmt.Errorf("GitHub labels retrieval failed: 404 Not Found")
	}
	if mgp.rateLimited {
		return nil, fmt.Errorf("GitHub API rate limit exceeded: 403 Forbidden")
	}

	// GitHub-specific labels with their typical descriptions
	labels := []*git.Label{
		{
			ID:          1,
			Name:        "bug",
			Description: "Something isn't working",
			URL:         fmt.Sprintf("https://api.github.com/repos/%s/%s/labels/bug", owner, repo),
		},
		{
			ID:          2,
			Name:        "enhancement",
			Description: "New feature or request",
			URL:         fmt.Sprintf("https://api.github.com/repos/%s/%s/labels/enhancement", owner, repo),
		},
		{
			ID:          3,
			Name:        "documentation",
			Description: "Improvements or additions to documentation",
			URL:         fmt.Sprintf("https://api.github.com/repos/%s/%s/labels/documentation", owner, repo),
		},
		{
			ID:          4,
			Name:        "good first issue",
			Description: "Good for newcomers",
			URL:         fmt.Sprintf("https://api.github.com/repos/%s/%s/labels/good%%20first%%20issue", owner, repo),
		},
		{
			ID:          5,
			Name:        "help wanted",
			Description: "Extra attention is needed",
			URL:         fmt.Sprintf("https://api.github.com/repos/%s/%s/labels/help%%20wanted", owner, repo),
		},
	}

	return labels, nil
}

// Helper methods to create GitHub-specific mock data

func (mgp *MockGitHubProvider) mockGitHubUser(id int, login string) *git.User {
	return &git.User{
		ID:    id,
		Login: login,
		Name:  fmt.Sprintf("GitHub User %d", id),
		Email: fmt.Sprintf("%s@github.com", login),
		Type:  "User",
	}
}

func (mgp *MockGitHubProvider) mockGitHubLabel(id int, name string) *git.Label {
	return &git.Label{
		ID:          id,
		Name:        name,
		Description: fmt.Sprintf("GitHub mock label: %s", name),
		URL:         fmt.Sprintf("https://api.github.com/repos/test/owner/labels/%s", name),
	}
}

func (mgp *MockGitHubProvider) mockGitHubBranch(ref, sha string) *git.Branch {
	return &git.Branch{
		Ref:  ref,
		SHA:  sha,
		Repo: &git.Repository{Owner: "test", Name: "repo"},
		User: mgp.mockGitHubUser(1, "github-user"),
	}
}
