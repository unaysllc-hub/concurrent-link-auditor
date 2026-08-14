package auditor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckPreservesInputOrderAndStatuses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	urls := []string{server.URL + "/ok", server.URL + "/missing"}
	results := Check(context.Background(), server.Client(), urls, 4)
	if len(results) != 2 || results[0].URL != urls[0] || results[1].URL != urls[1] {
		t.Fatalf("unexpected result order: %#v", results)
	}
	if results[0].StatusCode != 204 || !results[0].Healthy() {
		t.Fatalf("expected healthy 204 result: %#v", results[0])
	}
	if results[1].StatusCode != 404 || results[1].Healthy() {
		t.Fatalf("expected unhealthy 404 result: %#v", results[1])
	}
}

func TestInvalidURLDoesNotSendRequest(t *testing.T) {
	results := Check(context.Background(), &http.Client{}, []string{"not-a-url"}, 1)
	if results[0].Error == "" {
		t.Fatal("expected invalid URL error")
	}
}

func TestContextTimeoutIsReported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	results := Check(ctx, server.Client(), []string{server.URL}, 1)
	if results[0].Error == "" {
		t.Fatal("expected timeout error")
	}
}
