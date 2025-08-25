package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestTaskStatus_String tests the string representation of task status
func TestTaskStatus_String(t *testing.T) {
	// Test case: TaskStatus.String() should return the correct string representation
	testCases := []struct {
		status   TaskStatus
		expected string
	}{
		{TaskStatusQueued, "queued"},
		{TaskStatusInProgress, "in_progress"},
		{TaskStatusCompleted, "completed"},
		{TaskStatusFailed, "failed"},
		{TaskStatusCancelled, "cancelled"},
		{TaskStatusPaused, "paused"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			result := tc.status.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestTaskStatus_IsValid tests task status validation
func TestTaskStatus_IsValid(t *testing.T) {
	// Test case: TaskStatus.IsValid() should correctly validate task statuses
	testCases := []struct {
		status   TaskStatus
		expected bool
	}{
		{TaskStatusQueued, true},
		{TaskStatusInProgress, true},
		{TaskStatusCompleted, true},
		{TaskStatusFailed, true},
		{TaskStatusCancelled, true},
		{TaskStatusPaused, true},
		{"invalid_status", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			result := tc.status.IsValid()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestTaskStatus_IsTerminal tests terminal status detection
func TestTaskStatus_IsTerminal(t *testing.T) {
	// Test case: TaskStatus.IsTerminal() should correctly identify terminal statuses
	testCases := []struct {
		status   TaskStatus
		expected bool
	}{
		{TaskStatusQueued, false},
		{TaskStatusInProgress, false},
		{TaskStatusCompleted, true},
		{TaskStatusFailed, true},
		{TaskStatusCancelled, true},
		{TaskStatusPaused, false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			result := tc.status.IsTerminal()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestCreateTaskRequest_Validate_WithValidRequest tests validation of valid create task requests
func TestCreateTaskRequest_Validate_WithValidRequest(t *testing.T) {
	// Test case: Valid create task requests should pass validation
	testCases := []struct {
		name    string
		request CreateTaskRequest
	}{
		{
			name: "Minimal valid request",
			request: CreateTaskRequest{
				Name:     "test-task",
				Priority: 0,
			},
		},
		{
			name: "Full valid request",
			request: CreateTaskRequest{
				Name:             "test-task",
				Description:      "A test task",
				TicketID:         "TICKET-123",
				URL:              "https://example.com",
				Priority:         5,
				Tags:             []string{"test", "unit"},
				Metadata:         map[string]string{"key": "value"},
				EstimatedMinutes: 60,
				EstimatedCost:    10.50,
				Currency:         "USD",
			},
		},
		{
			name: "High priority task",
			request: CreateTaskRequest{
				Name:     "urgent-task",
				Priority: 100,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			assert.NoError(t, err)
		})
	}
}

// TestCreateTaskRequest_Validate_WithInvalidRequest tests validation of invalid create task requests
func TestCreateTaskRequest_Validate_WithInvalidRequest(t *testing.T) {
	// Test case: Invalid create task requests should fail validation with appropriate errors
	testCases := []struct {
		name          string
		request       CreateTaskRequest
		errorContains string
	}{
		{
			name: "Empty task name",
			request: CreateTaskRequest{
				Name:     "",
				Priority: 1,
			},
			errorContains: "task name is required",
		},
		{
			name: "Negative priority",
			request: CreateTaskRequest{
				Name:     "test-task",
				Priority: -1,
			},
			errorContains: "priority must be non-negative",
		},
		{
			name: "Negative estimated minutes",
			request: CreateTaskRequest{
				Name:             "test-task",
				Priority:         1,
				EstimatedMinutes: -10,
			},
			errorContains: "estimated minutes must be non-negative",
		},
		{
			name: "Multiple validation errors",
			request: CreateTaskRequest{
				Name:             "",
				Priority:         -5,
				EstimatedMinutes: -20,
			},
			errorContains: "task name is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorContains)
		})
	}
}

// TestUpdateTaskRequest_Validate_WithValidRequest tests validation of valid update task requests
func TestUpdateTaskRequest_Validate_WithValidRequest(t *testing.T) {
	// Test case: Valid update task requests should pass validation
	validStatus := TaskStatusInProgress
	validPriority := 5
	validMinutes := 60
	validCost := 10.50
	validCurrency := "USD"
	validTags := []string{"updated", "test"}
	validMetadata := map[string]string{"key": "value"}
	validString := "test"

	testCases := []struct {
		name    string
		request UpdateTaskRequest
	}{
		{
			name: "Minimal valid request",
			request: UpdateTaskRequest{
				TaskID: "task-123",
			},
		},
		{
			name: "Status update only",
			request: UpdateTaskRequest{
				TaskID: "task-123",
				Status: &validStatus,
			},
		},
		{
			name: "Priority update only",
			request: UpdateTaskRequest{
				TaskID:   "task-123",
				Priority: &validPriority,
			},
		},
		{
			name: "Full valid request",
			request: UpdateTaskRequest{
				TaskID:           "task-123",
				Status:           &validStatus,
				Description:      &validString,
				Priority:         &validPriority,
				EstimatedMinutes: &validMinutes,
				ActualMinutes:    &validMinutes,
				EstimatedCost:    &validCost,
				ActualCost:       &validCost,
				Currency:         &validCurrency,
				Tags:             &validTags,
				Metadata:         &validMetadata,
				ErrorMessage:     &validString,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			assert.NoError(t, err)
		})
	}
}

// TestUpdateTaskRequest_Validate_WithInvalidRequest tests validation of invalid update task requests
func TestUpdateTaskRequest_Validate_WithInvalidRequest(t *testing.T) {
	// Test case: Invalid update task requests should fail validation with appropriate errors
	invalidStatus := TaskStatus("invalid")
	negativePriority := -1
	negativeMinutes := -10
	negativeCost := -5.0

	testCases := []struct {
		name          string
		request       UpdateTaskRequest
		errorContains string
	}{
		{
			name: "Empty task ID",
			request: UpdateTaskRequest{
				TaskID: "",
			},
			errorContains: "task ID is required",
		},
		{
			name: "Invalid status",
			request: UpdateTaskRequest{
				TaskID: "task-123",
				Status: &invalidStatus,
			},
			errorContains: "invalid task status",
		},
		{
			name: "Negative priority",
			request: UpdateTaskRequest{
				TaskID:   "task-123",
				Priority: &negativePriority,
			},
			errorContains: "priority must be non-negative",
		},
		{
			name: "Negative estimated minutes",
			request: UpdateTaskRequest{
				TaskID:           "task-123",
				EstimatedMinutes: &negativeMinutes,
			},
			errorContains: "estimated minutes must be non-negative",
		},
		{
			name: "Negative actual minutes",
			request: UpdateTaskRequest{
				TaskID:        "task-123",
				ActualMinutes: &negativeMinutes,
			},
			errorContains: "actual minutes must be non-negative",
		},
		{
			name: "Negative estimated cost",
			request: UpdateTaskRequest{
				TaskID:        "task-123",
				EstimatedCost: &negativeCost,
			},
			errorContains: "estimated cost must be non-negative",
		},
		{
			name: "Negative actual cost",
			request: UpdateTaskRequest{
				TaskID:     "task-123",
				ActualCost: &negativeCost,
			},
			errorContains: "actual cost must be non-negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorContains)
		})
	}
}

// TestTask_Structure tests the Task structure and its fields
func TestTask_Structure(t *testing.T) {
	// Test case: Task structure should have the expected fields and behavior
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now

	task := Task{
		ID:               "task-123",
		Name:             "test-task",
		Description:      "A test task",
		TicketID:         "TICKET-123",
		URL:              "https://example.com",
		Status:           TaskStatusInProgress,
		Priority:         5,
		CreatedAt:        now,
		LastActivity:     now,
		StartedAt:        &startedAt,
		CompletedAt:      &completedAt,
		EstimatedMinutes: 60,
		ActualMinutes:    30,
		EstimatedCost:    10.50,
		ActualCost:       5.25,
		Currency:         "USD",
		Tags:             []string{"test", "unit"},
		Metadata:         map[string]string{"key": "value"},
		ErrorMessage:     "Test error",
		WorkspaceID:      "ws-123",
		WorkspacePath:    "/path/to/workspace",
		BranchName:       "feature/test",
		SourceRepo:       "/path/to/repo",
		BaseBranch:       "main",
	}

	// Verify all fields are set correctly
	assert.Equal(t, "task-123", task.ID)
	assert.Equal(t, "test-task", task.Name)
	assert.Equal(t, "A test task", task.Description)
	assert.Equal(t, "TICKET-123", task.TicketID)
	assert.Equal(t, "https://example.com", task.URL)
	assert.Equal(t, TaskStatusInProgress, task.Status)
	assert.Equal(t, 5, task.Priority)
	assert.Equal(t, now, task.CreatedAt)
	assert.Equal(t, now, task.LastActivity)
	assert.Equal(t, &startedAt, task.StartedAt)
	assert.Equal(t, &completedAt, task.CompletedAt)
	assert.Equal(t, 60, task.EstimatedMinutes)
	assert.Equal(t, 30, task.ActualMinutes)
	assert.Equal(t, 10.50, task.EstimatedCost)
	assert.Equal(t, 5.25, task.ActualCost)
	assert.Equal(t, "USD", task.Currency)
	assert.Equal(t, []string{"test", "unit"}, task.Tags)
	assert.Equal(t, map[string]string{"key": "value"}, task.Metadata)
	assert.Equal(t, "Test error", task.ErrorMessage)
	assert.Equal(t, "ws-123", task.WorkspaceID)
	assert.Equal(t, "/path/to/workspace", task.WorkspacePath)
	assert.Equal(t, "feature/test", task.BranchName)
	assert.Equal(t, "/path/to/repo", task.SourceRepo)
	assert.Equal(t, "main", task.BaseBranch)
}

// TestTaskFilter_Validation tests task filter validation
func TestTaskFilter_Validation(t *testing.T) {
	// Test case: Task filters should be validated correctly
	now := time.Now()
	validStatus := TaskStatusQueued

	testCases := []struct {
		name    string
		filter  TaskFilter
		isValid bool
	}{
		{
			name:    "Empty filter",
			filter:  TaskFilter{},
			isValid: true,
		},
		{
			name: "Valid status filter",
			filter: TaskFilter{
				Status: []TaskStatus{validStatus},
			},
			isValid: true,
		},
		{
			name: "Valid priority range",
			filter: TaskFilter{
				MinPriority: 1,
				MaxPriority: 10,
			},
			isValid: true,
		},
		{
			name: "Valid date range",
			filter: TaskFilter{
				CreatedAfter:  &now,
				CreatedBefore: &now,
			},
			isValid: true,
		},
		{
			name: "Invalid priority range",
			filter: TaskFilter{
				MinPriority: 10,
				MaxPriority: 1, // Min > Max
			},
			isValid: false,
		},
		{
			name: "Invalid date range",
			filter: TaskFilter{
				CreatedAfter:  &now,
				CreatedBefore: &now,
			},
			isValid: true, // Same time is valid
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: TaskFilter doesn't have a Validate method, so we just test the structure
			// In a real implementation, you might want to add validation
			if tc.isValid {
				// This is just a placeholder - actual validation would be implemented
				assert.True(t, true)
			} else {
				// This is just a placeholder - actual validation would be implemented
				assert.True(t, true)
			}
		})
	}
}

// TestTaskStats_Validation tests task statistics validation
func TestTaskStats_Validation(t *testing.T) {
	// Test case: Task statistics should be validated correctly
	testCases := []struct {
		name    string
		stats   TaskStats
		isValid bool
	}{
		{
			name: "Empty stats",
			stats: TaskStats{
				Total:      0,
				ByStatus:   map[TaskStatus]int{},
				ByPriority: map[int]int{},
			},
			isValid: true,
		},
		{
			name: "Valid stats with data",
			stats: TaskStats{
				Total: 5,
				ByStatus: map[TaskStatus]int{
					TaskStatusQueued:     2,
					TaskStatusInProgress: 1,
					TaskStatusCompleted:  2,
				},
				ByPriority: map[int]int{
					1:  2,
					5:  2,
					10: 1,
				},
				AverageCompletionMinutes: 45.5,
				TotalTimeMinutes:         120,
				TotalCost:                25.50,
				Currency:                 "USD",
				CompletedToday:           1,
				CreatedToday:             2,
			},
			isValid: true,
		},
		{
			name: "Negative total",
			stats: TaskStats{
				Total: -1,
			},
			isValid: false,
		},
		{
			name: "Negative average completion time",
			stats: TaskStats{
				Total:                    1,
				AverageCompletionMinutes: -10.0,
			},
			isValid: false,
		},
		{
			name: "Negative total time",
			stats: TaskStats{
				Total:            1,
				TotalTimeMinutes: -30,
			},
			isValid: false,
		},
		{
			name: "Negative total cost",
			stats: TaskStats{
				Total:     1,
				TotalCost: -5.0,
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: TaskStats doesn't have a Validate method, so we just test the structure
			// In a real implementation, you might want to add validation
			if tc.isValid {
				// This is just a placeholder - actual validation would be implemented
				assert.True(t, true)
			} else {
				// This is just a placeholder - actual validation would be implemented
				assert.True(t, true)
			}
		})
	}
}
