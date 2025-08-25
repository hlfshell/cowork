package gitprovider

import (
	"context"
	"testing"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
)

// TestDuckTyping demonstrates that all providers implement the same interface
func TestDuckTyping(t *testing.T) {
	providers := []git.GitProvider{
		NewMockGitHubProvider(),
		NewMockGitLabProvider(),
		NewMockBitbucketProvider(),
	}

	for _, provider := range providers {
		t.Run(provider.GetProviderType().String(), func(t *testing.T) {
			ctx := context.Background()

			// Test that all providers can perform the same operations
			err := provider.TestAuth(ctx)
			assert.NoError(t, err)

			repo, err := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
			assert.NoError(t, err)
			assert.NotNil(t, repo)
			assert.Equal(t, "testowner", repo.Owner)
			assert.Equal(t, "testrepo", repo.Name)

			issues, err := provider.GetIssues(ctx, "testowner", "testrepo", &git.IssueListOptions{})
			assert.NoError(t, err)
			assert.Len(t, issues, 2)

			prs, err := provider.GetPullRequests(ctx, "testowner", "testrepo", &git.PullRequestListOptions{})
			assert.NoError(t, err)
			assert.Len(t, prs, 2)

			labels, err := provider.GetLabels(ctx, "testowner", "testrepo")
			assert.NoError(t, err)
			assert.Len(t, labels, 5)
		})
	}
}

// TestProviderSpecificBehavior verifies that each provider maintains its unique characteristics
func TestProviderSpecificBehavior(t *testing.T) {
	t.Run("GitHub_Specific_Behavior", func(t *testing.T) {
		provider := NewMockGitHubProvider()
		ctx := context.Background()

		// GitHub-specific characteristics
		repo, _ := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
		assert.Equal(t, "main", repo.DefaultBranch)
		assert.Contains(t, repo.URL, "api.github.com")

		issues, _ := provider.GetIssues(ctx, "testowner", "testrepo", &git.IssueListOptions{})
		assert.Equal(t, "open", issues[0].State)
		assert.Equal(t, "closed", issues[1].State)

		prs, _ := provider.GetPullRequests(ctx, "testowner", "testrepo", &git.PullRequestListOptions{})
		assert.Equal(t, "open", prs[0].State)
		assert.Equal(t, "closed", prs[1].State) // GitHub uses "closed" for merged PRs

		labels, _ := provider.GetLabels(ctx, "testowner", "testrepo")
		labelNames := make([]string, len(labels))
		for i, label := range labels {
			labelNames[i] = label.Name
		}
		assert.Contains(t, labelNames, "good first issue")
		assert.Contains(t, labelNames, "help wanted")
	})

	t.Run("GitLab_Specific_Behavior", func(t *testing.T) {
		provider := NewMockGitLabProvider()
		ctx := context.Background()

		// GitLab-specific characteristics
		repo, _ := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
		assert.Equal(t, "master", repo.DefaultBranch)
		assert.Contains(t, repo.URL, "gitlab.com/api/v4/projects")
		assert.Contains(t, repo.URL, "%2F") // URL encoding

		issues, _ := provider.GetIssues(ctx, "testowner", "testrepo", &git.IssueListOptions{})
		assert.Equal(t, "opened", issues[0].State) // GitLab uses "opened"
		assert.Equal(t, "closed", issues[1].State)

		prs, _ := provider.GetPullRequests(ctx, "testowner", "testrepo", &git.PullRequestListOptions{})
		assert.Equal(t, "opened", prs[0].State) // GitLab uses "opened"
		assert.Equal(t, "merged", prs[1].State) // GitLab uses "merged" for merged MRs

		labels, _ := provider.GetLabels(ctx, "testowner", "testrepo")
		labelNames := make([]string, len(labels))
		for i, label := range labels {
			labelNames[i] = label.Name
		}
		assert.Contains(t, labelNames, "wontfix")
	})

	t.Run("Bitbucket_Specific_Behavior", func(t *testing.T) {
		provider := NewMockBitbucketProvider()
		ctx := context.Background()

		// Bitbucket-specific characteristics
		repo, _ := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
		assert.Equal(t, "main", repo.DefaultBranch)
		assert.Contains(t, repo.URL, "api.bitbucket.org/2.0/repositories")

		issues, _ := provider.GetIssues(ctx, "testowner", "testrepo", &git.IssueListOptions{})
		assert.Equal(t, "open", issues[0].State)
		assert.Equal(t, "resolved", issues[1].State) // Bitbucket uses "resolved"

		prs, _ := provider.GetPullRequests(ctx, "testowner", "testrepo", &git.PullRequestListOptions{})
		assert.Equal(t, "OPEN", prs[0].State)   // Bitbucket uses uppercase
		assert.Equal(t, "MERGED", prs[1].State) // Bitbucket uses uppercase

		labels, _ := provider.GetLabels(ctx, "testowner", "testrepo")
		labelNames := make([]string, len(labels))
		for i, label := range labels {
			labelNames[i] = label.Name
		}
		assert.Contains(t, labelNames, "invalid")
	})
}

// TestErrorHandling verifies that each provider returns appropriate error messages
func TestErrorHandling(t *testing.T) {
	t.Run("GitHub_Error_Handling", func(t *testing.T) {
		provider := NewMockGitHubProviderWithFailure("TestAuth")
		ctx := context.Background()

		err := provider.TestAuth(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitHub authentication failed: 401 Unauthorized")

		rateLimitedProvider := NewMockGitHubProviderWithRateLimit()
		err = rateLimitedProvider.TestAuth(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitHub API rate limit exceeded: 403 Forbidden")
	})

	t.Run("GitLab_Error_Handling", func(t *testing.T) {
		provider := NewMockGitLabProviderWithFailure("TestAuth")
		ctx := context.Background()

		err := provider.TestAuth(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitLab authentication failed: 401 Unauthorized")

		rateLimitedProvider := NewMockGitLabProviderWithRateLimit()
		err = rateLimitedProvider.TestAuth(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitLab API rate limit exceeded: 429 Too Many Requests")
	})

	t.Run("Bitbucket_Error_Handling", func(t *testing.T) {
		provider := NewMockBitbucketProviderWithFailure("TestAuth")
		ctx := context.Background()

		err := provider.TestAuth(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Bitbucket authentication failed: 401 Unauthorized")

		rateLimitedProvider := NewMockBitbucketProviderWithRateLimit()
		err = rateLimitedProvider.TestAuth(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Bitbucket API rate limit exceeded: 429 Too Many Requests")
	})
}

// TestCrossProviderFunctionality tests that the same code works across all providers
func TestCrossProviderFunctionality(t *testing.T) {
	providers := []git.GitProvider{
		NewMockGitHubProvider(),
		NewMockGitLabProvider(),
		NewMockBitbucketProvider(),
	}

	for _, provider := range providers {
		t.Run(provider.GetProviderType().String()+"_CrossProvider", func(t *testing.T) {
			ctx := context.Background()

			// Test issue creation and retrieval
			createRequest := &git.CreateIssueRequest{
				Title:     "Cross-provider test issue",
				Body:      "This issue tests cross-provider functionality",
				Assignees: []string{"user1"},
				Labels:    []string{"bug"},
			}

			issue, err := provider.CreateIssue(ctx, "testowner", "testrepo", createRequest)
			assert.NoError(t, err)
			assert.NotNil(t, issue)
			assert.Equal(t, "Cross-provider test issue", issue.Title)

			// Test pull request creation and retrieval
			prRequest := &git.CreatePullRequestRequest{
				Title: "Cross-provider test PR",
				Body:  "This PR tests cross-provider functionality",
				Head:  "feature-branch",
				Base:  "main",
				Draft: false,
			}

			pr, err := provider.CreatePullRequest(ctx, "testowner", "testrepo", prRequest)
			assert.NoError(t, err)
			assert.NotNil(t, pr)
			assert.Equal(t, "Cross-provider test PR", pr.Title)

			// Test comment creation
			commentRequest := &git.CreateCommentRequest{
				Body: "Cross-provider test comment",
			}

			comment, err := provider.CreateComment(ctx, "testowner", "testrepo", issue.Number, commentRequest)
			assert.NoError(t, err)
			assert.NotNil(t, comment)
			assert.Equal(t, "Cross-provider test comment", comment.Body)

			// Test issue update
			updateRequest := &git.UpdateIssueRequest{
				Title: stringPtr("Updated cross-provider issue"),
				Body:  stringPtr("Updated body for cross-provider test"),
			}

			updatedIssue, err := provider.UpdateIssue(ctx, "testowner", "testrepo", issue.Number, updateRequest)
			assert.NoError(t, err)
			assert.NotNil(t, updatedIssue)
			assert.Equal(t, "Updated cross-provider issue", updatedIssue.Title)
		})
	}
}

// TestProviderTypeConsistency ensures each provider returns the correct type
func TestProviderTypeConsistency(t *testing.T) {
	githubProvider := NewMockGitHubProvider()
	assert.Equal(t, git.ProviderGitHub, githubProvider.GetProviderType())

	gitlabProvider := NewMockGitLabProvider()
	assert.Equal(t, git.ProviderGitLab, gitlabProvider.GetProviderType())

	bitbucketProvider := NewMockBitbucketProvider()
	assert.Equal(t, git.ProviderBitbucket, bitbucketProvider.GetProviderType())
}

// TestURLPatterns verifies that each provider uses the correct URL patterns
func TestURLPatterns(t *testing.T) {
	t.Run("GitHub_URL_Patterns", func(t *testing.T) {
		provider := NewMockGitHubProvider()
		ctx := context.Background()

		repo, _ := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
		assert.Contains(t, repo.URL, "https://api.github.com/repos/testowner/testrepo")

		issues, _ := provider.GetIssues(ctx, "testowner", "testrepo", &git.IssueListOptions{})
		assert.Contains(t, issues[0].URL, "https://api.github.com/repos/testowner/testrepo/issues/1")

		prs, _ := provider.GetPullRequests(ctx, "testowner", "testrepo", &git.PullRequestListOptions{})
		assert.Contains(t, prs[0].URL, "https://api.github.com/repos/testowner/testrepo/pulls/1")
	})

	t.Run("GitLab_URL_Patterns", func(t *testing.T) {
		provider := NewMockGitLabProvider()
		ctx := context.Background()

		repo, _ := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
		assert.Contains(t, repo.URL, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo")

		issues, _ := provider.GetIssues(ctx, "testowner", "testrepo", &git.IssueListOptions{})
		assert.Contains(t, issues[0].URL, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/issues/1")

		prs, _ := provider.GetPullRequests(ctx, "testowner", "testrepo", &git.PullRequestListOptions{})
		assert.Contains(t, prs[0].URL, "https://gitlab.com/api/v4/projects/testowner%2Ftestrepo/merge_requests/1")
	})

	t.Run("Bitbucket_URL_Patterns", func(t *testing.T) {
		provider := NewMockBitbucketProvider()
		ctx := context.Background()

		repo, _ := provider.GetRepositoryInfo(ctx, "testowner", "testrepo")
		assert.Contains(t, repo.URL, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo")

		issues, _ := provider.GetIssues(ctx, "testowner", "testrepo", &git.IssueListOptions{})
		assert.Contains(t, issues[0].URL, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/issues/1")

		prs, _ := provider.GetPullRequests(ctx, "testowner", "testrepo", &git.PullRequestListOptions{})
		assert.Contains(t, prs[0].URL, "https://api.bitbucket.org/2.0/repositories/testowner/testrepo/pullrequests/1")
	})
}
