package main

import (
	"flag"
	"os"
)

type config struct {
	RedisAddr         string
	DiscordWebhookURL string
	Filename          string
}

func LoadConfig() *config {
	var redisAddr string
	var filename string
	var discordWebhookURL string
	flag.StringVar(&redisAddr, "redis-addr", "", "Address of the Redis server (e.g., localhost:6379)")
	flag.StringVar(&filename, "filename", "", "Path to the file to store seen codes")
	flag.StringVar(&discordWebhookURL, "discord-webhook-url", "", "Discord webhook URL for notifications")
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

	return &config{
		RedisAddr:         redisAddr,
		DiscordWebhookURL: discordWebhookURL,
		Filename:          filename,
	}
}
