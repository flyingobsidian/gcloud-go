package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// migratedLeafPaths returns the argument paths of every command in the
// migrated groups, e.g. {"tasks", "queues", "list"}.
func migratedLeafPaths(t *testing.T) [][]string {
	t.Helper()
	inv, _, _ := newTestInvocation(t, nil, "")
	var paths [][]string
	var walk func(c *cobra.Command, path []string)
	walk = func(c *cobra.Command, path []string) {
		path = append(path[:len(path):len(path)], c.Name())
		paths = append(paths, path)
		for _, sub := range c.Commands() {
			walk(sub, path)
		}
	}
	for _, g := range migratedGroups(inv) {
		walk(g, nil)
	}
	return paths
}

type invocationResult struct{ stdout, stderr, err string }

// runIsolated runs args in a fresh invocation from newInv, giving up after a
// few seconds.
func runIsolated(newInv func() (*invocation, *bytes.Buffer, *bytes.Buffer), args ...string) invocationResult {
	inv, stdout, stderr := newInv()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	root := newRootCmd(inv)
	root.SetArgs(args)
	inv.ctx = ctx
	err := root.ExecuteContext(ctx)
	r := invocationResult{stdout: stdout.String(), stderr: stderr.String()}
	if err != nil {
		r.err = err.Error()
	}
	return r
}

// Every migrated command keeps its state in its invocation (#2070): run
// --help for each group, and each list command against an empty fake API,
// once on its own and then several times at once, and check that the
// concurrent runs print exactly what the lone run printed. Run with -race
// to catch state shared through package variables.
func TestMigratedGroupsConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("runs every migrated command")
	}
	empty := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{}")
	})
	newInv := newFakeAPIFactory(t, empty)
	// Commands that call REST endpoints directly would otherwise reach the
	// real Google APIs.
	redirectGoogleAPIs(t, empty)
	var cases [][]string
	for _, p := range migratedLeafPaths(t) {
		if len(p) == 1 {
			cases = append(cases, slices.Concat(p, []string{"--help"}))
		}
		if p[len(p)-1] == "list" {
			cases = append(cases, slices.Concat(p, []string{"--project=p"}))
		}
	}
	want := make([]invocationResult, len(cases))
	for i, args := range cases {
		want[i] = runIsolated(newInv, args...)
	}
	const copies = 3
	sem := make(chan struct{}, 16)
	var wg sync.WaitGroup
	for i, args := range cases {
		for range copies {
			wg.Go(func() {
				sem <- struct{}{}
				defer func() { <-sem }()
				if got := runIsolated(newInv, args...); got != want[i] {
					t.Errorf("%s: concurrent run = %+v, want %+v", strings.Join(args, " "), got, want[i])
				}
			})
		}
	}
	wg.Wait()
}

// redirectGoogleAPIs sends every request to a *.googleapis.com host to api
// for the rest of the test, by replacing http.DefaultTransport.
func redirectGoogleAPIs(t *testing.T, api http.Handler) {
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	base := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Hostname(), ".googleapis.com") {
			r = r.Clone(r.Context())
			r.URL.Scheme, r.URL.Host = target.Scheme, target.Host
		}
		return base.RoundTrip(r)
	})
	t.Cleanup(func() { http.DefaultTransport = base })
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
