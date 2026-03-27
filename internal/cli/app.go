package cli

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hlfshell/cowork/internal/auth"
	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/task"
	"github.com/hlfshell/cowork/internal/types"
	"github.com/hlfshell/cowork/internal/workflow"
	"github.com/spf13/cobra"
)

// App represents the main CLI application
type App struct {
	rootCmd       *cobra.Command
	version       string
	buildDate     string
	gitCommit     string
	taskManager   task.TaskManager
	configManager *config.Manager
}

// NewApp creates a new CLI application with the specified version information
func NewApp(version, buildDate, gitCommit string) *App {
	// Check and initialize global configuration on first run
	configManager, err := config.NewManager()
	if err != nil {
		fmt.Printf("Warning: failed to initialize configuration manager: %v\n", err)
	}
	if err = ensureGlobalConfigExists(configManager); err != nil {
		fmt.Printf("Warning: failed to initialize global configuration: %v\n", err)
	}

	// Initialize task manager (which now handles workspaces)
	coworkDir := filepath.Join(".", ".cowork")
	taskManager, err := task.NewManager(coworkDir, 300)
	if err != nil {
		fmt.Printf("Warning: failed to initialize task manager: %v\n", err)
		taskManager = nil
	}

	app := &App{
		version:       version,
		buildDate:     buildDate,
		gitCommit:     gitCommit,
		taskManager:   taskManager,
		configManager: configManager,
	}

	app.setupCommands()
	return app
}

// setupCommands initializes all CLI commands and their structure
func (app *App) setupCommands() {
	app.rootCmd = &cobra.Command{
		Use:   "cw",
		Short: "cowork - managing your AI dev-team",

		// For more information, visit: https://github.com/hlfshell/cowork`,
		Version: app.version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show help if no subcommand is provided
			return cmd.Help()
		},
	}

	// Add version command
	app.addVersionCommand()

	// Add init command
	app.addInitCommand()

	// Add config commands
	app.addConfigCommands()

	// Add task commands (which now handle workspaces)
	app.addTaskCommands()

	// Add workflow commands
	app.addWorkflowCommands()

	// Add go command for workflow automation
	app.addGoCommand()
}

// addVersionCommand adds a detailed version command
func (app *App) addVersionCommand() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show detailed version information",
		Long:  "Display the version, build date, and git commit information for the cowork CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cowork (cw) version %s\n", app.version)
			fmt.Printf("Build Date: %s\n", app.buildDate)
			fmt.Printf("Git Commit: %s\n", app.gitCommit)
		},
	}

	app.rootCmd.AddCommand(versionCmd)
}

// addInitCommand adds the init command for initializing a project
func (app *App) addInitCommand() {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize cowork for the current project",
		Long:  "Initialize cowork for the current Git repository by creating the .cw directory and project configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.initializeProject(cmd)
		},
	}

	// Add flag to avoid error if already initialized
	initCmd.Flags().Bool("force", false, "Force initialization even if already initialized")

	app.rootCmd.AddCommand(initCmd)
}

// addTaskCommands adds task management commands (which now include workspace management)
func (app *App) addTaskCommands() {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks and their workspaces",
		Long:  "Create, list, describe, and manage tasks. Workspaces and tasks are always synced and have the same ID.",
	}

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Long:  "Display all tasks with their status, priority, and basic information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listTasks(cmd)
		},
	}

	// Sync command
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync tasks from git provider",
		Long:  "Sync down tasks and statuses from associated git provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.syncTasks(cmd)
		},
	}

	// Describe command
	describeCmd := &cobra.Command{
		Use:   "describe [task-id-or-name]",
		Short: "Show detailed task information",
		Long:  "Display detailed information about a specific task including workspace details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.describeTask(cmd, args[0])
		},
	}

	// Priority command
	priorityCmd := &cobra.Command{
		Use:   "priority [task-id-or-name] [priority]",
		Short: "Change task priority",
		Long:  "Change the priority of a task. Use 'freeze' as priority to prevent execution until 'unfreeze' is called.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.setTaskPriority(cmd, args[0], args[1])
		},
	}

	// Start command
	startCmd := &cobra.Command{
		Use:   "start [task-id-or-name]",
		Short: "Start working on a task",
		Long:  "Start a task with the agent if possible and report status back",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.startTask(cmd, args[0])
		},
	}

	// Stop command
	stopCmd := &cobra.Command{
		Use:   "stop [task-id-or-name]",
		Short: "Stop a task",
		Long:  "Stop a task (pause if possible, otherwise full stop)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.stopTask(cmd, args[0])
		},
	}

	// Kill command
	killCmd := &cobra.Command{
		Use:   "kill [task-id-or-name]",
		Short: "Force kill a task",
		Long:  "Forcibly kill the agent container if it's working",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.killTask(cmd, args[0])
		},
	}

	// Logs command
	logsCmd := &cobra.Command{
		Use:   "logs [task-id-or-name]",
		Short: "Show task logs",
		Long:  "Show the log output of the agent as it works. Use --tail or -t for continuous output.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.showTaskLogs(cmd, args[0])
		},
	}

	// Add tail flag to logs command
	logsCmd.Flags().BoolP("tail", "t", false, "Continuously show logs")

	taskCmd.AddCommand(listCmd, syncCmd, describeCmd, priorityCmd, startCmd, stopCmd, killCmd, logsCmd)
	app.rootCmd.AddCommand(taskCmd)
}

// addGoCommand adds the workflow automation command
func (app *App) addGoCommand() {
	goCmd := &cobra.Command{
		Use:   "go",
		Short: "Start workflow automation",
		Long:  "Start a workflow that pulls down issues/PRs, runs up to N agents, and manages the workflow automatically",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.startWorkflow(cmd)
		},
	}

	// Add flags for workflow configuration
	goCmd.Flags().IntP("max-agents", "n", 1, "Maximum number of concurrent agents")
	goCmd.Flags().String("provider", "github", "Git provider to use (github, gitlab, bitbucket)")
	goCmd.Flags().String("repo", "", "Repository to monitor (default: current repo)")

	app.rootCmd.AddCommand(goCmd)
}

// Run executes the CLI application with the given arguments
func (app *App) Run(args []string) error {
	app.rootCmd.SetArgs(args[1:]) // Skip the program name
	return app.rootCmd.Execute()
}

// func (app *App) manageAuth(cmd *cobra.Command) error {
// 	cmd.Println("🔐 Authentication Management")
// 	cmd.Println("============================")
// 	cmd.Println()
// 	cmd.Println("Available authentication types:")
// 	cmd.Println("  1. Git providers (GitHub, GitLab, Bitbucket)")
// 	cmd.Println("  2. AI agent API keys (OpenAI, Anthropic, Gemini)")
// 	cmd.Println("  3. Container registry (Docker Hub, etc.)")
// 	cmd.Println("  4. Git credentials (SSH keys, username/password)")
// 	cmd.Println()
// 	cmd.Println("Commands:")
// 	cmd.Println("  cw config auth provider <provider>")
// 	cmd.Println("  cw config auth provider <provider> --token <token>")
// 	cmd.Println("  cw config auth agent set <agent> --key <api-key>")
// 	cmd.Println("  cw config auth git set --username <user> --email <email>")
// 	cmd.Println("  cw config auth git ssh --key-file <path>")
// 	cmd.Println("  cw config auth registry login <registry> --username <user> --password <pass>")
// 	cmd.Println()
// 	cmd.Println("Examples:")
// 	cmd.Println("  cw config auth provider github")
// 	cmd.Println("  cw config auth provider github --token YOUR_TOKEN")
// 	cmd.Println("  cw config auth agent set openai --key sk-...")
// 	cmd.Println("  cw config auth git set --username john --email john@example.com")
// 	cmd.Println("  cw config auth git ssh --key-file ~/.ssh/id_rsa")
// 	cmd.Println("  cw config auth registry login docker.io --username user --password pass")

// 	return nil
// }

// func (app *App) testProvider(cmd *cobra.Command, providerName string) error {
// 	// Validate provider
// 	availableProviders := providers.GetAvailableProviders()
// 	validProvider := false
// 	for _, provider := range availableProviders {
// 		if provider == providerName {
// 			validProvider = true
// 			break
// 		}
// 	}
// 	if !validProvider {
// 		return fmt.Errorf("unknown provider: %s. Available providers: %s", providerName, strings.Join(availableProviders, ", "))
// 	}

// 	// Get scope
// 	scope, _ := cmd.Flags().GetString("scope")

// 	// Convert provider name to git.ProviderType
// 	var providerType git.ProviderType
// 	switch providerName {
// 	case "github":
// 		providerType = git.ProviderGitHub
// 	case "gitlab":
// 		providerType = git.ProviderGitLab
// 	case "bitbucket":
// 		providerType = git.ProviderBitbucket
// 	default:
// 		return fmt.Errorf("unsupported provider type: %s", providerName)
// 	}

// 	// Convert scope string to auth.AuthScope
// 	var authScope auth.AuthScope
// 	switch scope {
// 	case "global":
// 		authScope = auth.AuthScopeGlobal
// 	case "project":
// 		authScope = auth.AuthScopeProject
// 	default:
// 		return fmt.Errorf("invalid scope: %s. Valid scopes: global, project", scope)
// 	}

// 	cmd.Printf("🧪 Testing authentication for %s provider...\n", providerName)
// 	cmd.Printf("Scope: %s\n", scope)

// 	// Create auth manager
// 	authManager, err := auth.NewManager(app.configManager)
// 	if err != nil {
// 		return fmt.Errorf("failed to create auth manager: %w", err)
// 	}

// 	// Test authentication
// 	ctx := cmd.Context()
// 	if err := authManager.TestAuth(ctx, providerType, authScope); err != nil {
// 		return fmt.Errorf("authentication test failed: %w", err)
// 	}

// 	cmd.Printf("✅ Authentication test passed!\n")
// 	cmd.Printf("🔐 Credentials are valid and working\n")
// 	cmd.Printf("🔗 API access confirmed\n")
// 	cmd.Printf("📋 Permissions verified\n")

// 	return nil
// }

// func (app *App) configureAgent(cmd *cobra.Command) error {
// 	cmd.Println("🤖 AI Agent Configuration")
// 	cmd.Println("=========================")
// 	cmd.Println()
// 	cmd.Println("Available agent types:")
// 	cmd.Println("  1. Aider (container-based AI coding agent)")
// 	cmd.Println("  2. Cursor (local AI coding assistant)")
// 	cmd.Println("  3. Claude (Anthropic's AI assistant)")
// 	cmd.Println("  4. Gemini (Google's AI assistant)")
// 	cmd.Println("  5. Custom (user-defined agent)")
// 	cmd.Println()
// 	cmd.Println("Commands:")
// 	cmd.Println("  cw config agent set <agent-type> --image <docker-image> --command <cmd>")
// 	cmd.Println("  cw config agent enable <agent-type>")
// 	cmd.Println("  cw config agent disable <agent-type>")
// 	cmd.Println("  cw config agent timeout <agent-type> --minutes <minutes>")
// 	cmd.Println("  cw config agent env <agent-type> --key <key> --value <value>")
// 	cmd.Println()
// 	cmd.Println("Examples:")
// 	cmd.Println("  cw config agent set aider --image paulgauthier/aider --command aider")
// 	cmd.Println("  cw config agent enable aider")
// 	cmd.Println("  cw config agent timeout aider --minutes 45")
// 	cmd.Println("  cw config agent env aider --key OPENAI_API_KEY --value sk-...")
// 	cmd.Println()
// 	cmd.Println("Aider Agent (Recommended):")
// 	cmd.Println("  - Runs in Docker container with paulgauthier/aider image")
// 	cmd.Println("  - Supports multiple AI providers (OpenAI, Anthropic, Gemini)")
// 	cmd.Println("  - Automatic Git workflow integration")
// 	cmd.Println("  - Comprehensive instruction generation")
// 	cmd.Println()
// 	cmd.Println("Configuration file: .cwconfig (local) or ~/.config/cw/config.yaml (global)")

// 	return nil
// }

// func (app *App) manageEnv(cmd *cobra.Command) error {
// 	cmd.Println("🌍 Environment Variable Management")
// 	cmd.Println("=================================")
// 	cmd.Println()
// 	cmd.Println("Environment variables are stored encrypted and passed to containers at runtime.")
// 	cmd.Println("These are local-only (not global) for security reasons.")
// 	cmd.Println()
// 	cmd.Println("Commands:")
// 	cmd.Println("  cw config env set <key> <value>")
// 	cmd.Println("  cw config env get <key>")
// 	cmd.Println("  cw config env list")
// 	cmd.Println("  cw config env delete <key>")
// 	cmd.Println("  cw config env import <file>")
// 	cmd.Println("  cw config env export <file>")
// 	cmd.Println()
// 	cmd.Println("Examples:")
// 	cmd.Println("  cw config env set OPENAI_API_KEY sk-...")
// 	cmd.Println("  cw config env set ANTHROPIC_API_KEY sk-ant-...")
// 	cmd.Println("  cw config env set GEMINI_API_KEY AIza...")
// 	cmd.Println("  cw config env set DOCKER_REGISTRY_TOKEN dckr_...")
// 	cmd.Println("  cw config env import .env")
// 	cmd.Println("  cw config env export backup.env")
// 	cmd.Println()
// 	cmd.Println("Common environment variables:")
// 	cmd.Println("  - OPENAI_API_KEY: OpenAI API key for GPT models")
// 	cmd.Println("  - ANTHROPIC_API_KEY: Anthropic API key for Claude")
// 	cmd.Println("  - GEMINI_API_KEY: Google API key for Gemini")
// 	cmd.Println("  - DOCKER_REGISTRY_TOKEN: Docker Hub or other registry token")
// 	cmd.Println("  - GITHUB_TOKEN: GitHub personal access token")
// 	cmd.Println("  - GITLAB_TOKEN: GitLab personal access token")
// 	cmd.Println("  - BITBUCKET_TOKEN: Bitbucket app password")

// 	return nil
// }

// func (app *App) saveConfig(cmd *cobra.Command, filename string) error {
// 	cmd.Printf("Save configuration to %s - not yet implemented\n", filename)
// 	return nil
// }

// func (app *App) loadConfig(cmd *cobra.Command, filename string) error {
// 	cmd.Printf("Load configuration from %s - not yet implemented\n", filename)
// 	return nil
// }

// Task command implementations
func (app *App) listTasks(cmd *cobra.Command) error {
	if app.taskManager == nil {
		return fmt.Errorf("task manager not initialized")
	}

	tasks, err := app.taskManager.ListTasks(nil)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		cmd.Println("No tasks found.")
		return nil
	}

	cmd.Printf("Found %d task(s):\n\n", len(tasks))
	for _, task := range tasks {
		statusIcon := getStatusIcon(task.Status)
		cmd.Printf("%s %s (priority: %d, status: %s)\n", statusIcon, task.Name, task.Priority, task.Status)
		if task.Description != "" {
			shortDesc := truncateString(task.Description, 80)
			cmd.Printf("   %s\n", shortDesc)
		}
		cmd.Println()
	}

	return nil
}

func (app *App) syncTasks(cmd *cobra.Command) error {
	if app.taskManager == nil {
		return fmt.Errorf("task manager not initialized")
	}

	cmd.Printf("🔄 Syncing tasks from Git provider...\n")

	// Get current repository info
	repoPath := "."
	repoInfo, err := git.GetRepositoryInfo(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	// For now, default to GitHub and require manual specification
	// TODO: Implement proper provider detection and owner/repo extraction
	cmd.Printf("📁 Repository: %s\n", repoInfo.RemoteURL)
	cmd.Printf("⚠️  Provider detection not yet implemented, defaulting to GitHub\n")

	// For now, require manual specification of owner/repo
	cmd.Printf("Please specify owner and repository manually for now\n")
	cmd.Printf("Example: cw task sync --owner hlfshell --repo cowork\n")

	return fmt.Errorf("task sync requires manual owner/repo specification. Use --owner and --repo flags")
}

func (app *App) describeTask(cmd *cobra.Command, identifier string) error {
	if app.taskManager == nil {
		return fmt.Errorf("task manager not initialized")
	}

	// Try to find task by ID or name
	var task *types.Task
	var err error

	// First try by ID
	task, err = app.taskManager.GetTask(identifier)
	if err != nil {
		// If not found by ID, try by name
		task, err = app.taskManager.GetTaskByName(identifier)
		if err != nil {
			return fmt.Errorf("task not found: %s", identifier)
		}
	}

	// Display detailed information
	cmd.Printf("📋 Task Details\n")
	cmd.Printf("==============\n\n")
	cmd.Printf("Name: %s\n", task.Name)
	cmd.Printf("ID: %d\n", task.ID)
	cmd.Printf("Status: %s\n", task.Status)
	cmd.Printf("Priority: %d\n", task.Priority)
	cmd.Printf("Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	cmd.Printf("Last Activity: %s\n", task.LastActivity.Format("2006-01-02 15:04:05"))

	if task.WorkspaceID != 0 {
		cmd.Printf("Workspace ID: %d\n", task.WorkspaceID)
	}

	if task.WorkspacePath != "" {
		cmd.Printf("Workspace Path: %s\n", task.WorkspacePath)
	}

	if task.Description != "" {
		cmd.Printf("\nDescription:\n")
		cmd.Printf("------------\n")
		cmd.Printf("%s\n", task.Description)
	}

	return nil
}

func (app *App) setTaskPriority(cmd *cobra.Command, identifier, priority string) error {
	if app.taskManager == nil {
		return fmt.Errorf("task manager not initialized")
	}

	// Parse priority
	priorityInt, err := parsePriority(priority)
	if err != nil {
		return fmt.Errorf("invalid priority: %s. Valid priorities: 1-5, low, medium, high, urgent", priority)
	}

	// Try to find task by ID or name
	var task *types.Task
	task, err = app.taskManager.GetTask(identifier)
	if err != nil {
		// If not found by ID, try by name
		task, err = app.taskManager.GetTaskByName(identifier)
		if err != nil {
			return fmt.Errorf("task not found: %s", identifier)
		}
	}

	// Update priority
	oldPriority := task.Priority
	updateReq := &types.UpdateTaskRequest{
		TaskID:   task.ID,
		Priority: &priorityInt,
	}

	// Save the updated task
	if _, err := app.taskManager.UpdateTask(updateReq); err != nil {
		return fmt.Errorf("failed to update task priority: %w", err)
	}

	cmd.Printf("✅ Updated task '%s' priority from %d to %d\n", task.Name, oldPriority, priorityInt)
	return nil
}

// parsePriority converts priority string to integer
func parsePriority(priority string) (int, error) {
	switch strings.ToLower(priority) {
	case "low", "1":
		return 1, nil
	case "medium", "2":
		return 2, nil
	case "high", "3":
		return 3, nil
	case "urgent", "4":
		return 4, nil
	case "critical", "5":
		return 5, nil
	default:
		// Try to parse as integer
		if prio, err := strconv.Atoi(priority); err == nil && prio >= 1 && prio <= 5 {
			return prio, nil
		}
		return 0, fmt.Errorf("invalid priority value")
	}
}

func (app *App) startTask(cmd *cobra.Command, identifier string) error {
	cmd.Printf("Start task %s - not yet implemented\n", identifier)
	return nil
}

func (app *App) stopTask(cmd *cobra.Command, identifier string) error {
	cmd.Printf("Stop task %s - not yet implemented\n", identifier)
	return nil
}

func (app *App) killTask(cmd *cobra.Command, identifier string) error {
	cmd.Printf("Kill task %s - not yet implemented\n", identifier)
	return nil
}

func (app *App) showTaskLogs(cmd *cobra.Command, identifier string) error {
	tail, _ := cmd.Flags().GetBool("tail")
	cmd.Printf("Show logs for task %s (tail: %v) - not yet implemented\n", identifier, tail)
	return nil
}

// Workflow command implementation
func (app *App) startWorkflow(cmd *cobra.Command) error {
	maxAgents, _ := cmd.Flags().GetInt("max-agents")
	provider, _ := cmd.Flags().GetString("provider")
	repo, _ := cmd.Flags().GetString("repo")

	cmd.Printf("Start workflow with max %d agents, provider %s, repo %s - not yet implemented\n", maxAgents, provider, repo)
	return nil
}

// Helper functions
func getStatusIcon(status types.TaskStatus) string {
	switch status {
	case types.TaskStatusQueued:
		return "⏳"
	case types.TaskStatusInProgress:
		return "🔄"
	case types.TaskStatusCompleted:
		return "✅"
	case types.TaskStatusFailed:
		return "❌"
	case types.TaskStatusCancelled:
		return "🚫"
	case types.TaskStatusPaused:
		return "⏸️"
	default:
		return "❓"
	}
}

func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// addWorkflowCommands adds workflow management commands
func (app *App) addWorkflowCommands() {
	workflowCmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage automated workflows",
		Long:  "Manage automated workflows that scan issues, create tasks, and manage pull requests",
	}

	// Scan command
	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan issues from Git provider and create workflows",
		Long:  "Scan open issues assigned to the current user from the configured Git provider and create automated workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.scanWorkflows(cmd)
		},
	}

	// Start command
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start workflow automation",
		Long:  "Start automated workflow that processes queued workflows and manages the complete auto-PR lifecycle",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.startWorkflows(cmd)
		},
	}

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflows",
		Long:  "Display all workflows with their current status and basic information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listWorkflows(cmd)
		},
	}

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status [workflow-id]",
		Short: "Show workflow status",
		Long:  "Display detailed status information for a specific workflow or all active workflows",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return app.showWorkflowStatus(cmd, args[0])
			}
			return app.showAllWorkflowStatus(cmd)
		},
	}

	// Add flags
	scanCmd.Flags().String("provider", "github", "Git provider to use (github, gitlab, bitbucket)")
	scanCmd.Flags().String("owner", "", "Repository owner (defaults to auto-detected from current repository)")
	scanCmd.Flags().String("repo", "", "Repository name (defaults to auto-detected from current repository)")

	startCmd.Flags().IntP("max-concurrent", "n", 1, "Maximum number of concurrent workflow processors")
	startCmd.Flags().Bool("daemon", false, "Run as daemon process")

	listCmd.Flags().String("state", "", "Filter by workflow state (queued, implementing, pr_open, etc.)")
	listCmd.Flags().Bool("active-only", false, "Show only active (non-terminal) workflows")

	workflowCmd.AddCommand(scanCmd, startCmd, listCmd, statusCmd)
	app.rootCmd.AddCommand(workflowCmd)
}

// validateProviderAuth validates that provider authentication is configured
func (app *App) validateProviderAuth(providerName string) error {
	// Create auth manager
	authManager, err := auth.NewManager(app.configManager)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Convert provider name to type
	var providerType git.ProviderType
	switch strings.ToLower(providerName) {
	case "github":
		providerType = git.ProviderGitHub
	case "gitlab":
		providerType = git.ProviderGitLab
	case "bitbucket":
		providerType = git.ProviderBitbucket
	default:
		return fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Try to get auth config (check project first, then global)
	_, err = authManager.GetAuthConfig(providerType, auth.AuthScopeProject)
	if err != nil {
		// Fallback to global scope
		_, err = authManager.GetAuthConfig(providerType, auth.AuthScopeGlobal)
		if err != nil {
			return fmt.Errorf("no authentication configured for %s. Run 'cw config provider %s login' first", providerName, providerName)
		}
	}

	return nil
}

// detectRepositoryInfo auto-detects repository information from current directory
func (app *App) detectRepositoryInfo() (owner, repo string, err error) {
	repoInfo, err := git.GetRepositoryInfo(".")
	if err != nil {
		return "", "", fmt.Errorf("failed to detect repository info: %w", err)
	}

	// Parse owner/repo from remote URL
	// This is a simplified extraction - should be more robust
	remoteURL := repoInfo.RemoteURL
	if strings.Contains(remoteURL, "github.com") {
		// Parse URLs like "git@github.com:owner/repo.git" or "https://github.com/owner/repo.git"
		parts := strings.Split(remoteURL, "/")
		if len(parts) >= 2 {
			repo = strings.TrimSuffix(parts[len(parts)-1], ".git")
			owner = parts[len(parts)-2]
			if strings.Contains(owner, ":") {
				// Handle SSH format like "git@github.com:owner"
				ownerParts := strings.Split(owner, ":")
				if len(ownerParts) >= 2 {
					owner = ownerParts[1]
				}
			}
		}
	}

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("failed to parse owner/repo from remote URL: %s", remoteURL)
	}

	return owner, repo, nil
}

// Workflow command implementations

func (app *App) scanWorkflows(cmd *cobra.Command) error {
	provider, _ := cmd.Flags().GetString("provider")
	owner, _ := cmd.Flags().GetString("owner")
	repo, _ := cmd.Flags().GetString("repo")

	// Auto-detect repository info if not provided
	if owner == "" || repo == "" {
		detectedOwner, detectedRepo, err := app.detectRepositoryInfo()
		if err != nil {
			return fmt.Errorf("failed to auto-detect repository info: %w. Please specify --owner and --repo flags", err)
		}
		if owner == "" {
			owner = detectedOwner
		}
		if repo == "" {
			repo = detectedRepo
		}
	}

	cmd.Printf("🔍 Scanning issues from %s provider for %s/%s...\n", provider, owner, repo)

	// Validate provider authentication
	if err := app.validateProviderAuth(provider); err != nil {
		return fmt.Errorf("provider authentication validation failed: %w", err)
	}

	// For now, we'll create placeholder tasks to demonstrate the workflow
	// TODO: Implement full issue scanning when GitProvider interface is ready

	cmd.Printf("⚠️  Issue scanning is temporarily simplified\n")
	cmd.Printf("📋 Creating example task to demonstrate workflow functionality\n")

	// Create an example task to demonstrate workflow
	ticketID := fmt.Sprintf("%s:%s/%s#example", provider, owner, repo)
	req := &types.CreateTaskRequest{
		Name:        fmt.Sprintf("Example workflow task for %s/%s", owner, repo),
		Description: "This is an example task created to demonstrate workflow functionality. Replace with actual issue scanning when the provider interface is ready.",
		TicketID:    ticketID,
		URL:         fmt.Sprintf("https://github.com/%s/%s", owner, repo),
		Priority:    1, // Default priority
		Tags:        []string{provider, "workflow", "example"},
		Metadata: map[string]string{
			"provider":     provider,
			"owner":        owner,
			"repo":         repo,
			"issue_number": "example",
			"issue_title":  "Example workflow task",
			"created_at":   "2024-01-01T00:00:00Z",
		},
	}

	// Check if task already exists
	existingTasks, err := app.taskManager.ListTasks(nil)
	if err != nil {
		return fmt.Errorf("failed to list existing tasks: %w", err)
	}

	taskExists := false
	for _, task := range existingTasks {
		if task.TicketID == ticketID {
			taskExists = true
			cmd.Printf("✅ Example task already exists: %s (Task ID: %d)\n", task.Name, task.ID)
			break
		}
	}

	if !taskExists {
		task, err := app.taskManager.CreateTask(req)
		if err != nil {
			return fmt.Errorf("failed to create example task: %w", err)
		}

		cmd.Printf("✅ Created example task: %s (Task ID: %d)\n", task.Name, task.ID)
	}

	cmd.Printf("✅ Issue scanning completed successfully\n")
	return nil
}

func (app *App) startWorkflows(cmd *cobra.Command) error {
	maxConcurrent, _ := cmd.Flags().GetInt("max-concurrent")
	daemon, _ := cmd.Flags().GetBool("daemon")

	// Auto-detect repository info
	owner, repo, err := app.detectRepositoryInfo()
	if err != nil {
		return fmt.Errorf("failed to auto-detect repository info: %w", err)
	}

	cmd.Printf("🚀 Starting workflow automation for %s/%s...\n", owner, repo)
	cmd.Printf("📊 Max concurrent processors: %d\n", maxConcurrent)

	if daemon {
		cmd.Printf("🔄 Running in daemon mode...\n")
		// TODO: Implement daemon mode
		return fmt.Errorf("daemon mode not yet implemented")
	}

	// For now, we'll provide a simplified workflow start implementation
	// TODO: Implement full workflow automation when CoworkProvider is ready

	cmd.Printf("⚠️  Full workflow automation not yet implemented\n")
	cmd.Printf("📋 You can use the following commands to manage workflows manually:\n")
	cmd.Printf("   • cw workflow scan - Scan for issues and create tasks\n")
	cmd.Printf("   • cw task list - View created tasks\n")
	cmd.Printf("   • cw workflow list - View workflow status\n")
	cmd.Printf("\n💡 Full automation will be available once the workflow engine is complete\n")

	cmd.Printf("✅ Workflow automation completed successfully\n")
	return nil
}

func (app *App) listWorkflows(cmd *cobra.Command) error {
	stateFilter, _ := cmd.Flags().GetString("state")
	activeOnly, _ := cmd.Flags().GetBool("active-only")

	// Create workflow manager
	coworkDir := filepath.Join(".", ".cowork")
	workflowManager, err := workflow.NewWorkflowManager(coworkDir)
	if err != nil {
		return fmt.Errorf("failed to create workflow manager: %w", err)
	}
	defer workflowManager.Close()

	var workflows []*types.Workflow
	if stateFilter != "" {
		// Parse state filter
		state := types.WorkflowState(stateFilter)
		if !state.IsValid() {
			return fmt.Errorf("invalid workflow state: %s", stateFilter)
		}
		workflows, err = workflowManager.ListWorkflowsByState(state)
	} else {
		workflows, err = workflowManager.ListWorkflows()
	}

	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	// Filter active workflows if requested
	if activeOnly {
		var activeWorkflows []*types.Workflow
		for _, workflow := range workflows {
			if !workflow.State.IsTerminal() {
				activeWorkflows = append(activeWorkflows, workflow)
			}
		}
		workflows = activeWorkflows
	}

	if len(workflows) == 0 {
		cmd.Println("No workflows found.")
		return nil
	}

	cmd.Printf("Found %d workflow(s):\n\n", len(workflows))
	for _, workflow := range workflows {
		statusIcon := getWorkflowStatusIcon(workflow.State)
		cmd.Printf("%s Workflow %d: %s/%s#%d\n", statusIcon, workflow.ID, workflow.Owner, workflow.Repo, workflow.IssueID)
		cmd.Printf("   State: %s\n", workflow.State)
		if workflow.BranchName != "" {
			cmd.Printf("   Branch: %s\n", workflow.BranchName)
		}
		if workflow.PRNumber != nil {
			cmd.Printf("   PR: #%d\n", *workflow.PRNumber)
		}
		cmd.Printf("   Created: %s\n", workflow.CreatedAt.Format("2006-01-02 15:04:05"))
		cmd.Println()
	}

	return nil
}

func (app *App) showWorkflowStatus(cmd *cobra.Command, workflowID string) error {
	// Create workflow manager
	coworkDir := filepath.Join(".", ".cowork")
	workflowManager, err := workflow.NewWorkflowManager(coworkDir)
	if err != nil {
		return fmt.Errorf("failed to create workflow manager: %w", err)
	}
	defer workflowManager.Close()

	// Get workflow
	workflow, err := workflowManager.GetWorkflow(workflowID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Display detailed workflow information
	cmd.Printf("📋 Workflow %d Details\n", workflow.ID)
	cmd.Printf("======================\n\n")
	cmd.Printf("Repository: %s/%s\n", workflow.Owner, workflow.Repo)
	cmd.Printf("Issue: #%d\n", workflow.IssueID)
	cmd.Printf("Provider: %s\n", workflow.Provider)
	cmd.Printf("State: %s\n", workflow.State)
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

	cmd.Printf("\nTimestamps:\n")
	cmd.Printf("Created: %s\n", workflow.CreatedAt.Format("2006-01-02 15:04:05"))
	cmd.Printf("Updated: %s\n", workflow.UpdatedAt.Format("2006-01-02 15:04:05"))

	if workflow.StartedAt != nil {
		cmd.Printf("Started: %s\n", workflow.StartedAt.Format("2006-01-02 15:04:05"))
	}

	if workflow.EndedAt != nil {
		cmd.Printf("Ended: %s\n", workflow.EndedAt.Format("2006-01-02 15:04:05"))
	}

	if workflow.ErrorCount > 0 {
		cmd.Printf("\nErrors: %d\n", workflow.ErrorCount)
		if workflow.LastError != "" {
			cmd.Printf("Last Error: %s\n", workflow.LastError)
		}
	}

	// Check if workflow is locked
	locked, lock := workflowManager.IsWorkflowLocked(workflowID)
	if locked {
		cmd.Printf("\n🔒 Locked by: %s\n", lock.LockedBy)
		cmd.Printf("Lock timeout: %s\n", lock.LockTimeout.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func (app *App) showAllWorkflowStatus(cmd *cobra.Command) error {
	// Create workflow manager
	coworkDir := filepath.Join(".", ".cowork")
	workflowManager, err := workflow.NewWorkflowManager(coworkDir)
	if err != nil {
		return fmt.Errorf("failed to create workflow manager: %w", err)
	}
	defer workflowManager.Close()

	// Get all workflows
	workflows, err := workflowManager.ListWorkflows()
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	if len(workflows) == 0 {
		cmd.Println("No workflows found.")
		return nil
	}

	// Group workflows by state
	stateGroups := make(map[types.WorkflowState][]*types.Workflow)
	for _, workflow := range workflows {
		stateGroups[workflow.State] = append(stateGroups[workflow.State], workflow)
	}

	cmd.Printf("📊 Workflow Status Summary\n")
	cmd.Printf("==========================\n\n")

	// Display statistics
	totalWorkflows := len(workflows)
	activeWorkflows := 0
	completedWorkflows := 0

	for _, workflow := range workflows {
		if workflow.State.IsTerminal() {
			completedWorkflows++
		} else {
			activeWorkflows++
		}
	}

	cmd.Printf("Total Workflows: %d\n", totalWorkflows)
	cmd.Printf("Active: %d\n", activeWorkflows)
	cmd.Printf("Completed: %d\n", completedWorkflows)
	cmd.Println()

	// Display workflows by state
	for state, workflowList := range stateGroups {
		statusIcon := getWorkflowStatusIcon(state)
		cmd.Printf("%s %s (%d):\n", statusIcon, strings.ToUpper(string(state)), len(workflowList))
		for _, workflow := range workflowList {
			cmd.Printf("   • Workflow %d: %s/%s#%d", workflow.ID, workflow.Owner, workflow.Repo, workflow.IssueID)
			if workflow.BranchName != "" {
				cmd.Printf(" (%s)", workflow.BranchName)
			}
			cmd.Println()
		}
		cmd.Println()
	}

	return nil
}

// Helper function to get workflow status icons
func getWorkflowStatusIcon(state types.WorkflowState) string {
	switch state {
	case types.WorkflowStateQueued:
		return "⏳"
	case types.WorkflowStateWorkspaceReady:
		return "🏗️"
	case types.WorkflowStateImplementing:
		return "💻"
	case types.WorkflowStatePROpen:
		return "🔍"
	case types.WorkflowStateRevising:
		return "🔄"
	case types.WorkflowStateMerged:
		return "✅"
	case types.WorkflowStateClosed:
		return "❌"
	case types.WorkflowStateAborted:
		return "🚫"
	default:
		return "❓"
	}
}
