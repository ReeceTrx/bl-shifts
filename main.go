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
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

func main() {
	cfg := LoadConfig()

	// Setup storage
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

	// Setup Reddit retriever
	retrieversList := []retrievers.Retriever{
		reddit.NewRetriever(
			"borderlandsshiftcodes",
			cfg.RedditClientID,
			cfg.RedditClientSecret,
			cfg.RedditUserAgent,
		),
	}

	// Setup Discord notifier
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
		slog.Info("checking for new shift codes")
		lastRunError = false
		runs++

		allCodes := []string{}
		var postTimestamp float64
		var postTitle string
		var redditErr error

		// Get codes from each retriever
		for _, retriever := range retrieversList {
			codes, createdUTC, title, err := retriever.GetCodes()
			if err != nil {
				slog.Error("failed to get codes from Reddit", "error", err)
				fmt.Println("Reddit fetch error:", err) // log in workflow
				redditErr = err
				lastRunError = true
				continue
			}
			fmt.Println("Reddit post title:", title)
			fmt.Println("Codes found:", codes)
			allCodes = append(allCodes, codes...)
			if createdUTC != 0 {
				postTimestamp = createdUTC
				postTitle = title
			}
		}

		if redditErr != nil && len(allCodes) == 0 {
			for _, notifier := range notifiersList {
				err := notifier.Send([]string{fmt.Sprintf("⚠️ Reddit fetch failed: %v", redditErr)})
				if err != nil {
					slog.Error("failed to send Discord notification", "error", err)
				}
			}
			continue
		}

		// Remove duplicates
		existingCodes := map[string]bool{}
		finalCodes := []string{}
		for _, code := range allCodes {
			if !existingCodes[code] {
				finalCodes = append(finalCodes, code)
				existingCodes[code] = true
			}
		}

		// Save new codes
		ctx := context.Background()
		codesToSend, err := store.FilterAndSaveCodes(ctx, finalCodes)
		if err != nil {
			slog.Error("failed to filter codes", "error", err)
			lastRunError = true
			continue
		}
		if len(codesToSend) == 0 {
			slog.Info("no new shift codes found in the latest post")
			continue
		}

		// Format post age
		postAge := ""
		if postTimestamp != 0 {
			duration := time.Since(time.Unix(int64(postTimestamp), 0))
			postAge = fmt.Sprintf("%.0f minutes ago", duration.Minutes())
		}

		// Prepare Discord message
		message := fmt.Sprintf(
			"**New Shift Codes**\nHere are the latest shift codes, redeem at https://shift.gearboxsoftware.com/rewards\n\n**Post:** %s\n%s",
			postTitle,
			strings.Join(codesToSend, "\n"),
		)
		if postAge != "" {
			message += fmt.Sprintf("\n\n*Post age: %s*", postAge)
		}

		slog.Info("sending new shift codes", "codes", strings.Join(codesToSend, ", "))

		// Send to Discord
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
