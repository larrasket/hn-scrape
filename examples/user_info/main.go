// user_info prints a Hacker News user's profile and their most recent
// submissions.
//
// Usage:
//
//	go run . pg
//	go run . pg 5
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/larrasket/hnscrape"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <username> [limit]", os.Args[0])
	}
	username := os.Args[1]
	limit := 10
	if len(os.Args) >= 3 {
		n, err := strconv.Atoi(os.Args[2])
		if err != nil || n <= 0 {
			log.Fatalf("invalid limit %q", os.Args[2])
		}
		limit = n
	}

	ctx := context.Background()
	client := hnscrape.NewClient(hnscrape.WithTimeout(20 * time.Second))

	user, err := client.GetUser(ctx, username)
	if err != nil {
		log.Fatalf("GetUser: %v", err)
	}

	fmt.Printf("User:    %s\n", user.ID)
	fmt.Printf("Karma:   %d\n", user.Karma)
	fmt.Printf("Since:   %s\n", user.GetCreatedTime().UTC().Format("2006-01-02"))
	if user.About != "" {
		fmt.Printf("About:   %s\n", user.About)
	}
	fmt.Printf("Posts:   %d total\n\n", len(user.Submitted))

	if len(user.Submitted) == 0 {
		fmt.Println("No submissions.")
		return
	}

	items, err := client.GetUserSubmissionsLimited(ctx, username, limit)
	if err != nil {
		log.Fatalf("GetUserSubmissionsLimited: %v", err)
	}

	fmt.Printf("Last %d submissions:\n", len(items))
	for i, item := range items {
		switch {
		case item.Title != "":
			fmt.Printf("  %2d. [%s] %s\n", i+1, item.Type, item.Title)
		case item.Text != "":
			preview := item.Text
			if len(preview) > 80 {
				preview = preview[:80] + "…"
			}
			fmt.Printf("  %2d. [comment] %s\n", i+1, preview)
		default:
			fmt.Printf("  %2d. [%d] (no content)\n", i+1, item.ID)
		}
	}
}
