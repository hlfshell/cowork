package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hlfshell/cowork/internal/types"
)

// TestSaveWorkspaceMetadata_WithValidWorkspace tests saving workspace metadata with valid workspace
func TestSaveWorkspaceMetadata_WithValidWorkspace(t *testing.T) {
	// Test case: Saving workspace metadata with valid workspace should succeed
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "test-workspace")

	// Create workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		t.Fatalf("Failed to create workspace directory: %v", err)
	}

	workspace := &types.Workspace{
		ID:           1,
		TaskName:     "test-task",
		Description:  "Test task description",
		TicketID:     "123",
		Path:         workspacePath,
		BranchName:   "task/test-task",
		SourceRepo:   "/path/to/repo",
		BaseBranch:   "main",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Status:       types.WorkspaceStatusReady,
		ContainerID:  "container-123",
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err := SaveWorkspaceMetadata(workspacePath, workspace)
	if err != nil {
		t.Fatalf("Failed to save workspace metadata: %v", err)
	}

	// Verify metadata file was created
	metadataPath := filepath.Join(workspacePath, MetadataFileName)
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Errorf("Metadata file was not created: %s", metadataPath)
	}

	// Verify file content contains expected data
	content, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, fmt.Sprintf("%d", workspace.ID)) {
		t.Errorf("Metadata file should contain workspace ID: %d", workspace.ID)
	}

	if !strings.Contains(contentStr, workspace.TaskName) {
		t.Errorf("Metadata file should contain task name: %s", workspace.TaskName)
	}

	if !strings.Contains(contentStr, workspace.Description) {
		t.Errorf("Metadata file should contain description: %s", workspace.Description)
	}

	if !strings.Contains(contentStr, workspace.TicketID) {
		t.Errorf("Metadata file should contain ticket ID: %s", workspace.TicketID)
	}
}

// TestSaveWorkspaceMetadata_WithNonExistentDirectory tests saving metadata to non-existent directory
func TestSaveWorkspaceMetadata_WithNonExistentDirectory(t *testing.T) {
	// Test case: Saving metadata to non-existent directory should fail
	nonExistentPath := "/path/that/does/not/exist"

	workspace := &types.Workspace{
		ID:       2,
		TaskName: "test-task",
	}

	err := SaveWorkspaceMetadata(nonExistentPath, workspace)
	if err == nil {
		t.Error("Expected error when saving to non-existent directory, got nil")
	}

	expectedError := "no such file or directory"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestLoadWorkspaceMetadata_WithValidFile tests loading workspace metadata from valid file
func TestLoadWorkspaceMetadata_WithValidFile(t *testing.T) {
	// Test case: Loading workspace metadata from valid file should succeed
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "test-workspace")

	// Create workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		t.Fatalf("Failed to create workspace directory: %v", err)
	}

	// Create a test workspace
	originalWorkspace := &types.Workspace{
		ID:           3,
		TaskName:     "test-task",
		Description:  "Test task description",
		TicketID:     "123",
		Path:         workspacePath,
		BranchName:   "task/test-task",
		SourceRepo:   "/path/to/repo",
		BaseBranch:   "main",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Status:       types.WorkspaceStatusReady,
		ContainerID:  "container-123",
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// Save the metadata
	err := SaveWorkspaceMetadata(workspacePath, originalWorkspace)
	if err != nil {
		t.Fatalf("Failed to save workspace metadata: %v", err)
	}

	// Load the metadata
	loadedWorkspace, err := LoadWorkspaceMetadata(workspacePath)
	if err != nil {
		t.Fatalf("Failed to load workspace metadata: %v", err)
	}

	if loadedWorkspace == nil {
		t.Fatal("Expected loaded workspace to not be nil")
	}

	// Verify all fields were loaded correctly
	if loadedWorkspace.ID != originalWorkspace.ID {
		t.Errorf("Expected ID %d, got %d", originalWorkspace.ID, loadedWorkspace.ID)
	}

	if loadedWorkspace.TaskName != originalWorkspace.TaskName {
		t.Errorf("Expected TaskName '%s', got '%s'", originalWorkspace.TaskName, loadedWorkspace.TaskName)
	}

	if loadedWorkspace.Description != originalWorkspace.Description {
		t.Errorf("Expected Description '%s', got '%s'", originalWorkspace.Description, loadedWorkspace.Description)
	}

	if loadedWorkspace.TicketID != originalWorkspace.TicketID {
		t.Errorf("Expected TicketID '%s', got '%s'", originalWorkspace.TicketID, loadedWorkspace.TicketID)
	}

	if loadedWorkspace.BranchName != originalWorkspace.BranchName {
		t.Errorf("Expected BranchName '%s', got '%s'", originalWorkspace.BranchName, loadedWorkspace.BranchName)
	}

	if loadedWorkspace.SourceRepo != originalWorkspace.SourceRepo {
		t.Errorf("Expected SourceRepo '%s', got '%s'", originalWorkspace.SourceRepo, loadedWorkspace.SourceRepo)
	}

	if loadedWorkspace.BaseBranch != originalWorkspace.BaseBranch {
		t.Errorf("Expected BaseBranch '%s', got '%s'", originalWorkspace.BaseBranch, loadedWorkspace.BaseBranch)
	}

	if loadedWorkspace.Status != originalWorkspace.Status {
		t.Errorf("Expected Status '%s', got '%s'", originalWorkspace.Status, loadedWorkspace.Status)
	}

	if loadedWorkspace.ContainerID != originalWorkspace.ContainerID {
		t.Errorf("Expected ContainerID '%s', got '%s'", originalWorkspace.ContainerID, loadedWorkspace.ContainerID)
	}

	// Verify metadata map
	if len(loadedWorkspace.Metadata) != len(originalWorkspace.Metadata) {
		t.Errorf("Expected %d metadata entries, got %d", len(originalWorkspace.Metadata), len(loadedWorkspace.Metadata))
	}

	for key, value := range originalWorkspace.Metadata {
		if loadedWorkspace.Metadata[key] != value {
			t.Errorf("Expected metadata key '%s' to have value '%s', got '%s'", key, value, loadedWorkspace.Metadata[key])
		}
	}

	// Verify path is set correctly
	if loadedWorkspace.Path != workspacePath {
		t.Errorf("Expected Path '%s', got '%s'", workspacePath, loadedWorkspace.Path)
	}
}

// TestLoadWorkspaceMetadata_WithNonExistentFile tests loading metadata from non-existent file
func TestLoadWorkspaceMetadata_WithNonExistentFile(t *testing.T) {
	// Test case: Loading metadata from non-existent file should fail
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "non-existent-workspace")

	workspace, err := LoadWorkspaceMetadata(workspacePath)
	if err == nil {
		t.Error("Expected error when loading from non-existent file, got nil")
	}

	if workspace != nil {
		t.Error("Expected workspace to be nil for non-existent file")
	}

	expectedError := "no such file or directory"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestLoadWorkspaceMetadata_WithInvalidJSON tests loading metadata from invalid JSON file
func TestLoadWorkspaceMetadata_WithInvalidJSON(t *testing.T) {
	// Test case: Loading metadata from invalid JSON file should fail
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "test-workspace")

	// Create workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		t.Fatalf("Failed to create workspace directory: %v", err)
	}

	// Create invalid JSON file
	metadataPath := filepath.Join(workspacePath, MetadataFileName)
	invalidJSON := `{"invalid": json, "missing": quotes}`
	if err := os.WriteFile(metadataPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	workspace, err := LoadWorkspaceMetadata(workspacePath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON, got nil")
	}

	if workspace != nil {
		t.Error("Expected workspace to be nil for invalid JSON")
	}

	expectedError := "invalid character"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestDiscoverWorkspaces_WithNoWorkspaces tests discovering workspaces when none exist
func TestDiscoverWorkspaces_WithNoWorkspaces(t *testing.T) {
	// Test case: Discovering workspaces when none exist should return empty list
	tempDir := t.TempDir()

	workspaces, err := DiscoverWorkspaces(tempDir)
	if err != nil {
		t.Fatalf("Failed to discover workspaces: %v", err)
	}

	if len(workspaces) != 0 {
		t.Errorf("Expected 0 workspaces, got %d", len(workspaces))
	}
}

// TestDiscoverWorkspaces_WithMultipleWorkspaces tests discovering multiple workspaces
func TestDiscoverWorkspaces_WithMultipleWorkspaces(t *testing.T) {
	// Test case: Discovering workspaces should return all valid workspaces
	tempDir := t.TempDir()

	// Create multiple workspace directories with valid metadata
	workspaceIDs := []string{"ws1", "ws2", "ws3"}
	expectedWorkspaces := make(map[string]*types.Workspace)

	for i, workspaceID := range workspaceIDs {
		workspacePath := filepath.Join(tempDir, workspaceID)

		// Create workspace directory
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			t.Fatalf("Failed to create workspace directory %s: %v", workspaceID, err)
		}

		// Create workspace with metadata
		workspace := &types.Workspace{
			ID:           i + 1,
			TaskName:     fmt.Sprintf("task-%d", i+1),
			Description:  fmt.Sprintf("Description for task %d", i+1),
			Path:         workspacePath,
			BranchName:   fmt.Sprintf("task/task-%d", i+1),
			SourceRepo:   "/path/to/repo",
			BaseBranch:   "main",
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
			Status:       types.WorkspaceStatusReady,
		}

		// Save metadata
		err := SaveWorkspaceMetadata(workspacePath, workspace)
		if err != nil {
			t.Fatalf("Failed to save workspace metadata for %s: %v", workspaceID, err)
		}

		expectedWorkspaces[workspaceID] = workspace
	}

	// Discover workspaces
	discoveredWorkspaces, err := DiscoverWorkspaces(tempDir)
	if err != nil {
		t.Fatalf("Failed to discover workspaces: %v", err)
	}

	if len(discoveredWorkspaces) != len(workspaceIDs) {
		t.Errorf("Expected %d workspaces, got %d", len(workspaceIDs), len(discoveredWorkspaces))
	}

	// Verify all expected workspaces were discovered
	foundWorkspaces := make(map[int]bool)
	for _, workspace := range discoveredWorkspaces {
		foundWorkspaces[workspace.ID] = true
	}

	for i, workspaceID := range workspaceIDs {
		if !foundWorkspaces[i+1] {
			t.Errorf("Expected workspace '%s' not found in discovered workspaces", workspaceID)
		}
	}
}

// TestDiscoverWorkspaces_WithInvalidMetadata tests discovering workspaces with invalid metadata
func TestDiscoverWorkspaces_WithInvalidMetadata(t *testing.T) {
	// Test case: Discovering workspaces should skip directories with invalid metadata
	tempDir := t.TempDir()

	// Create a workspace directory with invalid metadata
	workspacePath := filepath.Join(tempDir, "invalid-workspace")
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		t.Fatalf("Failed to create workspace directory: %v", err)
	}

	// Create invalid metadata file
	metadataPath := filepath.Join(workspacePath, MetadataFileName)
	invalidJSON := `{"invalid": json}`
	if err := os.WriteFile(metadataPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	// Create a valid workspace for comparison
	validWorkspacePath := filepath.Join(tempDir, "valid-workspace")
	if err := os.MkdirAll(validWorkspacePath, 0755); err != nil {
		t.Fatalf("Failed to create valid workspace directory: %v", err)
	}

	validWorkspace := &types.Workspace{
		ID:       999,
		TaskName: "valid-task",
		Path:     validWorkspacePath,
		Status:   types.WorkspaceStatusReady,
	}

	err := SaveWorkspaceMetadata(validWorkspacePath, validWorkspace)
	if err != nil {
		t.Fatalf("Failed to save valid workspace metadata: %v", err)
	}

	// Discover workspaces
	discoveredWorkspaces, err := DiscoverWorkspaces(tempDir)
	if err != nil {
		t.Fatalf("Failed to discover workspaces: %v", err)
	}

	// Should only find the valid workspace
	if len(discoveredWorkspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(discoveredWorkspaces))
	}

	if discoveredWorkspaces[0].ID != 999 {
		t.Errorf("Expected workspace ID 999, got %d", discoveredWorkspaces[0].ID)
	}
}

// TestUpdateWorkspaceMetadata_WithValidWorkspace tests updating workspace metadata
func TestUpdateWorkspaceMetadata_WithValidWorkspace(t *testing.T) {
	// Test case: Updating workspace metadata should succeed
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "test-workspace")

	// Create workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		t.Fatalf("Failed to create workspace directory: %v", err)
	}

	// Create initial workspace
	workspace := &types.Workspace{
		ID:       1,
		TaskName: "test-task",
		Path:     workspacePath,
		Status:   types.WorkspaceStatusReady,
	}

	// Save initial metadata
	err := SaveWorkspaceMetadata(workspacePath, workspace)
	if err != nil {
		t.Fatalf("Failed to save initial workspace metadata: %v", err)
	}

	// Update workspace
	workspace.Status = types.WorkspaceStatusActive
	workspace.Description = "Updated description"

	err = UpdateWorkspaceMetadata(workspace)
	if err != nil {
		t.Fatalf("Failed to update workspace metadata: %v", err)
	}

	// Load and verify the update
	updatedWorkspace, err := LoadWorkspaceMetadata(workspacePath)
	if err != nil {
		t.Fatalf("Failed to load updated workspace metadata: %v", err)
	}

	if updatedWorkspace.Status != types.WorkspaceStatusActive {
		t.Errorf("Expected status '%s', got '%s'", types.WorkspaceStatusActive, updatedWorkspace.Status)
	}

	if updatedWorkspace.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", updatedWorkspace.Description)
	}
}

// TestTruncateDescription_WithShortDescription tests truncating short descriptions
func TestTruncateDescription_WithShortDescription(t *testing.T) {
	// Test case: Truncating short descriptions should return the original description
	description := "Short description"
	maxLength := 50

	result := TruncateDescription(description, maxLength)
	if result != description {
		t.Errorf("Expected original description '%s', got '%s'", description, result)
	}
}

// TestTruncateDescription_WithLongDescription tests truncating long descriptions
func TestTruncateDescription_WithLongDescription(t *testing.T) {
	// Test case: Truncating long descriptions should truncate at word boundary
	description := "This is a very long description that should be truncated at a word boundary when it exceeds the maximum length"
	maxLength := 30

	result := TruncateDescription(description, maxLength)

	if len(result) > maxLength+3 { // +3 for "..."
		t.Errorf("Expected truncated description to be <= %d characters, got %d: '%s'", maxLength+3, len(result), result)
	}

	if !strings.HasSuffix(result, "...") {
		t.Errorf("Expected truncated description to end with '...', got '%s'", result)
	}

	// Should truncate at a word boundary
	if strings.Contains(result, "boundary") {
		t.Errorf("Expected description to be truncated before 'boundary', got '%s'", result)
	}
}

// TestTruncateDescription_WithVeryLongWord tests truncating descriptions with very long words
func TestTruncateDescription_WithVeryLongWord(t *testing.T) {
	// Test case: Truncating descriptions with very long words should truncate mid-word
	description := "This is a description with a verylongwordthatdoesnothaveanyspacesinit"
	maxLength := 20

	result := TruncateDescription(description, maxLength)

	if len(result) > maxLength+3 { // +3 for "..."
		t.Errorf("Expected truncated description to be <= %d characters, got %d: '%s'", maxLength+3, len(result), result)
	}

	if !strings.HasSuffix(result, "...") {
		t.Errorf("Expected truncated description to end with '...', got '%s'", result)
	}
}

// TestReadDescriptionFromFile_WithValidFile tests reading description from valid file
func TestReadDescriptionFromFile_WithValidFile(t *testing.T) {
	// Test case: Reading description from valid file should succeed
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "description.txt")

	expectedDescription := "This is a test description with multiple lines.\nIt should be read correctly from the file."

	err := os.WriteFile(filePath, []byte(expectedDescription), 0644)
	if err != nil {
		t.Fatalf("Failed to write description file: %v", err)
	}

	description, err := ReadDescriptionFromFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read description from file: %v", err)
	}

	// Should trim whitespace
	expectedTrimmed := strings.TrimSpace(expectedDescription)
	if description != expectedTrimmed {
		t.Errorf("Expected description '%s', got '%s'", expectedTrimmed, description)
	}
}

// TestReadDescriptionFromFile_WithNonExistentFile tests reading description from non-existent file
func TestReadDescriptionFromFile_WithNonExistentFile(t *testing.T) {
	// Test case: Reading description from non-existent file should fail
	nonExistentPath := "/path/that/does/not/exist/description.txt"

	description, err := ReadDescriptionFromFile(nonExistentPath)
	if err == nil {
		t.Error("Expected error when reading from non-existent file, got nil")
	}

	if description != "" {
		t.Error("Expected description to be empty for non-existent file")
	}

	expectedError := "no such file or directory"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestReadDescriptionFromFile_WithEmptyFile tests reading description from empty file
func TestReadDescriptionFromFile_WithEmptyFile(t *testing.T) {
	// Test case: Reading description from empty file should return empty string
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.txt")

	// Create empty file
	err := os.WriteFile(filePath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	description, err := ReadDescriptionFromFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read empty file: %v", err)
	}

	if description != "" {
		t.Errorf("Expected empty description, got '%s'", description)
	}
}
