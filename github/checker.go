package github

import (
	"fmt"
	"time"
)

type CheckResult struct {
	PRsNeedingReview    []PullRequest
	StaleOwnPRs         []PullRequest
	AssignedIssues      []Issue
	UnreadNotifications []Notification
	FailedWorkflows     []WorkflowRun
}

type DailyDigest struct {
	PRsOpened       []PullRequest
	PRsMerged       []PullRequest
	ReviewsGiven    []Review
	IssuesResolved  []Issue
	FailedWorkflows []WorkflowRun
	Date            time.Time
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
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	for _, notif := range notifications {
		if notif.Unread {
			result.UnreadNotifications = append(result.UnreadNotifications, notif)
		}
	}

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
	digest := &DailyDigest{
		Date: time.Now(),
	}

	// Get PRs opened in the last 24 hours
	// This is a simplified version - would need proper date filtering
	ownPRs, err := c.GetUserPullRequests(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user PRs: %w", err)
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	for _, pr := range ownPRs {
		if pr.CreatedAt.After(yesterday) {
			digest.PRsOpened = append(digest.PRsOpened, pr)
		}
	}

	// TODO: Implement other digest features
	// - PRs merged
	// - Reviews given
	// - Issues resolved
	// - Failed workflows

	return digest, nil
}

func (r *CheckResult) HasAlerts() bool {
	return len(r.PRsNeedingReview) > 0 ||
		len(r.StaleOwnPRs) > 0 ||
		len(r.AssignedIssues) > 0 ||
		len(r.UnreadNotifications) > 0 ||
		len(r.FailedWorkflows) > 0
}

func (r *CheckResult) GetAlertCount() int {
	return len(r.PRsNeedingReview) +
		len(r.StaleOwnPRs) +
		len(r.AssignedIssues) +
		len(r.UnreadNotifications) +
		len(r.FailedWorkflows)
}
