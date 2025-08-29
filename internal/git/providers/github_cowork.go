package gitprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/task"
	"github.com/hlfshell/cowork/internal/types"
	"github.com/hlfshell/cowork/internal/workspace"
)

// Branch name generation constants
const (
	// MaxBranchNameLength is the maximum length for branch names (25 characters)
	MaxBranchNameLength = 25

	// DefaultTaskName is the default name used when no valid task name can be generated
	DefaultTaskName = "task"
)

// GitHubCoworkProvider implements the CoworkProvider interface for GitHub
type GitHubCoworkProvider struct {
	*GitHubProvider
	taskManager      task.TaskManager
	workspaceManager workspace.WorkspaceManager
	currentUser      string
}

// NewGitHubCoworkProvider creates a new GitHub cowork provider instance
func NewGitHubCoworkProvider(token, baseURL string, taskManager task.TaskManager, workspaceManager workspace.WorkspaceManager) (*GitHubCoworkProvider, error) {
	baseProvider, err := NewGitHubProvider(token, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create base GitHub provider: %w", err)
	}

	// Get current user for filtering assigned issues
	currentUser, _, err := baseProvider.client.Users.Get(context.Background(), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return &GitHubCoworkProvider{
		GitHubProvider:   baseProvider,
		taskManager:      taskManager,
		workspaceManager: workspaceManager,
		currentUser:      *currentUser.Login,
	}, nil
}

// ScanOpenIssues scans all open issues assigned to the current user
func (gcp *GitHubCoworkProvider) ScanOpenIssues(ctx context.Context, owner, repo string) ([]*git.Issue, error) {
	// Get issues assigned to current user that are open
	issues, _, err := gcp.client.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		State:     "open",
		Assignee:  gcp.currentUser,
		Sort:      "updated",
		Direction: "desc",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	var gitIssues []*git.Issue
	for _, issue := range issues {
		// Skip pull requests (GitHub API returns both issues and PRs)
		if issue.PullRequestLinks != nil {
			continue
		}

		gitIssue := gcp.convertGitHubIssue(issue)
		gitIssues = append(gitIssues, gitIssue)
	}

	return gitIssues, nil
}

// CreateTaskFromIssue creates a new task from a provider issue
func (gcp *GitHubCoworkProvider) CreateTaskFromIssue(ctx context.Context, owner, repo string, issue *git.Issue) (*types.Task, error) {
	// Check if task already exists for this issue
	existingTask, err := gcp.GetTaskByIssue(ctx, owner, repo, issue.Number)
	if err == nil && existingTask != nil {
		return existingTask, nil
	}

	// Generate task name from issue
	taskName := gcp.generateTaskName(issue)

	// Create task request
	req := &types.CreateTaskRequest{
		Name:        taskName,
		Description: issue.Body,
		TicketID:    fmt.Sprintf("github:%s/%s#%d", owner, repo, issue.Number),
		URL:         issue.URL,
		Priority:    gcp.calculatePriority(issue),
		Tags:        gcp.extractTags(issue),
		Metadata: map[string]string{
			"provider":     "github",
			"owner":        owner,
			"repo":         repo,
			"issue_number": fmt.Sprintf("%d", issue.Number),
			"issue_title":  issue.Title,
			"issue_author": gcp.getIssueAuthor(issue),
			"created_at":   issue.CreatedAt.Format(time.RFC3339),
		},
	}

	task, err := gcp.taskManager.CreateTask(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create task from issue: %w", err)
	}

	return task, nil
}

// GetTaskByIssue retrieves a task that was created from a specific issue
func (gcp *GitHubCoworkProvider) GetTaskByIssue(ctx context.Context, owner, repo string, issueNumber int) (*types.Task, error) {
	ticketID := fmt.Sprintf("github:%s/%s#%d", owner, repo, issueNumber)

	// List all tasks and find the one with matching ticket ID
	tasks, err := gcp.taskManager.ListTasks(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range tasks {
		if task.TicketID == ticketID {
			return task, nil
		}
	}

	return nil, nil // No task found
}

// CreateWorkspaceForTask creates a workspace for a task when it starts
func (gcp *GitHubCoworkProvider) CreateWorkspaceForTask(ctx context.Context, task *types.Task, owner, repo string) (*types.Workspace, error) {
	// Get the original issue to generate branch name
	issueNumber, err := gcp.extractIssueNumber(task.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract issue number from ticket ID: %w", err)
	}

	issue, err := gcp.GetIssue(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Generate branch name
	branchName := gcp.GenerateBranchName(issue)

	// Create workspace request
	req := &types.CreateWorkspaceRequest{
		TaskName:    task.Name,
		Description: task.Description,
		TicketID:    task.TicketID,
		SourceRepo:  fmt.Sprintf("https://github.com/%s/%s.git", owner, repo),
		BaseBranch:  "main",  // TODO: Get from repository info
		TaskID:      task.ID, // Ensure workspace shares task ID
		Metadata: map[string]string{
			"provider":     "github",
			"owner":        owner,
			"repo":         repo,
			"issue_number": fmt.Sprintf("%d", issueNumber),
			"branch_name":  branchName,
		},
	}

	workspace, err := gcp.workspaceManager.CreateWorkspace(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Update task with workspace information
	updateReq := &types.UpdateTaskRequest{
		TaskID:        task.ID,
		WorkspaceID:   &workspace.ID,
		WorkspacePath: &workspace.Path,
		BranchName:    &branchName,
		SourceRepo:    &req.SourceRepo,
		BaseBranch:    &req.BaseBranch,
	}

	_, err = gcp.taskManager.UpdateTask(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update task with workspace info: %w", err)
	}

	return workspace, nil
}

// GenerateBranchName generates a branch name for a task based on the issue
func (gcp *GitHubCoworkProvider) GenerateBranchName(issue *git.Issue) string {
	// Convert issue title to branch name
	branchName := strings.ToLower(issue.Title)

	// Replace underscores with hyphens
	branchName = strings.ReplaceAll(branchName, "_", "-")

	// Replace multiple spaces with single hyphens
	branchName = strings.Join(strings.Fields(branchName), "-")

	// Remove special characters that are not valid in Git branch names
	// Keep only letters, numbers, and hyphens
	branchName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, branchName)

	// Remove multiple consecutive hyphens
	for strings.Contains(branchName, "--") {
		branchName = strings.ReplaceAll(branchName, "--", "-")
	}

	// Remove leading and trailing hyphens
	branchName = strings.Trim(branchName, "-")

	// Limit length to keep it short and readable (25 characters max)
	if len(branchName) > MaxBranchNameLength {
		branchName = branchName[:MaxBranchNameLength]
		// Remove trailing hyphens again after truncation
		branchName = strings.TrimRight(branchName, "-")
	}

	// Ensure it's not empty
	if branchName == "" {
		branchName = DefaultTaskName
	}

	return branchName
}

// CreatePullRequestForTask creates a pull request for a completed task
func (gcp *GitHubCoworkProvider) CreatePullRequestForTask(ctx context.Context, task *types.Task, owner, repo string, workspace *types.Workspace) (*git.PullRequest, error) {
	// Check if PR already exists
	existingPR, err := gcp.GetPullRequestForTask(ctx, task, owner, repo)
	if err == nil && existingPR != nil {
		return existingPR, nil
	}

	// Get the original issue
	issueNumber, err := gcp.extractIssueNumber(task.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract issue number: %w", err)
	}

	issue, err := gcp.GetIssue(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Create PR request
	prReq := &git.CreatePullRequestRequest{
		Title: fmt.Sprintf("Fix #%d: %s", issue.Number, issue.Title),
		Body:  fmt.Sprintf("This PR addresses issue #%d: %s\n\nCloses #%d", issue.Number, issue.Title, issue.Number),
		Head:  task.BranchName,
		Base:  task.BaseBranch,
	}

	pr, err := gcp.CreatePullRequest(ctx, owner, repo, prReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	// Link PR to issue
	err = gcp.LinkPullRequestToIssue(ctx, owner, repo, pr, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to link PR to issue: %w", err)
	}

	return pr, nil
}

// GetPullRequestForTask retrieves the pull request associated with a task
func (gcp *GitHubCoworkProvider) GetPullRequestForTask(ctx context.Context, task *types.Task, owner, repo string) (*git.PullRequest, error) {
	if task.BranchName == "" {
		return nil, nil
	}

	// Get PR by branch name
	pr, err := gcp.GetPullRequestByBranch(ctx, owner, repo, task.BranchName)
	if err != nil {
		return nil, nil // No PR found
	}

	return pr, nil
}

// LinkPullRequestToIssue links a pull request to the original issue
func (gcp *GitHubCoworkProvider) LinkPullRequestToIssue(ctx context.Context, owner, repo string, pr *git.PullRequest, issue *git.Issue) error {
	// Add a comment to the issue linking to the PR
	commentBody := fmt.Sprintf("Related pull request: #%d", pr.Number)

	commentReq := &git.CreateCommentRequest{
		Body: commentBody,
	}

	_, err := gcp.CreateComment(ctx, owner, repo, issue.Number, commentReq)
	if err != nil {
		return fmt.Errorf("failed to create linking comment: %w", err)
	}

	return nil
}

// ScanPullRequestsForTasks scans pull requests associated with known tasks
func (gcp *GitHubCoworkProvider) ScanPullRequestsForTasks(ctx context.Context, owner, repo string, knownTaskIDs []string) ([]*git.PullRequest, error) {
	// Get all open PRs
	prs, _, err := gcp.client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	var relevantPRs []*git.PullRequest
	for _, pr := range prs {
		// Check if this PR is associated with any known task
		for _, taskID := range knownTaskIDs {
			task, err := gcp.taskManager.GetTask(taskID)
			if err != nil {
				continue
			}

			if task.BranchName == *pr.Head.Ref {
				gitPR := gcp.convertGitHubPullRequest(pr)
				relevantPRs = append(relevantPRs, gitPR)
				break
			}
		}
	}

	return relevantPRs, nil
}

// GetPullRequestUpdates retrieves recent updates for a pull request
func (gcp *GitHubCoworkProvider) GetPullRequestUpdates(ctx context.Context, owner, repo string, prNumber int, since time.Time) (*git.PullRequestUpdate, error) {
	// Get new comments
	comments, _, err := gcp.client.Issues.ListComments(ctx, owner, repo, prNumber, &github.IssueListCommentsOptions{
		Since: &since,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	// Get new reviews
	reviews, _, err := gcp.client.PullRequests.ListReviews(ctx, owner, repo, prNumber, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	// Filter reviews since the given time
	var newReviews []*git.Review
	for _, review := range reviews {
		if review.SubmittedAt.After(since) {
			gitReview := gcp.convertGitHubReview(review)
			newReviews = append(newReviews, gitReview)
		}
	}

	// Convert comments
	var newComments []*git.Comment
	for _, comment := range comments {
		gitComment := gcp.convertGitHubComment(comment)
		newComments = append(newComments, gitComment)
	}

	return &git.PullRequestUpdate{
		PRNumber:      prNumber,
		NewComments:   newComments,
		NewReviews:    newReviews,
		StatusChanges: make(map[string]interface{}), // TODO: Track status changes
		UpdatedAt:     time.Now(),
	}, nil
}

// UpdateTaskFromPullRequest updates a task based on pull request changes
func (gcp *GitHubCoworkProvider) UpdateTaskFromPullRequest(ctx context.Context, task *types.Task, pr *git.PullRequest, updates *git.PullRequestUpdate) error {
	// Update task description with new comments/reviews
	if len(updates.NewComments) > 0 || len(updates.NewReviews) > 0 {
		// TODO: Update task with new feedback
		// This could involve updating the task description, adding notes, etc.
	}

	return nil
}

// SyncTaskStatusToProvider updates the provider (issue/PR) with task status changes
func (gcp *GitHubCoworkProvider) SyncTaskStatusToProvider(ctx context.Context, task *types.Task, owner, repo string) error {
	issueNumber, err := gcp.extractIssueNumber(task.TicketID)
	if err != nil {
		return fmt.Errorf("failed to extract issue number: %w", err)
	}

	// Update issue labels based on task status
	var labels []string
	switch task.Status {
	case types.TaskStatusInProgress:
		labels = []string{"in-progress"}
	case types.TaskStatusCompleted:
		labels = []string{"completed"}
	case types.TaskStatusFailed:
		labels = []string{"failed"}
	}

	if len(labels) > 0 {
		updateReq := &git.UpdateIssueRequest{
			Labels: &labels,
		}

		_, err = gcp.UpdateIssue(ctx, owner, repo, issueNumber, updateReq)
		if err != nil {
			return fmt.Errorf("failed to update issue labels: %w", err)
		}
	}

	return nil
}

// GetProviderMetadata retrieves provider-specific metadata for a task
func (gcp *GitHubCoworkProvider) GetProviderMetadata(ctx context.Context, task *types.Task, owner, repo string) (map[string]interface{}, error) {
	issueNumber, err := gcp.extractIssueNumber(task.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract issue number: %w", err)
	}

	issue, err := gcp.GetIssue(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	metadata := map[string]interface{}{
		"issue_number": issue.Number,
		"issue_title":  issue.Title,
		"issue_state":  issue.State,
		"labels":       issue.Labels,
		"assignees":    issue.Assignees,
		"created_at":   issue.CreatedAt,
		"updated_at":   issue.UpdatedAt,
	}

	// Add PR information if available
	pr, err := gcp.GetPullRequestForTask(ctx, task, owner, repo)
	if err == nil && pr != nil {
		metadata["pr_number"] = pr.Number
		metadata["pr_state"] = pr.State
		metadata["pr_merged"] = pr.Merged
		metadata["pr_draft"] = pr.Draft
	}

	return metadata, nil
}

// Helper methods

func (gcp *GitHubCoworkProvider) generateTaskName(issue *git.Issue) string {
	// Use just the issue title for task names - they can have spaces and be more readable
	return issue.Title
}

func (gcp *GitHubCoworkProvider) calculatePriority(issue *git.Issue) int {
	// Simple priority calculation based on labels
	for _, label := range issue.Labels {
		switch strings.ToLower(label.Name) {
		case "urgent", "critical", "high":
			return 3
		case "medium":
			return 2
		case "low":
			return 1
		}
	}
	return 0 // Default priority
}

// getIssueAuthor safely gets the issue author, handling nil cases
func (gcp *GitHubCoworkProvider) getIssueAuthor(issue *git.Issue) string {
	if issue.Author != nil {
		return issue.Author.Login
	}
	return "unknown"
}

func (gcp *GitHubCoworkProvider) extractTags(issue *git.Issue) []string {
	var tags []string
	for _, label := range issue.Labels {
		tags = append(tags, label.Name)
	}
	return tags
}

func (gcp *GitHubCoworkProvider) extractIssueNumber(ticketID string) (int, error) {
	// Extract issue number from ticket ID like "github:owner/repo#123"
	parts := strings.Split(ticketID, "#")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid ticket ID format: %s", ticketID)
	}

	var issueNumber int
	_, err := fmt.Sscanf(parts[1], "%d", &issueNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to parse issue number: %w", err)
	}

	return issueNumber, nil
}

// Conversion methods (these would need to be implemented based on your existing conversion logic)
func (gcp *GitHubCoworkProvider) convertGitHubIssue(issue *github.Issue) *git.Issue {
	if issue == nil {
		return nil
	}

	gitIssue := &git.Issue{
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		Body:      issue.GetBody(),
		State:     issue.GetState(),
		URL:       issue.GetHTMLURL(),
		CreatedAt: issue.GetCreatedAt().Time,
		UpdatedAt: issue.GetUpdatedAt().Time,
	}

	// Convert author
	if issue.User != nil {
		gitIssue.Author = &git.User{
			Login: issue.User.GetLogin(),
			ID:    int(issue.User.GetID()),
		}
	}

	// Convert assignees
	if issue.Assignees != nil {
		for _, assignee := range issue.Assignees {
			gitIssue.Assignees = append(gitIssue.Assignees, &git.User{
				Login: assignee.GetLogin(),
				ID:    int(assignee.GetID()),
			})
		}
	}

	// Convert labels
	if issue.Labels != nil {
		for _, label := range issue.Labels {
			gitIssue.Labels = append(gitIssue.Labels, &git.Label{
				ID:   int(label.GetID()),
				Name: label.GetName(),
				URL:  label.GetURL(),
			})
		}
	}

	return gitIssue
}

func (gcp *GitHubCoworkProvider) convertGitHubPullRequest(pr *github.PullRequest) *git.PullRequest {
	// TODO: Implement conversion from github.PullRequest to git.PullRequest
	return &git.PullRequest{}
}

func (gcp *GitHubCoworkProvider) convertGitHubComment(comment *github.IssueComment) *git.Comment {
	// TODO: Implement conversion from github.IssueComment to git.Comment
	return &git.Comment{}
}

func (gcp *GitHubCoworkProvider) convertGitHubReview(review *github.PullRequestReview) *git.Review {
	// TODO: Implement conversion from github.PullRequestReview to git.Review
	return &git.Review{}
}
