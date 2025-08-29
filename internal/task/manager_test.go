package task

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hlfshell/cowork/internal/types"
)

// TestNewManager_WithValidDirectory tests creating a new task manager with a valid directory
func TestNewManager_WithValidDirectory(t *testing.T) {
	// Test case: Creating a task manager with a valid directory should succeed
	// and create the necessary directory structure
	tempDir := t.TempDir()
	gitTimeoutSeconds := 30

	manager, err := NewManager(tempDir, gitTimeoutSeconds)

	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, tempDir, manager.cwDir)
	assert.Equal(t, gitTimeoutSeconds, manager.gitTimeoutSeconds)

	// Verify directory structure was created
	assert.DirExists(t, filepath.Join(tempDir, TaskNotesDirName))
	assert.DirExists(t, filepath.Join(tempDir, WorkspacesDirName))
	// Tasks file is created when first task is saved, not on initialization
}

// TestNewManager_WithEmptyDirectory tests creating a task manager with an empty directory path
func TestNewManager_WithEmptyDirectory(t *testing.T) {
	// Test case: Creating a task manager with an empty directory should fail
	// with a descriptive error message
	gitTimeoutSeconds := 30

	manager, err := NewManager("", gitTimeoutSeconds)

	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "cw directory path is required")
}

// TestNewManager_WithInvalidDirectory tests creating a task manager with an invalid directory
func TestNewManager_WithInvalidDirectory(t *testing.T) {
	// Test case: Creating a task manager with an invalid directory should fail
	// when the system cannot create the directory structure
	invalidPath := "/invalid/path/that/should/not/exist/and/cannot/be/created"
	gitTimeoutSeconds := 30

	manager, err := NewManager(invalidPath, gitTimeoutSeconds)

	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "failed to create .cw directory")
}

// TestCreateTask_WithValidRequest tests creating a task with valid parameters
func TestCreateTask_WithValidRequest(t *testing.T) {
	// Test case: Creating a task with valid parameters should succeed
	// and return a task with the expected properties
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	request := &types.CreateTaskRequest{
		Name:        "test-task",
		Description: "A test task for unit testing",
		Priority:    5,
		Tags:        []string{"test", "unit"},
		Metadata: map[string]string{
			"test_key": "test_value",
		},
	}

	task, err := manager.CreateTask(request)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, request.Name, task.Name)
	assert.Equal(t, request.Description, task.Description)
	assert.Equal(t, request.Priority, task.Priority)
	assert.Equal(t, request.Tags, task.Tags)
	assert.Equal(t, request.Metadata, task.Metadata)
	assert.Equal(t, types.TaskStatusQueued, task.Status)
	assert.NotEmpty(t, task.ID)
	assert.NotZero(t, task.CreatedAt)
	assert.NotZero(t, task.LastActivity)
}

// TestCreateTask_WithInvalidRequest tests creating a task with invalid parameters
func TestCreateTask_WithInvalidRequest(t *testing.T) {
	// Test case: Creating a task with invalid parameters should fail
	// with appropriate validation errors
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		request       *types.CreateTaskRequest
		errorContains string
	}{
		{
			name: "Empty task name",
			request: &types.CreateTaskRequest{
				Name:     "",
				Priority: 1,
			},
			errorContains: "task name is required",
		},
		{
			name: "Negative priority",
			request: &types.CreateTaskRequest{
				Name:     "test-task",
				Priority: -1,
			},
			errorContains: "priority must be non-negative",
		},
		{
			name: "Negative estimated minutes",
			request: &types.CreateTaskRequest{
				Name:             "test-task",
				Priority:         1,
				EstimatedMinutes: -10,
			},
			errorContains: "estimated minutes must be non-negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			task, err := manager.CreateTask(tc.request)

			assert.Error(t, err)
			assert.Nil(t, task)
			assert.Contains(t, err.Error(), tc.errorContains)
		})
	}
}

// TestGetTask_WithValidID tests retrieving a task by its ID
func TestGetTask_WithValidID(t *testing.T) {
	// Test case: Retrieving a task by valid ID should return the correct task
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create a task first
	request := &types.CreateTaskRequest{
		Name:        "test-task",
		Description: "A test task",
		Priority:    3,
	}
	createdTask, err := manager.CreateTask(request)
	require.NoError(t, err)

	// Retrieve the task
	retrievedTask, err := manager.GetTask(fmt.Sprintf("%d", createdTask.ID))

	assert.NoError(t, err)
	assert.NotNil(t, retrievedTask)
	assert.Equal(t, createdTask.ID, retrievedTask.ID)
	assert.Equal(t, createdTask.Name, retrievedTask.Name)
	assert.Equal(t, createdTask.Description, retrievedTask.Description)
}

// TestGetTask_WithInvalidID tests retrieving a task with an invalid ID
func TestGetTask_WithInvalidID(t *testing.T) {
	// Test case: Retrieving a task with invalid ID should return an error
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	task, err := manager.GetTask("invalid-id")

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "task not found")
}

// TestGetTaskByName_WithValidName tests retrieving a task by its name
func TestGetTaskByName_WithValidName(t *testing.T) {
	// Test case: Retrieving a task by valid name should return the correct task
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create a task first
	request := &types.CreateTaskRequest{
		Name:        "unique-task-name",
		Description: "A test task",
		Priority:    3,
	}
	createdTask, err := manager.CreateTask(request)
	require.NoError(t, err)

	// Retrieve the task by name
	retrievedTask, err := manager.GetTaskByName(createdTask.Name)

	assert.NoError(t, err)
	assert.NotNil(t, retrievedTask)
	assert.Equal(t, createdTask.ID, retrievedTask.ID)
	assert.Equal(t, createdTask.Name, retrievedTask.Name)
}

// TestGetTaskByName_WithInvalidName tests retrieving a task with an invalid name
func TestGetTaskByName_WithInvalidName(t *testing.T) {
	// Test case: Retrieving a task with invalid name should return an error
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	task, err := manager.GetTaskByName("non-existent-task")

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "task not found")
}

// TestListTasks_WithNoFilters tests listing all tasks without filters
func TestListTasks_WithNoFilters(t *testing.T) {
	// Test case: Listing tasks without filters should return all tasks
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create multiple tasks
	taskNames := []string{"task-1", "task-2", "task-3"}
	for _, name := range taskNames {
		request := &types.CreateTaskRequest{
			Name:        name,
			Description: "Test task",
			Priority:    1,
		}
		_, err := manager.CreateTask(request)
		require.NoError(t, err)
	}

	// List all tasks
	tasks, err := manager.ListTasks(nil)

	assert.NoError(t, err)
	assert.Len(t, tasks, 3)

	// Verify all tasks are present
	taskNamesMap := make(map[string]bool)
	for _, task := range tasks {
		taskNamesMap[task.Name] = true
	}
	for _, name := range taskNames {
		assert.True(t, taskNamesMap[name], "Task %s should be in the list", name)
	}
}

// TestDeleteTask_WithValidID tests deleting a task by its ID
func TestDeleteTask_WithValidID(t *testing.T) {
	// Test case: Deleting a task by valid ID should succeed
	// and remove the task from the system
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create a task first
	request := &types.CreateTaskRequest{
		Name:        "task-to-delete",
		Description: "A task that will be deleted",
		Priority:    1,
	}
	task, err := manager.CreateTask(request)
	require.NoError(t, err)

	// Verify task exists
	_, err = manager.GetTask(fmt.Sprintf("%d", task.ID))
	require.NoError(t, err)

	// Delete the task
	err = manager.DeleteTask(fmt.Sprintf("%d", task.ID))

	assert.NoError(t, err)

	// Verify task no longer exists
	_, err = manager.GetTask(fmt.Sprintf("%d", task.ID))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

// TestDeleteTask_WithInvalidID tests deleting a task with an invalid ID
func TestDeleteTask_WithInvalidID(t *testing.T) {
	// Test case: Deleting a task with invalid ID should return an error
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	err = manager.DeleteTask("invalid-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

// TestCompleteTask_WithValidID tests completing a task
func TestCompleteTask_WithValidID(t *testing.T) {
	// Test case: Completing a task should change its status to completed
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create a task first
	request := &types.CreateTaskRequest{
		Name:        "task-to-complete",
		Description: "A task that will be completed",
		Priority:    1,
	}
	task, err := manager.CreateTask(request)
	require.NoError(t, err)

	// Complete the task
	err = manager.CompleteTask(fmt.Sprintf("%d", task.ID))

	assert.NoError(t, err)

	// Verify task status changed
	completedTask, err := manager.GetTask(fmt.Sprintf("%d", task.ID))
	require.NoError(t, err)
	assert.Equal(t, types.TaskStatusCompleted, completedTask.Status)
	assert.NotZero(t, completedTask.CompletedAt)
}

// TestFailTask_WithValidID tests failing a task
func TestFailTask_WithValidID(t *testing.T) {
	// Test case: Failing a task should change its status to failed
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create a task first
	request := &types.CreateTaskRequest{
		Name:        "task-to-fail",
		Description: "A task that will be failed",
		Priority:    1,
	}
	task, err := manager.CreateTask(request)
	require.NoError(t, err)

	// Fail the task
	err = manager.FailTask(fmt.Sprintf("%d", task.ID), "Test failure")

	assert.NoError(t, err)

	// Verify task status changed
	failedTask, err := manager.GetTask(fmt.Sprintf("%d", task.ID))
	require.NoError(t, err)
	assert.Equal(t, types.TaskStatusFailed, failedTask.Status)
	assert.NotZero(t, failedTask.CompletedAt)
}

// TestCancelTask_WithValidID tests cancelling a task
func TestCancelTask_WithValidID(t *testing.T) {
	// Test case: Cancelling a task should change its status to cancelled
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create a task first
	request := &types.CreateTaskRequest{
		Name:        "task-to-cancel",
		Description: "A task that will be cancelled",
		Priority:    1,
	}
	task, err := manager.CreateTask(request)
	require.NoError(t, err)

	// Cancel the task
	err = manager.CancelTask(fmt.Sprintf("%d", task.ID))

	assert.NoError(t, err)

	// Verify task status changed
	cancelledTask, err := manager.GetTask(fmt.Sprintf("%d", task.ID))
	require.NoError(t, err)
	assert.Equal(t, types.TaskStatusCancelled, cancelledTask.Status)
	assert.NotZero(t, cancelledTask.CompletedAt)
}

// TestPauseAndResumeTask tests pausing and resuming a task
func TestPauseAndResumeTask(t *testing.T) {
	// Test case: Pausing and resuming a task should change its status appropriately
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create a task first
	request := &types.CreateTaskRequest{
		Name:        "task-to-pause",
		Description: "A task that will be paused and resumed",
		Priority:    1,
	}
	task, err := manager.CreateTask(request)
	require.NoError(t, err)

	// Pause the task
	err = manager.PauseTask(fmt.Sprintf("%d", task.ID))
	assert.NoError(t, err)

	// Verify task is paused
	pausedTask, err := manager.GetTask(fmt.Sprintf("%d", task.ID))
	require.NoError(t, err)
	assert.Equal(t, types.TaskStatusPaused, pausedTask.Status)

	// Resume the task
	err = manager.ResumeTask(fmt.Sprintf("%d", task.ID))
	assert.NoError(t, err)

	// Verify task is resumed (should be back to queued)
	resumedTask, err := manager.GetTask(fmt.Sprintf("%d", task.ID))
	require.NoError(t, err)
	assert.Equal(t, types.TaskStatusQueued, resumedTask.Status)
}

// TestGetNextQueuedTask_WithMultipleTasks tests getting the next queued task
func TestGetNextQueuedTask_WithMultipleTasks(t *testing.T) {
	// Test case: Getting the next queued task should return the highest priority task
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create tasks with different priorities
	_, err = manager.CreateTask(&types.CreateTaskRequest{
		Name:     "low-priority",
		Priority: 1,
	})
	require.NoError(t, err)

	highPriorityTask, err := manager.CreateTask(&types.CreateTaskRequest{
		Name:     "high-priority",
		Priority: 10,
	})
	require.NoError(t, err)

	_, err = manager.CreateTask(&types.CreateTaskRequest{
		Name:     "medium-priority",
		Priority: 5,
	})
	require.NoError(t, err)

	// Get next queued task
	nextTask, err := manager.GetNextQueuedTask()

	assert.NoError(t, err)
	assert.NotNil(t, nextTask)
	assert.Equal(t, highPriorityTask.ID, nextTask.ID)
	assert.Equal(t, types.TaskStatusQueued, nextTask.Status)
}

// TestGetNextQueuedTask_WithNoQueuedTasks tests getting next queued task when none exist
func TestGetNextQueuedTask_WithNoQueuedTasks(t *testing.T) {
	// Test case: Getting the next queued task when no queued tasks exist
	// should return an error
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Create a task and complete it
	task, err := manager.CreateTask(&types.CreateTaskRequest{
		Name:     "completed-task",
		Priority: 1,
	})
	require.NoError(t, err)

	err = manager.CompleteTask(fmt.Sprintf("%d", task.ID))
	require.NoError(t, err)

	// Try to get next queued task
	nextTask, err := manager.GetNextQueuedTask()

	assert.Error(t, err)
	assert.Nil(t, nextTask)
	assert.Contains(t, err.Error(), "no queued tasks found")
}

// TestTaskPersistence tests that tasks persist across manager instances
func TestTaskPersistence(t *testing.T) {
	// Test case: Tasks should persist when a new manager instance is created
	// for the same directory
	tempDir := t.TempDir()

	// Create first manager and add a task
	manager1, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	task, err := manager1.CreateTask(&types.CreateTaskRequest{
		Name:        "persistent-task",
		Description: "A task that should persist",
		Priority:    5,
	})
	require.NoError(t, err)

	// Create second manager for the same directory
	manager2, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Verify task exists in second manager
	retrievedTask, err := manager2.GetTask(fmt.Sprintf("%d", task.ID))
	assert.NoError(t, err)
	assert.Equal(t, task.ID, retrievedTask.ID)
	assert.Equal(t, task.Name, retrievedTask.Name)
	assert.Equal(t, task.Description, retrievedTask.Description)
}

// TestConcurrentTaskCreation tests creating tasks concurrently
func TestConcurrentTaskCreation(t *testing.T) {
	// Test case: Creating tasks concurrently should work correctly
	// without race conditions
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	const numTasks = 10
	results := make(chan error, numTasks)

	// Create tasks concurrently
	for i := 0; i < numTasks; i++ {
		go func(index int) {
			request := &types.CreateTaskRequest{
				Name:        fmt.Sprintf("concurrent-task-%d", index),
				Description: fmt.Sprintf("Task created concurrently %d", index),
				Priority:    index,
			}
			_, err := manager.CreateTask(request)
			results <- err
		}(i)
	}

	// Wait for all tasks to be created
	for i := 0; i < numTasks; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Verify all tasks were created
	tasks, err := manager.ListTasks(nil)
	assert.NoError(t, err)
	assert.Len(t, tasks, numTasks)
}

// TestTaskValidation tests task validation rules
func TestTaskValidation(t *testing.T) {
	// Test case: Task validation should enforce business rules
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir, 30)
	require.NoError(t, err)

	// Test that task names must be unique
	request1 := &types.CreateTaskRequest{
		Name:     "duplicate-name",
		Priority: 1,
	}
	task1, err := manager.CreateTask(request1)
	require.NoError(t, err)
	assert.NotNil(t, task1)

	request2 := &types.CreateTaskRequest{
		Name:     "duplicate-name", // Same name
		Priority: 1,
	}
	task2, err := manager.CreateTask(request2)
	assert.Error(t, err)
	assert.Nil(t, task2)
	assert.Contains(t, err.Error(), "task with name")
}
