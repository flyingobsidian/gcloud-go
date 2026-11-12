package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/flyingobsidian/gcloud-go/internal/config"
)

// tokenServer is a fake OAuth token endpoint that counts refreshes.
func tokenServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := n.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"tok-%d","token_type":"Bearer","expires_in":3600}`, i)
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

// storeUser writes an authorized_user credential for account under dir.
func storeUser(t *testing.T, dir, account, tokenURI string) {
	t.Helper()
	data, err := json.Marshal(map[string]string{
		"type":          "authorized_user",
		"client_id":     "id",
		"client_secret": "secret",
		"refresh_token": "refresh-" + account,
		"token_uri":     tokenURI,
	})
	if err != nil {
		t.Fatal(err)
	}
	credDir := filepath.Join(dir, "credentials")
	if err := os.MkdirAll(credDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(credDir, account+".json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestTokenProviderCachesPerAccountAndScopes(t *testing.T) {
	srv, refreshes := tokenServer(t)
	dir := t.TempDir()
	storeUser(t, dir, "a@example.com", srv.URL)
	storeUser(t, dir, "b@example.com", srv.URL)
	p := NewTokenProvider(config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": dir}, "/nonexistent"))

	ctx := context.Background()
	ts1, err := p.TokenSource(ctx, "a@example.com", "s1", "s2")
	if err != nil {
		t.Fatal(err)
	}
	ts2, err := p.TokenSource(ctx, "a@example.com", "s2", "s1", "s1")
	if err != nil {
		t.Fatal(err)
	}
	tok1, err := ts1.Token()
	if err != nil {
		t.Fatal(err)
	}
	tok2, err := ts2.Token()
	if err != nil {
		t.Fatal(err)
	}
	if tok1.AccessToken != tok2.AccessToken {
		t.Errorf("same account and scopes got different tokens %q and %q", tok1.AccessToken, tok2.AccessToken)
	}
	if got := refreshes.Load(); got != 1 {
		t.Errorf("refreshes = %d, want 1", got)
	}

	tsB, err := p.TokenSource(ctx, "b@example.com", "s1", "s2")
	if err != nil {
		t.Fatal(err)
	}
	tokB, err := tsB.Token()
	if err != nil {
		t.Fatal(err)
	}
	if tokB.AccessToken == tok1.AccessToken {
		t.Errorf("different accounts share token %q", tokB.AccessToken)
	}
}

func TestTokenProviderSurvivesCancelledContext(t *testing.T) {
	srv, _ := tokenServer(t)
	dir := t.TempDir()
	storeUser(t, dir, "a@example.com", srv.URL)
	p := NewTokenProvider(config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": dir}, "/nonexistent"))

	ctx, cancel := context.WithCancel(context.Background())
	ts, err := p.TokenSource(ctx, "a@example.com", "s")
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := ts.Token(); err != nil {
		t.Fatalf("Token after the creating context was cancelled: %v", err)
	}
}

func TestTokenProviderUsesActiveAccountFromEnv(t *testing.T) {
	dir := t.TempDir()
	storeUser(t, dir, "active@example.com", "http://unused")
	if err := os.MkdirAll(filepath.Join(dir, "configurations"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configurations", "config_default"),
		[]byte("[core]\naccount = active@example.com\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/nonexistent/process-env.json")

	p := NewTokenProvider(config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": dir}, "/nonexistent"))
	if _, err := p.TokenSource(context.Background(), "", "s"); err != nil {
		t.Fatalf("TokenSource with the active account: %v", err)
	}
}

func TestTokenProviderNoAccount(t *testing.T) {
	p := NewTokenProvider(config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": t.TempDir()}, "/nonexistent"))
	if _, err := p.TokenSource(context.Background(), "", "s"); err == nil {
		t.Fatal("TokenSource with no account succeeded, want error")
	}
}

func TestTokenProviderConcurrent(t *testing.T) {
	srv, refreshes := tokenServer(t)
	dir := t.TempDir()
	storeUser(t, dir, "a@example.com", srv.URL)
	p := NewTokenProvider(config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": dir}, "/nonexistent"))

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			ts, err := p.TokenSource(context.Background(), "a@example.com", "s")
			if err != nil {
				t.Error(err)
				return
			}
			if _, err := ts.Token(); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if got := refreshes.Load(); got != 1 {
		t.Errorf("refreshes = %d, want 1", got)
	}
}
