# Real-Time Commit Notifications Setup Guide

⚠️ **DEPRECATED APPROACH**: This documentation describes the old scheduled commit tracking system. 

✅ **NEW APPROACH**: We now use real-time GitHub Actions for instant commit notifications!

## 🚀 New Real-Time System

The new system provides:
- **Instant notifications** - triggers immediately when you push
- **Zero configuration** - just add one workflow file per repository
- **No scheduling needed** - uses GitHub's push event triggers
- **Better reliability** - no missed commits due to timing

### Quick Setup for Real-Time Notifications:

1. **Use the GitHub Action**: Add this to any repository's `.github/workflows/discord-notify.yml`:
   ```yaml
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
   ```

2. **Deploy to all repositories**: Use the included script:
   ```bash
   ./deploy-to-all-repos.sh
   ```

3. **Add webhook secret**: Set `DISCORD_WEBHOOK` secret in each repository

For detailed instructions, see `ACTION_USAGE.md`.

---

# ⚠️ Legacy Scheduled System (Deprecated)

The information below describes the old 2-hour scheduled system which has been **removed** in favor of real-time notifications.

## 🚀 Quick Setup

### Option 1: Track ALL Your Repositories

1. **Set environment variables** (for local testing):
```bash
export TRACK_COMMITS_REALTIME=true
export COMMIT_LOOKBACK_MINUTES=120
```

2. **For GitHub Actions**, add these as repository variables:
   - Go to Settings → Secrets and variables → Actions → Variables
   - Add:
     - `TRACK_COMMITS_REALTIME` = `true`
     - `COMMIT_LOOKBACK_MINUTES` = `120`

### Option 2: Track Specific Repositories Only

1. **Set environment variables** with repository list:
```bash
export TRACK_COMMITS_REALTIME=true
export COMMIT_LOOKBACK_MINUTES=120
export TRACKED_REPOSITORIES="repo1,owner/repo2,repo3"
```

2. **For GitHub Actions**, add as repository variables:
   - `TRACKED_REPOSITORIES` = `"gh-notify,torvalds/linux,facebook/react"`

## ⚙️ Configuration Options

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TRACK_COMMITS_REALTIME` | `false` | Enable real-time commit tracking in instant checks |
| `COMMIT_LOOKBACK_MINUTES` | `120` | How far back to check for commits (in minutes) |
| `TRACKED_REPOSITORIES` | *(empty)* | Comma-separated list of repos to track. Empty = all repos |
| `TRACK_ALL_COMMITS` | `true` | Enable commit tracking in daily digests |

### Repository Format

For `TRACKED_REPOSITORIES`, you can use:
- **Just repo name**: `"my-repo"` (assumes your username)
- **Full path**: `"owner/repo"`
- **Multiple repos**: `"repo1,owner/repo2,my-repo"`

### Examples

```bash
# Track only your personal projects
TRACKED_REPOSITORIES="my-website,my-api,my-cli-tool"

# Track specific popular repos you contribute to
TRACKED_REPOSITORIES="kubernetes/kubernetes,golang/go,microsoft/vscode"

# Mixed - your repos and others
TRACKED_REPOSITORIES="my-app,facebook/react,google/gson"
```

## 📊 How It Works

### Instant Checks (Every 2 hours)
When `TRACK_COMMITS_REALTIME=true`:
1. Fetches commits from last `COMMIT_LOOKBACK_MINUTES` minutes
2. Filters by repositories in `TRACKED_REPOSITORIES` (or all if empty)
3. Checks cache to avoid duplicate notifications
4. Sends Discord notification for new commits only
5. Updates cache to prevent re-notification

### Daily Digests
When `TRACK_ALL_COMMITS=true`:
- **Morning digest**: Shows yesterday's commits for context
- **Evening digest**: Shows today's commits as accomplishments

## 🔔 Discord Notification Format

Commits appear in Discord with:
```
💻 Recent Commits
• [abc1234](link) in **repo-name**
  `feat: added new feature`
• [def5678](link) in **another-repo**
  `fix: resolved critical bug`
```

## 🛠️ Testing Your Configuration

### Local Testing
```bash
# 1. Set up environment
cp .env.example .env
# Edit .env with your configuration

# 2. Test with different lookback periods
export TRACK_COMMITS_REALTIME=true
export COMMIT_LOOKBACK_MINUTES=60  # Last hour
export TRACKED_REPOSITORIES="gh-notify"
CHECK_TYPE=instant ./gh-notify

# 3. Test with all repositories
export TRACKED_REPOSITORIES=""  # Empty = all repos
CHECK_TYPE=instant ./gh-notify
```

### GitHub Actions Testing
1. Go to Actions tab
2. Select "Scheduled GitHub Notifier"
3. Click "Run workflow"
4. Choose "instant" to test immediately

## 📈 Performance Considerations

### API Rate Limits
- GitHub API has rate limits (5000 requests/hour for authenticated requests)
- The app automatically:
  - Skips archived and forked repos
  - Limits checks to reduce API calls
  - Handles errors gracefully

### Recommendations
- **For many repos** (>20): Use longer `COMMIT_LOOKBACK_MINUTES` (120-240)
- **For few repos** (<5): Can use shorter periods (30-60 minutes)
- **High activity repos**: Consider specific tracking instead of all repos

## 🔍 Troubleshooting

### Not receiving commit notifications?

1. **Check if tracking is enabled**:
   ```bash
   echo $TRACK_COMMITS_REALTIME  # Should be "true"
   ```

2. **Verify repository access**:
   - Ensure your GitHub token has `repo` scope
   - Check if repositories are public or you have access

3. **Check lookback period**:
   - If commits are older than `COMMIT_LOOKBACK_MINUTES`, they won't be detected
   - Increase the value if needed

4. **Review debug logs**:
   ```bash
   CHECK_TYPE=instant ./gh-notify 2>&1 | grep -i commit
   ```

### Too many notifications?

1. **Increase cooldown** by keeping default 24-hour period
2. **Track specific repos** instead of all
3. **Adjust lookback period** to be more selective

### Cache issues?

The cache prevents duplicate notifications:
```bash
# View cache content
cat cache.json | jq .

# Clear cache to reset (will re-notify everything)
rm cache.json
```

## 📚 Advanced Usage

### Combining with Other Notifications

Commit tracking works alongside other notifications:
- PR reviews
- Issue assignments  
- Workflow failures
- Repository invitations

All appear in the same Discord message when detected.

### Custom Schedule for Commit Checks

Edit `.github/workflows/scheduled-notify.yml`:
```yaml
- cron: '*/30 * * * *'  # Every 30 minutes for commits
```

Then adjust:
```bash
COMMIT_LOOKBACK_MINUTES=35  # Slightly more than schedule interval
```

## 🎉 Ready to Go!

Your GitHub Notifier can now track commits across multiple repositories. Adjust the configuration based on your needs and enjoy comprehensive GitHub activity monitoring!