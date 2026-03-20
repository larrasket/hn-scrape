package hnscrape

import (
	"testing"
	"time"
)

func TestItem_TypeChecks(t *testing.T) {
	tests := []struct {
		itemType ItemType
		isStory  bool
		isCmt    bool
		isJob    bool
		isPoll   bool
		isPollOp bool
	}{
		{ItemTypeStory, true, false, false, false, false},
		{ItemTypeComment, false, true, false, false, false},
		{ItemTypeJob, false, false, true, false, false},
		{ItemTypePoll, false, false, false, true, false},
		{ItemTypePollOpt, false, false, false, false, true},
	}
	for _, tt := range tests {
		item := &Item{Type: tt.itemType}
		if got := item.IsStory(); got != tt.isStory {
			t.Errorf("%s IsStory() = %v, want %v", tt.itemType, got, tt.isStory)
		}
		if got := item.IsComment(); got != tt.isCmt {
			t.Errorf("%s IsComment() = %v, want %v", tt.itemType, got, tt.isCmt)
		}
		if got := item.IsJob(); got != tt.isJob {
			t.Errorf("%s IsJob() = %v, want %v", tt.itemType, got, tt.isJob)
		}
		if got := item.IsPoll(); got != tt.isPoll {
			t.Errorf("%s IsPoll() = %v, want %v", tt.itemType, got, tt.isPoll)
		}
		if got := item.IsPollOption(); got != tt.isPollOp {
			t.Errorf("%s IsPollOption() = %v, want %v", tt.itemType, got, tt.isPollOp)
		}
	}
}

func TestItem_GetCreatedTime(t *testing.T) {
	ts := int64(1705315800) // 2024-01-15 10:30:00 UTC
	item := &Item{Time: ts}
	got := item.GetCreatedTime()
	want := time.Unix(ts, 0)
	if !got.Equal(want) {
		t.Errorf("GetCreatedTime() = %v, want %v", got, want)
	}
}

func TestItem_KidsMethods(t *testing.T) {
	noKids := &Item{}
	if noKids.HasKids() {
		t.Error("HasKids() should be false when Kids is nil")
	}
	if noKids.KidsCount() != 0 {
		t.Errorf("KidsCount() = %d, want 0", noKids.KidsCount())
	}

	withKids := &Item{Kids: []int64{1, 2, 3}}
	if !withKids.HasKids() {
		t.Error("HasKids() should be true when Kids is non-empty")
	}
	if withKids.KidsCount() != 3 {
		t.Errorf("KidsCount() = %d, want 3", withKids.KidsCount())
	}
}

func TestUser_GetCreatedTime(t *testing.T) {
	ts := int64(1705315800)
	u := &User{Created: ts}
	got := u.GetCreatedTime()
	want := time.Unix(ts, 0)
	if !got.Equal(want) {
		t.Errorf("GetCreatedTime() = %v, want %v", got, want)
	}
}
