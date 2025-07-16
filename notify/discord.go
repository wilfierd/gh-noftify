package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordMessage struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

type Embed struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Color       int     `json:"color,omitempty"`
	Timestamp   string  `json:"timestamp,omitempty"`
	Footer      *Footer `json:"footer,omitempty"`
	Author      *Author `json:"author,omitempty"`
	Fields      []Field `json:"fields,omitempty"`
}

type Footer struct {
	Text string `json:"text,omitempty"`
}

type Author struct {
	Name    string `json:"name,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordNotifier struct {
	webhookURL string
	httpClient *http.Client
}

func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (d *DiscordNotifier) SendMessage(message *DiscordMessage) error {
	if d.webhookURL == "" {
		return fmt.Errorf("discord webhook URL is not configured")
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := d.httpClient.Post(d.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord API error: %s", resp.Status)
	}

	return nil
}

func (d *DiscordNotifier) SendSimpleMessage(content string) error {
	return d.SendMessage(&DiscordMessage{
		Content: content,
	})
}

// SendEmbedMessage sends a Discord embed message. If authorName and authorAvatarURL are provided, sets the author/avatar.
func (d *DiscordNotifier) SendEmbedMessage(title, description string, color int, fields []Field, authorName, authorAvatarURL string) error {
	embed := Embed{
		Title:       title,
		Description: description,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields:      fields,
		Footer: &Footer{
			Text: "GitHub Notifier",
		},
	}

	// Add author avatar if available
	if authorName != "" && authorAvatarURL != "" {
		embed.Author = &Author{
			Name:    authorName,
			IconURL: authorAvatarURL,
		}
	}

	return d.SendMessage(&DiscordMessage{
		Embeds: []Embed{embed},
	})
}

// Color constants for Discord embeds
const (
	ColorRed    = 0xFF0000 // For errors/failures
	ColorYellow = 0xFFFF00 // For warnings
	ColorGreen  = 0x00FF00 // For success
	ColorBlue   = 0x0099FF // For info
	ColorPurple = 0x9966CC // For daily digest
	ColorOrange = 0xFF9900 // For alerts
)
