package types

import (
	"fmt"
	"time"
)

// WorkspaceStatus represents the current state of a workspace
type WorkspaceStatus string

const (
	// WorkspaceStatusCreating indicates the workspace is being created
	WorkspaceStatusCreating WorkspaceStatus = "creating"

	// WorkspaceStatusReady indicates the workspace is ready for use
	WorkspaceStatusReady WorkspaceStatus = "ready"

	// WorkspaceStatusActive indicates the workspace is currently in use
	WorkspaceStatusActive WorkspaceStatus = "active"

	// WorkspaceStatusError indicates the workspace encountered an error
	WorkspaceStatusError WorkspaceStatus = "error"

	// WorkspaceStatusCleaning indicates the workspace is being cleaned up
	WorkspaceStatusCleaning WorkspaceStatus = "cleaning"
)

// String returns the string representation of the workspace status
func (ws WorkspaceStatus) String() string {
	return string(ws)
}

// IsValid checks if the workspace status is valid
func (ws WorkspaceStatus) IsValid() bool {
	switch ws {
	case WorkspaceStatusCreating, WorkspaceStatusReady, WorkspaceStatusActive, WorkspaceStatusError, WorkspaceStatusCleaning:
		return true
	default:
		return false
	}
}

// CreateWorkspaceRequest contains the parameters for creating a new workspace
type CreateWorkspaceRequest struct {
	// Human-readable name for the task
	TaskName string `json:"task_name"`

	// Description of what the task is trying to accomplish
	Description string `json:"description,omitempty"`

	// External ticket ID (optional)
	TicketID string `json:"ticket_id,omitempty"`

	// Source repository URL
	SourceRepo string `json:"source_repo"`

	// Base branch to clone from (defaults to "main")
	BaseBranch string `json:"base_branch"`

	// Parent directory for workspace creation
	ParentDir string `json:"parent_dir"`

	// Metadata to attach to the workspace
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Validate checks if the create workspace request is valid
func (req *CreateWorkspaceRequest) Validate() error {
	if req.TaskName == "" {
		return fmt.Errorf("task name is required")
	}

	if req.SourceRepo == "" {
		return fmt.Errorf("source repository URL is required")
	}

	if req.BaseBranch == "" {
		req.BaseBranch = "main"
	}

	return nil
}

// Workspace represents an isolated workspace for a specific task
type Workspace struct {
	// Unique identifier for the workspace
	ID string `json:"id"`

	// Human-readable name for the task
	TaskName string `json:"task_name"`

	// Description of what the task is trying to accomplish
	Description string `json:"description,omitempty"`

	// External ticket ID (e.g., GitHub #123)
	TicketID string `json:"ticket_id,omitempty"`

	// Path to the workspace directory
	Path string `json:"path"`

	// Git branch name for this workspace
	BranchName string `json:"branch_name"`

	// Source repository URL
	SourceRepo string `json:"source_repo"`

	// Base branch (usually main/master)
	BaseBranch string `json:"base_branch"`

	// Creation timestamp
	CreatedAt time.Time `json:"created_at"`

	// Last activity timestamp
	LastActivity time.Time `json:"last_activity"`

	// Current status of the workspace
	Status WorkspaceStatus `json:"status"`

	// Associated container ID (if any)
	ContainerID string `json:"container_id,omitempty"`

	// Metadata for the workspace
	Metadata map[string]string `json:"metadata,omitempty"`
}
