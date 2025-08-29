package workflow

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/task"
	"github.com/hlfshell/cowork/internal/types"
	"github.com/hlfshell/cowork/internal/workspace"
)

// Engine orchestrates the complete auto-PR workflow
type Engine struct {
	workflowManager  *WorkflowManager
	taskManager      task.TaskManager
	workspaceManager workspace.WorkspaceManager
	coworkProvider   git.CoworkProvider
	owner            string
	repo             string
	processID        string
}

// NewEngine creates a new workflow engine
func NewEngine(workflowManager *WorkflowManager, taskManager task.TaskManager, workspaceManager workspace.WorkspaceManager, coworkProvider git.CoworkProvider, owner, repo string) *Engine {
	return &Engine{
		workflowManager:  workflowManager,
		taskManager:      taskManager,
		workspaceManager: workspaceManager,
		coworkProvider:   coworkProvider,
		owner:            owner,
		repo:             repo,
		processID:        fmt.Sprintf("engine-%d", os.Getpid()),
	}
}

// ProcessWorkflow processes a workflow through its complete lifecycle
func (e *Engine) ProcessWorkflow(ctx context.Context, workflowID string) error {
	log.Printf("🚀 Starting workflow processing for %s", workflowID)

	// Get the workflow
	workflow, err := e.workflowManager.GetWorkflow(workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Try to acquire lock
	lock, err := e.workflowManager.LockWorkflow(workflowID, e.processID, DefaultLockTimeout)
	if err != nil {
		return fmt.Errorf("failed to acquire workflow lock: %w", err)
	}

	// Ensure we release the lock when done
	defer func() {
		if err := e.workflowManager.UnlockWorkflow(workflowID, e.processID); err != nil {
			log.Printf("⚠️  Failed to release workflow lock: %v", err)
		}
	}()

	log.Printf("🔒 Acquired lock for workflow %s (timeout: %s)", workflowID, lock.LockTimeout.Format(time.RFC3339))

	// Process workflow based on current state
	switch workflow.State {
	case types.WorkflowStateQueued:
		return e.processQueuedWorkflow(ctx, workflow)
	case types.WorkflowStateWorkspaceReady:
		return e.processWorkspaceReadyWorkflow(ctx, workflow)
	case types.WorkflowStateImplementing:
		return e.processImplementingWorkflow(ctx, workflow)
	case types.WorkflowStatePROpen:
		return e.processPROpenWorkflow(ctx, workflow)
	case types.WorkflowStateRevising:
		return e.processRevisingWorkflow(ctx, workflow)
	default:
		return fmt.Errorf("workflow %s is in terminal state %s", workflowID, workflow.State)
	}
}

// processQueuedWorkflow handles workflows in QUEUED state
func (e *Engine) processQueuedWorkflow(ctx context.Context, workflow *types.Workflow) error {
	log.Printf("📋 Processing queued workflow %d", workflow.ID)

	// Get the issue to understand what needs to be done
	issue, err := e.coworkProvider.GetIssue(ctx, e.owner, e.repo, workflow.IssueID)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Create or get the associated task
	task, err := e.createOrGetTask(workflow, issue)
	if err != nil {
		return fmt.Errorf("failed to create/get task: %w", err)
	}

	// Update workflow with task ID
	updateReq := &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		TaskID:     &task.ID,
	}
	_, err = e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to update workflow with task ID: %w", err)
	}

	// Create workspace
	workspace, err := e.createWorkspace(workflow, issue)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	// Update workflow with workspace information
	updateReq = &types.UpdateWorkflowRequest{
		WorkflowID:  workflow.ID,
		WorkspaceID: &workspace.ID,
		BranchName:  &workflow.BranchName,
	}
	_, err = e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to update workflow with workspace info: %w", err)
	}

	// Transition to WORKSPACE_READY
	state := types.WorkflowStateWorkspaceReady
	updateReq = &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		State:      &state,
	}
	_, err = e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to transition workflow to workspace_ready: %w", err)
	}

	log.Printf("✅ Workflow %d transitioned to workspace_ready", workflow.ID)
	return nil
}

// processWorkspaceReadyWorkflow handles workflows in WORKSPACE_READY state
func (e *Engine) processWorkspaceReadyWorkflow(ctx context.Context, workflow *types.Workflow) error {
	log.Printf("🏗️  Processing workspace ready workflow %d", workflow.ID)

	// Check if task is ready to be worked on
	task, err := e.taskManager.GetTask(fmt.Sprintf("%d", workflow.TaskID))
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// If task is not in progress, start it
	if task.Status != types.TaskStatusInProgress {
		// Update task status to in progress
		taskStatus := types.TaskStatusInProgress
		updateReq := &types.UpdateTaskRequest{
			TaskID: task.ID,
			Status: &taskStatus,
		}
		_, err = e.taskManager.UpdateTask(updateReq)
		if err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}
	}

	// Transition to IMPLEMENTING
	state := types.WorkflowStateImplementing
	updateReq := &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		State:      &state,
	}
	_, err = e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to transition workflow to implementing: %w", err)
	}

	log.Printf("✅ Workflow %d transitioned to implementing", workflow.ID)
	return nil
}

// processImplementingWorkflow handles workflows in IMPLEMENTING state
func (e *Engine) processImplementingWorkflow(ctx context.Context, workflow *types.Workflow) error {
	log.Printf("💻 Processing implementing workflow %d", workflow.ID)

	// Check task status
	task, err := e.taskManager.GetTask(fmt.Sprintf("%d", workflow.TaskID))
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// If task is completed, create PR
	if task.Status == types.TaskStatusCompleted {
		return e.createPullRequest(ctx, workflow, task)
	}

	// If task is failed, handle failure
	if task.Status == types.TaskStatusFailed {
		return e.handleTaskFailure(ctx, workflow, task)
	}

	// Task is still in progress, nothing to do
	log.Printf("⏳ Task %d is still in progress, waiting for completion", task.ID)
	return nil
}

// processPROpenWorkflow handles workflows in PR_OPEN state
func (e *Engine) processPROpenWorkflow(ctx context.Context, workflow *types.Workflow) error {
	log.Printf("🔍 Processing PR open workflow %d", workflow.ID)

	// Get the task to pass to the provider
	task, err := e.taskManager.GetTask(fmt.Sprintf("%d", workflow.TaskID))
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check for PR updates
	pr, err := e.coworkProvider.GetPullRequestForTask(ctx, task, e.owner, e.repo)
	if err != nil {
		return fmt.Errorf("failed to get pull request: %w", err)
	}

	if pr == nil {
		return fmt.Errorf("no pull request found for workflow %d", workflow.ID)
	}

	// Check if PR was merged or closed
	if pr.State == "closed" || pr.Merged {
		return e.handlePRCompleted(ctx, workflow, pr)
	}

	// Check for new feedback
	updates, err := e.coworkProvider.GetPullRequestUpdates(ctx, e.owner, e.repo, pr.Number, workflow.LastEventTS)
	if err != nil {
		return fmt.Errorf("failed to get PR updates: %w", err)
	}

	// If there are updates, handle them
	if len(updates.NewComments) > 0 || len(updates.NewReviews) > 0 {
		return e.handlePRFeedback(ctx, workflow, pr, updates)
	}

	log.Printf("⏳ No updates for PR #%d, waiting for feedback", pr.Number)
	return nil
}

// processRevisingWorkflow handles workflows in REVISING state
func (e *Engine) processRevisingWorkflow(ctx context.Context, workflow *types.Workflow) error {
	log.Printf("🔄 Processing revising workflow %d", workflow.ID)

	// Get the workspace
	workspace, err := e.workspaceManager.GetWorkspace(workflow.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Sync with base branch
	err = e.syncWithBaseBranch(workspace.Path, workflow.BaseBranch)
	if err != nil {
		return fmt.Errorf("failed to sync with base branch: %w", err)
	}

	// Update task to trigger agent work
	taskStatus := types.TaskStatusInProgress
	updateReq := &types.UpdateTaskRequest{
		TaskID: workflow.TaskID,
		Status: &taskStatus,
	}
	_, err = e.taskManager.UpdateTask(updateReq)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Transition back to PR_OPEN
	state := types.WorkflowStatePROpen
	workflowUpdateReq := &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		State:      &state,
	}
	_, err = e.workflowManager.UpdateWorkflow(workflowUpdateReq)
	if err != nil {
		return fmt.Errorf("failed to transition workflow to pr_open: %w", err)
	}

	log.Printf("✅ Workflow %d transitioned back to pr_open for revision", workflow.ID)
	return nil
}

// createOrGetTask creates a task for the workflow or gets existing one
func (e *Engine) createOrGetTask(workflow *types.Workflow, issue *git.Issue) (*types.Task, error) {
	// If workflow already has a task ID, get that task
	if workflow.TaskID != 0 {
		task, err := e.taskManager.GetTask(fmt.Sprintf("%d", workflow.TaskID))
		if err == nil {
			return task, nil
		}
		// Task not found, continue to create new one
	}

	// Create new task from issue
	task, err := e.coworkProvider.CreateTaskFromIssue(context.Background(), e.owner, e.repo, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to create task from issue: %w", err)
	}

	return task, nil
}

// createWorkspace creates a workspace for the workflow
func (e *Engine) createWorkspace(workflow *types.Workflow, issue *git.Issue) (*types.Workspace, error) {
	// Generate branch name
	branchName := e.coworkProvider.GenerateBranchName(issue)
	workflow.BranchName = branchName

	// Get the task to pass to the provider
	task, err := e.taskManager.GetTask(fmt.Sprintf("%d", workflow.TaskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Create workspace
	workspace, err := e.coworkProvider.CreateWorkspaceForTask(context.Background(), task, e.owner, e.repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Set up the branch
	err = e.setupBranch(workspace.Path, workflow.BaseBranch, branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to set up branch: %w", err)
	}

	return workspace, nil
}

// setupBranch sets up the feature branch for the workflow
func (e *Engine) setupBranch(workspacePath, baseBranch, branchName string) error {
	// Change to workspace directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(workspacePath); err != nil {
		return fmt.Errorf("failed to change to workspace directory: %w", err)
	}

	// Fetch latest changes
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch origin: %w", err)
	}

	// Checkout base branch
	cmd = exec.Command("git", "checkout", baseBranch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout base branch: %w", err)
	}

	// Pull latest changes
	cmd = exec.Command("git", "pull", "origin", baseBranch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}

	// Create and checkout feature branch
	cmd = exec.Command("git", "checkout", "-b", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create feature branch: %w", err)
	}

	return nil
}

// createPullRequest creates a pull request for the completed task
func (e *Engine) createPullRequest(ctx context.Context, workflow *types.Workflow, task *types.Task) error {
	log.Printf("📝 Creating pull request for workflow %d", workflow.ID)

	// Get the workspace
	workspace, err := e.workspaceManager.GetWorkspace(workflow.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Push the branch
	err = e.pushBranch(workspace.Path, workflow.BranchName)
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	// Create pull request
	pr, err := e.coworkProvider.CreatePullRequestForTask(ctx, task, e.owner, e.repo, workspace)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	// Update workflow with PR information
	updateReq := &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		PRNumber:   &pr.Number,
	}
	_, err = e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to update workflow with PR number: %w", err)
	}

	// Transition to PR_OPEN
	state := types.WorkflowStatePROpen
	updateReq = &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		State:      &state,
	}
	_, err = e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to transition workflow to pr_open: %w", err)
	}

	log.Printf("✅ Created PR #%d for workflow %d", pr.Number, workflow.ID)
	return nil
}

// pushBranch pushes the feature branch to origin
func (e *Engine) pushBranch(workspacePath, branchName string) error {
	// Change to workspace directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(workspacePath); err != nil {
		return fmt.Errorf("failed to change to workspace directory: %w", err)
	}

	// Push branch to origin
	cmd := exec.Command("git", "push", "-u", "origin", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	return nil
}

// handleTaskFailure handles task failures
func (e *Engine) handleTaskFailure(ctx context.Context, workflow *types.Workflow, task *types.Task) error {
	log.Printf("❌ Handling task failure for workflow %d", workflow.ID)

	// Transition to ABORTED
	state := types.WorkflowStateAborted
	updateReq := &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		State:      &state,
		LastError:  &task.ErrorMessage,
	}
	_, err := e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to transition workflow to aborted: %w", err)
	}

	log.Printf("✅ Workflow %d aborted due to task failure", workflow.ID)
	return nil
}

// handlePRCompleted handles when a PR is merged or closed
func (e *Engine) handlePRCompleted(ctx context.Context, workflow *types.Workflow, pr *git.PullRequest) error {
	log.Printf("🏁 Handling PR completion for workflow %d", workflow.ID)

	var state types.WorkflowState
	if pr.Merged {
		state = types.WorkflowStateMerged
		log.Printf("✅ PR #%d was merged", pr.Number)
	} else {
		state = types.WorkflowStateClosed
		log.Printf("❌ PR #%d was closed without merging", pr.Number)
	}

	// Transition to terminal state
	updateReq := &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		State:      &state,
	}
	_, err := e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to transition workflow to %s: %w", state, err)
	}

	// Clean up workspace if merged
	if pr.Merged {
		err = e.cleanupWorkspace(workflow)
		if err != nil {
			log.Printf("⚠️  Failed to cleanup workspace: %v", err)
		}
	}

	log.Printf("✅ Workflow %d completed with state %s", workflow.ID, state)
	return nil
}

// handlePRFeedback handles feedback on a pull request
func (e *Engine) handlePRFeedback(ctx context.Context, workflow *types.Workflow, pr *git.PullRequest, updates *git.PullRequestUpdate) error {
	log.Printf("💬 Handling PR feedback for workflow %d", workflow.ID)

	// Classify feedback intent
	intent := e.classifyFeedbackIntent(updates)

	switch intent {
	case types.FeedbackIntentAsk:
		// Handle questions/clarifications
		return e.handleAskFeedback(ctx, workflow, pr, updates)
	case types.FeedbackIntentChange, types.FeedbackIntentBlocker:
		// Handle requested changes
		return e.handleChangeFeedback(ctx, workflow, pr, updates)
	default:
		log.Printf("ℹ️  No actionable feedback found for workflow %d", workflow.ID)
		return nil
	}
}

// classifyFeedbackIntent classifies the intent of feedback
func (e *Engine) classifyFeedbackIntent(updates *git.PullRequestUpdate) types.FeedbackIntent {
	// Check for blocker conditions first
	for _, review := range updates.NewReviews {
		if review.State == "changes_requested" {
			return types.FeedbackIntentBlocker
		}
	}

	// Check for change requests in comments
	for _, comment := range updates.NewComments {
		body := strings.ToLower(comment.Body)
		if strings.Contains(body, "change") || strings.Contains(body, "fix") || strings.Contains(body, "update") {
			return types.FeedbackIntentChange
		}
	}

	// Default to ask for questions/clarifications
	return types.FeedbackIntentAsk
}

// handleAskFeedback handles questions and clarifications
func (e *Engine) handleAskFeedback(ctx context.Context, workflow *types.Workflow, pr *git.PullRequest, updates *git.PullRequestUpdate) error {
	log.Printf("❓ Handling ask feedback for workflow %d", workflow.ID)

	// For now, just log the questions
	// In the future, this could trigger agent responses
	for _, comment := range updates.NewComments {
		log.Printf("Question on PR #%d: %s", pr.Number, comment.Body)
	}

	return nil
}

// handleChangeFeedback handles requested changes
func (e *Engine) handleChangeFeedback(ctx context.Context, workflow *types.Workflow, pr *git.PullRequest, updates *git.PullRequestUpdate) error {
	log.Printf("🔧 Handling change feedback for workflow %d", workflow.ID)

	// Transition to REVISING
	state := types.WorkflowStateRevising
	updateReq := &types.UpdateWorkflowRequest{
		WorkflowID: workflow.ID,
		State:      &state,
	}
	_, err := e.workflowManager.UpdateWorkflow(updateReq)
	if err != nil {
		return fmt.Errorf("failed to transition workflow to revising: %w", err)
	}

	log.Printf("✅ Workflow %d transitioned to revising for feedback", workflow.ID)
	return nil
}

// syncWithBaseBranch syncs the feature branch with the base branch
func (e *Engine) syncWithBaseBranch(workspacePath, baseBranch string) error {
	// Change to workspace directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(workspacePath); err != nil {
		return fmt.Errorf("failed to change to workspace directory: %w", err)
	}

	// Fetch latest changes
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch origin: %w", err)
	}

	// Try rebase first
	cmd = exec.Command("git", "rebase", "origin/"+baseBranch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// If rebase fails, try merge
		log.Printf("⚠️  Rebase failed, trying merge")
		cmd = exec.Command("git", "merge", "origin/"+baseBranch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to sync with base branch: %w", err)
		}
	}

	return nil
}

// cleanupWorkspace cleans up the workspace after successful merge
func (e *Engine) cleanupWorkspace(workflow *types.Workflow) error {
	log.Printf("🧹 Cleaning up workspace for workflow %d", workflow.ID)

	// Delete the workspace
	err := e.workspaceManager.DeleteWorkspace(workflow.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	// Delete the remote branch
	err = e.deleteRemoteBranch(workflow.BranchName)
	if err != nil {
		log.Printf("⚠️  Failed to delete remote branch: %v", err)
	}

	return nil
}

// deleteRemoteBranch deletes the remote branch
func (e *Engine) deleteRemoteBranch(branchName string) error {
	// This would require authentication and proper Git operations
	// For now, just log the intention
	log.Printf("🗑️  Would delete remote branch: %s", branchName)
	return nil
}
