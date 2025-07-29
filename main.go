package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/wilfierd/gh-notify/cache"
	"github.com/wilfierd/gh-notify/config"
	"github.com/wilfierd/gh-notify/github"
	"github.com/wilfierd/gh-notify/notify"
)

func main() {
	// Load .env file if exists (silent fail for production)
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("DEBUG: No .env file found: %v\n", err)
	} else {
		fmt.Println("DEBUG: .env file loaded successfully")
	}

	// Debug environment variables
	fmt.Printf("DEBUG: CHECK_TYPE = '%s'\n", os.Getenv("CHECK_TYPE"))
	fmt.Printf("DEBUG: SCHEDULE_TYPE = '%s'\n", os.Getenv("SCHEDULE_TYPE"))
	fmt.Printf("DEBUG: CHECK_INTERVAL = '%s'\n", os.Getenv("CHECK_INTERVAL"))
	fmt.Printf("DEBUG: DAILY_REPORT_TIME = '%s'\n", os.Getenv("DAILY_REPORT_TIME"))
	fmt.Printf("DEBUG: Current time = %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// Load configuration
	cfg := config.Load()

	// Debug config values
	fmt.Printf("DEBUG: cfg.CheckInterval = %v\n", cfg.CheckInterval)
	fmt.Printf("DEBUG: cfg.DailyReportTime = '%s'\n", cfg.DailyReportTime)

	// Validate required configuration
	if cfg.GitHubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}
	if cfg.DiscordWebhook == "" {
		log.Fatal("DISCORD_WEBHOOK environment variable is required")
	}

	// Determine what to run based on time and environment
	now := time.Now()
	checkType := os.Getenv("CHECK_TYPE")
	scheduleType := os.Getenv("SCHEDULE_TYPE")

	// Auto-detect based on schedule trigger
	if checkType == "auto" {
		switch scheduleType {
		case "0 0 * * *": // 7 AM Vietnam time - Morning digest
			checkType = "morning"
		case "0 14 * * *": // 9 PM Vietnam time - Evening digest
			checkType = "evening"
		default: // Every 30 minutes - Instant checks
			checkType = "instant"
		}
	}

	fmt.Printf("DEBUG: Determined check type: %s (from schedule: %s)\n", checkType, scheduleType)

	// Load or create cache state
	state, err := cache.LoadState(cfg.CacheFile)
	if err != nil {
		log.Fatalf("Failed to load cache state: %v", err)
	}

	// Debug cache state
	fmt.Printf("DEBUG: Last check time = %s\n", state.LastCheck.Format("2006-01-02 15:04:05"))
	fmt.Printf("DEBUG: Last daily report = %s\n", state.LastDailyReport.Format("2006-01-02 15:04:05"))
	fmt.Printf("DEBUG: Time since last check = %v\n", time.Since(state.LastCheck))

	// Determine what to run based on check type
	shouldRunMorningDigest := checkType == "morning" || checkType == "both"
	shouldRunEveningDigest := checkType == "evening" || checkType == "both"
	shouldRunDailyReport := shouldRunMorningDigest || shouldRunEveningDigest
	shouldRunInstantCheck := checkType == "instant" || checkType == "both" || (checkType == "auto" && scheduleType == "")

	fmt.Printf("DEBUG: shouldRun conditions:\n")
	fmt.Printf("  - Check type: %s\n", checkType)
	fmt.Printf("  - Should run morning digest: %t\n", shouldRunMorningDigest)
	fmt.Printf("  - Should run evening digest: %t\n", shouldRunEveningDigest)
	fmt.Printf("  - Should run daily report: %t\n", shouldRunDailyReport)
	fmt.Printf("  - Should run instant: %t\n", shouldRunInstantCheck)

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
	fmt.Printf("Daily report: %t, Instant check: %t\n", shouldRunDailyReport, shouldRunInstantCheck)

	hasNewAlerts := false

	// Run instant checks
	if shouldRunInstantCheck {
		if newAlertsFound, err := runInstantChecks(githubClient, discordNotifier, state, username); err != nil {
			log.Printf("Error running instant checks: %v", err)
			// Send error notification
			errorMsg := notify.FormatErrorMessage(err)
			discordNotifier.SendSimpleMessage(errorMsg)
		} else {
			hasNewAlerts = newAlertsFound
		}
		state.LastCheck = now
	}

	// Run daily report (morning or evening)
	if shouldRunDailyReport {
		isEvening := shouldRunEveningDigest
		if err := runDailyReport(githubClient, discordNotifier, state, username, isEvening); err != nil {
			log.Printf("Error running daily report: %v", err)
			// Send error notification
			errorMsg := notify.FormatErrorMessage(err)
			discordNotifier.SendSimpleMessage(errorMsg)
		}
		state.LastDailyReport = now
	}

	// Clean up old entries to keep cache size manageable
	state.CleanupOldEntries(7 * 24 * time.Hour) // Keep 7 days of history

	// Save state only if we had updates (new alerts or daily report)
	shouldSaveCache := shouldRunInstantCheck && hasNewAlerts || shouldRunDailyReport
	if shouldSaveCache {
		if err := state.Save(cfg.CacheFile); err != nil {
			log.Printf("Warning: Failed to save cache state: %v", err)
		}
		fmt.Println("Cache state updated")
	} else {
		fmt.Println("No cache updates needed")
	}

	fmt.Println("GitHub Notifier completed successfully")
}

func runInstantChecks(githubClient *github.Client, discordNotifier *notify.DiscordNotifier, state *cache.State, username string) (bool, error) {
	fmt.Println("Running instant checks...")

	// Get current alerts
	result, err := githubClient.CheckForAlerts(username)
	if err != nil {
		return false, fmt.Errorf("failed to check for alerts: %w", err)
	}

	if !result.HasAlerts() {
		fmt.Println("No alerts found")
		return false, nil
	}

	// Filter for NEW alerts only - don't spam duplicates
	cooldownDuration := 24 * time.Hour // Only notify once per day per alert
	hasNewAlerts := false

	// Check PR reviews - only NEW ones
	var newPRsNeedingReview []interface{}
	for _, pr := range result.PRsNeedingReview {
		key := fmt.Sprintf("review_request_%d", pr.Number)
		if !state.IsNotificationSent(key, cooldownDuration) {
			newPRsNeedingReview = append(newPRsNeedingReview, pr)
			state.MarkNotificationSent(key)
			hasNewAlerts = true
		}
	}

	// Check stale PRs - only NEW ones
	var newStaleOwnPRs []interface{}
	for _, pr := range result.StaleOwnPRs {
		key := fmt.Sprintf("stale_pr_%d", pr.Number)
		if !state.IsNotificationSent(key, cooldownDuration) {
			newStaleOwnPRs = append(newStaleOwnPRs, pr)
			state.MarkNotificationSent(key)
			hasNewAlerts = true
		}
	}

	// Check assigned issues - only NEW ones
	var newAssignedIssues []interface{}
	for _, issue := range result.AssignedIssues {
		key := fmt.Sprintf("assigned_issue_%d", issue.Number)
		if !state.IsNotificationSent(key, cooldownDuration) {
			newAssignedIssues = append(newAssignedIssues, issue)
			state.MarkNotificationSent(key)
			hasNewAlerts = true
		}
	}

	// Check repository invitations - only NEW ones
	var newRepositoryInvitations []interface{}
	for _, invitation := range result.RepositoryInvitations {
		key := fmt.Sprintf("invitation_%d", invitation.ID)
		if !state.IsNotificationSent(key, cooldownDuration) {
			newRepositoryInvitations = append(newRepositoryInvitations, invitation)
			state.MarkNotificationSent(key)
			hasNewAlerts = true
		}
	}

	// Check unread notifications - only NEW ones
	var newUnreadNotifications []interface{}
	for _, notification := range result.UnreadNotifications {
		key := fmt.Sprintf("notification_%s", notification.ID)
		if !state.IsNotificationSent(key, cooldownDuration) {
			newUnreadNotifications = append(newUnreadNotifications, notification)
			state.MarkNotificationSent(key)
			hasNewAlerts = true
		}
	}

	// Check failed workflows - only NEW ones
	var newFailedWorkflows []interface{}
	for _, workflow := range result.FailedWorkflows {
		key := fmt.Sprintf("workflow_%d", workflow.ID)
		if !state.IsNotificationSent(key, cooldownDuration) {
			newFailedWorkflows = append(newFailedWorkflows, workflow)
			state.MarkNotificationSent(key)
			hasNewAlerts = true
		}
	}

	// Only send notification if there are NEW alerts
	if !hasNewAlerts {
		fmt.Println("No new alerts found (all previously notified)")
		return false, nil
	}

	// Create filtered result with only new alerts
	filteredResult := &github.CheckResult{
		PRsNeedingReview:      []github.PullRequest{},
		StaleOwnPRs:           []github.PullRequest{},
		AssignedIssues:        []github.Issue{},
		UnreadNotifications:   []github.Notification{},
		RepositoryInvitations: []github.Invitation{},
		FailedWorkflows:       []github.WorkflowRun{},
	}

	// Convert filtered items back to their original types
	for _, pr := range newPRsNeedingReview {
		filteredResult.PRsNeedingReview = append(filteredResult.PRsNeedingReview, pr.(github.PullRequest))
	}
	for _, pr := range newStaleOwnPRs {
		filteredResult.StaleOwnPRs = append(filteredResult.StaleOwnPRs, pr.(github.PullRequest))
	}
	for _, issue := range newAssignedIssues {
		filteredResult.AssignedIssues = append(filteredResult.AssignedIssues, issue.(github.Issue))
	}
	for _, invitation := range newRepositoryInvitations {
		filteredResult.RepositoryInvitations = append(filteredResult.RepositoryInvitations, invitation.(github.Invitation))
	}
	for _, notification := range newUnreadNotifications {
		filteredResult.UnreadNotifications = append(filteredResult.UnreadNotifications, notification.(github.Notification))
	}
	for _, workflow := range newFailedWorkflows {
		filteredResult.FailedWorkflows = append(filteredResult.FailedWorkflows, workflow.(github.WorkflowRun))
	}

	// Format and send alert message only for NEW items
	message, err := notify.FormatInstantAlert(filteredResult)
	if err != nil {
		return false, fmt.Errorf("failed to format alert: %w", err)
	}

	if message != nil {
		if err := discordNotifier.SendMessage(message); err != nil {
			return false, fmt.Errorf("failed to send Discord message: %w", err)
		}
		totalNewCount := len(newPRsNeedingReview) + len(newStaleOwnPRs) + len(newAssignedIssues) + len(newRepositoryInvitations) + len(newUnreadNotifications) + len(newFailedWorkflows)
		fmt.Printf("Sent instant alert with %d NEW items (filtered duplicates)\n", totalNewCount)
	}

	return true, nil
}

func runDailyReport(githubClient *github.Client, discordNotifier *notify.DiscordNotifier, state *cache.State, username string, isEvening bool) error {
	if isEvening {
		fmt.Println("Running evening digest...")
	} else {
		fmt.Println("Running morning digest...")
	}

	// Generate daily digest with evening flag
	digest, err := githubClient.GenerateDailyDigest(username)
	if err != nil {
		return fmt.Errorf("failed to generate daily digest: %w", err)
	}

	// Set the evening flag manually since we can't pass it through API
	digest.IsEvening = isEvening

	// Format and send daily digest
	message, err := notify.FormatDailyDigest(digest, username)
	if err != nil {
		return fmt.Errorf("failed to format daily digest: %w", err)
	}

	if err := discordNotifier.SendMessage(message); err != nil {
		return fmt.Errorf("failed to send Discord message: %w", err)
	}

	if isEvening {
		fmt.Println("Sent evening digest")
	} else {
		fmt.Println("Sent morning digest")
	}
	return nil
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Check if running in GitHub Actions
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		fmt.Println("Running in GitHub Actions environment")
	}
}
