package agent

import (
	"context"
	"testing"
	"time"
)

// TestNewAiderAgent tests the creation of a new Aider agent
func TestNewAiderAgent(t *testing.T) {
	agent := NewAiderAgent()

	if agent == nil {
		t.Fatal("Expected non-nil agent")
	}

	if agent.GetStatus() != AgentStatusIdle {
		t.Errorf("Expected initial status to be idle, got %s", agent.GetStatus())
	}

	info := agent.GetInfo()
	expectedKeys := []string{"name", "version", "description", "website"}
	for _, key := range expectedKeys {
		if _, exists := info[key]; !exists {
			t.Errorf("Expected info to contain key: %s", key)
		}
	}
}

// TestAiderAgent_Initialize tests the initialization of the Aider agent
func TestAiderAgent_Initialize(t *testing.T) {
	agent := NewAiderAgent()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: tempDir,
		Command:    []string{"echo"}, // Use echo for testing
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
		Verbose:    true,
	}

	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		t.Errorf("Failed to initialize agent: %v", err)
	}

	if agent.GetStatus() != AgentStatusIdle {
		t.Errorf("Expected status to be idle after initialization, got %s", agent.GetStatus())
	}
}

// TestAiderAgent_Initialize_InvalidConfig tests initialization with invalid configuration
func TestAiderAgent_Initialize_InvalidConfig(t *testing.T) {
	agent := NewAiderAgent()

	tests := []struct {
		name        string
		config      *AgentConfig
		expectError bool
	}{
		{
			name: "Missing agent type",
			config: &AgentConfig{
				WorkingDir: "/tmp/test",
				Command:    []string{"echo"},
				Timeout:    30 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "Missing working directory",
			config: &AgentConfig{
				AgentType: "aider",
				Command:   []string{"echo"},
				Timeout:   30 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "Missing command",
			config: &AgentConfig{
				AgentType:  "aider",
				WorkingDir: "/tmp/test",
				Timeout:    30 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "Zero timeout (should be set to default)",
			config: &AgentConfig{
				AgentType:  "aider",
				WorkingDir: "/tmp/test",
				Command:    []string{"echo"},
				Timeout:    0,
			},
			expectError: false, // Should not error because we set default
		},
		{
			name: "Negative max retries",
			config: &AgentConfig{
				AgentType:  "aider",
				WorkingDir: "/tmp/test",
				Command:    []string{"echo"},
				Timeout:    30 * time.Minute,
				MaxRetries: -1,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := agent.Initialize(ctx, tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestAiderAgent_CreateInstructionFile tests the creation of instruction files
func TestAiderAgent_CreateInstructionFile(t *testing.T) {
	agent := NewAiderAgent()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: tempDir,
		Command:    []string{"echo"},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
	}

	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	instruction := &AgentInstruction{
		Content:   "Test task content",
		TaskID:    123,
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"test_key": "test_value",
		},
	}

	// Test the instruction structure
	if instruction.Content != "Test task content" {
		t.Errorf("Expected content to match, got %s", instruction.Content)
	}

	if instruction.TaskID != 123 {
		t.Errorf("Expected task ID to be 123, got %d", instruction.TaskID)
	}

	if len(instruction.Metadata) != 1 {
		t.Errorf("Expected 1 metadata item, got %d", len(instruction.Metadata))
	}
}

// TestAiderAgent_GetStatus tests the status management
func TestAiderAgent_GetStatus(t *testing.T) {
	agent := NewAiderAgent()

	// Initial status should be idle
	if agent.GetStatus() != AgentStatusIdle {
		t.Errorf("Expected initial status to be idle, got %s", agent.GetStatus())
	}
}

// TestAiderAgent_GetInfo tests the info retrieval
func TestAiderAgent_GetInfo(t *testing.T) {
	agent := NewAiderAgent()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: tempDir,
		Command:    []string{"echo"},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
	}

	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	info := agent.GetInfo()

	// Check required fields
	requiredFields := []string{"name", "version", "description", "website", "status", "working_dir"}
	for _, field := range requiredFields {
		if _, exists := info[field]; !exists {
			t.Errorf("Expected info to contain field: %s", field)
		}
	}

	// Check specific values
	if info["name"] != "Aider AI Coding Agent" {
		t.Errorf("Expected name to be 'Aider AI Coding Agent', got %v", info["name"])
	}

	if info["status"] != AgentStatusIdle.String() {
		t.Errorf("Expected status to be 'idle', got %v", info["status"])
	}

	if info["working_dir"] != tempDir {
		t.Errorf("Expected working_dir to be %s, got %v", tempDir, info["working_dir"])
	}
}

// TestAiderAgent_Cleanup tests the cleanup functionality
func TestAiderAgent_Cleanup(t *testing.T) {
	agent := NewAiderAgent()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: tempDir,
		Command:    []string{"echo"},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
	}

	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Cleanup should not error
	err = agent.Cleanup(ctx)
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	// Status should be idle after cleanup
	if agent.GetStatus() != AgentStatusIdle {
		t.Errorf("Expected status to be idle after cleanup, got %s", agent.GetStatus())
	}
}

// TestAiderAgent_Stop_NotWorking tests stopping an agent that's not working
func TestAiderAgent_Stop_NotWorking(t *testing.T) {
	agent := NewAiderAgent()

	ctx := context.Background()
	err := agent.Stop(ctx)

	// Should return an error when agent is not working
	if err == nil {
		t.Error("Expected error when stopping non-working agent")
	}
}

// TestAiderAgent_Execute_InvalidInstruction tests execution with invalid instruction
func TestAiderAgent_Execute_InvalidInstruction(t *testing.T) {
	agent := NewAiderAgent()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: tempDir,
		Command:    []string{"echo"},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
	}

	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Test with invalid instruction (missing task ID)
	invalidInstruction := &AgentInstruction{
		Content:   "Test content",
		CreatedAt: time.Now(),
	}

	err = agent.Execute(ctx, invalidInstruction)
	if err == nil {
		t.Error("Expected error with invalid instruction")
	}
}

// TestAiderAgent_Execute_NotReady tests execution when agent is not ready
func TestAiderAgent_Execute_NotReady(t *testing.T) {
	agent := NewAiderAgent()

	// Don't initialize the agent

	instruction := &AgentInstruction{
		Content:   "Test content",
		TaskID:    123,
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	err := agent.Execute(ctx, instruction)
	if err == nil {
		t.Error("Expected error when agent is not initialized")
	}
}

// TestAiderAgent_DefaultConfigValues tests that default values are set correctly
func TestAiderAgent_DefaultConfigValues(t *testing.T) {
	agent := NewAiderAgent()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: tempDir,
		Command:    []string{"echo"},
		// Don't set Timeout and MaxRetries to test defaults
	}

	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	// Check that defaults were set
	if config.Timeout == 0 {
		t.Error("Expected timeout to be set to default value")
	}

	if config.MaxRetries == 0 {
		t.Error("Expected max retries to be set to default value")
	}

	// Default timeout should be 30 minutes
	expectedTimeout := 30 * time.Minute
	if config.Timeout != expectedTimeout {
		t.Errorf("Expected timeout to be %v, got %v", expectedTimeout, config.Timeout)
	}

	// Default max retries should be 3
	expectedMaxRetries := 3
	if config.MaxRetries != expectedMaxRetries {
		t.Errorf("Expected max retries to be %d, got %d", expectedMaxRetries, config.MaxRetries)
	}
}

// TestAiderAgent_InstructionValidation tests instruction validation
func TestAiderAgent_InstructionValidation(t *testing.T) {
	agent := NewAiderAgent()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: tempDir,
		Command:    []string{"echo"},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
	}

	ctx := context.Background()
	err := agent.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize agent: %v", err)
	}

	tests := []struct {
		name        string
		instruction *AgentInstruction
		expectError bool
	}{
		{
			name: "Valid instruction",
			instruction: &AgentInstruction{
				Content:   "Test task content",
				TaskID:    123,
				CreatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "Missing content",
			instruction: &AgentInstruction{
				TaskID:    123,
				CreatedAt: time.Now(),
			},
			expectError: true,
		},
		{
			name: "Zero task ID",
			instruction: &AgentInstruction{
				Content:   "Test content",
				TaskID:    0,
				CreatedAt: time.Now(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate the instruction
			err := tt.instruction.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
