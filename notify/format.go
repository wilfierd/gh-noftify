package notify

import (
	"fmt"
	"strings"
	"time"

	"github.com/wilfierd/gh-notify/github"
)

func FormatInstantAlert(result *github.CheckResult) (*DiscordMessage, error) {
	if !result.HasAlerts() {
		return nil, nil
	}

	var fields []Field
	alertCount := result.GetAlertCount()

	// PRs needing review
	if len(result.PRsNeedingReview) > 0 {
		var prList []string
		for _, pr := range result.PRsNeedingReview {
			prList = append(prList, fmt.Sprintf("• [#%d %s](%s)", pr.Number, pr.Title, pr.HTMLURL))
		}
		fields = append(fields, Field{
			Name:   "🔍 PRs waiting for your review",
			Value:  strings.Join(prList, "\n"),
			Inline: false,
		})
	}

	// Stale own PRs
	if len(result.StaleOwnPRs) > 0 {
		var prList []string
		for _, pr := range result.StaleOwnPRs {
			daysSince := int(time.Since(pr.UpdatedAt).Hours() / 24)
			prList = append(prList, fmt.Sprintf("• [#%d %s](%s) (%d days old)",
				pr.Number, pr.Title, pr.HTMLURL, daysSince))
		}
		fields = append(fields, Field{
			Name:   "⏰ Your PRs need attention",
			Value:  strings.Join(prList, "\n"),
			Inline: false,
		})
	}

	// Assigned issues
	if len(result.AssignedIssues) > 0 {
		var issueList []string
		for _, issue := range result.AssignedIssues {
			issueList = append(issueList, fmt.Sprintf("• [#%d %s](%s)",
				issue.Number, issue.Title, issue.HTMLURL))
		}
		fields = append(fields, Field{
			Name:   "📋 Issues assigned to you",
			Value:  strings.Join(issueList, "\n"),
			Inline: false,
		})
	}

	// Unread notifications
	if len(result.UnreadNotifications) > 0 {
		var notifList []string
		for _, notif := range result.UnreadNotifications {
			notifList = append(notifList, fmt.Sprintf("• %s: %s (%s)",
				getNotificationIcon(notif.Reason), notif.Subject.Title, notif.Repository.Name))
		}
		// Show only first 5 to avoid spam
		if len(notifList) > 5 {
			notifList = notifList[:5]
			notifList = append(notifList, fmt.Sprintf("... and %d more", len(result.UnreadNotifications)-5))
		}
		fields = append(fields, Field{
			Name:   "📬 Unread notifications",
			Value:  strings.Join(notifList, "\n"),
			Inline: false,
		})
	}

	// Failed workflows
	if len(result.FailedWorkflows) > 0 {
		var workflowList []string
		for _, workflow := range result.FailedWorkflows {
			workflowList = append(workflowList, fmt.Sprintf("• [%s](%s) in %s ❌",
				workflow.Name, workflow.HTMLURL, workflow.Repository.Name))
		}
		fields = append(fields, Field{
			Name:   "🚨 Failed workflows",
			Value:  strings.Join(workflowList, "\n"),
			Inline: false,
		})
	}

	return &DiscordMessage{
		Embeds: []Embed{
			{
				Title:       fmt.Sprintf("🔔 GitHub Alerts (%d items)", alertCount),
				Description: "Here are some items that need your attention:",
				Color:       ColorOrange,
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields:      fields,
				Footer: &Footer{
					Text: "GitHub Notifier",
				},
			},
		},
	}, nil
}

func FormatDailyDigest(digest *github.DailyDigest, username string) (*DiscordMessage, error) {
	var fields []Field

	dateStr := digest.Date.Format("2006-01-02")

	// PRs opened
	if len(digest.PRsOpened) > 0 {
		var prList []string
		for _, pr := range digest.PRsOpened {
			prList = append(prList, fmt.Sprintf("• [#%d %s](%s)", pr.Number, pr.Title, pr.HTMLURL))
		}
		fields = append(fields, Field{
			Name:   "📤 PRs you opened",
			Value:  strings.Join(prList, "\n"),
			Inline: false,
		})
	}

	// PRs merged
	if len(digest.PRsMerged) > 0 {
		var prList []string
		for _, pr := range digest.PRsMerged {
			prList = append(prList, fmt.Sprintf("• [#%d %s](%s)", pr.Number, pr.Title, pr.HTMLURL))
		}
		fields = append(fields, Field{
			Name:   "✅ PRs merged",
			Value:  strings.Join(prList, "\n"),
			Inline: false,
		})
	}

	// Reviews given
	if len(digest.ReviewsGiven) > 0 {
		fields = append(fields, Field{
			Name:   "🔍 Reviews given",
			Value:  fmt.Sprintf("%d reviews completed", len(digest.ReviewsGiven)),
			Inline: true,
		})
	}

	// Issues resolved
	if len(digest.IssuesResolved) > 0 {
		fields = append(fields, Field{
			Name:   "🐛 Issues resolved",
			Value:  fmt.Sprintf("%d issues closed", len(digest.IssuesResolved)),
			Inline: true,
		})
	}

	// Failed workflows
	if len(digest.FailedWorkflows) > 0 {
		var workflowList []string
		for _, workflow := range digest.FailedWorkflows {
			workflowList = append(workflowList, fmt.Sprintf("• [%s](%s) in %s ❌",
				workflow.Name, workflow.HTMLURL, workflow.Repository.Name))
		}
		fields = append(fields, Field{
			Name:   "🚨 Failed workflows",
			Value:  strings.Join(workflowList, "\n"),
			Inline: false,
		})
	}

	// If no activity, add a default message
	if len(fields) == 0 {
		fields = append(fields, Field{
			Name:   "🌙 Quiet day",
			Value:  "No significant GitHub activity today",
			Inline: false,
		})
	}

	return &DiscordMessage{
		Embeds: []Embed{
			{
				Title:       fmt.Sprintf("🧠 Daily Digest – %s", dateStr),
				Description: fmt.Sprintf("Here's your GitHub activity summary for today, %s!", username),
				Color:       ColorPurple,
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields:      fields,
				Footer: &Footer{
					Text: "GitHub Notifier • Daily Report",
				},
			},
		},
	}, nil
}

func getNotificationIcon(reason string) string {
	switch reason {
	case "review_requested":
		return "👀"
	case "mention":
		return "💬"
	case "assign":
		return "📋"
	case "comment":
		return "💭"
	case "push":
		return "📤"
	case "ci_activity":
		return "🔧"
	default:
		return "🔔"
	}
}

func FormatCommitNotification(sha, message, author, repoName, commitURL, repoURL string) (*DiscordMessage, error) {
	shortSHA := sha
	if len(sha) > 7 {
		shortSHA = sha[:7]
	}

	return &DiscordMessage{
		Embeds: []Embed{
			{
				Title:       "📝 New Commit Pushed",
				Description: fmt.Sprintf("Here's the latest commit from **%s**!", author),
				Color:       ColorBlue,
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields: []Field{
					{
						Name:   "🚀 Commit Details",
						Value:  fmt.Sprintf("**[%s](%s)** %s", shortSHA, commitURL, message),
						Inline: false,
					},
					{
						Name:   "👤 Author",
						Value:  author,
						Inline: true,
					},
					{
						Name:   "📂 Repository",
						Value:  fmt.Sprintf("[%s](%s)", repoName, repoURL),
						Inline: true,
					},
				},
				Footer: &Footer{
					Text: "GitHub Notifier • Commit Tracker",
				},
			},
		},
	}, nil
}

func FormatSimpleAlert(title, message string) string {
	return fmt.Sprintf("🔔 **%s**\n%s", title, message)
}

func FormatErrorMessage(err error) string {
	return fmt.Sprintf("🚨 **Error**\n```\n%s\n```", err.Error())
}
