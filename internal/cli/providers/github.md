# GitHub Authentication Help

## What GitHub is used for:
- Reading and commenting on pull requests
- Reading and commenting on issues
- Creating pull requests linked to issues
- Filtering issues by assignee
- Checking build status and merge status of pull requests

## How to get a GitHub Personal Access Token:
1. Go to GitHub.com and sign in to your account
2. Click your profile picture → Settings
3. Scroll down to 'Developer settings' (bottom left)
4. Click 'Personal access tokens' → 'Fine-grained tokens'
5. Click 'Generate new token'
6. Give your token a descriptive name (e.g., "Cowork CLI")
7. Set an appropriate expiration date
8. Select the repositories you want to grant access to
9. Under "Repository permissions", select the following:
   - **Pull requests**: Read and write
   - **Issues**: Read and write
   - **Commit statuses**: Read
10. Click 'Generate token'
11. Copy the token immediately (you won't see it again!)

## How to use the login command:
```bash
# Using token method (recommended):
coork config provider github login --token YOUR_TOKEN_HERE

# Interactive mode (will prompt for token):
cowork config provider github login --method token

# For project-specific authentication:
cowork conifg provider github login --token YOUR_TOKEN_HERE --scope project

# Using basic authentication:
cowork config provider github login --method basic --username YOUR_USERNAME --password YOUR_PASSWORD
```

## Useful links:
- GitHub Personal Access Tokens: https://github.com/settings/tokens
- GitHub API Documentation: https://docs.github.com/en/rest
- GitHub Fine-grained Tokens: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-fine-grained-personal-access-token

## Security notes:
- Never share your personal access token
- These minimal permissions ensure cowork can only read/write PRs and issues, and read commit statuses
- Regularly rotate your tokens
- The fine-grained token approach provides better security than classic tokens
