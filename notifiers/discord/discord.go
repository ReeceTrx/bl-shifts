package discord

import (
	"bl-shifts/notifiers"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type DiscordNotifier struct {
	WebhookURL string
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

// NewNotifier creates a new DiscordNotifier with the given webhook URL.
func NewNotifier(webhookURL string) notifiers.Notifier {
	return &DiscordNotifier{
		WebhookURL: webhookURL,
	}
}

// Send sends the given shift codes to the Discord webhook in batches of 10.
func (d *DiscordNotifier) Send(codes []string) error {
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
		req, err := http.NewRequest(http.MethodPost, d.WebhookURL, strings.NewReader(string(payload)))
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
