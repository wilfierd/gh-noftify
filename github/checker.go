package github

import (
	"fmt"
	"time"
)

type CheckResult struct {
	PRsNeedingReview      []PullRequest
	StaleOwnPRs           []PullRequest
	AssignedIssues        []Issue
	UnreadNotifications   []Notification
	FailedWorkflows       []WorkflowRun
	RepositoryInvitations []Invitation
}

type DailyDigest struct {
	PRsOpened             []PullRequest
	PRsMerged             []PullRequest
	PRsReviewed           []PullRequest
	IssuesOpened          []Issue
	IssuesClosed          []Issue
	CommitsToday          int
	FailedWorkflows       []WorkflowRun
	PendingReviews        []PullRequest
	AssignedIssues        []Issue
	RepositoryInvitations []Invitation // Add invitations to daily digest
	Date                  time.Time
	IsEvening             bool // true for evening digest, false for morning
}

func (c *Client) CheckForAlerts(username string) (*CheckResult, error) {
	result := &CheckResult{}

	// Get PRs that need review
	reviewRequests, err := c.GetReviewRequests(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get review requests: %w", err)
	}
	result.PRsNeedingReview = reviewRequests

	// Get user's own PRs that might be stale
	ownPRs, err := c.GetUserPullRequests(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user PRs: %w", err)
	}

	// Filter for stale PRs (older than 2 days, no recent activity)
	staleDuration := 48 * time.Hour
	for _, pr := range ownPRs {
		if time.Since(pr.UpdatedAt) > staleDuration && !pr.Draft {
			result.StaleOwnPRs = append(result.StaleOwnPRs, pr)
		}
	}

	// Get assigned issues
	assignedIssues, err := c.GetAssignedIssues(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigned issues: %w", err)
	}
	result.AssignedIssues = assignedIssues

	// Get unread notifications
	notifications, err := c.GetNotifications()
	if err != nil {
		// Don't fail the whole check if notifications fail due to permissions
		fmt.Printf("Warning: failed to get notifications: %v\n", err)
		// Continue with empty notifications
		notifications = []Notification{}
	}

	for _, notif := range notifications {
		if notif.Unread {
			result.UnreadNotifications = append(result.UnreadNotifications, notif)
		}
	}

	// Get repository invitations
	invitations, err := c.GetRepositoryInvitations()
	if err != nil {
		// Don't fail the whole check if invitations fail
		fmt.Printf("Warning: failed to get repository invitations: %v\n", err)
		invitations = []Invitation{}
	}
	result.RepositoryInvitations = invitations

	// Get recent workflow failures
	failedWorkflows, err := c.GetRecentWorkflowRuns(username)
	if err != nil {
		// Don't fail the whole check if workflows fail
		fmt.Printf("Warning: failed to get workflow runs: %v\n", err)
	}
	result.FailedWorkflows = failedWorkflows

	return result, nil
}

func (c *Client) GenerateDailyDigest(username string) (*DailyDigest, error) {
	now := time.Now()

	// Determine if this is evening digest (after 12 PM UTC = 7 PM Vietnam)
	isEvening := now.Hour() >= 12

	digest := &DailyDigest{
		Date:      now,
		IsEvening: isEvening,
	}

	// Get time range for "today" (last 24 hours)
	yesterday := now.AddDate(0, 0, -1)

	if isEvening {
		// Evening digest: Show what was accomplished today
		// Get PRs opened today
		ownPRs, err := c.GetUserPullRequests(username)
		if err != nil {
			return nil, fmt.Errorf("failed to get user PRs: %w", err)
		}

		for _, pr := range ownPRs {
			if pr.CreatedAt.After(yesterday) {
				digest.PRsOpened = append(digest.PRsOpened, pr)
			}
			// Check if PR was merged today
			if pr.State == "closed" && pr.UpdatedAt.After(yesterday) {
				digest.PRsMerged = append(digest.PRsMerged, pr)
			}
		}

		// Get issues worked on today
		issues, err := c.GetUserIssues(username)
		if err != nil {
			return nil, fmt.Errorf("failed to get user issues: %w", err)
		}

		for _, issue := range issues {
			if issue.CreatedAt.After(yesterday) {
				digest.IssuesOpened = append(digest.IssuesOpened, issue)
			}
			if issue.State == "closed" && issue.UpdatedAt.After(yesterday) {
				digest.IssuesClosed = append(digest.IssuesClosed, issue)
			}
		}

		// Get commit count for today (simplified)
		// This would require additional API calls to get accurate commit count
		digest.CommitsToday = len(digest.PRsOpened) // Simplified estimation

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
	}

	return digest, nil
}

func (r *CheckResult) HasAlerts() bool {
	return len(r.PRsNeedingReview) > 0 ||
		len(r.StaleOwnPRs) > 0 ||
		len(r.AssignedIssues) > 0 ||
		len(r.UnreadNotifications) > 0 ||
		len(r.FailedWorkflows) > 0 ||
		len(r.RepositoryInvitations) > 0
}

func (r *CheckResult) GetAlertCount() int {
	return len(r.PRsNeedingReview) +
		len(r.StaleOwnPRs) +
		len(r.AssignedIssues) +
		len(r.UnreadNotifications) +
		len(r.FailedWorkflows) +
		len(r.RepositoryInvitations)
}
