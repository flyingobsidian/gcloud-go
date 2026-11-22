package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// Every group gcloud-python 589.0.0 gave a labels-only update has one.
func TestOracleLabelsUpdateCommands(t *testing.T) {
	for group := range odbLabelsTargets() {
		g := oracleSubgroup(group)
		if g == nil {
			t.Errorf("oracle-database %s missing", group)
			continue
		}
		if findSub(g, "update") == nil {
			t.Errorf("oracle-database %s update missing", group)
		}
	}
}

// odbLabelsFakeAPI serves a cloud VM cluster with labels env=dev,team=a and
// records the requests and bodies.
type odbLabelsFakeAPI struct {
	opDone   bool
	requests []string
	bodies   []string
}

func (f *odbLabelsFakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.requests = append(f.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	f.bodies = append(f.bodies, string(body))
	w.Header().Set("Content-Type", "application/json")
	const name = "projects/p/locations/us-east4/cloudVmClusters/vm1"
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/"+name:
		fmt.Fprintf(w, `{"name":%q,"labels":{"env":"dev","team":"a"}}`, name)
	case r.Method == http.MethodPatch && r.URL.Path == "/v1/"+name:
		fmt.Fprintf(w, `{"name":"projects/p/locations/us-east4/operations/op1","done":%v}`, f.opDone)
	default:
		http.Error(w, r.Method+" "+r.URL.Path, http.StatusNotFound)
	}
}

func TestOracleLabelsUpdate(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want map[string]string
	}{
		{"update", []string{"--update-labels", "env=prod,owner=b"}, map[string]string{"env": "prod", "team": "a", "owner": "b"}},
		{"remove", []string{"--remove-labels", "team"}, map[string]string{"env": "dev"}},
		{"clear", []string{"--clear-labels"}, nil},
		{"clear and update", []string{"--clear-labels", "--update-labels", "x=y"}, map[string]string{"x": "y"}},
		{"set", []string{"--labels", "only=1"}, map[string]string{"only": "1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := &odbLabelsFakeAPI{opDone: true}
			inv, stdout, stderr := newFakeAPIInvocation(t, api)
			args := append([]string{"oracle-database", "cloud-vm-clusters", "update", "vm1", "--location", "us-east4",
				"--project", "p", "--request-id", "r1", "--format", "value(name)"}, tc.args...)
			if err := runInvocation(inv, args...); err != nil {
				t.Fatal(err)
			}
			if len(api.requests) != 3 || !strings.HasPrefix(api.requests[1], "PATCH ") {
				t.Fatalf("requests = %v, want GET, PATCH, GET", api.requests)
			}
			if q := api.requests[1]; !strings.Contains(q, "updateMask=labels") || !strings.Contains(q, "requestId=r1") {
				t.Errorf("patch = %s, want updateMask=labels and requestId=r1", q)
			}
			var sent struct{ Labels map[string]string }
			if err := json.Unmarshal([]byte(api.bodies[1]), &sent); err != nil || !reflect.DeepEqual(sent.Labels, tc.want) {
				t.Errorf("patch body = %s, want labels %v", api.bodies[1], tc.want)
			}
			wantErr := "Request issued for: [vm1]\n" +
				"Waiting for operation [projects/p/locations/us-east4/operations/op1] to complete...done.\n" +
				"Updated cloudVmCluster [vm1].\n"
			if stderr.String() != wantErr {
				t.Errorf("stderr = %q, want %q", stderr, wantErr)
			}
			if stdout.String() != "projects/p/locations/us-east4/cloudVmClusters/vm1\n" {
				t.Errorf("stdout = %q, want the updated resource", stdout)
			}
		})
	}
}

func TestOracleLabelsUpdateAsync(t *testing.T) {
	api := &odbLabelsFakeAPI{}
	inv, stdout, stderr := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "oracle-database", "cloud-vm-clusters", "update",
		"projects/p/locations/us-east4/cloudVmClusters/vm1", "--async", "--update-labels", "a=b",
		"--format", "value(name)"); err != nil {
		t.Fatal(err)
	}
	if len(api.requests) != 2 || strings.Contains(api.requests[1], "requestId") {
		t.Errorf("requests = %v, want GET and PATCH without requestId", api.requests)
	}
	wantErr := "Request issued for: [vm1]\n" +
		"Check operation [projects/p/locations/us-east4/operations/op1] for status.\n" +
		"Updated cloudVmCluster [vm1].\n"
	if stderr.String() != wantErr {
		t.Errorf("stderr = %q, want %q", stderr, wantErr)
	}
	if stdout.String() != "projects/p/locations/us-east4/operations/op1\n" {
		t.Errorf("stdout = %q, want the operation", stdout)
	}
}

// The argument errors match gcloud-python 589.0.0's.
func TestOracleLabelsUpdateErrors(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, "Must specify at least one valid parameter to update."},
		{[]string{"--labels", "a=b", "--update-labels", "c=d"}, "argument --labels: At most one of --labels | --update-labels --clear-labels | --remove-labels can be specified."},
		{[]string{"--labels", "a=b", "--clear-labels"}, "argument --labels: At most one of --labels | --update-labels --clear-labels | --remove-labels can be specified."},
		{[]string{"--clear-labels", "--remove-labels", "a"}, "argument --clear-labels: At most one of --clear-labels | --remove-labels can be specified."},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			api := &odbLabelsFakeAPI{}
			inv, _, _ := newFakeAPIInvocation(t, api)
			args := append([]string{"oracle-database", "cloud-vm-clusters", "update", "vm1", "--location", "us-east4", "--project", "p"}, tc.args...)
			err := runInvocation(inv, args...)
			if err == nil || err.Error() != tc.want {
				t.Errorf("err = %v, want %q", err, tc.want)
			}
			if len(api.requests) != 0 {
				t.Errorf("requests = %v, want none", api.requests)
			}
		})
	}
}
