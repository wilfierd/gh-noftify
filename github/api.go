package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HTMLURL   string    `json:"html_url"`
	Draft     bool      `json:"draft"`
	Reviews   []Review  `json:"reviews,omitempty"`
}

type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	User      User      `json:"user"`
	Assignee  *User     `json:"assignee"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HTMLURL   string    `json:"html_url"`
	Comments  int       `json:"comments"`
}

type Review struct {
	ID          int       `json:"id"`
	State       string    `json:"state"`
	User        User      `json:"user"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type Notification struct {
	ID         string    `json:"id"`
	Unread     bool      `json:"unread"`
	Reason     string    `json:"reason"`
	UpdatedAt  time.Time `json:"updated_at"`
	Subject    Subject   `json:"subject"`
	Repository Repo      `json:"repository"`
}

type Subject struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type Repo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Fork     bool   `json:"fork"`
	Archived bool   `json:"archived"`
}

type Invitation struct {
	ID         int       `json:"id"`
	Repository Repo      `json:"repository"`
	Invitee    User      `json:"invitee"`
	Inviter    User      `json:"inviter"`
	CreatedAt  time.Time `json:"created_at"`
	HTMLURL    string    `json:"html_url"`
}

// GetExpirationDate calculates when the invitation expires (7 days from creation)
func (inv *Invitation) GetExpirationDate() time.Time {
	return inv.CreatedAt.AddDate(0, 0, 7) // GitHub repo invitations expire after 7 days
}

// GetDaysUntilExpiration returns how many days until the invitation expires
func (inv *Invitation) GetDaysUntilExpiration() int {
	expiration := inv.GetExpirationDate()
	daysLeft := int(time.Until(expiration).Hours() / 24)
	if daysLeft < 0 {
		return 0 // Already expired
	}
	return daysLeft
}

// IsExpired checks if the invitation has expired
func (inv *Invitation) IsExpired() bool {
	return time.Now().After(inv.GetExpirationDate())
}

type WorkflowRun struct {
	ID         int       `json:"id"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	CreatedAt  time.Time `json:"created_at"`
	HTMLURL    string    `json:"html_url"`
	Name       string    `json:"name"`
	Repository Repo      `json:"repository"`
}

type Commit struct {
	SHA        string    `json:"sha"`
	Message    string    `json:"message"`
	Author     User      `json:"author"`
	Committer  User      `json:"committer"`
	Date       time.Time `json:"date"`
	URL        string    `json:"html_url"`
	Repository Repo      `json:"repository,omitempty"`
}

type CommitResponse struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	HTMLURL string `json:"html_url"`
	Author  User   `json:"author"`
}

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.github.com",
	}
}

func (c *Client) makeRequest(method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *Client) GetUserPullRequests(username string) ([]PullRequest, error) {
	url := fmt.Sprintf("%s/search/issues?q=type:pr+author:%s+state:open", c.baseURL, username)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []PullRequest `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Items, nil
}

func (c *Client) GetReviewRequests(username string) ([]PullRequest, error) {
	url := fmt.Sprintf("%s/search/issues?q=type:pr+review-requested:%s+state:open", c.baseURL, username)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get review requests: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []PullRequest `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Items, nil
}

func (c *Client) GetAssignedIssues(username string) ([]Issue, error) {
	url := fmt.Sprintf("%s/search/issues?q=type:issue+assignee:%s+state:open", c.baseURL, username)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigned issues: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []Issue `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Items, nil
}

func (c *Client) GetNotifications() ([]Notification, error) {
	url := fmt.Sprintf("%s/notifications?all=false&participating=false", c.baseURL)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body first
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			// Handle specific permission errors
			if resp.StatusCode == http.StatusForbidden && errorResp.Message == "Resource not accessible by personal access token" {
				return nil, fmt.Errorf("GitHub token missing 'notifications' permission. Please regenerate token with proper permissions")
			}
			return nil, fmt.Errorf("GitHub API error: %s (status: %d)", errorResp.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var notifications []Notification
	if err := json.Unmarshal(body, &notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %w", err)
	}

	return notifications, nil
}

func (c *Client) GetRepositoryInvitations() ([]Invitation, error) {
	url := fmt.Sprintf("%s/user/repository_invitations", c.baseURL)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository invitations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var invitations []Invitation
	if err := json.NewDecoder(resp.Body).Decode(&invitations); err != nil {
		return nil, fmt.Errorf("failed to decode invitations: %w", err)
	}

	return invitations, nil
}

func (c *Client) GetRecentWorkflowRuns(username string) ([]WorkflowRun, error) {
	// Get user repositories first
	repos, err := c.GetUserRepositories(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repositories: %w", err)
	}

	var allFailedWorkflows []WorkflowRun

	// Check workflows for each repository (limit to avoid API rate limits)
	maxRepos := 5 // Conservative limit to avoid rate limits
	repoCount := 0

	for _, repo := range repos {
		if repoCount >= maxRepos {
			break
		}

		// Skip archived, forked, and private repositories to reduce API calls
		if repo.Archived || repo.Fork || repo.Private {
			continue
		}

		repoCount++

		// Get workflow runs for this repository (only failed ones, last 3 results)
		url := fmt.Sprintf("%s/repos/%s/actions/runs?status=failure&per_page=3",
			c.baseURL, repo.FullName)

		resp, err := c.makeRequest("GET", url, nil)
		if err != nil {
			// Log warning but continue with other repos
			fmt.Printf("Warning: failed to get workflow runs for %s: %v\n", repo.FullName, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// Skip repos where we don't have access to workflows (404, 403)
			continue
		}

		var workflowResponse struct {
			WorkflowRuns []struct {
				ID         int       `json:"id"`
				Name       string    `json:"name"`
				Status     string    `json:"status"`
				Conclusion string    `json:"conclusion"`
				CreatedAt  time.Time `json:"created_at"`
				HTMLURL    string    `json:"html_url"`
			} `json:"workflow_runs"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&workflowResponse); err != nil {
			// Log warning but continue
			fmt.Printf("Warning: failed to decode workflow runs for %s: %v\n", repo.FullName, err)
			continue
		}

		// Convert to WorkflowRun format and filter recent failures
		cutoff := time.Now().AddDate(0, 0, -3)
		for _, run := range workflowResponse.WorkflowRuns {
			if run.CreatedAt.After(cutoff) && run.Conclusion == "failure" {
				workflowRun := WorkflowRun{
					ID:         run.ID,
					Name:       run.Name,
					Status:     run.Status,
					Conclusion: run.Conclusion,
					CreatedAt:  run.CreatedAt,
					HTMLURL:    run.HTMLURL,
					Repository: repo,
				}
				allFailedWorkflows = append(allFailedWorkflows, workflowRun)
			}
		}
	}

	return allFailedWorkflows, nil
}

func (c *Client) GetUser() (*User, error) {
	url := fmt.Sprintf("%s/user", c.baseURL)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

func (c *Client) GetRecentCommits(repo string, since time.Time) ([]Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/commits?since=%s",
		c.baseURL, repo, since.Format(time.RFC3339))

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var commitResponses []CommitResponse
	if err := json.NewDecoder(resp.Body).Decode(&commitResponses); err != nil {
		return nil, fmt.Errorf("failed to decode commits: %w", err)
	}

	commits := make([]Commit, len(commitResponses))
	for i, cr := range commitResponses {
		commits[i] = Commit{
			SHA:     cr.SHA,
			Message: cr.Commit.Message,
			Author:  cr.Author,
			Date:    cr.Commit.Author.Date,
			URL:     cr.HTMLURL,
			Repository: Repo{
				FullName: repo,
				Name:     repo[strings.LastIndex(repo, "/")+1:],
			},
		}
	}

	return commits, nil
}

func (c *Client) GetUserByUsername(username string) (*User, error) {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, username)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", username, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

func (c *Client) GetUserIssues(username string) ([]Issue, error) {
	url := fmt.Sprintf("%s/search/issues?q=type:issue+author:%s", c.baseURL, username)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user issues: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []Issue `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Items, nil
}

func (c *Client) GetUserRepositories(username string) ([]Repo, error) {
	url := fmt.Sprintf("%s/users/%s/repos?type=all&sort=updated&per_page=100", c.baseURL, username)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode repositories: %w", err)
	}

	return repos, nil
}

func (c *Client) GetRecentCommitsFromAllRepos(username string, since time.Time) ([]Commit, error) {
	// Get all user repositories
	repos, err := c.GetUserRepositories(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repositories: %w", err)
	}

	var allCommits []Commit

	// Iterate through repositories and get recent commits
	for _, repo := range repos {
		// Skip archived and forked repositories to reduce noise
		if repo.Archived || repo.Fork {
			continue
		}

		commits, err := c.GetRecentCommits(repo.FullName, since)
		if err != nil {
			// Log error but continue with other repos
			fmt.Printf("Warning: failed to get commits for repo %s: %v\n", repo.FullName, err)
			continue
		}

		// Filter commits by the username to only include user's own commits
		for _, commit := range commits {
			if commit.Author.Login == username {
				// Set repository information
				commit.Repository = repo
				allCommits = append(allCommits, commit)
			}
		}
	}

	return allCommits, nil
}
