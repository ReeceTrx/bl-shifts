package main

import (
	"flag"
	"os"
	"strconv"
)

type config struct {
	RedisAddr         string
	DiscordWebhookURL string
	Filename          string
	IntervalMinutes   int
	RedditClientID    string
	RedditClientSecret string
	RedditUserAgent   string
}

func LoadConfig() *config {
	var redisAddr string
	var filename string
	var discordWebhookURL string
	var intervalMinutes int
	var redditClientID string
	var redditClientSecret string
	var redditUserAgent string

	// Command-line flags (optional)
	flag.StringVar(&redisAddr, "redis-addr", "", "Address of the Redis server (e.g., localhost:6379)")
	flag.StringVar(&filename, "filename", "", "Path to the file to store seen codes")
	flag.StringVar(&discordWebhookURL, "discord-webhook-url", "", "Discord webhook URL for notifications")
	flag.IntVar(&intervalMinutes, "interval-minutes", 0, "Interval in minutes between checks (default: run once and exit)")
	flag.StringVar(&redditClientID, "reddit-client-id", "", "Reddit client ID")
	flag.StringVar(&redditClientSecret, "reddit-client-secret", "", "Reddit client secret")
	flag.StringVar(&redditUserAgent, "reddit-user-agent", "", "Reddit user agent")
	flag.Parse()

	// Override with environment variables if present
	if envRedisAddr := os.Getenv("REDIS_ADDR"); envRedisAddr != "" {
		redisAddr = envRedisAddr
	}
	if envFilename := os.Getenv("FILENAME"); envFilename != "" {
		filename = envFilename
	}
	if envDiscordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL"); envDiscordWebhookURL != "" {
		discordWebhookURL = envDiscordWebhookURL
	}
	if envInterval := os.Getenv("INTERVAL_MINUTES"); envInterval != "" {
		if val, err := strconv.Atoi(envInterval); err == nil {
			intervalMinutes = val
		}
	}
	if envRedditID := os.Getenv("REDDIT_CLIENT_ID"); envRedditID != "" {
		redditClientID = envRedditID
	}
	if envRedditSecret := os.Getenv("REDDIT_CLIENT_SECRET"); envRedditSecret != "" {
		redditClientSecret = envRedditSecret
	}
	if envRedditUA := os.Getenv("REDDIT_USER_AGENT"); envRedditUA != "" {
		redditUserAgent = envRedditUA
	}

	return &config{
		RedisAddr:          redisAddr,
		DiscordWebhookURL:  discordWebhookURL,
		Filename:           filename,
		IntervalMinutes:    intervalMinutes,
		RedditClientID:     redditClientID,
		RedditClientSecret: redditClientSecret,
		RedditUserAgent:    redditUserAgent,
	}
}
