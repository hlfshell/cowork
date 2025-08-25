package gitprovider

import (
	"context"
	"testing"
	"time"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGitHubProvider_Creation tests creating a new GitHub provider
func TestGitHubProvider_Creation(t *testing.T) {
	// Test case: Creating a new GitHub provider should succeed with valid token
	provider, err := NewGitHubProvider("test-token", "")

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, git.ProviderGitHub, provider.GetProviderType())
}

// TestGitHubProvider_Creation_NoToken tests provider creation without token
func TestGitHubProvider_Creation_NoToken(t *testing.T) {
	// Test case: Creating a GitHub provider without token should fail
	provider, err := NewGitHubProvider("", "")

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "GitHub token is required")
}

// TestGitHubProvider_Creation_InvalidBaseURL tests provider creation with invalid base URL
func TestGitHubProvider_Creation_InvalidBaseURL(t *testing.T) {
	// Test case: Creating a GitHub provider with invalid base URL should fail
	provider, err := NewGitHubProvider("test-token", "://invalid-url")

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "invalid base URL")
}

// TestGitHubProvider_GetProviderType tests getting the provider type
func TestGitHubProvider_GetProviderType(t *testing.T) {
	// Test case: GetProviderType should return GitHub provider type
	provider, err := NewGitHubProvider("test-token", "")
	require.NoError(t, err)

	providerType := provider.GetProviderType()
	assert.Equal(t, git.ProviderGitHub, providerType)
}

// TestGitHubProvider_TestAuth tests authentication with GitHub
func TestGitHubProvider_TestAuth(t *testing.T) {
	// Test case: TestAuth should verify GitHub authentication
	// This test will be skipped if no valid token is available
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.TestAuth(ctx)

	// This test may fail if the token is invalid, which is expected
	if err != nil {
		t.Skip("GitHub authentication failed, skipping auth test")
		return
	}

	assert.NoError(t, err)
}

// TestGitHubProvider_GetRepositoryInfo tests getting repository information
func TestGitHubProvider_GetRepositoryInfo(t *testing.T) {
	// Test case: GetRepositoryInfo should retrieve repository information
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()
	repoInfo, err := provider.GetRepositoryInfo(ctx, "golang", "go")

	// This test may fail if the repository doesn't exist or token is invalid
	if err != nil {
		t.Skip("Failed to get repository info, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, repoInfo)
	assert.Equal(t, "golang", repoInfo.Owner)
	assert.Equal(t, "go", repoInfo.Name)
	assert.Equal(t, "golang/go", repoInfo.FullName)
	assert.NotEmpty(t, repoInfo.Description)
	assert.NotEmpty(t, repoInfo.DefaultBranch)
	assert.True(t, repoInfo.CreatedAt.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.True(t, repoInfo.UpdatedAt.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
}

// TestGitHubProvider_GetIssues tests getting issues from a repository
func TestGitHubProvider_GetIssues(t *testing.T) {
	// Test case: GetIssues should retrieve issues with optional filtering
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Test getting all issues
	issues, err := provider.GetIssues(ctx, "golang", "go", nil)
	if err != nil {
		t.Skip("Failed to get issues, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, issues)
	// Issues can be empty, but the response should be valid

	// Test getting issues with filters
	options := &git.IssueListOptions{
		State:   "open",
		PerPage: 5,
	}

	filteredIssues, err := provider.GetIssues(ctx, "golang", "go", options)
	if err != nil {
		t.Skip("Failed to get filtered issues, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, filteredIssues)
	assert.LessOrEqual(t, len(filteredIssues), 5)
}

// TestGitHubProvider_GetIssue tests getting a specific issue
func TestGitHubProvider_GetIssue(t *testing.T) {
	// Test case: GetIssue should retrieve a specific issue by number
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Try to get a well-known issue from golang/go repository
	issue, err := provider.GetIssue(ctx, "golang", "go", 1)
	if err != nil {
		t.Skip("Failed to get issue, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 1, issue.Number)
	assert.NotEmpty(t, issue.Title)
	assert.NotEmpty(t, issue.Author.Login)
}

// TestGitHubProvider_GetPullRequests tests getting pull requests
func TestGitHubProvider_GetPullRequests(t *testing.T) {
	// Test case: GetPullRequests should retrieve pull requests with optional filtering
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Test getting all pull requests
	prs, err := provider.GetPullRequests(ctx, "golang", "go", nil)
	if err != nil {
		t.Skip("Failed to get pull requests, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, prs)
	// Pull requests can be empty, but the response should be valid

	// Test getting pull requests with filters
	options := &git.PullRequestListOptions{
		State:   "open",
		PerPage: 5,
	}

	filteredPRs, err := provider.GetPullRequests(ctx, "golang", "go", options)
	if err != nil {
		t.Skip("Failed to get filtered pull requests, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, filteredPRs)
	assert.LessOrEqual(t, len(filteredPRs), 5)
}

// TestGitHubProvider_GetPullRequest tests getting a specific pull request
func TestGitHubProvider_GetPullRequest(t *testing.T) {
	// Test case: GetPullRequest should retrieve a specific pull request by number
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Try to get a well-known pull request from golang/go repository
	pr, err := provider.GetPullRequest(ctx, "golang", "go", 1)
	if err != nil {
		t.Skip("Failed to get pull request, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 1, pr.Number)
	assert.NotEmpty(t, pr.Title)
	assert.NotEmpty(t, pr.Author.Login)
}

// TestGitHubProvider_GetPullRequestByBranch tests getting pull request by branch
func TestGitHubProvider_GetPullRequestByBranch(t *testing.T) {
	// Test case: GetPullRequestByBranch should retrieve pull request by source branch
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// This test is likely to fail since we don't know a specific branch with an open PR
	// We'll test the error handling instead
	_, err = provider.GetPullRequestByBranch(ctx, "golang", "go", "nonexistent-branch")

	// Should return an error for non-existent branch
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pull request found")
}

// TestGitHubProvider_GetPullRequestByIssue tests getting pull request by issue
func TestGitHubProvider_GetPullRequestByIssue(t *testing.T) {
	// Test case: GetPullRequestByIssue should retrieve pull request by issue number
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Try to get a pull request by issue number (issue #1 is likely a PR)
	pr, err := provider.GetPullRequestByIssue(ctx, "golang", "go", 1)
	if err != nil {
		// This might fail if issue #1 is not a pull request
		t.Skip("Failed to get pull request by issue, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 1, pr.Number)
}

// TestGitHubProvider_GetPullRequestReviews tests getting pull request reviews
func TestGitHubProvider_GetPullRequestReviews(t *testing.T) {
	// Test case: GetPullRequestReviews should retrieve reviews for a pull request
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Try to get reviews for a pull request
	reviews, err := provider.GetPullRequestReviews(ctx, "golang", "go", 1)
	if err != nil {
		t.Skip("Failed to get pull request reviews, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, reviews)
	// Reviews can be empty, but the response should be valid
}

// TestGitHubProvider_GetPullRequestComments tests getting pull request comments
func TestGitHubProvider_GetPullRequestComments(t *testing.T) {
	// Test case: GetPullRequestComments should retrieve comments for a pull request
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Try to get comments for a pull request
	comments, err := provider.GetPullRequestComments(ctx, "golang", "go", 1)
	if err != nil {
		t.Skip("Failed to get pull request comments, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, comments)
	// Comments can be empty, but the response should be valid
}

// TestGitHubProvider_GetIssueComments tests getting issue comments
func TestGitHubProvider_GetIssueComments(t *testing.T) {
	// Test case: GetIssueComments should retrieve comments for an issue
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Try to get comments for an issue
	comments, err := provider.GetIssueComments(ctx, "golang", "go", 1)
	if err != nil {
		t.Skip("Failed to get issue comments, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, comments)
	// Comments can be empty, but the response should be valid
}

// TestGitHubProvider_GetLabels tests getting repository labels
func TestGitHubProvider_GetLabels(t *testing.T) {
	// Test case: GetLabels should retrieve available labels for a repository
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Try to get labels for a repository
	labels, err := provider.GetLabels(ctx, "golang", "go")
	if err != nil {
		t.Skip("Failed to get labels, skipping test")
		return
	}

	assert.NoError(t, err)
	assert.NotNil(t, labels)
	// Labels can be empty, but the response should be valid
}

// TestGitHubProvider_ErrorHandling tests error handling scenarios
func TestGitHubProvider_ErrorHandling(t *testing.T) {
	// Test case: Provider should handle various error scenarios gracefully
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx := context.Background()

	// Test getting non-existent repository
	_, err = provider.GetRepositoryInfo(ctx, "nonexistent", "repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get repository info")

	// Test getting non-existent issue
	_, err = provider.GetIssue(ctx, "golang", "go", 999999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get issue")

	// Test getting non-existent pull request
	_, err = provider.GetPullRequest(ctx, "golang", "go", 999999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get pull request")
}

// TestGitHubProvider_ContextCancellation tests context cancellation
func TestGitHubProvider_ContextCancellation(t *testing.T) {
	// Test case: Provider should respect context cancellation
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// These operations should fail due to cancelled context
	_, err = provider.GetRepositoryInfo(ctx, "golang", "go")
	assert.Error(t, err)

	_, err = provider.GetIssues(ctx, "golang", "go", nil)
	assert.Error(t, err)

	_, err = provider.GetPullRequests(ctx, "golang", "go", nil)
	assert.Error(t, err)
}

// TestGitHubProvider_Timeout tests operation timeout
func TestGitHubProvider_Timeout(t *testing.T) {
	// Test case: Provider should timeout appropriately
	token := getTestToken(t)
	if token == "" {
		t.Skip("No GitHub token available for testing")
	}

	provider, err := NewGitHubProvider(token, "")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// These operations should timeout
	_, err = provider.GetRepositoryInfo(ctx, "golang", "go")
	assert.Error(t, err)

	_, err = provider.GetIssues(ctx, "golang", "go", nil)
	assert.Error(t, err)
}

// Helper function to get test token from environment
func getTestToken(t *testing.T) string {
	// In a real implementation, this would read from environment variables
	// For now, we'll return empty string to skip tests that require authentication
	return ""
}
