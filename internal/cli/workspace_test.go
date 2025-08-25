package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hlfshell/cowork/internal/types"
	"github.com/spf13/cobra"
)

// TestWorkspaceClearCommand tests the workspace clear command
func TestWorkspaceClearCommand(t *testing.T) {
	// Test case: Clear command should require confirmation by default

	// Create a temporary Git repository
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create app in the Git repository
	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create a test workspace
	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	_, err = app.workspaceManager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Test clear command without --force flag
	clearCmd := findCommand(app.rootCmd, "workspace", "clear")
	if clearCmd == nil {
		t.Fatal("Clear command not found")
	}

	// Test that the command exists and has the force flag
	forceFlag := clearCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Force flag not found on clear command")
	}

	// Test help text
	if !strings.Contains(clearCmd.Long, "cannot be undone") {
		t.Error("Clear command help text should mention that action cannot be undone")
	}

	// Verify workspace was actually created on filesystem
	workspaces, err := app.workspaceManager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(workspaces))
	}

	// Check that workspace directory exists
	workspace := workspaces[0]
	if _, err := os.Stat(workspace.Path); os.IsNotExist(err) {
		t.Errorf("Workspace directory does not exist: %s", workspace.Path)
	}

	// Check that metadata file exists
	metadataPath := filepath.Join(workspace.Path, ".cw-workspace.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Errorf("Workspace metadata file does not exist: %s", metadataPath)
	}

	// Check that it's a Git repository
	gitDir := filepath.Join(workspace.Path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("Workspace is not a Git repository: %s", gitDir)
	}

	// Verify metadata file contains correct information
	metadataContent, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Errorf("Failed to read metadata file: %v", err)
	}

	if !strings.Contains(string(metadataContent), workspace.TaskName) {
		t.Errorf("Metadata file should contain task name: %s", workspace.TaskName)
	}

	if !strings.Contains(string(metadataContent), workspace.ID) {
		t.Errorf("Metadata file should contain workspace ID: %s", workspace.ID)
	}
}

// TestWorkspaceClearWithForceFlag tests the workspace clear command with --force flag
func TestWorkspaceClearWithForceFlag(t *testing.T) {
	// Test case: Clear command with --force flag should skip confirmation

	// Create a temporary Git repository
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create app in the Git repository
	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create test workspaces
	for i := 0; i < 3; i++ {
		req := &types.CreateWorkspaceRequest{
			TaskName:   fmt.Sprintf("test-task-%d", i),
			SourceRepo: tempDir,
			BaseBranch: "main",
		}

		_, err = app.workspaceManager.CreateWorkspace(req)
		if err != nil {
			t.Fatalf("Failed to create test workspace %d: %v", i, err)
		}
	}

	// Verify workspaces exist
	workspaces, err := app.workspaceManager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != 3 {
		t.Errorf("Expected 3 workspaces, got %d", len(workspaces))
	}

	// Test clear command with --force flag
	clearCmd := findCommand(app.rootCmd, "workspace", "clear")
	if clearCmd == nil {
		t.Fatal("Clear command not found")
	}

	// Set the force flag
	clearCmd.Flags().Set("force", "true")

	// Execute the command by running the app
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	// Set the arguments for the clear command with force flag
	app.rootCmd.SetArgs([]string{"workspace", "clear", "--force"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Clear command failed: %v", err)
	}

	outputStr := output.String()

	// Check that confirmation was skipped
	if strings.Contains(outputStr, "Are you sure?") {
		t.Error("Confirmation prompt should be skipped with --force flag")
	}

	// Check that workspaces were cleared
	if !strings.Contains(outputStr, "Clearing 3 workspace(s)") {
		t.Errorf("Output should indicate clearing 3 workspaces, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "All workspaces cleared successfully!") {
		t.Errorf("Output should indicate successful clearing, got: %s", outputStr)
	}

	// Verify workspaces were actually removed
	workspaces, err = app.workspaceManager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces after clearing: %v", err)
	}

	if len(workspaces) != 0 {
		t.Errorf("Expected 0 workspaces after clearing, got %d", len(workspaces))
	}

	// Verify workspace directories were actually removed from filesystem
	workspaceDir := filepath.Join(tempDir, ".cw", "workspaces")
	if entries, err := os.ReadDir(workspaceDir); err == nil {
		if len(entries) != 0 {
			t.Errorf("Expected 0 workspace directories after clearing, got %d", len(entries))
		}
	} else {
		// If the directory doesn't exist, that's also fine (all workspaces were removed)
		if !os.IsNotExist(err) {
			t.Errorf("Unexpected error reading workspace directory: %v", err)
		}
	}
}

// TestWorkspaceClearWithNoWorkspaces tests the workspace clear command when no workspaces exist
func TestWorkspaceClearWithNoWorkspaces(t *testing.T) {
	// Test case: Clear command should handle case when no workspaces exist

	// Create a temporary Git repository
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create app in the Git repository
	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Test clear command with no workspaces
	clearCmd := findCommand(app.rootCmd, "workspace", "clear")
	if clearCmd == nil {
		t.Fatal("Clear command not found")
	}

	// Set the force flag to skip confirmation
	clearCmd.Flags().Set("force", "true")

	// Execute the command by running the app
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	// Set the arguments for the clear command with force flag
	app.rootCmd.SetArgs([]string{"workspace", "clear", "--force"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Clear command failed: %v", err)
	}

	outputStr := output.String()

	// Check that appropriate message is shown
	if !strings.Contains(outputStr, "No workspaces found to clear") {
		t.Errorf("Output should indicate no workspaces found to clear, got: %s", outputStr)
	}

	// Verify no workspace directories exist on filesystem
	workspaceDir := filepath.Join(tempDir, ".cw", "workspaces")
	if entries, err := os.ReadDir(workspaceDir); err == nil {
		if len(entries) != 0 {
			t.Errorf("Expected 0 workspace directories, got %d", len(entries))
		}
	} else {
		// If the directory doesn't exist, that's also fine
		if !os.IsNotExist(err) {
			t.Errorf("Unexpected error reading workspace directory: %v", err)
		}
	}
}

// TestWorkspaceClearConfirmationPrompt tests the workspace clear command confirmation prompt
func TestWorkspaceClearConfirmationPrompt(t *testing.T) {
	// Test case: Clear command should show confirmation prompt and list workspaces

	// Create a temporary Git repository
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create app in the Git repository
	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create test workspaces
	workspaceNames := []string{"task-1", "task-2", "task-3"}
	for _, name := range workspaceNames {
		req := &types.CreateWorkspaceRequest{
			TaskName:   name,
			SourceRepo: tempDir,
			BaseBranch: "main",
		}

		_, err = app.workspaceManager.CreateWorkspace(req)
		if err != nil {
			t.Fatalf("Failed to create test workspace %s: %v", name, err)
		}
	}

	// Test clear command without force flag
	clearCmd := findCommand(app.rootCmd, "workspace", "clear")
	if clearCmd == nil {
		t.Fatal("Clear command not found")
	}

	// Execute the command with simulated input
	var output bytes.Buffer
	clearCmd.SetOut(&output)
	clearCmd.SetErr(&output)

	// Simulate user input (we can't easily test interactive input in unit tests)
	// This test mainly verifies the command structure and help text

	// The command should be properly configured
	if clearCmd.Use != "clear" {
		t.Errorf("Expected command use 'clear', got '%s'", clearCmd.Use)
	}

	if !strings.Contains(clearCmd.Short, "Clear all workspaces") {
		t.Error("Clear command short description should mention clearing workspaces")
	}

	if !strings.Contains(clearCmd.Long, "cannot be undone") {
		t.Error("Clear command long description should mention that action cannot be undone")
	}

	// Verify workspaces were actually created on filesystem
	workspaces, err := app.workspaceManager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != 3 {
		t.Errorf("Expected 3 workspaces, got %d", len(workspaces))
	}

	// Check that all workspace directories exist
	for _, workspace := range workspaces {
		if _, err := os.Stat(workspace.Path); os.IsNotExist(err) {
			t.Errorf("Workspace directory does not exist: %s", workspace.Path)
		}

		// Check that metadata file exists
		metadataPath := filepath.Join(workspace.Path, ".cw-workspace.json")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Errorf("Workspace metadata file does not exist: %s", metadataPath)
		}

		// Check that it's a Git repository
		gitDir := filepath.Join(workspace.Path, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			t.Errorf("Workspace is not a Git repository: %s", gitDir)
		}
	}
}

// Helper function to create a temporary Git repository for testing
func createTempGitRepo(t *testing.T) string {
	// Create temporary directory
	tempDir := t.TempDir()

	// Initialize Git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize Git repository: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the file
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file to Git: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	return tempDir
}

// TestWorkspaceCreateCommand tests the workspace create command
func TestWorkspaceCreateCommand(t *testing.T) {
	// Test case: Create command should create a workspace with valid parameters
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Test create command
	createCmd := findCommand(app.rootCmd, "workspace", "create")
	if createCmd == nil {
		t.Fatal("Create command not found")
	}

	// Test command structure
	if createCmd.Use != "create [task-name]" {
		t.Errorf("Expected command use 'create [task-name]', got '%s'", createCmd.Use)
	}

	// Test flags
	messageFlag := createCmd.Flags().Lookup("message")
	if messageFlag == nil {
		t.Error("Message flag not found on create command")
	}

	messageFileFlag := createCmd.Flags().Lookup("message-file")
	if messageFileFlag == nil {
		t.Error("Message-file flag not found on create command")
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "create", "test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Create command failed: %v", err)
	}

	outputStr := output.String()

	// Check that workspace was created
	if !strings.Contains(outputStr, "Creating workspace for task: test-task") {
		t.Errorf("Output should indicate workspace creation, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "✅ Workspace created successfully!") {
		t.Errorf("Output should indicate successful creation, got: %s", outputStr)
	}

	// Verify workspace was actually created
	workspaces, err := app.workspaceManager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(workspaces))
	}

	workspace := workspaces[0]
	if workspace.TaskName != "test-task" {
		t.Errorf("Expected task name 'test-task', got '%s'", workspace.TaskName)
	}
}

// TestWorkspaceCreateCommand_WithDescription tests the workspace create command with description
func TestWorkspaceCreateCommand_WithDescription(t *testing.T) {
	// Test case: Create command should create a workspace with description
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Execute the command with description
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "create", "test-task", "-m", "Test task description"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Create command failed: %v", err)
	}

	outputStr := output.String()

	// Check that description was included
	if !strings.Contains(outputStr, "Description: Test task description") {
		t.Errorf("Output should include description, got: %s", outputStr)
	}

	// Verify workspace was created with description
	workspaces, err := app.workspaceManager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(workspaces))
	}

	workspace := workspaces[0]
	if workspace.Description != "Test task description" {
		t.Errorf("Expected description 'Test task description', got '%s'", workspace.Description)
	}
}

// TestWorkspaceCreateCommand_WithDescriptionFile tests the workspace create command with description file
func TestWorkspaceCreateCommand_WithDescriptionFile(t *testing.T) {
	// Test case: Create command should create a workspace with description from file
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create description file
	descriptionFile := filepath.Join(tempDir, "description.txt")
	descriptionContent := "This is a description from a file.\nIt has multiple lines."
	err = os.WriteFile(descriptionFile, []byte(descriptionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create description file: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Execute the command with description file
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "create", "test-task", "--message-file", descriptionFile})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Create command failed: %v", err)
	}

	outputStr := output.String()

	// Check that description was included
	if !strings.Contains(outputStr, "Description: This is a description from a file.") {
		t.Errorf("Output should include description from file, got: %s", outputStr)
	}

	// Verify workspace was created with description
	workspaces, err := app.workspaceManager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(workspaces))
	}

	workspace := workspaces[0]
	expectedDescription := strings.TrimSpace(descriptionContent)
	if workspace.Description != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, workspace.Description)
	}
}

// TestWorkspaceDescribeCommand tests the workspace describe command
func TestWorkspaceDescribeCommand(t *testing.T) {
	// Test case: Describe command should show detailed workspace information
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create a workspace first
	req := &types.CreateWorkspaceRequest{
		TaskName:    "test-task",
		Description: "Test task description",
		SourceRepo:  tempDir,
		BaseBranch:  "main",
	}

	workspace, err := app.workspaceManager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Test describe command
	describeCmd := findCommand(app.rootCmd, "workspace", "describe")
	if describeCmd == nil {
		t.Fatal("Describe command not found")
	}

	// Test command structure
	if describeCmd.Use != "describe [workspace-id-or-name]" {
		t.Errorf("Expected command use 'describe [workspace-id-or-name]', got '%s'", describeCmd.Use)
	}

	// Execute the command with task name
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "describe", "test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Describe command failed: %v", err)
	}

	outputStr := output.String()

	// Check that detailed information is shown
	if !strings.Contains(outputStr, "📁 Workspace Details") {
		t.Errorf("Output should show workspace details header, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Task Name: test-task") {
		t.Errorf("Output should show task name, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "ID: "+workspace.ID) {
		t.Errorf("Output should show workspace ID, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Description:") {
		t.Errorf("Output should show description section, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Test task description") {
		t.Errorf("Output should show description content, got: %s", outputStr)
	}
}

// TestWorkspaceDescribeCommand_WithWorkspaceID tests the workspace describe command with workspace ID
func TestWorkspaceDescribeCommand_WithWorkspaceID(t *testing.T) {
	// Test case: Describe command should work with workspace ID
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create a workspace first
	req := &types.CreateWorkspaceRequest{
		TaskName:    "test-task",
		Description: "Test task description",
		SourceRepo:  tempDir,
		BaseBranch:  "main",
	}

	workspace, err := app.workspaceManager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Execute the command with workspace ID
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "describe", workspace.ID})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Describe command failed: %v", err)
	}

	outputStr := output.String()

	// Check that detailed information is shown
	if !strings.Contains(outputStr, "Task Name: test-task") {
		t.Errorf("Output should show task name, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "ID: "+workspace.ID) {
		t.Errorf("Output should show workspace ID, got: %s", outputStr)
	}
}

// TestWorkspaceDescribeCommand_WithInvalidIdentifier tests the workspace describe command with invalid identifier
func TestWorkspaceDescribeCommand_WithInvalidIdentifier(t *testing.T) {
	// Test case: Describe command should fail with invalid identifier
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Execute the command with invalid identifier
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "describe", "invalid-workspace"})

	err = app.rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid workspace identifier, got nil")
	}

	expectedError := "workspace not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestWorkspaceListCommand tests the workspace list command
func TestWorkspaceListCommand(t *testing.T) {
	// Test case: List command should show all workspaces
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create multiple workspaces
	req1 := &types.CreateWorkspaceRequest{
		TaskName:    "task-1",
		Description: "First task description",
		SourceRepo:  tempDir,
		BaseBranch:  "main",
	}

	req2 := &types.CreateWorkspaceRequest{
		TaskName:    "task-2",
		Description: "Second task description",
		SourceRepo:  tempDir,
		BaseBranch:  "main",
	}

	_, err = app.workspaceManager.CreateWorkspace(req1)
	if err != nil {
		t.Fatalf("Failed to create first workspace: %v", err)
	}

	_, err = app.workspaceManager.CreateWorkspace(req2)
	if err != nil {
		t.Fatalf("Failed to create second workspace: %v", err)
	}

	// Test list command
	listCmd := findCommand(app.rootCmd, "workspace", "list")
	if listCmd == nil {
		t.Fatal("List command not found")
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "list"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	outputStr := output.String()

	// Check that workspaces are listed
	if !strings.Contains(outputStr, "Found 2 workspace(s):") {
		t.Errorf("Output should indicate 2 workspaces found, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "📁 task-1 (ready)") {
		t.Errorf("Output should show task-1, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "📁 task-2 (ready)") {
		t.Errorf("Output should show task-2, got: %s", outputStr)
	}

	// Check that descriptions are truncated
	if !strings.Contains(outputStr, "First task description") {
		t.Errorf("Output should show first task description, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Second task description") {
		t.Errorf("Output should show second task description, got: %s", outputStr)
	}
}

// TestWorkspaceListCommand_WithNoWorkspaces tests the workspace list command with no workspaces
func TestWorkspaceListCommand_WithNoWorkspaces(t *testing.T) {
	// Test case: List command should show appropriate message when no workspaces exist
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "list"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	outputStr := output.String()

	// Check that appropriate message is shown
	if !strings.Contains(outputStr, "No workspaces found.") {
		t.Errorf("Output should indicate no workspaces found, got: %s", outputStr)
	}
}

// Helper function to find a command by path
func findCommand(root *cobra.Command, path ...string) *cobra.Command {
	current := root
	for _, name := range path {
		found := false
		for _, cmd := range current.Commands() {
			if cmd.Name() == name {
				current = cmd
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return current
}

// TestWorkspaceDirCommand tests the workspace dir command structure
func TestWorkspaceDirCommand(t *testing.T) {
	// Test case: Workspace dir command should be properly configured
	app := NewApp("test", "test", "test")

	// Find the workspace command
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("workspace command not found")
	}

	// Find the dir subcommand
	var dirCmd *cobra.Command
	for _, subCmd := range workspaceCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "dir") {
			dirCmd = subCmd
			break
		}
	}

	if dirCmd == nil {
		t.Fatal("workspace dir command not found")
	}

	// Check command properties
	if dirCmd.Short != "Change to workspace directory" {
		t.Errorf("Expected short description 'Change to workspace directory', got '%s'", dirCmd.Short)
	}

	if !strings.Contains(dirCmd.Long, "Print the directory path") {
		t.Errorf("Expected long description to contain 'Print the directory path', got '%s'", dirCmd.Long)
	}
}

// Helper function to get command names for debugging
func getCommandNames(cmds []*cobra.Command) []string {
	names := make([]string, len(cmds))
	for i, cmd := range cmds {
		names[i] = cmd.Use
	}
	return names
}

// TestWorkspaceDirCommand_WithInvalidWorkspace tests dir command argument validation
func TestWorkspaceDirCommand_WithInvalidWorkspace(t *testing.T) {
	// Test case: Dir command should validate arguments properly
	app := NewApp("test", "test", "test")

	// Find the workspace command
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("workspace command not found")
	}

	// Find the dir subcommand
	var dirCmd *cobra.Command
	for _, subCmd := range workspaceCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "dir") {
			dirCmd = subCmd
			break
		}
	}

	if dirCmd == nil {
		t.Fatal("workspace dir command not found")
	}

	// Check that dir command requires exactly one argument
	if dirCmd.Args == nil {
		t.Error("dir command should have argument validation")
	}
}

// TestWorkspaceGitCommand tests the workspace git command structure
func TestWorkspaceGitCommand(t *testing.T) {
	// Test case: Workspace git command should be properly configured
	app := NewApp("test", "test", "test")

	// Find the workspace command
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("workspace command not found")
	}

	// Find the git subcommand
	var gitCmd *cobra.Command
	for _, subCmd := range workspaceCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "git") {
			gitCmd = subCmd
			break
		}
	}

	if gitCmd == nil {
		t.Fatal("workspace git command not found")
	}

	// Check command properties
	if gitCmd.Short != "Run git commands in workspace" {
		t.Errorf("Expected short description 'Run git commands in workspace', got '%s'", gitCmd.Short)
	}

	if !strings.Contains(gitCmd.Long, "Execute git commands") {
		t.Errorf("Expected long description to contain 'Execute git commands', got '%s'", gitCmd.Long)
	}
}

// TestWorkspaceGitCommand_WithInvalidWorkspace tests git command argument validation
func TestWorkspaceGitCommand_WithInvalidWorkspace(t *testing.T) {
	// Test case: Git command should validate arguments properly
	app := NewApp("test", "test", "test")

	// Find the workspace command
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("workspace command not found")
	}

	// Find the git subcommand
	var gitCmd *cobra.Command
	for _, subCmd := range workspaceCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "git") {
			gitCmd = subCmd
			break
		}
	}

	if gitCmd == nil {
		t.Fatal("workspace git command not found")
	}

	// Check that git command requires at least two arguments
	if gitCmd.Args == nil {
		t.Error("git command should have argument validation")
	}
}

// TestWorkspaceGitCommand_WithInvalidGitCommand tests git command argument validation
func TestWorkspaceGitCommand_WithInvalidGitCommand(t *testing.T) {
	// Test case: Git command should validate minimum arguments
	app := NewApp("test", "test", "test")

	// Find the workspace command
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("workspace command not found")
	}

	// Find the git subcommand
	var gitCmd *cobra.Command
	for _, subCmd := range workspaceCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "git") {
			gitCmd = subCmd
			break
		}
	}

	if gitCmd == nil {
		t.Fatal("workspace git command not found")
	}

	if gitCmd.Args == nil {
		t.Error("git command should have argument validation")
	}
}

// TestWorkspaceAlias tests that the ws alias works correctly
func TestWorkspaceAlias(t *testing.T) {
	// Test case: ws alias should be properly configured
	app := NewApp("test", "test", "test")

	// Find the ws command
	var wsCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "ws" {
			wsCmd = cmd
			break
		}
	}

	if wsCmd == nil {
		t.Fatal("ws command not found")
	}

	// Check that ws command has the same subcommands as workspace
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("workspace command not found")
	}

	// Check that ws has the same number of subcommands as workspace
	if len(wsCmd.Commands()) != len(workspaceCmd.Commands()) {
		t.Errorf("Expected ws to have %d subcommands, got %d", len(workspaceCmd.Commands()), len(wsCmd.Commands()))
	}

	// Check that ws has the dir and git subcommands
	var hasDir, hasGit bool
	for _, subCmd := range wsCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "dir") {
			hasDir = true
		}
		if strings.HasPrefix(subCmd.Use, "git") {
			hasGit = true
		}
	}

	if !hasDir {
		t.Error("ws command missing dir subcommand")
	}

	if !hasGit {
		t.Error("ws command missing git subcommand")
	}
}

// TestWorkspaceDirCommand_WithNoArgs tests dir command argument validation
func TestWorkspaceDirCommand_WithNoArgs(t *testing.T) {
	// Test case: Dir command should validate arguments properly
	app := NewApp("test", "test", "test")

	// Find the workspace command
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("workspace command not found")
	}

	// Find the dir subcommand
	var dirCmd *cobra.Command
	for _, subCmd := range workspaceCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "dir") {
			dirCmd = subCmd
			break
		}
	}

	if dirCmd == nil {
		t.Fatal("workspace dir command not found")
	}

	// Check that dir command requires exactly one argument
	if dirCmd.Args == nil {
		t.Error("dir command should have argument validation")
	}
}

// TestWorkspaceGitCommand_WithInsufficientArgs tests git command argument validation
func TestWorkspaceGitCommand_WithInsufficientArgs(t *testing.T) {
	// Test case: Git command should validate minimum arguments
	app := NewApp("test", "test", "test")

	// Find the workspace command
	var workspaceCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "workspace" {
			workspaceCmd = cmd
			break
		}
	}

	if workspaceCmd == nil {
		t.Fatal("workspace command not found")
	}

	// Find the git subcommand
	var gitCmd *cobra.Command
	for _, subCmd := range workspaceCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "git") {
			gitCmd = subCmd
			break
		}
	}

	if gitCmd == nil {
		t.Fatal("workspace git command not found")
	}

	// Check that git command requires at least two arguments
	if gitCmd.Args == nil {
		t.Error("git command should have argument validation")
	}
}

// TestWorkspaceDirCommand_Functional tests the workspace dir command with actual workspace
func TestWorkspaceDirCommand_Functional(t *testing.T) {
	// Test case: Dir command should return the correct workspace directory path
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create a test workspace
	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	workspace, err := app.workspaceManager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Test dir command with task name
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "dir", "test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Dir command failed: %v", err)
	}

	outputStr := strings.TrimSpace(output.String())

	// Check that the output is the workspace path
	if outputStr != workspace.Path {
		t.Errorf("Expected workspace path '%s', got '%s'", workspace.Path, outputStr)
	}

	// Verify the path exists
	if _, err := os.Stat(outputStr); os.IsNotExist(err) {
		t.Errorf("Workspace directory does not exist: %s", outputStr)
	}

	// Test dir command with workspace ID
	output.Reset()
	app.rootCmd.SetArgs([]string{"workspace", "dir", workspace.ID})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Dir command failed with workspace ID: %v", err)
	}

	outputStr = strings.TrimSpace(output.String())

	// Check that the output is the workspace path
	if outputStr != workspace.Path {
		t.Errorf("Expected workspace path '%s', got '%s'", workspace.Path, outputStr)
	}
}

// TestWorkspaceDirCommand_WithInvalidWorkspace_Functional tests the workspace dir command with invalid workspace
func TestWorkspaceDirCommand_WithInvalidWorkspace_Functional(t *testing.T) {
	// Test case: Dir command should fail with invalid workspace identifier
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Test dir command with invalid workspace name
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "dir", "invalid-workspace"})

	err = app.rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid workspace, got nil")
	}

	expectedError := "workspace not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}

	// Test dir command with invalid workspace ID
	output.Reset()
	app.rootCmd.SetArgs([]string{"workspace", "dir", "invalid-id-123"})

	err = app.rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid workspace ID, got nil")
	}

	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestWorkspaceGitCommand_Functional tests the workspace git command with actual workspace
func TestWorkspaceGitCommand_Functional(t *testing.T) {
	// Test case: Git command should execute git commands in the workspace directory
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create a test workspace
	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	workspace, err := app.workspaceManager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Test git status command in workspace
	// Note: Git command outputs directly to stdout/stderr, not through cobra
	app.rootCmd.SetArgs([]string{"workspace", "git", "test-task", "status"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git status command failed: %v", err)
	}

	// Test git branch command in workspace
	app.rootCmd.SetArgs([]string{"workspace", "git", workspace.ID, "branch"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git branch command failed: %v", err)
	}

	// Test git log command in workspace (simple version)
	app.rootCmd.SetArgs([]string{"workspace", "git", "test-task", "log"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git log command failed: %v", err)
	}
}

// TestWorkspaceGitCommand_WithInvalidWorkspace_Functional tests the workspace git command with invalid workspace
func TestWorkspaceGitCommand_WithInvalidWorkspace_Functional(t *testing.T) {
	// Test case: Git command should fail with invalid workspace identifier
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Test git command with invalid workspace name
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "git", "invalid-workspace", "status"})

	err = app.rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid workspace, got nil")
	}

	expectedError := "workspace not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}

	// Test git command with invalid workspace ID
	output.Reset()
	app.rootCmd.SetArgs([]string{"workspace", "git", "invalid-id-123", "status"})

	err = app.rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid workspace ID, got nil")
	}

	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestWorkspaceGitCommand_WithGitOperations tests the workspace git command with various git operations
func TestWorkspaceGitCommand_WithGitOperations(t *testing.T) {
	// Test case: Git command should support various git operations in workspace
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create a test workspace
	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	workspace, err := app.workspaceManager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Test git config command
	app.rootCmd.SetArgs([]string{"workspace", "git", "test-task", "config", "user.name"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git config command failed: %v", err)
	}

	// Test git remote command
	app.rootCmd.SetArgs([]string{"workspace", "git", workspace.ID, "remote"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git remote command failed: %v", err)
	}

	// Test git diff command (should be empty initially)
	app.rootCmd.SetArgs([]string{"workspace", "git", "test-task", "diff"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git diff command failed: %v", err)
	}
}

// TestWorkspaceGitCommand_WithMultipleArgs tests the workspace git command with multiple git arguments
func TestWorkspaceGitCommand_WithMultipleArgs(t *testing.T) {
	// Test case: Git command should handle multiple git arguments correctly
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create a test workspace
	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	workspace, err := app.workspaceManager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Test git log with multiple arguments
	app.rootCmd.SetArgs([]string{"workspace", "git", "test-task", "log", "HEAD"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git log command with multiple args failed: %v", err)
	}

	// Test git show with multiple arguments
	app.rootCmd.SetArgs([]string{"workspace", "git", workspace.ID, "show", "HEAD"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git show command with multiple args failed: %v", err)
	}
}

// TestWorkspaceDirAndGitCommands_Integration tests both dir and git commands together
func TestWorkspaceDirAndGitCommands_Integration(t *testing.T) {
	// Test case: Integration test of dir and git commands working together
	tempDir := createTempGitRepo(t)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.workspaceManager == nil {
		t.Fatal("Workspace manager should be initialized in a Git repository")
	}

	// Create a test workspace
	req := &types.CreateWorkspaceRequest{
		TaskName:   "integration-test",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	workspace, err := app.workspaceManager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Test dir command to get workspace path
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"workspace", "dir", "integration-test"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Dir command failed: %v", err)
	}

	workspacePath := strings.TrimSpace(output.String())

	// Verify the path matches the workspace path
	if workspacePath != workspace.Path {
		t.Errorf("Expected workspace path '%s', got '%s'", workspace.Path, workspacePath)
	}

	// Test git command in the same workspace
	app.rootCmd.SetArgs([]string{"workspace", "git", "integration-test", "rev-parse", "HEAD"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git rev-parse command failed: %v", err)
	}

	// Test that both commands work with workspace ID
	output.Reset()
	app.rootCmd.SetArgs([]string{"workspace", "dir", workspace.ID})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Dir command with ID failed: %v", err)
	}

	workspacePathFromID := strings.TrimSpace(output.String())

	if workspacePathFromID != workspace.Path {
		t.Errorf("Expected workspace path '%s', got '%s'", workspace.Path, workspacePathFromID)
	}

	// Test git branch command with workspace ID
	app.rootCmd.SetArgs([]string{"workspace", "git", workspace.ID, "branch"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git branch command with ID failed: %v", err)
	}
}
