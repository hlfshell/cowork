package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hlfshell/cowork/internal/config"
	"github.com/hlfshell/cowork/internal/git"
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
}

// NewApp creates a new CLI application with the specified version information
func NewApp(version, buildDate, gitCommit string) *App {
	// Check and initialize global configuration on first run
	configManager := config.NewManager()
	if err := ensureGlobalConfigExists(configManager); err != nil {
		fmt.Printf("Warning: failed to initialize global configuration: %v\n", err)
	}

	// Initialize workspace manager
	workspaceManager, err := workspace.NewManager(300)
	if err != nil {
		fmt.Printf("Warning: failed to initialize workspace manager: %v\n", err)
		workspaceManager = nil
	}

	app := &App{
		version:          version,
		buildDate:        buildDate,
		gitCommit:        gitCommit,
		workspaceManager: workspaceManager,
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

	// Add workspace commands (placeholder for future implementation)
	app.addWorkspaceCommands()

	// Add ticket commands (placeholder for future implementation)
	app.addTicketCommands()

	// Add agent commands (placeholder for future implementation)
	app.addAgentCommands()

	// Add config commands
	app.addConfigCommands()
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

	workspaceCmd.AddCommand(createCmd, listCmd, describeCmd, clearCmd)
	app.rootCmd.AddCommand(workspaceCmd)
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
