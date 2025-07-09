#!/bin/bash
# Setup script for GitHub Notifier (Linux/Mac)

echo "🚀 Setting up GitHub Notifier..."
echo

# Check if template exists
if [ ! -f "test-env.sh.template" ]; then
    echo "❌ test-env.sh.template not found!"
    echo "Please make sure you're in the right directory"
    exit 1
fi

# Copy template if test-env.sh doesn't exist
if [ ! -f "test-env.sh" ]; then
    echo "📋 Creating test-env.sh from template..."
    cp test-env.sh.template test-env.sh
    chmod +x test-env.sh
    echo "✅ Created test-env.sh"
    echo
    echo "🔧 Please edit test-env.sh and fill in your actual values:"
    echo "  - GITHUB_TOKEN: Get from https://github.com/settings/tokens"
    echo "  - DISCORD_WEBHOOK: Get from Discord Server Settings"
    echo "  - GITHUB_USERNAME: Your GitHub username"
    echo
    echo "💡 After editing, run: ./test.sh"
    read -p "Press Enter to continue..."
else
    echo "✅ test-env.sh already exists"
    echo
    echo "🧪 Ready to test! Run: ./test.sh"
    read -p "Press Enter to continue..."
fi
