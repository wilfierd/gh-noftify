@echo off
REM GitHub Notifier Test Environment Variables
REM Replace these values with your actual tokens and webhooks

echo Setting up environment variables...

REM GitHub Personal Access Token (replace with your actual token)
set GITHUB_TOKEN=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

REM Discord Webhook URL (replace with your actual webhook)
set DISCORD_WEBHOOK=https://discord.com/api/webhooks/1392337776149790780/VqIyOXdhd5mpXMfjP58qzGAfKujMtc5LR4pHA_xI-xM9nk_okcduIkhfXOpK6gsSWtlG

REM GitHub Username (replace with your actual username)
set GITHUB_USERNAME=wilfierd

REM Optional: Set other environment variables 
set CHECK_INTERVAL=5m
set DAILY_REPORT_TIME=02:00
set CACHE_FILE=cache.json
set TIMEZONE=Asia/Ho_Chi_Minh

echo Environment variables set successfully!
echo.
echo GITHUB_TOKEN: %GITHUB_TOKEN%
echo DISCORD_WEBHOOK: %DISCORD_WEBHOOK%
echo GITHUB_USERNAME: %GITHUB_USERNAME%
echo.
echo You can now run: test.bat
echo.
pause
