//go:build integration

// Integration tests hit the real Hacker News API and website. Run with:
//
//	go test -tags integration -v ./...
//
// Tests that read dead/flagged posts require HN credentials:
//
//	HN_TEST_USER=youruser HN_TEST_PASS=yourpass go test -tags integration -v ./...
package hnscrape

import (
	"context"
	"os"
	"testing"
	"time"
)

// integrationClient returns a client for real-network tests.
func integrationClient(t *testing.T) *Client {
	t.Helper()
	return NewClient(WithTimeout(20 * time.Second))
}

// integrationClientAuthed returns a logged-in client. Skips if no credentials.
func integrationClientAuthed(t *testing.T) *Client {
	t.Helper()
	user := os.Getenv("HN_TEST_USER")
	pass := os.Getenv("HN_TEST_PASS")
	if user == "" || pass == "" {
		t.Skip("set HN_TEST_USER and HN_TEST_PASS to run authenticated tests")
	}
	c := integrationClient(t)
	if err := c.Login(context.Background(), user, pass); err != nil {
		t.Fatalf("Login: %v", err)
	}
	if c.UserCookie() == "" {
		t.Fatal("Login succeeded but UserCookie() is empty")
	}
	t.Logf("logged in as %s", user)
	return c
}

func TestIntegration_GetTopStories(t *testing.T) {
	c := integrationClient(t)
	ids, err := c.GetTopStories(context.Background())
	if err != nil {
		t.Fatalf("GetTopStories: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("expected at least one story")
	}
	t.Logf("top stories: %d IDs, first=%d", len(ids), ids[0])
}

func TestIntegration_GetItem(t *testing.T) {
	// Item 1 is the first HN post ever — it never changes.
	c := integrationClient(t)
	item, err := c.GetItem(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetItem(1): %v", err)
	}
	if item.ID != 1 {
		t.Errorf("ID = %d, want 1", item.ID)
	}
	if item.By == "" {
		t.Error("expected non-empty By")
	}
	t.Logf("item 1: by=%s title=%q url=%s", item.By, item.Title, item.URL)
}

func TestIntegration_GetMaxItemID(t *testing.T) {
	c := integrationClient(t)
	maxID, err := c.GetMaxItemID(context.Background())
	if err != nil {
		t.Fatalf("GetMaxItemID: %v", err)
	}
	if maxID < 1_000_000 {
		t.Errorf("maxID = %d; suspiciously low", maxID)
	}
	t.Logf("max item ID: %d", maxID)
}

func TestIntegration_GetUser(t *testing.T) {
	c := integrationClient(t)
	// pg is a permanent account that will always exist.
	user, err := c.GetUser(context.Background(), "pg")
	if err != nil {
		t.Fatalf("GetUser(pg): %v", err)
	}
	if user.ID != "pg" {
		t.Errorf("ID = %q", user.ID)
	}
	if user.Karma <= 0 {
		t.Errorf("Karma = %d; expected positive", user.Karma)
	}
	t.Logf("pg: karma=%d created=%s", user.Karma, user.GetCreatedTime().Format("2006-01-02"))
}

func TestIntegration_GetUpdates(t *testing.T) {
	c := integrationClient(t)
	updates, err := c.GetUpdates(context.Background())
	if err != nil {
		t.Fatalf("GetUpdates: %v", err)
	}
	t.Logf("updates: %d items, %d profiles", len(updates.Items), len(updates.Profiles))
}

func TestIntegration_Login(t *testing.T) {
	// Just verifies login works and returns a cookie.
	integrationClientAuthed(t)
}

func TestIntegration_FlaggedPost(t *testing.T) {
	c := integrationClientAuthed(t)
	ctx := context.Background()

	maxID, err := c.GetMaxItemID(ctx)
	if err != nil {
		t.Fatalf("GetMaxItemID: %v", err)
	}

	// Scan backwards for a dead item. There's almost always one in the last
	// few hundred IDs.
	const searchRange = 500
	var deadID int64
	for id := maxID; id > maxID-searchRange; id-- {
		item, err := c.GetItem(ctx, id)
		if err != nil {
			continue
		}
		if item.Dead && !item.Deleted {
			deadID = id
			break
		}
	}
	if deadID == 0 {
		t.Skipf("no dead item found in last %d IDs", searchRange)
	}

	// Now fetch with scraping: GetItemWithScraping should recover more data
	// than the bare API response.
	item, err := c.GetItemWithScraping(ctx, deadID)
	if err != nil {
		t.Fatalf("GetItemWithScraping(%d): %v", deadID, err)
	}

	t.Logf("dead item %d: by=%q title=%q url=%q type=%s deleted=%v",
		deadID, item.By, item.Title, item.URL, item.Type, item.Deleted)

	if item.Deleted {
		t.Log("item was deleted between API check and scrape — skip assertions")
		return
	}

	// The URL field must not be an HN-internal link.
	if item.URL != "" {
		for _, bad := range []string{"flag?", "vote?", "hide?", "reply?", "item?", "user?", "from?"} {
			if len(item.URL) >= len(bad) && item.URL[:len(bad)] == bad {
				t.Errorf("URL = %q looks like an HN-internal link, not an external URL", item.URL)
			}
		}
	}

	// We should at minimum know who posted it.
	if item.By == "" {
		t.Errorf("By is empty for dead item %d", deadID)
	}
}

func TestIntegration_TopStoriesWithDetails(t *testing.T) {
	c := integrationClient(t)
	stories, err := c.GetTopStoriesWithDetails(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTopStoriesWithDetails: %v", err)
	}
	if len(stories) != 5 {
		t.Fatalf("expected 5 stories, got %d", len(stories))
	}
	for i, s := range stories {
		if s.Title == "" {
			t.Errorf("story %d (id=%d): empty title", i+1, s.ID)
		}
		t.Logf("%d. [%d pts] %s", i+1, s.Score, s.Title)
	}
}
