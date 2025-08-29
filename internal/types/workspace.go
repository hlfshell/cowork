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

	// Task ID if this workspace is created for a specific task (optional)
	// If provided, the workspace will share the same ID as the task
	TaskID int `json:"task_id,omitempty"`

	// Container configuration (optional)
	ContainerConfig *ContainerConfig `json:"container_config,omitempty"`
}

// ContainerConfig defines the container configuration for a workspace
type ContainerConfig struct {
	// Container image to use
	Image string `json:"image"`

	// Container name (optional, will be auto-generated if not provided)
	Name string `json:"name,omitempty"`

	// Working directory inside the container
	WorkingDir string `json:"working_dir,omitempty"`

	// Command to run in the container
	Command []string `json:"command,omitempty"`

	// Arguments for the command
	Args []string `json:"args,omitempty"`

	// Environment variables
	Environment map[string]string `json:"environment,omitempty"`

	// Port mappings (host:container)
	Ports map[string]string `json:"ports,omitempty"`

	// Additional volume mounts (host:container)
	Volumes map[string]string `json:"volumes,omitempty"`

	// Network to connect to
	Network string `json:"network,omitempty"`

	// User to run as
	User string `json:"user,omitempty"`

	// Whether to run in privileged mode
	Privileged bool `json:"privileged,omitempty"`

	// Whether to run in detached mode
	Detached bool `json:"detached,omitempty"`

	// Whether to remove container on exit
	Remove bool `json:"remove,omitempty"`

	// Whether to allocate a TTY
	TTY bool `json:"tty,omitempty"`

	// Whether to run interactively
	Interactive bool `json:"interactive,omitempty"`

	// Labels to attach to the container
	Labels map[string]string `json:"labels,omitempty"`
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
	// If this workspace was created for a task, this ID matches the task ID
	ID int `json:"id,string"`

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

	// Container configuration (if any)
	ContainerConfig *ContainerConfig `json:"container_config,omitempty"`

	// Container status information
	ContainerStatus *ContainerStatus `json:"container_status,omitempty"`

	// Metadata for the workspace
	Metadata map[string]string `json:"metadata,omitempty"`

	// Task ID if this workspace was created for a specific task
	// If this is a standalone workspace, this will be 0
	TaskID int `json:"task_id,omitempty"`

	// IsTaskWorkspace indicates whether this workspace was created for a task
	// This is a computed field based on whether TaskID is set
	IsTaskWorkspace bool `json:"is_task_workspace"`
}

// ContainerStatus represents the current status of a workspace container
type ContainerStatus struct {
	// Container ID
	ID string `json:"id"`

	// Container name
	Name string `json:"name"`

	// Container image
	Image string `json:"image"`

	// Container status (running, stopped, etc.)
	Status string `json:"status"`

	// Container creation time
	Created string `json:"created"`

	// Exposed ports
	Ports []string `json:"ports"`

	// Container labels
	Labels map[string]string `json:"labels"`
}
