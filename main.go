package main

import (
	"bl-shifts/notifiers"
	"bl-shifts/notifiers/discord"
	"bl-shifts/retrievers"
	"bl-shifts/retrievers/reddit"
	"bl-shifts/store"
	"bl-shifts/store/file"
	"bl-shifts/store/redis"
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"
	"time"
)

func main() {
	cfg := LoadConfig()

	// Set up storage (Redis or file)
	var store store.Store
	if cfg.RedisAddr != "" && cfg.Filename == "" {
		store = redis.NewStore(cfg.RedisAddr)
	} else if cfg.Filename != "" && cfg.RedisAddr == "" {
		store = file.NewStore(cfg.Filename)
	} else if cfg.Filename == "" && cfg.RedisAddr == "" {
		slog.Warn("no storage method specified, defaulting to file 'codes.json'")
		store = file.NewStore("codes.json")
	} else {
		slog.Error("either REDIS_ADDR or FILENAME must be set, but not both")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Set up Reddit retriever for latest post
	retrievers := []retrievers.Retriever{
		reddit.NewRetriever(
			"borderlandsshiftcodes",
			cfg.RedditClientID,
			cfg.RedditClientSecret,
			cfg.RedditUserAgent,
		),
	}

	// Set up Discord notifier
	notifiersList := []notifiers.Notifier{}
	if cfg.DiscordWebhookURL != "" {
		notifiersList = append(notifiersList, discord.NewNotifier(cfg.DiscordWebhookURL))
	}

	lastRunError := false
	runs := 0

	for {
		if cfg.IntervalMinutes == 0 && runs > 0 {
			break
		}
		if runs > 0 {
			slog.Info("waiting for next interval", "minutes", cfg.IntervalMinutes)
			time.Sleep(time.Duration(cfg.IntervalMinutes) * time.Minute)
		}
		runs++
		slog.Info("checking for new shift codes")
		lastRunError = false

		allCodes := []string{}
		for _, retriever := range retrievers {
			codes, err := retriever.GetCodes()
			if err != nil {
				slog.Error("failed to get codes from retriever", "error", err)
				lastRunError = true
				continue
			}
			// Only use codes from the latest post
			allCodes = append(allCodes, codes...)
			slog.Info("codes retrieved from Reddit", "codes", codes)
		}

		if lastRunError {
			continue
		}

		// Filter already-sent codes
		ctx := context.Background()
		codesToSend, err := store.FilterAndSaveCodes(ctx, allCodes)
		if err != nil {
			slog.Error("failed to filter codes", "error", err)
			lastRunError = true
			continue
		}

		if len(codesToSend) == 0 {
			slog.Info("no new shift codes found")
			continue
		}

		// Prepare Discord message
		message := "🎁 New Shift Code(s):\n" + strings.Join(codesToSend, "\n") +
			"\n\nRedeem at https://shift.gearboxsoftware.com/rewards"

		for _, notifier := range notifiersList {
			err := notifier.Send([]string{message})
			if err != nil {
				slog.Error("failed to send notification", "error", err)
				lastRunError = true
			}
		}
	}

	if lastRunError {
		os.Exit(1)
	}
}
