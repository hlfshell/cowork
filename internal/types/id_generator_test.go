package types

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDGenerator(t *testing.T) {
	// Reset generators for clean test
	ResetIDGenerators()

	t.Run("GenerateTaskID", func(t *testing.T) {
		// Test that task IDs are sequential integers
		id1 := GenerateTaskID()
		id2 := GenerateTaskID()
		id3 := GenerateTaskID()

		expected1 := 1
		expected2 := 2
		expected3 := 3

		if id1 != expected1 {
			t.Errorf("Expected %d, got %d", expected1, id1)
		}
		if id2 != expected2 {
			t.Errorf("Expected %d, got %d", expected2, id2)
		}
		if id3 != expected3 {
			t.Errorf("Expected %d, got %d", expected3, id3)
		}
	})

	t.Run("GenerateWorkspaceID", func(t *testing.T) {
		// Test that standalone workspace IDs start from 1000
		id1 := GenerateWorkspaceID()
		id2 := GenerateWorkspaceID()
		id3 := GenerateWorkspaceID()

		expected1 := 1000 // Standalone workspaces start from 1000
		expected2 := 1001
		expected3 := 1002

		if id1 != expected1 {
			t.Errorf("Expected %d, got %d", expected1, id1)
		}
		if id2 != expected2 {
			t.Errorf("Expected %d, got %d", expected2, id2)
		}
		if id3 != expected3 {
			t.Errorf("Expected %d, got %d", expected3, id3)
		}
	})

	t.Run("GenerateWorkflowID", func(t *testing.T) {
		// Test that workflow IDs are sequential integers
		id1 := GenerateWorkflowID()
		id2 := GenerateWorkflowID()
		id3 := GenerateWorkflowID()

		expected1 := 4 // After task-1, task-2, task-3 (workspaces now start from 1000)
		expected2 := 5
		expected3 := 6

		if id1 != expected1 {
			t.Errorf("Expected %d, got %d", expected1, id1)
		}
		if id2 != expected2 {
			t.Errorf("Expected %d, got %d", expected2, id2)
		}
		if id3 != expected3 {
			t.Errorf("Expected %d, got %d", expected3, id3)
		}
	})

	t.Run("ResetIDGenerators", func(t *testing.T) {
		// Test that reset works correctly
		_ = GenerateTaskID()
		_ = GenerateWorkspaceID()
		_ = GenerateWorkflowID()

		ResetIDGenerators()

		newTaskID := GenerateTaskID()
		newWsID := GenerateWorkspaceID()
		newWfID := GenerateWorkflowID()

		if newTaskID != 1 {
			t.Errorf("After reset, task ID should be 1, got: %d", newTaskID)
		}
		if newWsID != 1000 {
			t.Errorf("After reset, workspace ID should be 1000, got: %d", newWsID)
		}
		if newWfID != 2 {
			t.Errorf("After reset, workflow ID should be 2, got: %d", newWfID)
		}
	})
}

// TestIDGenerator_LinkedTaskWorkspaceIDs tests that tasks and standalone workspaces have separate sequences
func TestIDGenerator_LinkedTaskWorkspaceIDs(t *testing.T) {
	// Test case: Task and standalone workspace IDs should have separate sequences
	// Reset generators to ensure clean state
	ResetIDGenerators()

	// Generate task and workspace IDs in sequence
	taskID1 := GenerateTaskID()
	workspaceID1 := GenerateWorkspaceID()
	taskID2 := GenerateTaskID()
	workspaceID2 := GenerateWorkspaceID()

	// Verify they have separate sequences
	assert.Equal(t, 1, taskID1)
	assert.Equal(t, 1000, workspaceID1) // Standalone workspaces start from 1000
	assert.Equal(t, 2, taskID2)
	assert.Equal(t, 1001, workspaceID2)

	// Verify the sequence continues correctly
	taskID3 := GenerateTaskID()
	workspaceID3 := GenerateWorkspaceID()
	assert.Equal(t, 3, taskID3)
	assert.Equal(t, 1002, workspaceID3)
}

// TestIDGenerator_NoReuse tests that IDs are not reused
func TestIDGenerator_NoReuse(t *testing.T) {
	// Test case: IDs should not be reused when generators are reset
	ResetIDGenerators()

	// Generate some IDs
	_ = GenerateTaskID()                  // Generate first task ID
	_ = GenerateWorkspaceID()             // Generate first workspace ID
	taskID2 := GenerateTaskID()           // Generate another task ID to advance sequence
	workspaceID2 := GenerateWorkspaceID() // Generate another workspace ID

	// Reset and generate new IDs
	ResetIDGenerators()
	taskID3 := GenerateTaskID()
	workspaceID3 := GenerateWorkspaceID()

	// Verify new IDs start from 1 again (standalone workspaces start from 1000)
	assert.Equal(t, 1, taskID3)
	assert.Equal(t, 1000, workspaceID3)

	// Verify old IDs are different from new IDs
	assert.NotEqual(t, taskID2, taskID3)           // taskID2 should be different from taskID3
	assert.NotEqual(t, workspaceID2, workspaceID3) // workspaceID2 should be different from workspaceID3
}

// TestIDGenerator_ConcurrentAccess tests concurrent access to ID generators
func TestIDGenerator_ConcurrentAccess(t *testing.T) {
	// Test case: ID generators should be thread-safe
	ResetIDGenerators()

	const numGoroutines = 10
	const idsPerGoroutine = 5

	// Channel to collect IDs
	taskIDs := make(chan int, numGoroutines*idsPerGoroutine)
	workspaceIDs := make(chan int, numGoroutines*idsPerGoroutine)

	// Start goroutines to generate IDs concurrently
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				taskIDs <- GenerateTaskID()
				workspaceIDs <- GenerateWorkspaceID()
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(taskIDs)
	close(workspaceIDs)

	// Collect all IDs
	var allTaskIDs []int
	var allWorkspaceIDs []int

	for id := range taskIDs {
		allTaskIDs = append(allTaskIDs, id)
	}
	for id := range workspaceIDs {
		allWorkspaceIDs = append(allWorkspaceIDs, id)
	}

	// Verify we got the expected number of IDs
	assert.Equal(t, numGoroutines*idsPerGoroutine, len(allTaskIDs))
	assert.Equal(t, numGoroutines*idsPerGoroutine, len(allWorkspaceIDs))

	// Verify all IDs are unique
	taskIDSet := make(map[int]bool)
	workspaceIDSet := make(map[int]bool)

	for _, id := range allTaskIDs {
		assert.False(t, taskIDSet[id], "Duplicate task ID found: %d", id)
		taskIDSet[id] = true
	}

	for _, id := range allWorkspaceIDs {
		assert.False(t, workspaceIDSet[id], "Duplicate workspace ID found: %d", id)
		workspaceIDSet[id] = true
	}

	// Verify IDs are in expected ranges (though order may vary due to concurrency)
	// Tasks start from 1, standalone workspaces start from 1000
	totalExpectedNumbers := make(map[int]bool)
	// Task IDs: 1 to numGoroutines*idsPerGoroutine
	for i := 1; i <= numGoroutines*idsPerGoroutine; i++ {
		totalExpectedNumbers[i] = true
	}
	// Workspace IDs: 1000 to 1000 + numGoroutines*idsPerGoroutine - 1
	for i := 1000; i < 1000+numGoroutines*idsPerGoroutine; i++ {
		totalExpectedNumbers[i] = true
	}

	// Collect all numbers from both task and workspace IDs
	allNumbers := make(map[int]bool)
	for _, id := range allTaskIDs {
		allNumbers[id] = true
	}

	for _, id := range allWorkspaceIDs {
		allNumbers[id] = true
	}

	// Verify all numbers are in the expected range
	for num := range allNumbers {
		assert.True(t, totalExpectedNumbers[num], "Unexpected number: %d", num)
	}
}

// TestIDGenerator_WorkflowIDsIndependent tests that workflow IDs are independent
func TestIDGenerator_WorkflowIDsIndependent(t *testing.T) {
	// Test case: Workflow IDs should be independent of task/workspace IDs
	ResetIDGenerators()

	// Generate some task and workspace IDs
	taskID1 := GenerateTaskID()
	workspaceID1 := GenerateWorkspaceID()

	// Generate workflow IDs
	workflowID1 := GenerateWorkflowID()
	workflowID2 := GenerateWorkflowID()

	// Verify workflow IDs are independent
	assert.Equal(t, 1, taskID1)
	assert.Equal(t, 1000, workspaceID1) // Standalone workspaces start from 1000
	assert.Equal(t, 2, workflowID1)
	assert.Equal(t, 3, workflowID2)

	// Generate more task/workspace IDs
	taskID2 := GenerateTaskID()
	workspaceID2 := GenerateWorkspaceID()

	// Verify task/workspace sequence continues
	assert.Equal(t, 4, taskID2)
	assert.Equal(t, 1001, workspaceID2)
}

func TestGenerateTaskID(t *testing.T) {
	ResetIDGenerators()

	// Generate multiple task IDs
	taskID1 := GenerateTaskID()
	taskID2 := GenerateTaskID()
	taskID3 := GenerateTaskID()

	// Verify they are sequential
	if taskID1 != 1 {
		t.Errorf("Expected task ID 1, got %d", taskID1)
	}
	if taskID2 != 2 {
		t.Errorf("Expected task ID 2, got %d", taskID2)
	}
	if taskID3 != 3 {
		t.Errorf("Expected task ID 3, got %d", taskID3)
	}
}

func TestGenerateWorkspaceID_TaskCreated(t *testing.T) {
	ResetIDGenerators()

	// Create a task
	taskID := GenerateTaskID()

	// Generate workspace ID for the task
	workspaceID := GenerateWorkspaceID(taskID)

	// Verify workspace shares the same ID as the task
	if workspaceID != taskID {
		t.Errorf("Expected workspace ID %d to match task ID %d", workspaceID, taskID)
	}
}

func TestGenerateWorkspaceID_Standalone(t *testing.T) {
	ResetIDGenerators()

	// Generate standalone workspace IDs
	workspaceID1 := GenerateWorkspaceID() // Should be 1000
	workspaceID2 := GenerateWorkspaceID() // Should be 1001
	workspaceID3 := GenerateWorkspaceID() // Should be 1002

	// Verify they are sequential and start from 1000
	if workspaceID1 != 1000 {
		t.Errorf("Expected standalone workspace ID 1000, got %d", workspaceID1)
	}
	if workspaceID2 != 1001 {
		t.Errorf("Expected standalone workspace ID 1001, got %d", workspaceID2)
	}
	if workspaceID3 != 1002 {
		t.Errorf("Expected standalone workspace ID 1002, got %d", workspaceID3)
	}
}

func TestGenerateWorkspaceID_Mixed(t *testing.T) {
	ResetIDGenerators()

	// Create a task
	taskID := GenerateTaskID() // Should be 1

	// Generate standalone workspace IDs
	standalone1 := GenerateWorkspaceID() // Should be 1000
	standalone2 := GenerateWorkspaceID() // Should be 1001

	// Generate workspace ID for the task
	taskWorkspaceID := GenerateWorkspaceID(taskID) // Should be 1 (same as task)

	// Verify the relationships
	if taskWorkspaceID != taskID {
		t.Errorf("Task workspace ID %d should match task ID %d", taskWorkspaceID, taskID)
	}
	if standalone1 != 1000 {
		t.Errorf("Expected standalone workspace ID 1000, got %d", standalone1)
	}
	if standalone2 != 1001 {
		t.Errorf("Expected standalone workspace ID 1001, got %d", standalone2)
	}
	if taskWorkspaceID == standalone1 {
		t.Errorf("Task workspace ID %d should not conflict with standalone workspace ID %d", taskWorkspaceID, standalone1)
	}
}
