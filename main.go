package main

import (
	"bl-shifts/notifiers"
	"bl-shifts/notifiers/discord"
	"bl-shifts/retrievers"
	"bl-shifts/retrievers/reddit"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
)

var (
	redisAddr      = os.Getenv("REDIS_ADDR")
	discordWebhook = os.Getenv("DISCORD_WEBHOOK_URL")
	rdb            *redis.Client
)

func main() {
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if discordWebhook == "" {
		slog.Warn("DISCORD_WEBHOOK_URL not set, will not send notifications")
	}

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

	codesToSend, err := storeAndFilterCodes(finalCodes)
	if err != nil {
		slog.Error("failed to get latest posts", "error", err)
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

// storeAndFilterCodes checks Redis for existing codes, stores new ones, and returns only the new codes
func storeAndFilterCodes(allCodes []string) ([]string, error) {
	codesToSend := []string{}

	ctx := context.Background()
	for _, code := range allCodes {
		exists, err := rdb.SIsMember(ctx, "shift_codes", code).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to check Redis for code %s: %w", code, err)
		}
		if !exists {
			codesToSend = append(codesToSend, code)
			rdb.SAdd(ctx, "shift_codes", code)
		}
	}

	return codesToSend, nil
}
