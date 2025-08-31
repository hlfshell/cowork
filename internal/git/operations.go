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
func (g *GitOperations) CloneRepository(req *types.CreateWorkspaceRequest, workspacePath string, authInfo *types.GitAuthInfo) error {
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
	return g.performFullClone(req, workspacePath, authInfo)
}

// performFullClone creates a complete independent clone of the repository
func (g *GitOperations) performFullClone(req *types.CreateWorkspaceRequest, workspacePath string, authInfo *types.GitAuthInfo) error {
	// Check if source is a local path or remote URL
	if isLocalPath(req.SourceRepo) {
		return g.cloneFromLocalPath(req.SourceRepo, workspacePath, req)
	}

	// Build the clone command for remote repository
	cloneArgs := []string{"clone", req.SourceRepo, workspacePath}

	// Execute the clone command with authentication
	cloneCmd := g.getAuthenticatedCommand(filepath.Dir(workspacePath), authInfo, cloneArgs...)

	output, err := cloneCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	// Checkout the specified branch if it's not the default
	if req.BaseBranch != "main" && req.BaseBranch != "master" {
		if err := g.checkoutBranch(workspacePath, req.BaseBranch, authInfo); err != nil {
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

	if err := g.createAndCheckoutBranch(workspacePath, branchName, authInfo); err != nil {
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
		if err := g.checkoutBranch(workspacePath, req.BaseBranch, nil); err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w", req.BaseBranch, err)
		}
	}

	// Create a new branch for the task
	branchName := g.generateBranchName(req.TaskName, req.TicketID)
	if err := g.createAndCheckoutBranch(workspacePath, branchName, nil); err != nil {
		return fmt.Errorf("failed to create task branch: %w", err)
	}

	return nil
}

// checkoutBranch checks out a specific branch in the repository
func (g *GitOperations) checkoutBranch(repoPath, branchName string, authInfo *types.GitAuthInfo) error {
	// First, try to checkout the branch directly
	checkoutCmd := g.getAuthenticatedCommand(repoPath, authInfo, "checkout", branchName)

	output, err := checkoutCmd.CombinedOutput()
	if err != nil {
		// If the branch doesn't exist locally, try to fetch and checkout
		if strings.Contains(string(output), "did not match any file") {
			// Fetch the branch from remote
			fetchCmd := g.getAuthenticatedCommand(repoPath, authInfo, "fetch", "origin", branchName)

			fetchOutput, fetchErr := fetchCmd.CombinedOutput()
			if fetchErr != nil {
				return fmt.Errorf("failed to fetch branch %s: %w, output: %s", branchName, fetchErr, string(fetchOutput))
			}

			// Try checkout again
			checkoutCmd = g.getAuthenticatedCommand(repoPath, authInfo, "checkout", branchName)

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
func (g *GitOperations) createAndCheckoutBranch(repoPath, branchName string, authInfo *types.GitAuthInfo) error {
	// Create and checkout the new branch
	checkoutCmd := g.getAuthenticatedCommand(repoPath, authInfo, "checkout", "-b", branchName)

	output, err := checkoutCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create and checkout branch %s: %w, output: %s", branchName, err, string(output))
	}

	return nil
}

// PushBranch pushes a branch to the remote repository
func (g *GitOperations) PushBranch(repoPath, branchName string, authInfo *types.GitAuthInfo) error {
	// Push branch to origin
	pushCmd := g.getAuthenticatedCommand(repoPath, authInfo, "push", "-u", "origin", branchName)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr

	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch %s: %w", branchName, err)
	}

	return nil
}

// PullBranch pulls the latest changes for a branch
func (g *GitOperations) PullBranch(repoPath, branchName string, authInfo *types.GitAuthInfo) error {
	// Pull latest changes
	pullCmd := g.getAuthenticatedCommand(repoPath, authInfo, "pull", "origin", branchName)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull branch %s: %w", branchName, err)
	}

	return nil
}

// FetchOrigin fetches the latest changes from origin
func (g *GitOperations) FetchOrigin(repoPath string, authInfo *types.GitAuthInfo) error {
	// Fetch latest changes
	fetchCmd := g.getAuthenticatedCommand(repoPath, authInfo, "fetch", "origin")
	fetchCmd.Stdout = os.Stdout
	fetchCmd.Stderr = os.Stderr

	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch origin: %w", err)
	}

	return nil
}

// getAuthenticatedCommand creates a git command with authentication context
func (g *GitOperations) getAuthenticatedCommand(workspacePath string, authInfo *types.GitAuthInfo, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = workspacePath

	// Configure authentication if provided
	if authInfo != nil {
		g.configureAuth(workspacePath, authInfo)
	}

	return cmd
}

// configureAuth configures git authentication for the workspace
func (g *GitOperations) configureAuth(workspacePath string, authInfo *types.GitAuthInfo) {
	switch authInfo.Method {
	case types.GitAuthMethodSSH:
		g.configureSSHAuth(workspacePath, authInfo)
	case types.GitAuthMethodHTTPS:
		g.configureHTTPSAuth(workspacePath, authInfo)
	}
}

// configureSSHAuth configures SSH authentication
func (g *GitOperations) configureSSHAuth(workspacePath string, authInfo *types.GitAuthInfo) {
	if authInfo.SSHKeyPath == "" {
		return
	}

	// Expand ~ to home directory
	sshKeyPath := authInfo.SSHKeyPath
	if strings.HasPrefix(sshKeyPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return
		}
		sshKeyPath = filepath.Join(homeDir, sshKeyPath[1:])
	}

	// Verify SSH key exists
	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		return
	}

	// Configure git to use the SSH key
	cmd := exec.Command("git", "config", "core.sshCommand", fmt.Sprintf("ssh -i %s", sshKeyPath))
	cmd.Dir = workspacePath
	cmd.Run() // Ignore errors
}

// configureHTTPSAuth configures HTTPS authentication
func (g *GitOperations) configureHTTPSAuth(workspacePath string, authInfo *types.GitAuthInfo) {
	if authInfo.Username == "" || (authInfo.Password == "" && authInfo.Token == "") {
		return
	}

	// Configure git credential helper
	cmd := exec.Command("git", "config", "credential.helper", "store")
	cmd.Dir = workspacePath
	cmd.Run() // Ignore errors

	// Create credential file
	credentialFile := filepath.Join(workspacePath, ".git", "credentials")

	var credentialContent string
	if authInfo.Token != "" {
		credentialContent = fmt.Sprintf("https://%s@", authInfo.Token)
	} else {
		credentialContent = fmt.Sprintf("https://%s:%s@", authInfo.Username, authInfo.Password)
	}

	os.WriteFile(credentialFile, []byte(credentialContent), 0600) // Ignore errors
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
