package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hlfshell/cowork/internal/types"
)

// Branch name generation constants
const (
	// MaxBranchNameLength is the maximum length for branch names (30 characters for git operations)
	MaxBranchNameLength = 30

	// DefaultTaskName is the default name used when no valid task name can be generated
	DefaultTaskName = "task"
)

// GitOperations provides Git-related functionality for workspace management
type GitOperations struct {
	// Default timeout for Git operations in seconds
	operationTimeoutSeconds int
}

// NewGitOperations creates a new GitOperations instance
func NewGitOperations(operationTimeoutSeconds int) *GitOperations {
	if operationTimeoutSeconds <= 0 {
		operationTimeoutSeconds = 300 // 5 minutes default
	}

	return &GitOperations{
		operationTimeoutSeconds: operationTimeoutSeconds,
	}
}

// CloneRepository clones a repository using full clone
func (g *GitOperations) CloneRepository(req *types.CreateWorkspaceRequest, workspacePath string) error {
	// Validate the request
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid create workspace request: %w", err)
	}

	// Ensure the workspace directory doesn't exist
	if _, err := os.Stat(workspacePath); err == nil {
		return fmt.Errorf("workspace directory already exists: %s", workspacePath)
	}

	// Create the workspace directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Perform the full clone
	return g.performFullClone(req, workspacePath)
}

// performFullClone creates a complete independent clone of the repository
func (g *GitOperations) performFullClone(req *types.CreateWorkspaceRequest, workspacePath string) error {
	// Check if source is a local path or remote URL
	if isLocalPath(req.SourceRepo) {
		return g.cloneFromLocalPath(req.SourceRepo, workspacePath, req)
	}

	// Build the clone command for remote repository
	cloneArgs := []string{"clone", req.SourceRepo, workspacePath}

	// Execute the clone command
	cloneCmd := exec.Command("git", cloneArgs...)
	cloneCmd.Dir = filepath.Dir(workspacePath)

	output, err := cloneCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	// Checkout the specified branch if it's not the default
	if req.BaseBranch != "main" && req.BaseBranch != "master" {
		if err := g.checkoutBranch(workspacePath, req.BaseBranch); err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w", req.BaseBranch, err)
		}
	}

	// Create a new branch for the task
	// Use branch name from metadata if provided, otherwise generate one
	var branchName string
	if req.Metadata != nil && req.Metadata["branch_name"] != "" {
		branchName = req.Metadata["branch_name"]
	} else {
		branchName = g.generateBranchName(req.TaskName, req.TicketID)
	}

	if err := g.createAndCheckoutBranch(workspacePath, branchName); err != nil {
		return fmt.Errorf("failed to create task branch: %w", err)
	}

	return nil
}

// isLocalPath checks if the given path is a local filesystem path
func isLocalPath(path string) bool {
	// Check if it's a URL (starts with http://, https://, git://, etc.)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "git://") || strings.HasPrefix(path, "ssh://") ||
		strings.HasPrefix(path, "file://") {
		return false
	}

	// Check if it's a relative or absolute path
	return filepath.IsAbs(path) || !strings.Contains(path, "://")
}

// cloneFromLocalPath clones from a local repository path
func (g *GitOperations) cloneFromLocalPath(sourcePath, workspacePath string, req *types.CreateWorkspaceRequest) error {
	// Execute the clone command
	cloneCmd := exec.Command("git", "clone", sourcePath, workspacePath)
	cloneCmd.Dir = filepath.Dir(workspacePath)

	output, err := cloneCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone from local path failed: %w, output: %s", err, string(output))
	}

	// Checkout the specified branch if it's not the default
	if req.BaseBranch != "main" && req.BaseBranch != "master" {
		if err := g.checkoutBranch(workspacePath, req.BaseBranch); err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w", req.BaseBranch, err)
		}
	}

	// Create a new branch for the task
	branchName := g.generateBranchName(req.TaskName, req.TicketID)
	if err := g.createAndCheckoutBranch(workspacePath, branchName); err != nil {
		return fmt.Errorf("failed to create task branch: %w", err)
	}

	return nil
}

// checkoutBranch checks out a specific branch in the repository
func (g *GitOperations) checkoutBranch(repoPath, branchName string) error {
	// First, try to checkout the branch directly
	checkoutCmd := exec.Command("git", "checkout", branchName)
	checkoutCmd.Dir = repoPath

	output, err := checkoutCmd.CombinedOutput()
	if err != nil {
		// If the branch doesn't exist locally, try to fetch and checkout
		if strings.Contains(string(output), "did not match any file") {
			// Fetch the branch from remote
			fetchCmd := exec.Command("git", "fetch", "origin", branchName)
			fetchCmd.Dir = repoPath

			fetchOutput, fetchErr := fetchCmd.CombinedOutput()
			if fetchErr != nil {
				return fmt.Errorf("failed to fetch branch %s: %w, output: %s", branchName, fetchErr, string(fetchOutput))
			}

			// Try checkout again
			checkoutCmd = exec.Command("git", "checkout", branchName)
			checkoutCmd.Dir = repoPath

			output, err = checkoutCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to checkout branch %s after fetch: %w, output: %s", branchName, err, string(output))
			}
		} else {
			return fmt.Errorf("failed to checkout branch %s: %w, output: %s", branchName, err, string(output))
		}
	}

	return nil
}

// createAndCheckoutBranch creates a new branch and checks it out
func (g *GitOperations) createAndCheckoutBranch(repoPath, branchName string) error {
	// Create and checkout the new branch
	checkoutCmd := exec.Command("git", "checkout", "-b", branchName)
	checkoutCmd.Dir = repoPath

	output, err := checkoutCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create and checkout branch %s: %w, output: %s", branchName, err, string(output))
	}

	return nil
}

// generateBranchName creates a branch name from the task name and ticket ID
func (g *GitOperations) generateBranchName(taskName, ticketID string) string {
	// Sanitize the task name for use in branch names
	sanitizedTaskName := g.sanitizeBranchName(taskName)

	if ticketID != "" {
		return fmt.Sprintf("task/%s-%s", sanitizedTaskName, ticketID)
	}

	return fmt.Sprintf("task/%s", sanitizedTaskName)
}

// sanitizeBranchName converts a task name into a valid Git branch name
func (g *GitOperations) sanitizeBranchName(taskName string) string {
	// Convert to lowercase for consistency
	sanitized := strings.ToLower(taskName)

	// Replace invalid characters with hyphens
	invalidChars := []string{" ", "/", "\\", ":", "*", "?", "\"", "<", ">", "|", "..", "#", "&", "(", ")", "[", "]", "{", "}", "!", "@", "$", "%", "^", "+", "=", "~", "`", "_"}
	for _, char := range invalidChars {
		sanitized = strings.ReplaceAll(sanitized, char, "-")
	}

	// Remove leading/trailing hyphens and dots
	sanitized = strings.Trim(sanitized, "-.")

	// Ensure it's not empty
	if sanitized == "" {
		sanitized = DefaultTaskName
	}

	// Limit length for better readability
	if len(sanitized) > MaxBranchNameLength {
		sanitized = sanitized[:MaxBranchNameLength]
		sanitized = strings.TrimRight(sanitized, "-")
	}

	// Remove multiple consecutive hyphens
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}

	return sanitized
}

// GetRepositoryInfo retrieves information about a Git repository
func (g *GitOperations) GetRepositoryInfo(repoPath string) (*RepositoryInfo, error) {
	return GetRepositoryInfo(repoPath)
}
