package workflow

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/task"
	"github.com/hlfshell/cowork/internal/types"
	"github.com/hlfshell/cowork/internal/workspace"
)

// Manager orchestrates the complete task-provider integration workflow
type Manager struct {
	coworkProvider   git.CoworkProvider
	taskManager      task.TaskManager
	workspaceManager workspace.WorkspaceManager
	owner            string
	repo             string
}

// NewManager creates a new workflow manager
func NewManager(coworkProvider git.CoworkProvider, taskManager task.TaskManager, workspaceManager workspace.WorkspaceManager, owner, repo string) *Manager {
	return &Manager{
		coworkProvider:   coworkProvider,
		taskManager:      taskManager,
		workspaceManager: workspaceManager,
		owner:            owner,
		repo:             repo,
	}
}

// ScanAndCreateTasks scans open issues and creates tasks for assigned issues
// This implements feature 1: scan all issues that are not in a closed state
// This implements feature 2: create tasks for issues assigned to the current user
func (wm *Manager) ScanAndCreateTasks(ctx context.Context) error {
	log.Printf("🔍 Scanning open issues for %s/%s", wm.owner, wm.repo)

	// Scan open issues assigned to current user
	issues, err := wm.coworkProvider.ScanOpenIssues(ctx, wm.owner, wm.repo)
	if err != nil {
		return fmt.Errorf("failed to scan open issues: %w", err)
	}

	log.Printf("📋 Found %d open issues assigned to current user", len(issues))

	// Create tasks for each issue
	for _, issue := range issues {
		// Check if task already exists
		existingTask, err := wm.coworkProvider.GetTaskByIssue(ctx, wm.owner, wm.repo, issue.Number)
		if err != nil {
			log.Printf("⚠️  Error checking existing task for issue #%d: %v", issue.Number, err)
			continue
		}

		if existingTask != nil {
			log.Printf("✅ Task already exists for issue #%d: %s", issue.Number, existingTask.Name)
			continue
		}

		// Create new task
		task, err := wm.coworkProvider.CreateTaskFromIssue(ctx, wm.owner, wm.repo, issue)
		if err != nil {
			log.Printf("❌ Failed to create task for issue #%d: %v", issue.Number, err)
			continue
		}

		log.Printf("✅ Created task for issue #%d: %s (ID: %d)", issue.Number, task.Name, task.ID)
	}

	return nil
}

// ProcessQueuedTasks processes tasks that are queued and ready to start
// This implements feature 3: create workspace when task is marked as running/started
func (wm *Manager) ProcessQueuedTasks(ctx context.Context) error {
	log.Printf("🚀 Processing queued tasks")

	// Get next queued task
	task, err := wm.taskManager.GetNextQueuedTask()
	if err != nil {
		if err.Error() == "no queued tasks found" {
			log.Printf("📭 No queued tasks found")
			return nil
		}
		return fmt.Errorf("failed to get next queued task: %w", err)
	}

	log.Printf("🎯 Processing task: %s (ID: %d)", task.Name, task.ID)

	// Check if workspace already exists
	if task.WorkspaceID != 0 {
		log.Printf("✅ Workspace already exists for task %s", task.Name)
		return nil
	}

	// Create workspace for the task
	workspace, err := wm.coworkProvider.CreateWorkspaceForTask(ctx, task, wm.owner, wm.repo)
	if err != nil {
		return fmt.Errorf("failed to create workspace for task %s: %w", task.Name, err)
	}

	log.Printf("✅ Created workspace for task %s: %s", task.Name, workspace.Path)

	// Update task status to in progress
	status := types.TaskStatusInProgress
	updateReq := &types.UpdateTaskRequest{
		TaskID: task.ID,
		Status: &status,
	}

	_, err = wm.taskManager.UpdateTask(updateReq)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	log.Printf("✅ Updated task %s status to in_progress", task.Name)

	return nil
}

// CompleteTaskAndCreatePR completes a task and creates a pull request
// This implements feature 5: create PR when agent reports completion
func (wm *Manager) CompleteTaskAndCreatePR(ctx context.Context, taskID string) error {
	log.Printf("🏁 Completing task and creating PR for task ID: %s", taskID)

	// Get the task
	task, err := wm.taskManager.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task has a workspace
	if task.WorkspaceID == 0 {
		return fmt.Errorf("task %s has no workspace", task.Name)
	}

	// Get the workspace
	workspace, err := wm.workspaceManager.GetWorkspace(task.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Create pull request
	pr, err := wm.coworkProvider.CreatePullRequestForTask(ctx, task, wm.owner, wm.repo, workspace)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	log.Printf("✅ Created pull request #%d for task %s", pr.Number, task.Name)

	// Update task status to completed
	status := types.TaskStatusCompleted
	updateReq := &types.UpdateTaskRequest{
		TaskID: task.ID,
		Status: &status,
	}

	_, err = wm.taskManager.UpdateTask(updateReq)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	log.Printf("✅ Updated task %s status to completed", task.Name)

	return nil
}

// ScanPullRequestsForUpdates scans pull requests for updates and triggers additional work
// This implements the PR monitoring flow features
func (wm *Manager) ScanPullRequestsForUpdates(ctx context.Context) error {
	log.Printf("🔍 Scanning pull requests for updates")

	// Get all tasks to find their IDs
	tasks, err := wm.taskManager.ListTasks(nil)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	var taskIDs []string
	for _, task := range tasks {
		taskIDs = append(taskIDs, fmt.Sprintf("%d", task.ID))
	}

	// Scan PRs associated with known tasks
	prs, err := wm.coworkProvider.ScanPullRequestsForTasks(ctx, wm.owner, wm.repo, taskIDs)
	if err != nil {
		return fmt.Errorf("failed to scan pull requests: %w", err)
	}

	log.Printf("📋 Found %d pull requests associated with tasks", len(prs))

	// Process each PR for updates
	for _, pr := range prs {
		err := wm.processPullRequestUpdates(ctx, pr)
		if err != nil {
			log.Printf("⚠️  Error processing PR #%d: %v", pr.Number, err)
			continue
		}
	}

	return nil
}

// processPullRequestUpdates processes updates for a specific pull request
func (wm *Manager) processPullRequestUpdates(ctx context.Context, pr *git.PullRequest) error {
	// Find the task associated with this PR
	task, err := wm.findTaskByPullRequest(ctx, pr)
	if err != nil {
		return fmt.Errorf("failed to find task for PR #%d: %w", pr.Number, err)
	}

	if task == nil {
		log.Printf("⚠️  No task found for PR #%d", pr.Number)
		return nil
	}

	// Get updates since last check (use task's last activity as reference)
	since := task.LastActivity
	updates, err := wm.coworkProvider.GetPullRequestUpdates(ctx, wm.owner, wm.repo, pr.Number, since)
	if err != nil {
		return fmt.Errorf("failed to get PR updates: %w", err)
	}

	// Check if there are any updates
	if len(updates.NewComments) == 0 && len(updates.NewReviews) == 0 {
		log.Printf("📭 No updates for PR #%d since %s", pr.Number, since.Format(time.RFC3339))
		return nil
	}

	log.Printf("📝 Found %d new comments and %d new reviews for PR #%d",
		len(updates.NewComments), len(updates.NewReviews), pr.Number)

	// Update task with new information
	err = wm.coworkProvider.UpdateTaskFromPullRequest(ctx, task, pr, updates)
	if err != nil {
		return fmt.Errorf("failed to update task from PR: %w", err)
	}

	// Check if we need to trigger additional work
	if wm.shouldTriggerAdditionalWork(updates) {
		log.Printf("🔄 Triggering additional work for task %s due to PR updates", task.Name)

		// Update task status to in progress to trigger agent work
		status := types.TaskStatusInProgress
		updateReq := &types.UpdateTaskRequest{
			TaskID: task.ID,
			Status: &status,
		}

		_, err = wm.taskManager.UpdateTask(updateReq)
		if err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		log.Printf("✅ Updated task %s status to in_progress for additional work", task.Name)
	}

	return nil
}

// findTaskByPullRequest finds a task associated with a pull request
func (wm *Manager) findTaskByPullRequest(ctx context.Context, pr *git.PullRequest) (*types.Task, error) {
	tasks, err := wm.taskManager.ListTasks(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range tasks {
		if task.BranchName == pr.Head.Ref {
			return task, nil
		}
	}

	return nil, nil
}

// shouldTriggerAdditionalWork determines if PR updates should trigger additional work
func (wm *Manager) shouldTriggerAdditionalWork(updates *git.PullRequestUpdate) bool {
	// Trigger additional work if there are:
	// 1. New comments (feedback)
	// 2. New reviews with changes requested
	// 3. Status changes that require attention

	if len(updates.NewComments) > 0 {
		return true
	}

	for _, review := range updates.NewReviews {
		if review.State == "changes_requested" {
			return true
		}
	}

	// Check status changes
	if status, ok := updates.StatusChanges["mergeable"]; ok {
		if mergeable, ok := status.(bool); ok && !mergeable {
			return true
		}
	}

	return false
}

// SyncTaskStatuses syncs task statuses to the provider
func (wm *Manager) SyncTaskStatuses(ctx context.Context) error {
	log.Printf("🔄 Syncing task statuses to provider")

	tasks, err := wm.taskManager.ListTasks(nil)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range tasks {
		err := wm.coworkProvider.SyncTaskStatusToProvider(ctx, task, wm.owner, wm.repo)
		if err != nil {
			log.Printf("⚠️  Error syncing status for task %s: %v", task.Name, err)
			continue
		}
	}

	return nil
}

// RunFullWorkflow runs the complete workflow in sequence
func (wm *Manager) RunFullWorkflow(ctx context.Context) error {
	log.Printf("🚀 Starting full workflow for %s/%s", wm.owner, wm.repo)

	// Step 1: Scan and create tasks from issues
	if err := wm.ScanAndCreateTasks(ctx); err != nil {
		return fmt.Errorf("failed to scan and create tasks: %w", err)
	}

	// Step 2: Process queued tasks (create workspaces)
	if err := wm.ProcessQueuedTasks(ctx); err != nil {
		return fmt.Errorf("failed to process queued tasks: %w", err)
	}

	// Step 3: Scan PRs for updates
	if err := wm.ScanPullRequestsForUpdates(ctx); err != nil {
		return fmt.Errorf("failed to scan pull requests: %w", err)
	}

	// Step 4: Sync task statuses
	if err := wm.SyncTaskStatuses(ctx); err != nil {
		return fmt.Errorf("failed to sync task statuses: %w", err)
	}

	log.Printf("✅ Full workflow completed successfully")
	return nil
}
