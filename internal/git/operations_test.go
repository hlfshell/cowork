package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hlfshell/cowork/internal/types"
)

// TestNewGitOperations_WithValidTimeout tests creating GitOperations with valid timeout
func TestNewGitOperations_WithValidTimeout(t *testing.T) {
	// Test case: Creating GitOperations with a valid timeout should succeed
	timeoutSeconds := 600
	gitOps := NewGitOperations(timeoutSeconds)

	if gitOps == nil {
		t.Error("Expected GitOperations to be created, got nil")
	}
}

// TestNewGitOperations_WithZeroTimeout tests creating GitOperations with zero timeout
func TestNewGitOperations_WithZeroTimeout(t *testing.T) {
	// Test case: Creating GitOperations with zero timeout should use default
	gitOps := NewGitOperations(0)

	if gitOps == nil {
		t.Error("Expected GitOperations to be created, got nil")
	}
}

// TestNewGitOperations_WithNegativeTimeout tests creating GitOperations with negative timeout
func TestNewGitOperations_WithNegativeTimeout(t *testing.T) {
	// Test case: Creating GitOperations with negative timeout should use default
	gitOps := NewGitOperations(-100)

	if gitOps == nil {
		t.Error("Expected GitOperations to be created, got nil")
	}
}

// TestGitOperations_sanitizeBranchName_WithValidNames tests branch name sanitization with valid names
func TestGitOperations_sanitizeBranchName_WithValidNames(t *testing.T) {
	// Test case: Valid branch names should be sanitized correctly
	gitOps := NewGitOperations(300)

	testCases := []struct {
		input    string
		expected string
	}{
		{"simple-task", "simple-task"},
		{"task with spaces", "task-with-spaces"},
		{"task/with/slashes", "task-with-slashes"},
		{"task:with:colons", "task-with-colons"},
		{"task*with*stars", "task-with-stars"},
		{"task?with?question", "task-with-question"},
		{"task\"with\"quotes", "task-with-quotes"},
		{"task<with>brackets", "task-with-brackets"},
		{"task|with|pipes", "task-with-pipes"},
		{"task..with..dots", "task-with-dots"},
		{"task-with-leading-hyphen", "task-with-leading-hyphen"},
		{"task-with-trailing-hyphen-", "task-with-trailing-hyphen"},
		{"task.with.leading.dots", "task.with.leading.dots"},
		{"task.with.trailing.dots.", "task.with.trailing.dots"},
		{"", "task"},
		{"a", "a"},
	}

	for _, tc := range testCases {
		result := gitOps.sanitizeBranchName(tc.input)
		if result != tc.expected {
			t.Errorf("Expected sanitized name '%s' for input '%s', got '%s'", tc.expected, tc.input, result)
		}
	}
}

// TestGitOperations_sanitizeBranchName_WithLongNames tests branch name sanitization with long names
func TestGitOperations_sanitizeBranchName_WithLongNames(t *testing.T) {
	// Test case: Long branch names should be truncated to 50 characters
	gitOps := NewGitOperations(300)

	// Create a long name that should be truncated
	longName := strings.Repeat("very-long-task-name-", 10) // This will be longer than 50 chars

	result := gitOps.sanitizeBranchName(longName)

	if len(result) > 50 {
		t.Errorf("Expected sanitized name to be <= 50 characters, got %d: '%s'", len(result), result)
	}

	// Should not end with a hyphen
	if strings.HasSuffix(result, "-") {
		t.Errorf("Expected sanitized name to not end with hyphen, got '%s'", result)
	}
}

// TestGitOperations_generateBranchName_WithTaskNameOnly tests branch name generation with task name only
func TestGitOperations_generateBranchName_WithTaskNameOnly(t *testing.T) {
	// Test case: Generating branch name with only task name should create correct format
	gitOps := NewGitOperations(300)

	taskName := "oauth-refresh"
	expectedPrefix := "task/oauth-refresh"

	result := gitOps.generateBranchName(taskName, "")

	if !strings.HasPrefix(result, expectedPrefix) {
		t.Errorf("Expected branch name to start with '%s', got '%s'", expectedPrefix, result)
	}

	if result != expectedPrefix {
		t.Errorf("Expected branch name '%s', got '%s'", expectedPrefix, result)
	}
}

// TestGitOperations_generateBranchName_WithTaskNameAndTicket tests branch name generation with task name and ticket
func TestGitOperations_generateBranchName_WithTaskNameAndTicket(t *testing.T) {
	// Test case: Generating branch name with task name and ticket should create correct format
	gitOps := NewGitOperations(300)

	taskName := "oauth-refresh"
	ticketID := "123"
	expected := "task/oauth-refresh-123"

	result := gitOps.generateBranchName(taskName, ticketID)

	if result != expected {
		t.Errorf("Expected branch name '%s', got '%s'", expected, result)
	}
}

// TestGitOperations_generateBranchName_WithSpecialCharacters tests branch name generation with special characters
func TestGitOperations_generateBranchName_WithSpecialCharacters(t *testing.T) {
	// Test case: Generating branch name with special characters should sanitize them
	gitOps := NewGitOperations(300)

	taskName := "fix/bug #123 with spaces"
	ticketID := "456"
	expected := "task/fix-bug--123-with-spaces-456"

	result := gitOps.generateBranchName(taskName, ticketID)

	if result != expected {
		t.Errorf("Expected branch name '%s', got '%s'", expected, result)
	}
}

// TestGitOperations_CloneRepository_WithInvalidRequest tests cloning with invalid request
func TestGitOperations_CloneRepository_WithInvalidRequest(t *testing.T) {
	// Test case: Cloning with invalid request should return error
	gitOps := NewGitOperations(300)

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workspacePath := filepath.Join(tempDir, "workspace")

	// Test with empty task name
	req := &types.CreateWorkspaceRequest{
		TaskName:   "",
		SourceRepo: "https://github.com/test/repo.git",
	}

	err = gitOps.CloneRepository(req, workspacePath)
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
	}

	expectedError := "task name is required"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestGitOperations_CloneRepository_WithExistingDirectory tests cloning with existing directory
func TestGitOperations_CloneRepository_WithExistingDirectory(t *testing.T) {
	// Test case: Cloning to existing directory should return error
	gitOps := NewGitOperations(300)

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workspacePath := filepath.Join(tempDir, "workspace")

	// Create the workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		t.Fatalf("Failed to create workspace directory: %v", err)
	}

	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: "https://github.com/test/repo.git",
	}

	err = gitOps.CloneRepository(req, workspacePath)
	if err == nil {
		t.Error("Expected error for existing directory, got nil")
	}

	expectedError := "workspace directory already exists"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}



// TestGitOperations_CloneRepository_WithValidRequest tests cloning with valid request
func TestGitOperations_CloneRepository_WithValidRequest(t *testing.T) {
	// Test case: Cloning with valid request should succeed (this will fail in test environment
	// since we don't have a real Git repository, but we can test the validation)
	gitOps := NewGitOperations(300)

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "cowork-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	workspacePath := filepath.Join(tempDir, "workspace")

	req := &types.CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: "https://github.com/test/repo.git",
	}

	// This will fail because the repository doesn't exist, but we can verify
	// that the validation passes and the directory is created
	err = gitOps.CloneRepository(req, workspacePath)

	// We expect this to fail because the repository doesn't exist, but the error
	// should be about the Git operation, not validation
	if err != nil {
		if strings.Contains(err.Error(), "task name is required") ||
			strings.Contains(err.Error(), "source repository URL is required") ||
			strings.Contains(err.Error(), "invalid isolation level") {
			t.Errorf("Expected Git operation error, got validation error: %v", err)
		}
		// The error should be about the Git operation failing, which is expected
		// since we don't have a real repository
	} else {
		t.Error("Expected error for non-existent repository, got nil")
	}
}

// TestGitOperations_GetRepositoryInfo_WithNonExistentPath tests getting repository info for non-existent path
func TestGitOperations_GetRepositoryInfo_WithNonExistentPath(t *testing.T) {
	// Test case: Getting repository info for non-existent path should return error
	gitOps := NewGitOperations(300)

	nonExistentPath := "/path/that/does/not/exist"

	_, err := gitOps.GetRepositoryInfo(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}

	expectedError := "failed to get current branch"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}
