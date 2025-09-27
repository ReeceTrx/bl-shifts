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

	// Storage setup
	var storeInstance store.Store
	if cfg.RedisAddr != "" && cfg.Filename == "" {
		storeInstance = redis.NewStore(cfg.RedisAddr)
	} else if cfg.Filename != "" && cfg.RedisAddr == "" {
		storeInstance = file.NewStore(cfg.Filename)
	} else if cfg.Filename == "" && cfg.RedisAddr == "" {
		slog.Warn("no storage method specified, defaulting to file 'codes.json'")
		storeInstance = file.NewStore("codes.json")
	} else {
		slog.Error("either REDIS_ADDR or FILENAME must be set, but not both")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Reddit retriever
	retrieversList := []retrievers.Retriever{
		reddit.NewRetriever(
			"borderlandsshiftcodes",
			cfg.RedditClientID,
			cfg.RedditClientSecret,
			cfg.RedditUserAgent,
		),
	}

	// Discord notifier
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

		for _, retriever := range retrieversList {
			// Get codes and post timestamp
			codes, createdUTC, err := retriever.GetCodes()
			if err != nil {
				slog.Error("failed to get codes from retriever", "error", err)
				lastRunError = true
				continue
			}
			if len(codes) == 0 {
				slog.Info("no new codes found in latest post")
				continue
			}

			// Filter out already sent codes
			ctx := context.Background()
			codesToSend, err := storeInstance.FilterAndSaveCodes(ctx, codes)
			if err != nil {
				slog.Error("failed to filter codes", "error", err)
				lastRunError = true
				continue
			}
			if len(codesToSend) == 0 {
				slog.Info("all codes already sent")
				continue
			}

			// Calculate post age
			postTime := time.Unix(int64(createdUTC), 0)
			age := time.Since(postTime).Round(time.Minute)

			// Build Discord message
			message := "🎁 New Shift Code(s) - posted " + age.String() + " ago:\n" +
				strings.Join(codesToSend, "\n") +
				"\n\nRedeem at https://shift.gearboxsoftware.com/rewards"

			for _, notifier := range notifiersList {
				err := notifier.Send([]string{message})
				if err != nil {
					slog.Error("failed to send notification", "error", err)
					lastRunError = true
				}
			}

			slog.Info("sent codes to Discord", "codes", codesToSend)
		}
	}

	if lastRunError {
		os.Exit(1)
	}
}
