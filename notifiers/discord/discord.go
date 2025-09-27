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

// Send sends the message array to Discord
func (d *Notifier) Send(messages []string) error {
	for _, msg := range messages {
		payload := map[string]string{"content": msg} // Only send your message; no test line
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
