cowork (cw) — Multi-Agent Repo Orchestrator

A Go-based CLI that lets one developer spin up many isolated, containerized workspaces on the same Git repository, wire them to AI coding agents, and drive them from tasks (GitHub/GitLab/Jira/Linear). It keeps your main checkout pristine while parallel “coworkers” code safely on branches you can review and merge.

🔧 Core Ideas

Isolated workspaces per task: worktree / linked-clone / full-clone.

Containerized dev envs per workspace: deterministic toolchains, separate processes, own ports.

Agent runners: easily point Cursor/Claude/Gemini/etc. at one workspace without touching others.

Task-first workflows: cw task 123 → fetch issue → name branch → create workspace+container → run agent → open PR.

Rules engine: .cwrules defines branch/commit naming, default images, agent bindings, and safety guards.

State & auditability: .cwstate tracks who/what/where for each workspace & container.

🧠 Terminology

Task: a local identifier (e.g., oauth-refresh) you assign to a workspace.

Task: an external issue ID (e.g., GitHub #42) bound to a task/branch.

Workspace: a checked-out copy for one branch (worktree or clone).

Isolation level: how the workspace is materialized:

worktree (lightweight, shared object store)

linked-clone (separate .git, shared objects via --reference)

full-clone (complete independence)

Container: a running dev environment for one workspace (Docker/Podman).

Agent: an external tool (Cursor, Claude CLI, Gemini) that edits code and/or runs commands.

📁 Project Files & Layout