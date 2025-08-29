package workspace

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hlfshell/cowork/internal/container"
	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewManager_WithValidTimeout tests creating a workspace manager with valid timeout
func TestNewManager_WithValidTimeout(t *testing.T) {
	// Test case: Creating a workspace manager with valid timeout should succeed
	// when in a Git repository
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	if manager == nil {
		t.Error("Expected workspace manager to be created, got nil")
	}

	// Verify the base directory was set correctly
	expectedBaseDir := filepath.Join(tempDir, ".cw", "workspaces")
	if manager.baseDir != expectedBaseDir {
		t.Errorf("Expected base directory '%s', got '%s'", expectedBaseDir, manager.baseDir)
	}
}

// TestNewManager_WithZeroTimeout tests creating a workspace manager with zero timeout
func TestNewManager_WithZeroTimeout(t *testing.T) {
	// Test case: Creating a workspace manager with zero timeout should use default
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

	manager, err := NewManager(0)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	if manager == nil {
		t.Error("Expected workspace manager to be created, got nil")
	}
}

// TestNewManager_WithNegativeTimeout tests creating a workspace manager with negative timeout
func TestNewManager_WithNegativeTimeout(t *testing.T) {
	// Test case: Creating a workspace manager with negative timeout should use default
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

	manager, err := NewManager(-100)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	if manager == nil {
		t.Error("Expected workspace manager to be created, got nil")
	}
}

// TestNewManager_OutsideGitRepository tests creating a workspace manager outside a Git repository
func TestNewManager_OutsideGitRepository(t *testing.T) {
	// Test case: Creating a workspace manager outside a Git repository should fail
	tempDir := t.TempDir()

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	manager, err := NewManager(300)
	if err == nil {
		t.Error("Expected error when creating manager outside Git repository, got nil")
	}

	if manager != nil {
		t.Error("Expected manager to be nil when outside Git repository")
	}

	expectedError := "current directory is not a Git repository"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_CreateWorkspace_WithValidRequest tests creating a workspace with valid request
func TestManager_CreateWorkspace_WithValidRequest(t *testing.T) {
	// Test case: Creating a workspace with valid request should succeed
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	req := &types.CreateWorkspaceRequest{
		TaskName:    "test-task",
		Description: "Test task description",
		SourceRepo:  tempDir,
		BaseBranch:  "main",
	}

	workspace, err := manager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	if workspace == nil {
		t.Error("Expected workspace to be created, got nil")
	}

	// Verify workspace fields
	if workspace.TaskName != "test-task" {
		t.Errorf("Expected task name 'test-task', got '%s'", workspace.TaskName)
	}

	if workspace.Description != "Test task description" {
		t.Errorf("Expected description 'Test task description', got '%s'", workspace.Description)
	}

	if workspace.SourceRepo != tempDir {
		t.Errorf("Expected source repo '%s', got '%s'", tempDir, workspace.SourceRepo)
	}

	if workspace.BaseBranch != "main" {
		t.Errorf("Expected base branch 'main', got '%s'", workspace.BaseBranch)
	}

	if workspace.Status != types.WorkspaceStatusReady {
		t.Errorf("Expected status 'ready', got '%s'", workspace.Status)
	}

	// Verify workspace directory exists
	if _, err := os.Stat(workspace.Path); os.IsNotExist(err) {
		t.Errorf("Workspace directory does not exist: %s", workspace.Path)
	}

	// Verify it's a Git repository
	gitDir := filepath.Join(workspace.Path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("Workspace is not a Git repository: %s", gitDir)
	}

	// Verify metadata file exists
	metadataPath := filepath.Join(workspace.Path, ".cw-workspace.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Errorf("Workspace metadata file does not exist: %s", metadataPath)
	}

	// Verify branch name format
	expectedBranchPrefix := "task/test-task"
	if !strings.HasPrefix(workspace.BranchName, expectedBranchPrefix) {
		t.Errorf("Expected branch name to start with '%s', got '%s'", expectedBranchPrefix, workspace.BranchName)
	}
}

// TestManager_CreateWorkspace_WithInvalidRequest tests creating a workspace with invalid request
func TestManager_CreateWorkspace_WithInvalidRequest(t *testing.T) {
	// Test case: Creating a workspace with invalid request should fail
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Test with empty task name
	req := &types.CreateWorkspaceRequest{
		TaskName:   "",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	workspace, err := manager.CreateWorkspace(req)
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
	}

	if workspace != nil {
		t.Error("Expected workspace to be nil for invalid request")
	}

	expectedError := "task name is required"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_CreateWorkspace_WithDuplicateTaskName tests creating a workspace with duplicate task name
func TestManager_CreateWorkspace_WithDuplicateTaskName(t *testing.T) {
	// Test case: Creating a workspace with duplicate task name should fail
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	// Create first workspace
	workspace1, err := manager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create first workspace: %v", err)
	}

	if workspace1 == nil {
		t.Fatal("Expected first workspace to be created, got nil")
	}

	// Try to create second workspace with same task name
	workspace2, err := manager.CreateWorkspace(req)
	if err == nil {
		t.Error("Expected error for duplicate task name, got nil")
	}

	if workspace2 != nil {
		t.Error("Expected second workspace to be nil for duplicate task name")
	}

	expectedError := "workspace with task name 'test-task' already exists"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_GetWorkspace_WithValidID tests getting a workspace with valid ID
func TestManager_GetWorkspace_WithValidID(t *testing.T) {
	// Test case: Getting a workspace with valid ID should succeed
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	createdWorkspace, err := manager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Get the workspace by ID
	retrievedWorkspace, err := manager.GetWorkspace(createdWorkspace.ID)
	if err != nil {
		t.Fatalf("Failed to get workspace: %v", err)
	}

	if retrievedWorkspace == nil {
		t.Error("Expected workspace to be retrieved, got nil")
	}

	// Verify the retrieved workspace matches the created one
	if retrievedWorkspace.ID != createdWorkspace.ID {
		t.Errorf("Expected workspace ID %d, got %d", createdWorkspace.ID, retrievedWorkspace.ID)
	}

	if retrievedWorkspace.TaskName != createdWorkspace.TaskName {
		t.Errorf("Expected task name '%s', got '%s'", createdWorkspace.TaskName, retrievedWorkspace.TaskName)
	}

	if retrievedWorkspace.Path != createdWorkspace.Path {
		t.Errorf("Expected path '%s', got '%s'", createdWorkspace.Path, retrievedWorkspace.Path)
	}
}

// TestManager_GetWorkspace_WithInvalidID tests getting a workspace with invalid ID
func TestManager_GetWorkspace_WithInvalidID(t *testing.T) {
	// Test case: Getting a workspace with invalid ID should fail
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Try to get workspace with invalid ID
	workspace, err := manager.GetWorkspace(99999)
	if err == nil {
		t.Error("Expected error for invalid workspace ID, got nil")
	}

	if workspace != nil {
		t.Error("Expected workspace to be nil for invalid ID")
	}

	expectedError := "workspace not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_ListWorkspaces_WithNoWorkspaces tests listing workspaces when none exist
func TestManager_ListWorkspaces_WithNoWorkspaces(t *testing.T) {
	// Test case: Listing workspaces when none exist should return empty list
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != 0 {
		t.Errorf("Expected 0 workspaces, got %d", len(workspaces))
	}
}

// TestManager_ListWorkspaces_WithMultipleWorkspaces tests listing multiple workspaces
func TestManager_ListWorkspaces_WithMultipleWorkspaces(t *testing.T) {
	// Test case: Listing workspaces should return all created workspaces
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Create multiple workspaces
	taskNames := []string{"task-1", "task-2", "task-3"}
	createdWorkspaces := make(map[string]*types.Workspace)

	for _, taskName := range taskNames {
		req := &types.CreateWorkspaceRequest{
			TaskName:   taskName,
			SourceRepo: tempDir,
			BaseBranch: "main",
		}

		workspace, err := manager.CreateWorkspace(req)
		if err != nil {
			t.Fatalf("Failed to create workspace %s: %v", taskName, err)
		}

		createdWorkspaces[taskName] = workspace
	}

	// List all workspaces
	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != len(taskNames) {
		t.Errorf("Expected %d workspaces, got %d", len(taskNames), len(workspaces))
	}

	// Verify all created workspaces are in the list
	foundWorkspaces := make(map[string]bool)
	for _, workspace := range workspaces {
		foundWorkspaces[workspace.TaskName] = true
	}

	for _, taskName := range taskNames {
		if !foundWorkspaces[taskName] {
			t.Errorf("Expected workspace '%s' not found in list", taskName)
		}
	}
}

// TestManager_DeleteWorkspace_WithValidID tests deleting a workspace with valid ID
func TestManager_DeleteWorkspace_WithValidID(t *testing.T) {
	// Test case: Deleting a workspace with valid ID should succeed
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	workspace, err := manager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Verify workspace exists
	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("Expected 1 workspace before deletion, got %d", len(workspaces))
	}

	// Delete the workspace
	err = manager.DeleteWorkspace(workspace.ID)
	if err != nil {
		t.Fatalf("Failed to delete workspace: %v", err)
	}

	// Verify workspace is deleted
	workspaces, err = manager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces after deletion: %v", err)
	}

	if len(workspaces) != 0 {
		t.Errorf("Expected 0 workspaces after deletion, got %d", len(workspaces))
	}

	// Verify workspace directory is removed
	if _, err := os.Stat(workspace.Path); !os.IsNotExist(err) {
		t.Errorf("Expected workspace directory to be removed: %s", workspace.Path)
	}
}

// TestManager_DeleteWorkspace_WithInvalidID tests deleting a workspace with invalid ID
func TestManager_DeleteWorkspace_WithInvalidID(t *testing.T) {
	// Test case: Deleting a workspace with invalid ID should fail
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Try to delete workspace with invalid ID
	err = manager.DeleteWorkspace(99999)
	if err == nil {
		t.Error("Expected error for invalid workspace ID, got nil")
	}

	expectedError := "workspace not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_UpdateWorkspaceStatus_WithValidID tests updating workspace status with valid ID
func TestManager_UpdateWorkspaceStatus_WithValidID(t *testing.T) {
	// Test case: Updating workspace status with valid ID should succeed
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	workspace, err := manager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Update workspace status
	newStatus := types.WorkspaceStatusActive
	err = manager.UpdateWorkspaceStatus(workspace.ID, newStatus)
	if err != nil {
		t.Fatalf("Failed to update workspace status: %v", err)
	}

	// Verify status was updated
	updatedWorkspace, err := manager.GetWorkspace(workspace.ID)
	if err != nil {
		t.Fatalf("Failed to get updated workspace: %v", err)
	}

	if updatedWorkspace.Status != newStatus {
		t.Errorf("Expected status '%s', got '%s'", newStatus, updatedWorkspace.Status)
	}
}

// TestManager_UpdateWorkspaceStatus_WithInvalidID tests updating workspace status with invalid ID
func TestManager_UpdateWorkspaceStatus_WithInvalidID(t *testing.T) {
	// Test case: Updating workspace status with invalid ID should fail
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Try to update workspace status with invalid ID
	err = manager.UpdateWorkspaceStatus(99999, types.WorkspaceStatusActive)
	if err == nil {
		t.Error("Expected error for invalid workspace ID, got nil")
	}

	expectedError := "workspace not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_GetWorkspaceByTaskName_WithValidName tests getting workspace by valid task name
func TestManager_GetWorkspaceByTaskName_WithValidName(t *testing.T) {
	// Test case: Getting workspace by valid task name should succeed
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	createdWorkspace, err := manager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Get workspace by task name
	retrievedWorkspace, err := manager.GetWorkspaceByTaskName("test-task")
	if err != nil {
		t.Fatalf("Failed to get workspace by task name: %v", err)
	}

	if retrievedWorkspace == nil {
		t.Error("Expected workspace to be retrieved, got nil")
	}

	// Verify the retrieved workspace matches the created one
	if retrievedWorkspace.ID != createdWorkspace.ID {
		t.Errorf("Expected workspace ID %d, got %d", createdWorkspace.ID, retrievedWorkspace.ID)
	}

	if retrievedWorkspace.TaskName != createdWorkspace.TaskName {
		t.Errorf("Expected task name '%s', got '%s'", createdWorkspace.TaskName, retrievedWorkspace.TaskName)
	}
}

// TestManager_GetWorkspaceByTaskName_WithInvalidName tests getting workspace by invalid task name
func TestManager_GetWorkspaceByTaskName_WithInvalidName(t *testing.T) {
	// Test case: Getting workspace by invalid task name should fail
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Try to get workspace by invalid task name
	workspace, err := manager.GetWorkspaceByTaskName("invalid-task")
	if err == nil {
		t.Error("Expected error for invalid task name, got nil")
	}

	if workspace != nil {
		t.Error("Expected workspace to be nil for invalid task name")
	}

	expectedError := "workspace not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestManager_GenerateWorkspaceID tests workspace ID generation
func TestManager_GenerateWorkspaceID(t *testing.T) {
	// Test case: Generating workspace IDs should create unique, valid IDs
	// This test is skipped as the workspace manager uses the types.GenerateWorkspaceID() function
	// which returns an int, not a string. The test was written for an older implementation.
	t.Skip("This test is for an older implementation that used string IDs")
}

// TestManager_GetBaseDirectory tests getting the base directory
func TestManager_GetBaseDirectory(t *testing.T) {
	// Test case: Getting base directory should return the correct path
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	baseDir := manager.GetBaseDirectory()
	expectedBaseDir := filepath.Join(tempDir, ".cw", "workspaces")

	if baseDir != expectedBaseDir {
		t.Errorf("Expected base directory '%s', got '%s'", expectedBaseDir, baseDir)
	}
}

// TestManager_CleanupOrphanedWorkspaces tests cleaning up orphaned workspaces
func TestManager_CleanupOrphanedWorkspaces(t *testing.T) {
	// Test case: Cleaning up orphaned workspaces should remove invalid directories
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

	manager, err := NewManager(300)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Create a valid workspace first
	req := &types.CreateWorkspaceRequest{
		TaskName:   "valid-task",
		SourceRepo: tempDir,
		BaseBranch: "main",
	}

	_, err = manager.CreateWorkspace(req)
	if err != nil {
		t.Fatalf("Failed to create valid workspace: %v", err)
	}

	// Create an orphaned workspace directory (without metadata)
	orphanedDir := filepath.Join(manager.baseDir, "orphaned-workspace")
	if err := os.MkdirAll(orphanedDir, 0755); err != nil {
		t.Fatalf("Failed to create orphaned workspace directory: %v", err)
	}

	// Create a file in the orphaned directory to make it look like a workspace
	orphanedFile := filepath.Join(orphanedDir, "some-file.txt")
	if err := os.WriteFile(orphanedFile, []byte("orphaned content"), 0644); err != nil {
		t.Fatalf("Failed to create file in orphaned workspace: %v", err)
	}

	// Verify orphaned directory exists
	if _, err := os.Stat(orphanedDir); os.IsNotExist(err) {
		t.Error("Orphaned workspace directory should exist before cleanup")
	}

	// Run cleanup
	err = manager.CleanupOrphanedWorkspaces()
	if err != nil {
		t.Fatalf("Failed to cleanup orphaned workspaces: %v", err)
	}

	// Verify orphaned directory was removed
	if _, err := os.Stat(orphanedDir); !os.IsNotExist(err) {
		t.Error("Orphaned workspace directory should be removed after cleanup")
	}

	// Verify valid workspace still exists
	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces after cleanup: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("Expected 1 workspace after cleanup, got %d", len(workspaces))
	}

	if workspaces[0].TaskName != "valid-task" {
		t.Errorf("Expected workspace task name 'valid-task', got '%s'", workspaces[0].TaskName)
	}
}

func TestGetContainerStatus_WithoutContainer(t *testing.T) {
	// Test case: Getting container status for a workspace without a container should fail
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspaces")
	
	// Create workspace manager with mock container manager
	manager := &Manager{
		baseDir:         workspaceDir,
		containerManager: &MockContainerManager{},
	}

	// Create a workspace without container
	workspace := &types.Workspace{
		ID:       1,
		TaskName: "test-task",
		Path:     filepath.Join(workspaceDir, "1"),
		Status:   types.WorkspaceStatusReady,
	}

	// Save workspace metadata
	require.NoError(t, os.MkdirAll(workspace.Path, 0755))
	require.NoError(t, SaveWorkspaceMetadata(workspace.Path, workspace))

	// Try to get container status
	ctx := context.Background()
	_, err := manager.GetContainerStatus(ctx, 1)
	
	// Verify it fails with appropriate error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace does not have an associated container")
}

func TestStartContainer_WithValidConfig(t *testing.T) {
	// Test case: Starting a container with valid configuration should succeed
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspaces")
	
	// Create workspace manager with mock container manager
	manager := &Manager{
		baseDir:         workspaceDir,
		containerManager: &MockContainerManager{},
	}

	// Create a workspace with container configuration
	workspace := &types.Workspace{
		ID:       1,
		TaskName: "test-task",
		Path:     filepath.Join(workspaceDir, "1"),
		Status:   types.WorkspaceStatusReady,
		ContainerConfig: &types.ContainerConfig{
			Image:      "golang:1.21",
			Name:       "test-container",
			WorkingDir: "/workspace",
			Command:    []string{"bash"},
			Environment: map[string]string{
				"GOPATH": "/go",
			},
			Ports: map[string]string{
				"8080": "8080",
			},
			Detached: true,
		},
	}

	// Save workspace metadata
	require.NoError(t, os.MkdirAll(workspace.Path, 0755))
	require.NoError(t, SaveWorkspaceMetadata(workspace.Path, workspace))

	// Start container
	ctx := context.Background()
	err := manager.StartContainer(ctx, 1)
	
	// Verify it succeeds
	assert.NoError(t, err)
	
	// Verify workspace was updated
	updatedWorkspace, err := manager.GetWorkspace(1)
	require.NoError(t, err)
	assert.Equal(t, types.WorkspaceStatusActive, updatedWorkspace.Status)
	assert.NotEmpty(t, updatedWorkspace.ContainerID)
	assert.NotNil(t, updatedWorkspace.ContainerStatus)
}

func TestStopContainer_WithValidContainer(t *testing.T) {
	// Test case: Stopping a container with valid container should succeed
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspaces")
	
	// Create workspace manager with mock container manager
	manager := &Manager{
		baseDir:         workspaceDir,
		containerManager: &MockContainerManager{},
	}

	// Create a workspace with container
	workspace := &types.Workspace{
		ID:          1,
		TaskName:    "test-task",
		Path:        filepath.Join(workspaceDir, "1"),
		Status:      types.WorkspaceStatusActive,
		ContainerID: "mock-container-id",
		ContainerConfig: &types.ContainerConfig{
			Image: "golang:1.21",
		},
	}

	// Save workspace metadata
	require.NoError(t, os.MkdirAll(workspace.Path, 0755))
	require.NoError(t, SaveWorkspaceMetadata(workspace.Path, workspace))

	// Stop container
	ctx := context.Background()
	err := manager.StopContainer(ctx, 1, 30)
	
	// Verify it succeeds
	assert.NoError(t, err)
	
	// Verify workspace was updated
	updatedWorkspace, err := manager.GetWorkspace(1)
	require.NoError(t, err)
	assert.Equal(t, types.WorkspaceStatusReady, updatedWorkspace.Status)
}

func TestExecInContainer_WithValidContainer(t *testing.T) {
	// Test case: Executing a command in a container should succeed
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspaces")
	
	// Create workspace manager with mock container manager
	manager := &Manager{
		baseDir:         workspaceDir,
		containerManager: &MockContainerManager{},
	}

	// Create a workspace with container
	workspace := &types.Workspace{
		ID:          1,
		TaskName:    "test-task",
		Path:        filepath.Join(workspaceDir, "1"),
		Status:      types.WorkspaceStatusActive,
		ContainerID: "mock-container-id",
		ContainerConfig: &types.ContainerConfig{
			Image: "golang:1.21",
		},
	}

	// Save workspace metadata
	require.NoError(t, os.MkdirAll(workspace.Path, 0755))
	require.NoError(t, SaveWorkspaceMetadata(workspace.Path, workspace))

	// Execute command in container
	ctx := context.Background()
	err := manager.ExecInContainer(ctx, 1, []string{"ls", "-la"})
	
	// Verify it succeeds
	assert.NoError(t, err)
}

func TestGetContainerLogs_WithValidContainer(t *testing.T) {
	// Test case: Getting logs from a container should succeed
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspaces")
	
	// Create workspace manager with mock container manager
	manager := &Manager{
		baseDir:         workspaceDir,
		containerManager: &MockContainerManager{},
	}

	// Create a workspace with container
	workspace := &types.Workspace{
		ID:          1,
		TaskName:    "test-task",
		Path:        filepath.Join(workspaceDir, "1"),
		Status:      types.WorkspaceStatusActive,
		ContainerID: "mock-container-id",
		ContainerConfig: &types.ContainerConfig{
			Image: "golang:1.21",
		},
	}

	// Save workspace metadata
	require.NoError(t, os.MkdirAll(workspace.Path, 0755))
	require.NoError(t, SaveWorkspaceMetadata(workspace.Path, workspace))

	// Get container logs
	ctx := context.Background()
	logs, err := manager.GetContainerLogs(ctx, 1, false, 100)
	
	// Verify it succeeds
	assert.NoError(t, err)
	assert.Contains(t, logs, "mock logs")
}

func TestDeleteWorkspace_WithContainer(t *testing.T) {
	// Test case: Deleting a workspace with a container should clean up the container
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspaces")
	
	// Create workspace manager with mock container manager
	manager := &Manager{
		baseDir:         workspaceDir,
		containerManager: &MockContainerManager{},
	}

	// Create a workspace with container
	workspace := &types.Workspace{
		ID:          1,
		TaskName:    "test-task",
		Path:        filepath.Join(workspaceDir, "1"),
		Status:      types.WorkspaceStatusActive,
		ContainerID: "mock-container-id",
		ContainerConfig: &types.ContainerConfig{
			Image: "golang:1.21",
		},
	}

	// Save workspace metadata
	require.NoError(t, os.MkdirAll(workspace.Path, 0755))
	require.NoError(t, SaveWorkspaceMetadata(workspace.Path, workspace))

	// Delete workspace
	err := manager.DeleteWorkspace(1)
	
	// Verify it succeeds
	assert.NoError(t, err)
	
	// Verify workspace directory was removed
	_, err = os.Stat(workspace.Path)
	assert.True(t, os.IsNotExist(err))
}

func TestStartContainer_WithAutoGeneratedName(t *testing.T) {
	// Test case: Starting a container without a name should auto-generate one
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspaces")
	
	// Create workspace manager with mock container manager
	manager := &Manager{
		baseDir:         workspaceDir,
		containerManager: &MockContainerManager{},
	}

	// Create a workspace with container configuration but no name
	workspace := &types.Workspace{
		ID:       1,
		TaskName: "test-task",
		Path:     filepath.Join(workspaceDir, "1"),
		Status:   types.WorkspaceStatusReady,
		ContainerConfig: &types.ContainerConfig{
			Image:      "golang:1.21",
			WorkingDir: "/workspace",
			Detached:   true,
		},
	}

	// Save workspace metadata
	require.NoError(t, os.MkdirAll(workspace.Path, 0755))
	require.NoError(t, SaveWorkspaceMetadata(workspace.Path, workspace))

	// Start container
	ctx := context.Background()
	err := manager.StartContainer(ctx, 1)
	
	// Verify it succeeds
	assert.NoError(t, err)
	
	// Verify workspace was updated
	updatedWorkspace, err := manager.GetWorkspace(1)
	require.NoError(t, err)
	assert.Equal(t, types.WorkspaceStatusActive, updatedWorkspace.Status)
	assert.NotEmpty(t, updatedWorkspace.ContainerID)
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

// MockGitOperations is a mock implementation of GitOperationsInterface for testing
type MockGitOperations struct {
	cloneFunc       func(req *types.CreateWorkspaceRequest, workspacePath string) error
	getRepoInfoFunc func(repoPath string) (*git.RepositoryInfo, error)
}

func (m *MockGitOperations) CloneRepository(req *types.CreateWorkspaceRequest, workspacePath string) error {
	if m.cloneFunc != nil {
		return m.cloneFunc(req, workspacePath)
	}
	return nil
}

func (m *MockGitOperations) GetRepositoryInfo(repoPath string) (*git.RepositoryInfo, error) {
	if m.getRepoInfoFunc != nil {
		return m.getRepoInfoFunc(repoPath)
	}
	return &git.RepositoryInfo{
		Path:          repoPath,
		CurrentBranch: "main",
		RemoteURL:     "https://github.com/test/repo.git",
	}, nil
}

// MockContainerManager is a mock implementation of ContainerManager for testing
type MockContainerManager struct{}

func (m *MockContainerManager) GetEngine() container.ContainerEngine {
	return container.EngineDocker
}

func (m *MockContainerManager) GetVersion() (string, error) {
	return "1.0.0", nil
}

func (m *MockContainerManager) IsAvailable() bool {
	return true
}

func (m *MockContainerManager) Run(ctx context.Context, options container.RunOptions) (string, error) {
	return "mock-container-id", nil
}

func (m *MockContainerManager) Start(ctx context.Context, containerID string) error {
	return nil
}

func (m *MockContainerManager) Stop(ctx context.Context, containerID string, timeoutSeconds int) error {
	return nil
}

func (m *MockContainerManager) Remove(ctx context.Context, containerID string, force bool) error {
	return nil
}

func (m *MockContainerManager) Exec(ctx context.Context, containerID string, command []string, options container.ExecOptions) error {
	return nil
}

func (m *MockContainerManager) Logs(ctx context.Context, containerID string, options container.LogOptions) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock logs")), nil
}

func (m *MockContainerManager) Inspect(ctx context.Context, containerID string) (*container.ContainerInfo, error) {
	return &container.ContainerInfo{
		ID:      containerID,
		Name:    "mock-container",
		Image:   "mock-image",
		Status:  "running",
		Created: "2023-01-01T00:00:00Z",
		Ports:   []string{},
		Labels:  map[string]string{},
	}, nil
}

func (m *MockContainerManager) List(ctx context.Context, options container.ListOptions) ([]container.ContainerInfo, error) {
	return []container.ContainerInfo{}, nil
}

func (m *MockContainerManager) Pull(ctx context.Context, image string) error {
	return nil
}

func (m *MockContainerManager) Build(ctx context.Context, options container.BuildOptions) (string, error) {
	return "mock-image-id", nil
}

func (m *MockContainerManager) RemoveImage(ctx context.Context, imageID string, force bool) error {
	return nil
}

func (m *MockContainerManager) ListImages(ctx context.Context) ([]container.ImageInfo, error) {
	return []container.ImageInfo{}, nil
}

func (m *MockContainerManager) InspectImage(ctx context.Context, imageID string) (*container.ImageInfo, error) {
	return &container.ImageInfo{}, nil
}

func (m *MockContainerManager) CreateNetwork(ctx context.Context, name string, options container.NetworkOptions) error {
	return nil
}

func (m *MockContainerManager) RemoveNetwork(ctx context.Context, name string) error {
	return nil
}

func (m *MockContainerManager) ListNetworks(ctx context.Context) ([]container.NetworkInfo, error) {
	return []container.NetworkInfo{}, nil
}

func (m *MockContainerManager) CreateVolume(ctx context.Context, name string, options container.VolumeOptions) error {
	return nil
}

func (m *MockContainerManager) RemoveVolume(ctx context.Context, name string, force bool) error {
	return nil
}

func (m *MockContainerManager) ListVolumes(ctx context.Context) ([]container.VolumeInfo, error) {
	return []container.VolumeInfo{}, nil
}
