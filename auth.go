package hnscrape

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Login authenticates with Hacker News and stores the session cookie on the
// client. After a successful login, the client can access dead and flagged
// posts via GetItem and GetItemWithScraping.
//
// The password is never stored; only the resulting session cookie is kept.
func (c *Client) Login(ctx context.Context, username, password string) error {
	loginURL := fmt.Sprintf("%s/login", c.hnURL)
	form := url.Values{
		"acct": {username},
		"pw":   {password},
		"goto": {"news"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Shallow-copy the HTTP client so we can disable redirect following
	// without affecting the shared instance. The Jar and Transport are
	// still shared, which is intentional.
	noRedirect := *c.httpClient
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := noRedirect.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "user" {
			c.userCookie = cookie.Value
			return nil
		}
	}

	return fmt.Errorf("login failed: no session cookie in response (wrong password?)")
}

// UserCookie returns the current session cookie value. An empty string means
// the client is not authenticated.
func (c *Client) UserCookie() string {
	return c.userCookie
}
