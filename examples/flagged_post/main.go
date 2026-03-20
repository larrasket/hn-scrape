// flagged_post fetches a dead or flagged Hacker News post that the API returns
// with no content. It logs in to retrieve the session cookie, then scrapes the
// actual page to recover the title, author, and score.
//
// Usage:
//
//	HN_USER=youruser HN_PASS=yourpassword go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/larrasket/hnscrape"
)

func main() {
	username := os.Getenv("HN_USER")
	password := os.Getenv("HN_PASS")
	if username == "" || password == "" {
		log.Fatal("set HN_USER and HN_PASS environment variables")
	}

	// Optional: specify an item ID via HN_ITEM_ID (defaults to a known
	// flagged post for demonstration).
	itemIDStr := os.Getenv("HN_ITEM_ID")
	var itemID int64 = 42 // replace with a real flagged item ID
	if itemIDStr != "" {
		var err error
		itemID, err = strconv.ParseInt(itemIDStr, 10, 64)
		if err != nil {
			log.Fatalf("invalid HN_ITEM_ID: %v", err)
		}
	}

	ctx := context.Background()
	client := hnscrape.NewClient()

	fmt.Printf("Logging in as %s...\n", username)
	if err := client.Login(ctx, username, password); err != nil {
		log.Fatalf("Login: %v", err)
	}
	fmt.Println("Logged in. Cookie obtained.")

	// GetItemWithScraping tries the API first, then falls back to scraping
	// the HTML page — necessary for dead/flagged items whose data is stripped
	// from the API response.
	item, err := client.GetItemWithScraping(ctx, itemID)
	if err != nil {
		log.Fatalf("GetItemWithScraping(%d): %v", itemID, err)
	}

	if item.Deleted {
		fmt.Printf("Item %d has been deleted.\n", itemID)
		return
	}

	fmt.Printf("\nItem %d\n", item.ID)
	fmt.Printf("  Title:  %s\n", item.Title)
	fmt.Printf("  By:     %s\n", item.By)
	fmt.Printf("  Score:  %d\n", item.Score)
	fmt.Printf("  URL:    %s\n", item.URL)
	fmt.Printf("  Dead:   %v\n", item.Dead)
	fmt.Printf("  Time:   %s\n", item.GetCreatedTime().UTC())
	if item.Text != "" {
		fmt.Printf("  Text:   %s\n", item.Text)
	}
}
