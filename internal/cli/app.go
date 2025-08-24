package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// App represents the main CLI application
type App struct {
	rootCmd   *cobra.Command
	version   string
	buildDate string
	gitCommit string
}

// NewApp creates a new CLI application with the specified version information
func NewApp(version, buildDate, gitCommit string) *App {
	app := &App{
		version:   version,
		buildDate: buildDate,
		gitCommit: gitCommit,
	}

	app.setupCommands()
	return app
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

	// Add workspace commands (placeholder for future implementation)
	app.addWorkspaceCommands()

	// Add ticket commands (placeholder for future implementation)
	app.addTicketCommands()

	// Add agent commands (placeholder for future implementation)
	app.addAgentCommands()
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

// addWorkspaceCommands adds workspace management commands
func (app *App) addWorkspaceCommands() {
	workspaceCmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage isolated workspaces",
		Long:  "Create, list, and manage isolated workspaces for different tasks",
	}

	// Create workspace command
	createCmd := &cobra.Command{
		Use:   "create [task-name]",
		Short: "Create a new isolated workspace",
		Long:  "Create a new isolated workspace for the specified task with the given isolation level",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskName := args[0]
			fmt.Printf("Creating workspace for task: %s\n", taskName)
			fmt.Println("(This feature is not yet implemented)")
			return nil
		},
	}

	// List workspaces command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all active workspaces",
		Long:  "Display all currently active workspaces and their status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing workspaces:")
			fmt.Println("(This feature is not yet implemented)")
			return nil
		},
	}

	workspaceCmd.AddCommand(createCmd, listCmd)
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
  cw ticket create <ticket-id>        # Create workspace from ticket
  cw agent run <type> <workspace>     # Run AI agent in workspace
  cw version                          # Show version information

For detailed help on any command:
  cw <command> --help

Examples:
  cw workspace create oauth-refresh
  cw ticket create 123
  cw agent run cursor oauth-refresh

This is a development version. Many features are not yet implemented.
`
	fmt.Print(instructions)
}
