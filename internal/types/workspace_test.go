package types

import (
	"testing"
	"time"
)



// TestWorkspaceStatus_String tests the String method for workspace statuses
func TestWorkspaceStatus_String(t *testing.T) {
	// Test case: String method should return the correct string representation
	// for each workspace status
	testCases := []struct {
		status   WorkspaceStatus
		expected string
	}{
		{WorkspaceStatusCreating, "creating"},
		{WorkspaceStatusReady, "ready"},
		{WorkspaceStatusActive, "active"},
		{WorkspaceStatusError, "error"},
		{WorkspaceStatusCleaning, "cleaning"},
		{WorkspaceStatus("custom"), "custom"},
	}

	for _, tc := range testCases {
		result := tc.status.String()
		if result != tc.expected {
			t.Errorf("Expected workspace status string '%s', got '%s'", tc.expected, result)
		}
	}
}

// TestWorkspaceStatus_IsValid tests the IsValid method for workspace statuses
func TestWorkspaceStatus_IsValid(t *testing.T) {
	// Test case: IsValid should return true for valid workspace statuses
	// and false for invalid ones
	testCases := []struct {
		status   WorkspaceStatus
		expected bool
	}{
		{WorkspaceStatusCreating, true},
		{WorkspaceStatusReady, true},
		{WorkspaceStatusActive, true},
		{WorkspaceStatusError, true},
		{WorkspaceStatusCleaning, true},
		{WorkspaceStatus("invalid"), false},
		{WorkspaceStatus(""), false},
	}

	for _, tc := range testCases {
		result := tc.status.IsValid()
		if result != tc.expected {
			t.Errorf("Expected IsValid() to return %t for status '%s', got %t", tc.expected, tc.status, result)
		}
	}
}

// TestCreateWorkspaceRequest_Validate_WithValidRequest tests validation with valid request
func TestCreateWorkspaceRequest_Validate_WithValidRequest(t *testing.T) {
	// Test case: A valid create workspace request should pass validation
	req := &CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: "https://github.com/test/repo.git",
	}
	
	err := req.Validate()
	if err != nil {
		t.Errorf("Expected valid request to pass validation, got error: %v", err)
	}
	
	// Verify defaults were set
	if req.BaseBranch != "main" {
		t.Errorf("Expected BaseBranch to default to 'main', got '%s'", req.BaseBranch)
	}
}

// TestCreateWorkspaceRequest_Validate_WithEmptyTaskName tests validation with empty task name
func TestCreateWorkspaceRequest_Validate_WithEmptyTaskName(t *testing.T) {
	// Test case: A request with empty task name should fail validation
	req := &CreateWorkspaceRequest{
		TaskName:   "",
		SourceRepo: "https://github.com/test/repo.git",
	}
	
	err := req.Validate()
	if err == nil {
		t.Error("Expected validation to fail with empty task name")
	}
	
	expectedError := "task name is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestCreateWorkspaceRequest_Validate_WithEmptySourceRepo tests validation with empty source repo
func TestCreateWorkspaceRequest_Validate_WithEmptySourceRepo(t *testing.T) {
	// Test case: A request with empty source repository should fail validation
	req := &CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: "",
	}
	
	err := req.Validate()
	if err == nil {
		t.Error("Expected validation to fail with empty source repository")
	}
	
	expectedError := "source repository URL is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestCreateWorkspaceRequest_Validate_WithCustomBaseBranch tests validation with custom base branch
func TestCreateWorkspaceRequest_Validate_WithCustomBaseBranch(t *testing.T) {
	// Test case: A request with custom base branch should preserve the value
	req := &CreateWorkspaceRequest{
		TaskName:   "test-task",
		SourceRepo: "https://github.com/test/repo.git",
		BaseBranch: "develop",
	}
	
	err := req.Validate()
	if err != nil {
		t.Errorf("Expected valid request to pass validation, got error: %v", err)
	}
	
	if req.BaseBranch != "develop" {
		t.Errorf("Expected BaseBranch to remain 'develop', got '%s'", req.BaseBranch)
	}
}

// TestWorkspace_Creation tests workspace creation and field access
func TestWorkspace_Creation(t *testing.T) {
	// Test case: Creating a workspace should set all fields correctly
	now := time.Now()
	workspace := &Workspace{
		ID:           "test-id",
		TaskName:     "test-task",
		TicketID:     "123",
		Path:         "/path/to/workspace",
		BranchName:   "task/test-task-123",
		SourceRepo:   "https://github.com/test/repo.git",
		BaseBranch:   "main",
		CreatedAt:    now,
		LastActivity: now,
		Status:       WorkspaceStatusReady,
		ContainerID:  "container-123",
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	
	// Verify all fields are set correctly
	if workspace.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", workspace.ID)
	}
	
	if workspace.TaskName != "test-task" {
		t.Errorf("Expected TaskName 'test-task', got '%s'", workspace.TaskName)
	}
	
	if workspace.TicketID != "123" {
		t.Errorf("Expected TicketID '123', got '%s'", workspace.TicketID)
	}
	
	if workspace.Path != "/path/to/workspace" {
		t.Errorf("Expected Path '/path/to/workspace', got '%s'", workspace.Path)
	}
	
	if workspace.BranchName != "task/test-task-123" {
		t.Errorf("Expected BranchName 'task/test-task-123', got '%s'", workspace.BranchName)
	}
	
	if workspace.SourceRepo != "https://github.com/test/repo.git" {
		t.Errorf("Expected SourceRepo 'https://github.com/test/repo.git', got '%s'", workspace.SourceRepo)
	}
	
	if workspace.BaseBranch != "main" {
		t.Errorf("Expected BaseBranch 'main', got '%s'", workspace.BaseBranch)
	}
	
	if workspace.CreatedAt != now {
		t.Errorf("Expected CreatedAt to match provided time")
	}
	
	if workspace.LastActivity != now {
		t.Errorf("Expected LastActivity to match provided time")
	}
	
	if workspace.Status != WorkspaceStatusReady {
		t.Errorf("Expected Status 'ready', got '%s'", workspace.Status)
	}
	
	if workspace.ContainerID != "container-123" {
		t.Errorf("Expected ContainerID 'container-123', got '%s'", workspace.ContainerID)
	}
	
	if len(workspace.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(workspace.Metadata))
	}
	
	if workspace.Metadata["key1"] != "value1" {
		t.Errorf("Expected metadata key1='value1', got '%s'", workspace.Metadata["key1"])
	}
	
	if workspace.Metadata["key2"] != "value2" {
		t.Errorf("Expected metadata key2='value2', got '%s'", workspace.Metadata["key2"])
	}
}
