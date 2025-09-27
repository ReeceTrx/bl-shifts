package discord

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Notifier struct {
	WebhookURL string
}

func NewNotifier(webhookURL string) *Notifier {
	return &Notifier{
		WebhookURL: webhookURL,
	}
}

type Embed struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Color       int    `json:"color,omitempty"`
	Footer      *Footer `json:"footer,omitempty"`
}

type Footer struct {
	Text string `json:"text,omitempty"`
}

type WebhookPayload struct {
	Embeds []Embed `json:"embeds,omitempty"`
}

// Send sends the message array to Discord as embeds
func (d *Notifier) Send(messages []string) error {
	for _, msg := range messages {
		payload := WebhookPayload{
			Embeds: []Embed{
				{
					Title:       "New Shift Codes",
					Description: msg,
					Color:       0x00ff00, // green
					Footer:      &Footer{Text: "BL-Shifts"},
				},
			},
		}

		body, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", d.WebhookURL, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}
