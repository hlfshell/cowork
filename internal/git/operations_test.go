package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hlfshell/cowork/internal/types"
	"github.com/stretchr/testify/assert"
)

// Test constants to match implementation
const (
	TestMaxBranchNameLength = 30
	TestDefaultTaskName     = "task"
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
	expected := "task/fix-bug-123-with-spaces-456"

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

// TestGitOperations_sanitizeBranchName_Improved tests the improved branch name sanitization
func TestGitOperations_sanitizeBranchName_Improved(t *testing.T) {
	// Test case: Branch name sanitization should be more concise and human-readable
	gitOps := NewGitOperations(300)

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		// Basic functionality
		{"Fix login bug", "fix-login-bug", "simple title with spaces"},
		{"Update API v2.0 endpoints", "update-api-v2.0-endpoints", "title with numbers and dots"},
		{"Implement OAuth2 Authentication", "implement-oauth2-authenticatio", "mixed case (truncated)"},

		// Special character handling
		{"Add user authentication & authorization!", "add-user-authentication-auth", "ampersand and exclamation (truncated)"},
		{"Fix bug (critical) in payment processing", "fix-bug-critical-in-payment", "parentheses (truncated)"},
		{"Update [API] documentation", "update-api-documentation", "brackets"},
		{"Bug: Fix memory leak in worker", "bug-fix-memory-leak-in-worker", "colon prefix"},
		{`Fix "undefined" error in console`, "fix-undefined-error-in-conso", "quotes"},
		{"Fix *important* security vulnerability", "fix-important-security-vulne", "asterisks (truncated)"},
		{"How to implement caching?", "how-to-implement-caching", "question mark"},
		{"Urgent! Fix production bug!", "urgent-fix-production-bug", "exclamation marks"},
		{"Update <Component> props", "update-component-props", "angle brackets"},
		{"Add logging | monitoring | alerting", "add-logging-monitoring-ale", "pipe characters (truncated)"},

		// Underscore and space handling
		{"Fix database_connection issues", "fix-database-connection-issues", "underscores converted to hyphens"},
		{"  Refactor   user   management   module  ", "refactor-user-management", "multiple spaces (truncated)"},

		// Edge cases
		{"Fix..double..dots..issue", "fix-double-dots-issue", "double dots removed"},
		{"-Fix-leading-trailing-hyphens-", "fix-leading-trailing-hyphens", "leading/trailing hyphens removed"},
		{"", "task", "empty string defaults to task"},
		{"A", "a", "single character"},
		{"This is a very long task name that should be truncated to fit within the 30 character limit", "this-is-a-very-long-task-name", "very long name truncated"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := gitOps.sanitizeBranchName(tc.input)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
		})
	}
}

// TestGitOperations_sanitizeBranchName_LengthLimit tests the 30 character length limit
func TestGitOperations_sanitizeBranchName_LengthLimit(t *testing.T) {
	// Test case: Branch names should be limited to 30 characters for better readability
	gitOps := NewGitOperations(300)

	// Test various lengths
	testCases := []struct {
		input  string
		maxLen int
		desc   string
	}{
		{"Short name", 11, "short name within limit"},
		{"This is a medium length task name", TestMaxBranchNameLength, "medium name at limit"},
		{"This is a very long task name that should be truncated to fit within the limit", TestMaxBranchNameLength, "long name truncated"},
		{"A" + strings.Repeat("b", 50), TestMaxBranchNameLength, "very long name truncated"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := gitOps.sanitizeBranchName(tc.input)
			assert.LessOrEqual(t, len(result), tc.maxLen, "Branch name '%s' exceeds length limit", result)
		})
	}
}

// TestGitOperations_sanitizeBranchName_ConsecutiveHyphens tests consecutive hyphen removal
func TestGitOperations_sanitizeBranchName_ConsecutiveHyphens(t *testing.T) {
	// Test case: Multiple consecutive hyphens should be reduced to single hyphens
	gitOps := NewGitOperations(300)

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"Fix--double--hyphens", "fix-double-hyphens", "double hyphens"},
		{"Fix---triple---hyphens", "fix-triple-hyphens", "triple hyphens"},
		{"Fix----quadruple----hyphens", "fix-quadruple-hyphens", "quadruple hyphens"},
		{"Fix-----many-----hyphens", "fix-many-hyphens", "many hyphens"},
		{"Fix--hyphens--at--start", "fix-hyphens-at-start", "hyphens throughout"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := gitOps.sanitizeBranchName(tc.input)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
			// Verify no consecutive hyphens remain
			assert.False(t, strings.Contains(result, "--"), "Result contains consecutive hyphens: %s", result)
		})
	}
}

// TestGitOperations_sanitizeBranchName_Lowercase tests lowercase conversion
func TestGitOperations_sanitizeBranchName_Lowercase(t *testing.T) {
	// Test case: All branch names should be converted to lowercase
	gitOps := NewGitOperations(300)

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"UPPERCASE", "uppercase", "all uppercase"},
		{"MixedCase", "mixedcase", "mixed case"},
		{"Title Case", "title-case", "title case with spaces"},
		{"CamelCase", "camelcase", "camel case"},
		{"snake_case", "snake-case", "snake case"},
		{"kebab-case", "kebab-case", "already kebab case"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := gitOps.sanitizeBranchName(tc.input)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
			// Verify result is lowercase
			assert.Equal(t, strings.ToLower(result), result, "Result is not lowercase: %s", result)
		})
	}
}

// TestGitOperations_generateBranchName_Improved tests the improved branch name generation
func TestGitOperations_generateBranchName_Improved(t *testing.T) {
	// Test case: Branch name generation should be more concise and human-readable
	gitOps := NewGitOperations(300)

	testCases := []struct {
		taskName string
		ticketID string
		expected string
		desc     string
	}{
		// Task name only
		{"Fix login bug", "", "task/fix-login-bug", "task name only"},
		{"Update API", "", "task/update-api", "simple task name"},
		{"Implement OAuth2", "", "task/implement-oauth2", "mixed case task name"},

		// Task name with ticket ID
		{"Fix login bug", "123", "task/fix-login-bug-123", "task name with ticket"},
		{"Update API", "GH-456", "task/update-api-GH-456", "task name with ticket prefix"},
		{"Implement OAuth2", "ISSUE-789", "task/implement-oauth2-ISSUE-789", "task name with issue prefix"},

		// Special characters in task name
		{"Fix bug (critical)", "123", "task/fix-bug-critical-123", "parentheses in task name"},
		{"Update [API] docs", "456", "task/update-api-docs-456", "brackets in task name"},
		{"Fix *important* bug", "789", "task/fix-important-bug-789", "asterisks in task name"},

		// Long task names
		{"This is a very long task name that should be truncated", "123", "task/this-is-a-very-long-task-name-123", "long task name"},
		{"Short", "123", "task/short-123", "short task name"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := gitOps.generateBranchName(tc.taskName, tc.ticketID)
			assert.Equal(t, tc.expected, result, "Task: %s, Ticket: %s", tc.taskName, tc.ticketID)
		})
	}
}
