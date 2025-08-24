package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectCurrentRepository_WithValidRepository tests detecting current repository in a valid Git repo
func TestDetectCurrentRepository_WithValidRepository(t *testing.T) {
	// Test case: Detecting current repository in a valid Git repository should succeed
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

	repoInfo, err := DetectCurrentRepository()
	if err != nil {
		t.Fatalf("Failed to detect current repository: %v", err)
	}

	if repoInfo == nil {
		t.Fatal("Expected repository info to not be nil")
	}

	// Verify repository info
	if repoInfo.Path != tempDir {
		t.Errorf("Expected repository path '%s', got '%s'", tempDir, repoInfo.Path)
	}

	if repoInfo.CurrentBranch == "" {
		t.Error("Expected current branch to not be empty")
	}

	// The remote URL might be empty for a local repository
	if repoInfo.RemoteURL != "" {
		t.Logf("Remote URL: %s", repoInfo.RemoteURL)
	}
}

// TestDetectCurrentRepository_OutsideRepository tests detecting current repository outside a Git repo
func TestDetectCurrentRepository_OutsideRepository(t *testing.T) {
	// Test case: Detecting current repository outside a Git repository should fail
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

	repoInfo, err := DetectCurrentRepository()
	if err == nil {
		t.Error("Expected error when detecting repository outside Git repo, got nil")
	}

	if repoInfo != nil {
		t.Error("Expected repository info to be nil outside Git repo")
	}

	expectedError := "current directory is not a Git repository"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestIsGitRepository_WithValidRepository tests checking if a directory is a Git repository
func TestIsGitRepository_WithValidRepository(t *testing.T) {
	// Test case: Checking if a valid Git repository directory should return true
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	isRepo := isGitRepository(tempDir)
	if !isRepo {
		t.Error("Expected valid Git repository to return true")
	}
}

// TestIsGitRepository_WithNonExistentDirectory tests checking if a non-existent directory is a Git repository
func TestIsGitRepository_WithNonExistentDirectory(t *testing.T) {
	// Test case: Checking if a non-existent directory should return false
	nonExistentPath := "/path/that/does/not/exist"

	isRepo := isGitRepository(nonExistentPath)
	if isRepo {
		t.Error("Expected non-existent directory to return false")
	}
}

// TestIsGitRepository_WithNonGitDirectory tests checking if a non-Git directory is a Git repository
func TestIsGitRepository_WithNonGitDirectory(t *testing.T) {
	// Test case: Checking if a non-Git directory should return false
	tempDir := t.TempDir()

	isRepo := isGitRepository(tempDir)
	if isRepo {
		t.Error("Expected non-Git directory to return false")
	}
}

// TestGetRepositoryInfo_WithValidRepository tests getting repository info from a valid Git repo
func TestGetRepositoryInfo_WithValidRepository(t *testing.T) {
	// Test case: Getting repository info from a valid Git repository should succeed
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	repoInfo, err := GetRepositoryInfo(tempDir)
	if err != nil {
		t.Fatalf("Failed to get repository info: %v", err)
	}

	if repoInfo == nil {
		t.Fatal("Expected repository info to not be nil")
	}

	// Verify repository info
	if repoInfo.Path != tempDir {
		t.Errorf("Expected repository path '%s', got '%s'", tempDir, repoInfo.Path)
	}

	if repoInfo.CurrentBranch == "" {
		t.Error("Expected current branch to not be empty")
	}

	// The remote URL might be empty for a local repository
	if repoInfo.RemoteURL != "" {
		t.Logf("Remote URL: %s", repoInfo.RemoteURL)
	}
}

// TestGetRepositoryInfo_WithNonExistentPath tests getting repository info from non-existent path
func TestGetRepositoryInfo_WithNonExistentPath(t *testing.T) {
	// Test case: Getting repository info from non-existent path should fail
	nonExistentPath := "/path/that/does/not/exist"

	repoInfo, err := GetRepositoryInfo(nonExistentPath)
	if err == nil {
		t.Error("Expected error when getting repository info from non-existent path, got nil")
	}

	if repoInfo != nil {
		t.Error("Expected repository info to be nil for non-existent path")
	}

	expectedError := "failed to get current branch"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestGetRepositoryInfo_WithNonGitDirectory tests getting repository info from non-Git directory
func TestGetRepositoryInfo_WithNonGitDirectory(t *testing.T) {
	// Test case: Getting repository info from non-Git directory should fail
	tempDir := t.TempDir()

	repoInfo, err := GetRepositoryInfo(tempDir)
	if err == nil {
		t.Error("Expected error when getting repository info from non-Git directory, got nil")
	}

	if repoInfo != nil {
		t.Error("Expected repository info to be nil for non-Git directory")
	}

	expectedError := "failed to get current branch"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
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
