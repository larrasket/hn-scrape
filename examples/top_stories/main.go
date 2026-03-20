// top_stories prints the current top 10 stories from Hacker News with their
// scores and comment counts.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/larrasket/hnscrape"
)

func main() {
	client := hnscrape.NewClient(
		hnscrape.WithTimeout(15 * time.Second),
	)
	ctx := context.Background()

	stories, err := client.GetTopStoriesWithDetails(ctx, 10)
	if err != nil {
		log.Fatalf("GetTopStoriesWithDetails: %v", err)
	}

	for i, s := range stories {
		fmt.Printf("%2d. [%d pts | %d comments] %s\n",
			i+1, s.Score, s.Descendants, s.Title)
		if s.URL != "" {
			fmt.Printf("    %s\n", s.URL)
		}
		fmt.Println()
	}
}
