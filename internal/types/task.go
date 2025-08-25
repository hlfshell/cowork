package types

import (
	"fmt"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	// TaskStatusQueued indicates the task is waiting to be processed
	TaskStatusQueued TaskStatus = "queued"

	// TaskStatusInProgress indicates the task is currently being worked on
	TaskStatusInProgress TaskStatus = "in_progress"

	// TaskStatusCompleted indicates the task has been completed
	TaskStatusCompleted TaskStatus = "completed"

	// TaskStatusFailed indicates the task encountered an error
	TaskStatusFailed TaskStatus = "failed"

	// TaskStatusCancelled indicates the task was cancelled
	TaskStatusCancelled TaskStatus = "cancelled"

	// TaskStatusPaused indicates the task is temporarily paused
	TaskStatusPaused TaskStatus = "paused"
)

// String returns the string representation of the task status
func (ts TaskStatus) String() string {
	return string(ts)
}

// IsValid checks if the task status is valid
func (ts TaskStatus) IsValid() bool {
	switch ts {
	case TaskStatusQueued, TaskStatusInProgress, TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled, TaskStatusPaused:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the task status is a terminal state
func (ts TaskStatus) IsTerminal() bool {
	switch ts {
	case TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// CreateTaskRequest contains the parameters for creating a new task
type CreateTaskRequest struct {
	// Human-readable name for the task (matches workspace name)
	Name string `json:"name"`

	// Description of what the task is trying to accomplish
	Description string `json:"description,omitempty"`

	// External ticket ID (optional)
	TicketID string `json:"ticket_id,omitempty"`

	// Optional URL related to the task
	URL string `json:"url,omitempty"`

	// Priority of the task (higher number = higher priority)
	Priority int `json:"priority" default:"0"`

	// Metadata for optional keys
	Metadata map[string]string `json:"metadata,omitempty"`

	// Tags for categorizing the task
	Tags []string `json:"tags,omitempty"`

	// Estimated completion time in minutes
	EstimatedMinutes int `json:"estimated_minutes,omitempty"`

	// Cost tracking for AI agent usage
	EstimatedCost float64 `json:"estimated_cost,omitempty"`
	Currency      string  `json:"currency,omitempty" default:"USD"`
}

// Validate checks if the create task request is valid
func (req *CreateTaskRequest) Validate() error {
	if req.Name == "" {
		return fmt.Errorf("task name is required")
	}

	if req.Priority < 0 {
		return fmt.Errorf("priority must be non-negative")
	}

	if req.EstimatedMinutes < 0 {
		return fmt.Errorf("estimated minutes must be non-negative")
	}

	return nil
}

// Task represents a queued task to be worked on
type Task struct {
	// Unique identifier for the task (matches workspace ID)
	ID string `json:"id"`

	// Human-readable name for the task (matches workspace name)
	Name string `json:"name"`

	// Description of what the task is trying to accomplish
	Description string `json:"description,omitempty"`

	// External ticket ID (e.g., GitHub #123)
	TicketID string `json:"ticket_id,omitempty"`

	// Optional URL related to the task
	URL string `json:"url,omitempty"`

	// Current status of the task
	Status TaskStatus `json:"status"`

	// Priority of the task (higher number = higher priority)
	Priority int `json:"priority"`

	// Creation timestamp
	CreatedAt time.Time `json:"created_at"`

	// Last activity timestamp
	LastActivity time.Time `json:"last_activity"`

	// Started timestamp (when status changed to in_progress)
	StartedAt *time.Time `json:"started_at,omitempty"`

	// Completed timestamp (when status changed to completed/failed/cancelled)
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Estimated completion time in minutes
	EstimatedMinutes int `json:"estimated_minutes,omitempty"`

	// Actual time spent in minutes
	ActualMinutes int `json:"actual_minutes,omitempty"`

	// Cost tracking for AI agent usage
	EstimatedCost float64 `json:"estimated_cost,omitempty"`
	ActualCost    float64 `json:"actual_cost,omitempty"`
	Currency      string  `json:"currency,omitempty" default:"USD"`

	// Tags for categorizing the task
	Tags []string `json:"tags,omitempty"`

	// Metadata for the task
	Metadata map[string]string `json:"metadata,omitempty"`

	// Error message if task failed
	ErrorMessage string `json:"error_message,omitempty"`

	// Workspace information (if workspace has been created)
	WorkspaceID   string `json:"workspace_id,omitempty"`
	WorkspacePath string `json:"workspace_path,omitempty"`
	BranchName    string `json:"branch_name,omitempty"`
	SourceRepo    string `json:"source_repo,omitempty"`
	BaseBranch    string `json:"base_branch,omitempty"`
}

// UpdateTaskRequest contains the parameters for updating a task
type UpdateTaskRequest struct {
	// Task ID to update
	TaskID string `json:"task_id"`

	// New status (optional)
	Status *TaskStatus `json:"status,omitempty"`

	// New description (optional)
	Description *string `json:"description,omitempty"`

	// New priority (optional)
	Priority *int `json:"priority,omitempty"`

	// New estimated minutes (optional)
	EstimatedMinutes *int `json:"estimated_minutes,omitempty"`

	// New actual minutes (optional)
	ActualMinutes *int `json:"actual_minutes,omitempty"`

	// New estimated cost (optional)
	EstimatedCost *float64 `json:"estimated_cost,omitempty"`

	// New actual cost (optional)
	ActualCost *float64 `json:"actual_cost,omitempty"`

	// New currency (optional)
	Currency *string `json:"currency,omitempty"`

	// New tags (optional)
	Tags *[]string `json:"tags,omitempty"`

	// New metadata (optional)
	Metadata *map[string]string `json:"metadata,omitempty"`

	// Error message (optional)
	ErrorMessage *string `json:"error_message,omitempty"`

	// Workspace information (optional)
	WorkspaceID   *string `json:"workspace_id,omitempty"`
	WorkspacePath *string `json:"workspace_path,omitempty"`
	BranchName    *string `json:"branch_name,omitempty"`
	SourceRepo    *string `json:"source_repo,omitempty"`
	BaseBranch    *string `json:"base_branch,omitempty"`
}

// Validate checks if the update task request is valid
func (req *UpdateTaskRequest) Validate() error {
	if req.TaskID == "" {
		return fmt.Errorf("task ID is required")
	}

	if req.Status != nil && !req.Status.IsValid() {
		return fmt.Errorf("invalid task status: %s", *req.Status)
	}

	if req.Priority != nil && *req.Priority < 0 {
		return fmt.Errorf("priority must be non-negative")
	}

	if req.EstimatedMinutes != nil && *req.EstimatedMinutes < 0 {
		return fmt.Errorf("estimated minutes must be non-negative")
	}

	if req.ActualMinutes != nil && *req.ActualMinutes < 0 {
		return fmt.Errorf("actual minutes must be non-negative")
	}

	if req.EstimatedCost != nil && *req.EstimatedCost < 0 {
		return fmt.Errorf("estimated cost must be non-negative")
	}

	if req.ActualCost != nil && *req.ActualCost < 0 {
		return fmt.Errorf("actual cost must be non-negative")
	}

	return nil
}

// TaskFilter contains criteria for filtering tasks
type TaskFilter struct {
	// Filter by status
	Status []TaskStatus `json:"status,omitempty"`

	// Filter by tags
	Tags []string `json:"tags,omitempty"`

	// Filter by priority range
	MinPriority int `json:"min_priority,omitempty"`
	MaxPriority int `json:"max_priority,omitempty"`

	// Filter by creation date range
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`

	// Filter by last activity range
	LastActivityAfter  *time.Time `json:"last_activity_after,omitempty"`
	LastActivityBefore *time.Time `json:"last_activity_before,omitempty"`

	// Search in name and description
	Search string `json:"search,omitempty"`

	// Limit number of results
	Limit int `json:"limit,omitempty"`

	// Offset for pagination
	Offset int `json:"offset,omitempty"`
}

// TaskStats contains statistics about tasks
type TaskStats struct {
	// Total number of tasks
	Total int `json:"total"`

	// Number of tasks by status
	ByStatus map[TaskStatus]int `json:"by_status"`

	// Number of tasks by priority
	ByPriority map[int]int `json:"by_priority"`

	// Average completion time in minutes
	AverageCompletionMinutes float64 `json:"average_completion_minutes"`

	// Total time spent on all tasks in minutes
	TotalTimeMinutes int `json:"total_time_minutes"`

	// Total cost spent on all tasks
	TotalCost float64 `json:"total_cost"`

	// Currency used for cost tracking
	Currency string `json:"currency"`

	// Number of tasks completed today
	CompletedToday int `json:"completed_today"`

	// Number of tasks created today
	CreatedToday int `json:"created_today"`
}
