package config

import (
	"os"
	"strconv"
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
}

func Load() *Config {
	checkInterval, _ := time.ParseDuration(getEnvOrDefault("CHECK_INTERVAL", "5m"))

	return &Config{
		GitHubToken:     getEnvOrDefault("GITHUB_TOKEN", ""),
		DiscordWebhook:  getEnvOrDefault("DISCORD_WEBHOOK", ""),
		Username:        getEnvOrDefault("GITHUB_USERNAME", ""),
		CheckInterval:   checkInterval,
		DailyReportTime: getEnvOrDefault("DAILY_REPORT_TIME", "02:00"), // 9h sáng VN = 2h UTC
		CacheFile:       getEnvOrDefault("CACHE_FILE", "cache.json"),
		Timezone:        getEnvOrDefault("TIMEZONE", "Asia/Ho_Chi_Minh"),
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
