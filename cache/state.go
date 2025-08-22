package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type State struct {
	LastCheck         time.Time            `json:"last_check"`
	LastDailyReport   time.Time            `json:"last_daily_report"`
	SentNotifications map[string]time.Time `json:"sent_notifications"`
	ProcessedPRs      map[string]bool      `json:"processed_prs"`
	ProcessedIssues   map[string]bool      `json:"processed_issues"`
	ProcessedNotifs   map[string]bool      `json:"processed_notifications"`
}

func NewState() *State {
	return &State{
		LastCheck:         time.Now(),
		LastDailyReport:   time.Now().AddDate(0, 0, -1), // Yesterday
		SentNotifications: make(map[string]time.Time),
		ProcessedPRs:      make(map[string]bool),
		ProcessedIssues:   make(map[string]bool),
		ProcessedNotifs:   make(map[string]bool),
	}
}

func LoadState(filepath string) (*State, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewState(), nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Initialize maps if they're nil
	if state.SentNotifications == nil {
		state.SentNotifications = make(map[string]time.Time)
	}

	return &state, nil
}

func (s *State) Save(filepath string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (s *State) IsNotificationSent(key string, cooldown time.Duration) bool {
	if lastSent, exists := s.SentNotifications[key]; exists {
		// Special case: if cooldown is 0, just check if it was ever sent (for workflow failures)
		if cooldown == 0 {
			fmt.Printf("DEBUG: One-time check for '%s': was sent before, blocking repeat\n", key)
			return true
		}

		timeSince := time.Since(lastSent)
		withinCooldown := timeSince < cooldown
		// Debug logging for cooldown check
		fmt.Printf("DEBUG: Cooldown check for '%s': last sent %.2f hours ago, cooldown %.2f hours, within cooldown: %t\n",
			key, timeSince.Hours(), cooldown.Hours(), withinCooldown)
		return withinCooldown
	}
	return false
}

func (s *State) MarkNotificationSent(key string) {
	s.SentNotifications[key] = time.Now()
}

func (s *State) CleanupOldEntries(maxAge time.Duration) bool {
	cutoff := time.Now().Add(-maxAge)
	removedAny := false

	// Clean up old notifications
	for key, timestamp := range s.SentNotifications {
		if timestamp.Before(cutoff) {
			delete(s.SentNotifications, key)
			removedAny = true
		}
	}

	return removedAny
}
