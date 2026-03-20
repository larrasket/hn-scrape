package hnscrape

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// GetItem fetches a single item by its ID
// If the item is dead and a user cookie is provided, it will attempt to scrape
// the content
func (c *Client) GetItem(ctx context.Context, id int64) (*Item, error) {
	var item Item
	endpoint := fmt.Sprintf("/item/%d", id)

	if err := c.get(ctx, endpoint, &item); err != nil {
		return nil, fmt.Errorf("failed to get item %d: %w", id, err)
	}

	// If item is already marked as deleted by the API, return it as is
	if item.Deleted {
		return &item, nil
	}

	if item.Dead && c.userCookie != "" {
		scrapedItem, err := c.scrapeDeadItem(ctx, id)

		// If scraping fails due to deletion, mark the original item as deleted
		if err != nil && errors.Is(err, ErrItemDeleted) {
			item.Deleted = true
			return &item, nil
		}

		if err != nil {
			return nil, fmt.
				Errorf("failed to scrape dead item %d with provided cookie: %w",
					id, err)
		}

		// If scraped item is deleted, mark as deleted
		if scrapedItem.Deleted {
			item.Deleted = true
			return &item, nil
		}

		return c.mergeItemData(&item, scrapedItem), nil
	}

	return &item, nil
}

// GetItemWithScraping explicitly attempts to scrape an item even if it's not
// marked as dead
func (c *Client) GetItemWithScraping(
	ctx context.Context, id int64) (*Item, error) {
	if c.userCookie == "" {
		return nil, fmt.Errorf("user cookie required for scraping")
	}

	apiItem, apiErr := c.GetItem(ctx, id)

	scrapedItem, scrapeErr := c.scrapeDeadItem(ctx, id)

	// Handle deletion errors specially
	if scrapeErr != nil && errors.Is(scrapeErr, ErrItemDeleted) {
		if apiItem != nil {
			apiItem.Deleted = true
			return apiItem, nil
		}
		return &Item{ID: id, Deleted: true}, nil
	}

	if apiErr != nil && scrapeErr != nil {
		return nil, fmt.
			Errorf("both API and scraping failed - API: %w, Scraping: %w",
				apiErr, scrapeErr)
	}

	if apiErr != nil && scrapeErr == nil {
		return scrapedItem, nil
	}

	if apiErr == nil && scrapeErr != nil {
		return apiItem, nil
	}

	return c.mergeItemData(apiItem, scrapedItem), nil
}

// mergeItemData merges data from API and scraped sources, preferring non-empty
// values
func (c *Client) mergeItemData(apiItem, scrapedItem *Item) *Item {
	merged := *apiItem

	if scrapedItem.Title != "" {
		merged.Title = scrapedItem.Title
	}
	if scrapedItem.Text != "" {
		merged.Text = scrapedItem.Text
	}
	if scrapedItem.URL != "" && scrapedItem.URL != "news" {
		merged.URL = scrapedItem.URL
	}
	if scrapedItem.By != "" {
		merged.By = scrapedItem.By
	}
	if scrapedItem.Score > 0 && scrapedItem.Score > merged.Score {
		merged.Score = scrapedItem.Score
	}
	if scrapedItem.Time > 0 {
		merged.Time = scrapedItem.Time
	}
	if scrapedItem.Parent > 0 {
		merged.Parent = scrapedItem.Parent
	}
	if len(scrapedItem.Kids) > 0 {
		merged.Kids = scrapedItem.Kids
	}

	if scrapedItem.Type == ItemTypeComment ||
		(merged.Parent > 0 && merged.Text != "" && merged.Title == "") {
		merged.Type = ItemTypeComment
	} else if scrapedItem.Type != "" {
		merged.Type = scrapedItem.Type
	}

	return &merged
}

// GetItems fetches multiple items by their IDs concurrently
func (c *Client) GetItems(ctx context.Context, ids []int64) ([]*Item, error) {
	if len(ids) == 0 {
		return []*Item{}, nil
	}

	// Use concurrent processing for better performance
	type result struct {
		item  *Item
		index int
		err   error
	}

	resultChan := make(chan result, len(ids))
	// Get concurrent limit from environment or use default
	concurrentLimit := 10 // default
	if val := os.Getenv("HNAPI_CONCURRENT_LIMIT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			concurrentLimit = parsed
		}
	}

	semaphore := make(chan struct{}, concurrentLimit)

	// Start goroutines for each item
	for i, id := range ids {
		go func(index int, itemID int64) {
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			item, err := c.GetItem(ctx, itemID)
			resultChan <- result{item: item, index: index, err: err}
		}(i, id)
	}

	// Collect results
	items := make([]*Item, len(ids))
	for i := 0; i < len(ids); i++ {
		res := <-resultChan
		if res.err != nil {
			return nil, fmt.Errorf("failed to get item %d: %w", ids[res.index], res.err)
		}
		items[res.index] = res.item
	}

	return items, nil
}

// GetMaxItemID fetches the current largest item ID
func (c *Client) GetMaxItemID(ctx context.Context) (int64, error) {
	var maxID int64

	if err := c.get(ctx, "/maxitem", &maxID); err != nil {
		return 0, fmt.Errorf("failed to get max item ID: %w", err)
	}

	return maxID, nil
}

// GetItemsFromTo fetches items in a range from startID to endID (inclusive)
// Note: This is a convenience method that's not part of the official API
func (c *Client) GetItemsFromTo(
	ctx context.Context, startID, endID int64) ([]*Item, error) {
	if startID > endID {
		return nil, fmt.Errorf("startID cannot be greater than endID")
	}

	ids := make([]int64, 0, endID-startID+1)
	for i := startID; i <= endID; i++ {
		ids = append(ids, i)
	}

	return c.GetItems(ctx, ids)
}
