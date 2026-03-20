package hnscrape

import (
	"context"
	"testing"
)

func TestGetUser(t *testing.T) {
	c, mux := newTestClient(t)
	want := &User{
		ID:      "pg",
		Karma:   155000,
		Created: 1160418111,
		About:   "Essays at paulgraham.com",
	}
	handleJSON(t, mux, "/v0/user/pg.json", want)

	got, err := c.GetUser(context.Background(), "pg")
	if err != nil {
		t.Fatalf("GetUser() error: %v", err)
	}
	if got.ID != "pg" {
		t.Errorf("ID = %q, want pg", got.ID)
	}
	if got.Karma != 155000 {
		t.Errorf("Karma = %d, want 155000", got.Karma)
	}
}

func TestGetUsers_Multiple(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/user/alice.json", &User{ID: "alice", Karma: 100})
	handleJSON(t, mux, "/v0/user/bob.json", &User{ID: "bob", Karma: 200})

	got, err := c.GetUsers(context.Background(), []string{"alice", "bob"})
	if err != nil {
		t.Fatalf("GetUsers() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	// Order must be preserved
	if got[0].ID != "alice" {
		t.Errorf("got[0].ID = %q, want alice", got[0].ID)
	}
	if got[1].ID != "bob" {
		t.Errorf("got[1].ID = %q, want bob", got[1].ID)
	}
}

func TestGetUsers_Empty(t *testing.T) {
	c, _ := newTestClient(t)
	got, err := c.GetUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetUsers() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestGetUserSubmissions(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/user/tester.json", &User{
		ID:        "tester",
		Submitted: []int64{1, 2, 3},
	})
	handleItemsJSON(t, mux, map[int64]*Item{
		1: {ID: 1, Title: "Post One"},
		2: {ID: 2, Title: "Post Two"},
		3: {ID: 3, Title: "Post Three"},
	})

	got, err := c.GetUserSubmissions(context.Background(), "tester")
	if err != nil {
		t.Fatalf("GetUserSubmissions() error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
}

func TestGetUserSubmissions_NoSubmissions(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/user/newbie.json", &User{ID: "newbie"})

	got, err := c.GetUserSubmissions(context.Background(), "newbie")
	if err != nil {
		t.Fatalf("GetUserSubmissions() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestGetUserSubmissionsLimited(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/user/prolific.json", &User{
		ID:        "prolific",
		Submitted: []int64{1, 2, 3, 4, 5},
	})
	handleItemsJSON(t, mux, map[int64]*Item{
		1: {ID: 1}, 2: {ID: 2}, 3: {ID: 3},
	})

	got, err := c.GetUserSubmissionsLimited(context.Background(), "prolific", 3)
	if err != nil {
		t.Fatalf("GetUserSubmissionsLimited() error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3 (limit applied)", len(got))
	}
}

func TestGetUserSubmissionsLimited_ZeroLimit(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/user/someone.json", &User{
		ID:        "someone",
		Submitted: []int64{1, 2},
	})
	handleItemsJSON(t, mux, map[int64]*Item{
		1: {ID: 1}, 2: {ID: 2},
	})

	// limit=0 means return all
	got, err := c.GetUserSubmissionsLimited(context.Background(), "someone", 0)
	if err != nil {
		t.Fatalf("GetUserSubmissionsLimited() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}
