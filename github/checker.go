package github

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type CheckResult struct {
	PRsNeedingReview      []PullRequest
	StaleOwnPRs           []PullRequest
	AssignedIssues        []Issue
	UnreadNotifications   []Notification
	FailedWorkflows       []WorkflowRun
	RepositoryInvitations []Invitation
	RecentCommits         []Commit // New field for real-time commit tracking
}

type DailyDigest struct {
	PRsOpened             []PullRequest
	PRsMerged             []PullRequest
	PRsReviewed           []PullRequest
	IssuesOpened          []Issue
	IssuesClosed          []Issue
	CommitsToday          []Commit // Changed from int to []Commit to include actual commits
	FailedWorkflows       []WorkflowRun
	PendingReviews        []PullRequest
	AssignedIssues        []Issue
	RepositoryInvitations []Invitation // Add invitations to daily digest
	Date                  time.Time
	IsEvening             bool // true for evening digest, false for morning
}

func (c *Client) CheckForAlerts(username string) (*CheckResult, error) {
	return c.CheckForAlertsWithCommits(username, false, nil, 0)
}

// CheckForAlertsWithCommits includes optional commit tracking based on configuration
func (c *Client) CheckForAlertsWithCommits(username string, trackCommits bool, trackedRepos []string, lookbackMinutes int) (*CheckResult, error) {
	result := &CheckResult{}

	// Create context with timeout for all API calls
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Note: Context is available for future use if we need to pass it to API calls
	_ = ctx

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	var mu sync.Mutex // Protect shared result struct

	// Channel to collect errors from goroutines
	errChan := make(chan error, 6) // Buffer for 6 potential errors

	fmt.Println("DEBUG: Starting parallel API calls...")
	startTime := time.Now()

	// 1. Get PRs that need review
	wg.Add(1)
	go func() {
		defer wg.Done()
		reviewRequests, err := c.GetReviewRequests(username)
		if err != nil {
			errChan <- fmt.Errorf("failed to get review requests: %w", err)
			return
		}
		mu.Lock()
		result.PRsNeedingReview = reviewRequests
		mu.Unlock()
		fmt.Println("DEBUG: Completed review requests")
	}()

	// 2. Get user's own PRs that might be stale
	wg.Add(1)
	go func() {
		defer wg.Done()
		ownPRs, err := c.GetUserPullRequests(username)
		if err != nil {
			errChan <- fmt.Errorf("failed to get user PRs: %w", err)
			return
		}

		// Filter for stale PRs (older than 2 days, no recent activity)
		staleDuration := 48 * time.Hour
		var stalePRs []PullRequest
		for _, pr := range ownPRs {
			if time.Since(pr.UpdatedAt) > staleDuration && !pr.Draft {
				stalePRs = append(stalePRs, pr)
			}
		}

		mu.Lock()
		result.StaleOwnPRs = stalePRs
		mu.Unlock()
		fmt.Println("DEBUG: Completed user PRs and stale filtering")
	}()

	// 3. Get assigned issues
	wg.Add(1)
	go func() {
		defer wg.Done()
		assignedIssues, err := c.GetAssignedIssues(username)
		if err != nil {
			errChan <- fmt.Errorf("failed to get assigned issues: %w", err)
			return
		}
		mu.Lock()
		result.AssignedIssues = assignedIssues
		mu.Unlock()
		fmt.Println("DEBUG: Completed assigned issues")
	}()

	// 4. Get unread notifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		notifications, err := c.GetNotifications()
		if err != nil {
			// Don't fail the whole check if notifications fail due to permissions
			fmt.Printf("Warning: failed to get notifications: %v\n", err)
			notifications = []Notification{}
		}

		var unreadNotifications []Notification
		for _, notif := range notifications {
			if notif.Unread {
				unreadNotifications = append(unreadNotifications, notif)
			}
		}

		mu.Lock()
		result.UnreadNotifications = unreadNotifications
		mu.Unlock()
		fmt.Println("DEBUG: Completed notifications")
	}()

	// 5. Get repository invitations
	wg.Add(1)
	go func() {
		defer wg.Done()
		invitations, err := c.GetRepositoryInvitations()
		if err != nil {
			// Don't fail the whole check if invitations fail
			fmt.Printf("Warning: failed to get repository invitations: %v\n", err)
			invitations = []Invitation{}
		}
		mu.Lock()
		result.RepositoryInvitations = invitations
		mu.Unlock()
		fmt.Println("DEBUG: Completed repository invitations")
	}()

	// 6. Get recent workflow failures
	wg.Add(1)
	go func() {
		defer wg.Done()
		failedWorkflows, err := c.GetRecentWorkflowRuns(username)
		if err != nil {
			// Don't fail the whole check if workflows fail
			fmt.Printf("Warning: failed to get workflow runs: %v\n", err)
			failedWorkflows = []WorkflowRun{}
		}
		mu.Lock()
		result.FailedWorkflows = failedWorkflows
		mu.Unlock()
		fmt.Println("DEBUG: Completed workflow runs")
	}()

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for errors from goroutines
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// If we have critical errors (not warnings), return them
	if len(errors) > 0 {
		// Check if any errors are critical (not just warnings)
		criticalErrors := []error{}
		for _, err := range errors {
			// Only treat as critical if it's not a warning message
			if err != nil {
				criticalErrors = append(criticalErrors, err)
			}
		}
		if len(criticalErrors) > 0 {
			return nil, fmt.Errorf("critical errors occurred: %v", criticalErrors)
		}
	}

	// Get recent commits if tracking is enabled (this runs after parallel calls)
	if trackCommits && lookbackMinutes > 0 {
		since := time.Now().Add(-time.Duration(lookbackMinutes) * time.Minute)
		fmt.Printf("DEBUG: Checking for commits since %s (last %d minutes)\n", since.Format(time.RFC3339), lookbackMinutes)

		recentCommits, err := c.GetRecentCommitsFromSelectedRepos(username, since, trackedRepos)
		if err != nil {
			// Don't fail the whole check if commit fetching fails
			fmt.Printf("Warning: failed to get recent commits: %v\n", err)
			recentCommits = []Commit{}
		}

		fmt.Printf("DEBUG: Found %d recent commits\n", len(recentCommits))
		result.RecentCommits = recentCommits
	}

	elapsed := time.Since(startTime)
	fmt.Printf("DEBUG: Parallel API calls completed in %v\n", elapsed)

	return result, nil
}

func (c *Client) GenerateDailyDigest(username string, trackAllCommits bool) (*DailyDigest, error) {
	now := time.Now()

	// Determine if this is evening digest (after 12 PM UTC = 7 PM Vietnam)
	// For manual testing, force morning mode when CHECK_TYPE=morning
	isEvening := now.Hour() >= 12
	if os.Getenv("CHECK_TYPE") == "morning" {
		isEvening = false
		fmt.Printf("DEBUG: Forced to morning mode for testing (CHECK_TYPE=morning)\n")
	}

	digest := &DailyDigest{
		Date:      now,
		IsEvening: isEvening,
	}

	if isEvening {
		// Evening digest: Show what was accomplished today
		// Get start of today (midnight) instead of last 24 hours
		startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		// Get PRs opened today
		ownPRs, err := c.GetUserPullRequests(username)
		if err != nil {
			return nil, fmt.Errorf("failed to get user PRs: %w", err)
		}

		for _, pr := range ownPRs {
			if pr.CreatedAt.After(startOfToday) {
				digest.PRsOpened = append(digest.PRsOpened, pr)
			}
			// Check if PR was merged today
			if pr.State == "closed" && pr.UpdatedAt.After(startOfToday) {
				digest.PRsMerged = append(digest.PRsMerged, pr)
			}
		}

		// Get issues worked on today
		issues, err := c.GetUserIssues(username)
		if err != nil {
			return nil, fmt.Errorf("failed to get user issues: %w", err)
		}

		for _, issue := range issues {
			if issue.CreatedAt.After(startOfToday) {
				digest.IssuesOpened = append(digest.IssuesOpened, issue)
			}
			if issue.State == "closed" && issue.UpdatedAt.After(startOfToday) {
				digest.IssuesClosed = append(digest.IssuesClosed, issue)
			}
		}

		// Get actual commits from all repositories for today (if enabled)
		if trackAllCommits {
			commits, err := c.GetRecentCommitsFromAllRepos(username, startOfToday)
			if err != nil {
				fmt.Printf("Warning: failed to get commits from all repos: %v\n", err)
				digest.CommitsToday = []Commit{} // Empty slice on error
			} else {
				digest.CommitsToday = commits
			}
		} else {
			digest.CommitsToday = []Commit{} // Empty if feature disabled
		}

	} else {
		// Morning digest: Show what needs attention today
		// Get pending review requests
		reviewRequests, err := c.GetReviewRequests(username)
		if err != nil {
			return nil, fmt.Errorf("failed to get review requests: %w", err)
		}
		digest.PendingReviews = reviewRequests

		// Get assigned issues
		assignedIssues, err := c.GetAssignedIssues(username)
		if err != nil {
			return nil, fmt.Errorf("failed to get assigned issues: %w", err)
		}
		digest.AssignedIssues = assignedIssues

		// Get repository invitations for morning digest
		invitations, err := c.GetRepositoryInvitations()
		if err != nil {
			// Don't fail the whole digest if invitations fail
			fmt.Printf("Warning: failed to get repository invitations for daily digest: %v\n", err)
			invitations = []Invitation{}
		}
		digest.RepositoryInvitations = invitations

		// Get recent commits for context (previous day for morning digest, if enabled)
		if trackAllCommits {
			oneDaysAgo := now.AddDate(0, 0, -1)
			commits, err := c.GetRecentCommitsFromAllRepos(username, oneDaysAgo)
			if err != nil {
				fmt.Printf("Warning: failed to get commits from all repos for morning digest: %v\n", err)
				digest.CommitsToday = []Commit{} // Empty slice on error
			} else {
				digest.CommitsToday = commits
			}
		} else {
			digest.CommitsToday = []Commit{} // Empty if feature disabled
		}
	}

	return digest, nil
}

func (r *CheckResult) HasAlerts() bool {
	return len(r.PRsNeedingReview) > 0 ||
		len(r.StaleOwnPRs) > 0 ||
		len(r.AssignedIssues) > 0 ||
		len(r.UnreadNotifications) > 0 ||
		len(r.FailedWorkflows) > 0 ||
		len(r.RepositoryInvitations) > 0 ||
		len(r.RecentCommits) > 0
}

func (r *CheckResult) GetAlertCount() int {
	return len(r.PRsNeedingReview) +
		len(r.StaleOwnPRs) +
		len(r.AssignedIssues) +
		len(r.UnreadNotifications) +
		len(r.FailedWorkflows) +
		len(r.RepositoryInvitations) +
		len(r.RecentCommits)
}
