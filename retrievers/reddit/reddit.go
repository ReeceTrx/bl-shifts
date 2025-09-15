package reddit

import (
	"bl-shifts/retrievers"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
)

type RedditRetriever struct {
	Subreddit string
}

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

var shiftCodeRe = regexp.MustCompile(`(\w{5}-){4}\w{5}`)

const SUBREDDIT_URL = "https://www.reddit.com/r/"

// NewRetriever creates a new RedditRetriever for the given subreddit.
func NewRetriever(subreddit string) retrievers.Retriever {
	return &RedditRetriever{
		Subreddit: subreddit,
	}
}

// GetCodes retrieves the latest shift codes from the subreddit.
func (r *RedditRetriever) GetCodes() ([]string, error) {
	srResponse, err := r.getLatestPosts()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest posts: %w", err)
	}
	allCodes := r.getCodesFromPosts(srResponse)

	finalCodes := []string{}
	codeSet := map[string]bool{}

	for _, code := range allCodes {
		if !codeSet[code] {
			finalCodes = append(finalCodes, code)
			codeSet[code] = true
		}
	}
	return finalCodes, nil
}

// getLatestPosts fetches the latest posts from the subreddit.
func (r *RedditRetriever) getLatestPosts() (*SubredditResponse, error) {
	newUrl := fmt.Sprintf("%s/%s.json?sort=new&t=day", SUBREDDIT_URL, r.Subreddit)
	req, err := http.NewRequest(http.MethodGet, newUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	response, err := r.doSubredditRequest(req)
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

// getCodesFromPosts extracts shift codes from the posts' text and comments if necessary.
func (r *RedditRetriever) getCodesFromPosts(srResponse *SubredditResponse) []string {
	codes := []string{}
	for _, child := range srResponse.Data.Children {
		matches := shiftCodeRe.FindAllString(child.Data.Text, -1)
		if len(matches) == 0 {
			slog.Debug("no shift codes found in post, checking comments", "post_id", child.Data.ID)
			comments, err := r.getCommentsForPost(child.Data.ID)
			if err != nil {
				slog.Error("failed to get comments for post", "error", err)
				continue
			}
			commentCodes := r.getCodesFromComments(comments)
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

// getCommentsForPost fetches the top level comments for a given post ID.
func (r *RedditRetriever) getCommentsForPost(postID string) (*SubredditCommentResponse, error) {
	commentsUrl := fmt.Sprintf("%s/%s/comments/%s.json", SUBREDDIT_URL, r.Subreddit, postID)
	req, err := http.NewRequest(http.MethodGet, commentsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	respose, err := r.doSubredditRequest(req)
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

// getCodesFromComments extracts shift codes from the comments.
func (r *RedditRetriever) getCodesFromComments(srResponse *SubredditCommentResponse) []string {
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

// doSubredditRequest performs the HTTP request to Reddit with appropriate headers.
func (r *RedditRetriever) doSubredditRequest(req *http.Request) ([]byte, error) {
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
