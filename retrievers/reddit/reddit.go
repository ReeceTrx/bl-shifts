package reddit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RedditRetriever struct {
	Subreddit    string
	ClientID     string
	ClientSecret string
	UserAgent    string
}

func NewRetriever(subreddit, clientID, clientSecret, userAgent string) *RedditRetriever {
	return &RedditRetriever{
		Subreddit:    subreddit,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		UserAgent:    userAgent,
	}
}

// Reddit API response for /new posts
type RedditPost struct {
	Data struct {
		Children []struct {
			Data struct {
				Title string `json:"title"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// GetCodes fetches the latest posts from the subreddit
func (r *RedditRetriever) GetCodes() ([]string, error) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/new.json?limit=10", r.Subreddit)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", r.UserAgent)

	// If you have ClientID and ClientSecret, you can use OAuth token here
	// For now, just authenticated with User-Agent (works for public subreddits)
	// For full OAuth, you'd need to implement token retrieval

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected error code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result RedditPost
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	codes := []string{}
	for _, child := range result.Data.Children {
		title := child.Data.Title
		codes = append(codes, title)
	}

	return codes, nil
}
