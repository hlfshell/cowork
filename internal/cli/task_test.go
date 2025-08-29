package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hlfshell/cowork/internal/types"
	"github.com/spf13/cobra"
)

// TestTaskCreateCommand tests the task create command
func TestTaskCreateCommand(t *testing.T) {
	// Test case: Create command should create a task with valid parameters
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Test create command
	createCmd := findCommand(app.rootCmd, "task", "create")
	if createCmd == nil {
		t.Fatal("Create command not found")
	}

	// Test command structure
	if createCmd.Use != "create [task-name]" {
		t.Errorf("Expected command use 'create [task-name]', got '%s'", createCmd.Use)
	}

	// Test flags
	messageFlag := createCmd.Flags().Lookup("message")
	if messageFlag == nil {
		t.Error("Message flag not found on create command")
	}

	priorityFlag := createCmd.Flags().Lookup("priority")
	if priorityFlag == nil {
		t.Error("Priority flag not found on create command")
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "create", "test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Create command failed: %v", err)
	}

	outputStr := output.String()

	// Check that task was created
	if !strings.Contains(outputStr, "Creating task: test-task") {
		t.Errorf("Output should indicate task creation, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "✅ Task created successfully!") {
		t.Errorf("Output should indicate successful creation, got: %s", outputStr)
	}

	// Verify task was actually created
	tasks, err := app.taskManager.ListTasks(nil)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Name != "test-task" {
		t.Errorf("Expected task name 'test-task', got '%s'", task.Name)
	}

	if task.Status != types.TaskStatusQueued {
		t.Errorf("Expected task status 'queued', got '%s'", task.Status)
	}
}

// TestTaskCreateCommand_WithDescription tests the task create command with description
func TestTaskCreateCommand_WithDescription(t *testing.T) {
	// Test case: Create command should create a task with description
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Execute the command with description
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "create", "test-task", "-m", "Test task description"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Create command failed: %v", err)
	}

	outputStr := output.String()

	// Check that description was included
	if !strings.Contains(outputStr, "Description: Test task description") {
		t.Errorf("Output should include description, got: %s", outputStr)
	}

	// Verify task was created with description
	tasks, err := app.taskManager.ListTasks(nil)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Description != "Test task description" {
		t.Errorf("Expected description 'Test task description', got '%s'", task.Description)
	}
}

// TestTaskCreateCommand_WithCostTracking tests the task create command with cost tracking
func TestTaskCreateCommand_WithCostTracking(t *testing.T) {
	// Test case: Create command should create a task with cost tracking
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Execute the command with cost tracking
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{
		"task", "create", "cost-task",
		"--estimated-cost", "25.50",
		"--currency", "EUR",
		"-m", "Task with cost tracking",
	})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Create command failed: %v", err)
	}

	outputStr := output.String()

	// Check that task was created successfully
	if !strings.Contains(outputStr, "✅ Task created successfully!") {
		t.Errorf("Output should indicate successful creation, got: %s", outputStr)
	}

	// Verify task was created with cost tracking
	tasks, err := app.taskManager.ListTasks(nil)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Name != "cost-task" {
		t.Errorf("Expected task name 'cost-task', got '%s'", task.Name)
	}

	if task.EstimatedCost != 25.50 {
		t.Errorf("Expected estimated cost 25.50, got %f", task.EstimatedCost)
	}

	if task.Currency != "EUR" {
		t.Errorf("Expected currency 'EUR', got '%s'", task.Currency)
	}

	if task.ActualCost != 0.0 {
		t.Errorf("Expected actual cost 0.0, got %f", task.ActualCost)
	}
}

// TestTaskListCommand tests the task list command
func TestTaskListCommand(t *testing.T) {
	// Test case: List command should show all tasks
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create multiple tasks
	req1 := &types.CreateTaskRequest{
		Name:        "task-1",
		Description: "First task description",
		Priority:    1,
	}

	req2 := &types.CreateTaskRequest{
		Name:        "task-2",
		Description: "Second task description",
		Priority:    2,
	}

	_, err = app.taskManager.CreateTask(req1)
	if err != nil {
		t.Fatalf("Failed to create first task: %v", err)
	}

	_, err = app.taskManager.CreateTask(req2)
	if err != nil {
		t.Fatalf("Failed to create second task: %v", err)
	}

	// Test list command
	listCmd := findCommand(app.rootCmd, "task", "list")
	if listCmd == nil {
		t.Fatal("List command not found")
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "list"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	outputStr := output.String()

	// Check that tasks are listed
	if !strings.Contains(outputStr, "Found 2 task(s):") {
		t.Errorf("Output should indicate 2 tasks found, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "⏳ task-1") {
		t.Errorf("Output should show task-1, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "⏳ task-2") {
		t.Errorf("Output should show task-2, got: %s", outputStr)
	}

	// Check that descriptions are shown
	if !strings.Contains(outputStr, "First task description") {
		t.Errorf("Output should show first task description, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Second task description") {
		t.Errorf("Output should show second task description, got: %s", outputStr)
	}
}

// TestTaskListCommand_WithNoTasks tests the task list command with no tasks
func TestTaskListCommand_WithNoTasks(t *testing.T) {
	// Test case: List command should show appropriate message when no tasks exist
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "list"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	outputStr := output.String()

	// Check that appropriate message is shown
	if !strings.Contains(outputStr, "No tasks found.") {
		t.Errorf("Output should indicate no tasks found, got: %s", outputStr)
	}
}

// TestTaskDescribeCommand tests the task describe command
func TestTaskDescribeCommand(t *testing.T) {
	// Test case: Describe command should show detailed task information
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task first
	req := &types.CreateTaskRequest{
		Name:        "test-task",
		Description: "Test task description",
		Priority:    5,
		TicketID:    "GH-123",
		URL:         "https://github.com/example/repo/issues/123",
	}

	task, err := app.taskManager.CreateTask(req)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Test describe command
	describeCmd := findCommand(app.rootCmd, "task", "describe")
	if describeCmd == nil {
		t.Fatal("Describe command not found")
	}

	// Test command structure
	if describeCmd.Use != "describe [task-id-or-name]" {
		t.Errorf("Expected command use 'describe [task-id-or-name]', got '%s'", describeCmd.Use)
	}

	// Execute the command with task name
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "describe", "test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Describe command failed: %v", err)
	}

	outputStr := output.String()

	// Check that detailed information is shown
	if !strings.Contains(outputStr, "📋 Task Details") {
		t.Errorf("Output should show task details header, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Name: test-task") {
		t.Errorf("Output should show task name, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "ID: "+fmt.Sprintf("%d", task.ID)) {
		t.Errorf("Output should show task ID, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Priority: 5") {
		t.Errorf("Output should show priority, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Ticket ID: GH-123") {
		t.Errorf("Output should show ticket ID, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "URL: https://github.com/example/repo/issues/123") {
		t.Errorf("Output should show URL, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Description:") {
		t.Errorf("Output should show description section, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Test task description") {
		t.Errorf("Output should show description content, got: %s", outputStr)
	}
}

// TestTaskNextCommand tests the task next command
func TestTaskNextCommand(t *testing.T) {
	// Test case: Next command should return the highest priority queued task
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create multiple tasks with different priorities
	req1 := &types.CreateTaskRequest{
		Name:     "low-priority",
		Priority: 1,
	}

	req2 := &types.CreateTaskRequest{
		Name:     "high-priority",
		Priority: 5,
	}

	req3 := &types.CreateTaskRequest{
		Name:     "medium-priority",
		Priority: 3,
	}

	_, err = app.taskManager.CreateTask(req1)
	if err != nil {
		t.Fatalf("Failed to create first task: %v", err)
	}

	_, err = app.taskManager.CreateTask(req2)
	if err != nil {
		t.Fatalf("Failed to create second task: %v", err)
	}

	_, err = app.taskManager.CreateTask(req3)
	if err != nil {
		t.Fatalf("Failed to create third task: %v", err)
	}

	// Test next command
	nextCmd := findCommand(app.rootCmd, "task", "next")
	if nextCmd == nil {
		t.Fatal("Next command not found")
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "next"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Next command failed: %v", err)
	}

	outputStr := output.String()

	// Check that the highest priority task is returned
	if !strings.Contains(outputStr, "🎯 Next Task in Queue") {
		t.Errorf("Output should show next task header, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Name: high-priority") {
		t.Errorf("Output should show highest priority task, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Priority: 5") {
		t.Errorf("Output should show highest priority, got: %s", outputStr)
	}

	// Verify it's the correct task
	if !strings.Contains(outputStr, "Status: queued") {
		t.Errorf("Output should show queued status, got: %s", outputStr)
	}
}

// TestTaskNextCommand_NoQueuedTasks tests the task next command with no queued tasks
func TestTaskNextCommand_NoQueuedTasks(t *testing.T) {
	// Test case: Next command should fail when no queued tasks exist
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task and move it to in_progress (not queued)
	req := &types.CreateTaskRequest{
		Name:     "in-progress-task",
		Priority: 5,
	}

	task, err := app.taskManager.CreateTask(req)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Move task to in_progress
	status := types.TaskStatusInProgress
	updateReq := &types.UpdateTaskRequest{
		TaskID: task.ID,
		Status: &status,
	}

	_, err = app.taskManager.UpdateTask(updateReq)
	if err != nil {
		t.Fatalf("Failed to update task status: %v", err)
	}

	// Test next command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "next"})

	err = app.rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for no queued tasks, got nil")
	}

	expectedError := "no queued tasks found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestTaskStatsCommand tests the task stats command
func TestTaskStatsCommand(t *testing.T) {
	// Test case: Stats command should show task statistics
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create some tasks with cost tracking
	req1 := &types.CreateTaskRequest{
		Name:          "task-1",
		Priority:      1,
		EstimatedCost: 10.0,
		Currency:      "USD",
	}

	req2 := &types.CreateTaskRequest{
		Name:          "task-2",
		Priority:      2,
		EstimatedCost: 15.0,
		Currency:      "USD",
	}

	_, err = app.taskManager.CreateTask(req1)
	if err != nil {
		t.Fatalf("Failed to create first task: %v", err)
	}

	_, err = app.taskManager.CreateTask(req2)
	if err != nil {
		t.Fatalf("Failed to create second task: %v", err)
	}

	// Test stats command
	statsCmd := findCommand(app.rootCmd, "task", "stats")
	if statsCmd == nil {
		t.Fatal("Stats command not found")
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "stats"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Stats command failed: %v", err)
	}

	outputStr := output.String()

	// Check that statistics are shown
	if !strings.Contains(outputStr, "📊 Task Statistics") {
		t.Errorf("Output should show statistics header, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Total Tasks: 2") {
		t.Errorf("Output should show total tasks, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Created Today: 2") {
		t.Errorf("Output should show created today, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Total Cost: 25.00 USD") {
		t.Errorf("Output should show total cost, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "By Status:") {
		t.Errorf("Output should show status breakdown, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "queued: 2") {
		t.Errorf("Output should show queued tasks, got: %s", outputStr)
	}
}

// TestTaskStartCommand tests the task start command
func TestTaskStartCommand(t *testing.T) {
	// Test case: Start command should create task, workspace, and move to in_progress
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Test start command
	startCmd := findCommand(app.rootCmd, "task", "start")
	if startCmd == nil {
		t.Fatal("Start command not found")
	}

	// Test command structure
	if startCmd.Use != "start [task-name]" {
		t.Errorf("Expected command use 'start [task-name]', got '%s'", startCmd.Use)
	}

	// Execute the command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "start", "test-task", "-m", "Test task description"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Start command failed: %v", err)
	}

	outputStr := output.String()

	// Check that task was created and started
	if !strings.Contains(outputStr, "Creating new task: test-task") {
		t.Errorf("Output should indicate task creation, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "✅ Task created and started successfully!") {
		t.Errorf("Output should indicate successful start, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Status: in_progress") {
		t.Errorf("Output should show in_progress status, got: %s", outputStr)
	}

	// Verify task was actually created and started
	tasks, err := app.taskManager.ListTasks(nil)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Name != "test-task" {
		t.Errorf("Expected task name 'test-task', got '%s'", task.Name)
	}

	if task.Status != types.TaskStatusInProgress {
		t.Errorf("Expected task status 'in_progress', got '%s'", task.Status)
	}

	if task.WorkspaceID == 0 {
		t.Errorf("Expected task to have workspace ID, got 0")
	}

	if task.WorkspacePath == "" {
		t.Errorf("Expected task to have workspace path, got empty")
	}

	// Verify workspace was actually created by checking the workspace path
	if task.WorkspacePath == "" {
		t.Error("Expected task to have workspace path")
	}

	if _, err := os.Stat(task.WorkspacePath); os.IsNotExist(err) {
		t.Errorf("Workspace directory should exist: %s", task.WorkspacePath)
	}
}

// TestTaskStartCommand_ExistingTask tests the task start command with existing task
func TestTaskStartCommand_ExistingTask(t *testing.T) {
	// Test case: Start command should start existing task and create workspace if needed
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task first (without workspace)
	req := &types.CreateTaskRequest{
		Name:        "existing-task",
		Description: "Existing task description",
		Priority:    5,
	}

	task, err := app.taskManager.CreateTask(req)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Verify task is in queued status
	if task.Status != types.TaskStatusQueued {
		t.Errorf("Expected task status 'queued', got '%s'", task.Status)
	}

	// Execute the start command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "start", "existing-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Start command failed: %v", err)
	}

	outputStr := output.String()

	// Check that workspace was created for existing task
	if !strings.Contains(outputStr, "Task 'existing-task' exists but has no workspace. Creating workspace...") {
		t.Errorf("Output should indicate workspace creation for existing task, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "✅ Started working on task: existing-task") {
		t.Errorf("Output should indicate task started, got: %s", outputStr)
	}

	// Verify task was updated
	updatedTask, err := app.taskManager.GetTask(fmt.Sprintf("%d", task.ID))
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}

	if updatedTask.Status != types.TaskStatusInProgress {
		t.Errorf("Expected task status 'in_progress', got '%s'", updatedTask.Status)
	}

	if updatedTask.WorkspaceID == 0 {
		t.Errorf("Expected task to have workspace ID, got 0")
	}

	// Verify workspace was created by checking the workspace path
	if updatedTask.WorkspacePath == "" {
		t.Error("Expected task to have workspace path")
	}

	if _, err := os.Stat(updatedTask.WorkspacePath); os.IsNotExist(err) {
		t.Errorf("Workspace directory should exist: %s", updatedTask.WorkspacePath)
	}
}

// TestTaskDirCommand tests the task dir command
func TestTaskDirCommand(t *testing.T) {
	// Test case: Dir command should return workspace path for task with workspace
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task with workspace using start command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "start", "test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Start command failed: %v", err)
	}

	// Test dir command
	dirCmd := findCommand(app.rootCmd, "task", "dir")
	if dirCmd == nil {
		t.Fatal("Dir command not found")
	}

	// Execute the dir command
	output.Reset()
	app.rootCmd.SetArgs([]string{"task", "dir", "test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Dir command failed: %v", err)
	}

	outputStr := strings.TrimSpace(output.String())

	// Check that the output is a valid path
	if outputStr == "" {
		t.Error("Expected workspace path, got empty string")
	}

	// Verify the path exists
	if _, err := os.Stat(outputStr); os.IsNotExist(err) {
		t.Errorf("Workspace directory does not exist: %s", outputStr)
	}
}

// TestTaskDirCommand_NoWorkspace tests the task dir command with task without workspace
func TestTaskDirCommand_NoWorkspace(t *testing.T) {
	// Test case: Dir command should fail for task without workspace
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task without workspace
	req := &types.CreateTaskRequest{
		Name:        "no-workspace-task",
		Description: "Task without workspace",
	}

	_, err = app.taskManager.CreateTask(req)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Test dir command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "dir", "no-workspace-task"})

	err = app.rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for task without workspace, got nil")
	}

	expectedError := "has no associated workspace"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestTaskCdCommand tests the task cd command
func TestTaskCdCommand(t *testing.T) {
	// Test case: Cd command should change to task workspace directory
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task with workspace using start command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "start", "cd-test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Start command failed: %v", err)
	}

	// Get the current directory before cd
	beforeDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Test cd command
	cdCmd := findCommand(app.rootCmd, "task", "cd")
	if cdCmd == nil {
		t.Fatal("Cd command not found")
	}

	// Execute the cd command
	output.Reset()
	app.rootCmd.SetArgs([]string{"task", "cd", "cd-test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Cd command failed: %v", err)
	}

	outputStr := output.String()

	// Check that the command reported success
	if !strings.Contains(outputStr, "Changed to workspace directory:") {
		t.Errorf("Output should indicate directory change, got: %s", outputStr)
	}

	// Get the current directory after cd
	afterDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Verify that we actually changed directories
	if beforeDir == afterDir {
		t.Error("Expected directory to change, but it didn't")
	}

	// Verify we're in a workspace directory
	if !strings.Contains(afterDir, ".cw/workspaces") {
		t.Errorf("Expected to be in workspace directory, got: %s", afterDir)
	}
}

// TestTaskCdCommand_NoWorkspace tests the task cd command with task without workspace
func TestTaskCdCommand_NoWorkspace(t *testing.T) {
	// Test case: Cd command should fail for task without workspace
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task without workspace
	req := &types.CreateTaskRequest{
		Name:        "no-workspace-task",
		Description: "Task without workspace",
	}

	_, err = app.taskManager.CreateTask(req)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Test cd command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "cd", "no-workspace-task"})

	err = app.rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for task without workspace, got nil")
	}

	expectedError := "has no associated workspace"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// TestTaskGitCommand tests the task git command
func TestTaskGitCommand(t *testing.T) {
	// Test case: Git command should run git commands in task workspace
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task with workspace using start command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "start", "test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Start command failed: %v", err)
	}

	// Test git command
	gitCmd := findCommand(app.rootCmd, "task", "git")
	if gitCmd == nil {
		t.Fatal("Git command not found")
	}

	// Execute the git command
	app.rootCmd.SetArgs([]string{"task", "git", "test-task", "status"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Git command failed: %v", err)
	}

	// The git command should have executed successfully
	// We can't easily capture the output since it goes directly to stdout/stderr
	// But if it didn't error, it means the command worked
}

// TestTaskDeleteCommand_WithWorkspace tests that deleting a task cleans up its workspace
func TestTaskDeleteCommand_WithWorkspace(t *testing.T) {
	// Test case: Delete command should remove task and its workspace
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task with workspace using start command
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{"task", "start", "delete-test-task"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Start command failed: %v", err)
	}

	// Get the task to verify it has a workspace
	tasks, err := app.taskManager.ListTasks(nil)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.WorkspacePath == "" {
		t.Fatal("Expected task to have workspace path")
	}

	// Verify workspace directory exists
	if _, err := os.Stat(task.WorkspacePath); os.IsNotExist(err) {
		t.Fatalf("Workspace directory should exist: %s", task.WorkspacePath)
	}

	// Test delete command
	deleteCmd := findCommand(app.rootCmd, "task", "delete")
	if deleteCmd == nil {
		t.Fatal("Delete command not found")
	}

	// Execute the delete command
	output.Reset()
	app.rootCmd.SetArgs([]string{"task", "delete", "delete-test-task", "--force"})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Delete command failed: %v", err)
	}

	outputStr := output.String()

	// Check that task was deleted
	if !strings.Contains(outputStr, "✅ Task") || !strings.Contains(outputStr, "deleted successfully!") {
		t.Errorf("Output should indicate successful deletion, got: %s", outputStr)
	}

	// Verify task was actually deleted
	tasks, err = app.taskManager.ListTasks(nil)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}

	// Verify workspace directory was also deleted
	if _, err := os.Stat(task.WorkspacePath); !os.IsNotExist(err) {
		t.Errorf("Workspace directory should have been deleted: %s", task.WorkspacePath)
	}
}

// TestTaskUpdateCommand_WithCostTracking tests updating task cost information
func TestTaskUpdateCommand_WithCostTracking(t *testing.T) {
	// Test case: Update command should update task cost information
	tempDir := createTempGitRepo(t)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	app := NewApp("test", "test", "test")
	if app.taskManager == nil {
		t.Fatal("Task manager should be initialized in a Git repository")
	}

	// Create a task first
	req := &types.CreateTaskRequest{
		Name:          "cost-update-task",
		Description:   "Task for cost update testing",
		EstimatedCost: 10.0,
		Currency:      "USD",
	}

	task, err := app.taskManager.CreateTask(req)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Test update command
	updateCmd := findCommand(app.rootCmd, "task", "update")
	if updateCmd == nil {
		t.Fatal("Update command not found")
	}

	// Execute the update command to change cost
	var output bytes.Buffer
	app.rootCmd.SetOut(&output)
	app.rootCmd.SetErr(&output)

	app.rootCmd.SetArgs([]string{
		"task", "update", "cost-update-task",
		"--actual-cost", "15.50",
		"--actual-minutes", "120",
	})

	err = app.rootCmd.Execute()
	if err != nil {
		t.Fatalf("Update command failed: %v", err)
	}

	outputStr := output.String()

	// Check that task was updated
	if !strings.Contains(outputStr, "✅ Task updated successfully!") {
		t.Errorf("Output should indicate successful update, got: %s", outputStr)
	}

	// Verify task was actually updated
	updatedTask, err := app.taskManager.GetTask(fmt.Sprintf("%d", task.ID))
	if err != nil {
		t.Fatalf("Failed to get updated task: %v", err)
	}

	if updatedTask.ActualCost != 15.50 {
		t.Errorf("Expected actual cost 15.50, got %f", updatedTask.ActualCost)
	}

	if updatedTask.ActualMinutes != 120 {
		t.Errorf("Expected actual minutes 120, got %d", updatedTask.ActualMinutes)
	}

	// Verify estimated cost and currency remain unchanged
	if updatedTask.EstimatedCost != 10.0 {
		t.Errorf("Expected estimated cost 10.0, got %f", updatedTask.EstimatedCost)
	}

	if updatedTask.Currency != "USD" {
		t.Errorf("Expected currency USD, got %s", updatedTask.Currency)
	}
}

// TestTaskAlias tests that the t alias works correctly
func TestTaskAlias(t *testing.T) {
	// Test case: t alias should be properly configured
	app := NewApp("test", "test", "test")

	// Find the t command
	var tCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "t" {
			tCmd = cmd
			break
		}
	}

	if tCmd == nil {
		t.Fatal("t command not found")
	}

	// Check that t command has the same subcommands as task
	var taskCmd *cobra.Command
	for _, cmd := range app.rootCmd.Commands() {
		if cmd.Use == "task" {
			taskCmd = cmd
			break
		}
	}

	if taskCmd == nil {
		t.Fatal("task command not found")
	}

	// Check that t has the same number of subcommands as task
	if len(tCmd.Commands()) != len(taskCmd.Commands()) {
		t.Errorf("Expected t to have %d subcommands, got %d", len(taskCmd.Commands()), len(tCmd.Commands()))
	}

	// Check that t has the key subcommands
	var hasStart, hasNext, hasCd, hasDir, hasGit bool
	for _, subCmd := range tCmd.Commands() {
		if strings.HasPrefix(subCmd.Use, "start") {
			hasStart = true
		}
		if strings.HasPrefix(subCmd.Use, "next") {
			hasNext = true
		}
		if strings.HasPrefix(subCmd.Use, "cd") {
			hasCd = true
		}
		if strings.HasPrefix(subCmd.Use, "dir") {
			hasDir = true
		}
		if strings.HasPrefix(subCmd.Use, "git") {
			hasGit = true
		}
	}

	if !hasStart {
		t.Error("t command missing start subcommand")
	}

	if !hasNext {
		t.Error("t command missing next subcommand")
	}

	if !hasCd {
		t.Error("t command missing cd subcommand")
	}

	if !hasDir {
		t.Error("t command missing dir subcommand")
	}

	if !hasGit {
		t.Error("t command missing git subcommand")
	}
}
