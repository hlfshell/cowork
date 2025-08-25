package gitprovider

import (
	"context"
	"testing"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestMockGitLabProvider_GetProviderType(t *testing.T) {
	provider := NewMockGitLabProvider()
	assert.Equal(t, git.ProviderGitLab, provider.GetProviderType())
}

func TestMockGitLabProvider_TestAuth_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.NoError(t, err)
}

func TestMockGitLabProvider_TestAuth_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("TestAuth")
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitLab authentication failed: 401 Unauthorized")
}

func TestMockGitLabProvider_TestAuth_RateLimit(t *testing.T) {
	provider := NewMockGitLabProviderWithRateLimit()
	ctx := context.Background()

	err := provider.TestAuth(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitLab API rate limit exceeded: 429 Too Many Requests")
}

func TestMockGitLabProvider_GetRepositoryInfo_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	repo, err := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "testowner", repo.Owner)
	assert.Equal(t, "testrepo", repo.Name)
	assert.Equal(t, "testowner/testrepo", repo.FullName)
	assert.Equal(t, "master", repo.DefaultBranch) // GitLab uses "master" as default
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo", repo.URL)
	assert.False(t, repo.Private)
}

func TestMockGitLabProvider_GetRepositoryInfo_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetRepositoryInfo")
	ctx := context.Background()

	repo, err := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "GitLab project not found: 404 Not Found")
}

func TestMockGitLabProvider_GetIssues_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()
	options := &git.IssueListOptions{State: "opened"}

	issues, err := provider.GetIssues(ctx, "testowner", "testrepo", options)
	assert.NoError(t, err)
	assert.Len(t, issues, 2)

	// Check first issue
	issue1 := issues[0]
	assert.Equal(t, 1, issue1.Number)
	assert.Equal(t, "GitLab Mock Issue 1", issue1.Title)
	assert.Equal(t, "opened", issue1.State) // GitLab uses "opened" instead of "open"
	assert.NotNil(t, issue1.Author)
	assert.Equal(t, "gitlab-user1", issue1.Author.Login)
	assert.Len(t, issue1.Assignees, 1)
	assert.Len(t, issue1.Labels, 2)
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/issues/1", issue1.URL)

	// Check second issue
	issue2 := issues[1]
	assert.Equal(t, 2, issue2.Number)
	assert.Equal(t, "GitLab Mock Issue 2", issue2.Title)
	assert.Equal(t, "closed", issue2.State)
	assert.NotNil(t, issue2.ClosedAt)
}

func TestMockGitLabProvider_GetIssues_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetIssues")
	ctx := context.Background()
	options := &git.IssueListOptions{State: "opened"}

	issues, err := provider.GetIssues(ctx, "testowner", "testrepo", options)
	assert.Error(t, err)
	assert.Nil(t, issues)
	assert.Contains(t, err.Error(), "GitLab issues retrieval failed: 500 Internal Server Error")
}

func TestMockGitLabProvider_GetIssue_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	issue, err := provider.GetIssue(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 123, issue.Number)
	assert.Equal(t, "GitLab Mock Issue 123", issue.Title)
	assert.Equal(t, "opened", issue.State) // GitLab uses "opened"
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/issues/123", issue.URL)
}

func TestMockGitLabProvider_GetIssue_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetIssue")
	ctx := context.Background()

	issue, err := provider.GetIssue(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "GitLab issue not found: 404 Not Found")
}

func TestMockGitLabProvider_CreateIssue_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
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
	assert.Equal(t, "opened", issue.State) // GitLab uses "opened"
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/issues/999", issue.URL)
}

func TestMockGitLabProvider_CreateIssue_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("CreateIssue")
	ctx := context.Background()
	request := &git.CreateIssueRequest{
		Title: "Test Issue",
		Body:  "Test issue body",
	}

	issue, err := provider.CreateIssue(ctx, "testowner", "testrepo", request)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "GitLab issue creation failed: 422 Unprocessable Entity")
}

func TestMockGitLabProvider_UpdateIssue_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
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
	assert.Equal(t, "opened", issue.State) // Mock returns "opened" regardless of update
}

func TestMockGitLabProvider_UpdateIssue_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("UpdateIssue")
	ctx := context.Background()
	updates := &git.UpdateIssueRequest{
		Title: stringPtr("Updated Title"),
	}

	issue, err := provider.UpdateIssue(ctx, "testowner", "testrepo", 123, updates)
	assert.Error(t, err)
	assert.Nil(t, issue)
	assert.Contains(t, err.Error(), "GitLab issue update failed: 404 Not Found")
}

func TestMockGitLabProvider_GetPullRequests_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()
	options := &git.PullRequestListOptions{State: "opened"}

	prs, err := provider.GetPullRequests(ctx, "testowner", "testrepo", options)
	assert.NoError(t, err)
	assert.Len(t, prs, 2)

	// Check first MR
	pr1 := prs[0]
	assert.Equal(t, 1, pr1.Number)
	assert.Equal(t, "GitLab Mock Merge Request 1", pr1.Title)
	assert.Equal(t, "opened", pr1.State) // GitLab uses "opened"
	assert.False(t, pr1.Merged)
	assert.NotNil(t, pr1.Head)
	assert.Equal(t, "feature-branch", pr1.Head.Ref)
	assert.NotNil(t, pr1.Base)
	assert.Equal(t, "master", pr1.Base.Ref) // GitLab uses "master"
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/merge_requests/1", pr1.URL)

	// Check second MR
	pr2 := prs[1]
	assert.Equal(t, 2, pr2.Number)
	assert.Equal(t, "GitLab Mock Merge Request 2", pr2.Title)
	assert.Equal(t, "merged", pr2.State) // GitLab uses "merged" for merged MRs
	assert.True(t, pr2.Merged)
	assert.NotNil(t, pr2.MergedAt)
}

func TestMockGitLabProvider_GetPullRequests_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetPullRequests")
	ctx := context.Background()
	options := &git.PullRequestListOptions{State: "opened"}

	prs, err := provider.GetPullRequests(ctx, "testowner", "testrepo", options)
	assert.Error(t, err)
	assert.Nil(t, prs)
	assert.Contains(t, err.Error(), "GitLab merge requests retrieval failed: 500 Internal Server Error")
}

func TestMockGitLabProvider_GetPullRequest_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequest(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "GitLab Mock Merge Request 123", pr.Title)
	assert.Equal(t, "opened", pr.State) // GitLab uses "opened"
	assert.False(t, pr.Merged)
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/merge_requests/123", pr.URL)
}

func TestMockGitLabProvider_GetPullRequest_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetPullRequest")
	ctx := context.Background()

	pr, err := provider.GetPullRequest(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitLab merge request not found: 404 Not Found")
}

func TestMockGitLabProvider_GetPullRequestByBranch_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequestByBranch(ctx, "testowner", "testrepo", "feature-branch")
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "GitLab Mock MR for branch feature-branch", pr.Title)
	assert.Equal(t, "opened", pr.State) // GitLab uses "opened"
	assert.Equal(t, "feature-branch", pr.Head.Ref)
	assert.Equal(t, "master", pr.Base.Ref) // GitLab uses "master"
}

func TestMockGitLabProvider_GetPullRequestByBranch_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetPullRequestByBranch")
	ctx := context.Background()

	pr, err := provider.GetPullRequestByBranch(ctx, "testowner", "testrepo", "feature-branch")
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitLab merge request not found for branch: 404 Not Found")
}

func TestMockGitLabProvider_GetPullRequestByIssue_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	pr, err := provider.GetPullRequestByIssue(ctx, "testowner", "testrepo", 456)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 456, pr.Number)
	assert.Equal(t, "GitLab Mock MR closing issue 456", pr.Title)
	assert.Equal(t, "opened", pr.State) // GitLab uses "opened"
	assert.Equal(t, "fix-issue", pr.Head.Ref)
}

func TestMockGitLabProvider_GetPullRequestByIssue_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetPullRequestByIssue")
	ctx := context.Background()

	pr, err := provider.GetPullRequestByIssue(ctx, "testowner", "testrepo", 456)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitLab merge request not found for issue: 404 Not Found")
}

func TestMockGitLabProvider_CreatePullRequest_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()
	request := &git.CreatePullRequestRequest{
		Title: "Test MR",
		Body:  "Test MR body",
		Head:  "feature-branch",
		Base:  "master",
		Draft: false,
	}

	pr, err := provider.CreatePullRequest(ctx, "testowner", "testrepo", request)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 999, pr.Number)
	assert.Equal(t, "Test MR", pr.Title)
	assert.Equal(t, "Test MR body", pr.Body)
	assert.Equal(t, "opened", pr.State) // GitLab uses "opened"
	assert.False(t, pr.Draft)
	assert.Equal(t, "feature-branch", pr.Head.Ref)
	assert.Equal(t, "master", pr.Base.Ref) // GitLab uses "master"
}

func TestMockGitLabProvider_CreatePullRequest_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("CreatePullRequest")
	ctx := context.Background()
	request := &git.CreatePullRequestRequest{
		Title: "Test MR",
		Head:  "feature-branch",
		Base:  "master",
	}

	pr, err := provider.CreatePullRequest(ctx, "testowner", "testrepo", request)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitLab merge request creation failed: 422 Unprocessable Entity")
}

func TestMockGitLabProvider_UpdatePullRequest_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()
	updates := &git.UpdatePullRequestRequest{
		Title: stringPtr("Updated MR Title"),
		Body:  stringPtr("Updated MR body"),
	}

	pr, err := provider.UpdatePullRequest(ctx, "testowner", "testrepo", 123, updates)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "Updated MR Title", pr.Title)
	assert.Equal(t, "opened", pr.State) // GitLab uses "opened"
}

func TestMockGitLabProvider_UpdatePullRequest_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("UpdatePullRequest")
	ctx := context.Background()
	updates := &git.UpdatePullRequestRequest{
		Title: stringPtr("Updated MR Title"),
	}

	pr, err := provider.UpdatePullRequest(ctx, "testowner", "testrepo", 123, updates)
	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Contains(t, err.Error(), "GitLab merge request update failed: 404 Not Found")
}

func TestMockGitLabProvider_GetPullRequestReviews_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	reviews, err := provider.GetPullRequestReviews(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, reviews, 2)

	// Check first review
	review1 := reviews[0]
	assert.Equal(t, 1, review1.ID)
	assert.Equal(t, "approved", review1.State)
	assert.Equal(t, "This looks good!", review1.Body)
	assert.Equal(t, "gitlab-reviewer1", review1.User.Login)
	assert.Equal(t, "abc123", review1.CommitID)
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/merge_requests/123/approvals", review1.URL)

	// Check second review
	review2 := reviews[1]
	assert.Equal(t, 2, review2.ID)
	assert.Equal(t, "changes_requested", review2.State)
	assert.Equal(t, "Please fix the formatting", review2.Body)
	assert.Equal(t, "gitlab-reviewer2", review2.User.Login)
}

func TestMockGitLabProvider_GetPullRequestReviews_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetPullRequestReviews")
	ctx := context.Background()

	reviews, err := provider.GetPullRequestReviews(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, reviews)
	assert.Contains(t, err.Error(), "GitLab merge request reviews retrieval failed: 404 Not Found")
}

func TestMockGitLabProvider_GetPullRequestComments_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	comments, err := provider.GetPullRequestComments(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	// Check first comment
	comment1 := comments[0]
	assert.Equal(t, 1, comment1.ID)
	assert.Equal(t, "Great work on this MR!", comment1.Body)
	assert.Equal(t, "gitlab-commenter1", comment1.User.Login)
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/merge_requests/123/notes/1", comment1.URL)

	// Check second comment
	comment2 := comments[1]
	assert.Equal(t, 2, comment2.ID)
	assert.Equal(t, "Can you add more tests?", comment2.Body)
	assert.Equal(t, "gitlab-commenter2", comment2.User.Login)
}

func TestMockGitLabProvider_GetPullRequestComments_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetPullRequestComments")
	ctx := context.Background()

	comments, err := provider.GetPullRequestComments(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Contains(t, err.Error(), "GitLab merge request comments retrieval failed: 404 Not Found")
}

func TestMockGitLabProvider_GetIssueComments_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	comments, err := provider.GetIssueComments(ctx, "testowner", "testrepo", 123)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	// Check first comment
	comment1 := comments[0]
	assert.Equal(t, 1, comment1.ID)
	assert.Equal(t, "This issue is important", comment1.Body)
	assert.Equal(t, "gitlab-commenter1", comment1.User.Login)
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/issues/123/notes/1", comment1.URL)

	// Check second comment
	comment2 := comments[1]
	assert.Equal(t, 2, comment2.ID)
	assert.Equal(t, "I can help with this", comment2.Body)
	assert.Equal(t, "gitlab-commenter2", comment2.User.Login)
}

func TestMockGitLabProvider_GetIssueComments_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetIssueComments")
	ctx := context.Background()

	comments, err := provider.GetIssueComments(ctx, "testowner", "testrepo", 123)
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Contains(t, err.Error(), "GitLab issue comments retrieval failed: 404 Not Found")
}

func TestMockGitLabProvider_CreateComment_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()
	request := &git.CreateCommentRequest{
		Body: "Test comment body",
	}

	comment, err := provider.CreateComment(ctx, "testowner", "testrepo", 123, request)
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, 999, comment.ID)
	assert.Equal(t, "Test comment body", comment.Body)
	assert.Equal(t, "gitlab-currentuser", comment.User.Login)
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/issues/123/notes/999", comment.URL)
}

func TestMockGitLabProvider_CreateComment_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("CreateComment")
	ctx := context.Background()
	request := &git.CreateCommentRequest{
		Body: "Test comment body",
	}

	comment, err := provider.CreateComment(ctx, "testowner", "testrepo", 123, request)
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Contains(t, err.Error(), "GitLab comment creation failed: 422 Unprocessable Entity")
}

func TestMockGitLabProvider_GetLabels_Success(t *testing.T) {
	provider := NewMockGitLabProvider()
	ctx := context.Background()

	labels, err := provider.GetLabels(ctx, "testowner", "testrepo")
	assert.NoError(t, err)
	assert.Len(t, labels, 5)

	// Check specific labels
	bugLabel := labels[0]
	assert.Equal(t, 1, bugLabel.ID)
	assert.Equal(t, "bug", bugLabel.Name)
	assert.Equal(t, "Something isn't working", bugLabel.Description)
	assert.Equal(t, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/labels/bug", bugLabel.URL)

	enhancementLabel := labels[1]
	assert.Equal(t, 2, enhancementLabel.ID)
	assert.Equal(t, "enhancement", enhancementLabel.Name)
	assert.Equal(t, "New feature or request", enhancementLabel.Description)

	wontfixLabel := labels[4]
	assert.Equal(t, 5, wontfixLabel.ID)
	assert.Equal(t, "wontfix", wontfixLabel.Name)
	assert.Equal(t, "This will not be worked on", wontfixLabel.Description)
}

func TestMockGitLabProvider_GetLabels_Failure(t *testing.T) {
	provider := NewMockGitLabProviderWithFailure("GetLabels")
	ctx := context.Background()

	labels, err := provider.GetLabels(ctx, "testowner", "testrepo")
	assert.Error(t, err)
	assert.Nil(t, labels)
	assert.Contains(t, err.Error(), "GitLab labels retrieval failed: 404 Not Found")
}
