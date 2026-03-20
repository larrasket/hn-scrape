package hnscrape

import (
	"context"
	"errors"
	"testing"
)

func TestGetItem_Normal(t *testing.T) {
	c, mux := newTestClient(t)
	want := &Item{
		ID:    42,
		Type:  ItemTypeStory,
		By:    "pg",
		Title: "My Startup",
		Score: 200,
		Time:  1705315800,
	}
	handleItemsJSON(t, mux, map[int64]*Item{42: want})

	got, err := c.GetItem(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetItem() error: %v", err)
	}
	if got.ID != want.ID || got.Title != want.Title || got.By != want.By {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGetItem_DeletedByAPI(t *testing.T) {
	c, mux := newTestClient(t)
	handleItemsJSON(t, mux, map[int64]*Item{
		7: {ID: 7, Deleted: true},
	})

	got, err := c.GetItem(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetItem() error: %v", err)
	}
	if !got.Deleted {
		t.Error("expected Deleted=true")
	}
}

func TestGetItem_Dead_WithScraping(t *testing.T) {
	c, mux := newTestClientWithCookie(t, "testuser&hash")

	// API returns item with dead=true and no title
	handleItemsJSON(t, mux, map[int64]*Item{
		55555: {ID: 55555, Dead: true, By: "spammer"},
	})
	handleHTMLItem(mux, map[int64]string{55555: flaggedStoryHTML})

	got, err := c.GetItem(context.Background(), 55555)
	if err != nil {
		t.Fatalf("GetItem() error: %v", err)
	}
	if got.Title != "Flagged Story" {
		t.Errorf("Title = %q, want Flagged Story", got.Title)
	}
}

func TestGetItem_Dead_ScrapeDeleted(t *testing.T) {
	// Dead item that turns out to be deleted when we scrape it
	c, mux := newTestClientWithCookie(t, "testuser&hash")

	handleItemsJSON(t, mux, map[int64]*Item{
		99999: {ID: 99999, Dead: true},
	})
	handleHTMLItem(mux, map[int64]string{99999: deletedItemHTML})

	got, err := c.GetItem(context.Background(), 99999)
	if err != nil {
		t.Fatalf("GetItem() error: %v", err)
	}
	if !got.Deleted {
		t.Error("expected Deleted=true after scrape confirms deletion")
	}
}

func TestGetItem_Dead_NoCookie(t *testing.T) {
	// Dead item without a cookie — should return the API item as-is
	c, mux := newTestClient(t)
	handleItemsJSON(t, mux, map[int64]*Item{
		55555: {ID: 55555, Dead: true},
	})

	got, err := c.GetItem(context.Background(), 55555)
	if err != nil {
		t.Fatalf("GetItem() error: %v", err)
	}
	if !got.Dead {
		t.Error("expected Dead=true")
	}
}

func TestGetItems_Multiple(t *testing.T) {
	c, mux := newTestClient(t)
	items := map[int64]*Item{
		1: {ID: 1, Title: "First"},
		2: {ID: 2, Title: "Second"},
		3: {ID: 3, Title: "Third"},
	}
	handleItemsJSON(t, mux, items)

	got, err := c.GetItems(context.Background(), []int64{1, 2, 3})
	if err != nil {
		t.Fatalf("GetItems() error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}
	// Results must preserve order
	if got[0].ID != 1 || got[1].ID != 2 || got[2].ID != 3 {
		t.Errorf("order not preserved: got IDs %d %d %d", got[0].ID, got[1].ID, got[2].ID)
	}
}

func TestGetItems_Empty(t *testing.T) {
	c, _ := newTestClient(t)
	got, err := c.GetItems(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetItems() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}

func TestGetMaxItemID(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/maxitem.json", int64(9999999))

	id, err := c.GetMaxItemID(context.Background())
	if err != nil {
		t.Fatalf("GetMaxItemID() error: %v", err)
	}
	if id != 9999999 {
		t.Errorf("id = %d, want 9999999", id)
	}
}

func TestGetItemsFromTo_Normal(t *testing.T) {
	c, mux := newTestClient(t)
	handleItemsJSON(t, mux, map[int64]*Item{
		10: {ID: 10}, 11: {ID: 11}, 12: {ID: 12},
	})

	got, err := c.GetItemsFromTo(context.Background(), 10, 12)
	if err != nil {
		t.Fatalf("GetItemsFromTo() error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}
}

func TestGetItemsFromTo_InvalidRange(t *testing.T) {
	c, _ := newTestClient(t)
	_, err := c.GetItemsFromTo(context.Background(), 100, 50)
	if err == nil {
		t.Error("expected error when startID > endID")
	}
}

func TestMergeItemData(t *testing.T) {
	c := NewClient()
	api := &Item{
		ID:    1,
		Type:  ItemTypeStory,
		By:    "apiauthor",
		Score: 5,
		Time:  1000,
	}
	scraped := &Item{
		ID:    1,
		Type:  ItemTypeStory,
		Title: "Scraped Title",
		By:    "scrapedauthor",
		URL:   "https://scraped.example.com",
		Score: 10,
		Time:  2000,
	}
	merged := c.mergeItemData(api, scraped)

	if merged.Title != "Scraped Title" {
		t.Errorf("Title = %q, want scraped value", merged.Title)
	}
	if merged.By != "scrapedauthor" {
		t.Errorf("By = %q, want scraped value", merged.By)
	}
	if merged.Score != 10 {
		t.Errorf("Score = %d, want 10 (higher scraped value)", merged.Score)
	}
	if merged.URL != "https://scraped.example.com" {
		t.Errorf("URL = %q", merged.URL)
	}
}

func TestMergeItemData_PreservesAPIFields(t *testing.T) {
	c := NewClient()
	api := &Item{
		ID:   1,
		By:   "original",
		Kids: []int64{10, 20},
		Type: ItemTypeStory,
	}
	scraped := &Item{ID: 1} // empty scraped result

	merged := c.mergeItemData(api, scraped)
	if merged.By != "original" {
		t.Errorf("By = %q, want original (API value preserved)", merged.By)
	}
	if len(merged.Kids) != 2 {
		t.Errorf("Kids preserved: len = %d, want 2", len(merged.Kids))
	}
}

func TestGetItemWithScraping(t *testing.T) {
	c, mux := newTestClientWithCookie(t, "user&hash")

	handleItemsJSON(t, mux, map[int64]*Item{
		55555: {ID: 55555, Dead: true},
	})
	handleHTMLItem(mux, map[int64]string{55555: flaggedStoryHTML})

	got, err := c.GetItemWithScraping(context.Background(), 55555)
	if err != nil {
		t.Fatalf("GetItemWithScraping() error: %v", err)
	}
	if got.Title != "Flagged Story" {
		t.Errorf("Title = %q, want Flagged Story", got.Title)
	}
}

func TestGetItemWithScraping_NoCookie(t *testing.T) {
	c, _ := newTestClient(t)
	_, err := c.GetItemWithScraping(context.Background(), 1)
	if err == nil {
		t.Error("expected error when no user cookie")
	}
}

func TestGetItemWithScraping_DeletedItem(t *testing.T) {
	c, mux := newTestClientWithCookie(t, "user&hash")
	handleItemsJSON(t, mux, map[int64]*Item{
		99999: {ID: 99999, Dead: true},
	})
	handleHTMLItem(mux, map[int64]string{99999: deletedItemHTML})

	got, err := c.GetItemWithScraping(context.Background(), 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be marked deleted, not return ErrItemDeleted
	if !got.Deleted {
		t.Error("expected Deleted=true")
	}
	if errors.Is(err, ErrItemDeleted) {
		t.Error("should not surface ErrItemDeleted from GetItemWithScraping")
	}
}
