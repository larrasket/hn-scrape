package hnscrape

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// newTestServer creates a test HTTP server with a ServeMux, cleaned up
// automatically when the test ends.
func newTestServer(t *testing.T) (*httptest.Server, *http.ServeMux) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, mux
}

// newTestClient returns a Client pointed at a local test server.
func newTestClient(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()
	srv, mux := newTestServer(t)
	c := NewClient(
		WithBaseURL(srv.URL+"/v0"),
		WithHackerNewsURL(srv.URL),
	)
	return c, mux
}

// newTestClientWithCookie returns a Client with a preset session cookie.
func newTestClientWithCookie(t *testing.T, cookie string) (*Client, *http.ServeMux) {
	t.Helper()
	srv, mux := newTestServer(t)
	c := NewClient(
		WithBaseURL(srv.URL+"/v0"),
		WithHackerNewsURL(srv.URL),
		WithUserCookie(cookie),
	)
	return c, mux
}

// handleJSON registers a handler that encodes v as JSON for the given path.
func handleJSON(t *testing.T, mux *http.ServeMux, path string, v any) {
	t.Helper()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(v); err != nil {
			t.Errorf("handleJSON %s: %v", path, err)
		}
	})
}

// handleItemsJSON registers a /v0/item/ handler that dispatches by item ID.
func handleItemsJSON(t *testing.T, mux *http.ServeMux, items map[int64]*Item) {
	t.Helper()
	mux.HandleFunc("/v0/item/", func(w http.ResponseWriter, r *http.Request) {
		seg := strings.TrimPrefix(r.URL.Path, "/v0/item/")
		seg = strings.TrimSuffix(seg, ".json")
		id, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		item, ok := items[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(item); err != nil {
			t.Errorf("handleItemsJSON %d: %v", id, err)
		}
	})
}

// handleHTMLItem registers a /item handler that dispatches by ?id= query param.
func handleHTMLItem(mux *http.ServeMux, pages map[int64]string) {
	mux.HandleFunc("/item", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		body, ok := pages[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(body)) //nolint:errcheck
	})
}
