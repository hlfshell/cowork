package types

import (
	"sync"
)

// IDGenerator provides short, human-readable sequential IDs
type IDGenerator struct {
	mu       sync.Mutex
	sequence int
}

// NewIDGenerator creates a new ID generator
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		sequence: 1,
	}
}

// GenerateID generates a simple incrementing integer ID
func (g *IDGenerator) GenerateID() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := g.sequence
	g.sequence++
	return id
}

// WorkspaceIDGenerator provides IDs for standalone workspaces (w1, w2, etc.)
type WorkspaceIDGenerator struct {
	mu       sync.Mutex
	sequence int
}

// NewWorkspaceIDGenerator creates a new workspace ID generator
func NewWorkspaceIDGenerator() *WorkspaceIDGenerator {
	return &WorkspaceIDGenerator{
		sequence: 1000, // Start from 1000 to avoid conflicts with task IDs
	}
}

// GenerateWorkspaceID generates a workspace ID for standalone workspaces
func (g *WorkspaceIDGenerator) GenerateWorkspaceID() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := g.sequence
	g.sequence++
	return id
}

// Global ID generators
var (
	sharedIDGenerator    = NewIDGenerator()          // Shared generator for tasks and task-created workspaces
	workspaceIDGenerator = NewWorkspaceIDGenerator() // Separate generator for standalone workspaces
)

// GenerateTaskID generates a task ID
func GenerateTaskID() int {
	return sharedIDGenerator.GenerateID()
}

// GenerateWorkspaceID generates a workspace ID
// If taskID is provided, returns the same ID (for task-created workspaces)
// Otherwise, generates a new standalone workspace ID
func GenerateWorkspaceID(taskID ...int) int {
	if len(taskID) > 0 && taskID[0] > 0 {
		// Task-created workspace: use the same ID as the task
		return taskID[0]
	}
	// Standalone workspace: generate a new ID
	return workspaceIDGenerator.GenerateWorkspaceID()
}

// GenerateWorkflowID generates a workflow ID
func GenerateWorkflowID() int {
	return sharedIDGenerator.GenerateID()
}

// ResetIDGenerators resets all ID generators (mainly for testing)
func ResetIDGenerators() {
	sharedIDGenerator = NewIDGenerator()
	workspaceIDGenerator = NewWorkspaceIDGenerator()
}
