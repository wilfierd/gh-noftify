package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wilfierd/gh-notify/github"
	"github.com/wilfierd/gh-notify/notify"
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
	githubToken := os.Getenv("GITHUB_TOKEN")

	if discordWebhook == "" {
		log.Fatal("DISCORD_WEBHOOK environment variable is required")
	}

	// Create Discord notifier
	discordNotifier := notify.NewDiscordNotifier(discordWebhook)

	// Get avatar URL if GitHub token is available
	var avatarURL string
	if githubToken != "" && author != "" {
		client := github.NewClient(githubToken)
		if user, err := client.GetUserByUsername(author); err == nil {
			avatarURL = user.AvatarURL
		}
	}

	// Format commit notification
	discordMessage, err := notify.FormatCommitNotification(sha, commitMessage, author, repoName, commitURL, repoURL, avatarURL)
	if err != nil {
		log.Fatalf("Failed to format commit notification: %v", err)
	}

	// Send notification
	if err := discordNotifier.SendMessage(discordMessage); err != nil {
		log.Fatalf("Failed to send commit notification: %v", err)
	}

	fmt.Println("Commit notification sent successfully!")
}
