package hnscrape

import (
	"context"
	"fmt"
)

// GetUpdates fetches the changed items and profiles
func (c *Client) GetUpdates(ctx context.Context) (*Updates, error) {
	var updates Updates

	if err := c.get(ctx, "/updates", &updates); err != nil {
		return nil, fmt.Errorf("failed to get updates: %w", err)
	}

	return &updates, nil
}

// GetChangedItems fetches only the changed items
func (c *Client) GetChangedItems(ctx context.Context) ([]int64, error) {
	updates, err := c.GetUpdates(ctx)
	if err != nil {
		return nil, err
	}

	return updates.Items, nil
}

// GetChangedProfiles fetches only the changed profiles
func (c *Client) GetChangedProfiles(ctx context.Context) ([]string, error) {
	updates, err := c.GetUpdates(ctx)
	if err != nil {
		return nil, err
	}

	return updates.Profiles, nil
}

// GetChangedItemsWithDetails fetches changed items with full details
func (c *Client) GetChangedItemsWithDetails(ctx context.Context) ([]*Item, error) {
	updates, err := c.GetUpdates(ctx)
	if err != nil {
		return nil, err
	}

	if len(updates.Items) == 0 {
		return []*Item{}, nil
	}

	return c.GetItems(ctx, updates.Items)
}

// GetChangedProfilesWithDetails fetches changed profiles with full details
func (c *Client) GetChangedProfilesWithDetails(ctx context.Context) ([]*User, error) {
	updates, err := c.GetUpdates(ctx)
	if err != nil {
		return nil, err
	}

	if len(updates.Profiles) == 0 {
		return []*User{}, nil
	}

	return c.GetUsers(ctx, updates.Profiles)
}
