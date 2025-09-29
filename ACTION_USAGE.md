# Discord Commit Notifier Action

Get instant Discord notifications every time you push commits to GitHub! 🚀

## 🎯 Quick Start

### Step 1: Set up Discord Webhook
1. Go to your Discord server settings
2. Navigate to Integrations → Webhooks
3. Create a new webhook and copy the URL

### Step 2: Add the Webhook to GitHub Secrets
1. Go to your repository Settings → Secrets and variables → Actions
2. Add a new secret named `DISCORD_WEBHOOK`
3. Paste your Discord webhook URL as the value

### Step 3: Create Workflow File
Create `.github/workflows/discord-notify.yml` in your repository:

```yaml
name: Discord Commit Notifications
on:
  push:
    branches: [ main, master, develop ]  # Customize branches as needed

jobs:
  notify:
    runs-on: ubuntu-latest
    steps:
      - name: Send Discord Notification
        uses: wilfierd/gh-noftify@v0.1.0
        with:
          discord-webhook: ${{ secrets.DISCORD_WEBHOOK }}
          sha: ${{ github.sha }}
          commit-message: ${{ github.event.head_commit.message }}
          commit-author: ${{ github.event.head_commit.author.name }}
          repository: ${{ github.repository }}
          commit-url: ${{ github.event.head_commit.url }}
          repo-url: ${{ github.event.repository.html_url }}
          actor: ${{ github.actor }}
          branch: ${{ github.ref_name }}
```

That's it! Every push will now send a notification to your Discord channel.

## 📚 Advanced Usage

### With All Options

```yaml
name: Discord Commit Notifications
on:
  push:
    branches: [ main ]

jobs:
  notify:
    runs-on: ubuntu-latest
    steps:
      - name: Send Discord Notification
        uses: wilfierd/gh-noftify@v0.1.0
        with:
          discord-webhook: ${{ secrets.DISCORD_WEBHOOK }}
          github-token: ${{ secrets.GITHUB_TOKEN }}  # For avatar support
          include-avatar: 'true'
          notification-title: '🎉 New Code Deployed!'
          # GitHub context
          sha: ${{ github.sha }}
          commit-message: ${{ github.event.head_commit.message }}
          commit-author: ${{ github.event.head_commit.author.name }}
          repository: ${{ github.repository }}
          commit-url: ${{ github.event.head_commit.url }}
          repo-url: ${{ github.event.repository.html_url }}
          actor: ${{ github.actor }}
          branch: ${{ github.ref_name }}
```

### Multiple Webhooks for Different Branches

```yaml
name: Discord Commit Notifications
on:
  push:

jobs:
  notify:
    runs-on: ubuntu-latest
    steps:
      - name: Determine Webhook
        id: webhook
        run: |
          if [[ "${{ github.ref }}" == "refs/heads/main" ]]; then
            echo "webhook=${{ secrets.DISCORD_WEBHOOK_PROD }}" >> $GITHUB_OUTPUT
          else
            echo "webhook=${{ secrets.DISCORD_WEBHOOK_DEV }}" >> $GITHUB_OUTPUT
          fi
      
      - name: Send Discord Notification
        uses: wilfierd/gh-noftify@v1
        with:
          discord-webhook: ${{ steps.webhook.outputs.webhook }}
```

### Conditional Notifications

```yaml
name: Discord Commit Notifications
on:
  push:
    branches: [ main ]

jobs:
  notify:
    runs-on: ubuntu-latest
    # Only notify for non-dependabot commits
    if: github.actor != 'dependabot[bot]'
    steps:
      - name: Send Discord Notification
        uses: wilfierd/gh-noftify@v1
        with:
          discord-webhook: ${{ secrets.DISCORD_WEBHOOK }}
```

## 🔧 Configuration Options

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `discord-webhook` | Discord webhook URL | ✅ Yes | - |
| `github-token` | GitHub token for API access (for avatars) | ❌ No | `${{ github.token }}` |
| `include-avatar` | Include GitHub user avatar in notification | ❌ No | `'true'` |
| `notification-title` | Custom notification title | ❌ No | `'🔔 New Commit Pushed'` |

## 🚀 Deploy to All Your Repositories

Want to add this to all your repositories at once? Here's a bash script:

```bash
#!/bin/bash
# deploy-notifier.sh

GITHUB_USERNAME="your-username"
REPOS=$(gh repo list $GITHUB_USERNAME --limit 100 --json name -q '.[].name')

for repo in $REPOS; do
  echo "Adding workflow to $repo..."
  
  # Clone repo
  git clone "git@github.com:$GITHUB_USERNAME/$repo.git" temp-repo
  cd temp-repo
  
  # Create workflow directory
  mkdir -p .github/workflows
  
  # Create workflow file
  cat > .github/workflows/discord-notify.yml << 'EOF'
name: Discord Commit Notifications
on:
  push:
    branches: [ main, master ]

jobs:
  notify:
    runs-on: ubuntu-latest
    steps:
      - uses: wilfierd/gh-noftify@v1
        with:
          discord-webhook: ${{ secrets.DISCORD_WEBHOOK }}
EOF
  
  # Commit and push
  git add .github/workflows/discord-notify.yml
  git commit -m "Add Discord commit notifications"
  git push
  
  # Cleanup
  cd ..
  rm -rf temp-repo
  
  echo "✅ Added to $repo"
done

echo "🎉 Done! Remember to add DISCORD_WEBHOOK secret to each repository."
```

## 🔍 Troubleshooting

### Not receiving notifications?

1. **Check workflow runs**: Go to Actions tab in your repository
2. **Verify webhook URL**: Make sure the secret is correctly set
3. **Check branch names**: Ensure your workflow triggers on the correct branches
4. **Review workflow logs**: Click on failed runs to see error messages

### Getting errors?

- **"DISCORD_WEBHOOK environment variable is required"**: The webhook secret is not set
- **"Failed to send commit notification"**: Check if the webhook URL is valid
- **Rate limiting**: Discord webhooks have rate limits; avoid pushing too frequently

## 📝 Examples of What You'll See in Discord

The notification will show:
- Repository name with link
- Branch name
- Commit author
- Commit message
- Direct link to the commit
- Author's GitHub avatar (if enabled)

## 🤝 Contributing

Found a bug or want a new feature? Open an issue or submit a PR at [wilfierd/gh-noftify](https://github.com/wilfierd/gh-noftify)

## 📄 License

This action is available under the MIT License.