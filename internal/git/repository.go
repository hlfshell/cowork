package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepositoryInfo contains information about a Git repository
type RepositoryInfo struct {
	Path              string `json:"path"`
	CurrentBranch     string `json:"current_branch"`
	RemoteURL         string `json:"remote_url"`
	LastCommit        string `json:"last_commit"`
	LastCommitMessage string `json:"last_commit_message"`
}

// DetectCurrentRepository detects if the current directory is a Git repository and returns its information
func DetectCurrentRepository() (*RepositoryInfo, error) {
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	
	// Check if this is a Git repository
	if !isGitRepository(currentDir) {
		return nil, fmt.Errorf("current directory is not a Git repository: %s", currentDir)
	}
	
	// Get repository information
	repoInfo, err := GetRepositoryInfo(currentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}
	
	return repoInfo, nil
}

// isGitRepository checks if the given directory is a Git repository
func isGitRepository(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetRepositoryInfo retrieves information about a Git repository
func GetRepositoryInfo(repoPath string) (*RepositoryInfo, error) {
	info := &RepositoryInfo{
		Path: repoPath,
	}
	
	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = repoPath
	
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	info.CurrentBranch = strings.TrimSpace(string(branchOutput))
	
	// Get remote URL
	remoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	remoteCmd.Dir = repoPath
	
	remoteOutput, err := remoteCmd.Output()
	if err == nil {
		info.RemoteURL = strings.TrimSpace(string(remoteOutput))
	}
	
	// Get last commit hash
	commitCmd := exec.Command("git", "rev-parse", "HEAD")
	commitCmd.Dir = repoPath
	
	commitOutput, err := commitCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get last commit: %w", err)
	}
	info.LastCommit = strings.TrimSpace(string(commitOutput))
	
	// Get last commit message
	messageCmd := exec.Command("git", "log", "-1", "--pretty=format:%s")
	messageCmd.Dir = repoPath
	
	messageOutput, err := messageCmd.Output()
	if err == nil {
		info.LastCommitMessage = strings.TrimSpace(string(messageOutput))
	}
	
	return info, nil
}
