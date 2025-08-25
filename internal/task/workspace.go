package task

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/types"
)

// CreateWorkspaceForTask creates a workspace for a task
func (m *Manager) CreateWorkspaceForTask(task *types.Task, description string) error {
	// Generate workspace ID (use task ID as base)
	workspaceID := task.ID

	// Create workspace directory
	workspacePath := filepath.Join(m.workspacesDir, workspaceID)
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Clean up any existing content
	if err := os.RemoveAll(workspacePath); err != nil {
		return fmt.Errorf("failed to clean workspace directory: %w", err)
	}
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return fmt.Errorf("failed to recreate workspace directory: %w", err)
	}

	// Get current repository information
	repoInfo, err := git.DetectCurrentRepository()
	if err != nil {
		return fmt.Errorf("failed to detect current repository: %w", err)
	}

	// Create branch name
	branchName := fmt.Sprintf("task/%s", task.Name)

	// Create worktree in the workspace directory
	cmd := exec.Command("git", "worktree", "add", workspacePath, branchName)
	cmd.Dir = repoInfo.Path
	if err := cmd.Run(); err != nil {
		// If worktree creation fails, try creating a new branch
		cmd = exec.Command("git", "checkout", "-b", branchName)
		cmd.Dir = repoInfo.Path
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}

		// Copy files to workspace directory (excluding .git and .cw)
		cmd = exec.Command("rsync", "-av", "--exclude=.git", "--exclude=.cw", ".", workspacePath+"/")
		cmd.Dir = repoInfo.Path
		if err := cmd.Run(); err != nil {
			// Fallback to cp if rsync is not available
			cmd = exec.Command("cp", "-r", ".", workspacePath)
			cmd.Dir = repoInfo.Path
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to copy files to workspace: %w", err)
			}
		}

		// Initialize git in workspace
		cmd = exec.Command("git", "init")
		cmd.Dir = workspacePath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize git in workspace: %w", err)
		}

		// Add remote
		cmd = exec.Command("git", "remote", "add", "origin", repoInfo.Path)
		cmd.Dir = workspacePath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add remote: %w", err)
		}
	}

	// Save workspace metadata
	workspaceMetadata := map[string]interface{}{
		"id":            workspaceID,
		"task_name":     task.Name,
		"description":   description,
		"branch_name":   branchName,
		"source_repo":   repoInfo.Path,
		"base_branch":   repoInfo.CurrentBranch,
		"created_at":    time.Now(),
		"last_activity": time.Now(),
		"status":        "ready",
		"metadata":      make(map[string]string),
	}

	metadataPath := filepath.Join(workspacePath, ".workspace.json")
	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(workspaceMetadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	// Update task with workspace information
	task.WorkspaceID = workspaceID
	task.WorkspacePath = workspacePath
	task.BranchName = branchName
	task.SourceRepo = repoInfo.Path
	task.BaseBranch = repoInfo.CurrentBranch

	return nil
}

// GetTaskWorkspacePath returns the workspace path for a task
func (m *Manager) GetTaskWorkspacePath(taskID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return "", fmt.Errorf("task not found: %s", taskID)
	}

	if task.WorkspacePath == "" {
		return "", fmt.Errorf("task '%s' has no associated workspace", task.Name)
	}

	return task.WorkspacePath, nil
}

// RunGitInTaskWorkspace runs git commands in a task's workspace
func (m *Manager) RunGitInTaskWorkspace(taskID string, gitArgs []string) error {
	workspacePath, err := m.GetTaskWorkspacePath(taskID)
	if err != nil {
		return err
	}

	// Run git command in the workspace directory
	cmd := exec.Command("git", gitArgs...)
	cmd.Dir = workspacePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
