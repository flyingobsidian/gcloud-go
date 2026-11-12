package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// fakeStorageLogging serves bucket and log-metric lookups that echo the
// requested name, for checking that concurrent invocations stay separate.
func fakeStorageLogging(t *testing.T) http.Handler {
	t.Helper()
	write := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		p := r.URL.Path
		switch {
		case strings.Contains(p, "/b/"):
			name := p[strings.LastIndex(p, "/b/")+3:]
			write(w, map[string]any{"name": name, "location": "EU"})
		case strings.HasSuffix(p, "/b"):
			write(w, map[string]any{"items": []map[string]any{{"name": "bucket-" + r.URL.Query().Get("project")}}})
		case strings.Contains(p, "/metrics/"):
			name := p[strings.Index(p, "projects/"):]
			write(w, map[string]any{"name": name, "filter": "severity>=ERROR"})
		default:
			http.NotFound(w, r)
		}
	})
}

func TestStorageBucketsListInProcess(t *testing.T) {
	inv, stdout, _ := newFakeAPIInvocation(t, fakeStorageLogging(t))
	if err := runInvocation(inv, "storage", "buckets", "list", "--project=p1", "--format=json"); err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("storage buckets list --format=json is not JSON: %v\n%s", err, stdout.String())
	}
	if len(got) != 1 || got[0]["name"] != "bucket-p1" {
		t.Errorf("storage buckets list = %v", got)
	}
}

func TestStorageLoggingConcurrentInProcess(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			inv, stdout, _ := newFakeAPIInvocation(t, fakeStorageLogging(t))
			var args []string
			var want string
			if i%2 == 0 {
				name := fmt.Sprintf("bucket-%d", i)
				args = []string{"storage", "buckets", "describe", "gs://" + name}
				want = name
			} else {
				project := fmt.Sprintf("p%d", i)
				args = []string{"logging", "metrics", "describe", "errors", "--project=" + project, "--format=json"}
				want = "projects/" + project + "/metrics/errors"
			}
			if err := runInvocation(inv, args...); err != nil {
				t.Error(err)
				return
			}
			var got struct{ Name string }
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Errorf("%s: output is not JSON: %v\n%s", strings.Join(args, " "), err, stdout.String())
				return
			}
			if got.Name != want {
				t.Errorf("%s: name = %q, want %q", strings.Join(args, " "), got.Name, want)
			}
		})
	}
	wg.Wait()
}
