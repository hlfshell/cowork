package gitprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/hlfshell/cowork/internal/git"
)

// MockProvider implements the GitProvider interface for testing
// This mock provides realistic responses based on GitHub and GitLab API documentation
type MockProvider struct {
	providerType git.ProviderType
	shouldFail   bool
	failMethod   string
}

// NewMockProvider creates a new mock provider for testing
func NewMockProvider(providerType git.ProviderType) *MockProvider {
	return &MockProvider{
		providerType: providerType,
		shouldFail:   false,
	}
}

// NewMockProviderWithFailure creates a mock provider that fails on specific methods
func NewMockProviderWithFailure(providerType git.ProviderType, failMethod string) *MockProvider {
	return &MockProvider{
		providerType: providerType,
		shouldFail:   true,
		failMethod:   failMethod,
	}
}

// GetProviderType returns the type of this provider
func (mp *MockProvider) GetProviderType() git.ProviderType {
	return mp.providerType
}

// TestAuth verifies if the provided authentication is valid
func (mp *MockProvider) TestAuth(ctx context.Context) error {
	if mp.shouldFail && mp.failMethod == "TestAuth" {
		return fmt.Errorf("mock authentication failed")
	}
	return nil
}

// GetRepositoryInfo retrieves basic information about a repository
func (mp *MockProvider) GetRepositoryInfo(ctx context.Context, owner, repo string) (*git.Repository, error) {
	if mp.shouldFail && mp.failMethod == "GetRepositoryInfo" {
		return nil, fmt.Errorf("mock repository info retrieval failed")
	}

	// Mock repository data based on GitHub/GitLab API responses
	return &git.Repository{
		Owner:         owner,
		Name:          repo,
		FullName:      fmt.Sprintf("%s/%s", owner, repo),
		Description:   "A mock repository for testing purposes",
		URL:           fmt.Sprintf("https://api.%s.com/repos/%s/%s", mp.providerType, owner, repo),
		Private:       false,
		DefaultBranch: "main",
		CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

// GetIssues retrieves issues from a repository with optional filtering
func (mp *MockProvider) GetIssues(ctx context.Context, owner, repo string, options *git.IssueListOptions) ([]*git.Issue, error) {
	if mp.shouldFail && mp.failMethod == "GetIssues" {
		return nil, fmt.Errorf("mock issues retrieval failed")
	}

	// Mock issues data based on GitHub/GitLab API responses
	issues := []*git.Issue{
		{
			Number:    1,
			Title:     "Mock Issue 1",
			Body:      "This is a mock issue for testing",
			State:     "open",
			Author:    mp.mockUser(1, "testuser1"),
			Assignees: []*git.User{mp.mockUser(2, "assignee1")},
			Labels:    []*git.Label{mp.mockLabel(1, "bug"), mp.mockLabel(2, "enhancement")},
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Now().UTC(),
			Comments:  git.Comment{ID: 5},
			URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/issues/1", mp.providerType, owner, repo),
		},
		{
			Number:    2,
			Title:     "Mock Issue 2",
			Body:      "Another mock issue for testing",
			State:     "closed",
			Author:    mp.mockUser(3, "testuser2"),
			Assignees: []*git.User{},
			Labels:    []*git.Label{mp.mockLabel(3, "documentation")},
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			ClosedAt:  &time.Time{},
			Comments:  git.Comment{ID: 2},
			URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/issues/2", mp.providerType, owner, repo),
		},
	}

	return issues, nil
}

// GetIssue retrieves a specific issue by ID
func (mp *MockProvider) GetIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.Issue, error) {
	if mp.shouldFail && mp.failMethod == "GetIssue" {
		return nil, fmt.Errorf("mock issue retrieval failed")
	}

	return &git.Issue{
		Number:    issueNumber,
		Title:     fmt.Sprintf("Mock Issue %d", issueNumber),
		Body:      fmt.Sprintf("This is mock issue %d for testing", issueNumber),
		State:     "open",
		Author:    mp.mockUser(1, "testuser"),
		Assignees: []*git.User{mp.mockUser(2, "assignee")},
		Labels:    []*git.Label{mp.mockLabel(1, "bug")},
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 3},
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/issues/%d", mp.providerType, owner, repo, issueNumber),
	}, nil
}

// CreateIssue creates a new issue in the repository
func (mp *MockProvider) CreateIssue(ctx context.Context, owner, repo string, issue *git.CreateIssueRequest) (*git.Issue, error) {
	if mp.shouldFail && mp.failMethod == "CreateIssue" {
		return nil, fmt.Errorf("mock issue creation failed")
	}

	return &git.Issue{
		Number:    999,
		Title:     issue.Title,
		Body:      issue.Body,
		State:     "open",
		Author:    mp.mockUser(1, "currentuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/issues/999", mp.providerType, owner, repo),
	}, nil
}

// UpdateIssue updates an existing issue
func (mp *MockProvider) UpdateIssue(ctx context.Context, owner, repo string, issueNumber int, updates *git.UpdateIssueRequest) (*git.Issue, error) {
	if mp.shouldFail && mp.failMethod == "UpdateIssue" {
		return nil, fmt.Errorf("mock issue update failed")
	}

	title := fmt.Sprintf("Updated Mock Issue %d", issueNumber)
	if updates.Title != nil {
		title = *updates.Title
	}

	return &git.Issue{
		Number:    issueNumber,
		Title:     title,
		Body:      "Updated mock issue body",
		State:     "open",
		Author:    mp.mockUser(1, "testuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/issues/%d", mp.providerType, owner, repo, issueNumber),
	}, nil
}

// GetPullRequests retrieves pull requests from a repository with optional filtering
func (mp *MockProvider) GetPullRequests(ctx context.Context, owner, repo string, options *git.PullRequestListOptions) ([]*git.PullRequest, error) {
	if mp.shouldFail && mp.failMethod == "GetPullRequests" {
		return nil, fmt.Errorf("mock pull requests retrieval failed")
	}

	// Mock pull requests data based on GitHub/GitLab API responses
	prs := []*git.PullRequest{
		{
			Number:    1,
			Title:     "Mock Pull Request 1",
			Body:      "This is a mock pull request for testing",
			State:     "open",
			Merged:    false,
			Author:    mp.mockUser(1, "testuser1"),
			Assignees: []*git.User{mp.mockUser(2, "assignee1")},
			Labels:    []*git.Label{mp.mockLabel(1, "feature")},
			Head:      mp.mockBranch("feature-branch", "abc123"),
			Base:      mp.mockBranch("main", "def456"),
			CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Now().UTC(),
			Comments:  git.Comment{ID: 3},
			URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/1", mp.providerType, owner, repo),
			Draft:     false,
			Mergeable: &[]bool{true}[0],
		},
		{
			Number:    2,
			Title:     "Mock Pull Request 2",
			Body:      "Another mock pull request for testing",
			State:     "closed",
			Merged:    true,
			MergedAt:  &time.Time{},
			Author:    mp.mockUser(3, "testuser2"),
			Assignees: []*git.User{},
			Labels:    []*git.Label{mp.mockLabel(2, "bugfix")},
			Head:      mp.mockBranch("bugfix-branch", "ghi789"),
			Base:      mp.mockBranch("main", "def456"),
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			ClosedAt:  &time.Time{},
			Comments:  git.Comment{ID: 1},
			URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/2", mp.providerType, owner, repo),
			Draft:     false,
			Mergeable: &[]bool{true}[0],
		},
	}

	return prs, nil
}

// GetPullRequest retrieves a specific pull request by number
func (mp *MockProvider) GetPullRequest(ctx context.Context, owner, repo string, prNumber int) (*git.PullRequest, error) {
	if mp.shouldFail && mp.failMethod == "GetPullRequest" {
		return nil, fmt.Errorf("mock pull request retrieval failed")
	}

	return &git.PullRequest{
		Number:    prNumber,
		Title:     fmt.Sprintf("Mock Pull Request %d", prNumber),
		Body:      fmt.Sprintf("This is mock pull request %d for testing", prNumber),
		State:     "open",
		Merged:    false,
		Author:    mp.mockUser(1, "testuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{mp.mockLabel(1, "feature")},
		Head:      mp.mockBranch("feature-branch", "abc123"),
		Base:      mp.mockBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 2},
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/%d", mp.providerType, owner, repo, prNumber),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestByBranch retrieves a pull request by source branch name
func (mp *MockProvider) GetPullRequestByBranch(ctx context.Context, owner, repo string, branchName string) (*git.PullRequest, error) {
	if mp.shouldFail && mp.failMethod == "GetPullRequestByBranch" {
		return nil, fmt.Errorf("mock pull request by branch retrieval failed")
	}

	return &git.PullRequest{
		Number:    123,
		Title:     fmt.Sprintf("Mock PR for branch %s", branchName),
		Body:      fmt.Sprintf("This is a mock pull request for branch %s", branchName),
		State:     "open",
		Merged:    false,
		Author:    mp.mockUser(1, "testuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mp.mockBranch(branchName, "abc123"),
		Base:      mp.mockBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/123", mp.providerType, owner, repo),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestByIssue retrieves a pull request that closes a specific issue
func (mp *MockProvider) GetPullRequestByIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.PullRequest, error) {
	if mp.shouldFail && mp.failMethod == "GetPullRequestByIssue" {
		return nil, fmt.Errorf("mock pull request by issue retrieval failed")
	}

	return &git.PullRequest{
		Number:    456,
		Title:     fmt.Sprintf("Mock PR closing issue %d", issueNumber),
		Body:      fmt.Sprintf("This pull request closes issue #%d", issueNumber),
		State:     "open",
		Merged:    false,
		Author:    mp.mockUser(1, "testuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mp.mockBranch("fix-issue", "abc123"),
		Base:      mp.mockBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/456", mp.providerType, owner, repo),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// CreatePullRequest creates a new pull request
func (mp *MockProvider) CreatePullRequest(ctx context.Context, owner, repo string, pr *git.CreatePullRequestRequest) (*git.PullRequest, error) {
	if mp.shouldFail && mp.failMethod == "CreatePullRequest" {
		return nil, fmt.Errorf("mock pull request creation failed")
	}

	return &git.PullRequest{
		Number:    999,
		Title:     pr.Title,
		Body:      pr.Body,
		State:     "open",
		Merged:    false,
		Author:    mp.mockUser(1, "currentuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mp.mockBranch(pr.Head, "abc123"),
		Base:      mp.mockBranch(pr.Base, "def456"),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 0},
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/999", mp.providerType, owner, repo),
		Draft:     pr.Draft,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// UpdatePullRequest updates an existing pull request
func (mp *MockProvider) UpdatePullRequest(ctx context.Context, owner, repo string, prNumber int, updates *git.UpdatePullRequestRequest) (*git.PullRequest, error) {
	if mp.shouldFail && mp.failMethod == "UpdatePullRequest" {
		return nil, fmt.Errorf("mock pull request update failed")
	}

	title := fmt.Sprintf("Updated Mock PR %d", prNumber)
	if updates.Title != nil {
		title = *updates.Title
	}

	return &git.PullRequest{
		Number:    prNumber,
		Title:     title,
		Body:      "Updated mock pull request body",
		State:     "open",
		Merged:    false,
		Author:    mp.mockUser(1, "testuser"),
		Assignees: []*git.User{},
		Labels:    []*git.Label{},
		Head:      mp.mockBranch("feature-branch", "abc123"),
		Base:      mp.mockBranch("main", "def456"),
		CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Now().UTC(),
		Comments:  git.Comment{ID: 1},
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/%d", mp.providerType, owner, repo, prNumber),
		Draft:     false,
		Mergeable: &[]bool{true}[0],
	}, nil
}

// GetPullRequestReviews retrieves reviews for a pull request
func (mp *MockProvider) GetPullRequestReviews(ctx context.Context, owner, repo string, prNumber int) ([]*git.Review, error) {
	if mp.shouldFail && mp.failMethod == "GetPullRequestReviews" {
		return nil, fmt.Errorf("mock pull request reviews retrieval failed")
	}

	reviews := []*git.Review{
		{
			ID:          1,
			User:        mp.mockUser(1, "reviewer1"),
			Body:        "This looks good!",
			State:       "approved",
			SubmittedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			CommitID:    "abc123",
			URL:         fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/%d/reviews/1", mp.providerType, owner, repo, prNumber),
		},
		{
			ID:          2,
			User:        mp.mockUser(2, "reviewer2"),
			Body:        "Please fix the formatting",
			State:       "changes_requested",
			SubmittedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			CommitID:    "def456",
			URL:         fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/%d/reviews/2", mp.providerType, owner, repo, prNumber),
		},
	}

	return reviews, nil
}

// GetPullRequestComments retrieves comments for a pull request
func (mp *MockProvider) GetPullRequestComments(ctx context.Context, owner, repo string, prNumber int) ([]*git.Comment, error) {
	if mp.shouldFail && mp.failMethod == "GetPullRequestComments" {
		return nil, fmt.Errorf("mock pull request comments retrieval failed")
	}

	comments := []*git.Comment{
		{
			ID:        1,
			User:      mp.mockUser(1, "commenter1"),
			Body:      "Great work on this PR!",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/%d/comments/1", mp.providerType, owner, repo, prNumber),
		},
		{
			ID:        2,
			User:      mp.mockUser(2, "commenter2"),
			Body:      "Can you add more tests?",
			CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/pulls/%d/comments/2", mp.providerType, owner, repo, prNumber),
		},
	}

	return comments, nil
}

// GetIssueComments retrieves comments for an issue
func (mp *MockProvider) GetIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*git.Comment, error) {
	if mp.shouldFail && mp.failMethod == "GetIssueComments" {
		return nil, fmt.Errorf("mock issue comments retrieval failed")
	}

	comments := []*git.Comment{
		{
			ID:        1,
			User:      mp.mockUser(1, "commenter1"),
			Body:      "This issue is important",
			CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/issues/%d/comments/1", mp.providerType, owner, repo, issueNumber),
		},
		{
			ID:        2,
			User:      mp.mockUser(2, "commenter2"),
			Body:      "I can help with this",
			CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/issues/%d/comments/2", mp.providerType, owner, repo, issueNumber),
		},
	}

	return comments, nil
}

// CreateComment creates a comment on an issue or pull request
func (mp *MockProvider) CreateComment(ctx context.Context, owner, repo string, issueNumber int, comment *git.CreateCommentRequest) (*git.Comment, error) {
	if mp.shouldFail && mp.failMethod == "CreateComment" {
		return nil, fmt.Errorf("mock comment creation failed")
	}

	return &git.Comment{
		ID:        999,
		User:      mp.mockUser(1, "currentuser"),
		Body:      comment.Body,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		URL:       fmt.Sprintf("https://api.%s.com/repos/%s/%s/issues/%d/comments/999", mp.providerType, owner, repo, issueNumber),
	}, nil
}

// GetLabels retrieves available labels for a repository
func (mp *MockProvider) GetLabels(ctx context.Context, owner, repo string) ([]*git.Label, error) {
	if mp.shouldFail && mp.failMethod == "GetLabels" {
		return nil, fmt.Errorf("mock labels retrieval failed")
	}

	labels := []*git.Label{
		{
			ID:          1,
			Name:        "bug",
			Description: "Something isn't working",
			URL:         fmt.Sprintf("https://api.%s.com/repos/%s/%s/labels/bug", mp.providerType, owner, repo),
		},
		{
			ID:          2,
			Name:        "enhancement",
			Description: "New feature or request",
			URL:         fmt.Sprintf("https://api.%s.com/repos/%s/%s/labels/enhancement", mp.providerType, owner, repo),
		},
		{
			ID:          3,
			Name:        "documentation",
			Description: "Improvements or additions to documentation",
			URL:         fmt.Sprintf("https://api.%s.com/repos/%s/%s/labels/documentation", mp.providerType, owner, repo),
		},
		{
			ID:          4,
			Name:        "feature",
			Description: "New feature",
			URL:         fmt.Sprintf("https://api.%s.com/repos/%s/%s/labels/feature", mp.providerType, owner, repo),
		},
	}

	return labels, nil
}

// Helper methods to create mock data

func (mp *MockProvider) mockUser(id int, login string) *git.User {
	return &git.User{
		ID:    id,
		Login: login,
		Name:  fmt.Sprintf("Test User %d", id),
		Email: fmt.Sprintf("%s@example.com", login),
		Type:  "User",
	}
}

func (mp *MockProvider) mockLabel(id int, name string) *git.Label {
	return &git.Label{
		ID:          id,
		Name:        name,
		Description: fmt.Sprintf("Mock label: %s", name),
		URL:         fmt.Sprintf("https://api.%s.com/repos/test/owner/labels/%s", mp.providerType, name),
	}
}

func (mp *MockProvider) mockBranch(ref, sha string) *git.Branch {
	return &git.Branch{
		Ref:  ref,
		SHA:  sha,
		Repo: &git.Repository{Owner: "test", Name: "repo"},
		User: mp.mockUser(1, "testuser"),
	}
}
