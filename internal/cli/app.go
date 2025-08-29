package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/task"
	"github.com/hlfshell/cowork/internal/types"
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
	configManager := config.NewManager()
	if err := ensureGlobalConfigExists(configManager); err != nil {
		fmt.Printf("Warning: failed to initialize global configuration: %v\n", err)
	}

	// Initialize task manager (which now handles workspaces)
	cwDir := filepath.Join(".", ".cw")
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

// ensureGlobalConfigExists checks if the global configuration file exists and creates it if it doesn't
func ensureGlobalConfigExists(configManager *config.Manager) error {
	// Check if global config file exists
	if _, err := os.Stat(configManager.GlobalConfigPath); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		configDir := filepath.Dir(configManager.GlobalConfigPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Create default configuration
		defaultConfig := configManager.GetDefaultConfig()
		if err := configManager.SaveGlobal(defaultConfig); err != nil {
			return fmt.Errorf("failed to create default global configuration: %w", err)
		}

		fmt.Printf("✅ Created default global configuration at: %s\n", configManager.GlobalConfigPath)
	}
	return nil
}

// setupCommands initializes all CLI commands and their structure
func (app *App) setupCommands() {
	app.rootCmd = &cobra.Command{
		Use:   "cw",
		Short: "Multi-Agent Repo Orchestrator",
		Long: `cowork (cw) — Multi-Agent Repo Orchestrator

A Go-based CLI that lets one developer spin up many isolated, containerized workspaces 
on the same Git repository, wire them to AI coding agents, and drive them from tickets 
(GitHub/GitLab/Jira/Linear). It keeps your main checkout pristine while parallel 
"coworkers" code safely on branches you can review and merge.

Core Features:
• Isolated workspaces per task (worktree/linked-clone/full-clone)
• Containerized dev environments per workspace
• Agent runners for Cursor/Claude/Gemini/etc.
• Ticket-first workflows
• Rules engine with .cwrules configuration
• State tracking with .cwstate

For more information, visit: https://github.com/hlfshell/cowork`,
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

// addConfigCommands adds configuration management commands
func (app *App) addConfigCommands() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
		Long:  "Configure global and local cw setups. All config attributes follow the order: flags > local .cw configs > global .cw configs",
	}

	// Show command
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration settings",
		Long:  "Display the current configuration settings in a human readable manner. Never show keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.showConfig(cmd)
		},
	}

	// Auth command
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication settings",
		Long:  "Handle authentication for git providers (GitHub, GitLab, Bitbucket) and API keys for agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.manageAuth(cmd)
		},
	}

	// Agent command
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Configure AI agent settings",
		Long:  "Configure settings for AI agents (docker image, command structure, keys, etc.)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.configureAgent(cmd)
		},
	}

	// Env command
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environment variables",
		Long:  "Save and store .env variables that are passed to containers at runtime. These are stored encrypted and can only be local.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.manageEnv(cmd)
		},
	}

	// Save command
	saveCmd := &cobra.Command{
		Use:   "save [filename]",
		Short: "Save configuration to YAML file",
		Long:  "Save current local configuration settings to a YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.saveConfig(cmd, args[0])
		},
	}

	// Load command
	loadCmd := &cobra.Command{
		Use:   "load [filename]",
		Short: "Load configuration from YAML file",
		Long:  "Load configuration settings from a YAML file into local settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.loadConfig(cmd, args[0])
		},
	}

	configCmd.AddCommand(showCmd, authCmd, agentCmd, envCmd, saveCmd, loadCmd)
	app.rootCmd.AddCommand(configCmd)
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

// initializeProject initializes cowork for the current project
func (app *App) initializeProject(cmd *cobra.Command) error {
	// Check if we're in a Git repository
	repoInfo, err := git.DetectCurrentRepository()
	if err != nil {
		return fmt.Errorf("failed to detect Git repository: %w", err)
	}

	cmd.Printf("✅ Detected Git repository: %s\n", repoInfo.Path)
	cmd.Printf("   Current branch: %s\n", repoInfo.CurrentBranch)

	// Check if already initialized
	cwDir := filepath.Join(repoInfo.Path, ".cw")
	if _, err := os.Stat(cwDir); err == nil {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return fmt.Errorf("cowork is already initialized in this repository. Use --force to reinitialize")
		}
		cmd.Printf("ℹ️  Reinitializing existing cowork setup\n")
	}

	// Create .cw directory
	if err := os.MkdirAll(cwDir, 0755); err != nil {
		return fmt.Errorf("failed to create .cw directory: %w", err)
	}
	cmd.Printf("✅ Created .cw directory: %s\n", cwDir)

	// Create workspaces directory
	workspacesDir := filepath.Join(cwDir, "workspaces")
	if err := os.MkdirAll(workspacesDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspaces directory: %w", err)
	}
	cmd.Printf("✅ Created workspaces directory: %s\n", workspacesDir)

	// Check if project config exists, create if not
	projectConfigPath := filepath.Join(repoInfo.Path, ".cwconfig")
	if _, err := os.Stat(projectConfigPath); os.IsNotExist(err) {
		// Create default project configuration
		defaultConfig := app.configManager.GetDefaultConfig()
		if err := app.configManager.SaveProject(defaultConfig); err != nil {
			return fmt.Errorf("failed to create project configuration: %w", err)
		}
		cmd.Printf("✅ Created project configuration: %s\n", projectConfigPath)
	} else {
		cmd.Printf("ℹ️  Project configuration already exists: %s\n", projectConfigPath)
	}

	cmd.Printf("\n🎉 Project initialized successfully!\n")
	cmd.Printf("   You can now use: cw task start <task-name>\n")
	cmd.Printf("   Configuration: cw config show\n")

	return nil
}

// Configuration command implementations
func (app *App) showConfig(cmd *cobra.Command) error {
	cmd.Println("Configuration management - not yet implemented")
	return nil
}

func (app *App) manageAuth(cmd *cobra.Command) error {
	cmd.Println("Authentication management - not yet implemented")
	return nil
}

func (app *App) configureAgent(cmd *cobra.Command) error {
	cmd.Println("Agent configuration - not yet implemented")
	return nil
}

func (app *App) manageEnv(cmd *cobra.Command) error {
	cmd.Println("Environment variable management - not yet implemented")
	return nil
}

func (app *App) saveConfig(cmd *cobra.Command, filename string) error {
	cmd.Printf("Save configuration to %s - not yet implemented\n", filename)
	return nil
}

func (app *App) loadConfig(cmd *cobra.Command, filename string) error {
	cmd.Printf("Load configuration from %s - not yet implemented\n", filename)
	return nil
}

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
	cmd.Println("Task synchronization - not yet implemented")
	return nil
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
	cmd.Printf("Set priority for task %s to %s - not yet implemented\n", identifier, priority)
	return nil
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
