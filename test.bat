@echo off
REM Test script for GitHub Notifier (Windows)
REM Chạy script này để test tool trước khi deploy

echo 🧪 Testing GitHub Notifier...

REM Check if test-env.bat exists and run it
if exist test-env.bat (
    echo 🔧 Loading environment from test-env.bat...
    call test-env.bat
) else (
    echo ⚠️  test-env.bat not found!
    echo 📋 Please copy test-env.bat.template to test-env.bat and fill in your values
    echo 💡 Run: copy test-env.bat.template test-env.bat
    pause
    exit /b 1
)

REM Kiểm tra Go version
echo 📋 Checking Go version...
go version

REM Kiểm tra environment variables
echo 🔐 Checking environment variables...
if "%GITHUB_TOKEN%"=="" (
    echo ❌ GITHUB_TOKEN is not set
    echo Please run: test-env.bat first or set manually
    exit /b 1
)

if "%DISCORD_WEBHOOK%"=="" (
    echo ❌ DISCORD_WEBHOOK is not set
    echo Please run: test-env.bat first or set manually
    exit /b 1
)

echo ✅ Environment variables are set

REM Download dependencies
echo 📦 Downloading dependencies...
go mod download

REM Build application
echo 🔨 Building application...
go build -o gh-notify.exe main.go

if %errorlevel% neq 0 (
    echo ❌ Build failed
    exit /b 1
)

echo ✅ Build successful

REM Run test
echo 🚀 Running test...
if "%GITHUB_USERNAME%"=="" (
    for /f "tokens=*" %%i in ('git config user.name') do set GITHUB_USERNAME=%%i
)
gh-notify.exe

if %errorlevel% equ 0 (
    echo ✅ Test completed successfully!
    echo 🎉 Your GitHub Notifier is ready to deploy!
) else (
    echo ❌ Test failed
    exit /b 1
)

REM Clean up
del gh-notify.exe
echo 🧹 Cleaned up test files
