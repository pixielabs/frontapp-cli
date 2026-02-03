package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"golang.org/x/oauth2"
)

type sequenceTokenSource struct {
	mu    sync.Mutex
	count int
}

func (s *sequenceTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count++

	token := "token1"
	if s.count > 1 {
		token = "token2"
	}

	return &oauth2.Token{AccessToken: token}, nil
}

func TestClientRetriesOn401(t *testing.T) {
	var requests int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++

		auth := r.Header.Get("Authorization")
		if auth == "Bearer token1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"me"}`))
	}))
	defer srv.Close()

	ts := &sequenceTokenSource{}
	client := NewClientWithBaseURL(ts, srv.URL)

	var out map[string]any
	if err := client.Get(context.Background(), "/me", &out); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
}
