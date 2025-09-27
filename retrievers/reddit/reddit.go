package reddit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

// RedditRetriever holds Reddit API info
type RedditRetriever struct {
	Subreddit    string
	ClientID     string
	ClientSecret string
	UserAgent    string
}

// RedditPost is used to parse Reddit JSON response
type RedditPost struct {
	Data struct {
		Children []struct {
			Data struct {
				Title string `json:"title"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
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

// getToken retrieves an OAuth token from Reddit
func (r *RedditRetriever) getToken() (string, error) {
	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(r.ClientID, r.ClientSecret)
	req.Header.Set("User-Agent", r.UserAgent)
	q := req.URL.Query()
	q.Add("grant_type", "client_credentials")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get token, status: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

// GetCodes fetches the latest post and extracts shift codes
func (r *RedditRetriever) GetCodes() ([]string, error) {
	token, err := r.getToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://oauth.reddit.com/r/%s/new.json?limit=1", r.Subreddit)
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

	if len(result.Data.Children) == 0 {
		return nil, nil // no posts
	}

	newestTitle := result.Data.Children[0].Data.Title

	// Regex to match BL shift codes (5 blocks of 5 chars)
	codeRegex := regexp.MustCompile(`[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}`)
	codes := codeRegex.FindAllString(newestTitle, -1)

	return codes, nil
}
