package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Login string `json:"login"`
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
	SHA       string    `json:"sha"`
	Message   string    `json:"message"`
	Author    User      `json:"author"`
	Committer User      `json:"committer"`
	Date      time.Time `json:"date"`
	URL       string    `json:"html_url"`
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
	url := fmt.Sprintf("%s/notifications?all=false", c.baseURL)

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

func (c *Client) GetRecentWorkflowRuns(username string) ([]WorkflowRun, error) {
	// This is a simplified version - in practice you'd need to iterate through repos
	url := fmt.Sprintf("%s/search/issues?q=type:pr+author:%s+created:>%s",
		c.baseURL, username, time.Now().AddDate(0, 0, -1).Format("2006-01-02"))

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	defer resp.Body.Close()

	// This would need to be implemented properly with actual workflow API
	return []WorkflowRun{}, nil
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
		}
	}

	return commits, nil
}
