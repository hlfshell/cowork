# Cowork Architecture Summary

## Workspace-Task Relationship

The Cowork application now implements the following architecture for workspaces and tasks:

### 1. Workspace Creation Rules

**Task-Created Workspaces:**
- When a workspace is created by a task, it shares the same ID as that task
- This ensures a 1:1 relationship between tasks and their workspaces
- Example: Task ID 5 creates Workspace ID 5

**Standalone Workspaces:**
- When a workspace is created without a task, it gets a unique ID from a separate sequence
- Standalone workspaces start from ID 1000 (w1, w2, etc. in practice)
- These workspaces can never be associated with a task
- Example: Standalone workspaces get IDs 1000, 1001, 1002, etc.

### 2. ID Generation System

The application uses two separate ID generators:

- **Shared Generator**: Used for tasks, task-created workspaces, and workflows
- **Workspace Generator**: Used only for standalone workspaces (starts from 1000)

This ensures:
- Task IDs: 1, 2, 3, 4, 5, ...
- Task-created workspace IDs: Same as their task ID
- Standalone workspace IDs: 1000, 1001, 1002, 1003, ...
- Workflow IDs: Continue from the shared sequence

### 3. Task-Workspace Association

**Tasks can optionally create workspaces:**
- Tasks don't necessarily create workspaces when created
- Workspaces are created when tasks are started (status changes to `in_progress`)
- The `CreateWorkspaceForTask()` method handles this creation

**Workspace-Task Relationship:**
- Task-created workspaces have `TaskID` field set to the task's ID
- Task-created workspaces have `IsTaskWorkspace` field set to `true`
- Standalone workspaces have `TaskID` field set to 0
- Standalone workspaces have `IsTaskWorkspace` field set to `false`

### 4. Provider Interface

The application provides a generic `CoworkProvider` interface that supports:
- **GitHub**: Fully implemented with `GitHubCoworkProvider`
- **GitLab**: Placeholder implementation
- **Bitbucket**: Placeholder implementation

Each provider implements:
- Issue scanning and task creation
- Workspace creation for tasks
- Pull request creation and management
- Branch name generation
- Task-provider synchronization

### 5. Manager Class

The `workflow.Manager` class orchestrates the common flow:

**Task Creation from Issues:**
- Scans open issues assigned to the current user
- Creates tasks from issues that don't already have associated tasks
- Links tasks to their source issues via ticket IDs

**Workspace Creation:**
- Creates workspaces when tasks are started
- Generates appropriate branch names based on issue titles
- Sets up Git repositories with proper branching

**Pull Request Management:**
- Creates pull requests when tasks are completed
- Links pull requests back to original issues
- Handles PR updates and review processes

### 6. Key Methods

**Task Management:**
- `CreateTask()` - Creates a new task
- `CreateWorkspaceForTask()` - Creates a workspace for a specific task
- `UpdateTask()` - Updates task status and metadata

**Workspace Management:**
- `CreateWorkspace()` - Creates a workspace (task-created or standalone)
- `GetWorkspace()` - Retrieves workspace by ID
- `ListWorkspaces()` - Lists all workspaces

**Provider Operations:**
- `ScanOpenIssues()` - Finds issues that should become tasks
- `CreateTaskFromIssue()` - Converts provider issues to tasks
- `CreateWorkspaceForTask()` - Creates workspace with proper task association
- `CreatePullRequestForTask()` - Creates PR for completed task

### 7. Data Flow Example

1. **Issue Scanning**: Provider scans GitHub issues assigned to user
2. **Task Creation**: Creates task from issue (ID: 5)
3. **Task Start**: When task status changes to `in_progress`
4. **Workspace Creation**: Creates workspace with same ID (ID: 5)
5. **Development**: User works in workspace
6. **Task Completion**: When task is marked as completed
7. **PR Creation**: Creates pull request linking back to original issue

### 8. Benefits of This Architecture

- **Clear Separation**: Task-created and standalone workspaces are clearly distinguished
- **No ID Conflicts**: Separate ID ranges prevent conflicts
- **Flexible Workflow**: Tasks can exist without workspaces, workspaces can exist without tasks
- **Provider Agnostic**: Same interface works for GitHub, GitLab, and Bitbucket
- **Automated Flow**: Manager class handles the complete workflow from issue to PR
- **Traceability**: Full traceability from issue → task → workspace → PR

This architecture ensures that the application works exactly as specified in the requirements while maintaining flexibility and avoiding ID conflicts.
