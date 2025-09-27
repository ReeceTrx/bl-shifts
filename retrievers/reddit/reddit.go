func (r *RedditRetriever) GetCodes() ([]string, error) {
    token, err := r.getToken()
    if err != nil {
        return nil, err
    }

    // Get only the newest post
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

    // Extract the title of the newest post
    newestTitle := result.Data.Children[0].Data.Title

    // Regex to match BL shift codes (5 blocks of 5 characters)
    codeRegex := regexp.MustCompile(`[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}`)
    codes := codeRegex.FindAllString(newestTitle, -1)

    return codes, nil
}
