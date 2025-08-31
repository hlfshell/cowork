# Bitbucket Authentication Help

## What Bitbucket is used for:
- Reading and commenting on pull requests
- Reading and commenting on issues
- Creating pull requests linked to issues
- Filtering issues by assignee
- Checking build status and merge status of pull requests

## How to get a Bitbucket App Password:
1. Go to Bitbucket.org and sign in to your account
2. Click your profile picture → Personal settings
3. Click 'App passwords' in the left sidebar
4. Click 'Create app password'
5. Fill in the password details:
   - Label: Give it a descriptive name (e.g., "Cowork CLI")
   - Permissions: Select the following:
     - ✅ Account: Read
     - ✅ Repositories: Read
     - ✅ Pull requests: Read, Write
     - ✅ Issues: Read, Write
     - ✅ Pipelines: Read
6. Click 'Create'
7. Copy the password immediately (you won't see it again!)

## How to use the login command:
```bash
# Using basic auth method (recommended for Bitbucket):
cw config auth provider bitbucket --method basic --username YOUR_USERNAME --password YOUR_APP_PASSWORD

# Interactive mode (will prompt for credentials):
cw config auth provider bitbucket --method basic

# For project-specific authentication:
cw config auth provider bitbucket --method basic --scope project --username YOUR_USERNAME --password YOUR_APP_PASSWORD

# Using token method (if you have a token):
cw config auth provider bitbucket --token YOUR_TOKEN_HERE
```

## Useful links:
- Bitbucket App Passwords: https://bitbucket.org/account/settings/app-passwords/
- Bitbucket API Documentation: https://developer.atlassian.com/cloud/bitbucket/rest/
- Bitbucket Permissions: https://support.atlassian.com/bitbucket-cloud/docs/repository-permissions/

## Security notes:
- Never share your app password
- These minimal permissions ensure cowork can only read/write PRs and issues, and read pipeline statuses
- Regularly rotate your app passwords
- Only grant access to repositories you actually need to work with
