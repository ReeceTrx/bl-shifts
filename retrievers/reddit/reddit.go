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

type RedditPost struct {
	Data struct {
		Children []struct {
			Data struct {
				Title string `json:"title"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
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

	// Regex to match BL shift codes (letters/numbers, adjust pattern if needed)
	codeRegex := regexp.MustCompile(`\b[A-Z0-9]{3,10}\b`)

	codes := []string{}
	for _, child := range result.Data.Children {
		matches := codeRegex.FindAllString(child.Data.Title, -1)
		for _, code := range matches {
			codes = append(codes, code)
		}
	}

	return codes, nil
}
