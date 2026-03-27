# Cowork Workflow System

## Overview

The Cowork Workflow System implements the complete "Assigned-Issue → Auto-PR Loop" as specified in the requirements. This system automatically manages issues assigned to Cowork's service account, from creating working branches and implementing changes with a coding agent, to opening PRs, iterating on human feedback, and finally handling merge/closure.

**Important**: The workflow system is **project-specific** and only works when you're inside a `.cw` initialized directory. All workflows are tied to the current project and repository.

## Architecture

### Core Components

1. **Workflow** - Represents an auto-PR workflow for a specific issue
2. **WorkflowManager** - Manages workflow persistence and state transitions
3. **WorkflowEngine** - Orchestrates the complete workflow lifecycle
4. **CoworkProvider** - Provider-agnostic interface for Git operations
5. **Lock System** - Prevents concurrent processing with watchdog timers

### State Machine

```
QUEUED → WORKSPACE_READY → IMPLEMENTING → PR_OPEN
                                    ↓
                              REVISING → PR_OPEN
                                    ↓
                              MERGED | CLOSED | ABORTED
```

### Key Features

- **Project-Specific**: Only works in `.cw` initialized directories
- **Idempotency**: Deduplicate by (repo, issue_id)
- **Lock Management**: Prevents concurrent processing with automatic cleanup
- **State Persistence**: All workflows stored in `.cw/workflows.json`
- **Provider Agnostic**: Supports GitHub, GitLab, Bitbucket
- **Task Integration**: Seamlessly integrates with existing task system
- **Error Handling**: Robust error handling with retry mechanisms

## File Structure

```
internal/
├── workflow/
│   ├── workflow_manager.go    # Workflow persistence and management
│   └── engine.go             # Workflow orchestration engine
├── types/
│   └── workflow.go           # Workflow types and state definitions
└── cli/
    └── workflow.go           # CLI commands for workflow management
```

## CLI Commands

**Prerequisites**: All workflow commands must be run from within a `.cw` initialized directory. If you get an error about the project not being initialized, run `cowork init` first.

**Project Detection**: Workflow commands automatically detect the current project from the Git remote origin. The system supports both SSH (`git@github.com:owner/repo.git`) and HTTPS (`https://github.com/owner/repo.git`) remote formats.

### Workflow Management

```bash
# Create a new workflow from an issue (uses current project)
cw workflow create issue-id [--provider github] [--base-branch main]

# Create a workflow from an existing task (uses current project)
cw workflow create-from-task task-id [--provider github] [--base-branch main]

# List all workflows
cw workflow list

# Show detailed workflow information
cw workflow show workflow-id

# Process a single workflow
cw workflow process-workflow workflow-id

# Run workflows continuously (uses current project)
cw workflow run-workflows [--provider github] [--poll-interval 30s] [--once]
```

### Lock Management

```bash
# List active workflow locks
cw workflow lock list

# Force unlock a workflow (admin override)
cw workflow lock force-unlock workflow-id

# Cleanup expired locks (automatic via watchdog)
cw workflow lock cleanup
```

### Legacy Commands (Still Available)

```bash
# Scan and create tasks from issues (uses current project)
cw workflow scan [--provider github]

# Process queued tasks (uses current project)
cw workflow process [--provider github]

# Complete task and create PR (uses current project)
cw workflow complete task-id [--provider github]

# Scan PRs for updates (uses current project)
cw workflow scan-prs [--provider github]

# Sync task statuses (uses current project)
cw workflow sync [--provider github]

# Run full workflow (uses current project)
cw workflow run [--provider github]
```

## Workflow Lifecycle

### 1. Issue Discovery
- Issue assigned to Cowork service account
- Workflow created in `QUEUED` state
- Idempotent: won't create duplicate workflows

### 2. Workspace Preparation
- Creates isolated workspace
- Sets up feature branch (`feature/<issue-key>-<slug>`)
- Syncs with base branch
- Transitions to `WORKSPACE_READY`

### 3. Implementation
- Creates/updates associated task
- Agent works on task
- Monitors task completion
- Transitions to `IMPLEMENTING`

### 4. Pull Request Creation
- Pushes feature branch
- Creates PR with issue title
- Links PR to issue
- Applies `cowork:active` label
- Transitions to `PR_OPEN`

### 5. Review Loop
- Monitors PR for comments/reviews
- Classifies feedback intent (ASK/CHANGE/BLOCKER)
- Handles questions vs. requested changes
- Transitions between `PR_OPEN` and `REVISING`

### 6. Completion
- On merge: transitions to `MERGED`
- On close: transitions to `CLOSED`
- Cleans up workspace and remote branch
- Removes `cowork:active` label

## Lock System

### Features
- **Process Isolation**: Only one process can work on a workflow at a time
- **Automatic Cleanup**: Watchdog timer removes expired locks every 5 minutes
- **Manual Override**: Admin can force unlock stuck workflows
- **Timeout Protection**: Default 30-minute lock timeout

### Lock States
```go
type WorkflowLock struct {
    WorkflowID  string    `json:"workflow_id"`
    LockedBy    string    `json:"locked_by"`
    LockedAt    time.Time `json:"locked_at"`
    LockTimeout time.Time `json:"lock_timeout"`
    ProcessID   int       `json:"process_id"`
}
```

## Configuration

### Workflow Configuration
```go
type WorkflowConfig struct {
    BranchNamingTemplate string        `json:"branch_naming_template"`
    RequiredChecks       []string      `json:"required_checks"`
    SyncStrategy         string        `json:"sync_strategy"` // "rebase" or "merge"
    Labels               []string      `json:"labels"`
    Timeouts             TimeoutConfig `json:"timeouts"`
    Safety               SafetyConfig  `json:"safety"`
}
```

### Default Configuration
- **Branch Naming**: `feature/<issue-key>-<slug>`
- **Sync Strategy**: Prefer rebase, fallback to merge
- **Required Checks**: Project-specific (configurable)
- **Lock Timeout**: 30 minutes
- **Watchdog Interval**: 5 minutes

## Error Handling

### Error Recovery
- **Transient Errors**: Automatic retry with exponential backoff
- **Permanent Errors**: Transition to `ABORTED` state
- **Lock Timeouts**: Automatic cleanup and retry
- **Git Conflicts**: Attempt rebase, fallback to merge

### Error States
- **ABORTED**: Manual intervention required
- **Error Count**: Track consecutive failures
- **Last Error**: Store detailed error message

## Integration Points

### Task System
- Workflows reference tasks via `TaskID`
- Task status changes trigger workflow transitions
- Easy to create workflows from existing tasks

### Workspace System
- Workflows create isolated workspaces
- Workspace paths stored in workflow metadata
- Automatic cleanup on completion

### Git Providers
- Provider-agnostic interface
- Supports GitHub, GitLab, Bitbucket
- Extensible for new providers

## Security Considerations

### Safety Rails
- Never force-push to base branches
- Only touch feature branches
- Validate all Git operations
- Sanitize user inputs

### Authentication
- Uses existing auth system
- Token-based provider authentication
- Secure credential storage

## Monitoring and Logging

### Logging
- Structured logging with emojis for visibility
- State transition logging
- Error logging with context
- Performance metrics

### Monitoring
- Workflow state distribution
- Processing times
- Error rates
- Lock statistics

## Future Enhancements

### Planned Features
- **Webhook Support**: Real-time event processing
- **Agent Integration**: Direct agent communication
- **Advanced Feedback**: AI-powered feedback classification
- **Metrics Dashboard**: Web-based monitoring
- **Configuration UI**: Visual workflow configuration

### Extensibility
- **Custom States**: User-defined workflow states
- **Custom Actions**: Plugin-based workflow actions
- **Multi-Repository**: Cross-repository workflows
- **Dependencies**: Workflow dependencies and ordering

## Usage Examples

### Basic Workflow Creation
```bash
# Create workflow for GitHub issue #123 (uses current project)
cw workflow create 123 --provider github --base-branch main

# Check workflow status
cw workflow show <workflow-id>

# Process workflow manually
cw workflow process-workflow <workflow-id>
```

### Continuous Processing
```bash
# Run workflows continuously for the current project
cw workflow run-workflows --provider github --poll-interval 60s

# Run once and exit
cw workflow run-workflows --once
```

### Lock Management
```bash
# Check for stuck workflows
cw workflow lock list

# Force unlock if needed
cw workflow lock force-unlock <workflow-id>
```

## Troubleshooting

### Common Issues

1. **Workflow Stuck in QUEUED**
   - Check if workspace creation is failing
   - Verify Git provider authentication
   - Check for lock conflicts

2. **PR Creation Fails**
   - Verify branch push permissions
   - Check PR creation permissions
   - Review issue linking requirements

3. **Lock Timeouts**
   - Check for zombie processes
   - Verify system time synchronization
   - Review lock timeout configuration

### Debug Commands
```bash
# Show detailed workflow info
cw workflow show <workflow-id>

# List all locks
cw workflow lock list

# Check workflow state
cw workflow list
```

## Contributing

### Development Guidelines
- Follow existing code patterns
- Add comprehensive tests
- Update documentation
- Use descriptive variable names
- Include error handling

### Testing
- Unit tests for all components
- Integration tests for workflows
- End-to-end tests for complete flows
- Mock external dependencies

This workflow system provides a robust, extensible foundation for automating the complete Git-based development loop, from issue assignment to PR merge, with comprehensive error handling, monitoring, and management capabilities.
