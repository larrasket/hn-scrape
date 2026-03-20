package hnscrape

import "time"

// ItemType represents the type of a Hacker News item
type ItemType string

const (
	ItemTypeJob     ItemType = "job"
	ItemTypeStory   ItemType = "story"
	ItemTypeComment ItemType = "comment"
	ItemTypePoll    ItemType = "poll"
	ItemTypePollOpt ItemType = "pollopt"
)

// Item represents a Hacker News item (story, comment, job, Ask HN, poll, etc.)
type Item struct {
	ID          int64    `json:"id"`
	Deleted     bool     `json:"deleted,omitempty"`
	Type        ItemType `json:"type,omitempty"`
	By          string   `json:"by,omitempty"`
	Time        int64    `json:"time,omitempty"`
	Text        string   `json:"text,omitempty"`
	Dead        bool     `json:"dead,omitempty"`
	Parent      int64    `json:"parent,omitempty"`
	Poll        int64    `json:"poll,omitempty"`
	Kids        []int64  `json:"kids,omitempty"`
	URL         string   `json:"url,omitempty"`
	Score       int      `json:"score,omitempty"`
	Title       string   `json:"title,omitempty"`
	Parts       []int64  `json:"parts,omitempty"`
	Descendants int      `json:"descendants,omitempty"`
}

// User represents a Hacker News user
type User struct {
	ID        string  `json:"id"`
	Created   int64   `json:"created,omitempty"`
	Karma     int     `json:"karma,omitempty"`
	About     string  `json:"about,omitempty"`
	Submitted []int64 `json:"submitted,omitempty"`
}

// Updates represents the changed items and profiles
type Updates struct {
	Items    []int64  `json:"items,omitempty"`
	Profiles []string `json:"profiles,omitempty"`
}

// StoryList represents a list of story IDs
type StoryList []int64

// Utility methods for Item

// IsStory checks if an item is a story
func (i *Item) IsStory() bool {
	return i.Type == ItemTypeStory
}

// IsComment checks if an item is a comment
func (i *Item) IsComment() bool {
	return i.Type == ItemTypeComment
}

// IsJob checks if an item is a job
func (i *Item) IsJob() bool {
	return i.Type == ItemTypeJob
}

// IsPoll checks if an item is a poll
func (i *Item) IsPoll() bool {
	return i.Type == ItemTypePoll
}

// IsPollOption checks if an item is a poll option
func (i *Item) IsPollOption() bool {
	return i.Type == ItemTypePollOpt
}

// GetCreatedTime returns the creation time as time.Time
func (i *Item) GetCreatedTime() time.Time {
	return time.Unix(i.Time, 0)
}

// HasKids checks if an item has child comments
func (i *Item) HasKids() bool {
	return len(i.Kids) > 0
}

// KidsCount returns the number of direct child comments
func (i *Item) KidsCount() int {
	return len(i.Kids)
}

// Utility methods for User

// GetCreatedTime returns the user creation time as time.Time
func (u *User) GetCreatedTime() time.Time {
	return time.Unix(u.Created, 0)
}
