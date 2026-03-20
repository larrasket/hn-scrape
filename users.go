package hnscrape

import (
	"context"
	"fmt"
)

// GetUser fetches a user by their username
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	var user User
	endpoint := fmt.Sprintf("/user/%s", username)

	if err := c.get(ctx, endpoint, &user); err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", username, err)
	}

	return &user, nil
}

// GetUsers fetches multiple users by their usernames concurrently
func (c *Client) GetUsers(ctx context.Context, usernames []string) ([]*User, error) {
	if len(usernames) == 0 {
		return []*User{}, nil
	}

	type result struct {
		user  *User
		index int
		err   error
	}

	resultChan := make(chan result, len(usernames))
	semaphore := make(chan struct{}, 10)

	for i, username := range usernames {
		go func(index int, name string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			user, err := c.GetUser(ctx, name)
			resultChan <- result{user: user, index: index, err: err}
		}(i, username)
	}

	users := make([]*User, len(usernames))
	for range usernames {
		res := <-resultChan
		if res.err != nil {
			return nil, fmt.Errorf("failed to get user %s: %w", usernames[res.index], res.err)
		}
		users[res.index] = res.user
	}

	return users, nil
}

// GetUserSubmissions fetches all items submitted by a user
func (c *Client) GetUserSubmissions(ctx context.Context,
	username string) ([]*Item, error) {
	user, err := c.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	if len(user.Submitted) == 0 {
		return []*Item{}, nil
	}

	return c.GetItems(ctx, user.Submitted)
}

// GetUserSubmissionsLimited fetches a limited number of items submitted by a
// user
func (c *Client) GetUserSubmissionsLimited(
	ctx context.Context, username string, limit int) ([]*Item, error) {
	user, err := c.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	if len(user.Submitted) == 0 {
		return []*Item{}, nil
	}

	ids := user.Submitted
	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}

	return c.GetItems(ctx, ids)
}
