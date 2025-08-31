package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hlfshell/cowork/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAiderAgent_NewAiderAgent tests the creation of a new Aider agent
func TestAiderAgent_NewAiderAgent(t *testing.T) {
	agent := NewAiderAgent()

	assert.NotNil(t, agent)
	assert.Equal(t, AgentStatusIdle, agent.GetStatus())
	assert.Equal(t, "aider", agent.GetName())
	assert.Equal(t, "latest", agent.GetVersion())

	info := agent.GetInfo()
	assert.Equal(t, "Aider AI Coding Agent", info["name"])
	assert.Equal(t, "paulgauthier/aider", info["container"])
}

// TestAiderAgent_Initialize tests agent initialization
func TestAiderAgent_Initialize(t *testing.T) {
	agent := NewAiderAgent()

	// Create a temporary working directory
	tempDir := t.TempDir()

	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: tempDir,
		Command:    []string{"aider"},
		Environment: map[string]string{
			"OPENAI_API_KEY": "test-key",
		},
		Timeout:    30 * time.Minute,
		MaxRetries: 3,
	}

	ctx := context.Background()
	err := agent.Initialize(ctx, config)

	// Note: This test will fail if Docker/Podman is not available
	// In a real environment, we'd mock the container manager
	if err != nil {
		t.Logf("Initialization failed (expected if no container engine available): %v", err)
		return
	}

	assert.Equal(t, AgentStatusIdle, agent.GetStatus())

	history := agent.GetHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "initialized", history[0].EventType)
}

// TestAiderAgent_GenerateInstructions tests instruction generation from tasks
func TestAiderAgent_GenerateInstructions(t *testing.T) {
	agent := NewAiderAgent()

	task := &types.CreateTaskRequest{
		Name:        "Implement OAuth refresh",
		Description: "Add OAuth token refresh logic to handle expired tokens automatically",
		Metadata: map[string]string{
			"priority":  "high",
			"component": "auth",
		},
		Tags: []string{"authentication", "oauth", "security"},
	}

	instructions, err := agent.GenerateInstructions(task)
	require.NoError(t, err)

	assert.Contains(t, instructions, "# Task: Implement OAuth refresh")
	assert.Contains(t, instructions, "Add OAuth token refresh logic to handle expired tokens automatically")
	assert.Contains(t, instructions, "Group files into commits based on functionality")
	assert.Contains(t, instructions, "Write meaningful commit messages")
	assert.Contains(t, instructions, "Push the branch remotely")
	assert.Contains(t, instructions, "**priority:** high")
	assert.Contains(t, instructions, "**component:** auth")
	assert.Contains(t, instructions, "authentication, oauth, security")
}

// TestAiderAgent_GenerateInstructions_InvalidTask tests instruction generation with invalid task type
func TestAiderAgent_GenerateInstructions_InvalidTask(t *testing.T) {
	agent := NewAiderAgent()

	// Pass invalid task type
	instructions, err := agent.GenerateInstructions("invalid task")

	assert.Error(t, err)
	assert.Empty(t, instructions)
	assert.Contains(t, err.Error(), "invalid task type")
}

// TestAiderAgent_CreateInstructionFile tests instruction file creation
func TestAiderAgent_CreateInstructionFile(t *testing.T) {
	agent := NewAiderAgent()

	instruction := &AgentInstruction{
		Content:   "Test instruction content",
		TaskID:    123,
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			"test": "value",
		},
	}

	filePath, err := agent.createInstructionFile(instruction)
	require.NoError(t, err)
	defer os.Remove(filePath)

	// Check file exists
	assert.FileExists(t, filePath)

	// Check file content
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, instruction.Content, string(content))
}

// TestAiderAgent_CreateEnvFile tests environment file creation
func TestAiderAgent_CreateEnvFile(t *testing.T) {
	agent := NewAiderAgent()

	// Initialize with environment variables
	config := &AgentConfig{
		AgentType:  "aider",
		WorkingDir: t.TempDir(),
		Environment: map[string]string{
			"OPENAI_API_KEY":    "sk-test-openai-key",
			"ANTHROPIC_API_KEY": "sk-ant-test-key",
			"GEMINI_API_KEY":    "test-gemini-key",
			"OTHER_VAR":         "should-not-be-included",
		},
	}

	agent.config = config

	envFile, err := agent.createEnvFile()
	require.NoError(t, err)
	defer os.Remove(envFile)

	// Check file exists
	assert.FileExists(t, envFile)

	// Check file content
	content, err := os.ReadFile(envFile)
	require.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "OPENAI_API_KEY=sk-test-openai-key")
	assert.Contains(t, contentStr, "ANTHROPIC_API_KEY=sk-ant-test-key")
	assert.Contains(t, contentStr, "GEMINI_API_KEY=test-gemini-key")
	assert.NotContains(t, contentStr, "OTHER_VAR")
}

// TestAiderAgent_GetHistory tests history tracking
func TestAiderAgent_GetHistory(t *testing.T) {
	agent := NewAiderAgent()

	// Initially empty
	history := agent.GetHistory()
	assert.Len(t, history, 0)

	// Add some history entries
	agent.addHistoryEntry("test_event", "Test description", map[string]interface{}{
		"key": "value",
	})

	history = agent.GetHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "test_event", history[0].EventType)
	assert.Equal(t, "Test description", history[0].Description)
	assert.Equal(t, "value", history[0].Data["key"])
}

// TestAiderAgent_GetLastResult tests result tracking
func TestAiderAgent_GetLastResult(t *testing.T) {
	agent := NewAiderAgent()

	// Initially nil
	result := agent.GetLastResult()
	assert.Nil(t, result)

	// Set a result
	testResult := &AgentResult{
		Success:     true,
		Summary:     "Test completed",
		CompletedAt: time.Now(),
		Duration:    5 * time.Minute,
	}

	agent.lastResult = testResult

	result = agent.GetLastResult()
	assert.NotNil(t, result)
	assert.Equal(t, testResult.Success, result.Success)
	assert.Equal(t, testResult.Summary, result.Summary)
}

// TestAiderAgent_Stop tests agent stopping
func TestAiderAgent_Stop(t *testing.T) {
	agent := NewAiderAgent()

	// Should fail when not working
	err := agent.Stop(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not currently working")

	// Set status to working
	agent.status = AgentStatusWorking

	err = agent.Stop(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, AgentStatusStopped, agent.GetStatus())
}

// TestAiderAgent_Cleanup tests agent cleanup
func TestAiderAgent_Cleanup(t *testing.T) {
	agent := NewAiderAgent()

	// Set status to working
	agent.status = AgentStatusWorking

	ctx := context.Background()
	err := agent.Cleanup(ctx)
	assert.NoError(t, err)
	assert.Equal(t, AgentStatusIdle, agent.GetStatus())
}

// TestCopyFile tests the copyFile utility function
func TestCopyFile(t *testing.T) {
	// Create source file
	srcFile := filepath.Join(t.TempDir(), "source.txt")
	srcContent := "test content"
	err := os.WriteFile(srcFile, []byte(srcContent), 0644)
	require.NoError(t, err)

	// Create destination path
	dstFile := filepath.Join(t.TempDir(), "destination.txt")

	// Copy file
	err = copyFile(srcFile, dstFile)
	assert.NoError(t, err)

	// Check destination file exists and has correct content
	assert.FileExists(t, dstFile)
	dstContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, srcContent, string(dstContent))
}

// TestCopyFile_SourceNotExists tests copyFile with non-existent source
func TestCopyFile_SourceNotExists(t *testing.T) {
	dstFile := filepath.Join(t.TempDir(), "destination.txt")

	err := copyFile("/non/existent/file", dstFile)
	assert.Error(t, err)
}

// TestAgentResult tests AgentResult structure
func TestAgentResult(t *testing.T) {
	result := &AgentResult{
		Success:       true,
		Summary:       "Test summary",
		Output:        "Test output",
		ModifiedFiles: []string{"file1.go", "file2.go"},
		CreatedFiles:  []string{"newfile.go"},
		DeletedFiles:  []string{},
		Error:         "",
		Metadata: map[string]string{
			"task_id": "123",
		},
		CompletedAt: time.Now(),
		Duration:    5 * time.Minute,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "Test summary", result.Summary)
	assert.Len(t, result.ModifiedFiles, 2)
	assert.Len(t, result.CreatedFiles, 1)
	assert.Len(t, result.DeletedFiles, 0)
	assert.Empty(t, result.Error)
	assert.Equal(t, "123", result.Metadata["task_id"])
}
