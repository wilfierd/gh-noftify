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

	// Repository invitations (only show NEW ones in instant alerts)
	if len(result.RepositoryInvitations) > 0 {
		var inviteList []string
		for _, invite := range result.RepositoryInvitations {
			if !invite.IsExpired() {
				daysLeft := invite.GetDaysUntilExpiration()
				var expiryText string
				if daysLeft == 0 {
					expiryText = "expires today"
				} else if daysLeft == 1 {
					expiryText = "expires tomorrow"
				} else {
					expiryText = fmt.Sprintf("expires in %d days", daysLeft)
				}
				inviteList = append(inviteList, fmt.Sprintf("• **%s** to [%s](%s) (%s)",
					invite.Inviter.Login, invite.Repository.FullName, invite.HTMLURL, expiryText))
			}
		}
		if len(inviteList) > 0 {
			fields = append(fields, Field{
				Name:   "📨 New Repository Invitations",
				Value:  strings.Join(inviteList, "\n"),
				Inline: false,
			})
		}
	}

	// Don't send empty notifications
	if len(fields) == 0 {
		return nil, nil
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

	var title, description string
	var color int

	if digest.IsEvening {
		// Evening Digest - Show accomplishments
		title = fmt.Sprintf("🌆 Evening Summary – %s", dateStr)
		description = fmt.Sprintf("Here's what you accomplished today, %s!", username)
		color = 0x00FF00 // Green for accomplishments

		hasActivity := false

		// PRs opened today
		if len(digest.PRsOpened) > 0 {
			var prList []string
			for _, pr := range digest.PRsOpened {
				prList = append(prList, fmt.Sprintf("• [#%d %s](%s)", pr.Number, pr.Title, pr.HTMLURL))
			}
			fields = append(fields, Field{
				Name:   "📤 Pull Requests Opened",
				Value:  strings.Join(prList, "\n"),
				Inline: false,
			})
			hasActivity = true
		}

		// PRs merged today
		if len(digest.PRsMerged) > 0 {
			var prList []string
			for _, pr := range digest.PRsMerged {
				prList = append(prList, fmt.Sprintf("• [#%d %s](%s)", pr.Number, pr.Title, pr.HTMLURL))
			}
			fields = append(fields, Field{
				Name:   "✅ Pull Requests Merged",
				Value:  strings.Join(prList, "\n"),
				Inline: false,
			})
			hasActivity = true
		}

		// Issues opened today
		if len(digest.IssuesOpened) > 0 {
			var issueList []string
			for _, issue := range digest.IssuesOpened {
				issueList = append(issueList, fmt.Sprintf("• [#%d %s](%s)", issue.Number, issue.Title, issue.HTMLURL))
			}
			fields = append(fields, Field{
				Name:   "🐛 Issues Opened",
				Value:  strings.Join(issueList, "\n"),
				Inline: false,
			})
			hasActivity = true
		}

		// Issues closed today
		if len(digest.IssuesClosed) > 0 {
			var issueList []string
			for _, issue := range digest.IssuesClosed {
				issueList = append(issueList, fmt.Sprintf("• [#%d %s](%s)", issue.Number, issue.Title, issue.HTMLURL))
			}
			fields = append(fields, Field{
				Name:   "✅ Issues Resolved",
				Value:  strings.Join(issueList, "\n"),
				Inline: false,
			})
			hasActivity = true
		}

		// Commit count
		if digest.CommitsToday > 0 {
			fields = append(fields, Field{
				Name:   "💻 Activity",
				Value:  fmt.Sprintf("%d commits/PRs today", digest.CommitsToday),
				Inline: true,
			})
			hasActivity = true
		}

		if !hasActivity {
			fields = append(fields, Field{
				Name:   "🌙 Quiet day",
				Value:  "No significant GitHub activity today",
				Inline: false,
			})
		}

	} else {
		// Morning Digest - Show what needs attention
		title = fmt.Sprintf("🌅 Morning Briefing – %s", dateStr)
		description = fmt.Sprintf("Good morning %s! Here's what needs your attention:", username)
		color = 0xFFAA00 // Orange for attention needed

		hasWork := false

		// Pending reviews
		if len(digest.PendingReviews) > 0 {
			var prList []string
			for _, pr := range digest.PendingReviews {
				prList = append(prList, fmt.Sprintf("• [#%d %s](%s)", pr.Number, pr.Title, pr.HTMLURL))
			}
			fields = append(fields, Field{
				Name:   "� Reviews Waiting",
				Value:  strings.Join(prList, "\n"),
				Inline: false,
			})
			hasWork = true
		}

		// Assigned issues
		if len(digest.AssignedIssues) > 0 {
			var issueList []string
			for _, issue := range digest.AssignedIssues {
				issueList = append(issueList, fmt.Sprintf("• [#%d %s](%s)", issue.Number, issue.Title, issue.HTMLURL))
			}
			fields = append(fields, Field{
				Name:   "📝 Issues Assigned",
				Value:  strings.Join(issueList, "\n"),
				Inline: false,
			})
			hasWork = true
		}

		// Repository invitations (show in morning digest with expiration info)
		if len(digest.RepositoryInvitations) > 0 {
			var inviteList []string
			for _, invite := range digest.RepositoryInvitations {
				if !invite.IsExpired() {
					daysLeft := invite.GetDaysUntilExpiration()
					var expiryText string
					if daysLeft == 0 {
						expiryText = "expires today"
					} else if daysLeft == 1 {
						expiryText = "expires tomorrow"
					} else {
						expiryText = fmt.Sprintf("expires in %d days", daysLeft)
					}
					inviteList = append(inviteList, fmt.Sprintf("• **%s** to [%s](%s) (%s)",
						invite.Inviter.Login, invite.Repository.FullName, invite.HTMLURL, expiryText))
				}
			}
			if len(inviteList) > 0 {
				fields = append(fields, Field{
					Name:   "📨 Pending Repository Invitations",
					Value:  strings.Join(inviteList, "\n"),
					Inline: false,
				})
				hasWork = true
			}
		}

		if !hasWork {
			fields = append(fields, Field{
				Name:   "✨ All clear!",
				Value:  "No pending reviews, assigned issues, or invitations",
				Inline: false,
			})
		}
	}

	// Failed workflows (show in both morning and evening)
	if len(digest.FailedWorkflows) > 0 {
		var workflowList []string
		for _, workflow := range digest.FailedWorkflows {
			workflowList = append(workflowList, fmt.Sprintf("• [%s](%s) in %s ❌",
				workflow.Name, workflow.HTMLURL, workflow.Repository.Name))
		}
		fields = append(fields, Field{
			Name:   "🚨 Failed Workflows",
			Value:  strings.Join(workflowList, "\n"),
			Inline: false,
		})
	}

	return &DiscordMessage{
		Embeds: []Embed{
			{
				Title:       title,
				Description: description,
				Color:       color,
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
	case "invitation":
		return "📩"
	case "repository_invitation":
		return "📨"
	case "team_mention":
		return "👥"
	case "security_alert":
		return "🚨"
	case "state_change":
		return "🔄"
	case "subscribed":
		return "🔔"
	case "author":
		return "✏️"
	default:
		return "🔔"
	}
}

func FormatCommitNotification(sha, message, author, repoName, commitURL, repoURL, avatarURL string) (*DiscordMessage, error) {
	shortSHA := sha
	if len(sha) > 7 {
		shortSHA = sha[:7]
	}

	embed := Embed{
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
				Name:   " Repository",
				Value:  fmt.Sprintf("[%s](%s)", repoName, repoURL),
				Inline: true,
			},
		},
		Footer: &Footer{
			Text: "GitHub Notifier • Commit Tracker",
		},
	}

	// Add author avatar if available
	if avatarURL != "" {
		embed.Author = &Author{
			Name:    author,
			IconURL: avatarURL,
		}
	}

	return &DiscordMessage{
		Embeds: []Embed{embed},
	}, nil
}

func FormatSimpleAlert(title, message string) string {
	return fmt.Sprintf("🔔 **%s**\n%s", title, message)
}

func FormatErrorMessage(err error) string {
	return fmt.Sprintf("🚨 **Error**\n```\n%s\n```", err.Error())
}
