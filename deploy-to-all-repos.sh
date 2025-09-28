#!/bin/bash

# Deploy Discord Commit Notifier to All Repositories
# This script adds the notification workflow to all your GitHub repositories

set -e

# Configuration
GITHUB_USERNAME="${1:-$(gh api user -q .login)}"
ACTION_VERSION="v1"  # Change this when you publish updates

echo "🚀 Discord Commit Notifier Deployment Script"
echo "============================================"
echo "This will add commit notifications to all repositories for: $GITHUB_USERNAME"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "❌ Error: GitHub CLI (gh) is not installed"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "❌ Error: Not authenticated with GitHub"
    echo "Run: gh auth login"
    exit 1
fi

# Confirm before proceeding
read -p "Do you want to continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Get all repositories
echo ""
echo "📋 Fetching your repositories..."
REPOS=$(gh repo list "$GITHUB_USERNAME" --limit 1000 --json name,isArchived,isFork -q '.[] | select(.isArchived == false and .isFork == false) | .name')

# Count repositories
REPO_COUNT=$(echo "$REPOS" | wc -l)
echo "Found $REPO_COUNT active repositories (excluding archived and forked)"
echo ""

# Create a temporary directory for work
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Counter for progress
CURRENT=0
SUCCESS=0
SKIPPED=0
FAILED=0

# Process each repository
for repo in $REPOS; do
    CURRENT=$((CURRENT + 1))
    echo "[$CURRENT/$REPO_COUNT] Processing $repo..."
    
    # Clone the repository
    if ! git clone --quiet --depth 1 "https://github.com/$GITHUB_USERNAME/$repo.git" "$TEMP_DIR/$repo" 2>/dev/null; then
        echo "  ❌ Failed to clone $repo"
        FAILED=$((FAILED + 1))
        continue
    fi
    
    cd "$TEMP_DIR/$repo"
    
    # Check if workflow already exists
    if [ -f ".github/workflows/discord-notify.yml" ]; then
        echo "  ⏭️  Workflow already exists, skipping"
        SKIPPED=$((SKIPPED + 1))
        cd ..
        rm -rf "$repo"
        continue
    fi
    
    # Create .github/workflows directory
    mkdir -p .github/workflows
    
    # Create the workflow file
    cat > .github/workflows/discord-notify.yml << EOF
name: Discord Commit Notifications
on:
  push:
    branches: [ main, master, develop ]
    # Ignore changes to docs and configs
    paths-ignore:
      - '**.md'
      - 'LICENSE'
      - '.gitignore'

jobs:
  notify:
    runs-on: ubuntu-latest
    # Skip for bot commits
    if: github.actor != 'dependabot[bot]' && github.actor != 'renovate[bot]'
    steps:
      - name: Send Discord Notification
        uses: wilfierd/gh-noftify@$ACTION_VERSION
        with:
          discord-webhook: \${{ secrets.DISCORD_WEBHOOK }}
          github-token: \${{ secrets.GITHUB_TOKEN }}
EOF
    
    # Add and commit the workflow
    git add .github/workflows/discord-notify.yml
    git commit -m "Add Discord commit notifications via gh-noftify action" --quiet
    
    # Push to repository
    if git push --quiet origin HEAD 2>/dev/null; then
        echo "  ✅ Successfully added workflow"
        SUCCESS=$((SUCCESS + 1))
    else
        echo "  ❌ Failed to push changes"
        FAILED=$((FAILED + 1))
    fi
    
    # Clean up
    cd ..
    rm -rf "$repo"
done

# Summary
echo ""
echo "============================================"
echo "📊 Deployment Summary"
echo "============================================"
echo "✅ Successful: $SUCCESS"
echo "⏭️  Skipped:    $SKIPPED"
echo "❌ Failed:     $FAILED"
echo "📦 Total:      $REPO_COUNT"
echo ""

# Instructions for setting up Discord webhook
if [ $SUCCESS -gt 0 ]; then
    echo "🎯 Next Steps:"
    echo "============================================"
    echo "1. Go to each repository's Settings → Secrets and variables → Actions"
    echo "2. Add a new secret named: DISCORD_WEBHOOK"
    echo "3. Set the value to your Discord webhook URL"
    echo ""
    echo "To add the webhook secret to all repos at once, you can use:"
    echo "gh secret set DISCORD_WEBHOOK -b 'YOUR_WEBHOOK_URL' -R $GITHUB_USERNAME/REPO_NAME"
    echo ""
    echo "Or use this one-liner for all repos:"
    echo "for repo in $REPOS; do gh secret set DISCORD_WEBHOOK -b 'YOUR_WEBHOOK_URL' -R $GITHUB_USERNAME/\$repo; done"
fi

echo ""
echo "🎉 Done! Your repositories are now ready for Discord notifications!"