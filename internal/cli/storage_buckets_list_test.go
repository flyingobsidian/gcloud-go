package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// storageBucketsListAPI serves one bucket with the fields gcloud-python's
// display record reformats.
func storageBucketsListAPI(t *testing.T, seen *[]string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = append(*seen, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[{
			"kind":"storage#bucket","id":"b1","name":"b1","projectNumber":"123",
			"selfLink":"https://www.googleapis.com/storage/v1/b/b1",
			"timeCreated":"2021-01-01T01:01:01.123Z","updated":"2021-01-02T01:01:01Z",
			"generation":"1234567890123456789","metageneration":"1",
			"location":"EU","locationType":"multi-region","storageClass":"STANDARD","rpo":"DEFAULT",
			"iamConfiguration":{"uniformBucketLevelAccess":{"enabled":false},"publicAccessPrevention":"inherited"},
			"softDeletePolicy":{"retentionDurationSeconds":"604800","effectiveTime":"2021-01-01T01:01:01.804Z"},
			"labels":{"env":"prod"},
			"cors":[{"origin":["*.example.com"],"method":["GET"]}],
			"lifecycle":{"rule":[{"action":{"type":"Delete"},"condition":{"age":1,"isLive":false}}]},
			"acl":[{"kind":"storage#bucketAccessControl","id":"b1/project-owners-123","selfLink":"x","bucket":"b1",
				"entity":"project-owners-123","role":"OWNER","etag":"CAE=","projectTeam":{"projectNumber":"123","team":"owners"}}]
		}]}`)
	})
}

// The default output matches gcloud-python 589.0.0's display record and
// yaml style.
func TestStorageBucketsListDefaultYAML(t *testing.T) {
	var seen []string
	inv, stdout, _ := newFakeAPIInvocation(t, storageBucketsListAPI(t, &seen))
	if err := runInvocation(inv, "storage", "buckets", "list", "--project=p"); err != nil {
		t.Fatal(err)
	}
	want := `---
acl:
- entity: project-owners-123
  projectTeam:
    projectNumber: '123'
    team: owners
  role: OWNER
cors_config:
- method:
  - GET
  origin:
  - '*.example.com'
creation_time: 2021-01-01T01:01:01+0000
default_storage_class: STANDARD
generation: 1234567890123456789
labels:
  env: prod
lifecycle_config:
  rule:
  - action:
      type: Delete
    condition:
      age: 1
      isLive: false
location: EU
location_type: multi-region
metageneration: 1
name: b1
public_access_prevention: inherited
rpo: DEFAULT
soft_delete_policy:
  effectiveTime: '2021-01-01T01:01:01.804000+00:00'
  retentionDurationSeconds: '604800'
storage_url: gs://b1/
uniform_bucket_level_access: false
update_time: 2021-01-02T01:01:01+0000
`
	if stdout.String() != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
	if len(seen) != 1 || !strings.Contains(seen[0], "projection=full") {
		t.Errorf("queries = %v, want projection=full", seen)
	}
}

func TestStorageBucketsListFormats(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--uri"}, "https://storage.googleapis.com/storage/v1/b/b1\n"},
		{[]string{"--format=value(name,creation_time)"}, "b1\t2021-01-01T01:01:01+0000\n"},
		{[]string{"--filter=location=EU", "--format=value(name)"}, "b1\n"},
		{[]string{"--filter=B1", "--format=value(name)"}, "b1\n"},
		{[]string{"--filter=location=US", "--format=value(name)"}, ""},
		{[]string{"--raw", "--format=value(storageClass,timeCreated)"}, "STANDARD\t2021-01-01T01:01:01.123000+00:00\n"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var seen []string
			inv, stdout, stderr := newFakeAPIInvocation(t, storageBucketsListAPI(t, &seen))
			if err := runInvocation(inv, append([]string{"storage", "buckets", "list", "--project=p"}, tc.args...)...); err != nil {
				t.Fatal(err)
			}
			if stdout.String() != tc.want {
				t.Errorf("stdout = %q, want %q", stdout, tc.want)
			}
			if stderr.Len() != 0 {
				t.Errorf("stderr = %q, want none", stderr)
			}
		})
	}
}

func TestStorageBucketsListEmpty(t *testing.T) {
	empty := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{}`)
	})
	inv, stdout, stderr := newFakeAPIInvocation(t, empty)
	if err := runInvocation(inv, "storage", "buckets", "list", "--project=p"); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 || stderr.String() != "Listed 0 items.\n" {
		t.Errorf("stdout = %q, stderr = %q, want only Listed 0 items.", stdout, stderr)
	}
	inv, stdout, _ = newFakeAPIInvocation(t, empty)
	if err := runInvocation(inv, "storage", "buckets", "list", "--project=p", "--format=json"); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "[]\n" {
		t.Errorf("--format=json stdout = %q, want []", stdout)
	}
}

// The quoting matches gcloud-python 589.0.0's yaml printer (checked with
// resource_printer.Print on the same values).
func TestEmitPyYAMLDocumentsQuoting(t *testing.T) {
	var buf bytes.Buffer
	err := emitPyYAMLDocuments(&buf, []map[string]any{{
		"a": "plain text", "b": "123", "c": "true", "d": "", "e": "it's", "f": "x: y",
		"g": "2021-01-01T00:00:00+00:00", "h": "2021-01-01T00:00:00+0000", "i": "-x", "j": "- x",
		"k": map[string]any{}, "l": []any{}, "m": nil, "n": "https://a.b/c:d",
		"o": "yes", "p": "0x1F", "q": "'x", "r": "1_000", "s": "?x", "t": "? x",
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := "---\na: plain text\nb: '123'\nc: 'true'\nd: ''\ne: it's\nf: 'x: y'\n" +
		"g: '2021-01-01T00:00:00+00:00'\nh: 2021-01-01T00:00:00+0000\ni: -x\nj: '- x'\n" +
		"k: {}\nl: []\nm: null\nn: https://a.b/c:d\no: yes\np: '0x1F'\nq: \"'x\"\nr: '1_000'\ns: ?x\nt: '? x'\n"
	if buf.String() != want {
		t.Errorf("got\n%s\nwant\n%s", buf.String(), want)
	}
}
