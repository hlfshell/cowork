# Cowork Workflow System

## Overview

The Cowork Workflow System provides a comprehensive integration between Git providers (GitHub, GitLab, Bitbucket) and the Cowork task management system. It automates the entire lifecycle from issue scanning to pull request creation and review handling.

## Architecture

### Core Components

1. **CoworkProvider Interface** (`internal/git/interfaces.go`)
   - Extends the base `GitProvider` interface with task-specific operations
   - Handles the complete workflow from issue to PR
   - Provider-agnostic design supporting GitHub, GitLab, and Bitbucket

2. **WorkflowManager** (`internal/workflow/manager.go`)
   - Orchestrates the complete workflow process
   - Manages task creation, workspace setup, and PR handling
   - Provides both individual step execution and full workflow runs

3. **Provider Implementations**
   - `GitHubCoworkProvider` (`internal/git/providers/github_cowork.go`)
   - Placeholder implementations for GitLab and Bitbucket
   - Each provider implements the complete `CoworkProvider` interface

4. **CLI Integration** (`internal/cli/workflow.go`)
   - Command-line interface for workflow operations
   - Individual commands for each workflow step
   - Full workflow execution command

## Workflow Features

### 1. Issue Scanning and Task Creation

**Feature**: Scan all open issues assigned to the current user and create tasks for them.

**Implementation**:
- `ScanOpenIssues()` - Retrieves open issues from the provider
- `CreateTaskFromIssue()` - Converts provider issues to Cowork tasks
- `GetTaskByIssue()` - Checks for existing tasks to avoid duplicates

**CLI Command**:
```bash
cw workflow scan owner/repo --provider github
```

### 2. Workspace Creation for Active Tasks

**Feature**: When a task is marked as running/started, create a workspace with a branch based on the issue.

**Implementation**:
- `CreateWorkspaceForTask()` - Creates workspace and Git branch
- `GenerateBranchName()` - Creates provider-specific branch names
- Automatic task status updates to "in_progress"

**CLI Command**:
```bash
cw workflow process owner/repo --provider github
```

### 3. Pull Request Creation

**Feature**: When an agent reports completion, create a PR if one doesn't exist and link it to the original issue.

**Implementation**:
- `CreatePullRequestForTask()` - Creates PR with proper linking
- `GetPullRequestForTask()` - Checks for existing PRs
- `LinkPullRequestToIssue()` - Adds comments and references

**CLI Command**:
```bash
cw workflow complete task-id owner/repo --provider github
```

### 4. PR Monitoring and Updates

**Feature**: Scan PRs for updates (comments, reviews) and trigger additional work when needed.

**Implementation**:
- `ScanPullRequestsForTasks()` - Finds PRs associated with known tasks
- `GetPullRequestUpdates()` - Retrieves recent comments and reviews
- `UpdateTaskFromPullRequest()` - Updates tasks with new feedback
- Automatic task reactivation for additional work

**CLI Command**:
```bash
cw workflow scan-prs owner/repo --provider github
```

### 5. Status Synchronization

**Feature**: Sync task statuses back to the provider (update issue labels, etc.).

**Implementation**:
- `SyncTaskStatusToProvider()` - Updates provider with task status
- `GetProviderMetadata()` - Retrieves provider-specific information
- Automatic label updates based on task status

**CLI Command**:
```bash
cw workflow sync owner/repo --provider github
```

## Complete Workflow

**Feature**: Run the entire workflow in sequence.

**Implementation**:
- `RunFullWorkflow()` - Executes all steps in order
- Comprehensive logging and error handling
- Graceful failure handling for individual steps

**CLI Command**:
```bash
cw workflow run owner/repo --provider github
```

## Usage Examples

### Basic Workflow

1. **Initialize and authenticate**:
   ```bash
   cw auth provider login github
   ```

2. **Run the complete workflow**:
   ```bash
   cw workflow run myorg/myrepo --provider github
   ```

### Step-by-Step Workflow

1. **Scan for new issues**:
   ```bash
   cw workflow scan myorg/myrepo --provider github
   ```

2. **Process queued tasks**:
   ```bash
   cw workflow process myorg/myrepo --provider github
   ```

3. **Monitor for PR updates**:
   ```bash
   cw workflow scan-prs myorg/myrepo --provider github
   ```

4. **Sync statuses**:
   ```bash
   cw workflow sync myorg/myrepo --provider github
   ```

### Task Completion

When an agent finishes work on a task:

```bash
cw workflow complete task-uuid myorg/myrepo --provider github
```

## Provider-Specific Features

### GitHub Provider

- **Branch Naming**: `issue-{number}-{title-slug}`
- **PR Titles**: `Fix #{number}: {title}`
- **Issue Linking**: Automatic comment addition
- **Label Updates**: Status-based label management
- **Review Handling**: Full review state management

### GitLab Provider (Planned)

- **Branch Naming**: `issue-{number}-{title-slug}`
- **MR Titles**: `Fix #{number}: {title}`
- **Issue Linking**: Automatic comment addition
- **Label Updates**: Status-based label management

### Bitbucket Provider (Planned)

- **Branch Naming**: `issue-{number}-{title-slug}`
- **PR Titles**: `Fix #{number}: {title}`
- **Issue Linking**: Automatic comment addition
- **Label Updates**: Status-based label management

## Configuration

### Authentication

The workflow system uses the existing authentication system:

```bash
# Set up GitHub authentication
cw auth provider login github

# Set up GitLab authentication
cw auth provider login gitlab

# Set up Bitbucket authentication
cw auth provider login bitbucket
```

### Repository Configuration

The system automatically detects repository information from the current Git repository, but you can specify:

- **Owner/Repo**: Passed as command arguments
- **Provider**: Specified via `--provider` flag
- **Base Branch**: Automatically detected (defaults to "main")

## Error Handling

The workflow system includes comprehensive error handling:

- **Graceful Degradation**: Individual step failures don't stop the entire workflow
- **Detailed Logging**: Each step provides clear status messages
- **Retry Logic**: Automatic retries for transient failures
- **Status Tracking**: Task status updates reflect workflow progress

## Integration Points

### Task Management

- **Task Creation**: Automatic task creation from provider issues
- **Status Updates**: Task status reflects workflow progress
- **Metadata Storage**: Provider information stored in task metadata
- **Workspace Linking**: Tasks linked to their workspaces

### Workspace Management

- **Automatic Creation**: Workspaces created when tasks start
- **Branch Management**: Automatic branch creation and management
- **Repository Cloning**: Automatic repository setup
- **Cleanup**: Workspace cleanup when tasks complete

### Agent Integration (Future)

- **Task Assignment**: Agents can pick up tasks in "in_progress" status
- **Status Updates**: Agents can update task status
- **Completion Handling**: Automatic PR creation when agents complete tasks
- **Feedback Loop**: PR updates trigger additional agent work

## Development Status

### Completed

- ✅ `CoworkProvider` interface definition
- ✅ `GitHubCoworkProvider` implementation
- ✅ `WorkflowManager` orchestration
- ✅ CLI command integration
- ✅ Complete workflow automation
- ✅ Error handling and logging

### In Progress

- 🔄 GitLab provider implementation
- 🔄 Bitbucket provider implementation
- 🔄 Authentication integration
- 🔄 Conversion methods for GitHub types

### Planned

- 📋 Webhook support for real-time updates
- 📋 Advanced PR review handling
- 📋 Multi-repository support
- 📋 Agent integration hooks
- 📋 Performance optimizations

## Contributing

To add support for a new provider:

1. Implement the `CoworkProvider` interface
2. Add conversion methods for provider-specific types
3. Add CLI integration in `createCoworkProvider()`
4. Add authentication support
5. Write comprehensive tests

## Testing

The workflow system includes:

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end workflow testing
- **Mock Providers**: Test implementations for development
- **Error Scenarios**: Comprehensive error condition testing

Run tests with:

```bash
go test ./internal/workflow/...
go test ./internal/git/providers/...
```
