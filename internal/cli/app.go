package cli

import (
	"fmt"
	"path/filepath"

	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/task"
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
	cwDir := filepath.Join(".", ".cowork")
	taskManager, err := task.NewManager(cwDir, 300)
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

	// // Add task commands (which now handle workspaces)
	// app.addTaskCommands()

	// // Add go command for workflow automation
	// app.addGoCommand()
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

// // addTaskCommands adds task management commands (which now include workspace management)
// func (app *App) addTaskCommands() {
// 	taskCmd := &cobra.Command{
// 		Use:   "task",
// 		Short: "Manage tasks and their workspaces",
// 		Long:  "Create, list, describe, and manage tasks. Workspaces and tasks are always synced and have the same ID.",
// 	}

// 	// List command
// 	listCmd := &cobra.Command{
// 		Use:   "list",
// 		Short: "List all tasks",
// 		Long:  "Display all tasks with their status, priority, and basic information",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.listTasks(cmd)
// 		},
// 	}

// 	// Sync command
// 	syncCmd := &cobra.Command{
// 		Use:   "sync",
// 		Short: "Sync tasks from git provider",
// 		Long:  "Sync down tasks and statuses from associated git provider",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.syncTasks(cmd)
// 		},
// 	}

// 	// Describe command
// 	describeCmd := &cobra.Command{
// 		Use:   "describe [task-id-or-name]",
// 		Short: "Show detailed task information",
// 		Long:  "Display detailed information about a specific task including workspace details",
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.describeTask(cmd, args[0])
// 		},
// 	}

// 	// Priority command
// 	priorityCmd := &cobra.Command{
// 		Use:   "priority [task-id-or-name] [priority]",
// 		Short: "Change task priority",
// 		Long:  "Change the priority of a task. Use 'freeze' as priority to prevent execution until 'unfreeze' is called.",
// 		Args:  cobra.ExactArgs(2),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.setTaskPriority(cmd, args[0], args[1])
// 		},
// 	}

// 	// Start command
// 	startCmd := &cobra.Command{
// 		Use:   "start [task-id-or-name]",
// 		Short: "Start working on a task",
// 		Long:  "Start a task with the agent if possible and report status back",
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.startTask(cmd, args[0])
// 		},
// 	}

// 	// Stop command
// 	stopCmd := &cobra.Command{
// 		Use:   "stop [task-id-or-name]",
// 		Short: "Stop a task",
// 		Long:  "Stop a task (pause if possible, otherwise full stop)",
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.stopTask(cmd, args[0])
// 		},
// 	}

// 	// Kill command
// 	killCmd := &cobra.Command{
// 		Use:   "kill [task-id-or-name]",
// 		Short: "Force kill a task",
// 		Long:  "Forcibly kill the agent container if it's working",
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.killTask(cmd, args[0])
// 		},
// 	}

// 	// Logs command
// 	logsCmd := &cobra.Command{
// 		Use:   "logs [task-id-or-name]",
// 		Short: "Show task logs",
// 		Long:  "Show the log output of the agent as it works. Use --tail or -t for continuous output.",
// 		Args:  cobra.ExactArgs(1),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.showTaskLogs(cmd, args[0])
// 		},
// 	}

// 	// Add tail flag to logs command
// 	logsCmd.Flags().BoolP("tail", "t", false, "Continuously show logs")

// 	taskCmd.AddCommand(listCmd, syncCmd, describeCmd, priorityCmd, startCmd, stopCmd, killCmd, logsCmd)
// 	app.rootCmd.AddCommand(taskCmd)
// }

// // addGoCommand adds the workflow automation command
// func (app *App) addGoCommand() {
// 	goCmd := &cobra.Command{
// 		Use:   "go",
// 		Short: "Start workflow automation",
// 		Long:  "Start a workflow that pulls down issues/PRs, runs up to N agents, and manages the workflow automatically",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return app.startWorkflow(cmd)
// 		},
// 	}

// 	// Add flags for workflow configuration
// 	goCmd.Flags().IntP("max-agents", "n", 1, "Maximum number of concurrent agents")
// 	goCmd.Flags().String("provider", "github", "Git provider to use (github, gitlab, bitbucket)")
// 	goCmd.Flags().String("repo", "", "Repository to monitor (default: current repo)")

// 	app.rootCmd.AddCommand(goCmd)
// }

// // Run executes the CLI application with the given arguments
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

// // Task command implementations
// func (app *App) listTasks(cmd *cobra.Command) error {
// 	if app.taskManager == nil {
// 		return fmt.Errorf("task manager not initialized")
// 	}

// 	tasks, err := app.taskManager.ListTasks(nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to list tasks: %w", err)
// 	}

// 	if len(tasks) == 0 {
// 		cmd.Println("No tasks found.")
// 		return nil
// 	}

// 	cmd.Printf("Found %d task(s):\n\n", len(tasks))
// 	for _, task := range tasks {
// 		statusIcon := getStatusIcon(task.Status)
// 		cmd.Printf("%s %s (priority: %d, status: %s)\n", statusIcon, task.Name, task.Priority, task.Status)
// 		if task.Description != "" {
// 			shortDesc := truncateString(task.Description, 80)
// 			cmd.Printf("   %s\n", shortDesc)
// 		}
// 		cmd.Println()
// 	}

// 	return nil
// }

// func (app *App) syncTasks(cmd *cobra.Command) error {
// 	if app.taskManager == nil {
// 		return fmt.Errorf("task manager not initialized")
// 	}

// 	cmd.Printf("🔄 Syncing tasks from Git provider...\n")

// 	// Get current repository info
// 	repoPath := "."
// 	repoInfo, err := git.GetRepositoryInfo(repoPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to get repository info: %w", err)
// 	}

// 	// For now, default to GitHub and require manual specification
// 	// TODO: Implement proper provider detection and owner/repo extraction
// 	cmd.Printf("📁 Repository: %s\n", repoInfo.RemoteURL)
// 	cmd.Printf("⚠️  Provider detection not yet implemented, defaulting to GitHub\n")

// 	// For now, require manual specification of owner/repo
// 	cmd.Printf("Please specify owner and repository manually for now\n")
// 	cmd.Printf("Example: cw task sync --owner hlfshell --repo cowork\n")

// 	return fmt.Errorf("task sync requires manual owner/repo specification. Use --owner and --repo flags")
// 	return nil
// }

// func (app *App) describeTask(cmd *cobra.Command, identifier string) error {
// 	if app.taskManager == nil {
// 		return fmt.Errorf("task manager not initialized")
// 	}

// 	// Try to find task by ID or name
// 	var task *types.Task
// 	var err error

// 	// First try by ID
// 	task, err = app.taskManager.GetTask(identifier)
// 	if err != nil {
// 		// If not found by ID, try by name
// 		task, err = app.taskManager.GetTaskByName(identifier)
// 		if err != nil {
// 			return fmt.Errorf("task not found: %s", identifier)
// 		}
// 	}

// 	// Display detailed information
// 	cmd.Printf("📋 Task Details\n")
// 	cmd.Printf("==============\n\n")
// 	cmd.Printf("Name: %s\n", task.Name)
// 	cmd.Printf("ID: %d\n", task.ID)
// 	cmd.Printf("Status: %s\n", task.Status)
// 	cmd.Printf("Priority: %d\n", task.Priority)
// 	cmd.Printf("Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
// 	cmd.Printf("Last Activity: %s\n", task.LastActivity.Format("2006-01-02 15:04:05"))

// 	if task.WorkspaceID != 0 {
// 		cmd.Printf("Workspace ID: %d\n", task.WorkspaceID)
// 	}

// 	if task.WorkspacePath != "" {
// 		cmd.Printf("Workspace Path: %s\n", task.WorkspacePath)
// 	}

// 	if task.Description != "" {
// 		cmd.Printf("\nDescription:\n")
// 		cmd.Printf("------------\n")
// 		cmd.Printf("%s\n", task.Description)
// 	}

// 	return nil
// }

// func (app *App) setTaskPriority(cmd *cobra.Command, identifier, priority string) error {
// 	if app.taskManager == nil {
// 		return fmt.Errorf("task manager not initialized")
// 	}

// 	// Parse priority
// 	priorityInt, err := parsePriority(priority)
// 	if err != nil {
// 		return fmt.Errorf("invalid priority: %s. Valid priorities: 1-5, low, medium, high, urgent", priority)
// 	}

// 	// Try to find task by ID or name
// 	var task *types.Task
// 	task, err = app.taskManager.GetTask(identifier)
// 	if err != nil {
// 		// If not found by ID, try by name
// 		task, err = app.taskManager.GetTaskByName(identifier)
// 		if err != nil {
// 			return fmt.Errorf("task not found: %s", identifier)
// 		}
// 	}

// 	// Update priority
// 	oldPriority := task.Priority
// 	task.Priority = priorityInt
// 	task.LastActivity = time.Now()

// 	// Save the updated task
// 	if err := app.taskManager.UpdateTask(task); err != nil {
// 		return fmt.Errorf("failed to update task priority: %w", err)
// 	}

// 	cmd.Printf("✅ Updated task '%s' priority from %d to %d\n", task.Name, oldPriority, priorityInt)
// 	return nil
// }

// // parsePriority converts priority string to integer
// func parsePriority(priority string) (int, error) {
// 	switch strings.ToLower(priority) {
// 	case "low", "1":
// 		return 1, nil
// 	case "medium", "2":
// 		return 2, nil
// 	case "high", "3":
// 		return 3, nil
// 	case "urgent", "4":
// 		return 4, nil
// 	case "critical", "5":
// 		return 5, nil
// 	default:
// 		// Try to parse as integer
// 		if prio, err := strconv.Atoi(priority); err == nil && prio >= 1 && prio <= 5 {
// 			return prio, nil
// 		}
// 		return 0, fmt.Errorf("invalid priority value")
// 	}
// }

// func (app *App) startTask(cmd *cobra.Command, identifier string) error {
// 	cmd.Printf("Start task %s - not yet implemented\n", identifier)
// 	return nil
// }

// func (app *App) stopTask(cmd *cobra.Command, identifier string) error {
// 	cmd.Printf("Stop task %s - not yet implemented\n", identifier)
// 	return nil
// }

// func (app *App) killTask(cmd *cobra.Command, identifier string) error {
// 	cmd.Printf("Kill task %s - not yet implemented\n", identifier)
// 	return nil
// }

// func (app *App) showTaskLogs(cmd *cobra.Command, identifier string) error {
// 	tail, _ := cmd.Flags().GetBool("tail")
// 	cmd.Printf("Show logs for task %s (tail: %v) - not yet implemented\n", identifier, tail)
// 	return nil
// }

// // Workflow command implementation
// func (app *App) startWorkflow(cmd *cobra.Command) error {
// 	maxAgents, _ := cmd.Flags().GetInt("max-agents")
// 	provider, _ := cmd.Flags().GetString("provider")
// 	repo, _ := cmd.Flags().GetString("repo")

// 	cmd.Printf("Start workflow with max %d agents, provider %s, repo %s - not yet implemented\n", maxAgents, provider, repo)
// 	return nil
// }

// // Helper functions
// func getStatusIcon(status types.TaskStatus) string {
// 	switch status {
// 	case types.TaskStatusQueued:
// 		return "⏳"
// 	case types.TaskStatusInProgress:
// 		return "🔄"
// 	case types.TaskStatusCompleted:
// 		return "✅"
// 	case types.TaskStatusFailed:
// 		return "❌"
// 	case types.TaskStatusCancelled:
// 		return "🚫"
// 	case types.TaskStatusPaused:
// 		return "⏸️"
// 	default:
// 		return "❓"
// 	}
// }

// func truncateString(s string, maxLength int) string {
// 	if len(s) <= maxLength {
// 		return s
// 	}
// 	return s[:maxLength-3] + "..."
// }
