package hnscrape

import (
	"context"
	"testing"
)

func TestGetUpdates(t *testing.T) {
	c, mux := newTestClient(t)
	want := &Updates{
		Items:    []int64{1, 2, 3},
		Profiles: []string{"alice", "bob"},
	}
	handleJSON(t, mux, "/v0/updates.json", want)

	got, err := c.GetUpdates(context.Background())
	if err != nil {
		t.Fatalf("GetUpdates() error: %v", err)
	}
	if len(got.Items) != 3 {
		t.Errorf("Items len = %d, want 3", len(got.Items))
	}
	if len(got.Profiles) != 2 {
		t.Errorf("Profiles len = %d, want 2", len(got.Profiles))
	}
}

func TestGetChangedItems(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/updates.json", &Updates{
		Items:    []int64{10, 20},
		Profiles: []string{"x"},
	})

	got, err := c.GetChangedItems(context.Background())
	if err != nil {
		t.Fatalf("GetChangedItems() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
	if got[0] != 10 || got[1] != 20 {
		t.Errorf("got %v, want [10, 20]", got)
	}
}

func TestGetChangedProfiles(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/updates.json", &Updates{
		Items:    []int64{1},
		Profiles: []string{"carol", "dave"},
	})

	got, err := c.GetChangedProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetChangedProfiles() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
	if got[0] != "carol" {
		t.Errorf("got[0] = %q, want carol", got[0])
	}
}

func TestGetChangedItemsWithDetails(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/updates.json", &Updates{
		Items: []int64{1, 2},
	})
	handleItemsJSON(t, mux, map[int64]*Item{
		1: {ID: 1, Title: "Updated Story"},
		2: {ID: 2, Title: "Another Update"},
	})

	got, err := c.GetChangedItemsWithDetails(context.Background())
	if err != nil {
		t.Fatalf("GetChangedItemsWithDetails() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestGetChangedItemsWithDetails_Empty(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/updates.json", &Updates{})

	got, err := c.GetChangedItemsWithDetails(context.Background())
	if err != nil {
		t.Fatalf("GetChangedItemsWithDetails() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d items", len(got))
	}
}

func TestGetChangedProfilesWithDetails(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/updates.json", &Updates{
		Profiles: []string{"alice"},
	})
	handleJSON(t, mux, "/v0/user/alice.json", &User{ID: "alice", Karma: 42})

	got, err := c.GetChangedProfilesWithDetails(context.Background())
	if err != nil {
		t.Fatalf("GetChangedProfilesWithDetails() error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != "alice" {
		t.Errorf("got[0].ID = %q, want alice", got[0].ID)
	}
}

func TestGetChangedProfilesWithDetails_Empty(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/updates.json", &Updates{})

	got, err := c.GetChangedProfilesWithDetails(context.Background())
	if err != nil {
		t.Fatalf("GetChangedProfilesWithDetails() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}
