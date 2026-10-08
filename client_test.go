package dogabot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetMeSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/me" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer dbk_live_test" {
			t.Fatalf("auth %q", got)
		}
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "dogabot-sdk-go/") {
			t.Fatalf("ua %q", r.Header.Get("User-Agent"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"user_id": "u1"}})
	}))
	defer srv.Close()

	c := New("dbk_live_test", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	out, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	m := out.(map[string]any)
	data := m["data"].(map[string]any)
	if data["user_id"] != "u1" {
		t.Fatalf("%v", out)
	}
}

func TestWriteRequiresIdempotency(t *testing.T) {
	c := New("dbk_live_test", WithBaseURL("http://example.invalid"))
	_, err := c.PostTerminalPlaceorder(context.Background(), map[string]any{"symbol": "BTC"})
	if err == nil || !strings.Contains(err.Error(), "Idempotency-Key") {
		t.Fatalf("got %v", err)
	}
}

func TestWriteSendsIdempotency(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Idempotency-Key") != "k1" {
			t.Fatalf("idem %q", r.Header.Get("Idempotency-Key"))
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "BTC") {
			t.Fatalf("body %s", body)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer srv.Close()

	c := New("dbk_live_test", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.PostTerminalPlaceorder(context.Background(), map[string]any{"symbol": "BTC"}, WithIdempotencyKey("k1"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	c := New("dbk_live_test", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithMaxRetries(0))
	_, err := c.GetMe(context.Background())
	ae, ok := err.(*APIError)
	if !ok || ae.StatusCode != 401 || ae.Message != "unauthorized" {
		t.Fatalf("%v", err)
	}
}

func TestOperationIDsCoverage(t *testing.T) {
	if len(OperationIDs) < 100 {
		t.Fatalf("expected full allowlist, got %d", len(OperationIDs))
	}
}
