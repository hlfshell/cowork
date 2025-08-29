package gitprovider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/stretchr/testify/assert"
)

// Import constants from the implementation
const (
	TestMaxBranchNameLength = 25
	TestDefaultTaskName     = "task"
)

func TestGitHubCoworkProvider_GenerateBranchName(t *testing.T) {
	provider := &GitHubCoworkProvider{}

	// Test cases covering various scenarios with realistic expectations
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		// Basic functionality
		{"Fix login bug", "fix-login-bug", "simple title with spaces"},
		{"Update API v2.0 endpoints", "update-api-v20-endpoints", "title with numbers and dots"},
		{"Implement OAuth2 Authentication", "implement-oauth2-authenti", "mixed case (truncated)"},

		// Special character handling
		{"Add user authentication & authorization!", "add-user-authentication-a", "ampersand and exclamation (truncated)"},
		{"Fix bug (critical) in payment processing", "fix-bug-critical-in-payme", "parentheses (truncated)"},
		{"Update [API] documentation", "update-api-documentation", "brackets"},
		{"Bug: Fix memory leak in worker", "bug-fix-memory-leak-in-wo", "colon prefix"},
		{`Fix "undefined" error in console`, "fix-undefined-error-in-co", "quotes"},
		{"Fix *important* security vulnerability", "fix-important-security-vu", "asterisks (truncated)"},
		{"How to implement caching?", "how-to-implement-caching", "question mark"},
		{"Urgent! Fix production bug!", "urgent-fix-production-bug", "exclamation marks"},
		{"Update <Component> props", "update-component-props", "angle brackets"},
		{"Add logging | monitoring | alerting", "add-logging-monitoring-al", "pipe characters (truncated)"},

		// Underscore and space handling
		{"Fix database_connection issues", "fix-database-connection-i", "underscores converted to hyphens"},
		{"  Refactor   user   management   module  ", "refactor-user-management", "multiple spaces (truncated)"},

		// Edge cases
		{"Fix..double..dots..issue", "fixdoubledotsissue", "double dots removed"},
		{"-Fix-leading-trailing-hyphens-", "fix-leading-trailing-hyph", "leading/trailing hyphens removed"},
		{".Fix.leading.trailing.dots.", "fixleadingtrailingdots", "leading/trailing dots removed"},
		{"Fix\tbug\nin\tauthentication", "fix-bug-in-authentication", "tabs and newlines"},

		// Empty and invalid inputs
		{"", "task", "empty title"},
		{"   ", "task", "only spaces"},
		{"---", "task", "only hyphens"},
		{"...", "task", "only dots"},
		{"!@#$%^&*()_+-=[]{}|;':\",./<>?", "task", "only special characters"},
		{"Valid@#$%^&*()Invalid", "validinvalid", "mixed valid and invalid characters"},

		// Unicode and emoji handling
		{"🐛 Fix bug in authentication 🔐", "fix-bug-in-authentication", "emojis removed"},
		{"Fix café authentication issue", "fix-caf-authentication-is", "unicode characters"},

		// Very long titles
		{"This is a very long title that should be truncated to thirty characters maximum to keep branch names short and manageable", "this-is-a-very-long-title", "very long title truncated"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			issue := &git.Issue{Title: tc.input}
			result := provider.GenerateBranchName(issue)

			// Test that the result is not longer than the maximum length
			if len(result) > TestMaxBranchNameLength {
				t.Errorf("Result too long: got %d characters, want <= %d", len(result), TestMaxBranchNameLength)
			}

			// Test that the result matches expected
			if result != tc.expected {
				t.Errorf("GenerateBranchName() = %q, want %q", result, tc.expected)
			}

			// Test that the result is a valid Git branch name
			if result != "" {
				if err := validateGitBranchName(result); err != nil {
					t.Errorf("Invalid Git branch name: %v", err)
				}
			}

			// Test that the result doesn't start or end with hyphens
			if result != "" && (strings.HasPrefix(result, "-") || strings.HasSuffix(result, "-")) {
				t.Errorf("Result starts or ends with hyphen: %q", result)
			}
		})
	}
}

// TestGitHubCoworkProvider_GenerateBranchName_Improved tests the improved branch name generation
func TestGitHubCoworkProvider_GenerateBranchName_Improved(t *testing.T) {
	provider := &GitHubCoworkProvider{}

	// Test cases covering the improved branch name generation
	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		// Basic functionality with improved length limits
		{"Fix login bug", "fix-login-bug", "simple title with spaces"},
		{"Update API v2.0 endpoints", "update-api-v20-endpoints", "title with numbers and dots"},
		{"Implement OAuth2 Authentication", "implement-oauth2-authenti", "mixed case (truncated to 25 chars)"},

		// Special character handling
		{"Add user authentication & authorization!", "add-user-authentication-a", "ampersand and exclamation (truncated)"},
		{"Fix bug (critical) in payment processing", "fix-bug-critical-in-payme", "parentheses (truncated)"},
		{"Update [API] documentation", "update-api-documentation", "brackets"},
		{"Bug: Fix memory leak in worker", "bug-fix-memory-leak-in-wo", "colon prefix (truncated)"},
		{`Fix "undefined" error in console`, "fix-undefined-error-in-co", "quotes (truncated)"},
		{"Fix *important* security vulnerability", "fix-important-security-vu", "asterisks (truncated)"},
		{"How to implement caching?", "how-to-implement-caching", "question mark"},
		{"Urgent! Fix production bug!", "urgent-fix-production-bug", "exclamation marks"},
		{"Update <Component> props", "update-component-props", "angle brackets"},
		{"Add logging | monitoring | alerting", "add-logging-monitoring-al", "pipe characters (truncated)"},

		// Underscore and space handling
		{"Fix database_connection issues", "fix-database-connection-i", "underscores converted to hyphens (truncated)"},
		{"  Refactor   user   management   module  ", "refactor-user-management", "multiple spaces (truncated)"},

		// Edge cases
		{"Fix..double..dots..issue", "fixdoubledotsissue", "double dots removed"},
		{"-Fix-leading-trailing-hyphens-", "fix-leading-trailing-hyph", "leading/trailing hyphens (truncated)"},
		{"", "task", "empty string defaults to task"},
		{"A", "a", "single character"},
		{"This is a very long task name that should be truncated to fit within the 25 character limit", "this-is-a-very-long-task", "very long name truncated to 25 chars"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Create a mock issue with the test title
			issue := &git.Issue{
				Title: tc.input,
			}

			result := provider.GenerateBranchName(issue)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
		})
	}
}

// TestGitHubCoworkProvider_GenerateBranchName_LengthLimit tests the 25 character length limit
func TestGitHubCoworkProvider_GenerateBranchName_LengthLimit(t *testing.T) {
	// Test case: Branch names should be limited to 25 characters for better readability
	provider := &GitHubCoworkProvider{}

	// Test various lengths
	testCases := []struct {
		input  string
		maxLen int
		desc   string
	}{
		{"Short name", 11, "short name within limit"},
		{"This is a medium length task name", 25, "medium name at limit"},
		{"This is a very long task name that should be truncated to fit within the limit", 25, "long name truncated"},
		{"A" + strings.Repeat("b", 50), 25, "very long name truncated"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			issue := &git.Issue{
				Title: tc.input,
			}
			result := provider.GenerateBranchName(issue)
			assert.LessOrEqual(t, len(result), tc.maxLen, "Branch name '%s' exceeds length limit", result)
		})
	}
}

// TestGitHubCoworkProvider_GenerateBranchName_ConsecutiveHyphens tests consecutive hyphen removal
func TestGitHubCoworkProvider_GenerateBranchName_ConsecutiveHyphens(t *testing.T) {
	// Test case: Multiple consecutive hyphens should be reduced to single hyphens
	provider := &GitHubCoworkProvider{}

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
			issue := &git.Issue{
				Title: tc.input,
			}
			result := provider.GenerateBranchName(issue)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
			// Verify no consecutive hyphens remain
			assert.False(t, strings.Contains(result, "--"), "Result contains consecutive hyphens: %s", result)
		})
	}
}

// TestGitHubCoworkProvider_GenerateBranchName_Lowercase tests lowercase conversion
func TestGitHubCoworkProvider_GenerateBranchName_Lowercase(t *testing.T) {
	// Test case: All branch names should be converted to lowercase
	provider := &GitHubCoworkProvider{}

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
			issue := &git.Issue{
				Title: tc.input,
			}
			result := provider.GenerateBranchName(issue)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
			// Verify result is lowercase
			assert.Equal(t, strings.ToLower(result), result, "Result is not lowercase: %s", result)
		})
	}
}

// TestGitHubCoworkProvider_GenerateBranchName_EmptyHandling tests empty string handling
func TestGitHubCoworkProvider_GenerateBranchName_EmptyHandling(t *testing.T) {
	// Test case: Empty or whitespace-only titles should default to "task"
	provider := &GitHubCoworkProvider{}

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{"", "task", "empty string"},
		{"   ", "task", "whitespace only"},
		{"\t\n\r", "task", "control characters only"},
		{"-", "task", "single hyphen"},
		{"--", "task", "double hyphens"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			issue := &git.Issue{
				Title: tc.input,
			}
			result := provider.GenerateBranchName(issue)
			assert.Equal(t, tc.expected, result, "Input: '%s'", tc.input)
		})
	}
}

// validateGitBranchName checks if a string is a valid Git branch name
func validateGitBranchName(name string) error {
	// Git branch name rules:
	// - Cannot start with '-'
	// - Cannot contain '..'
	// - Cannot contain control characters
	// - Cannot contain spaces, tabs, newlines, or other whitespace
	// - Cannot contain '~', '^', ':', '?', '*', '[', '\\'

	if name == "" {
		return nil // Empty is valid for our use case
	}

	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("branch name cannot start with '-'")
	}

	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name cannot contain '..'")
	}

	// Check for invalid characters
	invalidChars := []rune{'~', '^', ':', '?', '*', '[', '\\', ' ', '\t', '\n', '\r'}
	for _, char := range invalidChars {
		if strings.ContainsRune(name, char) {
			return fmt.Errorf("branch name contains invalid character: %c", char)
		}
	}

	// Check for control characters
	for _, char := range name {
		if char < 32 || char == 127 {
			return fmt.Errorf("branch name contains control character: %d", char)
		}
	}

	return nil
}
