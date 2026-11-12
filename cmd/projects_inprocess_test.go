package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// fakeCRM serves the Cloud Resource Manager v3 calls used by projects list,
// describe and get-ancestors. Every project has the ID it is asked for and
// sits in folder 10, which is in organization 20.
func fakeCRM(t *testing.T) http.Handler {
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
		case p == "/v3/projects:search":
			write(w, map[string]any{"projects": []map[string]any{
				{"name": "projects/111", "projectId": "alpha-proj", "displayName": "Alpha", "state": "ACTIVE"},
				{"name": "projects/222", "projectId": "beta-proj", "displayName": "Beta", "state": "ACTIVE"},
			}})
		case strings.HasPrefix(p, "/v3/projects/"):
			id := strings.TrimPrefix(p, "/v3/projects/")
			write(w, map[string]any{"name": "projects/" + id, "projectId": id, "parent": "folders/10"})
		case p == "/v3/folders/10":
			write(w, map[string]any{"name": "folders/10", "parent": "organizations/20"})
		default:
			http.NotFound(w, r)
		}
	})
}

func TestProjectsListInProcess(t *testing.T) {
	inv, stdout, _ := newFakeAPIInvocation(t, fakeCRM(t))
	if err := runInvocation(inv, "projects", "list"); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%-30s %-30s %-20s %s\n", "PROJECT_ID", "NAME", "PROJECT_NUMBER", "STATE") +
		fmt.Sprintf("%-30s %-30s %-20s %s\n", "alpha-proj", "Alpha", "111", "ACTIVE") +
		fmt.Sprintf("%-30s %-30s %-20s %s\n", "beta-proj", "Beta", "222", "ACTIVE")
	if got := stdout.String(); got != want {
		t.Errorf("projects list:\n%s\nwant:\n%s", got, want)
	}

	stdout.Reset()
	if err := runInvocation(inv, "projects", "list", "--format=json"); err != nil {
		t.Fatal(err)
	}
	var projects []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &projects); err != nil {
		t.Fatalf("projects list --format=json is not JSON: %v\n%s", err, stdout.String())
	}
	if len(projects) != 2 || projects[1]["projectId"] != "beta-proj" {
		t.Errorf("projects list --format=json = %v", projects)
	}
}

func TestProjectsConcurrentInProcess(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			inv, stdout, _ := newFakeAPIInvocation(t, fakeCRM(t))
			id := fmt.Sprintf("project-%02d", i)
			if err := runInvocation(inv, "projects", "get-ancestors", id); err != nil {
				t.Error(err)
				return
			}
			want := fmt.Sprintf("%-40s %s\n%-40s %s\n%-40s %s\n%-40s %s\n",
				"ID", "TYPE", id, "project", "10", "folder", "20", "organization")
			if got := stdout.String(); got != want {
				t.Errorf("get-ancestors %s:\n%s\nwant:\n%s", id, got, want)
			}
		})
	}
	wg.Wait()
}

func TestProjectsDescribeError(t *testing.T) {
	inv, _, _ := newFakeAPIInvocation(t, http.NotFoundHandler())
	err := runInvocation(inv, "projects", "describe", "missing-proj")
	if err == nil || !strings.Contains(err.Error(), "describing project") {
		t.Errorf("err = %v, want a describing project error", err)
	}
}
