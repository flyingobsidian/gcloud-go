package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
)

// testTokens returns a TokenProvider whose only account,
// user@example.com, gets tokens from a fake endpoint, and a counter of the
// refreshes made against it.
func testTokens(t *testing.T) (*auth.TokenProvider, *atomic.Int32) {
	t.Helper()
	var refreshes atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := refreshes.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"tok-%d","token_type":"Bearer","expires_in":3600}`, n)
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	data, err := json.Marshal(map[string]string{
		"type":          "authorized_user",
		"client_id":     "id",
		"client_secret": "secret",
		"refresh_token": "refresh",
		"token_uri":     srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "credentials"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials", "user@example.com.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	env := config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": dir}, "/nonexistent")
	return auth.NewTokenProvider(env), &refreshes
}

// apiServer is a fake API endpoint that records the Authorization header of
// each request.
func apiServer(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var auths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auths = append(auths, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &auths
}

func TestClientsAuthenticateAndShareTokens(t *testing.T) {
	tokens, refreshes := testTokens(t)
	srv, auths := apiServer(t)
	ctx := context.Background()

	for range 2 {
		c := &Clients{Tokens: tokens, Account: "user@example.com"}
		svc, err := c.DNSAt(ctx, srv.URL+"/")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := svc.ManagedZones.List("my-project").Context(ctx).Do(); err != nil {
			t.Fatal(err)
		}
	}
	if len(*auths) != 2 || (*auths)[0] != "Bearer tok-1" || (*auths)[1] != "Bearer tok-1" {
		t.Errorf("Authorization headers = %q, want two of %q", *auths, "Bearer tok-1")
	}
	if got := refreshes.Load(); got != 1 {
		t.Errorf("token refreshes = %d, want 1", got)
	}
}

// TestClientsDebugOutKeepsEndpoint checks that turning on request logging
// doesn't drop options such as a regional endpoint.
func TestClientsDebugOutKeepsEndpoint(t *testing.T) {
	tokens, _ := testTokens(t)
	srv, auths := apiServer(t)
	var debug bytes.Buffer
	c := &Clients{Tokens: tokens, Account: "user@example.com", DebugOut: &debug}

	ctx := context.Background()
	svc, err := c.DNSAt(ctx, srv.URL+"/")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ManagedZones.List("my-project").Context(ctx).Do(); err != nil {
		t.Fatal(err)
	}
	if len(*auths) != 1 {
		t.Fatalf("API requests = %d, want 1", len(*auths))
	}
	if !strings.Contains(debug.String(), srv.URL) {
		t.Errorf("debug output does not mention %s:\n%s", srv.URL, debug.String())
	}
}

func TestClientsNoCredentials(t *testing.T) {
	env := config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": t.TempDir()}, "/nonexistent")
	c := &Clients{Tokens: auth.NewTokenProvider(env)}
	_, err := c.Scheduler(context.Background())
	if err == nil || !strings.HasPrefix(err.Error(), "obtaining credentials: ") {
		t.Errorf("err = %v, want an \"obtaining credentials\" error", err)
	}
}
