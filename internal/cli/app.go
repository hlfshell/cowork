package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/task"
	"github.com/hlfshell/cowork/internal/types"
	"github.com/hlfshell/cowork/internal/workspace"
	"github.com/spf13/cobra"
)

// App represents the main CLI application
type App struct {
	rootCmd          *cobra.Command
	version          string
	buildDate        string
	gitCommit        string
	workspaceManager workspace.WorkspaceManager
	taskManager      task.TaskManager
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
		version:          version,
		buildDate:        buildDate,
		gitCommit:        gitCommit,
		workspaceManager: nil, // No longer needed - tasks handle workspaces
		taskManager:      taskManager,
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

	// Add task commands (which now handle workspaces)
	app.addTaskCommands()

	// Add ticket commands (placeholder for future implementation)
	app.addTicketCommands()

	// Add agent commands (placeholder for future implementation)
	app.addAgentCommands()

	// Add config commands
	app.addConfigCommands()

	// Add auth commands
	app.addAuthCommands()
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

	app.rootCmd.AddCommand(initCmd)
}

// addWorkspaceCommands adds workspace management commands
func (app *App) addWorkspaceCommands() {
	workspaceCmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage isolated workspaces",
		Long:  "Create, list, describe, and manage isolated workspaces for different tasks",
	}

	// Create workspace command
	createCmd := &cobra.Command{
		Use:   "create [task-name]",
		Short: "Create a new isolated workspace",
		Long:  "Create a new isolated workspace for the specified task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.workspaceManager == nil {
				return fmt.Errorf("workspace manager not initialized")
			}

			taskName := args[0]

			// Get description from flags
			message, _ := cmd.Flags().GetString("message")
			messageFile, _ := cmd.Flags().GetString("message-file")

			var description string
			if messageFile != "" {
				// Read description from file
				content, err := workspace.ReadDescriptionFromFile(messageFile)
				if err != nil {
					return fmt.Errorf("failed to read description from file: %w", err)
				}
				description = content
			} else if message != "" {
				description = message
			}

			// Get current repository information
			repoInfo, err := git.DetectCurrentRepository()
			if err != nil {
				return fmt.Errorf("failed to detect current repository: %w", err)
			}

			// Create workspace request using current repository
			req := &types.CreateWorkspaceRequest{
				TaskName:    taskName,
				Description: description,
				SourceRepo:  repoInfo.Path,
				BaseBranch:  repoInfo.CurrentBranch,
			}

			cmd.Printf("Creating workspace for task: %s\n", taskName)
			workspace, err := app.workspaceManager.CreateWorkspace(req)
			if err != nil {
				return fmt.Errorf("failed to create workspace: %w", err)
			}

			cmd.Printf("✅ Workspace created successfully!\n")
			cmd.Printf("   ID: %s\n", workspace.ID)
			cmd.Printf("   Path: %s\n", workspace.Path)
			cmd.Printf("   Branch: %s\n", workspace.BranchName)
			cmd.Printf("   Status: %s\n", workspace.Status)
			if description != "" {
				if len(description) > 100 {
					cmd.Printf("   Description: %s...\n", description[:100])
				} else {
					cmd.Printf("   Description: %s\n", description)
				}
			}

			return nil
		},
	}

	// Add message flags to create command
	createCmd.Flags().StringP("message", "m", "", "Description of what the task is trying to accomplish")
	createCmd.Flags().String("message-file", "", "File containing the description of what the task is trying to accomplish")

	// Describe workspace command
	describeCmd := &cobra.Command{
		Use:   "describe [workspace-id-or-name]",
		Short: "Show detailed information about a workspace",
		Long:  "Display detailed information about a specific workspace including full description, path, and metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.workspaceManager == nil {
				return fmt.Errorf("workspace manager not initialized")
			}

			identifier := args[0]

			// Try to find workspace by ID or task name
			var workspace *types.Workspace
			var err error

			// First try by ID
			workspace, err = app.workspaceManager.GetWorkspace(identifier)
			if err != nil {
				// If not found by ID, try by task name
				workspace, err = app.workspaceManager.GetWorkspaceByTaskName(identifier)
				if err != nil {
					return fmt.Errorf("workspace not found: %s", identifier)
				}
			}

			// Display detailed information
			cmd.Printf("📁 Workspace Details\n")
			cmd.Printf("==================\n\n")
			cmd.Printf("Task Name: %s\n", workspace.TaskName)
			cmd.Printf("ID: %s\n", workspace.ID)
			cmd.Printf("Status: %s\n", workspace.Status)
			cmd.Printf("Path: %s\n", workspace.Path)
			cmd.Printf("Branch: %s\n", workspace.BranchName)
			cmd.Printf("Source Repo: %s\n", workspace.SourceRepo)
			cmd.Printf("Base Branch: %s\n", workspace.BaseBranch)
			cmd.Printf("Created: %s\n", workspace.CreatedAt.Format("2006-01-02 15:04:05"))
			cmd.Printf("Last Activity: %s\n", workspace.LastActivity.Format("2006-01-02 15:04:05"))

			if workspace.TicketID != "" {
				cmd.Printf("Ticket ID: %s\n", workspace.TicketID)
			}

			if workspace.ContainerID != "" {
				cmd.Printf("Container ID: %s\n", workspace.ContainerID)
			}

			if workspace.Description != "" {
				cmd.Printf("\nDescription:\n")
				cmd.Printf("------------\n")
				cmd.Printf("%s\n", workspace.Description)
			}

			if len(workspace.Metadata) > 0 {
				cmd.Printf("\nMetadata:\n")
				cmd.Printf("---------\n")
				for key, value := range workspace.Metadata {
					cmd.Printf("  %s: %s\n", key, value)
				}
			}

			return nil
		},
	}

	// List workspaces command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all active workspaces",
		Long:  "Display all currently active workspaces and their status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.workspaceManager == nil {
				return fmt.Errorf("workspace manager not initialized")
			}

			workspaces, err := app.workspaceManager.ListWorkspaces()
			if err != nil {
				return fmt.Errorf("failed to list workspaces: %w", err)
			}

			if len(workspaces) == 0 {
				cmd.Println("No workspaces found.")
				return nil
			}

			cmd.Printf("Found %d workspace(s):\n\n", len(workspaces))
			for _, ws := range workspaces {
				cmd.Printf("📁 %s (%s)\n", ws.TaskName, ws.Status)
				if ws.Description != "" {
					shortDesc := workspace.TruncateDescription(ws.Description, 80)
					cmd.Printf("   %s\n", shortDesc)
				}
				cmd.Println()
			}

			return nil
		},
	}

	// Clear workspaces command
	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all workspaces",
		Long:  "Remove all workspaces in the current repository. This action cannot be undone.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.workspaceManager == nil {
				return fmt.Errorf("workspace manager not initialized")
			}

			// Check if --force flag is set
			force, _ := cmd.Flags().GetBool("force")

			// Get current workspaces
			workspaces, err := app.workspaceManager.ListWorkspaces()
			if err != nil {
				return fmt.Errorf("failed to list workspaces: %w", err)
			}

			if len(workspaces) == 0 {
				cmd.Println("No workspaces found to clear.")
				return nil
			}

			// Show confirmation prompt unless --force is used
			if !force {
				cmd.Printf("This will remove %d workspace(s):\n", len(workspaces))
				for _, ws := range workspaces {
					cmd.Printf("  - %s (ID: %s)\n", ws.TaskName, ws.ID)
				}
				cmd.Print("\nThis action cannot be undone. Are you sure? (y/N): ")

				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					cmd.Println("Operation cancelled.")
					return nil
				}
			}

			// Clear all workspaces
			cmd.Printf("Clearing %d workspace(s)...\n", len(workspaces))
			for _, ws := range workspaces {
				if err := app.workspaceManager.DeleteWorkspace(ws.ID); err != nil {
					cmd.Printf("Warning: failed to delete workspace %s: %v\n", ws.TaskName, err)
				} else {
					cmd.Printf("✅ Removed workspace: %s\n", ws.TaskName)
				}
			}

			cmd.Println("All workspaces cleared successfully!")
			return nil
		},
	}

	// Add --force flag to clear command
	clearCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Directory command - change to workspace directory
	dirCmd := &cobra.Command{
		Use:   "dir [workspace-id-or-name]",
		Short: "Change to workspace directory",
		Long:  "Print the directory path of a workspace for use with cd or other commands",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.workspaceManager == nil {
				return fmt.Errorf("workspace manager not initialized")
			}

			identifier := args[0]

			// Try to find workspace by ID or task name
			var workspace *types.Workspace
			var err error

			// First try by ID
			workspace, err = app.workspaceManager.GetWorkspace(identifier)
			if err != nil {
				// If not found by ID, try by task name
				workspace, err = app.workspaceManager.GetWorkspaceByTaskName(identifier)
				if err != nil {
					return fmt.Errorf("workspace not found: %s", identifier)
				}
			}

			// Print the workspace directory path
			cmd.Printf("%s\n", workspace.Path)
			return nil
		},
	}

	// Git command - run git commands in workspace
	gitCmd := &cobra.Command{
		Use:   "git [workspace-id-or-name] [git-command...]",
		Short: "Run git commands in workspace",
		Long:  "Execute git commands in the context of a specific workspace",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.workspaceManager == nil {
				return fmt.Errorf("workspace manager not initialized")
			}

			identifier := args[0]
			gitArgs := args[1:]

			// Try to find workspace by ID or task name
			var workspace *types.Workspace
			var err error

			// First try by ID
			workspace, err = app.workspaceManager.GetWorkspace(identifier)
			if err != nil {
				// If not found by ID, try by task name
				workspace, err = app.workspaceManager.GetWorkspaceByTaskName(identifier)
				if err != nil {
					return fmt.Errorf("workspace not found: %s", identifier)
				}
			}

			// Change to workspace directory and run git command
			gitCmd := exec.Command("git", gitArgs...)
			gitCmd.Dir = workspace.Path
			gitCmd.Stdout = os.Stdout
			gitCmd.Stderr = os.Stderr
			gitCmd.Stdin = os.Stdin

			return gitCmd.Run()
		},
	}

	workspaceCmd.AddCommand(createCmd, listCmd, describeCmd, clearCmd, dirCmd, gitCmd)
	app.rootCmd.AddCommand(workspaceCmd)

	// Add alias for workspace command
	wsCmd := &cobra.Command{
		Use:    "ws",
		Short:  "Alias for workspace command",
		Long:   "Alias for the workspace command. Use 'cw ws' instead of 'cw workspace'",
		Hidden: true, // Hide from help since it's just an alias
	}

	// Copy all subcommands from workspace to ws
	for _, subCmd := range workspaceCmd.Commands() {
		wsCmd.AddCommand(subCmd)
	}

	app.rootCmd.AddCommand(wsCmd)
}

// addTicketCommands adds ticket management commands
func (app *App) addTicketCommands() {
	ticketCmd := &cobra.Command{
		Use:   "ticket",
		Short: "Manage tickets and create workspaces from them",
		Long:  "Fetch ticket information and create workspaces automatically from GitHub/GitLab/Jira/Linear tickets",
	}

	// Create workspace from ticket command
	createFromTicketCmd := &cobra.Command{
		Use:   "create [ticket-id]",
		Short: "Create workspace from ticket",
		Long:  "Fetch ticket information and create a workspace with appropriate branch naming",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ticketID := args[0]
			fmt.Printf("Creating workspace from ticket: %s\n", ticketID)
			fmt.Println("(This feature is not yet implemented)")
			return nil
		},
	}

	ticketCmd.AddCommand(createFromTicketCmd)
	app.rootCmd.AddCommand(ticketCmd)
}

// addTaskCommands adds task management commands
func (app *App) addTaskCommands() {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage queued tasks",
		Long:  "Create, list, describe, and manage queued tasks for AI agents to work on",
	}

	// Create task command
	createCmd := &cobra.Command{
		Use:   "create [task-name]",
		Short: "Create a new queued task",
		Long:  "Create a new task that will be queued for AI agents to work on",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			taskName := args[0]

			// Get description from flags
			message, _ := cmd.Flags().GetString("message")
			messageFile, _ := cmd.Flags().GetString("message-file")
			priority, _ := cmd.Flags().GetInt("priority")
			ticketID, _ := cmd.Flags().GetString("ticket-id")
			url, _ := cmd.Flags().GetString("url")
			estimatedMinutes, _ := cmd.Flags().GetInt("estimated-minutes")
			estimatedCost, _ := cmd.Flags().GetFloat64("estimated-cost")
			currency, _ := cmd.Flags().GetString("currency")

			var description string
			if messageFile != "" {
				// Read description from file
				content, err := os.ReadFile(messageFile)
				if err != nil {
					return fmt.Errorf("failed to read description from file: %w", err)
				}
				description = string(content)
			} else if message != "" {
				description = message
			}

			// Create task request
			req := &types.CreateTaskRequest{
				Name:             taskName,
				Description:      description,
				TicketID:         ticketID,
				URL:              url,
				Priority:         priority,
				EstimatedMinutes: estimatedMinutes,
				EstimatedCost:    estimatedCost,
				Currency:         currency,
				Metadata:         make(map[string]string),
				Tags:             []string{},
			}

			cmd.Printf("Creating task: %s\n", taskName)
			task, err := app.taskManager.CreateTask(req)
			if err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}

			cmd.Printf("✅ Task created successfully!\n")
			cmd.Printf("   ID: %s\n", task.ID)
			cmd.Printf("   Name: %s\n", task.Name)
			cmd.Printf("   Status: %s\n", task.Status)
			cmd.Printf("   Priority: %d\n", task.Priority)
			if description != "" {
				if len(description) > 100 {
					cmd.Printf("   Description: %s...\n", description[:100])
				} else {
					cmd.Printf("   Description: %s\n", description)
				}
			}

			return nil
		},
	}

	// Add flags to create command
	createCmd.Flags().StringP("message", "m", "", "Description of what the task is trying to accomplish")
	createCmd.Flags().String("message-file", "", "File containing the description of what the task is trying to accomplish")
	createCmd.Flags().IntP("priority", "p", 0, "Priority of the task (higher number = higher priority)")
	createCmd.Flags().String("ticket-id", "", "External ticket ID (e.g., GitHub #123)")
	createCmd.Flags().String("url", "", "Optional URL related to the task")
	createCmd.Flags().Int("estimated-minutes", 0, "Estimated completion time in minutes")
	createCmd.Flags().Float64("estimated-cost", 0, "Estimated cost for AI agent usage")
	createCmd.Flags().String("currency", "USD", "Currency for cost tracking")

	// Start task command (creates task and workspace, moves to in_progress)
	startCmd := &cobra.Command{
		Use:   "start [task-name]",
		Short: "Start working on a task",
		Long:  "Create a task, create a workspace if needed, and move the task to in_progress status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			taskName := args[0]

			// Get description from flags
			message, _ := cmd.Flags().GetString("message")
			messageFile, _ := cmd.Flags().GetString("message-file")
			priority, _ := cmd.Flags().GetInt("priority")
			ticketID, _ := cmd.Flags().GetString("ticket-id")
			url, _ := cmd.Flags().GetString("url")
			estimatedMinutes, _ := cmd.Flags().GetInt("estimated-minutes")
			estimatedCost, _ := cmd.Flags().GetFloat64("estimated-cost")
			currency, _ := cmd.Flags().GetString("currency")

			var description string
			if messageFile != "" {
				// Read description from file
				content, err := os.ReadFile(messageFile)
				if err != nil {
					return fmt.Errorf("failed to read description from file: %w", err)
				}
				description = string(content)
			} else if message != "" {
				description = message
			}

			// Check if task already exists
			existingTask, err := app.taskManager.GetTaskByName(taskName)
			if err == nil {
				// Task exists, check if it has a workspace
				if existingTask.WorkspaceID == "" {
					// No workspace, create one
					cmd.Printf("Task '%s' exists but has no workspace. Creating workspace...\n", taskName)

					// Create workspace for the task
					err = app.taskManager.CreateWorkspaceForTask(existingTask, description)
					if err != nil {
						return fmt.Errorf("failed to create workspace: %w", err)
					}

					// Save the updated task
					_, err = app.taskManager.UpdateTask(&types.UpdateTaskRequest{
						TaskID:        existingTask.ID,
						WorkspaceID:   &existingTask.WorkspaceID,
						WorkspacePath: &existingTask.WorkspacePath,
						BranchName:    &existingTask.BranchName,
						SourceRepo:    &existingTask.SourceRepo,
						BaseBranch:    &existingTask.BaseBranch,
					})
					if err != nil {
						return fmt.Errorf("failed to update task with workspace info: %w", err)
					}

					cmd.Printf("✅ Created workspace for existing task\n")
					cmd.Printf("   Workspace ID: %s\n", existingTask.WorkspaceID)
					cmd.Printf("   Workspace Path: %s\n", existingTask.WorkspacePath)
				}

				// Move task to in_progress
				status := types.TaskStatusInProgress
				updateReq := &types.UpdateTaskRequest{
					TaskID: existingTask.ID,
					Status: &status,
				}
				updatedTask, err := app.taskManager.UpdateTask(updateReq)
				if err != nil {
					return fmt.Errorf("failed to update task status: %w", err)
				}

				cmd.Printf("✅ Started working on task: %s\n", taskName)
				cmd.Printf("   Task ID: %s\n", updatedTask.ID)
				cmd.Printf("   Status: %s\n", updatedTask.Status)
				if updatedTask.WorkspaceID != "" {
					cmd.Printf("   Workspace ID: %s\n", updatedTask.WorkspaceID)
				}

				return nil
			}

			// Task doesn't exist, create it
			cmd.Printf("Creating new task: %s\n", taskName)

			// Create task request
			taskReq := &types.CreateTaskRequest{
				Name:             taskName,
				Description:      description,
				TicketID:         ticketID,
				URL:              url,
				Priority:         priority,
				EstimatedMinutes: estimatedMinutes,
				EstimatedCost:    estimatedCost,
				Currency:         currency,
				Metadata:         make(map[string]string),
				Tags:             []string{},
			}

			task, err := app.taskManager.CreateTask(taskReq)
			if err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}

			// Create workspace for the task
			err = app.taskManager.CreateWorkspaceForTask(task, description)
			if err != nil {
				return fmt.Errorf("failed to create workspace: %w", err)
			}

			// Update task with workspace information and move to in_progress
			status := types.TaskStatusInProgress
			updateReq := &types.UpdateTaskRequest{
				TaskID:        task.ID,
				Status:        &status,
				WorkspaceID:   &task.WorkspaceID,
				WorkspacePath: &task.WorkspacePath,
				BranchName:    &task.BranchName,
				SourceRepo:    &task.SourceRepo,
				BaseBranch:    &task.BaseBranch,
			}
			updatedTask, err := app.taskManager.UpdateTask(updateReq)
			if err != nil {
				return fmt.Errorf("failed to update task: %w", err)
			}

			cmd.Printf("✅ Task created and started successfully!\n")
			cmd.Printf("   Task ID: %s\n", updatedTask.ID)
			cmd.Printf("   Task Name: %s\n", updatedTask.Name)
			cmd.Printf("   Status: %s\n", updatedTask.Status)
			cmd.Printf("   Workspace ID: %s\n", task.WorkspaceID)
			cmd.Printf("   Workspace Path: %s\n", task.WorkspacePath)
			cmd.Printf("   Branch: %s\n", task.BranchName)
			if description != "" {
				if len(description) > 100 {
					cmd.Printf("   Description: %s...\n", description[:100])
				} else {
					cmd.Printf("   Description: %s\n", description)
				}
			}

			return nil
		},
	}

	// Add flags to start command (same as create command)
	startCmd.Flags().StringP("message", "m", "", "Description of what the task is trying to accomplish")
	startCmd.Flags().String("message-file", "", "File containing the description of what the task is trying to accomplish")
	startCmd.Flags().IntP("priority", "p", 0, "Priority of the task (higher number = higher priority)")
	startCmd.Flags().String("ticket-id", "", "External ticket ID (e.g., GitHub #123)")
	startCmd.Flags().String("url", "", "Optional URL related to the task")
	startCmd.Flags().Int("estimated-minutes", 0, "Estimated completion time in minutes")
	startCmd.Flags().Float64("estimated-cost", 0, "Estimated cost for AI agent usage")
	startCmd.Flags().String("currency", "USD", "Currency for cost tracking")

	// List tasks command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Long:  "Display all tasks with their status, priority, and basic information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			// Get filter flags
			statusFilter, _ := cmd.Flags().GetStringSlice("status")
			searchFilter, _ := cmd.Flags().GetString("search")

			// Build filter
			var filter *types.TaskFilter
			if len(statusFilter) > 0 || searchFilter != "" {
				filter = &types.TaskFilter{
					Search: searchFilter,
				}
				if len(statusFilter) > 0 {
					for _, status := range statusFilter {
						filter.Status = append(filter.Status, types.TaskStatus(status))
					}
				}
			}

			tasks, err := app.taskManager.ListTasks(filter)
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
				if task.EstimatedCost > 0 {
					cmd.Printf("   Estimated Cost: %.2f %s\n", task.EstimatedCost, task.Currency)
				}
				cmd.Println()
			}

			return nil
		},
	}

	// Add filter flags to list command
	listCmd.Flags().StringSlice("status", []string{}, "Filter by status (queued, in_progress, completed, failed, cancelled, paused)")

	listCmd.Flags().String("search", "", "Search in task name and description")

	// Describe task command
	describeCmd := &cobra.Command{
		Use:   "describe [task-id-or-name]",
		Short: "Show detailed information about a task",
		Long:  "Display detailed information about a specific task including full description, notes, and metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			identifier := args[0]

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
			cmd.Printf("ID: %s\n", task.ID)
			cmd.Printf("Status: %s\n", task.Status)
			cmd.Printf("Priority: %d\n", task.Priority)
			cmd.Printf("Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
			cmd.Printf("Last Activity: %s\n", task.LastActivity.Format("2006-01-02 15:04:05"))

			if task.StartedAt != nil {
				cmd.Printf("Started: %s\n", task.StartedAt.Format("2006-01-02 15:04:05"))
			}

			if task.CompletedAt != nil {
				cmd.Printf("Completed: %s\n", task.CompletedAt.Format("2006-01-02 15:04:05"))
			}

			if task.TicketID != "" {
				cmd.Printf("Ticket ID: %s\n", task.TicketID)
			}

			if task.URL != "" {
				cmd.Printf("URL: %s\n", task.URL)
			}

			if task.EstimatedCost > 0 {
				cmd.Printf("Estimated Cost: %.2f %s\n", task.EstimatedCost, task.Currency)
			}

			if task.ActualCost > 0 {
				cmd.Printf("Actual Cost: %.2f %s\n", task.ActualCost, task.Currency)
			}

			if task.EstimatedMinutes > 0 {
				cmd.Printf("Estimated Time: %d minutes\n", task.EstimatedMinutes)
			}

			if task.ActualMinutes > 0 {
				cmd.Printf("Actual Time: %d minutes\n", task.ActualMinutes)
			}

			if task.WorkspaceID != "" {
				cmd.Printf("Workspace ID: %s\n", task.WorkspaceID)
			}

			if task.WorkspacePath != "" {
				cmd.Printf("Workspace Path: %s\n", task.WorkspacePath)
			}

			if task.Description != "" {
				cmd.Printf("\nDescription:\n")
				cmd.Printf("------------\n")
				cmd.Printf("%s\n", task.Description)
			}

			if len(task.Tags) > 0 {
				cmd.Printf("\nTags:\n")
				cmd.Printf("-----\n")
				for _, tag := range task.Tags {
					cmd.Printf("  %s\n", tag)
				}
			}

			if len(task.Metadata) > 0 {
				cmd.Printf("\nMetadata:\n")
				cmd.Printf("---------\n")
				for key, value := range task.Metadata {
					cmd.Printf("  %s: %s\n", key, value)
				}
			}

			if task.ErrorMessage != "" {
				cmd.Printf("\nError Message:\n")
				cmd.Printf("--------------\n")
				cmd.Printf("%s\n", task.ErrorMessage)
			}

			return nil
		},
	}

	// Update task command
	updateCmd := &cobra.Command{
		Use:   "update [task-id-or-name]",
		Short: "Update task properties",
		Long:  "Update task status, priority, description, or other properties",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			identifier := args[0]

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

			// Get update flags
			status, _ := cmd.Flags().GetString("status")
			priority, _ := cmd.Flags().GetInt("priority")
			description, _ := cmd.Flags().GetString("description")
			actualMinutes, _ := cmd.Flags().GetInt("actual-minutes")
			actualCost, _ := cmd.Flags().GetFloat64("actual-cost")

			// Build update request
			req := &types.UpdateTaskRequest{
				TaskID: task.ID,
			}

			if status != "" {
				taskStatus := types.TaskStatus(status)
				req.Status = &taskStatus
			}

			if priority >= 0 {
				req.Priority = &priority
			}

			if description != "" {
				req.Description = &description
			}

			if actualCost >= 0 {
				req.ActualCost = &actualCost
			}

			if actualMinutes >= 0 {
				req.ActualMinutes = &actualMinutes
			}

			// Update the task
			updatedTask, err := app.taskManager.UpdateTask(req)
			if err != nil {
				return fmt.Errorf("failed to update task: %w", err)
			}

			cmd.Printf("✅ Task updated successfully!\n")
			cmd.Printf("   ID: %s\n", updatedTask.ID)
			cmd.Printf("   Name: %s\n", updatedTask.Name)
			cmd.Printf("   Status: %s\n", updatedTask.Status)
			cmd.Printf("   Priority: %d\n", updatedTask.Priority)

			return nil
		},
	}

	// Add flags to update command
	updateCmd.Flags().String("status", "", "New status (queued, in_progress, completed, failed, cancelled, paused)")
	updateCmd.Flags().Int("priority", -1, "New priority")
	updateCmd.Flags().String("description", "", "New description")
	updateCmd.Flags().Int("actual-minutes", -1, "Actual time spent in minutes")
	updateCmd.Flags().Float64("actual-cost", -1, "Actual cost spent")

	// Delete task command
	deleteCmd := &cobra.Command{
		Use:   "delete [task-id-or-name]",
		Short: "Delete a task",
		Long:  "Remove a task from the queue. This action cannot be undone.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			identifier := args[0]

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

			// Check if --force flag is set
			force, _ := cmd.Flags().GetBool("force")

			if !force {
				cmd.Printf("Are you sure you want to delete task '%s'? This action cannot be undone. (y/N): ", task.Name)
				// For now, we'll just inform the user. In a real implementation,
				// you might want to use a proper prompt library
				cmd.Printf("   Use --force flag to skip confirmation\n")
				return nil
			}

			err = app.taskManager.DeleteTask(task.ID)
			if err != nil {
				return fmt.Errorf("failed to delete task: %w", err)
			}

			cmd.Printf("✅ Task '%s' deleted successfully!\n", task.Name)
			return nil
		},
	}

	// Add --force flag to delete command
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Task cd command - change to task workspace directory
	taskCdCmd := &cobra.Command{
		Use:   "cd [task-id-or-name]",
		Short: "Change to task workspace directory",
		Long:  "Change the current directory to a task's workspace directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			identifier := args[0]

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

			// Get workspace path
			workspacePath, err := app.taskManager.GetTaskWorkspacePath(task.ID)
			if err != nil {
				return err
			}

			// Change to the workspace directory
			if err := os.Chdir(workspacePath); err != nil {
				return fmt.Errorf("failed to change to workspace directory: %w", err)
			}

			cmd.Printf("Changed to workspace directory: %s\n", workspacePath)
			return nil
		},
	}

	// Task dir command - print task workspace directory
	taskDirCmd := &cobra.Command{
		Use:   "dir [task-id-or-name]",
		Short: "Print task workspace directory",
		Long:  "Print the directory path of a task's workspace for use with cd or other commands",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			identifier := args[0]

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

			// Get workspace path
			workspacePath, err := app.taskManager.GetTaskWorkspacePath(task.ID)
			if err != nil {
				return err
			}

			// Print the workspace directory path
			cmd.Printf("%s\n", workspacePath)
			return nil
		},
	}

	// Task git command - run git commands in task workspace
	taskGitCmd := &cobra.Command{
		Use:   "git [task-id-or-name] [git-command...]",
		Short: "Run git commands in task workspace",
		Long:  "Execute git commands in the context of a specific task's workspace",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			identifier := args[0]
			gitArgs := args[1:]

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

			// Run git command in task workspace
			return app.taskManager.RunGitInTaskWorkspace(task.ID, gitArgs)
		},
	}

	// Next command
	nextCmd := &cobra.Command{
		Use:   "next",
		Short: "Get the next task in the queue",
		Long:  "Return the highest priority queued task",
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			nextTask, err := app.taskManager.GetNextQueuedTask()
			if err != nil {
				return fmt.Errorf("failed to get next task: %w", err)
			}

			cmd.Printf("🎯 Next Task in Queue\n")
			cmd.Printf("   Name: %s\n", nextTask.Name)
			cmd.Printf("   ID: %s\n", nextTask.ID)
			cmd.Printf("   Priority: %d\n", nextTask.Priority)
			cmd.Printf("   Status: %s\n", nextTask.Status)
			if nextTask.Description != "" {
				if len(nextTask.Description) > 100 {
					cmd.Printf("   Description: %s...\n", nextTask.Description[:100])
				} else {
					cmd.Printf("   Description: %s\n", nextTask.Description)
				}
			}

			return nil
		},
	}

	// Stats command
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show task statistics",
		Long:  "Display statistics about tasks including counts by status and completion times",
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.taskManager == nil {
				return fmt.Errorf("task manager not initialized")
			}

			stats, err := app.taskManager.GetTaskStats(nil)
			if err != nil {
				return fmt.Errorf("failed to get task stats: %w", err)
			}

			cmd.Printf("📊 Task Statistics\n")
			cmd.Printf("=================\n\n")
			cmd.Printf("Total Tasks: %d\n", stats.Total)
			cmd.Printf("Created Today: %d\n", stats.CreatedToday)
			cmd.Printf("Completed Today: %d\n", stats.CompletedToday)
			cmd.Printf("Total Time Spent: %d minutes\n", stats.TotalTimeMinutes)
			cmd.Printf("Average Completion Time: %.1f minutes\n", stats.AverageCompletionMinutes)
			if stats.TotalCost > 0 {
				cmd.Printf("Total Cost: %.2f %s\n", stats.TotalCost, stats.Currency)
			}

			if len(stats.ByStatus) > 0 {
				cmd.Printf("\nBy Status:\n")
				cmd.Printf("----------\n")
				for status, count := range stats.ByStatus {
					cmd.Printf("  %s: %d\n", status, count)
				}
			}

			if len(stats.ByPriority) > 0 {
				cmd.Printf("\nBy Priority:\n")
				cmd.Printf("------------\n")
				for priority, count := range stats.ByPriority {
					cmd.Printf("  Priority %d: %d\n", priority, count)
				}
			}

			return nil
		},
	}

	taskCmd.AddCommand(createCmd, startCmd, listCmd, describeCmd, updateCmd, deleteCmd, taskCdCmd, taskDirCmd, taskGitCmd, nextCmd, statsCmd)
	app.rootCmd.AddCommand(taskCmd)

	// Add alias for task command
	tCmd := &cobra.Command{
		Use:    "t",
		Short:  "Alias for task command",
		Long:   "Alias for the task command. Use 'cw t' instead of 'cw task'",
		Hidden: true, // Hide from help since it's just an alias
	}

	// Copy all subcommands from task to t
	for _, subCmd := range taskCmd.Commands() {
		tCmd.AddCommand(subCmd)
	}

	app.rootCmd.AddCommand(tCmd)
}

// addAgentCommands adds agent management commands
func (app *App) addAgentCommands() {
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage AI agents",
		Long:  "Configure and run AI agents (Cursor, Claude, Gemini) in isolated workspaces",
	}

	// Run agent command
	runCmd := &cobra.Command{
		Use:   "run [agent-type] [workspace-name]",
		Short: "Run an AI agent in a workspace",
		Long:  "Start an AI agent (cursor, claude, gemini) in the specified workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentType := args[0]
			workspaceName := args[1]
			fmt.Printf("Running %s agent in workspace: %s\n", agentType, workspaceName)
			fmt.Println("(This feature is not yet implemented)")
			return nil
		},
	}

	agentCmd.AddCommand(runCmd)
	app.rootCmd.AddCommand(agentCmd)
}

// Run executes the CLI application with the given arguments
func (app *App) Run(args []string) error {
	app.rootCmd.SetArgs(args[1:]) // Skip the program name
	return app.rootCmd.Execute()
}

// printInstructions displays the main instructions for using the cowork CLI
func (app *App) printInstructions() {
	instructions := `
cowork (cw) — Multi-Agent Repo Orchestrator

Quick Start:
  cw workspace create <task-name>     # Create a new isolated workspace
  cw workspace list                   # List all workspaces
  cw workspace describe <id-or-name>  # Show detailed workspace information
  cw workspace clear                  # Clear all workspaces (with confirmation)
  cw ticket create <ticket-id>        # Create workspace from ticket
  cw agent run <type> <workspace>     # Run AI agent in workspace
  cw version                          # Show version information

For detailed help on any command:
  cw <command> --help

Examples:
  cw workspace create oauth-refresh
  cw workspace clear --force          # Clear all workspaces without confirmation
  cw ticket create 123
  cw agent run cursor oauth-refresh

This is a development version. Many features are not yet implemented.
`
	fmt.Print(instructions)
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

	// Create .cw directory
	cwDir := filepath.Join(repoInfo.Path, ".cw")
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
	configManager := config.NewManager()
	projectConfigPath := filepath.Join(repoInfo.Path, ".cwconfig")
	if _, err := os.Stat(projectConfigPath); os.IsNotExist(err) {
		// Create default project configuration
		defaultConfig := configManager.GetDefaultConfig()
		if err := configManager.SaveProject(defaultConfig); err != nil {
			return fmt.Errorf("failed to create project configuration: %w", err)
		}
		cmd.Printf("✅ Created project configuration: %s\n", projectConfigPath)
	} else {
		cmd.Printf("ℹ️  Project configuration already exists: %s\n", projectConfigPath)
	}

	// Check for .gitignore and offer to add .cw/ to it
	gitignorePath := filepath.Join(repoInfo.Path, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		// .gitignore exists, check if .cw/ is already in it
		content, err := os.ReadFile(gitignorePath)
		if err != nil {
			return fmt.Errorf("failed to read .gitignore: %w", err)
		}

		if !containsLine(string(content), ".cw/") {
			cmd.Printf("📝 Found .gitignore file. Would you like to add .cw/ to it? (y/N): ")

			// For now, we'll just inform the user. In a real implementation,
			// you might want to use a proper prompt library
			cmd.Printf("   To add .cw/ to .gitignore, run: echo '.cw/' >> .gitignore\n")
		} else {
			cmd.Printf("✅ .cw/ is already in .gitignore\n")
		}
	} else {
		cmd.Printf("📝 No .gitignore found. Consider creating one and adding .cw/ to it.\n")
	}

	cmd.Printf("\n🎉 Project initialized successfully!\n")
	cmd.Printf("   You can now use: cw workspace create <task-name>\n")
	cmd.Printf("   Configuration: cw config show\n")

	return nil
}

// containsLine checks if a string contains a specific line
func containsLine(content, line string) bool {
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == strings.TrimSpace(line) {
			return true
		}
	}
	return false
}

// getStatusIcon returns an appropriate icon for the task status
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

// truncateString truncates a string to the specified length and adds "..." if needed
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}
