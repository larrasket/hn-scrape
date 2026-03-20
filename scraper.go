package hnscrape

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	HackerNewsURL = "https://news.ycombinator.com"
)

// scrapeDeadItem fetches a dead item by scraping the HTML from Hacker News
func (c *Client) scrapeDeadItem(ctx context.Context, id int64) (*Item, error) {
	if c.userCookie == "" {
		return nil, fmt.Errorf("user cookie required to access dead items")
	}

	itemURL := fmt.Sprintf("%s/item?id=%d", c.hnURL, id)

	req, err := http.NewRequestWithContext(ctx, "GET", itemURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for scraping: %w", err)
	}

	if err := c.addUserCookie(req); err != nil {
		return nil, fmt.Errorf("failed to add user cookie: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform scraping request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.
			Errorf("scraping request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	item, err := c.parseItemFromHTML(string(body), id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse item from HTML: %w", err)
	}

	if item.Deleted {
		return nil, fmt.Errorf("item %d: %w", id, ErrItemDeleted)
	}

	return item, nil
}

// addUserCookie adds the user session cookie to the request.
func (c *Client) addUserCookie(req *http.Request) error {
	if c.userCookie == "" {
		return fmt.Errorf("no user cookie provided")
	}
	req.AddCookie(&http.Cookie{Name: "user", Value: c.userCookie})
	return nil
}

// parseItemFromHTML extracts item data from HTML content
func (c *Client) parseItemFromHTML(htmlContent string,
	itemID int64) (*Item, error) {
	item := &Item{ID: itemID}

	itemHTML := c.extractItemHTML(htmlContent, itemID)
	if itemHTML == "" {
		itemHTML = htmlContent
	}

	doc, err := html.Parse(strings.NewReader(itemHTML))
	if err == nil {
		c.parseHTMLNode(doc, item)
	}

	if item.Title == "" || item.By == "" || item.Time == 0 ||
		(item.Text == "" && item.Type != ItemTypeStory) {
		// Use the original full HTML for regex fallback — the extracted
		// item HTML only covers the matching element, not the surrounding
		// subtext rows that hold author, score, and timestamp.
		c.parseWithRegex(htmlContent, item)
	}

	return item, nil
}

// extractItemHTML extracts the HTML section relevant to a specific item ID
func (c *Client) extractItemHTML(htmlContent string, itemID int64) string {
	patterns := []string{
		fmt.Sprintf(`(?s)<tr[^>]*id=['"]%d['"][^>]*>.*?</tr>`, itemID),
		fmt.Sprintf(`(?s)<tr[^>]*class="athing[^"]*"[^>]*id=['"]%d['"][^>]*>.*?</table>`, itemID),
		fmt.Sprintf(`(?s)id=['"]%d['"].*?(?:</tr>|</table>|</div>)`, itemID),
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		if match := regex.FindString(htmlContent); match != "" {
			return match
		}
	}

	return ""
}

// parseHTMLNode recursively parses HTML nodes to extract item data
func (c *Client) parseHTMLNode(node *html.Node, item *Item) {
	if node.Type == html.ElementNode {
		switch node.Data {
		case "title":
			if title := c.getTextContent(node); title != "" {
				titleText := strings.TrimSpace(title)
				titleText = strings.TrimSuffix(titleText, " | Hacker News")
				if item.Title == "" && titleText != "" && titleText != "Hacker News" {
					item.Title = titleText
				}
			}
		case "span":
			class := c.getAttr(node, "class")
			switch class {
			case "titleline":
				if item.Title == "" {
					title := c.extractTitleFromTitleline(node)
					if title != "" {
						item.Title = strings.TrimSpace(title)
					}
				}
			case "hnuser":
				if item.By == "" {
					item.By = strings.TrimSpace(c.getTextContent(node))
				}
			case "age":
				if timeStr := c.getAttr(node, "title"); timeStr != "" {
					if parsedTime, err := time.Parse("2006-01-02T15:04:05",
						timeStr); err == nil {
						item.Time = parsedTime.Unix()
					}
				}
			case "score":
				if scoreText := c.getTextContent(node); scoreText != "" {
					if score, err := strconv.Atoi(strings.
						Fields(scoreText)[0]); err == nil {
						item.Score = score
					}
				}
			case "commtext", "comment-tree":
				if item.Text == "" {
					htmlText := c.getInnerHTML(node)
					textContent := strings.TrimSpace(htmlText)

					// Check for deleted content
					if c.isDeletedText(textContent) {
						item.Deleted = true
						return
					}

					item.Text = textContent
					if item.Text != "" && item.Type == "" {
						item.Type = ItemTypeComment
					}
				}
			}
		case "div":
			class := c.getAttr(node, "class")
			if strings.Contains(class, "commtext") || class == "comment" {
				if item.Text == "" {
					htmlText := c.getInnerHTML(node)
					textContent := strings.TrimSpace(htmlText)

					// Check for deleted content
					if c.isDeletedText(textContent) {
						item.Deleted = true
						return
					}

					item.Text = textContent
					if item.Text != "" && item.Type == "" {
						item.Type = ItemTypeComment
					}
				}
			}
		case "a":
			href := c.getAttr(node, "href")
			if strings.Contains(href, "item?id=") && item.Parent == 0 {
				if parentID := c.extractIDFromURL(href); parentID > 0 &&
					parentID != item.ID {
					item.Parent = parentID
				}
			}
			// Only accept absolute HTTP/HTTPS URLs as the item URL.
			// All HN-internal links (item?, user?, flag?, vote?, hide?,
			// reply?, from?, etc.) are relative and must be skipped.
			if (strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")) &&
				!strings.Contains(href, "ycombinator.com") {
				if item.URL == "" {
					item.URL = href
					if item.Type == "" {
						item.Type = ItemTypeStory
					}
				}
			}
		}
	}

	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		if c.isDeletedText(text) {
			item.Deleted = true
			return
		}
	}

	if !item.Deleted {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			c.parseHTMLNode(child, item)
			if item.Deleted {
				return
			}
		}
	}
}

// extractTitleFromTitleline extracts the title from a titleline span element
func (c *Client) extractTitleFromTitleline(node *html.Node) string {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "a" {
			href := c.getAttr(child, "href")
			if href != "" && !strings.HasPrefix(href, "user?") && !strings.HasPrefix(href, "item?") {
				title := c.getTextContent(child)
				return strings.TrimSpace(title)
			}
		}
		if childTitle := c.extractTitleFromTitleline(child); childTitle != "" {
			return childTitle
		}
	}
	return ""
}

// parseWithRegex uses regex to extract data as a fallback method
func (c *Client) parseWithRegex(htmlContent string, item *Item) {
	itemIdStr := strconv.FormatInt(item.ID, 10)

	isComment := strings.Contains(htmlContent, `class="comtr"`) ||
		(strings.Contains(htmlContent, `class="athing"`) &&
			!strings.Contains(htmlContent, `class="submission"`)) ||
		strings.Contains(htmlContent, `class="commtext"`) ||
		strings.Contains(htmlContent, `>parent</a>`)

	if isComment {
		item.Type = ItemTypeComment
		c.parseComment(htmlContent, item, itemIdStr)
	} else {
		item.Type = ItemTypeStory
		c.parseStory(htmlContent, item, itemIdStr)
	}
}

// parseComment extracts comment-specific data
func (c *Client) parseComment(htmlContent string, item *Item, itemIdStr string) {
	if c.isDeletedContent(htmlContent) {
		item.Deleted = true
		return
	}

	if item.Text == "" {
		textRegex := regexp.MustCompile(
			`<div[^>]*class="commtext[^"]*"[^>]*>(.*?)</div>`)
		if matches := textRegex.
			FindStringSubmatch(htmlContent); len(matches) > 1 {
			textContent := strings.TrimSpace(matches[1])

			if c.isDeletedText(textContent) {
				item.Deleted = true
				return
			}

			item.Text = textContent
		}
	}

	if item.By == "" {
		authorRegex := regexp.
			MustCompile(`<a[^>]*class="hnuser"[^>]*>([^<]+)</a>`)
		if matches := authorRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			item.By = strings.TrimSpace(matches[1])
		}
	}

	if item.Score == 0 {
		scorePattern := fmt.Sprintf(`<span[^>]*class="score"[^>]*id="score_%s"[^>]*>(\d+)\s+points?</span>`, itemIdStr)
		scoreRegex := regexp.MustCompile(scorePattern)
		if matches := scoreRegex.
			FindStringSubmatch(htmlContent); len(matches) > 1 {
			if score, err := strconv.Atoi(matches[1]); err == nil {
				item.Score = score
			}
		}
	}

	if item.Time == 0 {
		timeRegex := regexp.
			MustCompile(`<span[^>]*class="age"[^>]*title="([^"]+)"`)
		if matches := timeRegex.
			FindStringSubmatch(htmlContent); len(matches) > 1 {
			if parsedTime, err := time.
				Parse("2006-01-02T15:04:05", matches[1]); err == nil {
				item.Time = parsedTime.Unix()
			}
		}
	}

	if item.Parent == 0 {
		parentRegex := regexp.
			MustCompile(`<a[^>]+href="item\?id=(\d+)"[^>]*>parent</a>`)
		if matches := parentRegex.
			FindStringSubmatch(htmlContent); len(matches) > 1 {
			if parentID, err := strconv.
				ParseInt(matches[1], 10, 64); err == nil &&
				parentID != item.ID {
				item.Parent = parentID
			}
		}
	}
}

// isDeletedContent checks if the HTML content indicates a deleted item
func (c *Client) isDeletedContent(htmlContent string) bool {
	deletedPatterns := []string{
		`<div[^>]*class="comment"[^>]*>\s*\[deleted\]\s*<`,
		`<div[^>]*class="commtext"[^>]*>\s*\[deleted\]\s*<`,
		`>\s*\[deleted\]\s*</div>`,
	}

	for _, pattern := range deletedPatterns {
		if matched, _ := regexp.MatchString(pattern, htmlContent); matched {
			return true
		}
	}

	return false
}

// isDeletedText checks if the extracted text content is deleted
func (c *Client) isDeletedText(text string) bool {
	cleanText := strings.TrimSpace(text)
	return cleanText == "[deleted]" || cleanText == "<p>[deleted]</p>"
}

// parseStory extracts story-specific data
func (c *Client) parseStory(htmlContent string, item *Item, itemIdStr string) {
	if item.Title == "" {
		titlelineRegex := regexp.MustCompile(`<span class="titleline"[^>]*>.*?<a[^>]+href="[^"]*"[^>]*>([^<]+)</a>`)
		if matches := titlelineRegex.
			FindStringSubmatch(htmlContent); len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			title = regexp.MustCompile(`^\[flagged\]\s*`).
				ReplaceAllString(title, "")
			title = regexp.MustCompile(`^\[dead\]\s*`).
				ReplaceAllString(title, "")
			if title != "" {
				item.Title = title
			}
		}

		if item.Title == "" {
			titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
			if matches := titleRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
				title := strings.TrimSpace(matches[1])
				title = strings.TrimSuffix(title, " | Hacker News")

				if !strings.Contains(title, "Hacker News") && title != "" {
					item.Title = title
				}
			}
		}
	}

	if item.By == "" {
		authorRegex := regexp.MustCompile(`<a[^>]*class="hnuser"[^>]*>([^<]+)</a>`)
		if matches := authorRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			item.By = strings.TrimSpace(matches[1])
		}
	}

	if item.Score == 0 {
		scorePattern := fmt.
			Sprintf(`<span[^>]*class="score"[^>]*id="score_%s"[^>]*>(\d+)\s+points?</span>`, itemIdStr)
		scoreRegex := regexp.MustCompile(scorePattern)
		if matches := scoreRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			if score, err := strconv.Atoi(matches[1]); err == nil {
				item.Score = score
			}
		}
	}

	if item.Time == 0 {
		timeRegex := regexp.MustCompile(`<span[^>]*class="age"[^>]*title="([^"]+)"`)
		if matches := timeRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			if parsedTime, err := time.Parse("2006-01-02T15:04:05", matches[1]); err == nil {
				item.Time = parsedTime.Unix()
			}
		}
	}

	if item.URL == "" {
		urlRegex := regexp.MustCompile(`<span class="titleline"[^>]*>.*?<a[^>]+href="([^"]+)"[^>]*rel="nofollow"`)
		if matches := urlRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			item.URL = matches[1]
		}
	}
}

// getAttr gets an attribute value from an HTML node
func (c *Client) getAttr(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// getInnerHTML extracts the innerHTML of a node, preserving HTML tags
func (c *Client) getInnerHTML(node *html.Node) string {
	var buf strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		html.Render(&buf, child)
	}
	return buf.String()
}

// getTextContent extracts all text content from a node and its children
func (c *Client) getTextContent(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}

	var text strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text.WriteString(c.getTextContent(child))
	}
	return text.String()
}

// extractIDFromURL extracts an ID number from a URL like "item?id=12345".
// It requires "id=" to be preceded by "?" or "&" so that "notid=123" is
// not mistakenly matched.
func (c *Client) extractIDFromURL(url string) int64 {
	re := regexp.MustCompile(`(?:^|[?&])id=(\d+)`)
	if matches := re.FindStringSubmatch(url); len(matches) > 1 {
		if id, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			return id
		}
	}
	return 0
}

// ParseItemFromHTML extracts item data from HTML content (exported for testing)
func (c *Client) ParseItemFromHTML(htmlContent string, itemID int64) (*Item, error) {
	return c.parseItemFromHTML(htmlContent, itemID)
}
