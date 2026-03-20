package hnscrape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	BaseURL        = "https://hacker-news.firebaseio.com/v0"
	DefaultTimeout = 30 * time.Second
)

// Exported error variables
var (
	// ErrItemDeleted is returned when an item has been deleted
	ErrItemDeleted = errors.New("item has been deleted")
)

// IsItemDeleted checks if an error indicates that an item was deleted
func IsItemDeleted(err error) bool {
	return errors.Is(err, ErrItemDeleted)
}

// Client represents a Hacker News API client
type Client struct {
	baseURL    string
	hnURL      string
	httpClient *http.Client
	userCookie string
}

// Option represents a client configuration option
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets a custom timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithBaseURL sets a custom base URL (useful for testing)
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHackerNewsURL sets a custom Hacker News base URL (useful for testing)
func WithHackerNewsURL(url string) Option {
	return func(c *Client) {
		c.hnURL = url
	}
}

// WithUserCookie sets a user cookie for accessing dead posts
func WithUserCookie(cookie string) Option {
	return func(c *Client) {
		c.userCookie = cookie
	}
}

// NewClient creates a new Hacker News API client
func NewClient(options ...Option) *Client {
	// Create cookie jar for handling cookies
	jar, _ := cookiejar.New(nil)

	client := &Client{
		baseURL: BaseURL,
		hnURL:   HackerNewsURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
			Jar:     jar,
		},
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// get performs a GET request to the API
func (c *Client) get(ctx context.Context, endpoint string, result any) error {
	url := fmt.Sprintf("%s%s.json", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
