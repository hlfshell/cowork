package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hlfshell/cowork/internal/auth"
	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
	gitprovider "github.com/hlfshell/cowork/internal/git/providers"
	"github.com/hlfshell/cowork/internal/types"
	"github.com/hlfshell/cowork/internal/workflow"
	"github.com/spf13/cobra"
)

// addWorkflowCommands adds workflow management commands
func (app *App) addWorkflowCommands() {
	workflowCmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage auto-PR workflows",
		Long:  "Create, manage, and monitor auto-PR workflows that automatically handle issues, tasks, and pull requests",
	}

	// Scan and create tasks command
	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan open issues and create tasks",
		Long:  "Scan all open issues assigned to the current user and create tasks for them in the current project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.scanAndCreateTasks(cmd)
		},
	}

	// Process queued tasks command
	processCmd := &cobra.Command{
		Use:   "process",
		Short: "Process queued tasks and create workspaces",
		Long:  "Process queued tasks by creating workspaces and updating their status in the current project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.processQueuedTasks(cmd)
		},
	}

	// Complete task and create PR command
	completeCmd := &cobra.Command{
		Use:   "complete [task-id]",
		Short: "Complete a task and create a pull request",
		Long:  "Mark a task as completed and create a pull request for the changes in the current project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.completeTaskAndCreatePR(cmd, args[0])
		},
	}

	// Scan PRs for updates command
	scanPRsCmd := &cobra.Command{
		Use:   "scan-prs",
		Short: "Scan pull requests for updates",
		Long:  "Scan pull requests associated with tasks and process any updates in the current project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.scanPullRequestsForUpdates(cmd)
		},
	}

	// Sync task statuses command
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync task statuses to provider",
		Long:  "Sync task statuses back to the Git provider (update issue labels, etc.) in the current project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.syncTaskStatuses(cmd)
		},
	}

	// Run full workflow command
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the complete workflow",
		Long:  "Run the complete workflow: scan issues, process tasks, scan PRs, and sync statuses in the current project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runFullWorkflow(cmd)
		},
	}

	// Add provider flag to commands that need it
	providerFlag := "provider"
	for _, cmd := range []*cobra.Command{scanCmd, processCmd, completeCmd, scanPRsCmd, syncCmd, runCmd} {
		cmd.Flags().String(providerFlag, "github", "Git provider (github, gitlab, bitbucket)")
	}

	// New workflow management commands
	// Create workflow command
	createCmd := &cobra.Command{
		Use:   "create [issue-id]",
		Short: "Create a new workflow from an issue",
		Long:  "Create a new auto-PR workflow for a specific issue in the current project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.createWorkflow(cmd, args[0])
		},
	}

	// Create workflow from task command
	createFromTaskCmd := &cobra.Command{
		Use:   "create-from-task [task-id]",
		Short: "Create a workflow from an existing task",
		Long:  "Create a new auto-PR workflow from an existing task in the current project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.createWorkflowFromTask(cmd, args[0])
		},
	}

	// List workflows command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflows",
		Long:  "List all workflows with their current states",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listWorkflows(cmd)
		},
	}

	// Show workflow command
	showCmd := &cobra.Command{
		Use:   "show [workflow-id]",
		Short: "Show workflow details",
		Long:  "Show detailed information about a specific workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.showWorkflow(cmd, args[0])
		},
	}

	// Process workflow command
	processWorkflowCmd := &cobra.Command{
		Use:   "process-workflow [workflow-id]",
		Short: "Process a workflow",
		Long:  "Process a workflow through its current state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.processWorkflow(cmd, args[0])
		},
	}

	// Run workflows command (continuous processing)
	runWorkflowsCmd := &cobra.Command{
		Use:   "run-workflows",
		Short: "Run workflows continuously",
		Long:  "Run workflows continuously for the current project, processing events and state transitions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runWorkflows(cmd)
		},
	}

	// Lock management commands
	lockCmd := &cobra.Command{
		Use:   "lock",
		Short: "Manage workflow locks",
		Long:  "Manage workflow locks and handle stuck workflows",
	}

	// List locks command
	listLocksCmd := &cobra.Command{
		Use:   "list",
		Short: "List active workflow locks",
		Long:  "List all active workflow locks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listWorkflowLocks(cmd)
		},
	}

	// Force unlock command
	forceUnlockCmd := &cobra.Command{
		Use:   "force-unlock [workflow-id]",
		Short: "Force unlock a workflow",
		Long:  "Force unlock a workflow (admin override)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.forceUnlockWorkflow(cmd, args[0])
		},
	}

	// Cleanup expired locks command
	cleanupLocksCmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup expired locks",
		Long:  "Cleanup expired workflow locks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.cleanupExpiredLocks(cmd)
		},
	}

	// Add provider flag to new commands
	for _, cmd := range []*cobra.Command{createCmd, createFromTaskCmd, runWorkflowsCmd} {
		cmd.Flags().String(providerFlag, "github", "Git provider (github, gitlab, bitbucket)")
	}

	// Add base branch flag
	baseBranchFlag := "base-branch"
	for _, cmd := range []*cobra.Command{createCmd, createFromTaskCmd} {
		cmd.Flags().String(baseBranchFlag, "main", "Base branch for the workflow")
	}

	// Add flags to run workflows command
	runWorkflowsCmd.Flags().Duration("poll-interval", 30*time.Second, "Polling interval for workflow processing")
	runWorkflowsCmd.Flags().Bool("once", false, "Process workflows once and exit")

	// Add commands to lock subcommand
	lockCmd.AddCommand(listLocksCmd, forceUnlockCmd, cleanupLocksCmd)

	// Add all commands to workflow command
	workflowCmd.AddCommand(
		scanCmd, processCmd, completeCmd, scanPRsCmd, syncCmd, runCmd,
		createCmd, createFromTaskCmd, listCmd, showCmd, processWorkflowCmd, runWorkflowsCmd, lockCmd,
	)
	app.rootCmd.AddCommand(workflowCmd)
}

// parseOwnerRepo parses owner/repo string into separate components
func (app *App) parseOwnerRepo(ownerRepo string) (string, string, error) {
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid owner/repo format: %s (expected owner/repo)", ownerRepo)
	}
	return parts[0], parts[1], nil
}

// getWorkflowManager creates a workflow manager for the current project
func (app *App) getWorkflowManager() (*workflow.WorkflowManager, error) {
	// Get current directory and check for .cw initialization
	cwDir := ".cw"
	if _, err := os.Stat(cwDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("cowork project not initialized in current directory. Run 'cowork init' first")
	}

	// Create workflow manager
	workflowManager, err := workflow.NewWorkflowManager(cwDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow manager: %w", err)
	}

	return workflowManager, nil
}

// getCurrentProjectInfo determines the current project's owner/repo from Git
func (app *App) getCurrentProjectInfo() (string, string, error) {
	// Check if we're in a Git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return "", "", fmt.Errorf("current directory is not a Git repository")
	}

	// Get the remote origin URL
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get Git remote origin: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse the remote URL to extract owner/repo
	// Handle different URL formats:
	// - https://github.com/owner/repo.git
	// - git@github.com:owner/repo.git
	// - https://gitlab.com/owner/repo.git
	// - git@gitlab.com:owner/repo.git

	var owner, repo string

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.Contains(remoteURL, "@") && strings.Contains(remoteURL, ":") {
		parts := strings.Split(remoteURL, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH remote URL format: %s", remoteURL)
		}
		repoPart := strings.TrimSuffix(parts[1], ".git")
		ownerRepo := strings.Split(repoPart, "/")
		if len(ownerRepo) != 2 {
			return "", "", fmt.Errorf("invalid repository format in SSH URL: %s", remoteURL)
		}
		owner = ownerRepo[0]
		repo = ownerRepo[1]
	} else {
		// Handle HTTPS format: https://github.com/owner/repo.git
		parts := strings.Split(remoteURL, "/")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid HTTPS remote URL format: %s", remoteURL)
		}
		repoPart := strings.TrimSuffix(parts[len(parts)-1], ".git")
		owner = parts[len(parts)-2]
		repo = repoPart
	}

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("failed to parse owner/repo from remote URL: %s", remoteURL)
	}

	return owner, repo, nil
}

// createCoworkProvider creates a cowork provider for the specified provider type
func (app *App) createCoworkProvider(providerType string) (git.CoworkProvider, error) {
	// Create auth manager
	configManager := config.NewManager()
	authManager, err := auth.NewManager(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Parse provider type
	var gitProviderType git.ProviderType
	switch providerType {
	case "github":
		gitProviderType = git.ProviderGitHub
	case "gitlab":
		gitProviderType = git.ProviderGitLab
	case "bitbucket":
		gitProviderType = git.ProviderBitbucket
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}

	// Try to get authentication - first project-specific, then global
	var authConfig *auth.AuthConfig
	var token string
	var baseURL string

	// Try project-specific authentication first
	authConfig, err = authManager.GetAuthConfig(gitProviderType, auth.AuthScopeProject)
	if err == nil && authConfig != nil {
		token = authConfig.Token
		baseURL = authConfig.BaseURL
	} else {
		// Fall back to global authentication
		authConfig, err = authManager.GetAuthConfig(gitProviderType, auth.AuthScopeGlobal)
		if err != nil {
			return nil, fmt.Errorf("no authentication configured for %s provider. Run 'cw auth provider login %s' first", providerType, providerType)
		}
		token = authConfig.Token
		baseURL = authConfig.BaseURL
	}

	// Set default base URL if not provided
	if baseURL == "" {
		switch providerType {
		case "github":
			baseURL = "https://api.github.com"
		case "gitlab":
			baseURL = "https://gitlab.com/api/v4"
		case "bitbucket":
			baseURL = "https://api.bitbucket.org/2.0"
		}
	}

	// Create the appropriate cowork provider
	switch providerType {
	case "github":
		return gitprovider.NewGitHubCoworkProvider(token, baseURL, app.taskManager, app.workspaceManager)
	case "gitlab":
		// TODO: Implement GitLab cowork provider
		return nil, fmt.Errorf("GitLab cowork provider not yet implemented")
	case "bitbucket":
		// TODO: Implement Bitbucket cowork provider
		return nil, fmt.Errorf("Bitbucket cowork provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}
}

// scanAndCreateTasks scans open issues and creates tasks
func (app *App) scanAndCreateTasks(cmd *cobra.Command) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	providerType, _ := cmd.Flags().GetString("provider")
	coworkProvider, err := app.createCoworkProvider(providerType)
	if err != nil {
		return err
	}

	workflowManager := workflow.NewManager(coworkProvider, app.taskManager, app.workspaceManager, owner, repo)

	ctx := context.Background()
	return workflowManager.ScanAndCreateTasks(ctx)
}

// processQueuedTasks processes queued tasks and creates workspaces
func (app *App) processQueuedTasks(cmd *cobra.Command) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	providerType, _ := cmd.Flags().GetString("provider")
	coworkProvider, err := app.createCoworkProvider(providerType)
	if err != nil {
		return err
	}

	workflowManager := workflow.NewManager(coworkProvider, app.taskManager, app.workspaceManager, owner, repo)

	ctx := context.Background()
	return workflowManager.ProcessQueuedTasks(ctx)
}

// completeTaskAndCreatePR completes a task and creates a pull request
func (app *App) completeTaskAndCreatePR(cmd *cobra.Command, taskID string) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	providerType, _ := cmd.Flags().GetString("provider")
	coworkProvider, err := app.createCoworkProvider(providerType)
	if err != nil {
		return err
	}

	workflowManager := workflow.NewManager(coworkProvider, app.taskManager, app.workspaceManager, owner, repo)

	ctx := context.Background()
	return workflowManager.CompleteTaskAndCreatePR(ctx, taskID)
}

// scanPullRequestsForUpdates scans pull requests for updates
func (app *App) scanPullRequestsForUpdates(cmd *cobra.Command) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	providerType, _ := cmd.Flags().GetString("provider")
	coworkProvider, err := app.createCoworkProvider(providerType)
	if err != nil {
		return err
	}

	workflowManager := workflow.NewManager(coworkProvider, app.taskManager, app.workspaceManager, owner, repo)

	ctx := context.Background()
	return workflowManager.ScanPullRequestsForUpdates(ctx)
}

// syncTaskStatuses syncs task statuses to the provider
func (app *App) syncTaskStatuses(cmd *cobra.Command) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	providerType, _ := cmd.Flags().GetString("provider")
	coworkProvider, err := app.createCoworkProvider(providerType)
	if err != nil {
		return err
	}

	workflowManager := workflow.NewManager(coworkProvider, app.taskManager, app.workspaceManager, owner, repo)

	ctx := context.Background()
	return workflowManager.SyncTaskStatuses(ctx)
}

// runFullWorkflow runs the complete workflow
func (app *App) runFullWorkflow(cmd *cobra.Command) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	providerType, _ := cmd.Flags().GetString("provider")
	coworkProvider, err := app.createCoworkProvider(providerType)
	if err != nil {
		return err
	}

	workflowManager := workflow.NewManager(coworkProvider, app.taskManager, app.workspaceManager, owner, repo)

	ctx := context.Background()
	return workflowManager.RunFullWorkflow(ctx)
}

// createWorkflow creates a new workflow from an issue
func (app *App) createWorkflow(cmd *cobra.Command, issueIDStr string) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	// Parse issue ID
	var issueID int
	_, err = fmt.Sscanf(issueIDStr, "%d", &issueID)
	if err != nil {
		return fmt.Errorf("invalid issue ID: %s", issueIDStr)
	}

	// Get flags
	providerType, _ := cmd.Flags().GetString("provider")
	baseBranch, _ := cmd.Flags().GetString("base-branch")

	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// Create workflow request
	req := &types.CreateWorkflowRequest{
		Owner:      owner,
		Repo:       repo,
		IssueID:    issueID,
		BaseBranch: baseBranch,
		Provider:   providerType,
		Config:     types.GetDefaultWorkflowConfig(),
	}

	// Create workflow
	workflow, err := workflowManager.CreateWorkflow(req)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	cmd.Printf("✅ Created workflow %d for issue %s/%s#%d\n", workflow.ID, owner, repo, issueID)
	cmd.Printf("   State: %s\n", workflow.State)
	cmd.Printf("   Provider: %s\n", workflow.Provider)
	cmd.Printf("   Base Branch: %s\n", workflow.BaseBranch)

	return nil
}

// createWorkflowFromTask creates a workflow from an existing task
func (app *App) createWorkflowFromTask(cmd *cobra.Command, taskID string) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	// Get the task
	task, err := app.taskManager.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Extract issue ID from task ticket ID
	// Expected format: "github:owner/repo#123"
	parts := strings.Split(task.TicketID, "#")
	if len(parts) != 2 {
		return fmt.Errorf("task %s does not have a valid ticket ID format", taskID)
	}

	var issueID int
	_, err = fmt.Sscanf(parts[1], "%d", &issueID)
	if err != nil {
		return fmt.Errorf("failed to parse issue ID from ticket ID: %w", err)
	}

	// Get flags
	providerType, _ := cmd.Flags().GetString("provider")
	baseBranch, _ := cmd.Flags().GetString("base-branch")

	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// Convert taskID string to int
	taskIDInt, err := strconv.Atoi(taskID)
	if err != nil {
		return fmt.Errorf("invalid task ID: %s", taskID)
	}

	// Create workflow request
	req := &types.CreateWorkflowRequest{
		Owner:      owner,
		Repo:       repo,
		IssueID:    issueID,
		BaseBranch: baseBranch,
		Provider:   providerType,
		Config:     types.GetDefaultWorkflowConfig(),
		TaskID:     taskIDInt,
	}

	// Create workflow
	workflow, err := workflowManager.CreateWorkflow(req)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	cmd.Printf("✅ Created workflow %d from task %s\n", workflow.ID, taskID)
	cmd.Printf("   State: %s\n", workflow.State)
	cmd.Printf("   Provider: %s\n", workflow.Provider)
	cmd.Printf("   Base Branch: %s\n", workflow.BaseBranch)

	return nil
}

// listWorkflows lists all workflows
func (app *App) listWorkflows(cmd *cobra.Command) error {
	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// Get workflows
	workflows, err := workflowManager.ListWorkflows()
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	if len(workflows) == 0 {
		cmd.Printf("No workflows found.\n")
		return nil
	}

	cmd.Printf("Workflows (%d):\n", len(workflows))
	cmd.Printf("%-36s %-20s %-15s %-10s %-15s\n", "ID", "Repository", "Issue", "State", "Provider")
	cmd.Printf("%s\n", strings.Repeat("-", 100))

	for _, w := range workflows {
		repo := fmt.Sprintf("%s/%s", w.Owner, w.Repo)
		issue := fmt.Sprintf("#%d", w.IssueID)
		cmd.Printf("%-36s %-20s %-15s %-10s %-15s\n", fmt.Sprintf("%d", w.ID), repo, issue, w.State, w.Provider)
	}

	return nil
}

// showWorkflow shows detailed information about a workflow
func (app *App) showWorkflow(cmd *cobra.Command, workflowID string) error {
	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// Get workflow
	workflow, err := workflowManager.GetWorkflow(workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Check if workflow is locked
	isLocked, lock := workflowManager.IsWorkflowLocked(fmt.Sprintf("%d", workflow.ID))

	cmd.Printf("Workflow: %d\n", workflow.ID)
	cmd.Printf("Repository: %s/%s\n", workflow.Owner, workflow.Repo)
	cmd.Printf("Issue: #%d\n", workflow.IssueID)
	cmd.Printf("State: %s\n", workflow.State)
	cmd.Printf("Provider: %s\n", workflow.Provider)
	cmd.Printf("Base Branch: %s\n", workflow.BaseBranch)

	if workflow.BranchName != "" {
		cmd.Printf("Feature Branch: %s\n", workflow.BranchName)
	}

	if workflow.PRNumber != nil {
		cmd.Printf("Pull Request: #%d\n", *workflow.PRNumber)
	}

	if workflow.TaskID != 0 {
		cmd.Printf("Task ID: %d\n", workflow.TaskID)
	}

	if workflow.WorkspaceID != 0 {
		cmd.Printf("Workspace ID: %d\n", workflow.WorkspaceID)
	}

	cmd.Printf("Created: %s\n", workflow.CreatedAt.Format(time.RFC3339))
	cmd.Printf("Updated: %s\n", workflow.UpdatedAt.Format(time.RFC3339))

	if workflow.StartedAt != nil {
		cmd.Printf("Started: %s\n", workflow.StartedAt.Format(time.RFC3339))
	}

	if workflow.EndedAt != nil {
		cmd.Printf("Ended: %s\n", workflow.EndedAt.Format(time.RFC3339))
	}

	if workflow.ErrorCount > 0 {
		cmd.Printf("Error Count: %d\n", workflow.ErrorCount)
		if workflow.LastError != "" {
			cmd.Printf("Last Error: %s\n", workflow.LastError)
		}
	}

	if isLocked && lock != nil {
		cmd.Printf("Locked: Yes (by %s until %s)\n", lock.LockedBy, lock.LockTimeout.Format(time.RFC3339))
	} else {
		cmd.Printf("Locked: No\n")
	}

	return nil
}

// processWorkflow processes a single workflow
func (app *App) processWorkflow(cmd *cobra.Command, workflowID string) error {
	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// Get workflow
	workflowObj, err := workflowManager.GetWorkflow(workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Create cowork provider
	coworkProvider, err := app.createCoworkProvider(workflowObj.Provider)
	if err != nil {
		return err
	}

	// Create workflow engine
	engine := workflow.NewEngine(workflowManager, app.taskManager, app.workspaceManager, coworkProvider, workflowObj.Owner, workflowObj.Repo)

	// Process workflow
	ctx := context.Background()
	err = engine.ProcessWorkflow(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to process workflow: %w", err)
	}

	cmd.Printf("✅ Successfully processed workflow %s\n", workflowID)
	return nil
}

// runWorkflows runs workflows continuously
func (app *App) runWorkflows(cmd *cobra.Command) error {
	// Get current project info
	owner, repo, err := app.getCurrentProjectInfo()
	if err != nil {
		return fmt.Errorf("failed to get current project info: %w", err)
	}

	// Get flags
	providerType, _ := cmd.Flags().GetString("provider")
	pollInterval, _ := cmd.Flags().GetDuration("poll-interval")
	once, _ := cmd.Flags().GetBool("once")

	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// Create cowork provider
	coworkProvider, err := app.createCoworkProvider(providerType)
	if err != nil {
		return err
	}

	// Create workflow engine
	engine := workflow.NewEngine(workflowManager, app.taskManager, app.workspaceManager, coworkProvider, owner, repo)

	cmd.Printf("🚀 Starting workflow runner for %s/%s\n", owner, repo)
	cmd.Printf("   Provider: %s\n", providerType)
	cmd.Printf("   Poll Interval: %s\n", pollInterval)
	cmd.Printf("   Mode: %s\n", map[bool]string{true: "once", false: "continuous"}[once])

	ctx := context.Background()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		// Get workflows for this repository
		workflows, err := workflowManager.ListWorkflows()
		if err != nil {
			cmd.Printf("❌ Failed to list workflows: %v\n", err)
			continue
		}

		// Filter workflows for this repository
		var relevantWorkflows []*types.Workflow
		for _, w := range workflows {
			if w.Owner == owner && w.Repo == repo && !w.State.IsTerminal() {
				relevantWorkflows = append(relevantWorkflows, w)
			}
		}

		if len(relevantWorkflows) > 0 {
			cmd.Printf("📋 Processing %d workflows...\n", len(relevantWorkflows))
		}

		// Process each workflow
		for _, w := range relevantWorkflows {
			// Check if workflow is locked
			isLocked, lock := workflowManager.IsWorkflowLocked(fmt.Sprintf("%d", w.ID))
			if isLocked {
				cmd.Printf("⏳ Workflow %d is locked by %s until %s\n", w.ID, lock.LockedBy, lock.LockTimeout.Format(time.RFC3339))
				continue
			}

			cmd.Printf("🔄 Processing workflow %d (state: %s)\n", w.ID, w.State)
			err := engine.ProcessWorkflow(ctx, fmt.Sprintf("%d", w.ID))
			if err != nil {
				cmd.Printf("❌ Failed to process workflow %d: %v\n", w.ID, err)
			} else {
				cmd.Printf("✅ Successfully processed workflow %d\n", w.ID)
			}
		}

		if once {
			break
		}

		// Wait for next tick
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// listWorkflowLocks lists active workflow locks
func (app *App) listWorkflowLocks(cmd *cobra.Command) error {
	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// Get all workflows
	workflows, err := workflowManager.ListWorkflows()
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	var lockedWorkflows []*types.Workflow
	for _, w := range workflows {
		isLocked, _ := workflowManager.IsWorkflowLocked(fmt.Sprintf("%d", w.ID))
		if isLocked {
			lockedWorkflows = append(lockedWorkflows, w)
		}
	}

	if len(lockedWorkflows) == 0 {
		cmd.Printf("No locked workflows found.\n")
		return nil
	}

	cmd.Printf("Locked Workflows (%d):\n", len(lockedWorkflows))
	cmd.Printf("%-36s %-20s %-15s %-10s %-20s %-20s\n", "ID", "Repository", "Issue", "State", "Locked By", "Lock Until")
	cmd.Printf("%s\n", strings.Repeat("-", 120))

	for _, w := range lockedWorkflows {
		_, lock := workflowManager.IsWorkflowLocked(fmt.Sprintf("%d", w.ID))
		repo := fmt.Sprintf("%s/%s", w.Owner, w.Repo)
		issue := fmt.Sprintf("#%d", w.IssueID)
		lockedBy := lock.LockedBy
		lockUntil := lock.LockTimeout.Format(time.RFC3339)
		cmd.Printf("%-36s %-20s %-15s %-10s %-20s %-20s\n", fmt.Sprintf("%d", w.ID), repo, issue, w.State, lockedBy, lockUntil)
	}

	return nil
}

// forceUnlockWorkflow forcefully unlocks a workflow
func (app *App) forceUnlockWorkflow(cmd *cobra.Command, workflowID string) error {
	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// Force unlock workflow
	err = workflowManager.ForceUnlockWorkflow(workflowID)
	if err != nil {
		return fmt.Errorf("failed to force unlock workflow: %w", err)
	}

	cmd.Printf("✅ Force unlocked workflow %s\n", workflowID)
	return nil
}

// cleanupExpiredLocks cleans up expired workflow locks
func (app *App) cleanupExpiredLocks(cmd *cobra.Command) error {
	// Create workflow manager
	workflowManager, err := app.getWorkflowManager()
	if err != nil {
		return err
	}
	defer workflowManager.Close()

	// The cleanup is handled automatically by the watchdog timer
	// This command just triggers a manual cleanup
	cmd.Printf("🧹 Cleanup is handled automatically by the watchdog timer.\n")
	cmd.Printf("   Check interval: %s\n", workflow.WatchdogInterval)
	cmd.Printf("   Lock timeout: %s\n", workflow.DefaultLockTimeout)

	return nil
}
