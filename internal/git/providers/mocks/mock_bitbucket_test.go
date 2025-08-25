package gitprovider

import (
	"context"
	"testing"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestMockBitbucketProvider_GetProviderType(t *testing.T) {
	provider := NewMockBitbucketProvider()
	assert.Equal(t, git.ProviderBitbucket, provider.GetProviderType())
}

func TestMockBitbucketProvider_TestAuth_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.NoError(t, err)
}

func TestMockBitbucketProvider_TestAuth_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("TestAuth")
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Bitbucket authentication failed: 401 Unauthorized")
}

func TestMockBitbucketProvider_TestAuth_RateLimit(t *testing.T) {
	provider := NewMockBitbucketProviderWithRateLimit()
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Bitbucket API rate limit exceeded: 429 Too Many Requests")
}

func TestMockBitbucketProvider_GetRepositoryInfo_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	repo, err := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "testowner", repo.Owner)
	assert.Equal(t, "testrepo", repo.Name)
	assert.Equal(t, "testowner/testrepo", repo.FullName)
	assert.Equal(t, "main", repo.DefaultBranch) // Bitbucket uses "main" as default
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo", repo.URL)
	assert.False(t, repo.Private)
}

func TestMockBitbucketProvider_GetRepositoryInfo_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetRepositoryInfo")
	ctx := context.Background()

	repo, err := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "Bitbucket repository not found: 404 Not Found")
}

func TestMockBitbucketProvider_GetIssues_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()
	options := &git.IssueListOptions{State: "open"}

	issues, err := provider.GetIssues(ctx, "testowner", "testrepo", options)
	assert.NoError(t, err)
	assert.Len(t, issues, 2)

	// Check first issue
	issue1 := issues[0]
	assert.Equal(t, 1, issue1.Number)
	assert.Equal(t, "Bitbucket Mock Issue 1", issue1.Title)
	assert.Equal(t, "open", issue1.State) // Bitbucket uses "open"
	assert.NotNil(t, issue1.Author)
	assert.Equal(t, "bitbucket-user1", issue1.Author.Login)
	assert.Len(t, issue1.Assignees, 1)
	assert.Len(t, issue1.Labels, 2)
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/issues/1", issue1.URL)

	// Check second issue
	issue2 := issues[1]
	assert.Equal(t, 2, issue2.Number)
	assert.Equal(t, "Bitbucket Mock Issue 2", issue2.Title)
	assert.Equal(t, "resolved", issue2.State) // Bitbucket uses "resolved" instead of "closed"
	assert.NotNil(t, issue2.ClosedAt)
}

func TestMockBitbucketProvider_GetIssues_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetIssues")
	ctx := context.Background()
	options := &git.IssueListOptions{State: "open"}

	issues, err := provider.GetIssues(ctx, "testowner", "testrepo", options)
	assert.Error(t, err)
	assert.Nil(t, issues)
	assert.Contains(t, err.Error(), "Bitbucket issues retrieval failed: 500 Internal Server Error")
}

func TestMockBitbucketProvider_GetIssue_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	issue, err := provider.GetIssue(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 123, issue.Number)
	assert.Equal(t, "Bitbucket Mock Issue 123", issue.Title)
	assert.Equal(t, "open", issue.State) // Bitbucket uses "open"
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/issues/123", issue.URL)
}

func TestMockBitbucketProvider_GetIssue_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetIssue")
	ctx := context.Background()

	issue, err := provider.GetIssue(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "Bitbucket issue not found: 404 Not Found")
}

func TestMockBitbucketProvider_CreateIssue_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
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
	assert.Equal(t, "open", issue.State) // Bitbucket uses "open"
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/issues/999", issue.URL)
}

func TestMockBitbucketProvider_CreateIssue_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("CreateIssue")
	ctx := context.Background()
	request := &git.CreateIssueRequest{
		Title: "Test Issue",
		Body:  "Test issue body",
	}

	issue, err := provider.CreateIssue(ctx, "testowner", "testrepo", request)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "Bitbucket issue creation failed: 422 Unprocessable Entity")
}

func TestMockBitbucketProvider_UpdateIssue_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()
	updates := &git.UpdateIssueRequest{
		Title: stringPtr("Updated Title"),
		Body:  stringPtr("Updated body"),
		State: stringPtr("resolved"),
	}

	issue, err := provider.UpdateIssue(ctx, "testowner", "testrepo", 123, updates)
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 123, issue.Number)
	assert.Equal(t, "Updated Title", issue.Title)
	assert.Equal(t, "open", issue.State) // Mock returns "open" regardless of update
}

func TestMockBitbucketProvider_UpdateIssue_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("UpdateIssue")
	ctx := context.Background()
	updates := &git.UpdateIssueRequest{
		Title: stringPtr("Updated Title"),
	}

	issue, err := provider.UpdateIssue(ctx, "testowner", "testrepo", 123, updates)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "Bitbucket issue update failed: 404 Not Found")
}

func TestMockBitbucketProvider_GetPullRequests_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()
	options := &git.PullRequestListOptions{State: "OPEN"}

	prs, err := provider.GetPullRequests(ctx, "testowner", "testrepo", options)
	assert.NoError(t, err)
	assert.Len(t, prs, 2)

	// Check first PR
	pr1 := prs[0]
	assert.Equal(t, 1, pr1.Number)
	assert.Equal(t, "Bitbucket Mock Pull Request 1", pr1.Title)
	assert.Equal(t, "OPEN", pr1.State) // Bitbucket uses uppercase states
	assert.False(t, pr1.Merged)
	assert.NotNil(t, pr1.Head)
	assert.Equal(t, "feature-branch", pr1.Head.Ref)
	assert.NotNil(t, pr1.Base)
	assert.Equal(t, "main", pr1.Base.Ref) // Bitbucket uses "main"
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/pullrequests/1", pr1.URL)

	// Check second PR
	pr2 := prs[1]
	assert.Equal(t, 2, pr2.Number)
	assert.Equal(t, "Bitbucket Mock Pull Request 2", pr2.Title)
	assert.Equal(t, "MERGED", pr2.State) // Bitbucket uses "MERGED" for merged PRs
	assert.True(t, pr2.Merged)
	assert.NotNil(t, pr2.MergedAt)
}

func TestMockBitbucketProvider_GetPullRequests_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetPullRequests")
	ctx := context.Background()
	options := &git.PullRequestListOptions{State: "OPEN"}

	prs, err := provider.GetPullRequests(ctx, "testowner", "testrepo", options)
	assert.Error(t, err)
	assert.Nil(t, prs)
	assert.Contains(t, err.Error(), "Bitbucket pull requests retrieval failed: 500 Internal Server Error")
}

func TestMockBitbucketProvider_GetPullRequest_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequest(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "Bitbucket Mock Pull Request 123", pr.Title)
	assert.Equal(t, "OPEN", pr.State) // Bitbucket uses uppercase states
	assert.False(t, pr.Merged)
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/pullrequests/123", pr.URL)
}

func TestMockBitbucketProvider_GetPullRequest_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetPullRequest")
	ctx := context.Background()

	pr, err := provider.GetPullRequest(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "Bitbucket pull request not found: 404 Not Found")
}

func TestMockBitbucketProvider_GetPullRequestByBranch_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequestByBranch(ctx, "testowner", "testrepo", "feature-branch")
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "Bitbucket Mock PR for branch feature-branch", pr.Title)
	assert.Equal(t, "OPEN", pr.State) // Bitbucket uses uppercase states
	assert.Equal(t, "feature-branch", pr.Head.Ref)
	assert.Equal(t, "main", pr.Base.Ref) // Bitbucket uses "main"
}

func TestMockBitbucketProvider_GetPullRequestByBranch_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetPullRequestByBranch")
	ctx := context.Background()

	pr, err := provider.GetPullRequestByBranch(ctx, "testowner", "testrepo", "feature-branch")
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "Bitbucket pull request not found for branch: 404 Not Found")
}

func TestMockBitbucketProvider_GetPullRequestByIssue_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequestByIssue(ctx, "testowner", "testrepo", 456)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 456, pr.Number)
	assert.Equal(t, "Bitbucket Mock PR closing issue 456", pr.Title)
	assert.Equal(t, "OPEN", pr.State) // Bitbucket uses uppercase states
	assert.Equal(t, "fix-issue", pr.Head.Ref)
}

func TestMockBitbucketProvider_GetPullRequestByIssue_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetPullRequestByIssue")
	ctx := context.Background()

	pr, err := provider.GetPullRequestByIssue(ctx, "testowner", "testrepo", 456)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "Bitbucket pull request not found for issue: 404 Not Found")
}

func TestMockBitbucketProvider_CreatePullRequest_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
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
	assert.Equal(t, "OPEN", pr.State) // Bitbucket uses uppercase states
	assert.False(t, pr.Draft)
	assert.Equal(t, "feature-branch", pr.Head.Ref)
	assert.Equal(t, "main", pr.Base.Ref) // Bitbucket uses "main"
}

func TestMockBitbucketProvider_CreatePullRequest_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("CreatePullRequest")
	ctx := context.Background()
	request := &git.CreatePullRequestRequest{
		Title: "Test PR",
		Head:  "feature-branch",
		Base:  "main",
	}

	pr, err := provider.CreatePullRequest(ctx, "testowner", "testrepo", request)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "Bitbucket pull request creation failed: 422 Unprocessable Entity")
}

func TestMockBitbucketProvider_UpdatePullRequest_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
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
	assert.Equal(t, "OPEN", pr.State) // Bitbucket uses uppercase states
}

func TestMockBitbucketProvider_UpdatePullRequest_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("UpdatePullRequest")
	ctx := context.Background()
	updates := &git.UpdatePullRequestRequest{
		Title: stringPtr("Updated PR Title"),
	}

	pr, err := provider.UpdatePullRequest(ctx, "testowner", "testrepo", 123, updates)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "Bitbucket pull request update failed: 404 Not Found")
}

func TestMockBitbucketProvider_GetPullRequestReviews_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	reviews, err := provider.GetPullRequestReviews(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, reviews, 2)

	// Check first review
	review1 := reviews[0]
	assert.Equal(t, 1, review1.ID)
	assert.Equal(t, "approved", review1.State)
	assert.Equal(t, "This looks good!", review1.Body)
	assert.Equal(t, "bitbucket-reviewer1", review1.User.Login)
	assert.Equal(t, "abc123", review1.CommitID)
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/pullrequests/123/approve", review1.URL)

	// Check second review
	review2 := reviews[1]
	assert.Equal(t, 2, review2.ID)
	assert.Equal(t, "changes_requested", review2.State)
	assert.Equal(t, "Please fix the formatting", review2.Body)
	assert.Equal(t, "bitbucket-reviewer2", review2.User.Login)
}

func TestMockBitbucketProvider_GetPullRequestReviews_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetPullRequestReviews")
	ctx := context.Background()

	reviews, err := provider.GetPullRequestReviews(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, reviews)
	assert.Contains(t, err.Error(), "Bitbucket pull request reviews retrieval failed: 404 Not Found")
}

func TestMockBitbucketProvider_GetPullRequestComments_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	comments, err := provider.GetPullRequestComments(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	// Check first comment
	comment1 := comments[0]
	assert.Equal(t, 1, comment1.ID)
	assert.Equal(t, "Great work on this PR!", comment1.Body)
	assert.Equal(t, "bitbucket-commenter1", comment1.User.Login)
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/pullrequests/123/comments/1", comment1.URL)

	// Check second comment
	comment2 := comments[1]
	assert.Equal(t, 2, comment2.ID)
	assert.Equal(t, "Can you add more tests?", comment2.Body)
	assert.Equal(t, "bitbucket-commenter2", comment2.User.Login)
}

func TestMockBitbucketProvider_GetPullRequestComments_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetPullRequestComments")
	ctx := context.Background()

	comments, err := provider.GetPullRequestComments(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Contains(t, err.Error(), "Bitbucket pull request comments retrieval failed: 404 Not Found")
}

func TestMockBitbucketProvider_GetIssueComments_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	comments, err := provider.GetIssueComments(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	// Check first comment
	comment1 := comments[0]
	assert.Equal(t, 1, comment1.ID)
	assert.Equal(t, "This issue is important", comment1.Body)
	assert.Equal(t, "bitbucket-commenter1", comment1.User.Login)
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/issues/123/comments/1", comment1.URL)

	// Check second comment
	comment2 := comments[1]
	assert.Equal(t, 2, comment2.ID)
	assert.Equal(t, "I can help with this", comment2.Body)
	assert.Equal(t, "bitbucket-commenter2", comment2.User.Login)
}

func TestMockBitbucketProvider_GetIssueComments_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetIssueComments")
	ctx := context.Background()

	comments, err := provider.GetIssueComments(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Contains(t, err.Error(), "Bitbucket issue comments retrieval failed: 404 Not Found")
}

func TestMockBitbucketProvider_CreateComment_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()
	request := &git.CreateCommentRequest{
		Body: "Test comment body",
	}

	comment, err := provider.CreateComment(ctx, "testowner", "testrepo", 123, request)
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, 999, comment.ID)
	assert.Equal(t, "Test comment body", comment.Body)
	assert.Equal(t, "bitbucket-currentuser", comment.User.Login)
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/issues/123/comments/999", comment.URL)
}

func TestMockBitbucketProvider_CreateComment_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("CreateComment")
	ctx := context.Background()
	request := &git.CreateCommentRequest{
		Body: "Test comment body",
	}

	comment, err := provider.CreateComment(ctx, "testowner", "testrepo", 123, request)
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Contains(t, err.Error(), "Bitbucket comment creation failed: 422 Unprocessable Entity")
}

func TestMockBitbucketProvider_GetLabels_Success(t *testing.T) {
	provider := NewMockBitbucketProvider()
	ctx := context.Background()

	labels, err := provider.GetLabels(ctx, "testowner", "testrepo")
	assert.NoError(t, err)
	assert.Len(t, labels, 5)

	// Check specific labels
	bugLabel := labels[0]
	assert.Equal(t, 1, bugLabel.ID)
	assert.Equal(t, "bug", bugLabel.Name)
	assert.Equal(t, "Something isn't working", bugLabel.Description)
	assert.Equal(t, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/labels/bug", bugLabel.URL)

	enhancementLabel := labels[1]
	assert.Equal(t, 2, enhancementLabel.ID)
	assert.Equal(t, "enhancement", enhancementLabel.Name)
	assert.Equal(t, "New feature or request", enhancementLabel.Description)

	invalidLabel := labels[4]
	assert.Equal(t, 5, invalidLabel.ID)
	assert.Equal(t, "invalid", invalidLabel.Name)
	assert.Equal(t, "This doesn't seem right", invalidLabel.Description)
}

func TestMockBitbucketProvider_GetLabels_Failure(t *testing.T) {
	provider := NewMockBitbucketProviderWithFailure("GetLabels")
	ctx := context.Background()

	labels, err := provider.GetLabels(ctx, "testowner", "testrepo")
	assert.Error(t, err)
	assert.Nil(t, labels)
	assert.Contains(t, err.Error(), "Bitbucket labels retrieval failed: 404 Not Found")
}
