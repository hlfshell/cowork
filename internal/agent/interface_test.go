package agent

import (
	"testing"
	"time"
)

// TestAgentStatus_String tests the String method for AgentStatus
func TestAgentStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   AgentStatus
		expected string
	}{
		{
			name:     "Idle status",
			status:   AgentStatusIdle,
			expected: "idle",
		},
		{
			name:     "Working status",
			status:   AgentStatusWorking,
			expected: "working",
		},
		{
			name:     "Completed status",
			status:   AgentStatusCompleted,
			expected: "completed",
		},
		{
			name:     "Failed status",
			status:   AgentStatusFailed,
			expected: "failed",
		},
		{
			name:     "Stopped status",
			status:   AgentStatusStopped,
			expected: "stopped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("AgentStatus.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAgentStatus_IsValid tests the IsValid method for AgentStatus
func TestAgentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   AgentStatus
		expected bool
	}{
		{
			name:     "Valid idle status",
			status:   AgentStatusIdle,
			expected: true,
		},
		{
			name:     "Valid working status",
			status:   AgentStatusWorking,
			expected: true,
		},
		{
			name:     "Valid completed status",
			status:   AgentStatusCompleted,
			expected: true,
		},
		{
			name:     "Valid failed status",
			status:   AgentStatusFailed,
			expected: true,
		},
		{
			name:     "Valid stopped status",
			status:   AgentStatusStopped,
			expected: true,
		},
		{
			name:     "Invalid status",
			status:   AgentStatus("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("AgentStatus.IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAgentStatus_IsTerminal tests the IsTerminal method for AgentStatus
func TestAgentStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name     string
		status   AgentStatus
		expected bool
	}{
		{
			name:     "Non-terminal idle status",
			status:   AgentStatusIdle,
			expected: false,
		},
		{
			name:     "Non-terminal working status",
			status:   AgentStatusWorking,
			expected: false,
		},
		{
			name:     "Terminal completed status",
			status:   AgentStatusCompleted,
			expected: true,
		},
		{
			name:     "Terminal failed status",
			status:   AgentStatusFailed,
			expected: true,
		},
		{
			name:     "Terminal stopped status",
			status:   AgentStatusStopped,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsTerminal()
			if result != tt.expected {
				t.Errorf("AgentStatus.IsTerminal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestInstructionType_String tests the String method for InstructionType
func TestInstructionType_String(t *testing.T) {
	tests := []struct {
		name     string
		instType InstructionType
		expected string
	}{
		{
			name:     "Task instruction type",
			instType: InstructionTypeTask,
			expected: "task",
		},
		{
			name:     "PR review instruction type",
			instType: InstructionTypePRReview,
			expected: "pr_review",
		},
		{
			name:     "User message instruction type",
			instType: InstructionTypeUserMessage,
			expected: "user_message",
		},
		{
			name:     "Issue instruction type",
			instType: InstructionTypeIssue,
			expected: "issue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.instType.String()
			if result != tt.expected {
				t.Errorf("InstructionType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestInstructionType_IsValid tests the IsValid method for InstructionType
func TestInstructionType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		instType InstructionType
		expected bool
	}{
		{
			name:     "Valid task instruction type",
			instType: InstructionTypeTask,
			expected: true,
		},
		{
			name:     "Valid PR review instruction type",
			instType: InstructionTypePRReview,
			expected: true,
		},
		{
			name:     "Valid user message instruction type",
			instType: InstructionTypeUserMessage,
			expected: true,
		},
		{
			name:     "Valid issue instruction type",
			instType: InstructionTypeIssue,
			expected: true,
		},
		{
			name:     "Invalid instruction type",
			instType: InstructionType("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.instType.IsValid()
			if result != tt.expected {
				t.Errorf("InstructionType.IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAgentInstruction_Validate tests the Validate method for AgentInstruction
func TestAgentInstruction_Validate(t *testing.T) {
	validTime := time.Now()

	tests := []struct {
		name        string
		instruction *AgentInstruction
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid instruction",
			instruction: &AgentInstruction{
				Content:   "Test task content",
				TaskID:    123,
				CreatedAt: validTime,
			},
			expectError: false,
		},
		{
			name: "Missing content",
			instruction: &AgentInstruction{
				TaskID:    123,
				CreatedAt: validTime,
			},
			expectError: true,
			errorMsg:    "instruction content is required",
		},
		{
			name: "Zero task ID",
			instruction: &AgentInstruction{
				Content:   "Test content",
				TaskID:    0,
				CreatedAt: validTime,
			},
			expectError: true,
			errorMsg:    "task ID must be positive",
		},
		{
			name: "Negative task ID",
			instruction: &AgentInstruction{
				Content:   "Test content",
				TaskID:    -1,
				CreatedAt: validTime,
			},
			expectError: true,
			errorMsg:    "task ID must be positive",
		},
		{
			name: "Missing content",
			instruction: &AgentInstruction{
				TaskID:    123,
				CreatedAt: validTime,
			},
			expectError: true,
			errorMsg:    "instruction content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.instruction.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestAgentConfig_Validate tests the Validate method for AgentConfig
func TestAgentConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *AgentConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid configuration",
			config: &AgentConfig{
				AgentType:  "aider",
				WorkingDir: "/tmp/test",
				Command:    []string{"aider"},
				Timeout:    30 * time.Minute,
				MaxRetries: 3,
			},
			expectError: false,
		},
		{
			name: "Missing agent type",
			config: &AgentConfig{
				WorkingDir: "/tmp/test",
				Command:    []string{"aider"},
				Timeout:    30 * time.Minute,
				MaxRetries: 3,
			},
			expectError: true,
			errorMsg:    "agent type is required",
		},
		{
			name: "Missing working directory",
			config: &AgentConfig{
				AgentType:  "aider",
				Command:    []string{"aider"},
				Timeout:    30 * time.Minute,
				MaxRetries: 3,
			},
			expectError: true,
			errorMsg:    "working directory is required",
		},
		{
			name: "Missing command",
			config: &AgentConfig{
				AgentType:  "aider",
				WorkingDir: "/tmp/test",
				Timeout:    30 * time.Minute,
				MaxRetries: 3,
			},
			expectError: true,
			errorMsg:    "command is required",
		},
		{
			name: "Zero timeout",
			config: &AgentConfig{
				AgentType:  "aider",
				WorkingDir: "/tmp/test",
				Command:    []string{"aider"},
				Timeout:    0,
				MaxRetries: 3,
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "Negative timeout",
			config: &AgentConfig{
				AgentType:  "aider",
				WorkingDir: "/tmp/test",
				Command:    []string{"aider"},
				Timeout:    -1 * time.Minute,
				MaxRetries: 3,
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "Negative max retries",
			config: &AgentConfig{
				AgentType:  "aider",
				WorkingDir: "/tmp/test",
				Command:    []string{"aider"},
				Timeout:    30 * time.Minute,
				MaxRetries: -1,
			},
			expectError: true,
			errorMsg:    "max retries must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
