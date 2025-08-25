package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hlfshell/cowork/internal/types"
)

const (
	// TasksFileName is the name of the tasks file in the .cw directory
	TasksFileName = "tasks.json"

	// TaskNotesDirName is the name of the directory containing task notes
	TaskNotesDirName = "task-notes"

	// WorkspacesDirName is the name of the directory containing task workspaces
	WorkspacesDirName = "workspaces"
)

// Manager implements the TaskManager interface with file-based storage
type Manager struct {
	// Path to the .cw directory
	cwDir string

	// Path to the tasks file
	tasksFilePath string

	// Path to the task notes directory
	taskNotesDir string

	// Path to the workspaces directory
	workspacesDir string

	// In-memory cache of tasks
	tasks map[string]*types.Task

	// Mutex for thread safety
	mu sync.RWMutex

	// Git timeout in seconds
	gitTimeoutSeconds int
}

// NewManager creates a new task manager
func NewManager(cwDir string, gitTimeoutSeconds int) (*Manager, error) {
	if cwDir == "" {
		return nil, fmt.Errorf("cw directory path is required")
	}

	// Create the .cw directory if it doesn't exist
	if err := os.MkdirAll(cwDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .cw directory: %w", err)
	}

	manager := &Manager{
		cwDir:             cwDir,
		tasksFilePath:     filepath.Join(cwDir, TasksFileName),
		taskNotesDir:      filepath.Join(cwDir, TaskNotesDirName),
		workspacesDir:     filepath.Join(cwDir, WorkspacesDirName),
		tasks:             make(map[string]*types.Task),
		gitTimeoutSeconds: gitTimeoutSeconds,
	}

	// Create the task notes directory if it doesn't exist
	if err := os.MkdirAll(manager.taskNotesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create task notes directory: %w", err)
	}

	// Create the workspaces directory if it doesn't exist
	if err := os.MkdirAll(manager.workspacesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspaces directory: %w", err)
	}

	// Load existing tasks
	if err := manager.loadTasks(); err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	return manager, nil
}

// loadTasks loads all tasks from the tasks file
func (m *Manager) loadTasks() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if tasks file exists
	if _, err := os.Stat(m.tasksFilePath); os.IsNotExist(err) {
		// File doesn't exist, start with empty tasks
		return nil
	}

	// Read the tasks file
	file, err := os.Open(m.tasksFilePath)
	if err != nil {
		return fmt.Errorf("failed to open tasks file: %w", err)
	}
	defer file.Close()

	// Decode the tasks from JSON
	var tasks []*types.Task
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		return fmt.Errorf("failed to decode tasks: %w", err)
	}

	// Load tasks into memory
	for _, task := range tasks {
		m.tasks[task.ID] = task
	}

	return nil
}

// saveTasks saves all tasks to the tasks file
func (m *Manager) saveTasks() error {
	// Convert tasks map to slice
	tasks := make([]*types.Task, 0, len(m.tasks))

	// We need to copy the tasks to avoid holding the lock during file I/O
	m.mu.RLock()
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	m.mu.RUnlock()

	// Sort tasks by creation time (newest first)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	// Create a temporary file
	tempFile := m.tasksFilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary tasks file: %w", err)
	}
	defer file.Close()

	// Encode tasks to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(tasks); err != nil {
		return fmt.Errorf("failed to encode tasks: %w", err)
	}

	// Close the file before renaming
	file.Close()

	// Atomically rename the temporary file to the actual file
	if err := os.Rename(tempFile, m.tasksFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary tasks file: %w", err)
	}

	return nil
}

// saveTasksUnlocked saves all tasks to the tasks file without acquiring locks
// This should only be called from methods that already hold the write lock
func (m *Manager) saveTasksUnlocked() error {
	// Convert tasks map to slice
	tasks := make([]*types.Task, 0, len(m.tasks))

	// We already hold the lock, so we can access m.tasks directly
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	// Sort tasks by creation time (newest first)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	// Create a temporary file
	tempFile := m.tasksFilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary tasks file: %w", err)
	}
	defer file.Close()

	// Encode tasks to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(tasks); err != nil {
		return fmt.Errorf("failed to encode tasks: %w", err)
	}

	// Close the file before renaming
	file.Close()

	// Atomically rename the temporary file to the actual file
	if err := os.Rename(tempFile, m.tasksFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary tasks file: %w", err)
	}

	return nil
}

// CreateTask creates a new task
func (m *Manager) CreateTask(req *types.CreateTaskRequest) (*types.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create task request: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if task with this name already exists
	for _, task := range m.tasks {
		if task.Name == req.Name {
			return nil, fmt.Errorf("task with name '%s' already exists", req.Name)
		}
	}

	// Generate unique ID
	taskID := uuid.New().String()

	// Create the task
	now := time.Now()
	task := &types.Task{
		ID:               taskID,
		Name:             req.Name,
		Description:      req.Description,
		TicketID:         req.TicketID,
		URL:              req.URL,
		Status:           types.TaskStatusQueued,
		Priority:         req.Priority,
		CreatedAt:        now,
		LastActivity:     now,
		EstimatedMinutes: req.EstimatedMinutes,
		ActualMinutes:    0,
		EstimatedCost:    req.EstimatedCost,
		ActualCost:       0.0,
		Currency:         req.Currency,
		Tags:             req.Tags,
		Metadata:         req.Metadata,
	}

	// Add to memory
	m.tasks[taskID] = task

	// Save to file (we already hold the lock, so use saveTasksUnlocked)
	if err := m.saveTasksUnlocked(); err != nil {
		// Remove from memory if save failed
		delete(m.tasks, taskID)
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(taskID string) (*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// GetTaskByName retrieves a task by name
func (m *Manager) GetTaskByName(taskName string) (*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, task := range m.tasks {
		if task.Name == taskName {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %s", taskName)
}

// ListTasks returns all tasks, optionally filtered
func (m *Manager) ListTasks(filter *types.TaskFilter) ([]*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tasks []*types.Task

	// Convert tasks map to slice
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	// Apply filters if provided
	if filter != nil {
		tasks = m.applyFilters(tasks, filter)
	}

	// Sort tasks by priority (highest first), then by creation time (newest first)
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	// Apply pagination if specified
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		end := start + filter.Limit
		if start >= len(tasks) {
			tasks = []*types.Task{}
		} else if end > len(tasks) {
			tasks = tasks[start:]
		} else {
			tasks = tasks[start:end]
		}
	}

	return tasks, nil
}

// applyFilters applies the given filters to the tasks
func (m *Manager) applyFilters(tasks []*types.Task, filter *types.TaskFilter) []*types.Task {
	var filtered []*types.Task

	for _, task := range tasks {
		// Filter by status
		if len(filter.Status) > 0 {
			found := false
			for _, status := range filter.Status {
				if task.Status == status {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by tags
		if len(filter.Tags) > 0 {
			found := false
			for _, filterTag := range filter.Tags {
				for _, taskTag := range task.Tags {
					if taskTag == filterTag {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by priority range
		if filter.MinPriority > 0 && task.Priority < filter.MinPriority {
			continue
		}
		if filter.MaxPriority > 0 && task.Priority > filter.MaxPriority {
			continue
		}

		// Filter by creation date range
		if filter.CreatedAfter != nil && task.CreatedAt.Before(*filter.CreatedAfter) {
			continue
		}
		if filter.CreatedBefore != nil && task.CreatedAt.After(*filter.CreatedBefore) {
			continue
		}

		// Filter by last activity range
		if filter.LastActivityAfter != nil && task.LastActivity.Before(*filter.LastActivityAfter) {
			continue
		}
		if filter.LastActivityBefore != nil && task.LastActivity.After(*filter.LastActivityBefore) {
			continue
		}

		// Filter by search term
		if filter.Search != "" {
			searchLower := strings.ToLower(filter.Search)
			nameLower := strings.ToLower(task.Name)
			descLower := strings.ToLower(task.Description)
			if !strings.Contains(nameLower, searchLower) && !strings.Contains(descLower, searchLower) {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	return filtered
}

// UpdateTask updates a task
func (m *Manager) UpdateTask(req *types.UpdateTaskRequest) (*types.Task, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid update task request: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[req.TaskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", req.TaskID)
	}

	// Update fields if provided
	if req.Status != nil {
		oldStatus := task.Status
		task.Status = *req.Status

		// Update timestamps based on status change
		now := time.Now()
		if *req.Status == types.TaskStatusInProgress && oldStatus == types.TaskStatusQueued {
			task.StartedAt = &now
		} else if (*req.Status == types.TaskStatusCompleted || *req.Status == types.TaskStatusFailed || *req.Status == types.TaskStatusCancelled) && !oldStatus.IsTerminal() {
			task.CompletedAt = &now
		}
	}

	if req.Description != nil {
		task.Description = *req.Description
	}

	if req.Priority != nil {
		task.Priority = *req.Priority
	}

	if req.EstimatedMinutes != nil {
		task.EstimatedMinutes = *req.EstimatedMinutes
	}

	if req.ActualMinutes != nil {
		task.ActualMinutes = *req.ActualMinutes
	}

	if req.EstimatedCost != nil {
		task.EstimatedCost = *req.EstimatedCost
	}

	if req.ActualCost != nil {
		task.ActualCost = *req.ActualCost
	}

	if req.Currency != nil {
		task.Currency = *req.Currency
	}

	if req.Tags != nil {
		task.Tags = *req.Tags
	}

	if req.Metadata != nil {
		task.Metadata = *req.Metadata
	}

	if req.ErrorMessage != nil {
		task.ErrorMessage = *req.ErrorMessage
	}

	if req.WorkspaceID != nil {
		task.WorkspaceID = *req.WorkspaceID
	}

	if req.WorkspacePath != nil {
		task.WorkspacePath = *req.WorkspacePath
	}

	if req.BranchName != nil {
		task.BranchName = *req.BranchName
	}

	if req.SourceRepo != nil {
		task.SourceRepo = *req.SourceRepo
	}

	if req.BaseBranch != nil {
		task.BaseBranch = *req.BaseBranch
	}

	// Update last activity
	task.LastActivity = time.Now()

	// Save to file (we already hold the lock, so use saveTasksUnlocked)
	if err := m.saveTasksUnlocked(); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}

// DeleteTask removes a task
func (m *Manager) DeleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Remove from memory
	delete(m.tasks, taskID)

	// Save to file (we already hold the lock, so use saveTasksUnlocked)
	if err := m.saveTasksUnlocked(); err != nil {
		return fmt.Errorf("failed to save tasks after deletion: %w", err)
	}

	// Remove task workspace directory
	workspaceDir := filepath.Join(m.workspacesDir, taskID)
	if err := os.RemoveAll(workspaceDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove task workspace directory: %w", err)
	}

	return nil
}

// GetTaskStats returns statistics about tasks
func (m *Manager) GetTaskStats(filter *types.TaskFilter) (*types.TaskStats, error) {
	tasks, err := m.ListTasks(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks for stats: %w", err)
	}

	stats := &types.TaskStats{
		Total:      len(tasks),
		ByStatus:   make(map[types.TaskStatus]int),
		ByPriority: make(map[int]int),
		TotalCost:  0.0,
		Currency:   "USD", // Default currency
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var totalCompletionMinutes int
	var completedTasks int

	for _, task := range tasks {
		// Count by status
		stats.ByStatus[task.Status]++

		// Count by priority
		stats.ByPriority[task.Priority]++

		// Count completed today
		if task.Status == types.TaskStatusCompleted && task.CompletedAt != nil {
			if task.CompletedAt.After(today) {
				stats.CompletedToday++
			}
			if task.ActualMinutes > 0 {
				totalCompletionMinutes += task.ActualMinutes
				completedTasks++
			}
		}

		// Count created today
		if task.CreatedAt.After(today) {
			stats.CreatedToday++
		}

		// Sum total time
		if task.ActualMinutes > 0 {
			stats.TotalTimeMinutes += task.ActualMinutes
		}

		// Sum total cost
		if task.EstimatedCost > 0 {
			stats.TotalCost += task.EstimatedCost
			if task.Currency != "" {
				stats.Currency = task.Currency
			}
		}
	}

	// Calculate average completion time
	if completedTasks > 0 {
		stats.AverageCompletionMinutes = float64(totalCompletionMinutes) / float64(completedTasks)
	}

	return stats, nil
}

// GetNextQueuedTask returns the next task in the queue based on priority
func (m *Manager) GetNextQueuedTask() (*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var nextTask *types.Task
	var highestPriority int = -1

	for _, task := range m.tasks {
		if task.Status == types.TaskStatusQueued && task.Priority > highestPriority {
			highestPriority = task.Priority
			nextTask = task
		}
	}

	if nextTask == nil {
		return nil, fmt.Errorf("no queued tasks found")
	}

	return nextTask, nil
}

// CompleteTask marks a task as completed
func (m *Manager) CompleteTask(taskID string) error {
	status := types.TaskStatusCompleted

	req := &types.UpdateTaskRequest{
		TaskID: taskID,
		Status: &status,
	}

	_, err := m.UpdateTask(req)
	return err
}

// FailTask marks a task as failed
func (m *Manager) FailTask(taskID string, errorMessage string) error {
	status := types.TaskStatusFailed

	req := &types.UpdateTaskRequest{
		TaskID:       taskID,
		Status:       &status,
		ErrorMessage: &errorMessage,
	}

	_, err := m.UpdateTask(req)
	return err
}

// CancelTask marks a task as cancelled
func (m *Manager) CancelTask(taskID string) error {
	status := types.TaskStatusCancelled

	req := &types.UpdateTaskRequest{
		TaskID: taskID,
		Status: &status,
	}

	_, err := m.UpdateTask(req)
	return err
}

// PauseTask pauses a task
func (m *Manager) PauseTask(taskID string) error {
	status := types.TaskStatusPaused

	req := &types.UpdateTaskRequest{
		TaskID: taskID,
		Status: &status,
	}

	_, err := m.UpdateTask(req)
	return err
}

// ResumeTask resumes a paused task
func (m *Manager) ResumeTask(taskID string) error {
	status := types.TaskStatusQueued

	req := &types.UpdateTaskRequest{
		TaskID: taskID,
		Status: &status,
	}

	_, err := m.UpdateTask(req)
	return err
}

// DeleteTaskWorkspace deletes the workspace for a task
func (m *Manager) DeleteTaskWorkspace(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.WorkspacePath == "" {
		return fmt.Errorf("task '%s' has no associated workspace", task.Name)
	}

	// Remove workspace directory
	if err := os.RemoveAll(task.WorkspacePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	// Clear workspace information from task
	task.WorkspaceID = ""
	task.WorkspacePath = ""
	task.BranchName = ""
	task.SourceRepo = ""
	task.BaseBranch = ""

	// Save updated task (we already hold the lock, so use saveTasksUnlocked)
	if err := m.saveTasksUnlocked(); err != nil {
		return fmt.Errorf("failed to save task after workspace deletion: %w", err)
	}

	return nil
}
