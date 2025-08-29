package task

import "github.com/hlfshell/cowork/internal/types"

// TaskManager defines the interface for task management operations
type TaskManager interface {
	// CreateTask creates a new task
	CreateTask(req *types.CreateTaskRequest) (*types.Task, error)

	// GetTask retrieves a task by ID
	GetTask(taskID string) (*types.Task, error)

	// GetTaskByName retrieves a task by name
	GetTaskByName(taskName string) (*types.Task, error)

	// ListTasks returns all tasks, optionally filtered
	ListTasks(filter *types.TaskFilter) ([]*types.Task, error)

	// UpdateTask updates an existing task
	UpdateTask(req *types.UpdateTaskRequest) (*types.Task, error)

	// DeleteTask removes a task
	DeleteTask(taskID string) error

	// GetTaskStats returns statistics about tasks
	GetTaskStats() (*types.TaskStats, error)

	// GetNextQueuedTask returns the next task in the queue based on priority
	GetNextQueuedTask() (*types.Task, error)

	// Task status management
	CompleteTask(taskID string) error
	FailTask(taskID string, errorMessage string) error
	CancelTask(taskID string) error
	PauseTask(taskID string) error
	ResumeTask(taskID string) error

	// Workspace management for tasks
	CreateTaskWorkspace(taskID string, req *types.CreateWorkspaceRequest) (*types.Workspace, error)
	CreateWorkspaceForTask(taskID string, req *types.CreateWorkspaceRequest) (*types.Workspace, error)
	GetTaskWorkspace(taskID string) (*types.Workspace, error)
	GetTaskWorkspacePath(taskID string) (string, error)
	RunGitInTaskWorkspace(taskID string, gitArgs []string) error
	DeleteTaskWorkspace(taskID string) error
}
