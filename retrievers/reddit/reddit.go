package reddit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

// RedditRetriever represents a Reddit API retriever
type RedditRetriever struct {
	Subreddit    string
	ClientID     string
	ClientSecret string
	UserAgent    string
}

// NewRetriever creates a new RedditRetriever
func NewRetriever(subreddit, clientID, clientSecret, userAgent string) *RedditRetriever {
	return &RedditRetriever{
		Subreddit:    subreddit,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		UserAgent:    userAgent,
	}
}

// RedditPost struct for parsing Reddit JSON
type RedditPost struct {
	Data struct {
		Children []struct {
			Data struct {
				Title      string  `json:"title"`
				CreatedUTC float64 `json:"created_utc"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// GetCodes fetches the newest post and extracts up to 3 BL shift codes
func (r *RedditRetriever) GetCodes() ([]string, float64, error) {
	token, err := r.getToken()
	if err != nil {
		return nil, 0, err
	}

	url := fmt.Sprintf("https://oauth.reddit.com/r/%s/new.json?limit=1", r.Subreddit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", r.UserAgent)
	req.Header.Set("Authorization", "bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("unexpected error code: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var result RedditPost
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}

	if len(result.Data.Children) == 0 {
		return nil, 0, nil
	}

	newest := result.Data.Children[0].Data
	title := newest.Title
	created := newest.CreatedUTC

	// Regex: match 5 blocks of 5 characters
	codeRegex := regexp.MustCompile(`[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}`)
	codes := codeRegex.FindAllString(title, -1)

	// Only latest 3 codes
	if len(codes) > 3 {
		codes = codes[:3]
	}

	return codes, created, nil
}

// Dummy getToken function — replace with real OAuth token retrieval
func (r *RedditRetriever) getToken() (string, error) {
	return "YOUR_REDDIT_OAUTH_TOKEN", nil
}
