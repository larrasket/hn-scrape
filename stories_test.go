package hnscrape

import (
	"context"
	"testing"
)

func TestGetTopStories(t *testing.T) {
	c, mux := newTestClient(t)
	want := StoryList{1, 2, 3, 4, 5}
	handleJSON(t, mux, "/v0/topstories.json", want)

	got, err := c.GetTopStories(context.Background())
	if err != nil {
		t.Fatalf("GetTopStories() error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, id := range want {
		if got[i] != id {
			t.Errorf("got[%d] = %d, want %d", i, got[i], id)
		}
	}
}

func TestGetNewStories(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/newstories.json", StoryList{10, 11})

	got, err := c.GetNewStories(context.Background())
	if err != nil {
		t.Fatalf("GetNewStories() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestGetBestStories(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/beststories.json", StoryList{20, 21, 22})

	got, err := c.GetBestStories(context.Background())
	if err != nil {
		t.Fatalf("GetBestStories() error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("len = %d, want 3", len(got))
	}
}

func TestGetAskStories(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/askstories.json", StoryList{30, 31})

	got, err := c.GetAskStories(context.Background())
	if err != nil {
		t.Fatalf("GetAskStories() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestGetShowStories(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/showstories.json", StoryList{40, 41, 42})

	got, err := c.GetShowStories(context.Background())
	if err != nil {
		t.Fatalf("GetShowStories() error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("len = %d, want 3", len(got))
	}
}

func TestGetJobStories(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/jobstories.json", StoryList{50})

	got, err := c.GetJobStories(context.Background())
	if err != nil {
		t.Fatalf("GetJobStories() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestGetTopStoriesWithDetails(t *testing.T) {
	c, mux := newTestClient(t)
	ids := StoryList{1, 2, 3, 4, 5}
	handleJSON(t, mux, "/v0/topstories.json", ids)
	handleItemsJSON(t, mux, map[int64]*Item{
		1: {ID: 1, Title: "A"},
		2: {ID: 2, Title: "B"},
		3: {ID: 3, Title: "C"},
		4: {ID: 4, Title: "D"},
		5: {ID: 5, Title: "E"},
	})

	// Limit to 3
	got, err := c.GetTopStoriesWithDetails(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetTopStoriesWithDetails() error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3 (limit applied)", len(got))
	}
}

func TestGetTopStoriesWithDetails_NoLimit(t *testing.T) {
	c, mux := newTestClient(t)
	ids := StoryList{1, 2}
	handleJSON(t, mux, "/v0/topstories.json", ids)
	handleItemsJSON(t, mux, map[int64]*Item{
		1: {ID: 1, Title: "A"},
		2: {ID: 2, Title: "B"},
	})

	// limit=0 means no limit
	got, err := c.GetTopStoriesWithDetails(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetTopStoriesWithDetails() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestGetBestStoriesWithDetails(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/beststories.json", StoryList{7, 8})
	handleItemsJSON(t, mux, map[int64]*Item{
		7: {ID: 7, Title: "Best 1"},
		8: {ID: 8, Title: "Best 2"},
	})

	got, err := c.GetBestStoriesWithDetails(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetBestStoriesWithDetails() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestGetNewStoriesWithDetails(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/newstories.json", StoryList{9})
	handleItemsJSON(t, mux, map[int64]*Item{
		9: {ID: 9, Title: "New 1"},
	})

	got, err := c.GetNewStoriesWithDetails(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetNewStoriesWithDetails() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}
