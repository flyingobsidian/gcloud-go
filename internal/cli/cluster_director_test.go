package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClusterDirectorCommandTree(t *testing.T) {
	g := newClusterDirectorCmds(legacyTestInvocation())
	assertSubcommands(t, g.clusterDirectorCmd, []string{"clusters", "operations", "locations"})
	assertSubcommands(t, g.cdClustersCmd, []string{"create", "delete", "describe", "list", "update"})
	assertSubcommands(t, g.cdOperationsCmd, []string{"describe"})
	assertSubcommands(t, g.cdLocationsCmd, []string{"describe", "list"})
}

// cdFakeAPI records requests and serves hypercomputecluster v1 responses.
type cdFakeAPI struct {
	requests []string
	bodies   []string
}

func (f *cdFakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.requests = append(f.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	f.bodies = append(f.bodies, string(body))
	w.Header().Set("Content-Type", "application/json")
	p := r.URL.Path
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(p, "/locations/us-central1/clusters"):
		if r.URL.Query().Get("pageToken") == "" {
			fmt.Fprint(w, `{"clusters":[{"name":"projects/p/locations/us-central1/clusters/a"}],"nextPageToken":"n"}`)
			return
		}
		fmt.Fprint(w, `{"clusters":[{"name":"projects/p/locations/us-central1/clusters/b"}]}`)
	case r.Method == http.MethodGet && strings.HasSuffix(p, "/clusters/c1"):
		fmt.Fprint(w, `{"name":"projects/p/locations/us-central1/clusters/c1","description":"d"}`)
	case r.Method == http.MethodPost && strings.HasSuffix(p, "/clusters"):
		fmt.Fprint(w, `{"name":"projects/p/locations/us-central1/operations/op1","done":true,"response":{"name":"projects/p/locations/us-central1/clusters/c1","description":"new"}}`)
	case r.Method == http.MethodPatch:
		fmt.Fprint(w, `{"name":"projects/p/locations/us-central1/operations/op2"}`)
	case r.Method == http.MethodDelete:
		fmt.Fprint(w, `{"name":"projects/p/locations/us-central1/operations/op3","done":true}`)
	case r.Method == http.MethodGet && strings.HasSuffix(p, "/projects/p/locations"):
		fmt.Fprint(w, `{"locations":[{"name":"projects/p/locations/us-central1","locationId":"us-central1"}]}`)
	default:
		http.Error(w, r.Method+" "+p, http.StatusNotFound)
	}
}

func TestClusterDirectorClustersList(t *testing.T) {
	api := &cdFakeAPI{}
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "cluster-director", "clusters", "list", "--location", "us-central1", "--project", "p", "--format", "value(name)"); err != nil {
		t.Fatal(err)
	}
	if want := "projects/p/locations/us-central1/clusters/a\nprojects/p/locations/us-central1/clusters/b\n"; stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func TestClusterDirectorClustersCreateConfig(t *testing.T) {
	api := &cdFakeAPI{}
	inv, stdout, stderr := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "cluster-director", "clusters", "create", "c1", "--location", "us-central1", "--project", "p",
		"--config", `{"description":"new","labels":{"env":"prod"}}`, "--format", "value(description)")
	if err != nil {
		t.Fatal(err)
	}
	if want := "POST /v1/projects/p/locations/us-central1/clusters?alt=json&clusterId=c1&prettyPrint=false"; len(api.requests) != 1 || api.requests[0] != want {
		t.Errorf("requests = %v, want [%s]", api.requests, want)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(api.bodies[0]), &sent); err != nil || sent["description"] != "new" {
		t.Errorf("body = %s", api.bodies[0])
	}
	if stdout.String() != "new\n" {
		t.Errorf("stdout = %q, want the created cluster", stdout)
	}
	wantErr := "Create request issued for: [c1]\nWaiting for operation [projects/p/locations/us-central1/operations/op1] to complete...done.\nCreated cluster [c1].\n"
	if stderr.String() != wantErr {
		t.Errorf("stderr = %q, want %q", stderr, wantErr)
	}
}

func TestClusterDirectorClustersCreateConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cluster.json")
	if err := os.WriteFile(path, []byte(`{"description":"from file"}`), 0600); err != nil {
		t.Fatal(err)
	}
	api := &cdFakeAPI{}
	inv, _, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "cluster-director", "clusters", "create", "c1", "--location", "us-central1", "--project", "p", "--config", path); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(api.bodies[0], `"from file"`) {
		t.Errorf("body = %s, want the file's cluster", api.bodies[0])
	}
}

func TestClusterDirectorClustersCreateUnknownField(t *testing.T) {
	api := &cdFakeAPI{}
	inv, _, _ := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "cluster-director", "clusters", "create", "c1", "--location", "us-central1", "--project", "p", "--config", `{"nope":1}`)
	if err == nil || !strings.Contains(err.Error(), `unknown field "nope"`) {
		t.Fatalf("err = %v, want an unknown field error", err)
	}
	if len(api.requests) != 0 {
		t.Errorf("requests = %v, want none", api.requests)
	}
}

func TestClusterDirectorClustersUpdateAsync(t *testing.T) {
	api := &cdFakeAPI{}
	inv, stdout, stderr := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "cluster-director", "clusters", "update", "c1", "--location", "us-central1", "--project", "p",
		"--config", `{"description":"x"}`, "--update-mask", "description", "--async", "--format", "value(name)")
	if err != nil {
		t.Fatal(err)
	}
	if want := "PATCH /v1/projects/p/locations/us-central1/clusters/c1?alt=json&prettyPrint=false&updateMask=description"; api.requests[0] != want {
		t.Errorf("request = %s, want %s", api.requests[0], want)
	}
	if stdout.String() != "projects/p/locations/us-central1/operations/op2\n" {
		t.Errorf("stdout = %q, want the operation", stdout)
	}
	if stderr.String() != "Update request issued for: [c1]\n" {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestClusterDirectorClustersUpdateNeedsMask(t *testing.T) {
	inv, _, _ := newFakeAPIInvocation(t, &cdFakeAPI{})
	err := runInvocation(inv, "cluster-director", "clusters", "update", "c1", "--location", "us-central1", "--project", "p", "--config", `{}`)
	if err == nil || !strings.Contains(err.Error(), "update-mask") {
		t.Fatalf("err = %v, want --update-mask required", err)
	}
}

func TestClusterDirectorClustersDelete(t *testing.T) {
	api := &cdFakeAPI{}
	inv, stdout, stderr := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "cluster-director", "clusters", "delete", "projects/p/locations/us-central1/clusters/c1"); err != nil {
		t.Fatal(err)
	}
	if want := "DELETE /v1/projects/p/locations/us-central1/clusters/c1?alt=json&prettyPrint=false"; api.requests[0] != want {
		t.Errorf("request = %s, want %s", api.requests[0], want)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want nothing", stdout)
	}
	if want := "Delete request issued for: [c1]\nWaiting for operation [projects/p/locations/us-central1/operations/op3] to complete...done.\nDeleted cluster [c1].\n"; stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}
}

func TestClusterDirectorLocationsList(t *testing.T) {
	inv, stdout, _ := newFakeAPIInvocation(t, &cdFakeAPI{})
	if err := runInvocation(inv, "cluster-director", "locations", "list", "--project", "p", "--format", "value(locationId)"); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "us-central1\n" {
		t.Errorf("stdout = %q", stdout)
	}
}
