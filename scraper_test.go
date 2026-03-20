package hnscrape

import (
	"context"
	"net/http"
	"testing"
)

// Minimal but realistic HN HTML snippets used across scraper tests.
// Kept close to what the real site serves so bugs in parsing real pages
// are caught here before they reach users.

const storyHTML = `<!DOCTYPE html>
<html><head><title>My Story Title | Hacker News</title></head>
<body><table id="hnmain"><tr><td>
<table class="fatitem">
<tr class="athing" id="12345">
  <td class="title">
    <span class="titleline">
      <a href="https://example.com/article" rel="nofollow">My Story Title</a>
      <span class="sitebit comhead">(<a href="from?site=example.com">example.com</a>)</span>
    </span>
  </td>
</tr>
<tr>
  <td class="subtext">
    <span class="score" id="score_12345">100 points</span>
    by <a href="user?id=testauthor" class="hnuser">testauthor</a>
    <span class="age" title="2024-01-15T10:30:00">3 hours ago</span>
    | <a href="item?id=12345">42 comments</a>
  </td>
</tr>
</table>
</td></tr></table></body></html>`

const commentHTML = `<!DOCTYPE html>
<html><head><title>Comment | Hacker News</title></head>
<body><table id="hnmain"><tr><td>
<table class="fatitem">
<tr class="athing comtr" id="67890">
  <td>
    <div>
      <span class="comhead">
        <a href="user?id=commenter" class="hnuser">commenter</a>
        <span class="age" title="2024-01-15T11:00:00">2 hours ago</span>
        | <a href="item?id=12345">parent</a>
      </span>
    </div>
    <div class="comment">
      <span class="commtext c00">This is a test comment.</span>
    </div>
  </td>
</tr>
</table>
</td></tr></table></body></html>`

const deletedItemHTML = `<!DOCTYPE html>
<html><head><title>Hacker News</title></head><body>
<table id="hnmain"><tr><td>
<table class="fatitem">
<tr class="athing comtr" id="99999">
  <td>
    <div class="comment">
      <span class="commtext c00">[deleted]</span>
    </div>
  </td>
</tr>
</table></td></tr></table></body></html>`

const flaggedStoryHTML = `<!DOCTYPE html>
<html><head><title>Flagged Story | Hacker News</title></head>
<body><table id="hnmain"><tr><td>
<table class="fatitem">
<tr class="athing" id="55555">
  <td class="title">
    <span class="titleline">
      <a href="https://flagged.example.com/" rel="nofollow">Flagged Story</a>
    </span>
  </td>
</tr>
<tr>
  <td class="subtext">
    <span class="score" id="score_55555">3 points</span>
    by <a href="user?id=spammer" class="hnuser">spammer</a>
    <span class="age" title="2024-01-15T09:00:00">5 hours ago</span>
  </td>
</tr>
</table>
</td></tr></table></body></html>`

// deadCommentHTML simulates a dead comment page as HN actually renders it —
// including flag/vote/hide links that must NOT be mistaken for the item URL.
const deadCommentHTML = `<!DOCTYPE html>
<html><head><title>Hacker News</title></head>
<body><table id="hnmain"><tr><td>
<table class="fatitem">
<tr class="athing comtr" id="47449485">
  <td>
    <div>
      <span class="comhead">
        <a href="user?id=justboy1987" class="hnuser">justboy1987</a>
        <span class="age" title="2026-03-20T01:00:00">1 hour ago</span>
        | <a href="item?id=12345">parent</a>
        | <a href="flag?id=47449485&auth=abc&goto=item%3Fid%3D47449485">flag</a>
        | <a href="hide?id=47449485&auth=abc&goto=item%3Fid%3D47449485">hide</a>
        | <a href="vote?id=47449485&how=up&auth=abc">upvote</a>
      </span>
    </div>
    <div class="comment">
      <span class="commtext c00">This is the actual comment content.</span>
    </div>
  </td>
</tr>
</table>
</td></tr></table></body></html>`


func TestParseItemFromHTML_Story(t *testing.T) {
	c := NewClient()
	item, err := c.ParseItemFromHTML(storyHTML, 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Title != "My Story Title" {
		t.Errorf("Title = %q, want %q", item.Title, "My Story Title")
	}
	if item.By != "testauthor" {
		t.Errorf("By = %q, want %q", item.By, "testauthor")
	}
	if item.Score != 100 {
		t.Errorf("Score = %d, want 100", item.Score)
	}
	if item.URL != "https://example.com/article" {
		t.Errorf("URL = %q, want https://example.com/article", item.URL)
	}
	if item.Time == 0 {
		t.Error("Time should not be zero")
	}
	if item.Deleted {
		t.Error("Deleted should be false")
	}
}

func TestParseItemFromHTML_Comment(t *testing.T) {
	c := NewClient()
	item, err := c.ParseItemFromHTML(commentHTML, 67890)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.By != "commenter" {
		t.Errorf("By = %q, want commenter", item.By)
	}
	if item.Parent != 12345 {
		t.Errorf("Parent = %d, want 12345", item.Parent)
	}
	if item.Time == 0 {
		t.Error("Time should not be zero")
	}
	if item.Type != ItemTypeComment {
		t.Errorf("Type = %q, want %q", item.Type, ItemTypeComment)
	}
	if item.Deleted {
		t.Error("Deleted should be false for a live comment")
	}
}

func TestParseItemFromHTML_DeletedItem(t *testing.T) {
	c := NewClient()
	item, err := c.ParseItemFromHTML(deletedItemHTML, 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !item.Deleted {
		t.Error("Deleted should be true for [deleted] content")
	}
}

func TestParseItemFromHTML_FlaggedStory(t *testing.T) {
	c := NewClient()
	item, err := c.ParseItemFromHTML(flaggedStoryHTML, 55555)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Title != "Flagged Story" {
		t.Errorf("Title = %q, want %q", item.Title, "Flagged Story")
	}
	if item.By != "spammer" {
		t.Errorf("By = %q, want spammer", item.By)
	}
	if item.URL != "https://flagged.example.com/" {
		t.Errorf("URL = %q", item.URL)
	}
	if item.Deleted {
		t.Error("Deleted should be false for a flagged but non-deleted story")
	}
}

func TestIsDeletedText(t *testing.T) {
	c := NewClient()
	tests := []struct {
		input string
		want  bool
	}{
		{"[deleted]", true},
		{"  [deleted]  ", true},
		{"<p>[deleted]</p>", true},
		{"some regular text", false},
		{"", false},
		{"[flagged]", false},
		{"deleted", false},
		{"[deleted] extra", false},
	}
	for _, tt := range tests {
		if got := c.isDeletedText(tt.input); got != tt.want {
			t.Errorf("isDeletedText(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestExtractIDFromURL(t *testing.T) {
	c := NewClient()
	tests := []struct {
		url  string
		want int64
	}{
		{"item?id=12345", 12345},
		{"https://news.ycombinator.com/item?id=99999", 99999},
		{"item?id=1", 1},
		{"item?notid=12345", 0},
		{"", 0},
		{"item?id=abc", 0},
	}
	for _, tt := range tests {
		if got := c.extractIDFromURL(tt.url); got != tt.want {
			t.Errorf("extractIDFromURL(%q) = %d, want %d", tt.url, got, tt.want)
		}
	}
}

func TestParseItemFromHTML_DeadComment(t *testing.T) {
	c := NewClient()
	item, err := c.ParseItemFromHTML(deadCommentHTML, 47449485)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.URL != "" {
		t.Errorf("URL = %q — flag/vote/hide links must not be treated as item URL", item.URL)
	}
	if item.By != "justboy1987" {
		t.Errorf("By = %q, want justboy1987", item.By)
	}
	if item.Parent != 12345 {
		t.Errorf("Parent = %d, want 12345", item.Parent)
	}
	if item.Deleted {
		t.Error("Deleted should be false")
	}
}

func TestAddUserCookie_Set(t *testing.T) {
	c := NewClient(WithUserCookie("testuser&abc123hash"))
	req, _ := http.NewRequest("GET", "https://news.ycombinator.com/", nil)

	if err := c.addUserCookie(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cookie, err := req.Cookie("user")
	if err != nil {
		t.Fatal("user cookie not found on request")
	}
	if cookie.Value != "testuser&abc123hash" {
		t.Errorf("cookie.Value = %q, want testuser&abc123hash", cookie.Value)
	}
}

func TestAddUserCookie_NoCookie(t *testing.T) {
	c := NewClient()
	req, _ := http.NewRequest("GET", "https://news.ycombinator.com/", nil)
	if err := c.addUserCookie(req); err == nil {
		t.Error("expected error when client has no cookie")
	}
}

func TestScrapeDeadItem(t *testing.T) {
	c, mux := newTestClientWithCookie(t, "testuser&testhash")
	handleHTMLItem(mux, map[int64]string{55555: flaggedStoryHTML})

	item, err := c.scrapeDeadItem(context.Background(), 55555)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Title != "Flagged Story" {
		t.Errorf("Title = %q, want Flagged Story", item.Title)
	}
	if item.By != "spammer" {
		t.Errorf("By = %q, want spammer", item.By)
	}
}

func TestScrapeDeadItem_NoCookie(t *testing.T) {
	c, _ := newTestClient(t)
	_, err := c.scrapeDeadItem(context.Background(), 12345)
	if err == nil {
		t.Error("expected error when client has no user cookie")
	}
}

func TestScrapeDeadItem_DeletedItem(t *testing.T) {
	c, mux := newTestClientWithCookie(t, "testuser&testhash")
	handleHTMLItem(mux, map[int64]string{99999: deletedItemHTML})

	_, err := c.scrapeDeadItem(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected ErrItemDeleted error")
	}
	if !IsItemDeleted(err) {
		t.Errorf("expected ErrItemDeleted, got: %v", err)
	}
}

func TestScrapeDeadItem_NotFound(t *testing.T) {
	c, mux := newTestClientWithCookie(t, "testuser&testhash")
	// Register a handler that always 404s
	mux.HandleFunc("/item", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	_, err := c.scrapeDeadItem(context.Background(), 12345)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}
