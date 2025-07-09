#!/bin/bash

# Test script for GitHub Notifier
# Chạy script này để test tool trước khi deploy

echo "🧪 Testing GitHub Notifier..."

# Kiểm tra Go version
echo "📋 Checking Go version..."
go version

# Kiểm tra environment variables
echo "🔐 Checking environment variables..."
if [ -z "$GITHUB_TOKEN" ]; then
    echo "❌ GITHUB_TOKEN is not set"
    echo "Please set: export GITHUB_TOKEN='your_token_here'"
    exit 1
fi

if [ -z "$DISCORD_WEBHOOK" ]; then
    echo "❌ DISCORD_WEBHOOK is not set"
    echo "Please set: export DISCORD_WEBHOOK='your_webhook_url'"
    exit 1
fi

echo "✅ Environment variables are set"

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download

# Build application
echo "🔨 Building application..."
go build -o gh-notify main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Run test
echo "🚀 Running test..."
export GITHUB_USERNAME=${GITHUB_USERNAME:-$(git config user.name)}
./gh-notify

if [ $? -eq 0 ]; then
    echo "✅ Test completed successfully!"
    echo "🎉 Your GitHub Notifier is ready to deploy!"
else
    echo "❌ Test failed"
    exit 1
fi

# Clean up
rm -f gh-notify
echo "🧹 Cleaned up test files"
