# GitHub Notifier

A lightweight GitHub Actions workflow that automatically sends Discord notifications for GitHub activity, featuring smart duplicate filtering and scheduled digest reports.

[![GitHub Actions](https://img.shields.io/github/actions/workflow/status/wilfierd/gh-noftify/notify.yml?branch=main)](https://github.com/wilfierd/gh-noftify/actions)
[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ Features

- **Scheduled Digests**: Automatic morning (7:00 AM) and evening (9:00 PM) reports for GMT+7
- **Real-time Alerts**: Instant notifications every 2 hours for new GitHub activity
- **Smart Filtering**: Prevents duplicate notifications with 24-hour cooldown
- **Discord Integration**: Clean, formatted messages sent directly to your Discord channel
- **Manual Control**: Run notifications on-demand with customizable check types
- **Efficient Caching**: Minimal repository commits, only when necessary
- **Multi-Repository Commit Tracking**: Monitor commits across all your repos or selected ones

##  Quick Setup

### 1. Repository Secrets

Add these secrets to your repository (`Settings` → `Secrets and variables` → `Actions`):

```
GH_TOKEN=ghp_your_github_personal_access_token
DISCORD_WEBHOOK=https://discord.com/api/webhooks/your_webhook_url
```

### 2. GitHub Token Permissions

Your `GH_TOKEN` needs these scopes:
- `repo` - Access repositories
- `notifications` - Read notifications
- `user:email` - Read user email

### 3. Discord Webhook

1. Go to your Discord server → `Server Settings` → `Integrations` → `Webhooks`
2. Create a new webhook and copy the URL
3. Add it as `DISCORD_WEBHOOK` secret

## Schedule


| Time (Vietnam, GMT+7) | Time (UTC) | Type | Description |
|-----------------------|------------|------|-------------|
| ~6:15–7:00 AM         | 23:15 UTC  | Morning Digest | What needs your attention today |
| ~9:00–9:30 PM         | 14:00 UTC  | Evening Digest | Your accomplishments summary |
| Every 2 hours         | Every 2 hours | Instant Check | New alerts only |

**Note:**
- The workflow is scheduled at `23:15 UTC` (6:15 AM Vietnam) for the morning digest and `14:00 UTC` (9:00 PM Vietnam) for the evening digest.
- This is done to compensate for GitHub Actions' delay (5–30 minutes), so notifications arrive before 7:00 AM and 9:30 PM local time.
- Actual notification time may vary slightly due to GitHub's scheduling lag.

## Manual Usage

Run the workflow manually with different options:

### Check Types
Go to the `Actions` tab → `GitHub Notifier` → `Run workflow` and choose a check type:

- **`instant`** – Check for new alerts only
- **`morning`** – Generate morning briefing (what needs attention)
- **`evening`** – Generate evening summary (accomplishments)
- **`commit`** – Send commit notification (on push)
- **`all`** – Run all notification types (instant, morning, evening, commit)

## Configuration

The workflow uses these environment variables:

```yaml
CHECK_TYPE: 'instant' | 'morning' | 'evening' | 'commit' | 'all'  # Notification type
CHECK_INTERVAL: '5m'                  # Time between instant checks
DAILY_REPORT_TIME: '07:00'            # Morning report time (Vietnam timezone) 
TIMEZONE: 'Asia/Ho_Chi_Minh'          # Timezone for reports
TRACK_COMMITS_REALTIME: 'false'       # Enable real-time commit notifications
COMMIT_LOOKBACK_MINUTES: '120'        # How far back to check for commits
TRACKED_REPOSITORIES: ''              # Comma-separated repos or empty for all
```

## Development

### Prerequisites

- Go 1.22+
- Git

## Contributing

🤝 Contributions are welcome! Fork the repository and submit a pull request.


