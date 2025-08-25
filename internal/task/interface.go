package task

import "github.com/hlfshell/cowork/internal/types"

// TaskManager defines the interface for task operations
type TaskManager interface {
	// CreateTask creates a new task
	CreateTask(req *types.CreateTaskRequest) (*types.Task, error)

	// GetTask retrieves a task by ID
	GetTask(taskID string) (*types.Task, error)

	// GetTaskByName retrieves a task by name
	GetTaskByName(taskName string) (*types.Task, error)

	// ListTasks returns all tasks, optionally filtered
	ListTasks(filter *types.TaskFilter) ([]*types.Task, error)

	// UpdateTask updates a task
	UpdateTask(req *types.UpdateTaskRequest) (*types.Task, error)

	// DeleteTask removes a task
	DeleteTask(taskID string) error

	// GetTaskStats returns statistics about tasks
	GetTaskStats(filter *types.TaskFilter) (*types.TaskStats, error)

	// GetNextQueuedTask returns the next task in the queue based on priority
	GetNextQueuedTask() (*types.Task, error)

	// CompleteTask marks a task as completed
	CompleteTask(taskID string) error

	// FailTask marks a task as failed
	FailTask(taskID string, errorMessage string) error

	// CancelTask marks a task as cancelled
	CancelTask(taskID string) error

	// PauseTask pauses a task
	PauseTask(taskID string) error

	// ResumeTask resumes a paused task
	ResumeTask(taskID string) error

	// CreateWorkspaceForTask creates a workspace for a task
	CreateWorkspaceForTask(task *types.Task, description string) error

	// GetTaskWorkspacePath returns the workspace path for a task
	GetTaskWorkspacePath(taskID string) (string, error)

	// RunGitInTaskWorkspace runs git commands in a task's workspace
	RunGitInTaskWorkspace(taskID string, gitArgs []string) error

	// DeleteTaskWorkspace deletes the workspace for a task
	DeleteTaskWorkspace(taskID string) error
}
