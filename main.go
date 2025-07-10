package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gh-notify/cache"
	"github.com/gh-notify/config"
	"github.com/gh-notify/github"
	"github.com/gh-notify/notify"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if exists (silent fail for production)
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Validate required configuration
	if cfg.GitHubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}
	if cfg.DiscordWebhook == "" {
		log.Fatal("DISCORD_WEBHOOK environment variable is required")
	}

	// Load or create cache state
	state, err := cache.LoadState(cfg.CacheFile)
	if err != nil {
		log.Fatalf("Failed to load cache state: %v", err)
	}

	// Initialize clients
	githubClient := github.NewClient(cfg.GitHubToken)
	discordNotifier := notify.NewDiscordNotifier(cfg.DiscordWebhook)

	// Get current user if username not provided
	username := cfg.Username
	if username == "" {
		user, err := githubClient.GetUser()
		if err != nil {
			log.Fatalf("Failed to get current user: %v", err)
		}
		username = user.Login
	}

	fmt.Printf("Running GitHub Notifier for user: %s\n", username)

	// Determine what to run based on time and last execution
	now := time.Now()

	// Check if it's time for daily report (around 9 AM Vietnam time)
	shouldRunDailyReport := shouldRunDaily(now, state.LastDailyReport, cfg.DailyReportTime)

	// Always run instant checks
	shouldRunInstantCheck := time.Since(state.LastCheck) >= cfg.CheckInterval

	fmt.Printf("Running GitHub Notifier for user: %s\n", username)
	fmt.Printf("Daily report: %t, Instant check: %t\n", shouldRunDailyReport, shouldRunInstantCheck)

	// Run instant checks
	if shouldRunInstantCheck {
		if err := runInstantChecks(githubClient, discordNotifier, state, username); err != nil {
			log.Printf("Error running instant checks: %v", err)
			// Send error notification
			errorMsg := notify.FormatErrorMessage(err)
			discordNotifier.SendSimpleMessage(errorMsg)
		}
		state.LastCheck = now
	}

	// Run daily report
	if shouldRunDailyReport {
		if err := runDailyReport(githubClient, discordNotifier, state, username); err != nil {
			log.Printf("Error running daily report: %v", err)
			// Send error notification
			errorMsg := notify.FormatErrorMessage(err)
			discordNotifier.SendSimpleMessage(errorMsg)
		}
		state.LastDailyReport = now
	}

	// Clean up old entries to keep cache size manageable
	state.CleanupOldEntries(7 * 24 * time.Hour) // Keep 7 days of history

	// Save state
	if err := state.Save(cfg.CacheFile); err != nil {
		log.Printf("Warning: Failed to save cache state: %v", err)
	}

	fmt.Println("GitHub Notifier completed successfully")
}

func runInstantChecks(githubClient *github.Client, discordNotifier *notify.DiscordNotifier, state *cache.State, username string) error {
	fmt.Println("Running instant checks...")

	// Get current alerts
	result, err := githubClient.CheckForAlerts(username)
	if err != nil {
		return fmt.Errorf("failed to check for alerts: %w", err)
	}

	if !result.HasAlerts() {
		fmt.Println("No alerts found")
		return nil
	}

	// Check each alert type for duplicates
	cooldownDuration := 2 * time.Hour // Don't spam same alert within 2 hours

	// Check PR reviews
	for _, pr := range result.PRsNeedingReview {
		key := fmt.Sprintf("review_request_%d", pr.Number)
		if !state.IsNotificationSent(key, cooldownDuration) {
			state.MarkNotificationSent(key)
		}
	}

	// Check stale PRs
	for _, pr := range result.StaleOwnPRs {
		key := fmt.Sprintf("stale_pr_%d", pr.Number)
		if !state.IsNotificationSent(key, cooldownDuration) {
			state.MarkNotificationSent(key)
		}
	}

	// Check assigned issues
	for _, issue := range result.AssignedIssues {
		key := fmt.Sprintf("assigned_issue_%d", issue.Number)
		if !state.IsNotificationSent(key, cooldownDuration) {
			state.MarkNotificationSent(key)
		}
	}

	// Format and send alert message
	message, err := notify.FormatInstantAlert(result)
	if err != nil {
		return fmt.Errorf("failed to format alert: %w", err)
	}

	if message != nil {
		if err := discordNotifier.SendMessage(message); err != nil {
			return fmt.Errorf("failed to send Discord message: %w", err)
		}
		fmt.Printf("Sent instant alert with %d items\n", result.GetAlertCount())
	}

	return nil
}

func runDailyReport(githubClient *github.Client, discordNotifier *notify.DiscordNotifier, state *cache.State, username string) error {
	fmt.Println("Running daily report...")

	// Generate daily digest
	digest, err := githubClient.GenerateDailyDigest(username)
	if err != nil {
		return fmt.Errorf("failed to generate daily digest: %w", err)
	}

	// Format and send daily digest
	message, err := notify.FormatDailyDigest(digest, username)
	if err != nil {
		return fmt.Errorf("failed to format daily digest: %w", err)
	}

	if err := discordNotifier.SendMessage(message); err != nil {
		return fmt.Errorf("failed to send Discord message: %w", err)
	}

	fmt.Println("Sent daily digest")
	return nil
}

func shouldRunDaily(now time.Time, lastRun time.Time, dailyTime string) bool {
	// Parse the daily time (e.g., "02:00" for 2 AM UTC = 9 AM Vietnam)
	targetTime, err := time.Parse("15:04", dailyTime)
	if err != nil {
		log.Printf("Invalid daily time format: %s", dailyTime)
		return false
	}

	// Create today's target time
	todayTarget := time.Date(now.Year(), now.Month(), now.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())

	// Check if we haven't run today and it's past the target time
	lastRunDate := lastRun.Format("2006-01-02")
	todayDate := now.Format("2006-01-02")

	return lastRunDate != todayDate && now.After(todayTarget)
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Check if running in GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		fmt.Println("Running in GitHub Actions environment")
	}
}
