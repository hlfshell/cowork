package workspace

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hlfshell/cowork/internal/git"
	"github.com/hlfshell/cowork/internal/types"
)

// Manager implements the WorkspaceManager interface
type Manager struct {
	// Git operations handler
	gitOps git.GitOperationsInterface

	// Base directory for all workspaces
	baseDir string

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewManager creates a new workspace manager for the current project
func NewManager(gitTimeoutSeconds int) (*Manager, error) {
	// Detect the current repository
	repoInfo, err := git.DetectCurrentRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to detect current repository: %w", err)
	}

	// Create project-specific workspace directory
	projectDir := filepath.Join(repoInfo.Path, ".cw", "workspaces")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project workspace directory: %w", err)
	}

	return &Manager{
		gitOps:  git.NewGitOperations(gitTimeoutSeconds),
		baseDir: projectDir,
	}, nil
}

// CreateWorkspace creates a new isolated workspace
func (m *Manager) CreateWorkspace(req *types.CreateWorkspaceRequest) (*types.Workspace, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if a workspace with this task name already exists
	existingWorkspaces, err := DiscoverWorkspaces(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing workspaces: %w", err)
	}

	for _, ws := range existingWorkspaces {
		if ws.TaskName == req.TaskName {
			return nil, fmt.Errorf("workspace with task name '%s' already exists", req.TaskName)
		}
	}

	// Generate a unique workspace ID
	workspaceID, err := m.generateWorkspaceID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate workspace ID: %w", err)
	}

	// Create the workspace path
	workspacePath := filepath.Join(m.baseDir, workspaceID)

	// Create the workspace object
	workspace := &types.Workspace{
		ID:           workspaceID,
		TaskName:     req.TaskName,
		Description:  req.Description,
		TicketID:     req.TicketID,
		Path:         workspacePath,
		SourceRepo:   req.SourceRepo,
		BaseBranch:   req.BaseBranch,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Status:       types.WorkspaceStatusCreating,
		Metadata:     req.Metadata,
	}

	// Clone the repository
	if err := m.gitOps.CloneRepository(req, workspacePath); err != nil {
		// Remove the workspace directory on failure
		os.RemoveAll(workspacePath)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repository information to populate branch name
	repoInfo, err := m.gitOps.GetRepositoryInfo(workspacePath)
	if err != nil {
		// This is not a critical error, but we should log it
		fmt.Printf("Warning: failed to get repository info: %v\n", err)
	} else {
		workspace.BranchName = repoInfo.CurrentBranch
	}

	// Update status to ready
	workspace.Status = types.WorkspaceStatusReady
	workspace.LastActivity = time.Now()

	// Save workspace metadata
	if err := SaveWorkspaceMetadata(workspacePath, workspace); err != nil {
		// This is not a critical error, but we should log it
		fmt.Printf("Warning: failed to save workspace metadata: %v\n", err)
	}

	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (m *Manager) GetWorkspace(workspaceID string) (*types.Workspace, error) {
	workspacePath := filepath.Join(m.baseDir, workspaceID)

	// Check if the workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace not found: %s", workspaceID)
	}

	// Load workspace metadata
	workspace, err := LoadWorkspaceMetadata(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace metadata: %w", err)
	}

	return workspace, nil
}

// ListWorkspaces returns all workspaces
func (m *Manager) ListWorkspaces() ([]*types.Workspace, error) {
	return DiscoverWorkspaces(m.baseDir)
}

// DeleteWorkspace removes a workspace and cleans up its resources
func (m *Manager) DeleteWorkspace(workspaceID string) error {
	workspacePath := filepath.Join(m.baseDir, workspaceID)

	// Check if the workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found: %s", workspaceID)
	}

	// Load workspace metadata to get the path
	workspace, err := LoadWorkspaceMetadata(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to load workspace metadata: %w", err)
	}

	// Update status to cleaning
	workspace.Status = types.WorkspaceStatusCleaning
	if err := UpdateWorkspaceMetadata(workspace); err != nil {
		// This is not a critical error, but we should log it
		fmt.Printf("Warning: failed to update workspace status: %v\n", err)
	}

	// Remove the workspace directory
	if err := os.RemoveAll(workspace.Path); err != nil {
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	return nil
}

// UpdateWorkspaceStatus updates the status of a workspace
func (m *Manager) UpdateWorkspaceStatus(workspaceID string, status types.WorkspaceStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid workspace status: %s", status)
	}

	workspacePath := filepath.Join(m.baseDir, workspaceID)

	// Check if the workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found: %s", workspaceID)
	}

	// Load workspace metadata
	workspace, err := LoadWorkspaceMetadata(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to load workspace metadata: %w", err)
	}

	// Update the status
	workspace.Status = status
	workspace.LastActivity = time.Now()

	// Save the updated metadata
	return UpdateWorkspaceMetadata(workspace)
}

// GetWorkspaceByTaskName retrieves a workspace by task name
func (m *Manager) GetWorkspaceByTaskName(taskName string) (*types.Workspace, error) {
	workspaces, err := DiscoverWorkspaces(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover workspaces: %w", err)
	}

	for _, workspace := range workspaces {
		if workspace.TaskName == taskName {
			return workspace, nil
		}
	}

	return nil, fmt.Errorf("workspace not found with task name: %s", taskName)
}

// generateWorkspaceID creates a unique workspace identifier
func (m *Manager) generateWorkspaceID() (string, error) {
	// Generate 8 random bytes
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	return hex.EncodeToString(bytes), nil
}

// GetBaseDirectory returns the base directory for workspaces
func (m *Manager) GetBaseDirectory() string {
	return m.baseDir
}

// CleanupOrphanedWorkspaces removes workspaces that exist on disk but not in memory
func (m *Manager) CleanupOrphanedWorkspaces() error {
	// Get all directories in the base directory
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base directory: %w", err)
	}

	// Check each directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceID := entry.Name()
		workspacePath := filepath.Join(m.baseDir, workspaceID)

		// Try to load metadata for this workspace
		_, err := LoadWorkspaceMetadata(workspacePath)
		if err != nil {
			// This workspace has invalid or missing metadata, remove it
			if err := os.RemoveAll(workspacePath); err != nil {
				fmt.Printf("Warning: failed to remove orphaned workspace %s: %v\n", workspaceID, err)
			} else {
				fmt.Printf("Removed orphaned workspace: %s\n", workspaceID)
			}
		}
	}

	return nil
}
