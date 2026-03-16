package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	dev        bool
	httpClient *http.Client
}

type Post struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
	Date string `json:"date"`
}

type postsResponse struct {
	Data  []Post `json:"data"`
	Total int    `json:"total"`
}

type authRequest struct {
	Strategy string `json:"strategy"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string `json:"accessToken"`
}

func NewClient(baseURL, token string, dev bool) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		dev:        dev,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	var rawBody []byte
	var bodyReader io.Reader
	if body != nil {
		var err error
		rawBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(rawBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	if c.dev {
		fmt.Printf("\n\033[33m→ %s %s\033[0m\n", method, c.baseURL+path)
		if rawBody != nil {
			fmt.Printf("  body: %s\n", truncate(string(rawBody), 500))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if c.dev {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		color := "\033[32m" // green
		if resp.StatusCode >= 400 {
			color = "\033[31m" // red
		}
		fmt.Printf("%s← %s\033[0m\n", color, resp.Status)
		if len(respBody) > 0 {
			fmt.Printf("  body: %s\n", truncate(string(respBody), 500))
		}
		fmt.Println()
	}

	return resp, nil
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func (c *Client) Login(email, password string) (string, error) {
	resp, err := c.do("POST", "/authentication", authRequest{
		Strategy: "local",
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed (%s)", resp.Status)
	}

	var ar authResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return "", err
	}
	return ar.AccessToken, nil
}

func (c *Client) GetPosts(limit int) ([]Post, error) {
	path := fmt.Sprintf("/posts?$sort[date]=-1&$limit=%d", limit)
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get posts failed (%s)", resp.Status)
	}

	var pr postsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	return pr.Data, nil
}

func (c *Client) CreatePost(body string, date time.Time) (*Post, error) {
	payload := map[string]any{
		"body": body,
		"date": date.Format("2006-01-02 15:04:05"),
	}

	resp, err := c.do("POST", "/posts", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create post failed (%s)", resp.Status)
	}

	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, err
	}
	return &post, nil
}

func (c *Client) SearchPosts(query string, limit int) ([]Post, error) {
	q := url.QueryEscape("%" + query + "%")
	path := fmt.Sprintf("/posts?body[$like]=%s&$limit=%d", q, limit)

	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed (%s)", resp.Status)
	}

	var pr postsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	return pr.Data, nil
}

func (c *Client) GetPostsHistory(month, day string) ([]Post, error) {
	path := fmt.Sprintf("/posts-history?md=%s-%s", month, day)
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get posts history failed (%s)", resp.Status)
	}

	var posts []Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (c *Client) GetHistory() (map[string]any, error) {
	resp, err := c.do("GET", "/posts-history?get=months", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get history failed (%s)", resp.Status)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
