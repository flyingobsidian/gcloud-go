package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	aw "google.golang.org/api/assuredworkloads/v1"
)

func TestAssuredParent(t *testing.T) {
	flagAssuredOrg = "123"
	flagAssuredLocation = "us-central1"
	if got := assuredParent(); got != "organizations/123/locations/us-central1" {
		t.Errorf("assuredParent = %q", got)
	}
	flagAssuredOrg = "organizations/456"
	if got := assuredParent(); got != "organizations/456/locations/us-central1" {
		t.Errorf("assuredParent with prefix = %q", got)
	}
	flagAssuredOrg = ""
	flagAssuredLocation = ""
}

func TestAssuredWorkloadName(t *testing.T) {
	flagAssuredOrg = "123"
	flagAssuredLocation = "us-central1"
	t.Cleanup(func() { flagAssuredOrg = ""; flagAssuredLocation = "" })

	if got := assuredWorkloadName("abc"); got != "organizations/123/locations/us-central1/workloads/abc" {
		t.Errorf("assuredWorkloadName(abc) = %q", got)
	}
	full := "organizations/999/locations/foo/workloads/xyz"
	if got := assuredWorkloadName(full); got != full {
		t.Errorf("passthrough failed: %q", got)
	}
}

func TestAssuredOperationRegion(t *testing.T) {
	flagAssuredLocation = "europe-west2"
	t.Cleanup(func() { flagAssuredLocation = "" })

	if got := assuredOperationRegion("organizations/1/locations/us-central1/operations/op1"); got != "us-central1" {
		t.Errorf("assuredOperationRegion(full) = %q, want us-central1", got)
	}
	if got := assuredOperationRegion("op1"); got != "europe-west2" {
		t.Errorf("assuredOperationRegion(id) = %q, want europe-west2", got)
	}
}

// Verifies the v1 client targets REGION-assuredworkloads.googleapis.com and
// logs "Using endpoint [...]" to stderr, as gcloud-python does.
func TestAssuredV1ServiceRegionalEndpoint(t *testing.T) {
	creds := filepath.Join(t.TempDir(), "creds.json")
	if err := os.WriteFile(creds, []byte(`{"type":"authorized_user","client_id":"c","client_secret":"s","refresh_token":"r"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", creds)

	cases := []struct {
		region, want string
	}{
		{"us-central1", "https://us-central1-assuredworkloads.googleapis.com/"},
		{"", "https://assuredworkloads.googleapis.com/"},
	}
	for _, c := range cases {
		var svc *aw.Service
		stderr := captureAssuredStderr(t, func() {
			var err error
			svc, err = assuredV1Service(context.Background(), c.region)
			if err != nil {
				t.Fatalf("assuredV1Service(%q): %v", c.region, err)
			}
		})
		if svc.BasePath != c.want {
			t.Errorf("assuredV1Service(%q).BasePath = %q, want %q", c.region, svc.BasePath, c.want)
		}
		if want := "Using endpoint [" + c.want + "]\n"; stderr != want {
			t.Errorf("assuredV1Service(%q) stderr = %q, want %q", c.region, stderr, want)
		}
	}
}

// captureAssuredStderr runs fn while capturing everything it writes to os.Stderr.
func captureAssuredStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()
	fn()
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured stderr: %v", err)
	}
	return string(out)
}
