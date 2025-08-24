package workspace

import "github.com/hlfshell/cowork/internal/types"

// WorkspaceManager defines the interface for workspace operations
type WorkspaceManager interface {
	// CreateWorkspace creates a new isolated workspace
	CreateWorkspace(req *types.CreateWorkspaceRequest) (*types.Workspace, error)

	// GetWorkspace retrieves a workspace by ID
	GetWorkspace(workspaceID string) (*types.Workspace, error)

	// ListWorkspaces returns all workspaces
	ListWorkspaces() ([]*types.Workspace, error)

	// DeleteWorkspace removes a workspace and cleans up its resources
	DeleteWorkspace(workspaceID string) error

	// UpdateWorkspaceStatus updates the status of a workspace
	UpdateWorkspaceStatus(workspaceID string, status types.WorkspaceStatus) error

	// GetWorkspaceByTaskName retrieves a workspace by task name
	GetWorkspaceByTaskName(taskName string) (*types.Workspace, error)
}
