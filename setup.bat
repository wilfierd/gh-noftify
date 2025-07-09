@echo off
REM Setup script for GitHub Notifier
REM Run this script to set up your development environment

echo 🚀 Setting up GitHub Notifier...
echo.

REM Check if template exists
if not exist test-env.bat.template (
    echo ❌ test-env.bat.template not found!
    echo Please make sure you're in the right directory
    pause
    exit /b 1
)

REM Copy template if test-env.bat doesn't exist
if not exist test-env.bat (
    echo 📋 Creating test-env.bat from template...
    copy test-env.bat.template test-env.bat
    echo ✅ Created test-env.bat
    echo.
    echo 🔧 Please edit test-env.bat and fill in your actual values:
    echo   - GITHUB_TOKEN: Get from https://github.com/settings/tokens
    echo   - DISCORD_WEBHOOK: Get from Discord Server Settings
    echo   - GITHUB_USERNAME: Your GitHub username
    echo.
    echo 💡 After editing, run: test.bat
    pause
) else (
    echo ✅ test-env.bat already exists
    echo.
    echo 🧪 Ready to test! Run: test.bat
    pause
)
