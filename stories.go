package hnscrape

import (
	"context"
	"fmt"
)

// StoryType represents different types of story lists
type StoryType string

const (
	TopStories  StoryType = "topstories"
	NewStories  StoryType = "newstories"
	BestStories StoryType = "beststories"
	AskStories  StoryType = "askstories"
	ShowStories StoryType = "showstories"
	JobStories  StoryType = "jobstories"
)

// getStories is a generic method to fetch any type of story list
func (c *Client) getStories(ctx context.Context, storyType StoryType) (StoryList, error) {
	var stories StoryList
	endpoint := fmt.Sprintf("/%s", string(storyType))

	if err := c.get(ctx, endpoint, &stories); err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", storyType, err)
	}

	return stories, nil
}

// GetTopStories fetches up to 500 top stories (includes jobs)
func (c *Client) GetTopStories(ctx context.Context) (StoryList, error) {
	return c.getStories(ctx, TopStories)
}

// GetNewStories fetches up to 500 new stories
func (c *Client) GetNewStories(ctx context.Context) (StoryList, error) {
	return c.getStories(ctx, NewStories)
}

// GetBestStories fetches up to 500 best stories
func (c *Client) GetBestStories(ctx context.Context) (StoryList, error) {
	return c.getStories(ctx, BestStories)
}

// GetAskStories fetches up to 200 Ask HN stories
func (c *Client) GetAskStories(ctx context.Context) (StoryList, error) {
	return c.getStories(ctx, AskStories)
}

// GetShowStories fetches up to 200 Show HN stories
func (c *Client) GetShowStories(ctx context.Context) (StoryList, error) {
	return c.getStories(ctx, ShowStories)
}

// GetJobStories fetches up to 200 job stories
func (c *Client) GetJobStories(ctx context.Context) (StoryList, error) {
	return c.getStories(ctx, JobStories)
}

// getStoriesWithDetails is a generic method to fetch stories with full details
func (c *Client) getStoriesWithDetails(ctx context.Context, storyType StoryType, limit int) ([]*Item, error) {
	storyIDs, err := c.getStories(ctx, storyType)
	if err != nil {
		return nil, err
	}

	if limit > 0 && limit < len(storyIDs) {
		storyIDs = storyIDs[:limit]
	}

	return c.GetItems(ctx, storyIDs)
}

// GetTopStoriesWithDetails fetches top stories with full item details
func (c *Client) GetTopStoriesWithDetails(ctx context.Context, limit int) ([]*Item, error) {
	return c.getStoriesWithDetails(ctx, TopStories, limit)
}

// GetNewStoriesWithDetails fetches new stories with full item details
func (c *Client) GetNewStoriesWithDetails(ctx context.Context, limit int) ([]*Item, error) {
	return c.getStoriesWithDetails(ctx, NewStories, limit)
}

// GetBestStoriesWithDetails fetches best stories with full item details
func (c *Client) GetBestStoriesWithDetails(ctx context.Context, limit int) ([]*Item, error) {
	return c.getStoriesWithDetails(ctx, BestStories, limit)
}
