package gitprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/hlfshell/cowork/internal/git"
)

// MockBitbucketProvider implements the GitProvider interface with Bitbucket-specific behavior
// This mock provides realistic responses based on Bitbucket API documentation
type MockBitbucketProvider struct {
	shouldFail  bool
	failMethod  string
	rateLimited bool
}

// NewMockBitbucketProvider creates a new Bitbucket mock provider for testing
func NewMockBitbucketProvider() *MockBitbucketProvider {
	return &MockBitbucketProvider{
		shouldFail:  false,
		rateLimited: false,
	}
}

// NewMockBitbucketProviderWithFailure creates a Bitbucket mock provider that fails on specific methods
func NewMockBitbucketProviderWithFailure(failMethod string) *MockBitbucketProvider {
	return &MockBitbucketProvider{
		shouldFail:  true,
		failMethod:  failMethod,
		rateLimited: false,
	}
}

// NewMockBitbucketProviderWithRateLimit creates a Bitbucket mock provider that simulates rate limiting
func NewMockBitbucketProviderWithRateLimit() *MockBitbucketProvider {
	return &MockBitbucketProvider{
		shouldFail:  false,
		rateLimited: true,
	}
}

// GetProviderType returns the type of this provider
func (mbp *MockBitbucketProvider) GetProviderType() git.ProviderType {
	return git.ProviderBitbucket
}

// TestAuth verifies if the provided authentication is valid
func (mbp *MockBitbucketProvider) TestAuth(ctx context.Context) error {
	if mbp.shouldFail && mbp.failMethod == "TestAuth" {
		return fmt.Errorf("Bitbucket authentication failed: 401 Unauthorized")
	}
	if mbp.rateLimited {
		return fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}
	return nil
}

// GetRepositoryInfo retrieves basic information about a Bitbucket repository
func (mbp *MockBitbucketProvider) GetRepositoryInfo(ctx context.Context, owner, repo string) (*git.Repository, error) {
	if mbp.shouldFail && mbp.failMethod == "GetRepositoryInfo" {
		return nil, fmt.Errorf("Bitbucket repository not found: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	// Bitbucket-specific repository data structure
	return &git.Repository{
		Owner:         owner,
		Name:          repo,
		FullName:      fmt.Sprintf("%s/%s", owner, repo),
		Description:   "A mock Bitbucket repository for testing purposes",
		URL:           fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", owner, repo),
		Private:       false,
		DefaultBranch: "main", // Bitbucket's default branch is typically "main"
		CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

// GetIssues retrieves issues from a Bitbucket repository with optional filtering
func (mbp *MockBitbucketProvider) GetIssues(ctx context.Context, owner, repo string, options *git.IssueListOptions) ([]*git.Issue, error) {
	if mbp.shouldFail && mbp.failMethod == "GetIssues" {
		return nil, fmt.Errorf("Bitbucket issues retrieval failed: 500 Internal Server Error")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	// Bitbucket-specific issues data structure
	issues := []*git.Issue{
		{
			Number:    1,
			Title:     "Bitbucket Mock Issue 1",
			Body:      "This is a mock Bitbucket issue for testing\n\n## Description\nThis issue demonstrates Bitbucket's markdown support.",
			State:     "open",
			Author:    mbp.mockBitbucketUser(1, "bitbucket-user1"),
			Assignees: []*git.User{mbp.mockBitbucketUser(2, "bitbucket-assignee1")},
			Labels:    []*git.Label{mbp.mockBitbucketLabel(1, "bug"), mbp.mockBitbucketLabel(2, "enhancement")},
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Now().UTC(),
			Comments:  git.Comment{ID: 5},
			URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/issues/1", owner, repo),
		},
		{
			Number:    2,
			Title:     "Bitbucket Mock Issue 2",
			Body:      "Another mock Bitbucket issue for testing",
			State:     "resolved", // Bitbucket uses "resolved" instead of "closed"
			Author:    mbp.mockBitbucketUser(3, "bitbucket-user2"),
			Assignees: []*git.User{},
			Labels:    []*git.Label{mbp.mockBitbucketLabel(3, "documentation")},
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			ClosedAt:  &time.Time{},
			Comments:  git.Comment{ID: 2},
			URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/issues/2", owner, repo),
		},
	}

	return issues, nil
}

// GetIssue retrieves a specific Bitbucket issue by ID
func (mbp *MockBitbucketProvider) GetIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.Issue, error) {
	if mbp.shouldFail && mbp.failMethod == "GetIssue" {
		return nil, fmt.Errorf("Bitbucket issue not found: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.Issue{
		Number:    issueNumber,
		Title:     fmt.Sprintf("Bitbucket Mock Issue %d", issueNumber),
		Body:      fmt.Sprintf("This is mock Bitbucket issue %d for testing", issueNumber),
		State:     "open",
		Author:    mbp.mockBitbucketUser(1, "bitbucket-user"),
		Assignees: []*git.User{mbp.mockBitbucketUser(2, "bitbucket-assignee")},
		Labels:    []*git.Label{mbp.mockBitbucketLabel(1, "bug")},
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 3},
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/issues/%d", owner, repo, issueNumber),
	}, nil
}

// CreateIssue creates a new Bitbucket issue
func (mbp *MockBitbucketProvider) CreateIssue(ctx context.Context, owner, repo string, issue *git.CreateIssueRequest) (*git.Issue, error) {
	if mbp.shouldFail && mbp.failMethod == "CreateIssue" {
		return nil, fmt.Errorf("Bitbucket issue creation failed: 422 Unprocessable Entity")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.Issue{
		Number:    999,
		Title:     issue.Title,
		Body:      issue.Body,
		State:     "open",
		Author:    mbp.mockBitbucketUser(1, "bitbucket-currentuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/issues/999", owner, repo),
	}, nil
}

// UpdateIssue updates an existing Bitbucket issue
func (mbp *MockBitbucketProvider) UpdateIssue(ctx context.Context, owner, repo string, issueNumber int, updates *git.UpdateIssueRequest) (*git.Issue, error) {
	if mbp.shouldFail && mbp.failMethod == "UpdateIssue" {
		return nil, fmt.Errorf("Bitbucket issue update failed: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	title := fmt.Sprintf("Updated Bitbucket Mock Issue %d", issueNumber)
	if updates.Title != nil {
		title = *updates.Title
	}

	return &git.Issue{
		Number:    issueNumber,
		Title:     title,
		Body:      "Updated Bitbucket mock issue body",
		State:     "open",
		Author:    mbp.mockBitbucketUser(1, "bitbucket-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/issues/%d", owner, repo, issueNumber),
	}, nil
}

// GetPullRequests retrieves pull requests from a Bitbucket repository
func (mbp *MockBitbucketProvider) GetPullRequests(ctx context.Context, owner, repo string, options *git.PullRequestListOptions) ([]*git.PullRequest, error) {
	if mbp.shouldFail && mbp.failMethod == "GetPullRequests" {
		return nil, fmt.Errorf("Bitbucket pull requests retrieval failed: 500 Internal Server Error")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	// Bitbucket-specific pull requests data structure
	prs := []*git.PullRequest{
		{
			Number:    1,
			Title:     "Bitbucket Mock Pull Request 1",
			Body:      "This is a mock Bitbucket pull request for testing\n\n## Changes\n- Added new feature\n- Updated documentation",
			State:     "OPEN", // Bitbucket uses uppercase states
			Merged:    false,
			Author:    mbp.mockBitbucketUser(1, "bitbucket-user1"),
			Assignees: []*git.User{mbp.mockBitbucketUser(2, "bitbucket-assignee1")},
			Labels:    []*git.Label{mbp.mockBitbucketLabel(1, "feature")},
			Head:      mbp.mockBitbucketBranch("feature-branch", "abc123"),
			Base:      mbp.mockBitbucketBranch("main", "def456"),
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Now().UTC(),
			Comments:  git.Comment{ID: 3},
			URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/1", owner, repo),
			Draft:     false,
			Mergeable: &[]bool{true}[0],
		},
		{
			Number:    2,
			Title:     "Bitbucket Mock Pull Request 2",
			Body:      "Another mock Bitbucket pull request for testing",
			State:     "MERGED", // Bitbucket uses "MERGED" instead of "closed"
			Merged:    true,
			MergedAt:  &time.Time{},
			Author:    mbp.mockBitbucketUser(3, "bitbucket-user2"),
			Assignees: []*git.User{},
			Labels:    []*git.Label{mbp.mockBitbucketLabel(2, "bugfix")},
			Head:      mbp.mockBitbucketBranch("bugfix-branch", "ghi789"),
			Base:      mbp.mockBitbucketBranch("main", "def456"),
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			ClosedAt:  &time.Time{},
			Comments:  git.Comment{ID: 1},
			URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/2", owner, repo),
			Draft:     false,
			Mergeable: &[]bool{true}[0],
		},
	}

	return prs, nil
}

// GetPullRequest retrieves a specific Bitbucket pull request by number
func (mbp *MockBitbucketProvider) GetPullRequest(ctx context.Context, owner, repo string, prNumber int) (*git.PullRequest, error) {
	if mbp.shouldFail && mbp.failMethod == "GetPullRequest" {
		return nil, fmt.Errorf("Bitbucket pull request not found: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.PullRequest{
		Number:    prNumber,
		Title:     fmt.Sprintf("Bitbucket Mock Pull Request %d", prNumber),
		Body:      fmt.Sprintf("This is mock Bitbucket pull request %d for testing", prNumber),
		State:     "OPEN", // Bitbucket uses uppercase states
		Merged:    false,
		Author:    mbp.mockBitbucketUser(1, "bitbucket-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{mbp.mockBitbucketLabel(1, "feature")},
		Head:      mbp.mockBitbucketBranch("feature-branch", "abc123"),
		Base:      mbp.mockBitbucketBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 2},
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d", owner, repo, prNumber),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestByBranch retrieves a Bitbucket pull request by source branch name
func (mbp *MockBitbucketProvider) GetPullRequestByBranch(ctx context.Context, owner, repo string, branchName string) (*git.PullRequest, error) {
	if mbp.shouldFail && mbp.failMethod == "GetPullRequestByBranch" {
		return nil, fmt.Errorf("Bitbucket pull request not found for branch: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.PullRequest{
		Number:    123,
		Title:     fmt.Sprintf("Bitbucket Mock PR for branch %s", branchName),
		Body:      fmt.Sprintf("This is a mock Bitbucket pull request for branch %s", branchName),
		State:     "OPEN", // Bitbucket uses uppercase states
		Merged:    false,
		Author:    mbp.mockBitbucketUser(1, "bitbucket-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mbp.mockBitbucketBranch(branchName, "abc123"),
		Base:      mbp.mockBitbucketBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/123", owner, repo),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestByIssue retrieves a Bitbucket pull request that closes a specific issue
func (mbp *MockBitbucketProvider) GetPullRequestByIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.PullRequest, error) {
	if mbp.shouldFail && mbp.failMethod == "GetPullRequestByIssue" {
		return nil, fmt.Errorf("Bitbucket pull request not found for issue: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.PullRequest{
		Number:    456,
		Title:     fmt.Sprintf("Bitbucket Mock PR closing issue %d", issueNumber),
		Body:      fmt.Sprintf("This Bitbucket pull request closes issue #%d", issueNumber),
		State:     "OPEN", // Bitbucket uses uppercase states
		Merged:    false,
		Author:    mbp.mockBitbucketUser(1, "bitbucket-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mbp.mockBitbucketBranch("fix-issue", "abc123"),
		Base:      mbp.mockBitbucketBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/456", owner, repo),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// CreatePullRequest creates a new Bitbucket pull request
func (mbp *MockBitbucketProvider) CreatePullRequest(ctx context.Context, owner, repo string, pr *git.CreatePullRequestRequest) (*git.PullRequest, error) {
	if mbp.shouldFail && mbp.failMethod == "CreatePullRequest" {
		return nil, fmt.Errorf("Bitbucket pull request creation failed: 422 Unprocessable Entity")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.PullRequest{
		Number:    999,
		Title:     pr.Title,
		Body:      pr.Body,
		State:     "OPEN", // Bitbucket uses uppercase states
		Merged:    false,
		Author:    mbp.mockBitbucketUser(1, "bitbucket-currentuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mbp.mockBitbucketBranch(pr.Head, "abc123"),
		Base:      mbp.mockBitbucketBranch(pr.Base, "def456"),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/999", owner, repo),
		Draft:     pr.Draft,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// UpdatePullRequest updates an existing Bitbucket pull request
func (mbp *MockBitbucketProvider) UpdatePullRequest(ctx context.Context, owner, repo string, prNumber int, updates *git.UpdatePullRequestRequest) (*git.PullRequest, error) {
	if mbp.shouldFail && mbp.failMethod == "UpdatePullRequest" {
		return nil, fmt.Errorf("Bitbucket pull request update failed: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	title := fmt.Sprintf("Updated Bitbucket Mock PR %d", prNumber)
	if updates.Title != nil {
		title = *updates.Title
	}

	return &git.PullRequest{
		Number:    prNumber,
		Title:     title,
		Body:      "Updated Bitbucket mock pull request body",
		State:     "OPEN", // Bitbucket uses uppercase states
		Merged:    false,
		Author:    mbp.mockBitbucketUser(1, "bitbucket-user"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mbp.mockBitbucketBranch("feature-branch", "abc123"),
		Base:      mbp.mockBitbucketBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d", owner, repo, prNumber),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestReviews retrieves reviews for a Bitbucket pull request
func (mbp *MockBitbucketProvider) GetPullRequestReviews(ctx context.Context, owner, repo string, prNumber int) ([]*git.Review, error) {
	if mbp.shouldFail && mbp.failMethod == "GetPullRequestReviews" {
		return nil, fmt.Errorf("Bitbucket pull request reviews retrieval failed: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	reviews := []*git.Review{
		{
			ID:          1,
			User:        mbp.mockBitbucketUser(1, "bitbucket-reviewer1"),
			Body:        "This looks good!",
			State:       "approved",
			SubmittedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			CommitID:    "abc123",
			URL:         fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d/approve", owner, repo, prNumber),
		},
		{
			ID:          2,
			User:        mbp.mockBitbucketUser(2, "bitbucket-reviewer2"),
			Body:        "Please fix the formatting",
			State:       "changes_requested",
			SubmittedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			CommitID:    "def456",
			URL:         fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d/approve", owner, repo, prNumber),
		},
	}

	return reviews, nil
}

// GetPullRequestComments retrieves comments for a Bitbucket pull request
func (mbp *MockBitbucketProvider) GetPullRequestComments(ctx context.Context, owner, repo string, prNumber int) ([]*git.Comment, error) {
	if mbp.shouldFail && mbp.failMethod == "GetPullRequestComments" {
		return nil, fmt.Errorf("Bitbucket pull request comments retrieval failed: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	comments := []*git.Comment{
		{
			ID:        1,
			User:      mbp.mockBitbucketUser(1, "bitbucket-commenter1"),
			Body:      "Great work on this PR!",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d/comments/1", owner, repo, prNumber),
		},
		{
			ID:        2,
			User:      mbp.mockBitbucketUser(2, "bitbucket-commenter2"),
			Body:      "Can you add more tests?",
			CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d/comments/2", owner, repo, prNumber),
		},
	}

	return comments, nil
}

// GetIssueComments retrieves comments for a Bitbucket issue
func (mbp *MockBitbucketProvider) GetIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*git.Comment, error) {
	if mbp.shouldFail && mbp.failMethod == "GetIssueComments" {
		return nil, fmt.Errorf("Bitbucket issue comments retrieval failed: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	comments := []*git.Comment{
		{
			ID:        1,
			User:      mbp.mockBitbucketUser(1, "bitbucket-commenter1"),
			Body:      "This issue is important",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/issues/%d/comments/1", owner, repo, issueNumber),
		},
		{
			ID:        2,
			User:      mbp.mockBitbucketUser(2, "bitbucket-commenter2"),
			Body:      "I can help with this",
			CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/issues/%d/comments/2", owner, repo, issueNumber),
		},
	}

	return comments, nil
}

// CreateComment creates a comment on a Bitbucket issue or pull request
func (mbp *MockBitbucketProvider) CreateComment(ctx context.Context, owner, repo string, issueNumber int, comment *git.CreateCommentRequest) (*git.Comment, error) {
	if mbp.shouldFail && mbp.failMethod == "CreateComment" {
		return nil, fmt.Errorf("Bitbucket comment creation failed: 422 Unprocessable Entity")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	return &git.Comment{
		ID:        999,
		User:      mbp.mockBitbucketUser(1, "bitbucket-currentuser"),
		Body:      comment.Body,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		URL:       fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/issues/%d/comments/999", owner, repo, issueNumber),
	}, nil
}

// GetLabels retrieves available labels for a Bitbucket repository
func (mbp *MockBitbucketProvider) GetLabels(ctx context.Context, owner, repo string) ([]*git.Label, error) {
	if mbp.shouldFail && mbp.failMethod == "GetLabels" {
		return nil, fmt.Errorf("Bitbucket labels retrieval failed: 404 Not Found")
	}
	if mbp.rateLimited {
		return nil, fmt.Errorf("Bitbucket API rate limit exceeded: 429 Too Many Requests")
	}

	// Bitbucket-specific labels with their typical descriptions
	labels := []*git.Label{
		{
			ID:          1,
			Name:        "bug",
			Description: "Something isn't working",
			URL:         fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/labels/bug", owner, repo),
		},
		{
			ID:          2,
			Name:        "enhancement",
			Description: "New feature or request",
			URL:         fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/labels/enhancement", owner, repo),
		},
		{
			ID:          3,
			Name:        "documentation",
			Description: "Improvements or additions to documentation",
			URL:         fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/labels/documentation", owner, repo),
		},
		{
			ID:          4,
			Name:        "help wanted",
			Description: "Extra attention is needed",
			URL:         fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/labels/help%%20wanted", owner, repo),
		},
		{
			ID:          5,
			Name:        "invalid",
			Description: "This doesn't seem right",
			URL:         fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/labels/invalid", owner, repo),
		},
	}

	return labels, nil
}

// Helper methods to create Bitbucket-specific mock data

func (mbp *MockBitbucketProvider) mockBitbucketUser(id int, login string) *git.User {
	return &git.User{
		ID:    id,
		Login: login,
		Name:  fmt.Sprintf("Bitbucket User %d", id),
		Email: fmt.Sprintf("%s@bitbucket.org", login),
		Type:  "User",
	}
}

func (mbp *MockBitbucketProvider) mockBitbucketLabel(id int, name string) *git.Label {
	return &git.Label{
		ID:          id,
		Name:        name,
		Description: fmt.Sprintf("Bitbucket mock label: %s", name),
		URL:         fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/test/owner/labels/%s", name),
	}
}

func (mbp *MockBitbucketProvider) mockBitbucketBranch(ref, sha string) *git.Branch {
	return &git.Branch{
		Ref:  ref,
		SHA:  sha,
		Repo: &git.Repository{Owner: "test", Name: "repo"},
		User: mbp.mockBitbucketUser(1, "bitbucket-user"),
	}
}
