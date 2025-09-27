package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GitHubToken     string
	DiscordWebhook  string
	Username        string
	CheckInterval   time.Duration
	DailyReportTime string
	CacheFile       string
	Timezone        string
	TrackAllCommits bool     // Enable tracking commits from all repositories
	TrackCommitsRealtime bool // Enable real-time commit tracking in instant checks
	TrackedRepositories []string // Specific repositories to track (empty = all repos)
	CommitLookbackMinutes int  // How far back to check for commits in minutes
}

func Load() *Config {
	checkInterval, _ := time.ParseDuration(getEnvOrDefault("CHECK_INTERVAL", "5m"))
	lookbackMinutes := 120 // Default 2 hours
	if val := getEnvOrDefault("COMMIT_LOOKBACK_MINUTES", ""); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			lookbackMinutes = parsed
		}
	}

	// Parse tracked repositories from comma-separated list
	var trackedRepos []string
	if repoList := getEnvOrDefault("TRACKED_REPOSITORIES", ""); repoList != "" {
		for _, repo := range strings.Split(repoList, ",") {
			repo = strings.TrimSpace(repo)
			if repo != "" {
				trackedRepos = append(trackedRepos, repo)
			}
		}
	}

	return &Config{
		GitHubToken:     getEnvOrDefault("GITHUB_TOKEN", ""),
		DiscordWebhook:  getEnvOrDefault("DISCORD_WEBHOOK", ""),
		Username:        getEnvOrDefault("GITHUB_USERNAME", ""),
		CheckInterval:   checkInterval,
		DailyReportTime: getEnvOrDefault("DAILY_REPORT_TIME", "02:00"), // 9h sáng VN = 2h UTC
		CacheFile:       getEnvOrDefault("CACHE_FILE", "cache.json"),
		Timezone:        getEnvOrDefault("TIMEZONE", "Asia/Ho_Chi_Minh"),
		TrackAllCommits: GetBoolEnv("TRACK_ALL_COMMITS", true), // Default enabled for daily digests
		TrackCommitsRealtime: GetBoolEnv("TRACK_COMMITS_REALTIME", false), // Default disabled for instant checks
		TrackedRepositories: trackedRepos, // Empty means track all repositories
		CommitLookbackMinutes: lookbackMinutes,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
