package cowork

import (
	"context"
	"testing"
	"time"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGitProvider is a mock implementation of GitProvider for testing
type MockGitProvider struct {
	mock.Mock
}

func (m *MockGitProvider) GetProviderType() git.ProviderType {
	args := m.Called()
	return args.Get(0).(git.ProviderType)
}

func (m *MockGitProvider) TestAuth(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGitProvider) GetRepositoryInfo(ctx context.Context, owner, repo string) (*git.Repository, error) {
	args := m.Called(ctx, owner, repo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.Repository), args.Error(1)
}

func (m *MockGitProvider) GetIssues(ctx context.Context, owner, repo string, options *git.IssueListOptions) ([]*git.Issue, error) {
	args := m.Called(ctx, owner, repo, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*git.Issue), args.Error(1)
}

func (m *MockGitProvider) GetIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.Issue, error) {
	args := m.Called(ctx, owner, repo, issueNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.Issue), args.Error(1)
}

func (m *MockGitProvider) CreateIssue(ctx context.Context, owner, repo string, issue *git.CreateIssueRequest) (*git.Issue, error) {
	args := m.Called(ctx, owner, repo, issue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.Issue), args.Error(1)
}

func (m *MockGitProvider) UpdateIssue(ctx context.Context, owner, repo string, issueNumber int, updates *git.UpdateIssueRequest) (*git.Issue, error) {
	args := m.Called(ctx, owner, repo, issueNumber, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.Issue), args.Error(1)
}

func (m *MockGitProvider) GetPullRequests(ctx context.Context, owner, repo string, options *git.PullRequestListOptions) ([]*git.PullRequest, error) {
	args := m.Called(ctx, owner, repo, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*git.PullRequest), args.Error(1)
}

func (m *MockGitProvider) GetPullRequest(ctx context.Context, owner, repo string, prNumber int) (*git.PullRequest, error) {
	args := m.Called(ctx, owner, repo, prNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.PullRequest), args.Error(1)
}

func (m *MockGitProvider) GetPullRequestByBranch(ctx context.Context, owner, repo string, branchName string) (*git.PullRequest, error) {
	args := m.Called(ctx, owner, repo, branchName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.PullRequest), args.Error(1)
}

func (m *MockGitProvider) GetPullRequestByIssue(ctx context.Context, owner, repo string, issueNumber int) (*git.PullRequest, error) {
	args := m.Called(ctx, owner, repo, issueNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.PullRequest), args.Error(1)
}

func (m *MockGitProvider) CreatePullRequest(ctx context.Context, owner, repo string, pr *git.CreatePullRequestRequest) (*git.PullRequest, error) {
	args := m.Called(ctx, owner, repo, pr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.PullRequest), args.Error(1)
}

func (m *MockGitProvider) UpdatePullRequest(ctx context.Context, owner, repo string, prNumber int, updates *git.UpdatePullRequestRequest) (*git.PullRequest, error) {
	args := m.Called(ctx, owner, repo, prNumber, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.PullRequest), args.Error(1)
}

func (m *MockGitProvider) GetPullRequestReviews(ctx context.Context, owner, repo string, prNumber int) ([]*git.Review, error) {
	args := m.Called(ctx, owner, repo, prNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*git.Review), args.Error(1)
}

func (m *MockGitProvider) GetPullRequestComments(ctx context.Context, owner, repo string, prNumber int) ([]*git.Comment, error) {
	args := m.Called(ctx, owner, repo, prNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*git.Comment), args.Error(1)
}

func (m *MockGitProvider) GetIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*git.Comment, error) {
	args := m.Called(ctx, owner, repo, issueNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*git.Comment), args.Error(1)
}

func (m *MockGitProvider) CreateComment(ctx context.Context, owner, repo string, issueNumber int, comment *git.CreateCommentRequest) (*git.Comment, error) {
	args := m.Called(ctx, owner, repo, issueNumber, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.Comment), args.Error(1)
}

func (m *MockGitProvider) GetLabels(ctx context.Context, owner, repo string) ([]*git.Label, error) {
	args := m.Called(ctx, owner, repo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*git.Label), args.Error(1)
}

// TestHandler_GenerateBranchName tests branch name generation
func TestHandler_GenerateBranchName(t *testing.T) {
	// Test case: Generate branch name from issue
	handler := &Handler{}
	
	issue := &git.Issue{
		Number: 123,
		Title:  "Fix authentication bug in login flow",
		Author: &git.User{Login: "testuser"},
		Labels: []*git.Label{
			{Name: "bug"},
			{Name: "high-priority"},
		},
		CreatedAt: time.Now(),
		URL:       "https://github.com/test/repo/issues/123",
	}

	branchName := handler.GenerateBranchName(issue)
	
	// Should generate a human-readable branch name
	assert.Contains(t, branchName, "fix-authentication-bug")
	assert.Contains(t, branchName, "123")
	assert.True(t, len(branchName) <= 50) // Reasonable length
}

// TestHandler_GenerateBranchName_SpecialCharacters tests branch name generation with special characters
func TestHandler_GenerateBranchName_SpecialCharacters(t *testing.T) {
	handler := &Handler{}
	
	issue := &git.Issue{
		Number: 456,
		Title:  "Update API endpoints (v2.0) - breaking changes!",
		Author: &git.User{Login: "testuser"},
		Labels: []*git.Label{},
		CreatedAt: time.Now(),
		URL:       "https://github.com/test/repo/issues/456",
	}

	branchName := handler.GenerateBranchName(issue)
	
	// Should handle special characters properly
	assert.Contains(t, branchName, "update-api-endpoints")
	assert.Contains(t, branchName, "456")
	assert.NotContains(t, branchName, "!")
	assert.NotContains(t, branchName, "(")
	assert.NotContains(t, branchName, ")")
}

// TestHandler_GenerateBranchName_EmptyTitle tests branch name generation with empty title
func TestHandler_GenerateBranchName_EmptyTitle(t *testing.T) {
	handler := &Handler{}
	
	issue := &git.Issue{
		Number: 789,
		Title:  "",
		Author: &git.User{Login: "testuser"},
		Labels: []*git.Label{},
		CreatedAt: time.Now(),
		URL:       "https://github.com/test/repo/issues/789",
	}

	branchName := handler.GenerateBranchName(issue)
	
	// Should default to "task" for empty title
	assert.Contains(t, branchName, "task")
	assert.Contains(t, branchName, "789")
}

// TestHandler_ExtractIssueNumber tests issue number extraction from ticket ID
func TestHandler_ExtractIssueNumber(t *testing.T) {
	handler := &Handler{}
	
	testCases := []struct {
		ticketID     string
		expectedNum  int
		description  string
	}{
		{"github:owner/repo#123", 123, "Valid GitHub ticket ID"},
		{"gitlab:owner/repo#456", 456, "Valid GitLab ticket ID"},
		{"bitbucket:owner/repo#789", 789, "Valid Bitbucket ticket ID"},
		{"invalid-format", 0, "Invalid format"},
		{"github:owner/repo#", 0, "Missing number"},
		{"", 0, "Empty string"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := handler.extractIssueNumber(tc.ticketID)
			assert.Equal(t, tc.expectedNum, result)
		})
	}
}

// TestHandler_GetProviderHost tests provider host resolution
func TestHandler_GetProviderHost(t *testing.T) {
	testCases := []struct {
		providerType git.ProviderType
		expectedHost string
		description  string
	}{
		{git.ProviderGitHub, "github.com", "GitHub provider"},
		{git.ProviderGitLab, "gitlab.com", "GitLab provider"},
		{git.ProviderBitbucket, "bitbucket.org", "Bitbucket provider"},
		{"unknown", "github.com", "Unknown provider defaults to GitHub"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			mockProvider := &MockGitProvider{}
			mockProvider.On("GetProviderType").Return(tc.providerType)
			
			handler := &Handler{provider: mockProvider}
			result := handler.getProviderHost()
			
			assert.Equal(t, tc.expectedHost, result)
		})
	}
}
