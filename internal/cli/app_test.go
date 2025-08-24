package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestNewApp_WithValidVersionInfo tests that NewApp creates a valid application
// with the provided version information and properly initializes all commands
func TestNewApp_WithValidVersionInfo(t *testing.T) {
	// Test case: Creating a new app with valid version information should succeed
	// and return an app with the correct version details
	expectedVersion := "1.2.3"
	expectedBuildDate := "2024-01-15"
	expectedGitCommit := "abc123"

	app := NewApp(expectedVersion, expectedBuildDate, expectedGitCommit)

	// Verify the app was created with correct version information
	if app.version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, app.version)
	}

	if app.buildDate != expectedBuildDate {
		t.Errorf("Expected build date %s, got %s", expectedBuildDate, app.buildDate)
	}

	if app.gitCommit != expectedGitCommit {
		t.Errorf("Expected git commit %s, got %s", expectedGitCommit, app.gitCommit)
	}

	// Verify the root command was created
	if app.rootCmd == nil {
		t.Error("Expected root command to be initialized, got nil")
	}

	// Verify the root command has the correct use and short description
	if app.rootCmd.Use != "cw" {
		t.Errorf("Expected root command use 'cw', got %s", app.rootCmd.Use)
	}

	expectedShort := "Multi-Agent Repo Orchestrator"
	if app.rootCmd.Short != expectedShort {
		t.Errorf("Expected root command short description '%s', got '%s'", expectedShort, app.rootCmd.Short)
	}
}

// TestNewApp_WithEmptyVersionInfo tests that NewApp handles empty version information gracefully
func TestNewApp_WithEmptyVersionInfo(t *testing.T) {
	// Test case: Creating a new app with empty version information should still succeed
	// and create a valid application structure
	app := NewApp("", "", "")

	// Verify the app was created successfully
	if app == nil {
		t.Error("Expected app to be created, got nil")
	}

	// Verify the root command was created
	if app.rootCmd == nil {
		t.Error("Expected root command to be initialized, got nil")
	}

	// Verify empty version information is handled
	if app.version != "" {
		t.Errorf("Expected empty version, got %s", app.version)
	}
}

// TestApp_CommandStructure tests that all expected commands are properly added to the application
func TestApp_CommandStructure(t *testing.T) {
	// Test case: The application should have all the expected top-level commands
	// including version, workspace, ticket, and agent commands
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Get all subcommands
	commands := app.rootCmd.Commands()

	// Define expected command names
	expectedCommands := []string{"version", "workspace", "ticket", "agent"}

	// Create a map of found commands for easy lookup
	foundCommands := make(map[string]bool)
	for _, cmd := range commands {
		foundCommands[cmd.Name()] = true
	}

	// Verify all expected commands are present
	for _, expectedCmd := range expectedCommands {
		if !foundCommands[expectedCmd] {
			t.Errorf("Expected command '%s' not found in application", expectedCmd)
		}
	}

	// Verify we have the expected number of commands
	expectedCommandCount := len(expectedCommands)
	if len(commands) < expectedCommandCount {
		t.Errorf("Expected at least %d commands, got %d", expectedCommandCount, len(commands))
	}
}

// TestApp_VersionCommand tests that the version command displays correct information
func TestApp_VersionCommand(t *testing.T) {
	// Test case: The version command should display the correct version information
	// when executed
	expectedVersion := "2.0.0"
	expectedBuildDate := "2024-02-01"
	expectedGitCommit := "def456"

	app := NewApp(expectedVersion, expectedBuildDate, expectedGitCommit)

	// Find the version command
	var versionCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "version" {
			versionCmd = cmd
			break
		}
	}

	if versionCmd == nil {
		t.Fatal("Version command not found")
	}

	// Verify the version command has the correct short description
	expectedShort := "Show detailed version information"
	if versionCmd.Short != expectedShort {
		t.Errorf("Expected version command short description '%s', got '%s'", expectedShort, versionCmd.Short)
	}

	// Verify the version command has the correct long description
	expectedLong := "Display the version, build date, and git commit information for the cowork CLI"
	if versionCmd.Long != expectedLong {
		t.Errorf("Expected version command long description '%s', got '%s'", expectedLong, versionCmd.Long)
	}
}

// TestApp_WorkspaceCommands tests that workspace commands are properly structured
func TestApp_WorkspaceCommands(t *testing.T) {
	// Test case: The workspace command should have the expected subcommands
	// including create and list commands
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the workspace command
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("Workspace command not found")
	}

	// Verify workspace command has the correct short description
	expectedShort := "Manage isolated workspaces"
	if workspaceCmd.Short != expectedShort {
		t.Errorf("Expected workspace command short description '%s', got '%s'", expectedShort, workspaceCmd.Short)
	}

	// Get workspace subcommands
	workspaceSubcommands := workspaceCmd.Commands()

	// Define expected subcommand names
	expectedSubcommands := []string{"create", "list"}

	// Create a map of found subcommands for easy lookup
	foundSubcommands := make(map[string]bool)
	for _, cmd := range workspaceSubcommands {
		foundSubcommands[cmd.Name()] = true
	}

	// Verify all expected subcommands are present
	for _, expectedSubcmd := range expectedSubcommands {
		if !foundSubcommands[expectedSubcmd] {
			t.Errorf("Expected workspace subcommand '%s' not found", expectedSubcmd)
		}
	}
}

// TestApp_TicketCommands tests that ticket commands are properly structured
func TestApp_TicketCommands(t *testing.T) {
	// Test case: The ticket command should have the expected subcommands
	// including create command
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the ticket command
	var ticketCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "ticket" {
			ticketCmd = cmd
			break
		}
	}

	if ticketCmd == nil {
		t.Fatal("Ticket command not found")
	}

	// Verify ticket command has the correct short description
	expectedShort := "Manage tickets and create workspaces from them"
	if ticketCmd.Short != expectedShort {
		t.Errorf("Expected ticket command short description '%s', got '%s'", expectedShort, ticketCmd.Short)
	}

	// Get ticket subcommands
	ticketSubcommands := ticketCmd.Commands()

	// Verify create subcommand is present
	foundCreate := false
	for _, cmd := range ticketSubcommands {
		if cmd.Name() == "create" {
			foundCreate = true
			break
		}
	}

	if !foundCreate {
		t.Error("Expected ticket subcommand 'create' not found")
	}
}

// TestApp_AgentCommands tests that agent commands are properly structured
func TestApp_AgentCommands(t *testing.T) {
	// Test case: The agent command should have the expected subcommands
	// including run command
	app := NewApp("1.0.0", "2024-01-01", "test")

	// Find the agent command
	var agentCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Name() == "agent" {
			agentCmd = cmd
			break
		}
	}

	if agentCmd == nil {
		t.Fatal("Agent command not found")
	}

	// Verify agent command has the correct short description
	expectedShort := "Manage AI agents"
	if agentCmd.Short != expectedShort {
		t.Errorf("Expected agent command short description '%s', got '%s'", expectedShort, agentCmd.Short)
	}

	// Get agent subcommands
	agentSubcommands := agentCmd.Commands()

	// Verify run subcommand is present
	foundRun := false
	for _, cmd := range agentSubcommands {
		if cmd.Name() == "run" {
			foundRun = true
			break
		}
	}

	if !foundRun {
		t.Error("Expected agent subcommand 'run' not found")
	}
}
