package gcp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

// TestWithRequestLogging checks that the logging transport sees the request
// after the auth transport has stamped its headers on, and that the request
// still reaches the server unchanged.
func TestWithRequestLogging(t *testing.T) {
	var gotAuth, gotUserProject string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUserProject = r.Header.Get("X-Goog-User-Project")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	var debug bytes.Buffer
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token", TokenType: "Bearer"})
	client, err := loggingHTTPClient(ctx, &debug,
		option.WithTokenSource(ts), option.WithQuotaProject("qp-123"))
	if err != nil {
		t.Fatalf("loggingHTTPClient: %v", err)
	}
	resp, err := client.Get(server.URL + "/v1/organizations/1234/sources/-/findings?alt=json")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if !strings.HasPrefix(gotAuth, "Bearer ") {
		t.Errorf("server saw Authorization %q, want a Bearer token", gotAuth)
	}
	if gotUserProject != "qp-123" {
		t.Errorf("server saw X-Goog-User-Project %q, want qp-123", gotUserProject)
	}

	logged := debug.String()
	for _, want := range []string{
		"curl -X GET '" + server.URL + "/v1/organizations/1234/sources/-/findings?alt=json'",
		`-H "Authorization: Bearer $(gcloud auth print-access-token)"`,
		"-H 'X-Goog-User-Project: qp-123'",
	} {
		if !strings.Contains(logged, want) {
			t.Errorf("debug output missing %q; got:\n%s", want, logged)
		}
	}
	if strings.Contains(logged, "test-token") {
		t.Error("debug output leaked the access token")
	}
}

// TestWithRequestLoggingDisabled checks the no-op path: with no debug writer
// the caller's options must reach the generated client untouched.
func TestWithRequestLoggingDisabled(t *testing.T) {
	in := []option.ClientOption{option.WithQuotaProject("qp-123")}
	out, err := withRequestLogging(context.Background(), nil, in...)
	if err != nil {
		t.Fatalf("withRequestLogging: %v", err)
	}
	if len(out) != 1 || out[0] != in[0] {
		t.Errorf("expected options to pass through unchanged, got %v", out)
	}
}

// TestWithRequestLoggingReplacesOptions checks that enabling logging hands the
// generated client a single prebuilt HTTP client instead of the auth options,
// which google.golang.org/api rejects in combination.
func TestWithRequestLoggingReplacesOptions(t *testing.T) {
	var debug bytes.Buffer
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token", TokenType: "Bearer"})
	out, err := withRequestLogging(context.Background(), &debug, option.WithTokenSource(ts))
	if err != nil {
		t.Fatalf("withRequestLogging: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected exactly one client option, got %d", len(out))
	}
}
