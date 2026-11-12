package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// fakeCompute serves instance lookups that echo the requested project, zone
// and name, and a one-disk aggregated list.
func fakeCompute(t *testing.T) http.Handler {
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
		parts := strings.Split(strings.Trim(p[strings.Index(p, "projects/"):], "/"), "/")
		switch {
		case len(parts) == 6 && parts[4] == "instances":
			write(w, map[string]any{
				"name":     parts[5],
				"zone":     "https://www.googleapis.com/compute/v1/projects/" + parts[1] + "/zones/" + parts[3],
				"selfLink": "https://www.googleapis.com/compute/v1/" + strings.Join(parts, "/"),
				"status":   "RUNNING",
			})
		case len(parts) == 4 && parts[2] == "aggregated" && parts[3] == "disks":
			write(w, map[string]any{"items": map[string]any{
				"zones/europe-west2-a": map[string]any{"disks": []map[string]any{{
					"name": "d1", "sizeGb": "10", "status": "READY",
					"zone": "https://www.googleapis.com/compute/v1/projects/" + parts[1] + "/zones/europe-west2-a",
					"type": "https://www.googleapis.com/compute/v1/projects/" + parts[1] + "/zones/europe-west2-a/diskTypes/pd-balanced",
				}}},
			}})
		default:
			http.NotFound(w, r)
		}
	})
}

func TestComputeDisksListInProcess(t *testing.T) {
	inv, stdout, _ := newFakeAPIInvocation(t, fakeCompute(t))
	if err := runInvocation(inv, "compute", "disks", "list", "--project=p1", "--format=value(name,sizeGb)"); err != nil {
		t.Fatal(err)
	}
	if want := "d1\t10\n"; stdout.String() != want {
		t.Errorf("compute disks list = %q, want %q", stdout.String(), want)
	}
}

func TestComputeConcurrentInProcess(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			inv, stdout, _ := newFakeAPIInvocation(t, fakeCompute(t))
			name := fmt.Sprintf("vm-%d", i)
			zone := fmt.Sprintf("europe-west2-%c", 'a'+i%3)
			if err := runInvocation(inv, "compute", "instances", "describe", name, "--project=p1", "--zone="+zone, "--format=json"); err != nil {
				t.Error(err)
				return
			}
			var got struct{ Name, Zone string }
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Errorf("describe %s: output is not JSON: %v\n%s", name, err, stdout.String())
				return
			}
			if got.Name != name || !strings.HasSuffix(got.Zone, "/zones/"+zone) {
				t.Errorf("describe %s in %s = %+v", name, zone, got)
			}
		})
	}
	wg.Wait()
}
