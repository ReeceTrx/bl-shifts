package main

import (
	"bl-shifts/notifiers"
	"bl-shifts/notifiers/discord"
	"bl-shifts/retrievers"
	"bl-shifts/retrievers/reddit"
	"bl-shifts/store/redis"
	"context"
	"log/slog"
	"os"
	"strings"
)

var (
	redisAddr      = os.Getenv("REDIS_ADDR")
	discordWebhook = os.Getenv("DISCORD_WEBHOOK_URL")
)

func main() {
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if discordWebhook == "" {
		slog.Warn("DISCORD_WEBHOOK_URL not set, will not send notifications")
	}

	store := redis.NewStore(redisAddr)
	retrievers := []retrievers.Retriever{
		reddit.NewRetriever("borderlandsshiftcodes"),
	}
	notifiers := []notifiers.Notifier{}
	if discordWebhook != "" {
		notifiers = append(notifiers, discord.NewNotifier(discordWebhook))
	}

	allCodes := []string{}

	for _, retriever := range retrievers {
		codes, err := retriever.GetCodes()
		if err != nil {
			slog.Error("failed to get codes from retriever", "error", err)
			continue
		}
		allCodes = append(allCodes, codes...)
	}

	existingCodes := map[string]bool{}
	finalCodes := []string{}
	for _, code := range allCodes {
		if !existingCodes[code] {
			finalCodes = append(finalCodes, code)
			existingCodes[code] = true
		}
	}

	ctx := context.Background()
	codesToSend, err := store.FilterAndSaveCodes(ctx, finalCodes)
	if err != nil {
		slog.Error("failed to filter codes", "error", err)
		os.Exit(1)
	}
	if len(codesToSend) == 0 {
		slog.Info("no new shift codes found")
		return
	}
	slog.Info("found new shift codes", "codes", strings.Join(codesToSend, ", "))

	for _, notifier := range notifiers {
		err := notifier.Send(codesToSend)
		if err != nil {
			slog.Error("failed to send notification", "error", err)
		}
	}
}
