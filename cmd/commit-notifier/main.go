package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/wilfierd/gh-notify/github"
	"github.com/wilfierd/gh-notify/notify"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()
	
	// Get environment variables from GitHub Actions
	sha := os.Getenv("GITHUB_SHA")
	commitMessage := os.Getenv("COMMIT_MESSAGE")
	author := os.Getenv("COMMIT_AUTHOR")
	repoName := os.Getenv("GITHUB_REPOSITORY")
	commitURL := os.Getenv("COMMIT_URL")
	repoURL := os.Getenv("REPO_URL")
	discordWebhook := os.Getenv("DISCORD_WEBHOOK")
	githubToken := os.Getenv("GH_TOKEN")

	if discordWebhook == "" {
		log.Fatal("DISCORD_WEBHOOK environment variable is required")
	}

	// Create Discord notifier
	discordNotifier := notify.NewDiscordNotifier(discordWebhook)

	// Get avatar URL if GitHub token is available
	var avatarURL string
	fmt.Printf("DEBUG: githubToken exists: %v, author: '%s'\n", githubToken != "", author)

	if githubToken != "" && author != "" {
		client := github.NewClient(githubToken)
		fmt.Printf("DEBUG: Fetching user info for: %s\n", author)
		if user, err := client.GetUserByUsername(author); err == nil {
			avatarURL = user.AvatarURL
			fmt.Printf("DEBUG: Successfully got avatar URL: %s\n", avatarURL)
		} else {
			fmt.Printf("DEBUG: Failed to get user info: %v\n", err)
		}
	} else {
		fmt.Printf("DEBUG: Skipping avatar fetch - missing token or author\n")
	}

	fmt.Printf("DEBUG: Final avatarURL: '%s'\n", avatarURL)

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
