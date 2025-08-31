# GitLab Authentication Help

## What GitLab is used for:
- Reading and commenting on merge requests
- Reading and commenting on issues
- Creating merge requests linked to issues
- Filtering issues by assignee
- Checking build status and merge status of merge requests

## How to get a GitLab Personal Access Token:
1. Go to GitLab.com (or your GitLab instance) and sign in
2. Click your profile picture → Preferences
3. Click 'Access Tokens' in the left sidebar
4. Fill in the token details:
   - Token name: Give it a descriptive name (e.g., "Cowork CLI")
   - Expiration date: Set an appropriate expiration
5. Select the following scopes:
   - ✅ api (Access your API)
6. Click 'Create personal access token'
7. Copy the token immediately (you won't see it again!)

## Important: Repository Permissions
After creating the token, you need to ensure the user has the **Reporter** role on the repositories you want to work with. The Reporter role provides:
- Read access to merge requests and issues
- Ability to comment on merge requests and issues
- Ability to create merge requests
- Read access to pipeline statuses

To set repository permissions:
1. Go to your repository/project
2. Navigate to Settings → Members
3. Add your user with the **Reporter** role
4. Or ask your project administrator to grant you the Reporter role

## How to use the login command:
```bash
# Using token method (recommended):
cw config auth provider gitlab --token YOUR_TOKEN_HERE

# Interactive mode (will prompt for token):
cw config auth provider gitlab --method token

# For project-specific authentication:
cw config auth provider gitlab --token YOUR_TOKEN_HERE --scope project

# Using basic authentication:
cw config auth provider gitlab --method basic --username YOUR_USERNAME --password YOUR_PASSWORD
```

## Useful links:
- GitLab Personal Access Tokens: https://gitlab.com/-/profile/personal_access_tokens
- GitLab API Documentation: https://docs.gitlab.com/ee/api/
- GitLab Permissions: https://docs.gitlab.com/ee/user/permissions.html

## Security notes:
- Never share your personal access token
- The Reporter role provides minimal necessary permissions for cowork functionality
- Regularly rotate your tokens
- Only grant access to repositories you actually need to work with
