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
}

func LoadConfig() *config {
	var redisAddr string
	var filename string
	var discordWebhookURL string
	var intervalMinutes int
	flag.StringVar(&redisAddr, "redis-addr", "", "Address of the Redis server (e.g., localhost:6379)")
	flag.StringVar(&filename, "filename", "", "Path to the file to store seen codes")
	flag.StringVar(&discordWebhookURL, "discord-webhook-url", "", "Discord webhook URL for notifications")
	flag.IntVar(&intervalMinutes, "interval-minutes", 0, "Interval in minutes between checks (default: run once and exit)")
	flag.Parse()
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

	return &config{
		RedisAddr:         redisAddr,
		DiscordWebhookURL: discordWebhookURL,
		Filename:          filename,
		IntervalMinutes:   intervalMinutes,
	}
}
