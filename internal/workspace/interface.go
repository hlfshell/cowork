package workspace

import (
	"context"

	"github.com/hlfshell/cowork/internal/types"
)

// WorkspaceManager defines the interface for workspace management operations
type WorkspaceManager interface {
	// CreateWorkspace creates a new isolated workspace
	CreateWorkspace(req *types.CreateWorkspaceRequest) (*types.Workspace, error)

	// GetWorkspace retrieves a workspace by ID
	GetWorkspace(workspaceID int) (*types.Workspace, error)

	// ListWorkspaces returns all workspaces
	ListWorkspaces() ([]*types.Workspace, error)

	// DeleteWorkspace removes a workspace and cleans up its resources
	DeleteWorkspace(workspaceID int) error

	// UpdateWorkspaceStatus updates the status of a workspace
	UpdateWorkspaceStatus(workspaceID int, status types.WorkspaceStatus) error

	// GetWorkspaceByTaskName retrieves a workspace by task name
	GetWorkspaceByTaskName(taskName string) (*types.Workspace, error)

	// GetBaseDirectory returns the base directory for workspaces
	GetBaseDirectory() string

	// CleanupOrphanedWorkspaces removes workspaces that exist on disk but not in memory
	CleanupOrphanedWorkspaces() error

	// Container management methods
	StartContainer(ctx context.Context, workspaceID int) error
	StopContainer(ctx context.Context, workspaceID int, timeoutSeconds int) error
	GetContainerStatus(ctx context.Context, workspaceID int) (*types.ContainerStatus, error)
	ExecInContainer(ctx context.Context, workspaceID int, command []string) error
	GetContainerLogs(ctx context.Context, workspaceID int, follow bool, tail int) (string, error)
}
