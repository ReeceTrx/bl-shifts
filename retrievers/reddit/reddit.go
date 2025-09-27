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

// RedditRetriever stores subreddit and Reddit API credentials
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

// RedditPost represents the structure of Reddit's /new.json response
type RedditPost struct {
	Data struct {
		Children []struct {
			Data struct {
				Title string `json:"title"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// getToken retrieves a Reddit OAuth token
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

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get token: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

// GetCodes fetches the latest posts and extracts valid BL shift codes
func (r *RedditRetriever) GetCodes() ([]string, error) {
	token, err := r.getToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://oauth.reddit.com/r/%s/new.json?limit=10", r.Subreddit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", r.UserAgent)
	req.Header.Set("Authorization", "bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected error code: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var result RedditPost
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Regex to match BL shift codes (adjust if needed)
	codeRegex := regexp.MustCompile(`\bBL[A-Z0-9]{2,10}\b`)

	codes := []string{}
	for _, child := range result.Data.Children {
		matches := codeRegex.FindAllString(child.Data.Title, -1)
		for _, code := range matches {
			codes = append(codes, code)
		}
	}

	return codes, nil
}
