package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type SubredditResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				Text string `json:"selftext"`
				Name string `json:"name"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type DiscordWebhookMessage struct {
	Embed []DiscordWebhookEmbed `json:"embeds"`
}

type DiscordWebhookEmbed struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Fields      []DiscordWebhookField `json:"fields"`
}

type DiscordWebhookField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

const SUBREDDIT_URL = "https://www.reddit.com/r/borderlandsshiftcodes.json"

var discordWebhook = os.Getenv("DISCORD_WEBHOOK_URL")
var redisAddr = os.Getenv("REDIS_ADDR")

var (
	rdb *redis.Client
	ctx = context.Background()
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

	shiftCodeRe := regexp.MustCompile(`(\w{5}-){4}\w{5}`)
	srResponse, err := getLatestPosts()
	if err != nil {
		slog.Error("failed to get latest posts", "error", err)
		os.Exit(1)
	}
	allCodes := []string{}
	for _, child := range srResponse.Data.Children {
		matches := shiftCodeRe.FindAllString(child.Data.Text, -1)
		if len(matches) == 0 {
			continue
		}
		for _, code := range matches {
			slog.Debug("found shift code", "code", code)
			allCodes = append(allCodes, code)
		}
	}
	codesToSend, err := storeAndFilterCodes(allCodes)
	if err != nil {
		slog.Error("failed to get latest posts", "error", err)
		os.Exit(1)
	}
	if len(codesToSend) == 0 {
		slog.Info("no new shift codes found")
		return
	}
	slog.Info("found new shift codes", "codes", strings.Join(codesToSend, ", "))
	if discordWebhook != "" {
		err = sendToDiscord(codesToSend)
		if err != nil {
			slog.Error("failed to send to discord", "error", err)
			os.Exit(1)
		}
		slog.Info("sent shift codes to discord")
	}
}
func storeAndFilterCodes(allCodes []string) ([]string, error) {
	codesToSend := []string{}

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

func getLatestPosts() (*SubredditResponse, error) {
	newUrl := fmt.Sprintf("%s?sort=new&t=day", SUBREDDIT_URL)
	req, err := http.NewRequest(http.MethodGet, newUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "shift-code-fetcher/0.1 by ImDevinC")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected error code: %d", resp.StatusCode)
	}
	var srResponse SubredditResponse
	if err := json.NewDecoder(resp.Body).Decode(&srResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &srResponse, nil
}

func sendToDiscord(codes []string) error {
	batchSize := 10
	for i := 0; i < len(codes); i += batchSize {
		end := min(i+batchSize, len(codes))
		batch := codes[i:end]
		message := DiscordWebhookMessage{
			Embed: []DiscordWebhookEmbed{
				{
					Title:       "New Shift Codes",
					Description: "Here are the latest shift codes",
					Fields:      []DiscordWebhookField{},
				},
			},
		}
		for _, code := range batch {
			message.Embed[0].Fields = append(message.Embed[0].Fields, DiscordWebhookField{
				Name:   "Code",
				Value:  code,
				Inline: false,
			})
		}
		payload, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		req, err := http.NewRequest(http.MethodPost, discordWebhook, strings.NewReader(string(payload)))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}
	return nil
}
