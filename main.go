package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-redis/redis/v8"
)

type SubredditResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				Text string `json:"selftext"`
				ID   string `json:"id"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type SubredditCommentResponse []struct {
	Data struct {
		Children []struct {
			Data struct {
				Body string `json:"body"`
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

const SUBREDDIT_URL = "https://www.reddit.com/r/borderlandsshiftcodes"

var (
	shiftCodeRe    = regexp.MustCompile(`(\w{5}-){4}\w{5}`)
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

	srResponse, err := getLatestPosts()
	if err != nil {
		slog.Error("failed to get latest posts", "error", err)
		os.Exit(1)
	}
	allCodes := getCodesFromPosts(srResponse)

	finalCodes := []string{}
	codeSet := map[string]bool{}

	for _, code := range allCodes {
		if !codeSet[code] {
			finalCodes = append(finalCodes, code)
			codeSet[code] = true
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
	if discordWebhook != "" {
		err = sendToDiscord(codesToSend)
		if err != nil {
			slog.Error("failed to send to discord", "error", err)
			os.Exit(1)
		}
		slog.Info("sent shift codes to discord")
	}
}

func getCodesFromPosts(srResponse *SubredditResponse) []string {
	codes := []string{}
	for _, child := range srResponse.Data.Children {
		matches := shiftCodeRe.FindAllString(child.Data.Text, -1)
		if len(matches) == 0 {
			slog.Debug("no shift codes found in post, checking comments", "post_id", child.Data.ID)
			comments, err := getCommentsForPost(child.Data.ID)
			if err != nil {
				slog.Error("failed to get comments for post", "error", err)
				continue
			}
			commentCodes := getCodesFromComments(comments)
			codes = append(codes, commentCodes...)
			continue
		}
		for _, code := range matches {
			slog.Debug("found shift code", "code", code)
			codes = append(codes, code)
		}
	}
	return codes
}

func getCodesFromComments(srResponse *SubredditCommentResponse) []string {
	codes := []string{}
	for _, comment := range *srResponse {
		for _, child := range comment.Data.Children {
			matches := shiftCodeRe.FindAllString(child.Data.Body, -1)
			if len(matches) == 0 {
				continue
			}
			for _, code := range matches {
				slog.Debug("found shift code", "code", code)
				codes = append(codes, code)
			}
		}
	}
	return codes
}

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

func getLatestPosts() (*SubredditResponse, error) {
	newUrl := fmt.Sprintf("%s.json?sort=new&t=day", SUBREDDIT_URL)
	req, err := http.NewRequest(http.MethodGet, newUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	response, err := doSubredditRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do subreddit request: %w", err)
	}

	var srResponse SubredditResponse
	err = json.Unmarshal(response, &srResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal subreddit response: %w", err)
	}
	return &srResponse, nil
}

func getCommentsForPost(postID string) (*SubredditCommentResponse, error) {
	commentsUrl := fmt.Sprintf("%s/comments/%s.json", SUBREDDIT_URL, postID)
	req, err := http.NewRequest(http.MethodGet, commentsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	respose, err := doSubredditRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do subreddit request: %w", err)
	}
	var commentsResponse SubredditCommentResponse
	err = json.Unmarshal(respose, &commentsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal comments response: %w", err)
	}
	return &commentsResponse, nil
}

func doSubredditRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", "shift-code-fetcher/0.1 by ImDevinC")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected error code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
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
