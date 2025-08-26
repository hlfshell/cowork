# cowork (cw) — Managing AI Coworkers

Spin up AI-powered "coworkers" that each take a task, get their own isolated dev container, and safely code alongside you — without ever messing up your main current git branch.

## ✨ What It Does

Coming soon

## 🚀 Why It's Cool

Think of it as hiring AI interns: they each get their own desk (workspace + container), their own set of tools, and a single task to focus on. You stay in control, review their work, and merge when ready.

## 🔐 Authentication

Cowork supports authentication with Git providers (GitHub, GitLab, Bitbucket) using SSH keys, API tokens, or basic authentication.

### Quick Authentication

```bash
# SSH key authentication (recommended)
cw auth git ssh

# Token authentication
cw auth provider login github --method token --token your-token

# Basic authentication
cw auth provider login github --method basic --username your-username --password your-password

# Project-scoped authentication
cw auth provider login github --scope project
```

### Getting GitHub Tokens

To authenticate with GitHub, you'll need a Personal Access Token:

1. **Go to GitHub Settings**: Visit https://github.com/settings/tokens
2. **Generate New Token**: Click "Generate new token (classic)"
3. **Configure Token**:
   - **Note**: Give it a descriptive name like "Cowork CLI"
   - **Expiration**: Choose an appropriate expiration (30 days, 90 days, or custom)
   - **Scopes**: Select the required permissions:
     - `repo` (Full control of private repositories)
     - `workflow` (Update GitHub Action workflows)
     - `write:packages` (Upload packages to GitHub Package Registry)
     - `delete:packages` (Delete packages from GitHub Package Registry)
     - `read:org` (Read organization data)
4. **Generate Token**: Click "Generate token"
5. **Copy Token**: Copy the token immediately (you won't see it again!)

**Security Note**: Keep your token secure and never commit it to version control.

### Using Your GitHub Token

Once you have your token, you can authenticate with GitHub:

```bash
# Authenticate with GitHub using your token
cw auth provider login github --method token --token ghp_your_token_here

# Test your authentication
cw auth provider test github

# View your authentication status
cw auth show
```

**Example**:
```bash
cw auth provider login github --method token --token ghp_1234567890abcdef1234567890abcdef12345678
```

### Authentication Methods

- **SSH Keys**: Secure key-based authentication (default)
- **API Token**: Personal access tokens for CI/CD and automation
- **Basic Auth**: Username/password for legacy systems

### SSH Key Setup

SSH authentication is the recommended method for Git operations:

```bash
# Configure SSH authentication (interactive)
cw auth git ssh

# Configure SSH authentication with custom key file
cw auth git ssh --key ~/.ssh/my_custom_key

# This will:
# - Set up your SSH key path (default: ~/.ssh/id_rsa)
# - Generate a new key if needed
# - Add the key to SSH agent
# - Test the connection to GitHub
```

### Scopes

- **Global**: Stored in `~/.config/cowork/auth/` (applies to all projects)
- **Project**: Stored in `./.cw/auth/` (applies only to current project)

### Testing Authentication

```bash
# Test Git authentication (SSH and HTTPS)
cw auth git test

# Test provider authentication
cw auth provider test github
```

All credentials are encrypted using AES-256-GCM and stored securely with proper file permissions.

### Troubleshooting Authentication

**Token Authentication Fails**:
- Verify your token is correct and hasn't expired
- Check that you have the required scopes (`repo`, `workflow`, etc.)
- Ensure your token has access to the repositories you're trying to access

**SSH Authentication Fails**:
- Verify your SSH key is added to your GitHub account
- Test your SSH connection: `ssh -T git@github.com`
- Check that your SSH key is in the SSH agent: `ssh-add -l`

**Permission Denied Errors**:
- Ensure your token has the correct scopes for the operations you're performing
- Check that you have access to the specific repositories or organizations
- Verify your GitHub account has the necessary permissions

## 🛠 Quick Start

Coming soon