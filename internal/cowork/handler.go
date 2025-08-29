package cowork

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/task"
	"github.com/hlfshell/cowork/internal/types"
	"github.com/hlfshell/cowork/internal/workspace"
)

// Handler implements cowork-specific operations using any GitProvider
type Handler struct {
	provider         git.GitProvider
	taskManager      task.TaskManager
	workspaceManager workspace.WorkspaceManager
}

// NewHandler creates a new cowork handler with the given provider and managers
func NewHandler(provider git.GitProvider, taskManager task.TaskManager, workspaceManager workspace.WorkspaceManager) *Handler {
	return &Handler{
		provider:         provider,
		taskManager:      taskManager,
		workspaceManager: workspaceManager,
	}
}

// ScanOpenIssues scans all open issues that should be converted to tasks
func (h *Handler) ScanOpenIssues(ctx context.Context, owner, repo string) ([]*git.Issue, error) {
	// Get all open issues
	issues, err := h.provider.GetIssues(ctx, owner, repo, &git.IssueListOptions{
		State: "open",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}

	// Filter out already processed issues
	var filteredIssues []*git.Issue
	for _, issue := range issues {
		// Note: The provider should filter out PRs, but we can add additional filtering here if needed

		// Check if task already exists for this issue
		existingTask, err := h.GetTaskByIssue(ctx, owner, repo, issue.Number)
		if err == nil && existingTask != nil {
			continue // Skip if task already exists
		}

		filteredIssues = append(filteredIssues, issue)
	}

	return filteredIssues, nil
}

// CreateTaskFromIssue creates a new task from a provider issue
func (h *Handler) CreateTaskFromIssue(ctx context.Context, owner, repo string, issue *git.Issue) (*types.Task, error) {
	// Check if task already exists for this issue
	existingTask, err := h.GetTaskByIssue(ctx, owner, repo, issue.Number)
	if err == nil && existingTask != nil {
		return existingTask, nil
	}

	// Generate task name from issue
	taskName := h.generateTaskName(issue)

	// Create task request
	req := &types.CreateTaskRequest{
		Name:        taskName,
		Description: issue.Body,
		TicketID:    fmt.Sprintf("%s:%s/%s#%d", h.provider.GetProviderType(), owner, repo, issue.Number),
		URL:         issue.URL,
		Priority:    h.calculatePriority(issue),
		Tags:        h.extractTags(issue),
		Metadata: map[string]string{
			"provider":     string(h.provider.GetProviderType()),
			"owner":        owner,
			"repo":         repo,
			"issue_number": fmt.Sprintf("%d", issue.Number),
			"issue_title":  issue.Title,
			"issue_author": h.getIssueAuthor(issue),
			"created_at":   issue.CreatedAt.Format(time.RFC3339),
		},
	}

	return h.taskManager.CreateTask(req)
}

// GetTaskByIssue retrieves a task that was created from a specific issue
func (h *Handler) GetTaskByIssue(ctx context.Context, owner, repo string, issueNumber int) (*types.Task, error) {
	// Search for task with matching ticket ID
	ticketID := fmt.Sprintf("%s:%s/%s#%d", h.provider.GetProviderType(), owner, repo, issueNumber)
	
	tasks, err := h.taskManager.ListTasks(&types.TaskFilter{
		Search: ticketID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	for _, task := range tasks {
		if task.TicketID == ticketID {
			return task, nil
		}
	}

	return nil, nil // No task found
}

// CreateWorkspaceForTask creates a workspace for a task
func (h *Handler) CreateWorkspaceForTask(ctx context.Context, task *types.Task, owner, repo string) (*types.Workspace, error) {
	// Get the original issue to generate branch name
	issueNumber := h.extractIssueNumber(task.TicketID)
	if issueNumber == 0 {
		return nil, fmt.Errorf("invalid ticket ID format: %s", task.TicketID)
	}

	issue, err := h.provider.GetIssue(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Generate branch name
	branchName := h.GenerateBranchName(issue)

	// Create workspace request
	req := &types.CreateWorkspaceRequest{
		TaskName:   task.Name,
		Description: task.Description,
		TicketID:   task.TicketID,
		SourceRepo: fmt.Sprintf("https://%s/%s/%s.git", h.getProviderHost(), owner, repo),
		BaseBranch: "main", // Default, could be made configurable
		Metadata: map[string]string{
			"branch_name": branchName,
			"issue_number": fmt.Sprintf("%d", issueNumber),
		},
	}

	return h.workspaceManager.CreateWorkspace(req)
}

// GenerateBranchName generates a branch name for a task based on the issue
func (h *Handler) GenerateBranchName(issue *git.Issue) string {
	// Create a human-readable branch name
	title := strings.ToLower(issue.Title)
	
	// Remove special characters and replace spaces with hyphens
	title = strings.ReplaceAll(title, " ", "-")
	title = strings.ReplaceAll(title, "_", "-")
	title = strings.ReplaceAll(title, ".", "")
	title = strings.ReplaceAll(title, ",", "")
	title = strings.ReplaceAll(title, ":", "")
	title = strings.ReplaceAll(title, ";", "")
	title = strings.ReplaceAll(title, "!", "")
	title = strings.ReplaceAll(title, "?", "")
	title = strings.ReplaceAll(title, "(", "")
	title = strings.ReplaceAll(title, ")", "")
	title = strings.ReplaceAll(title, "[", "")
	title = strings.ReplaceAll(title, "]", "")
	title = strings.ReplaceAll(title, "{", "")
	title = strings.ReplaceAll(title, "}", "")
	title = strings.ReplaceAll(title, "/", "-")
	title = strings.ReplaceAll(title, "\\", "-")
	
	// Remove multiple consecutive hyphens
	for strings.Contains(title, "--") {
		title = strings.ReplaceAll(title, "--", "-")
	}
	
	// Trim hyphens from start and end
	title = strings.Trim(title, "-")
	
	// Limit length
	if len(title) > 30 {
		title = title[:30]
		title = strings.Trim(title, "-")
	}
	
	// Ensure it's not empty
	if title == "" {
		title = "task"
	}
	
	return fmt.Sprintf("task/%s-%d", title, issue.Number)
}

// CreatePullRequestForTask creates a pull request for a completed task
func (h *Handler) CreatePullRequestForTask(ctx context.Context, task *types.Task, owner, repo string, workspace *types.Workspace) (*git.PullRequest, error) {
	// Get the original issue
	issueNumber := h.extractIssueNumber(task.TicketID)
	if issueNumber == 0 {
		return nil, fmt.Errorf("invalid ticket ID format: %s", task.TicketID)
	}

	issue, err := h.provider.GetIssue(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Check if PR already exists
	existingPR, err := h.GetPullRequestForTask(ctx, task, owner, repo)
	if err == nil && existingPR != nil {
		return existingPR, nil
	}

	// Create pull request
	branchName := workspace.Metadata["branch_name"]
	if branchName == "" {
		branchName = h.GenerateBranchName(issue)
	}

	prReq := &git.CreatePullRequestRequest{
		Title:  fmt.Sprintf("Fix: %s", issue.Title),
		Body:   h.generatePRBody(task, issue),
		Head:   branchName,
		Base:   "main", // Default, could be made configurable
		Draft:  false,
	}

	return h.provider.CreatePullRequest(ctx, owner, repo, prReq)
}

// GetPullRequestForTask retrieves the pull request associated with a task
func (h *Handler) GetPullRequestForTask(ctx context.Context, task *types.Task, owner, repo string) (*git.PullRequest, error) {
	// Try to find PR by branch name
	branchName := task.Metadata["branch_name"]
	if branchName == "" {
		// Generate branch name from issue
		issueNumber := h.extractIssueNumber(task.TicketID)
		if issueNumber == 0 {
			return nil, nil
		}

		issue, err := h.provider.GetIssue(ctx, owner, repo, issueNumber)
		if err != nil {
			return nil, nil
		}
		branchName = h.GenerateBranchName(issue)
	}

	return h.provider.GetPullRequestByBranch(ctx, owner, repo, branchName)
}

// ScanPullRequestsForTasks scans pull requests associated with known tasks
func (h *Handler) ScanPullRequestsForTasks(ctx context.Context, owner, repo string, knownTaskIDs []string) ([]*git.PullRequest, error) {
	// Get all open pull requests
	prs, err := h.provider.GetPullRequests(ctx, owner, repo, &git.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	// Filter PRs that are associated with known tasks
	var taskPRs []*git.PullRequest
	for _, pr := range prs {
		// Check if this PR is associated with any known task
		for _, taskID := range knownTaskIDs {
			task, err := h.taskManager.GetTask(taskID)
			if err != nil {
				continue
			}

			// Check if PR branch matches task's expected branch
			expectedBranch := task.Metadata["branch_name"]
			if expectedBranch != "" && pr.Head.Ref == expectedBranch {
				taskPRs = append(taskPRs, pr)
				break
			}
		}
	}

	return taskPRs, nil
}

// Helper methods

func (h *Handler) generateTaskName(issue *git.Issue) string {
	title := issue.Title
	if len(title) > 50 {
		title = title[:47] + "..."
	}
	return title
}

func (h *Handler) calculatePriority(issue *git.Issue) int {
	// Default priority
	priority := 5

	// Check for priority labels
	for _, label := range issue.Labels {
		switch strings.ToLower(label.Name) {
		case "urgent", "critical", "p0":
			priority = 1
		case "high", "p1":
			priority = 2
		case "medium", "p2":
			priority = 3
		case "low", "p3":
			priority = 4
		}
	}

	return priority
}

func (h *Handler) extractTags(issue *git.Issue) []string {
	var tags []string
	for _, label := range issue.Labels {
		// Skip priority labels
		if strings.HasPrefix(strings.ToLower(label.Name), "p") {
			continue
		}
		tags = append(tags, label.Name)
	}
	return tags
}

func (h *Handler) getIssueAuthor(issue *git.Issue) string {
	if issue.Author != nil {
		return issue.Author.Login
	}
	return "unknown"
}

func (h *Handler) extractIssueNumber(ticketID string) int {
	// Parse ticket ID format: provider:owner/repo#number
	parts := strings.Split(ticketID, "#")
	if len(parts) != 2 {
		return 0
	}
	
	var number int
	fmt.Sscanf(parts[1], "%d", &number)
	return number
}

func (h *Handler) getProviderHost() string {
	switch h.provider.GetProviderType() {
	case git.ProviderGitHub:
		return "github.com"
	case git.ProviderGitLab:
		return "gitlab.com"
	case git.ProviderBitbucket:
		return "bitbucket.org"
	default:
		return "github.com" // Default fallback
	}
}

func (h *Handler) generatePRBody(task *types.Task, issue *git.Issue) string {
	return fmt.Sprintf(`## Summary

This PR addresses issue #%d: %s

## Changes

- [Task completed by Cowork](%s)

## Related

- Closes #%d

---
*This PR was automatically generated by Cowork*`, 
		issue.Number, issue.Title, task.URL, issue.Number)
}
