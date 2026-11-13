package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// fakeWorkflows serves the parts of the Workflows and Workflow Executions
// APIs the workflows commands use.
func fakeWorkflows(t *testing.T) http.Handler {
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
		p := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch parts := strings.Split(p, "/"); {
		case len(parts) == 5 && parts[4] == "workflows":
			write(w, map[string]any{"workflows": []map[string]any{
				{"name": p + "/wf1", "state": "ACTIVE", "revisionId": "000001-abc"},
			}})
		case len(parts) == 7 && parts[6] == "executions":
			write(w, map[string]any{"executions": []map[string]any{
				{"name": p + "/e1", "state": "SUCCEEDED"},
			}})
		case len(parts) == 8 && parts[6] == "executions":
			write(w, map[string]any{"name": p, "state": "SUCCEEDED", "result": `"ok"`})
		default:
			http.NotFound(w, r)
		}
	})
}

func TestWorkflowsListInProcess(t *testing.T) {
	inv, stdout, stderr := newFakeAPIInvocation(t, fakeWorkflows(t))
	if err := runInvocation(inv, "workflows", "list", "--project=p1", "--format=json"); err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("workflows list --format=json is not JSON: %v\n%s", err, stdout.String())
	}
	if len(got) != 1 || got[0]["name"] != "projects/p1/locations/us-central1/workflows/wf1" {
		t.Errorf("workflows list = %v", got)
	}
	if want := "WARNING: The default location(us-central1) was used since the location flag was not specified.\n"; stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}

	stdout.Reset()
	stderr.Reset()
	if err := runInvocation(inv, "workflows", "executions", "list", "wf1", "--project=p1", "--location=europe-west2", "--format=value(name)"); err != nil {
		t.Fatal(err)
	}
	if want := "projects/p1/locations/europe-west2/workflows/wf1/executions/e1\n"; stdout.String() != want {
		t.Errorf("executions list = %q, want %q", stdout.String(), want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestWorkflowsConcurrentInProcess(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			inv, stdout, _ := newFakeAPIInvocation(t, fakeWorkflows(t))
			id := fmt.Sprintf("e%d", i)
			if err := runInvocation(inv, "workflows", "executions", "describe", id,
				"--project=p1", "--location=l1", "--workflow=wf1", "--format=value(name)"); err != nil {
				t.Error(err)
				return
			}
			if want := "projects/p1/locations/l1/workflows/wf1/executions/" + id + "\n"; stdout.String() != want {
				t.Errorf("describe %s = %q, want %q", id, stdout.String(), want)
			}
		})
	}
	wg.Wait()
}
