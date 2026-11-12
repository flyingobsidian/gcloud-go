package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// fakeResourceManager serves the folders and tag keys parts of the CRM v3 API.
func fakeResourceManager(t *testing.T) http.Handler {
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
		switch p := r.URL.Path; {
		case p == "/v3/folders":
			parent := r.URL.Query().Get("parent")
			write(w, map[string]any{"folders": []map[string]any{
				{"name": "folders/30", "displayName": "Eng", "parent": parent, "state": "ACTIVE"},
			}})
		case strings.HasPrefix(p, "/v3/folders/"):
			id := strings.TrimPrefix(p, "/v3/folders/")
			write(w, map[string]any{"name": "folders/" + id, "displayName": "folder-" + id})
		case p == "/v3/tagKeys":
			write(w, map[string]any{"tagKeys": []map[string]any{
				{"name": "tagKeys/40", "shortName": "env", "namespacedName": "20/env"},
			}})
		case strings.HasPrefix(p, "/v3/tagKeys/"):
			id := strings.TrimPrefix(p, "/v3/tagKeys/")
			write(w, map[string]any{"name": "tagKeys/" + id, "shortName": "key-" + id})
		default:
			http.NotFound(w, r)
		}
	})
}

func TestResourceManagerListInProcess(t *testing.T) {
	inv, stdout, _ := newFakeAPIInvocation(t, fakeResourceManager(t))
	if err := runInvocation(inv, "resource-manager", "folders", "list", "--organization=20"); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%-30s %-15s %-20s %s\n", "DISPLAY_NAME", "ID", "PARENT", "STATE") +
		fmt.Sprintf("%-30s %-15s %-20s %s\n", "Eng", "30", "organizations/20", "ACTIVE")
	if got := stdout.String(); got != want {
		t.Errorf("folders list:\n%s\nwant:\n%s", got, want)
	}

	stdout.Reset()
	if err := runInvocation(inv, "resource-manager", "tags", "keys", "list", "--parent=organizations/20", "--format=json"); err != nil {
		t.Fatal(err)
	}
	var keys []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &keys); err != nil {
		t.Fatalf("tags keys list --format=json is not JSON: %v\n%s", err, stdout.String())
	}
	if len(keys) != 1 || keys[0]["namespacedName"] != "20/env" {
		t.Errorf("tags keys list --format=json = %v", keys)
	}
}

func TestResourceManagerConcurrentInProcess(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			inv, stdout, _ := newFakeAPIInvocation(t, fakeResourceManager(t))
			id := fmt.Sprint(1000 + i)
			args := []string{"resource-manager", "folders", "describe", id}
			want := fmt.Sprintf("displayName: folder-%s\nname: folders/%s\n", id, id)
			if i%2 == 1 {
				args = []string{"resource-manager", "tags", "keys", "describe", "tagKeys/" + id}
				want = fmt.Sprintf("name: tagKeys/%s\nshortName: key-%s\n", id, id)
			}
			if err := runInvocation(inv, args...); err != nil {
				t.Error(err)
				return
			}
			if got := stdout.String(); got != want {
				t.Errorf("%s = %q, want %q", strings.Join(args, " "), got, want)
			}
		})
	}
	wg.Wait()
}
