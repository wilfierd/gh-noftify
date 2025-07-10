package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gh-notify/notify"
)

func main() {
	// Get environment variables from GitHub Actions
	sha := os.Getenv("GITHUB_SHA")
	commitMessage := os.Getenv("COMMIT_MESSAGE")
	author := os.Getenv("COMMIT_AUTHOR")
	repoName := os.Getenv("GITHUB_REPOSITORY")
	commitURL := os.Getenv("COMMIT_URL")
	repoURL := os.Getenv("REPO_URL")
	discordWebhook := os.Getenv("DISCORD_WEBHOOK")

	if discordWebhook == "" {
		log.Fatal("DISCORD_WEBHOOK environment variable is required")
	}

	// Create Discord notifier
	discordNotifier := notify.NewDiscordNotifier(discordWebhook)

	// Format commit notification
	discordMessage, err := notify.FormatCommitNotification(sha, commitMessage, author, repoName, commitURL, repoURL)
	if err != nil {
		log.Fatalf("Failed to format commit notification: %v", err)
	}

	// Send notification
	if err := discordNotifier.SendMessage(discordMessage); err != nil {
		log.Fatalf("Failed to send commit notification: %v", err)
	}

	fmt.Println("Commit notification sent successfully!")
}
