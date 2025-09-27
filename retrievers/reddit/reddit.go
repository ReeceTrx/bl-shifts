package reddit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// RedditRetriever fetches posts from a subreddit
type RedditRetriever struct {
	Subreddit    string
	ClientID     string
	ClientSecret string
	UserAgent    string
}

// RedditPost is the structure for parsing Reddit API response
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

// getToken fetches a valid OAuth token from Reddit
func (r *RedditRetriever) getToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(r.ClientID, r.ClientSecret)
	req.Header.Set("User-Agent", r.UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 403 || strings.Contains(string(body), "Whoa there") {
		return "", fmt.Errorf("reddit blocked the request (status %d). Possibly too many requests from this IP", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get Reddit token: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

// GetCodes fetches the latest post and extracts up to 3 BL shift codes
func (r *RedditRetriever) GetCodes() ([]string, float64, string, error) {
	token, err := r.getToken()
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to get Reddit token: %w", err)
	}

	url := fmt.Sprintf("https://oauth.reddit.com/r/%s/new.json?limit=1", r.Subreddit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, "", err
	}

	req.Header.Set("User-Agent", r.UserAgent)
	req.Header.Set("Authorization", "bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// Detect Reddit blocks / rate-limiting
	if resp.StatusCode == 403 || strings.Contains(string(body), "Whoa there") {
		return nil, 0, "", fmt.Errorf("reddit blocked the request (status %d). Possibly too many requests from this IP", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, 0, "", fmt.Errorf("unexpected error code from Reddit: %d", resp.StatusCode)
	}

	var result RedditPost
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, "", err
	}

	if len(result.Data.Children) == 0 {
		return nil, 0, "", nil // no posts
	}

	newest := result.Data.Children[0].Data
	title := newest.Title
	created := newest.CreatedUTC

	// Regex: match 5 blocks of 5 characters
	codeRegex := regexp.MustCompile(`[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}`)
	codes := codeRegex.FindAllString(title, -1)

	// Return only the latest 3 codes
	if len(codes) > 3 {
		codes = codes[:3]
	}

	return codes, created, title, nil
}

// NewRetriever creates a new RedditRetriever instance
func NewRetriever(subreddit, clientID, clientSecret, userAgent string) *RedditRetriever {
	return &RedditRetriever{
		Subreddit:    subreddit,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		UserAgent:    userAgent,
	}
}
