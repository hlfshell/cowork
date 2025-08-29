package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/hlfshell/cowork/internal/types"
)

const (
	// WorkflowsFileName is the name of the workflows file
	WorkflowsFileName = "workflows.json"

	// WorkflowEventsFileName is the name of the workflow events file
	WorkflowEventsFileName = "workflow_events.json"

	// WorkflowLocksFileName is the name of the workflow locks file
	WorkflowLocksFileName = "workflow_locks.json"

	// DefaultLockTimeout is the default timeout for workflow locks
	DefaultLockTimeout = 30 * time.Minute

	// WatchdogInterval is the interval for the watchdog timer
	WatchdogInterval = 5 * time.Minute
)

// WorkflowManager manages workflows and their state transitions
type WorkflowManager struct {
	// Path to the .cw directory
	cwDir string

	// Path to the workflows file
	workflowsFilePath string

	// Path to the workflow events file
	eventsFilePath string

	// Path to the workflow locks file
	locksFilePath string

	// In-memory cache of workflows
	workflows map[string]*types.Workflow

	// In-memory cache of events
	events map[string]*types.WorkflowEvent

	// In-memory cache of locks
	locks map[string]*types.WorkflowLock

	// Mutex for thread safety
	mu sync.RWMutex

	// Watchdog timer for cleaning up expired locks
	watchdogTicker *time.Ticker
	watchdogDone   chan bool
}

// NewWorkflowManager creates a new workflow manager
func NewWorkflowManager(cwDir string) (*WorkflowManager, error) {
	if cwDir == "" {
		return nil, fmt.Errorf("cw directory path is required")
	}

	// Check if .cw directory exists and is initialized
	if _, err := os.Stat(cwDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("cowork project not initialized. Run 'cowork init' first")
	}

	// Check if .cw directory contains required files (indicating it's initialized)
	// Look for either config.json or tasks.json to indicate initialization
	configFile := filepath.Join(cwDir, "config.json")
	tasksFile := filepath.Join(cwDir, "tasks.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if _, err := os.Stat(tasksFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("cowork project not properly initialized. Run 'cowork init' first")
		}
	}

	manager := &WorkflowManager{
		cwDir:             cwDir,
		workflowsFilePath: filepath.Join(cwDir, WorkflowsFileName),
		eventsFilePath:    filepath.Join(cwDir, WorkflowEventsFileName),
		locksFilePath:     filepath.Join(cwDir, WorkflowLocksFileName),
		workflows:         make(map[string]*types.Workflow),
		events:            make(map[string]*types.WorkflowEvent),
		locks:             make(map[string]*types.WorkflowLock),
		watchdogDone:      make(chan bool),
	}

	// Load existing workflows, events, and locks
	if err := manager.loadWorkflows(); err != nil {
		return nil, fmt.Errorf("failed to load workflows: %w", err)
	}

	if err := manager.loadEvents(); err != nil {
		return nil, fmt.Errorf("failed to load events: %w", err)
	}

	if err := manager.loadLocks(); err != nil {
		return nil, fmt.Errorf("failed to load locks: %w", err)
	}

	// Start watchdog timer
	manager.startWatchdog()

	return manager, nil
}

// Close stops the workflow manager and cleans up resources
func (wm *WorkflowManager) Close() error {
	// Stop watchdog timer
	if wm.watchdogTicker != nil {
		wm.watchdogTicker.Stop()
		close(wm.watchdogDone)
	}

	return nil
}

// startWatchdog starts the watchdog timer for cleaning up expired locks
func (wm *WorkflowManager) startWatchdog() {
	wm.watchdogTicker = time.NewTicker(WatchdogInterval)
	go func() {
		for {
			select {
			case <-wm.watchdogTicker.C:
				wm.cleanupExpiredLocks()
			case <-wm.watchdogDone:
				return
			}
		}
	}()
}

// cleanupExpiredLocks removes expired locks
func (wm *WorkflowManager) cleanupExpiredLocks() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	var expiredLocks []string
	for workflowID, lock := range wm.locks {
		if lock.IsExpired() {
			expiredLocks = append(expiredLocks, workflowID)
		}
	}

	for _, workflowID := range expiredLocks {
		delete(wm.locks, workflowID)
	}

	if len(expiredLocks) > 0 {
		wm.saveLocksUnlocked()
	}
}

// loadWorkflows loads all workflows from the workflows file
func (wm *WorkflowManager) loadWorkflows() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Check if workflows file exists
	if _, err := os.Stat(wm.workflowsFilePath); os.IsNotExist(err) {
		// File doesn't exist, start with empty workflows
		return nil
	}

	// Read the workflows file
	file, err := os.Open(wm.workflowsFilePath)
	if err != nil {
		return fmt.Errorf("failed to open workflows file: %w", err)
	}
	defer file.Close()

	// Decode the workflows from JSON
	var workflows []*types.Workflow
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&workflows); err != nil {
		return fmt.Errorf("failed to decode workflows: %w", err)
	}

	// Load workflows into memory
	for _, workflow := range workflows {
		wm.workflows[fmt.Sprintf("%d", workflow.ID)] = workflow
	}

	return nil
}

// saveWorkflows saves all workflows to the workflows file
func (wm *WorkflowManager) saveWorkflows() error {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// Convert workflows map to slice
	var workflows []*types.Workflow
	for _, workflow := range wm.workflows {
		workflows = append(workflows, workflow)
	}

	// Create temporary file
	tempFile := wm.workflowsFilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary workflows file: %w", err)
	}
	defer file.Close()

	// Encode workflows to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(workflows); err != nil {
		return fmt.Errorf("failed to encode workflows: %w", err)
	}

	// Close file before renaming
	file.Close()

	// Rename temporary file to actual file
	if err := os.Rename(tempFile, wm.workflowsFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary workflows file: %w", err)
	}

	return nil
}

// loadEvents loads all events from the events file
func (wm *WorkflowManager) loadEvents() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Check if events file exists
	if _, err := os.Stat(wm.eventsFilePath); os.IsNotExist(err) {
		// File doesn't exist, start with empty events
		return nil
	}

	// Read the events file
	file, err := os.Open(wm.eventsFilePath)
	if err != nil {
		return fmt.Errorf("failed to open events file: %w", err)
	}
	defer file.Close()

	// Decode the events from JSON
	var events []*types.WorkflowEvent
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&events); err != nil {
		return fmt.Errorf("failed to decode events: %w", err)
	}

	// Load events into memory
	for _, event := range events {
		wm.events[event.ID] = event
	}

	return nil
}

// saveEvents saves all events to the events file
func (wm *WorkflowManager) saveEvents() error {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// Convert events map to slice
	var events []*types.WorkflowEvent
	for _, event := range wm.events {
		events = append(events, event)
	}

	// Create temporary file
	tempFile := wm.eventsFilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary events file: %w", err)
	}
	defer file.Close()

	// Encode events to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(events); err != nil {
		return fmt.Errorf("failed to encode events: %w", err)
	}

	// Close file before renaming
	file.Close()

	// Rename temporary file to actual file
	if err := os.Rename(tempFile, wm.eventsFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary events file: %w", err)
	}

	return nil
}

// loadLocks loads all locks from the locks file
func (wm *WorkflowManager) loadLocks() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Check if locks file exists
	if _, err := os.Stat(wm.locksFilePath); os.IsNotExist(err) {
		// File doesn't exist, start with empty locks
		return nil
	}

	// Read the locks file
	file, err := os.Open(wm.locksFilePath)
	if err != nil {
		return fmt.Errorf("failed to open locks file: %w", err)
	}
	defer file.Close()

	// Decode the locks from JSON
	var locks []*types.WorkflowLock
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&locks); err != nil {
		return fmt.Errorf("failed to decode locks: %w", err)
	}

	// Load locks into memory
	for _, lock := range locks {
		wm.locks[fmt.Sprintf("%d", lock.WorkflowID)] = lock
	}

	return nil
}

// saveLocks saves all locks to the locks file
func (wm *WorkflowManager) saveLocks() error {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// Convert locks map to slice
	var locks []*types.WorkflowLock
	for _, lock := range wm.locks {
		locks = append(locks, lock)
	}

	// Create temporary file
	tempFile := wm.locksFilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary locks file: %w", err)
	}
	defer file.Close()

	// Encode locks to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(locks); err != nil {
		return fmt.Errorf("failed to encode locks: %w", err)
	}

	// Close file before renaming
	file.Close()

	// Rename temporary file to actual file
	if err := os.Rename(tempFile, wm.locksFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary locks file: %w", err)
	}

	return nil
}

// CreateWorkflow creates a new workflow
func (wm *WorkflowManager) CreateWorkflow(req *types.CreateWorkflowRequest) (*types.Workflow, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create workflow request: %w", err)
	}

	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Check if workflow already exists for this issue
	for _, workflow := range wm.workflows {
		if workflow.Owner == req.Owner && workflow.Repo == req.Repo && workflow.IssueID == req.IssueID {
			return workflow, nil // Workflow already exists, return existing workflow
		}
	}

	// Generate unique ID
	workflowID := types.GenerateWorkflowID()

	// Create the workflow
	now := time.Now()
	workflow := &types.Workflow{
		ID:          workflowID,
		Owner:       req.Owner,
		Repo:        req.Repo,
		IssueID:     req.IssueID,
		BaseBranch:  req.BaseBranch,
		State:       types.WorkflowStateQueued,
		Provider:    req.Provider,
		Config:      req.Config,
		LastEventTS: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		ErrorCount:  0,
		Metadata:    make(map[string]string),
		LockTimeout: now.Add(DefaultLockTimeout),
	}

	// Set task ID if provided
	if req.TaskID != 0 {
		workflow.TaskID = req.TaskID
	}

	// Add to memory
	wm.workflows[fmt.Sprintf("%d", workflowID)] = workflow

	// Save to file
	if err := wm.saveWorkflowsUnlocked(); err != nil {
		// Remove from memory if save failed
		delete(wm.workflows, fmt.Sprintf("%d", workflowID))
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	return workflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (wm *WorkflowManager) GetWorkflow(workflowID string) (*types.Workflow, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workflow, exists := wm.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	return workflow, nil
}

// GetWorkflowByIssue retrieves a workflow by issue information
func (wm *WorkflowManager) GetWorkflowByIssue(owner, repo string, issueID int) (*types.Workflow, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for _, workflow := range wm.workflows {
		if workflow.Owner == owner && workflow.Repo == repo && workflow.IssueID == issueID {
			return workflow, nil
		}
	}

	return nil, fmt.Errorf("workflow not found for issue %s/%s#%d", owner, repo, issueID)
}

// ListWorkflows returns all workflows
func (wm *WorkflowManager) ListWorkflows() ([]*types.Workflow, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var workflows []*types.Workflow
	for _, workflow := range wm.workflows {
		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

// ListWorkflowsByState returns workflows filtered by state
func (wm *WorkflowManager) ListWorkflowsByState(state types.WorkflowState) ([]*types.Workflow, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var workflows []*types.Workflow
	for _, workflow := range wm.workflows {
		if workflow.State == state {
			workflows = append(workflows, workflow)
		}
	}

	return workflows, nil
}

// UpdateWorkflow updates a workflow
func (wm *WorkflowManager) UpdateWorkflow(req *types.UpdateWorkflowRequest) (*types.Workflow, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid update workflow request: %w", err)
	}

	wm.mu.Lock()
	defer wm.mu.Unlock()

	workflow, exists := wm.workflows[fmt.Sprintf("%d", req.WorkflowID)]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %d", req.WorkflowID)
	}

	// Update fields if provided
	if req.State != nil {
		if !workflow.State.CanTransitionTo(*req.State) {
			return nil, fmt.Errorf("invalid state transition from %s to %s", workflow.State, *req.State)
		}
		workflow.State = *req.State
	}

	if req.BranchName != nil {
		workflow.BranchName = *req.BranchName
	}

	if req.PRNumber != nil {
		workflow.PRNumber = req.PRNumber
	}

	if req.TaskID != nil {
		workflow.TaskID = *req.TaskID
	}

	if req.WorkspaceID != nil {
		workflow.WorkspaceID = *req.WorkspaceID
	}

	if req.ErrorCount != nil {
		workflow.ErrorCount = *req.ErrorCount
	}

	if req.LastError != nil {
		workflow.LastError = *req.LastError
	}

	// Update timestamps
	workflow.UpdatedAt = time.Now()
	workflow.LastEventTS = time.Now()

	// Set started/ended timestamps based on state
	if workflow.StartedAt == nil && (workflow.State == types.WorkflowStateWorkspaceReady || workflow.State == types.WorkflowStateImplementing) {
		now := time.Now()
		workflow.StartedAt = &now
	}

	if workflow.EndedAt == nil && workflow.State.IsTerminal() {
		now := time.Now()
		workflow.EndedAt = &now
	}

	// Save to file
	if err := wm.saveWorkflowsUnlocked(); err != nil {
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	return workflow, nil
}

// DeleteWorkflow removes a workflow
func (wm *WorkflowManager) DeleteWorkflow(workflowID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.workflows[workflowID]; !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	delete(wm.workflows, workflowID)

	// Save to file
	if err := wm.saveWorkflowsUnlocked(); err != nil {
		return fmt.Errorf("failed to save workflows: %w", err)
	}

	return nil
}

// LockWorkflow attempts to acquire a lock on a workflow
func (wm *WorkflowManager) LockWorkflow(workflowID, lockedBy string, timeout time.Duration) (*types.WorkflowLock, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Check if workflow exists
	if _, exists := wm.workflows[workflowID]; !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Check if workflow is already locked
	if existingLock, exists := wm.locks[workflowID]; exists {
		if existingLock.IsValid() {
			return nil, fmt.Errorf("workflow %s is already locked by %s until %s", workflowID, existingLock.LockedBy, existingLock.LockTimeout.Format(time.RFC3339))
		}
		// Lock is expired, remove it
		delete(wm.locks, workflowID)
	}

	// Get current process ID
	processID := os.Getpid()

	// Create new lock
	now := time.Now()
	workflowIDInt, _ := strconv.Atoi(workflowID)
	lock := &types.WorkflowLock{
		WorkflowID:  workflowIDInt,
		LockedBy:    lockedBy,
		LockedAt:    now,
		LockTimeout: now.Add(timeout),
		ProcessID:   processID,
	}

	// Add to memory
	wm.locks[workflowID] = lock

	// Save to file
	if err := wm.saveLocksUnlocked(); err != nil {
		// Remove from memory if save failed
		delete(wm.locks, workflowID)
		return nil, fmt.Errorf("failed to save lock: %w", err)
	}

	return lock, nil
}

// UnlockWorkflow releases a lock on a workflow
func (wm *WorkflowManager) UnlockWorkflow(workflowID, lockedBy string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	lock, exists := wm.locks[workflowID]
	if !exists {
		return fmt.Errorf("no lock found for workflow: %s", workflowID)
	}

	if lock.LockedBy != lockedBy {
		return fmt.Errorf("workflow %s is locked by %s, not %s", workflowID, lock.LockedBy, lockedBy)
	}

	delete(wm.locks, workflowID)

	// Save to file
	if err := wm.saveLocksUnlocked(); err != nil {
		return fmt.Errorf("failed to save locks: %w", err)
	}

	return nil
}

// ForceUnlockWorkflow forcefully releases a lock on a workflow (admin override)
func (wm *WorkflowManager) ForceUnlockWorkflow(workflowID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.locks[workflowID]; !exists {
		return fmt.Errorf("no lock found for workflow: %s", workflowID)
	}

	delete(wm.locks, workflowID)

	// Save to file
	if err := wm.saveLocksUnlocked(); err != nil {
		return fmt.Errorf("failed to save locks: %w", err)
	}

	return nil
}

// GetWorkflowLock retrieves the lock for a workflow
func (wm *WorkflowManager) GetWorkflowLock(workflowID string) (*types.WorkflowLock, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	lock, exists := wm.locks[workflowID]
	if !exists {
		return nil, fmt.Errorf("no lock found for workflow: %s", workflowID)
	}

	return lock, nil
}

// IsWorkflowLocked checks if a workflow is currently locked
func (wm *WorkflowManager) IsWorkflowLocked(workflowID string) (bool, *types.WorkflowLock) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	lock, exists := wm.locks[workflowID]
	if !exists {
		return false, nil
	}

	if lock.IsExpired() {
		return false, nil
	}

	return true, lock
}

// CreateEvent creates a new workflow event
func (wm *WorkflowManager) CreateEvent(eventType, provider, owner, repo string, issueID int, data map[string]interface{}) (*types.WorkflowEvent, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Generate unique ID
	eventID := fmt.Sprintf("event-%d", time.Now().UnixNano())

	// Create the event
	now := time.Now()
	event := &types.WorkflowEvent{
		ID:        eventID,
		Type:      eventType,
		Provider:  provider,
		Timestamp: now,
		Owner:     owner,
		Repo:      repo,
		IssueID:   issueID,
		Data:      data,
		Processed: false,
	}

	// Add to memory
	wm.events[eventID] = event

	// Save to file
	if err := wm.saveEventsUnlocked(); err != nil {
		// Remove from memory if save failed
		delete(wm.events, eventID)
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	return event, nil
}

// GetEvent retrieves an event by ID
func (wm *WorkflowManager) GetEvent(eventID string) (*types.WorkflowEvent, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	event, exists := wm.events[eventID]
	if !exists {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}

	return event, nil
}

// ListUnprocessedEvents returns all unprocessed events
func (wm *WorkflowManager) ListUnprocessedEvents() ([]*types.WorkflowEvent, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var events []*types.WorkflowEvent
	for _, event := range wm.events {
		if !event.Processed {
			events = append(events, event)
		}
	}

	return events, nil
}

// MarkEventProcessed marks an event as processed
func (wm *WorkflowManager) MarkEventProcessed(eventID, workflowID string, errorMsg string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	event, exists := wm.events[eventID]
	if !exists {
		return fmt.Errorf("event not found: %s", eventID)
	}

	event.Processed = true
	event.JobID = workflowID
	if errorMsg != "" {
		event.Error = errorMsg
	}

	// Save to file
	if err := wm.saveEventsUnlocked(); err != nil {
		return fmt.Errorf("failed to save events: %w", err)
	}

	return nil
}

// saveWorkflowsUnlocked saves workflows without acquiring the lock (assumes lock is already held)
func (wm *WorkflowManager) saveWorkflowsUnlocked() error {
	// Convert workflows map to slice
	var workflows []*types.Workflow
	for _, workflow := range wm.workflows {
		workflows = append(workflows, workflow)
	}

	// Create temporary file
	tempFile := wm.workflowsFilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary workflows file: %w", err)
	}
	defer file.Close()

	// Encode workflows to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(workflows); err != nil {
		return fmt.Errorf("failed to encode workflows: %w", err)
	}

	// Close file before renaming
	file.Close()

	// Rename temporary file to actual file
	if err := os.Rename(tempFile, wm.workflowsFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary workflows file: %w", err)
	}

	return nil
}

// saveEventsUnlocked saves events without acquiring the lock (assumes lock is already held)
func (wm *WorkflowManager) saveEventsUnlocked() error {
	// Convert events map to slice
	var events []*types.WorkflowEvent
	for _, event := range wm.events {
		events = append(events, event)
	}

	// Create temporary file
	tempFile := wm.eventsFilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary events file: %w", err)
	}
	defer file.Close()

	// Encode events to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(events); err != nil {
		return fmt.Errorf("failed to encode events: %w", err)
	}

	// Close file before renaming
	file.Close()

	// Rename temporary file to actual file
	if err := os.Rename(tempFile, wm.eventsFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary events file: %w", err)
	}

	return nil
}

// saveLocksUnlocked saves locks without acquiring the lock (assumes lock is already held)
func (wm *WorkflowManager) saveLocksUnlocked() error {
	// Convert locks map to slice
	var locks []*types.WorkflowLock
	for _, lock := range wm.locks {
		locks = append(locks, lock)
	}

	// Create temporary file
	tempFile := wm.locksFilePath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary locks file: %w", err)
	}
	defer file.Close()

	// Encode locks to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(locks); err != nil {
		return fmt.Errorf("failed to encode locks: %w", err)
	}

	// Close file before renaming
	file.Close()

	// Rename temporary file to actual file
	if err := os.Rename(tempFile, wm.locksFilePath); err != nil {
		return fmt.Errorf("failed to rename temporary locks file: %w", err)
	}

	return nil
}
