# GitHub Notifier

A lightweight GitHub Actions workflow that automatically sends Discord notifications for GitHub activity, featuring smart duplicate filtering and scheduled digest reports.

[![GitHub Actions](https://img.shields.io/github/actions/workflow/status/wilfierd/gh-noftify/notify.yml?branch=main)](https://github.com/wilfierd/gh-noftify/actions)
[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ Features

- **Scheduled Digests**: Automatic morning (8:00 AM) and evening (9:00 PM) reports
- **Real-time Alerts**: Instant notifications every 1 hours for new GitHub activity
- **Smart Filtering**: Prevents duplicate notifications with 24-hour cooldown
- **Discord Integration**: Clean, formatted messages sent directly to your Discord channel
- **Manual Control**: Run notifications on-demand with customizable check types
- **Efficient Caching**: Minimal repository commits, only when necessary

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

| Time (Vietnam) | Time (UTC) | Type | Description |
|---------------|------------|------|-------------|
| 8:00 AM | 0:00 UTC | Morning Digest | What needs your attention today |
| 9:00 PM | 14:00 UTC | Evening Digest | Your accomplishments summary |
| Every 60 min | Every 60 min | Instant Check | New alerts only |

## Manual Usage

Run the workflow manually with different options:

### Check Types
1. Go to `Actions` tab → `GitHub Notifier` → `Run workflow`
2. Choose check type:
   - **`instant`** - Check for new alerts only
   - **`morning`** - Generate morning briefing (what needs attention)
   - **`evening`** - Generate evening summary (accomplishments)  
   - **`both`** - Run both instant and daily checks

## Configuration

The workflow uses these environment variables:

```yaml
CHECK_TYPE: 'both'                    # instant, morning, evening, or both
CHECK_INTERVAL: '5m'                  # Time between checks
DAILY_REPORT_TIME: '02:00'           # Daily report time (Vietnam timezone)
CACHE_FILE: 'cache.json'             # Cache file location
TIMEZONE: 'Asia/Ho_Chi_Minh'         # Timezone for reports
```

## Development

### Prerequisites

- Go 1.22+
- Git

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/wilfierd/gh-noftify.git
   cd gh-noftify
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Create environment file**
   ```bash
   cp .env.example .env
   # Edit .env with your actual values
   ```

4. **Build and run**
   ```bash
   go build -o gh-notify main.go
   ./gh-notify
   ```
## Contributing

🤝 Contributions are welcome! Fork the repository and submit a pull request.


