package gitprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/hlfshell/cowork/internal/git"
)

// MockGitLabProvider implements the GitProvider interface with GitLab-specific behavior
// This mock provides realistic responses based on GitLab API documentation
type MockGitLabProvider struct {
	shouldFail  bool
	failMethod  string
	rateLimited bool
}

// NewMockGitLabProvider creates a new GitLab mock provider for testing
func NewMockGitLabProvider() *MockGitLabProvider {
	return &MockGitLabProvider{
		shouldFail:  false,
		rateLimited: false,
	}
}

// NewMockGitLabProviderWithFailure creates a GitLab mock provider that fails on specific methods
func NewMockGitLabProviderWithFailure(failMethod string) *MockGitLabProvider {
	return &MockGitLabProvider{
		shouldFail:  true,
		failMethod:  failMethod,
		rateLimited: false,
	}
}

// NewMockGitLabProviderWithRateLimit creates a GitLab mock provider that simulates rate limiting
func NewMockGitLabProviderWithRateLimit() *MockGitLabProvider {
	return &MockGitLabProvider{
		shouldFail:  false,
		rateLimited: true,
	}
}

// GetProviderType returns the type of this provider
func (mlp *MockGitLabProvider) GetProviderType() git.ProviderType {
	return git.ProviderGitLab
}

// TestAuth verifies if the provided authentication is valid
func (mlp *MockGitLabProvider) TestAuth(ctx context.Context) error {
	if mlp.shouldFail && mlp.failMethod == "TestAuth" {
		return fmt.Errorf("GitLab authentication failed: 401 Unauthorized")
	}
	if mlp.rateLimited {
		return fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}
	return nil
}

// GetRepositoryInfo retrieves basic information about a GitLab repository
func (mlp *MockGitLabProvider) GetRepositoryInfo(ctx context.Context, owner, repo string) (*git.Repository, error) {
	if mlp.shouldFail && mlp.failMethod == "GetRepositoryInfo" {
		return nil, fmt.Errorf("GitLab project not found: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	// GitLab-specific repository data structure (GitLab calls them "projects")
	return &git.Repository{
		Owner:         owner,
		Name:          repo,
		FullName:      fmt.Sprintf("%s/%s", owner, repo),
		Description:   "A mock GitLab project for testing purposes",
		URL:           fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s", owner, repo),
		Private:       false,
		DefaultBranch: "master", // GitLab's default branch is typically "master"
		CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

// GetIssues retrieves issues from a GitLab repository with optional filtering
func (mlp *MockGitLabProvider) GetIssues(ctx context.Context, owner, repo string, options *git.IssueListOptions) ([]*git.Issue, error) {
	if mlp.shouldFail && mlp.failMethod == "GetIssues" {
		return nil, fmt.Errorf("GitLab issues retrieval failed: 500 Internal Server Error")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	// GitLab-specific issues data structure
	issues := []*git.Issue{
		{
			Number:    1,
			Title:     "GitLab Mock Issue 1",
			Body:      "This is a mock GitLab issue for testing\n\n## Description\nThis issue demonstrates GitLab's markdown support.",
			State:     "opened", // GitLab uses "opened" instead of "open"
			Author:    mlp.mockGitLabUser(1, "gitlab-user1"),
			Assignees: []*git.User{mlp.mockGitLabUser(2, "gitlab-assignee1")},
			Labels:    []*git.Label{mlp.mockGitLabLabel(1, "bug"), mlp.mockGitLabLabel(2, "enhancement")},
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Now().UTC(),
			Comments:  git.Comment{ID: 5},
			URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/issues/1", owner, repo),
		},
		{
			Number:    2,
			Title:     "GitLab Mock Issue 2",
			Body:      "Another mock GitLab issue for testing",
			State:     "closed", // GitLab uses "closed" for closed issues
			Author:    mlp.mockGitLabUser(3, "gitlab-user2"),
			Assignees: []*git.User{},
			Labels:    []*git.Label{mlp.mockGitLabLabel(3, "documentation")},
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			ClosedAt:  &time.Time{},
			Comments:  git.Comment{ID: 2},
			URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/issues/2", owner, repo),
		},
	}

	return issues, nil
}

// GetIssue retrieves a specific GitLab issue by ID
func (mlp *MockGitLabProvider) GetIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.Issue, error) {
	if mlp.shouldFail && mlp.failMethod == "GetIssue" {
		return nil, fmt.Errorf("GitLab issue not found: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.Issue{
		Number:    issueNumber,
		Title:     fmt.Sprintf("GitLab Mock Issue %d", issueNumber),
		Body:      fmt.Sprintf("This is mock GitLab issue %d for testing", issueNumber),
		State:     "opened", // GitLab uses "opened" instead of "open"
		Author:    mlp.mockGitLabUser(1, "gitlab-user"),
		Assignees: []*git.User{mlp.mockGitLabUser(2, "gitlab-assignee")},
		Labels:    []*git.Label{mlp.mockGitLabLabel(1, "bug")},
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 3},
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/issues/%d", owner, repo, issueNumber),
	}, nil
}

// CreateIssue creates a new GitLab issue
func (mlp *MockGitLabProvider) CreateIssue(ctx context.Context, owner, repo string, issue *git.CreateIssueRequest) (*git.Issue, error) {
	if mlp.shouldFail && mlp.failMethod == "CreateIssue" {
		return nil, fmt.Errorf("GitLab issue creation failed: 422 Unprocessable Entity")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.Issue{
		Number:    999,
		Title:     issue.Title,
		Body:      issue.Body,
		State:     "opened", // GitLab uses "opened" instead of "open"
		Author:    mlp.mockGitLabUser(1, "gitlab-currentuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/issues/999", owner, repo),
	}, nil
}

// UpdateIssue updates an existing GitLab issue
func (mlp *MockGitLabProvider) UpdateIssue(ctx context.Context, owner, repo string, issueNumber int, updates *git.UpdateIssueRequest) (*git.Issue, error) {
	if mlp.shouldFail && mlp.failMethod == "UpdateIssue" {
		return nil, fmt.Errorf("GitLab issue update failed: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	title := fmt.Sprintf("Updated GitLab Mock Issue %d", issueNumber)
	if updates.Title != nil {
		title = *updates.Title
	}

	return &git.Issue{
		Number:    issueNumber,
		Title:     title,
		Body:      "Updated GitLab mock issue body",
		State:     "opened", // GitLab uses "opened" instead of "open"
		Author:    mlp.mockGitLabUser(1, "gitlab-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/issues/%d", owner, repo, issueNumber),
	}, nil
}

// GetPullRequests retrieves merge requests from a GitLab repository
func (mlp *MockGitLabProvider) GetPullRequests(ctx context.Context, owner, repo string, options *git.PullRequestListOptions) ([]*git.PullRequest, error) {
	if mlp.shouldFail && mlp.failMethod == "GetPullRequests" {
		return nil, fmt.Errorf("GitLab merge requests retrieval failed: 500 Internal Server Error")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	// GitLab-specific merge requests data structure (GitLab calls them "merge requests")
	prs := []*git.PullRequest{
		{
			Number:    1,
			Title:     "GitLab Mock Merge Request 1",
			Body:      "This is a mock GitLab merge request for testing\n\n## Changes\n- Added new feature\n- Updated documentation",
			State:     "opened", // GitLab uses "opened" instead of "open"
			Merged:    false,
			Author:    mlp.mockGitLabUser(1, "gitlab-user1"),
			Assignees: []*git.User{mlp.mockGitLabUser(2, "gitlab-assignee1")},
			Labels:    []*git.Label{mlp.mockGitLabLabel(1, "feature")},
			Head:      mlp.mockGitLabBranch("feature-branch", "abc123"),
			Base:      mlp.mockGitLabBranch("master", "def456"), // GitLab typically uses "master"
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Now().UTC(),
			Comments:  git.Comment{ID: 3},
			URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/1", owner, repo),
			Draft:     false,
			Mergeable: &[]bool{true}[0],
		},
		{
			Number:    2,
			Title:     "GitLab Mock Merge Request 2",
			Body:      "Another mock GitLab merge request for testing",
			State:     "merged", // GitLab uses "merged" instead of "closed"
			Merged:    true,
			MergedAt:  &time.Time{},
			Author:    mlp.mockGitLabUser(3, "gitlab-user2"),
			Assignees: []*git.User{},
			Labels:    []*git.Label{mlp.mockGitLabLabel(2, "bugfix")},
			Head:      mlp.mockGitLabBranch("bugfix-branch", "ghi789"),
			Base:      mlp.mockGitLabBranch("master", "def456"),
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			ClosedAt:  &time.Time{},
			Comments:  git.Comment{ID: 1},
			URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/2", owner, repo),
			Draft:     false,
			Mergeable: &[]bool{true}[0],
		},
	}

	return prs, nil
}

// GetPullRequest retrieves a specific GitLab merge request by number
func (mlp *MockGitLabProvider) GetPullRequest(ctx context.Context, owner, repo string, prNumber int) (*git.PullRequest, error) {
	if mlp.shouldFail && mlp.failMethod == "GetPullRequest" {
		return nil, fmt.Errorf("GitLab merge request not found: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.PullRequest{
		Number:    prNumber,
		Title:     fmt.Sprintf("GitLab Mock Merge Request %d", prNumber),
		Body:      fmt.Sprintf("This is mock GitLab merge request %d for testing", prNumber),
		State:     "opened", // GitLab uses "opened" instead of "open"
		Merged:    false,
		Author:    mlp.mockGitLabUser(1, "gitlab-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{mlp.mockGitLabLabel(1, "feature")},
		Head:      mlp.mockGitLabBranch("feature-branch", "abc123"),
		Base:      mlp.mockGitLabBranch("master", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 2},
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/%d", owner, repo, prNumber),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestByBranch retrieves a GitLab merge request by source branch name
func (mlp *MockGitLabProvider) GetPullRequestByBranch(ctx context.Context, owner, repo string, branchName string) (*git.PullRequest, error) {
	if mlp.shouldFail && mlp.failMethod == "GetPullRequestByBranch" {
		return nil, fmt.Errorf("GitLab merge request not found for branch: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.PullRequest{
		Number:    123,
		Title:     fmt.Sprintf("GitLab Mock MR for branch %s", branchName),
		Body:      fmt.Sprintf("This is a mock GitLab merge request for branch %s", branchName),
		State:     "opened", // GitLab uses "opened" instead of "open"
		Merged:    false,
		Author:    mlp.mockGitLabUser(1, "gitlab-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mlp.mockGitLabBranch(branchName, "abc123"),
		Base:      mlp.mockGitLabBranch("master", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/123", owner, repo),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestByIssue retrieves a GitLab merge request that closes a specific issue
func (mlp *MockGitLabProvider) GetPullRequestByIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.PullRequest, error) {
	if mlp.shouldFail && mlp.failMethod == "GetPullRequestByIssue" {
		return nil, fmt.Errorf("GitLab merge request not found for issue: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.PullRequest{
		Number:    456,
		Title:     fmt.Sprintf("GitLab Mock MR closing issue %d", issueNumber),
		Body:      fmt.Sprintf("This GitLab merge request closes issue #%d", issueNumber),
		State:     "opened", // GitLab uses "opened" instead of "open"
		Merged:    false,
		Author:    mlp.mockGitLabUser(1, "gitlab-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mlp.mockGitLabBranch("fix-issue", "abc123"),
		Base:      mlp.mockGitLabBranch("master", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/456", owner, repo),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// CreatePullRequest creates a new GitLab merge request
func (mlp *MockGitLabProvider) CreatePullRequest(ctx context.Context, owner, repo string, pr *git.CreatePullRequestRequest) (*git.PullRequest, error) {
	if mlp.shouldFail && mlp.failMethod == "CreatePullRequest" {
		return nil, fmt.Errorf("GitLab merge request creation failed: 422 Unprocessable Entity")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.PullRequest{
		Number:    999,
		Title:     pr.Title,
		Body:      pr.Body,
		State:     "opened", // GitLab uses "opened" instead of "open"
		Merged:    false,
		Author:    mlp.mockGitLabUser(1, "gitlab-currentuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mlp.mockGitLabBranch(pr.Head, "abc123"),
		Base:      mlp.mockGitLabBranch(pr.Base, "def456"),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/999", owner, repo),
		Draft:     pr.Draft,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// UpdatePullRequest updates an existing GitLab merge request
func (mlp *MockGitLabProvider) UpdatePullRequest(ctx context.Context, owner, repo string, prNumber int, updates *git.UpdatePullRequestRequest) (*git.PullRequest, error) {
	if mlp.shouldFail && mlp.failMethod == "UpdatePullRequest" {
		return nil, fmt.Errorf("GitLab merge request update failed: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	title := fmt.Sprintf("Updated GitLab Mock MR %d", prNumber)
	if updates.Title != nil {
		title = *updates.Title
	}

	return &git.PullRequest{
		Number:    prNumber,
		Title:     title,
		Body:      "Updated GitLab mock merge request body",
		State:     "opened", // GitLab uses "opened" instead of "open"
		Merged:    false,
		Author:    mlp.mockGitLabUser(1, "gitlab-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mlp.mockGitLabBranch("feature-branch", "abc123"),
		Base:      mlp.mockGitLabBranch("master", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/%d", owner, repo, prNumber),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestReviews retrieves reviews for a GitLab merge request
func (mlp *MockGitLabProvider) GetPullRequestReviews(ctx context.Context, owner, repo string, prNumber int) ([]*git.Review, error) {
	if mlp.shouldFail && mlp.failMethod == "GetPullRequestReviews" {
		return nil, fmt.Errorf("GitLab merge request reviews retrieval failed: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	reviews := []*git.Review{
		{
			ID:          1,
			User:        mlp.mockGitLabUser(1, "gitlab-reviewer1"),
			Body:        "This looks good!",
			State:       "approved",
			SubmittedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			CommitID:    "abc123",
			URL:         fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/%d/approvals", owner, repo, prNumber),
		},
		{
			ID:          2,
			User:        mlp.mockGitLabUser(2, "gitlab-reviewer2"),
			Body:        "Please fix the formatting",
			State:       "changes_requested",
			SubmittedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			CommitID:    "def456",
			URL:         fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/%d/approvals", owner, repo, prNumber),
		},
	}

	return reviews, nil
}

// GetPullRequestComments retrieves comments for a GitLab merge request
func (mlp *MockGitLabProvider) GetPullRequestComments(ctx context.Context, owner, repo string, prNumber int) ([]*git.Comment, error) {
	if mlp.shouldFail && mlp.failMethod == "GetPullRequestComments" {
		return nil, fmt.Errorf("GitLab merge request comments retrieval failed: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	comments := []*git.Comment{
		{
			ID:        1,
			User:      mlp.mockGitLabUser(1, "gitlab-commenter1"),
			Body:      "Great work on this MR!",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/%d/notes/1", owner, repo, prNumber),
		},
		{
			ID:        2,
			User:      mlp.mockGitLabUser(2, "gitlab-commenter2"),
			Body:      "Can you add more tests?",
			CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/merge_requests/%d/notes/2", owner, repo, prNumber),
		},
	}

	return comments, nil
}

// GetIssueComments retrieves comments for a GitLab issue
func (mlp *MockGitLabProvider) GetIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*git.Comment, error) {
	if mlp.shouldFail && mlp.failMethod == "GetIssueComments" {
		return nil, fmt.Errorf("GitLab issue comments retrieval failed: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	comments := []*git.Comment{
		{
			ID:        1,
			User:      mlp.mockGitLabUser(1, "gitlab-commenter1"),
			Body:      "This issue is important",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/issues/%d/notes/1", owner, repo, issueNumber),
		},
		{
			ID:        2,
			User:      mlp.mockGitLabUser(2, "gitlab-commenter2"),
			Body:      "I can help with this",
			CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/issues/%d/notes/2", owner, repo, issueNumber),
		},
	}

	return comments, nil
}

// CreateComment creates a comment on a GitLab issue or merge request
func (mlp *MockGitLabProvider) CreateComment(ctx context.Context, owner, repo string, issueNumber int, comment *git.CreateCommentRequest) (*git.Comment, error) {
	if mlp.shouldFail && mlp.failMethod == "CreateComment" {
		return nil, fmt.Errorf("GitLab comment creation failed: 422 Unprocessable Entity")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.Comment{
		ID:        999,
		User:      mlp.mockGitLabUser(1, "gitlab-currentuser"),
		Body:      comment.Body,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		URL:       fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/issues/%d/notes/999", owner, repo, issueNumber),
	}, nil
}

// GetLabels retrieves available labels for a GitLab repository
func (mlp *MockGitLabProvider) GetLabels(ctx context.Context, owner, repo string) ([]*git.Label, error) {
	if mlp.shouldFail && mlp.failMethod == "GetLabels" {
		return nil, fmt.Errorf("GitLab labels retrieval failed: 404 Not Found")
	}
	if mlp.rateLimited {
		return nil, fmt.Errorf("GitLab API rate limit exceeded: 429 Too Many Requests")
	}

	// GitLab-specific labels with their typical descriptions
	labels := []*git.Label{
		{
			ID:          1,
			Name:        "bug",
			Description: "Something isn't working",
			URL:         fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/labels/bug", owner, repo),
		},
		{
			ID:          2,
			Name:        "enhancement",
			Description: "New feature or request",
			URL:         fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/labels/enhancement", owner, repo),
		},
		{
			ID:          3,
			Name:        "documentation",
			Description: "Improvements or additions to documentation",
			URL:         fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/labels/documentation", owner, repo),
		},
		{
			ID:          4,
			Name:        "help wanted",
			Description: "Extra attention is needed",
			URL:         fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/labels/help%%20wanted", owner, repo),
		},
		{
			ID:          5,
			Name:        "wontfix",
			Description: "This will not be worked on",
			URL:         fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s/labels/wontfix", owner, repo),
		},
	}

	return labels, nil
}

// Helper methods to create GitLab-specific mock data

func (mlp *MockGitLabProvider) mockGitLabUser(id int, login string) *git.User {
	return &git.User{
		ID:    id,
		Login: login,
		Name:  fmt.Sprintf("GitLab User %d", id),
		Email: fmt.Sprintf("%s@gitlab.com", login),
		Type:  "User",
	}
}

func (mlp *MockGitLabProvider) mockGitLabLabel(id int, name string) *git.Label {
	return &git.Label{
		ID:          id,
		Name:        name,
		Description: fmt.Sprintf("GitLab mock label: %s", name),
		URL:         fmt.Sprintf("https://gitlab.com/api/v4/projects/test/owner/labels/%s", name),
	}
}

func (mlp *MockGitLabProvider) mockGitLabBranch(ref, sha string) *git.Branch {
	return &git.Branch{
		Ref:  ref,
		SHA:  sha,
		Repo: &git.Repository{Owner: "test", Name: "repo"},
		User: mlp.mockGitLabUser(1, "gitlab-user"),
	}
}
