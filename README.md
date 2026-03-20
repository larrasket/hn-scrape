# hnscrape

Go client for the [Hacker News Firebase API](https://github.com/HackerNews/API) with support for reading dead and flagged posts.

When HN kills a post it strips the content from the API response. This library scrapes the HTML page for those posts (with your session cookie) so you can still get at the data.

## Install

```
go get github.com/larrasket/hnscrape
```

## Usage

```go
client := hnscrape.NewClient()
ctx := context.Background()

// Standard API stuff
stories, err := client.GetTopStoriesWithDetails(ctx, 10)
item, err    := client.GetItem(ctx, 12345)
user, err    := client.GetUser(ctx, "pg")
updates, err := client.GetUpdates(ctx)
```

### Flagged / dead posts

Dead items returned by the API have no title, no author, no score. To recover that data you need to be logged in:

```go
client := hnscrape.NewClient()

if err := client.Login(ctx, "youruser", "yourpassword"); err != nil {
    log.Fatal(err)
}

// GetItem automatically scrapes when the API returns dead=true
item, err := client.GetItem(ctx, deadItemID)

// force scraping regardless of the dead flag
item, err = client.GetItemWithScraping(ctx, itemID)
```

If you already have your HN session cookie (the `user` cookie value from your browser), skip the login:

```go
client := hnscrape.NewClient(hnscrape.WithUserCookie("youruser&yourhash"))
```

For `GetItem` to trigger scraping your HN account needs **Show Dead** enabled in your profile settings.
    
### Options

```go
hnscrape.WithTimeout(10 * time.Second)
hnscrape.WithHTTPClient(myClient)
hnscrape.WithUserCookie("user&hash")
```


## Notes

- Concurrent fetches (`GetItems`, `GetUsers`) default to 10 parallel requests. Override with `HNAPI_CONCURRENT_LIMIT`.  
- The HTML scraper is best-effort. HN occasionally changes its markup. If parsing breaks, open an issue with the item ID.  
- The Firebase API is read-only and has no official rate limits, but be reasonable.

