package gitprovider

import (
	"context"
	"testing"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestMockGitHubProvider_GetProviderType(t *testing.T) {
	provider := NewMockGitHubProvider()
	assert.Equal(t, git.ProviderGitHub, provider.GetProviderType())
}

func TestMockGitHubProvider_TestAuth_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.NoError(t, err)
}

func TestMockGitHubProvider_TestAuth_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("TestAuth")
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub authentication failed: 401 Unauthorized")
}

func TestMockGitHubProvider_TestAuth_RateLimit(t *testing.T) {
	provider := NewMockGitHubProviderWithRateLimit()
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub API rate limit exceeded: 403 Forbidden")
}

func TestMockGitHubProvider_GetRepositoryInfo_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	repo, err := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "testowner", repo.Owner)
	assert.Equal(t, "testrepo", repo.Name)
	assert.Equal(t, "testowner/testrepo", repo.FullName)
	assert.Equal(t, "main", repo.DefaultBranch)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo", repo.URL)
	assert.False(t, repo.Private)
}

func TestMockGitHubProvider_GetRepositoryInfo_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetRepositoryInfo")
	ctx := context.Background()

	repo, err := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "GitHub repository not found: 404 Not Found")
}

func TestMockGitHubProvider_GetIssues_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()
	options := &git.IssueListOptions{State: "open"}

	issues, err := provider.GetIssues(ctx, "testowner", "testrepo", options)
	assert.NoError(t, err)
	assert.Len(t, issues, 2)

	// Check first issue
	issue1 := issues[0]
	assert.Equal(t, 1, issue1.Number)
	assert.Equal(t, "GitHub Mock Issue 1", issue1.Title)
	assert.Equal(t, "open", issue1.State)
	assert.NotNil(t, issue1.Author)
	assert.Equal(t, "github-user1", issue1.Author.Login)
	assert.Len(t, issue1.Assignees, 1)
	assert.Len(t, issue1.Labels, 2)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/issues/1", issue1.URL)

	// Check second issue
	issue2 := issues[1]
	assert.Equal(t, 2, issue2.Number)
	assert.Equal(t, "GitHub Mock Issue 2", issue2.Title)
	assert.Equal(t, "closed", issue2.State)
	assert.NotNil(t, issue2.ClosedAt)
}

func TestMockGitHubProvider_GetIssues_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetIssues")
	ctx := context.Background()
	options := &git.IssueListOptions{State: "open"}

	issues, err := provider.GetIssues(ctx, "testowner", "testrepo", options)
	assert.Error(t, err)
	assert.Nil(t, issues)
	assert.Contains(t, err.Error(), "GitHub issues retrieval failed: 500 Internal Server Error")
}

func TestMockGitHubProvider_GetIssue_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	issue, err := provider.GetIssue(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 123, issue.Number)
	assert.Equal(t, "GitHub Mock Issue 123", issue.Title)
	assert.Equal(t, "open", issue.State)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/issues/123", issue.URL)
}

func TestMockGitHubProvider_GetIssue_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetIssue")
	ctx := context.Background()

	issue, err := provider.GetIssue(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "GitHub issue not found: 404 Not Found")
}

func TestMockGitHubProvider_CreateIssue_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()
	request := &git.CreateIssueRequest{
		Title:     "Test Issue",
		Body:      "Test issue body",
		Assignees: []string{"user1", "user2"},
		Labels:    []string{"bug", "enhancement"},
	}

	issue, err := provider.CreateIssue(ctx, "testowner", "testrepo", request)
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 999, issue.Number)
	assert.Equal(t, "Test Issue", issue.Title)
	assert.Equal(t, "Test issue body", issue.Body)
	assert.Equal(t, "open", issue.State)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/issues/999", issue.URL)
}

func TestMockGitHubProvider_CreateIssue_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("CreateIssue")
	ctx := context.Background()
	request := &git.CreateIssueRequest{
		Title: "Test Issue",
		Body:  "Test issue body",
	}

	issue, err := provider.CreateIssue(ctx, "testowner", "testrepo", request)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "GitHub issue creation failed: 422 Unprocessable Entity")
}

func TestMockGitHubProvider_UpdateIssue_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()
	updates := &git.UpdateIssueRequest{
		Title: stringPtr("Updated Title"),
		Body:  stringPtr("Updated body"),
		State: stringPtr("closed"),
	}

	issue, err := provider.UpdateIssue(ctx, "testowner", "testrepo", 123, updates)
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 123, issue.Number)
	assert.Equal(t, "Updated Title", issue.Title)
	assert.Equal(t, "open", issue.State) // Mock returns "open" regardless of update
}

func TestMockGitHubProvider_UpdateIssue_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("UpdateIssue")
	ctx := context.Background()
	updates := &git.UpdateIssueRequest{
		Title: stringPtr("Updated Title"),
	}

	issue, err := provider.UpdateIssue(ctx, "testowner", "testrepo", 123, updates)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "GitHub issue update failed: 404 Not Found")
}

func TestMockGitHubProvider_GetPullRequests_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()
	options := &git.PullRequestListOptions{State: "open"}

	prs, err := provider.GetPullRequests(ctx, "testowner", "testrepo", options)
	assert.NoError(t, err)
	assert.Len(t, prs, 2)

	// Check first PR
	pr1 := prs[0]
	assert.Equal(t, 1, pr1.Number)
	assert.Equal(t, "GitHub Mock Pull Request 1", pr1.Title)
	assert.Equal(t, "open", pr1.State)
	assert.False(t, pr1.Merged)
	assert.NotNil(t, pr1.Head)
	assert.Equal(t, "feature-branch", pr1.Head.Ref)
	assert.NotNil(t, pr1.Base)
	assert.Equal(t, "main", pr1.Base.Ref)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/pulls/1", pr1.URL)

	// Check second PR
	pr2 := prs[1]
	assert.Equal(t, 2, pr2.Number)
	assert.Equal(t, "GitHub Mock Pull Request 2", pr2.Title)
	assert.Equal(t, "closed", pr2.State)
	assert.True(t, pr2.Merged)
	assert.NotNil(t, pr2.MergedAt)
}

func TestMockGitHubProvider_GetPullRequests_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetPullRequests")
	ctx := context.Background()
	options := &git.PullRequestListOptions{State: "open"}

	prs, err := provider.GetPullRequests(ctx, "testowner", "testrepo", options)
	assert.Error(t, err)
	assert.Nil(t, prs)
	assert.Contains(t, err.Error(), "GitHub pull requests retrieval failed: 500 Internal Server Error")
}

func TestMockGitHubProvider_GetPullRequest_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequest(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "GitHub Mock Pull Request 123", pr.Title)
	assert.Equal(t, "open", pr.State)
	assert.False(t, pr.Merged)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/pulls/123", pr.URL)
}

func TestMockGitHubProvider_GetPullRequest_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetPullRequest")
	ctx := context.Background()

	pr, err := provider.GetPullRequest(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitHub pull request not found: 404 Not Found")
}

func TestMockGitHubProvider_GetPullRequestByBranch_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequestByBranch(ctx, "testowner", "testrepo", "feature-branch")
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "GitHub Mock PR for branch feature-branch", pr.Title)
	assert.Equal(t, "open", pr.State)
	assert.Equal(t, "feature-branch", pr.Head.Ref)
	assert.Equal(t, "main", pr.Base.Ref)
}

func TestMockGitHubProvider_GetPullRequestByBranch_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetPullRequestByBranch")
	ctx := context.Background()

	pr, err := provider.GetPullRequestByBranch(ctx, "testowner", "testrepo", "feature-branch")
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitHub pull request not found for branch: 404 Not Found")
}

func TestMockGitHubProvider_GetPullRequestByIssue_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequestByIssue(ctx, "testowner", "testrepo", 456)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 456, pr.Number)
	assert.Equal(t, "GitHub Mock PR closing issue 456", pr.Title)
	assert.Equal(t, "open", pr.State)
	assert.Equal(t, "fix-issue", pr.Head.Ref)
}

func TestMockGitHubProvider_GetPullRequestByIssue_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetPullRequestByIssue")
	ctx := context.Background()

	pr, err := provider.GetPullRequestByIssue(ctx, "testowner", "testrepo", 456)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitHub pull request not found for issue: 404 Not Found")
}

func TestMockGitHubProvider_CreatePullRequest_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()
	request := &git.CreatePullRequestRequest{
		Title: "Test PR",
		Body:  "Test PR body",
		Head:  "feature-branch",
		Base:  "main",
		Draft: false,
	}

	pr, err := provider.CreatePullRequest(ctx, "testowner", "testrepo", request)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 999, pr.Number)
	assert.Equal(t, "Test PR", pr.Title)
	assert.Equal(t, "Test PR body", pr.Body)
	assert.Equal(t, "open", pr.State)
	assert.False(t, pr.Draft)
	assert.Equal(t, "feature-branch", pr.Head.Ref)
	assert.Equal(t, "main", pr.Base.Ref)
}

func TestMockGitHubProvider_CreatePullRequest_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("CreatePullRequest")
	ctx := context.Background()
	request := &git.CreatePullRequestRequest{
		Title: "Test PR",
		Head:  "feature-branch",
		Base:  "main",
	}

	pr, err := provider.CreatePullRequest(ctx, "testowner", "testrepo", request)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitHub pull request creation failed: 422 Unprocessable Entity")
}

func TestMockGitHubProvider_UpdatePullRequest_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()
	updates := &git.UpdatePullRequestRequest{
		Title: stringPtr("Updated PR Title"),
		Body:  stringPtr("Updated PR body"),
	}

	pr, err := provider.UpdatePullRequest(ctx, "testowner", "testrepo", 123, updates)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "Updated PR Title", pr.Title)
	assert.Equal(t, "open", pr.State)
}

func TestMockGitHubProvider_UpdatePullRequest_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("UpdatePullRequest")
	ctx := context.Background()
	updates := &git.UpdatePullRequestRequest{
		Title: stringPtr("Updated PR Title"),
	}

	pr, err := provider.UpdatePullRequest(ctx, "testowner", "testrepo", 123, updates)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitHub pull request update failed: 404 Not Found")
}

func TestMockGitHubProvider_GetPullRequestReviews_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	reviews, err := provider.GetPullRequestReviews(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, reviews, 2)

	// Check first review
	review1 := reviews[0]
	assert.Equal(t, 1, review1.ID)
	assert.Equal(t, "approved", review1.State)
	assert.Equal(t, "This looks good! 👍", review1.Body)
	assert.Equal(t, "github-reviewer1", review1.User.Login)
	assert.Equal(t, "abc123", review1.CommitID)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/pulls/123/reviews/1", review1.URL)

	// Check second review
	review2 := reviews[1]
	assert.Equal(t, 2, review2.ID)
	assert.Equal(t, "changes_requested", review2.State)
	assert.Equal(t, "Please fix the formatting", review2.Body)
	assert.Equal(t, "github-reviewer2", review2.User.Login)
}

func TestMockGitHubProvider_GetPullRequestReviews_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetPullRequestReviews")
	ctx := context.Background()

	reviews, err := provider.GetPullRequestReviews(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, reviews)
	assert.Contains(t, err.Error(), "GitHub pull request reviews retrieval failed: 404 Not Found")
}

func TestMockGitHubProvider_GetPullRequestComments_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	comments, err := provider.GetPullRequestComments(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	// Check first comment
	comment1 := comments[0]
	assert.Equal(t, 1, comment1.ID)
	assert.Equal(t, "Great work on this PR! 🎉", comment1.Body)
	assert.Equal(t, "github-commenter1", comment1.User.Login)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/pulls/123/comments/1", comment1.URL)

	// Check second comment
	comment2 := comments[1]
	assert.Equal(t, 2, comment2.ID)
	assert.Equal(t, "Can you add more tests?", comment2.Body)
	assert.Equal(t, "github-commenter2", comment2.User.Login)
}

func TestMockGitHubProvider_GetPullRequestComments_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetPullRequestComments")
	ctx := context.Background()

	comments, err := provider.GetPullRequestComments(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Contains(t, err.Error(), "GitHub pull request comments retrieval failed: 404 Not Found")
}

func TestMockGitHubProvider_GetIssueComments_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	comments, err := provider.GetIssueComments(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	// Check first comment
	comment1 := comments[0]
	assert.Equal(t, 1, comment1.ID)
	assert.Equal(t, "This issue is important", comment1.Body)
	assert.Equal(t, "github-commenter1", comment1.User.Login)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/issues/123/comments/1", comment1.URL)

	// Check second comment
	comment2 := comments[1]
	assert.Equal(t, 2, comment2.ID)
	assert.Equal(t, "I can help with this", comment2.Body)
	assert.Equal(t, "github-commenter2", comment2.User.Login)
}

func TestMockGitHubProvider_GetIssueComments_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetIssueComments")
	ctx := context.Background()

	comments, err := provider.GetIssueComments(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Contains(t, err.Error(), "GitHub issue comments retrieval failed: 404 Not Found")
}

func TestMockGitHubProvider_CreateComment_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()
	request := &git.CreateCommentRequest{
		Body: "Test comment body",
	}

	comment, err := provider.CreateComment(ctx, "testowner", "testrepo", 123, request)
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, 999, comment.ID)
	assert.Equal(t, "Test comment body", comment.Body)
	assert.Equal(t, "github-currentuser", comment.User.Login)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/issues/123/comments/999", comment.URL)
}

func TestMockGitHubProvider_CreateComment_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("CreateComment")
	ctx := context.Background()
	request := &git.CreateCommentRequest{
		Body: "Test comment body",
	}

	comment, err := provider.CreateComment(ctx, "testowner", "testrepo", 123, request)
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Contains(t, err.Error(), "GitHub comment creation failed: 422 Unprocessable Entity")
}

func TestMockGitHubProvider_GetLabels_Success(t *testing.T) {
	provider := NewMockGitHubProvider()
	ctx := context.Background()

	labels, err := provider.GetLabels(ctx, "testowner", "testrepo")
	assert.NoError(t, err)
	assert.Len(t, labels, 5)

	// Check specific labels
	bugLabel := labels[0]
	assert.Equal(t, 1, bugLabel.ID)
	assert.Equal(t, "bug", bugLabel.Name)
	assert.Equal(t, "Something isn't working", bugLabel.Description)
	assert.Equal(t, "https://api.github.com/repos/testowner/testrepo/labels/bug", bugLabel.URL)

	enhancementLabel := labels[1]
	assert.Equal(t, 2, enhancementLabel.ID)
	assert.Equal(t, "enhancement", enhancementLabel.Name)
	assert.Equal(t, "New feature or request", enhancementLabel.Description)

	goodFirstIssueLabel := labels[3]
	assert.Equal(t, 4, goodFirstIssueLabel.ID)
	assert.Equal(t, "good first issue", goodFirstIssueLabel.Name)
	assert.Equal(t, "Good for newcomers", goodFirstIssueLabel.Description)
}

func TestMockGitHubProvider_GetLabels_Failure(t *testing.T) {
	provider := NewMockGitHubProviderWithFailure("GetLabels")
	ctx := context.Background()

	labels, err := provider.GetLabels(ctx, "testowner", "testrepo")
	assert.Error(t, err)
	assert.Nil(t, labels)
	assert.Contains(t, err.Error(), "GitHub labels retrieval failed: 404 Not Found")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
