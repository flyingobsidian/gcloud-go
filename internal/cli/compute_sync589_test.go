package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	compute "google.golang.org/api/compute/v1"
)

// --- #1948: gcloud-python 583.0.0–589.0.0 Compute Engine changes ---

func TestConstructMetadataDictStdin(t *testing.T) {
	file := filepath.Join(t.TempDir(), "startup.sh")
	if err := os.WriteFile(file, []byte("from file"), 0600); err != nil {
		t.Fatal(err)
	}
	inv, _, _ := newTestInvocation(t, nil, "from stdin")
	s := newComputeCmds(inv)

	got, err := s.constructMetadataDict(map[string]string{"a": "b"}, map[string]string{"user-data": "-", "startup-script": file})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"a": "b", "user-data": "from stdin", "startup-script": "from file"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if _, err := s.constructMetadataDict(nil, map[string]string{"x": "-", "y": "-"}); err == nil ||
		err.Error() != "Cannot specify [-] for multiple keys in [--metadata-from-file]." {
		t.Errorf("two stdin keys: err = %v", err)
	}
	if _, err := s.constructMetadataDict(map[string]string{"k": "v"}, map[string]string{"k": file}); err == nil ||
		err.Error() != "Encountered duplicate metadata key [k]." {
		t.Errorf("duplicate key: err = %v", err)
	}
	if _, err := s.addMetadataDict(nil, nil); err == nil ||
		err.Error() != "At least one of [--metadata] or [--metadata-from-file] must be provided." {
		t.Errorf("no metadata: err = %v", err)
	}
}

func TestBuildMetadataSorted(t *testing.T) {
	if buildMetadata(nil) != nil {
		t.Error("empty metadata should be nil")
	}
	m := buildMetadata(map[string]string{"b": "2", "a": "1"})
	if len(m.Items) != 2 || m.Items[0].Key != "a" || *m.Items[1].Value != "2" {
		t.Errorf("items = %+v", m.Items)
	}
}

func TestSerialOutputLost(t *testing.T) {
	for _, tc := range []struct {
		start, gotStart, next int64
		want                  bool
	}{
		{start: 0, gotStart: 0, next: 100},
		{start: 10, gotStart: 50, next: 100, want: true},
		{start: -20, gotStart: 80, next: 100},
		{start: -20, gotStart: 90, next: 100, want: true},
		{start: -200, gotStart: 0, next: 100},
		{start: -200, gotStart: 5, next: 100, want: true},
	} {
		out := &compute.SerialPortOutput{Start: tc.gotStart, Next: tc.next}
		if got := serialOutputLost(tc.start, out); got != tc.want {
			t.Errorf("serialOutputLost(%d, start=%d next=%d) = %v, want %v", tc.start, tc.gotStart, tc.next, got, tc.want)
		}
	}
}

func TestGetSerialPortOutputEpilog(t *testing.T) {
	var query string
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"contents":"boot log","start":"90","next":"100"}`)
	})
	inv, stdout, stderr := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "compute", "instances", "get-serial-port-output", "vm", "--project", "p", "--zone", "z", "--start=-20"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(query, "start=-20") {
		t.Errorf("query = %q, want start=-20", query)
	}
	if stdout.String() != "boot log\n" {
		t.Errorf("stdout = %q", stdout)
	}
	want := "WARNING: Some serial port output was lost due to a limited buffer. The oldest byte of output returned was at offset 90.\n" +
		"\nSpecify --start=100 in the next get-serial-port-output invocation to get only the new output starting from here.\n"
	if stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}
}

func TestManagedListInstancesPaging(t *testing.T) {
	for _, tc := range []struct {
		name      string
		limit     int
		pageSize  int64
		filter    string
		wantMax   int64
		wantLimit int
	}{
		{name: "none"},
		{name: "limit", limit: 5, wantMax: 5, wantLimit: 5},
		{name: "page-size", pageSize: 50, wantMax: 50},
		{name: "smaller wins", limit: 5, pageSize: 50, wantMax: 5, wantLimit: 5},
		{name: "capped", pageSize: 2000, wantMax: 500},
		{name: "filter disables", limit: 5, pageSize: 50, filter: "status=RUNNING"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inv, _, _ := newTestInvocation(t, nil, "")
			inv.flags.filter = tc.filter
			s := newComputeCmds(inv)
			s.flagManagedListInstLimit, s.flagManagedListInstPageSize = tc.limit, tc.pageSize
			gotMax, gotLimit := s.managedListInstancesPaging()
			if gotMax != tc.wantMax || gotLimit != tc.wantLimit {
				t.Errorf("paging = (%d, %d), want (%d, %d)", gotMax, gotLimit, tc.wantMax, tc.wantLimit)
			}
		})
	}
}

func TestManagedListInstancesTable(t *testing.T) {
	var queries []string
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query().Get("maxResults")+"/"+r.URL.Query().Get("pageToken"))
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("pageToken") == "" {
			fmt.Fprint(w, `{"managedInstances":[{"name":"vm-1","instance":"https://compute.googleapis.com/compute/v1/projects/p/zones/z1/instances/vm-1","instanceStatus":"RUNNING","currentAction":"NONE","version":{"instanceTemplate":"global/instanceTemplates/t1","name":"v1"},"scheduling":{"terminationTimestamp":"2026-11-20T00:00:00Z"}}],"nextPageToken":"n"}`)
			return
		}
		fmt.Fprint(w, `{"managedInstances":[{"name":"vm-2","instance":"https://compute.googleapis.com/compute/v1/projects/p/zones/z1/instances/vm-2","instanceStatus":"STOPPING","currentAction":"DELETING","lastAttempt":{"errors":{"errors":[{"code":"E1","message":"boom"}]}}}]}`)
	})
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "compute", "instance-groups", "managed", "list-instances", "mig", "--project", "p", "--zone", "z1"); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	for _, want := range []string{"NAME", "ZONE", "STATUS", "HEALTH_STATE", "ACTION", "INSTANCE_TEMPLATE", "VERSION_NAME", "LAST_ERROR", "TERMINATION_TIMESTAMP", "vm-1", "t1", "v1", "Error E1: boom", "2026-11-20T00:00:00Z"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "GRACEFUL_SHUTDOWN_TIMESTAMP") {
		t.Errorf("GRACEFUL_SHUTDOWN_TIMESTAMP shown with no values:\n%s", out)
	}
	if len(queries) != 2 || queries[0] != "/" {
		t.Errorf("queries = %v, want two pages without maxResults", queries)
	}

	queries = nil
	inv, stdout, _ = newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "compute", "instance-groups", "managed", "list-instances", "mig", "--project", "p", "--zone", "z1", "--limit", "1", "--format", "value(name)"); err != nil {
		t.Fatal(err)
	}
	if len(queries) != 1 || queries[0] != "1/" {
		t.Errorf("queries = %v, want one page with maxResults=1", queries)
	}
	if stdout.String() != "vm-1\n" {
		t.Errorf("stdout = %q, want only vm-1", stdout)
	}
}

func TestManagedNoGracefulShutdown(t *testing.T) {
	var got []string
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Method+" "+r.URL.Query().Get("noGracefulShutdown"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"op","status":"DONE"}`)
	})
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"delete", "mig", "--region", "r", "--quiet"}, "DELETE "},
		{[]string{"delete", "mig", "--region", "r", "--quiet", "--no-graceful-shutdown"}, "DELETE true"},
		{[]string{"recreate-instances", "mig", "--region", "r", "--instances", "vm-1", "--no-graceful-shutdown"}, "POST true"},
	} {
		got = nil
		inv, _, _ := newFakeAPIInvocation(t, api)
		if err := runInvocation(inv, append([]string{"compute", "instance-groups", "managed", "--project", "p"}, tc.args...)...); err != nil {
			t.Fatalf("%v: %v", tc.args, err)
		}
		if len(got) == 0 || got[0] != tc.want {
			t.Errorf("%v: requests = %v, want first %q", tc.args, got, tc.want)
		}
	}
}

func TestImageViewsList(t *testing.T) {
	var paths []string
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		project := strings.Split(strings.TrimPrefix(r.URL.Path, "/projects/"), "/")[0]
		if project == "p" {
			fmt.Fprint(w, `{"items":[{"name":"v1","selfLink":"https://x/projects/p/regions/r1/imageViews/v1","image":{"name":"mine","family":"fam","status":"READY","selfLink":"https://x/projects/p/global/images/mine"}},{"name":"v2","image":{"name":"old","status":"READY","deprecated":{"state":"DEPRECATED"}}}]}`)
			return
		}
		fmt.Fprint(w, `{"items":[]}`)
	})
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "compute", "image-views", "list", "--project", "p", "--region", "r1", "--no-standard-images"); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || !strings.HasSuffix(paths[0], "/projects/p/regions/r1/imageViews") {
		t.Errorf("paths = %v, want only project p", paths)
	}
	out := stdout.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "mine") || !strings.Contains(out, "fam") || strings.Contains(out, "old") {
		t.Errorf("stdout = %q, want mine without the deprecated image", out)
	}

	paths = nil
	inv, _, _ = newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "compute", "image-views", "list", "--project", "p", "--region", "r1", "--show-deprecated", "--format", "value(name)"); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1+len(computePublicImageProjects()) {
		t.Errorf("listed %d projects, want p and the public image projects", len(paths))
	}
}

func TestImageViewsRegionRequired(t *testing.T) {
	inv, _, _ := newFakeAPIInvocation(t, http.NotFoundHandler())
	err := runInvocation(inv, "compute", "project-views", "describe", "--project", "p")
	if err == nil || !strings.HasPrefix(err.Error(), "The required property [region] is not currently set.") {
		t.Fatalf("err = %v", err)
	}
}

func TestProjectViewsDescribe(t *testing.T) {
	var path string
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"project":{"name":"pv"}}`)
	})
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "compute", "project-views", "describe", "--project", "p", "--region", "r1", "--format", "value(project.name)"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "/projects/p/regions/r1/projectViews") || stdout.String() != "pv\n" {
		t.Errorf("path = %q, stdout = %q", path, stdout)
	}
}
