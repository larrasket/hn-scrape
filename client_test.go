package hnscrape

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient()
	if c.baseURL != BaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, BaseURL)
	}
	if c.hnURL != HackerNewsURL {
		t.Errorf("hnURL = %q, want %q", c.hnURL, HackerNewsURL)
	}
	if c.httpClient == nil {
		t.Fatal("httpClient should not be nil")
	}
	if c.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", c.httpClient.Timeout, DefaultTimeout)
	}
	if c.httpClient.Jar == nil {
		t.Error("Jar should not be nil")
	}
}

func TestWithBaseURL(t *testing.T) {
	c := NewClient(WithBaseURL("https://custom.example.com"))
	if c.baseURL != "https://custom.example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestWithHackerNewsURL(t *testing.T) {
	c := NewClient(WithHackerNewsURL("https://hn.example.com"))
	if c.hnURL != "https://hn.example.com" {
		t.Errorf("hnURL = %q", c.hnURL)
	}
}

func TestWithTimeout(t *testing.T) {
	c := NewClient(WithTimeout(5 * time.Second))
	if c.httpClient.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", c.httpClient.Timeout)
	}
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 99 * time.Second}
	c := NewClient(WithHTTPClient(custom))
	if c.httpClient != custom {
		t.Error("httpClient should be the custom client")
	}
}

func TestWithUserCookie(t *testing.T) {
	c := NewClient(WithUserCookie("user&hash"))
	if c.userCookie != "user&hash" {
		t.Errorf("userCookie = %q", c.userCookie)
	}
	if c.UserCookie() != "user&hash" {
		t.Errorf("UserCookie() = %q", c.UserCookie())
	}
}

func TestGet_Success(t *testing.T) {
	c, mux := newTestClient(t)
	handleJSON(t, mux, "/v0/test.json", map[string]string{"key": "value"})

	var result map[string]string
	err := c.get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("result[key] = %q, want value", result["key"])
	}
}

func TestGet_HTTPError(t *testing.T) {
	srv, mux := newTestServer(t)
	mux.HandleFunc("/v0/broken.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})
	c := NewClient(WithBaseURL(srv.URL + "/v0"))

	var result any
	err := c.get(context.Background(), "/broken", &result)
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}

func TestGet_InvalidJSON(t *testing.T) {
	srv, mux := newTestServer(t)
	mux.HandleFunc("/v0/bad.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json")) //nolint:errcheck
	})
	c := NewClient(WithBaseURL(srv.URL + "/v0"))

	var result json.RawMessage
	err := c.get(context.Background(), "/bad", &result)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGet_ContextCancel(t *testing.T) {
	// Server that hangs until context is cancelled
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithBaseURL(srv.URL + "/v0"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result any
	err := c.get(ctx, "/slow", &result)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestIsItemDeleted(t *testing.T) {
	if IsItemDeleted(ErrItemDeleted) != true {
		t.Error("IsItemDeleted(ErrItemDeleted) should be true")
	}
	if IsItemDeleted(nil) != false {
		t.Error("IsItemDeleted(nil) should be false")
	}
}

func TestLogin_Success(t *testing.T) {
	srv, mux := newTestServer(t)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "user",
			Value: "testuser&deadbeef",
		})
		http.Redirect(w, r, "/news", http.StatusFound)
	})

	c := NewClient(WithHackerNewsURL(srv.URL))
	if err := c.Login(context.Background(), "testuser", "correctpw"); err != nil {
		t.Fatalf("Login() error: %v", err)
	}
	if c.UserCookie() != "testuser&deadbeef" {
		t.Errorf("UserCookie() = %q, want testuser&deadbeef", c.UserCookie())
	}
}

func TestLogin_Failure(t *testing.T) {
	srv, mux := newTestServer(t)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// No Set-Cookie — simulate bad credentials
		http.Redirect(w, r, "/login?bad=t", http.StatusFound)
	})

	c := NewClient(WithHackerNewsURL(srv.URL))
	err := c.Login(context.Background(), "nobody", "wrongpw")
	if err == nil {
		t.Fatal("Login() should fail when no session cookie is returned")
	}
}
