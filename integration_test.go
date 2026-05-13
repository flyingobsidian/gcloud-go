//go:build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Integration tests require:
//   - A built gcloud-go binary in the current directory
//   - GCLOUD_TEST_PROJECT set to a GCP project
//   - GCLOUD_TEST_ZONE set to a compute zone
//   - GCLOUD_TEST_CRED_FILE set to a service account JSON key file
//   - GCLOUD_TEST_INSTANCE set to an existing instance name (for ssh/stop/start)
//   - GCLOUD_TEST_INSTANCE_GROUP set (optional, for instance group tests)
//
// Run with: go test -tags=integration -v

var (
	binary    = "./gcloud-go"
	project   string
	zone      string
	credFile  string
	instance  string
	instGroup string
)

func TestMain(m *testing.M) {
	project = os.Getenv("GCLOUD_TEST_PROJECT")
	zone = os.Getenv("GCLOUD_TEST_ZONE")
	credFile = os.Getenv("GCLOUD_TEST_CRED_FILE")
	instance = os.Getenv("GCLOUD_TEST_INSTANCE")
	instGroup = os.Getenv("GCLOUD_TEST_INSTANCE_GROUP")

	os.Exit(m.Run())
}

func skipIfMissing(t *testing.T, vars ...string) {
	t.Helper()
	for _, v := range vars {
		if os.Getenv(v) == "" {
			t.Skipf("skipping: %s not set", v)
		}
	}
}

func run(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command(binary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command %v failed: %v\noutput: %s", args, err, out)
	}
	return string(out)
}

func TestIntegration_AuthLogin(t *testing.T) {
	skipIfMissing(t, "GCLOUD_TEST_CRED_FILE")

	out := run(t, "auth", "login", "--cred-file", credFile)
	if !strings.Contains(out, "Activated service account") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestIntegration_CreateCredConfig(t *testing.T) {
	outFile := t.TempDir() + "/cred-config.json"

	run(t, "iam", "workload-identity-pools", "create-cred-config",
		"//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/pool/providers/prov",
		"--output-file", outFile,
		"--credential-source-file", "/var/run/token",
		"--service-account", "sa@project.iam.gserviceaccount.com",
	)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg["type"] != "external_account" {
		t.Errorf("type = %v", cfg["type"])
	}
}

func TestIntegration_InstancesStop(t *testing.T) {
	skipIfMissing(t, "GCLOUD_TEST_PROJECT", "GCLOUD_TEST_ZONE", "GCLOUD_TEST_CRED_FILE", "GCLOUD_TEST_INSTANCE")

	// First ensure we're authenticated.
	run(t, "auth", "login", "--cred-file", credFile)

	out := run(t, "compute", "instances", "stop", instance,
		"--project", project, "--zone", zone)
	if !strings.Contains(out, "Stopped") && !strings.Contains(out, "Stopping") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestIntegration_InstancesStart(t *testing.T) {
	skipIfMissing(t, "GCLOUD_TEST_PROJECT", "GCLOUD_TEST_ZONE", "GCLOUD_TEST_CRED_FILE", "GCLOUD_TEST_INSTANCE")

	run(t, "auth", "login", "--cred-file", credFile)

	out := run(t, "compute", "instances", "start", instance,
		"--project", project, "--zone", zone)
	if !strings.Contains(out, "Started") && !strings.Contains(out, "Starting") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestIntegration_UnmanagedListInstances(t *testing.T) {
	skipIfMissing(t, "GCLOUD_TEST_PROJECT", "GCLOUD_TEST_ZONE", "GCLOUD_TEST_CRED_FILE", "GCLOUD_TEST_INSTANCE_GROUP")

	run(t, "auth", "login", "--cred-file", credFile)

	out := run(t, "compute", "instance-groups", "unmanaged", "list-instances",
		instGroup, "--project", project, "--zone", zone)
	if !strings.Contains(out, "NAME") && !strings.Contains(out, "No instances") {
		t.Errorf("unexpected output: %s", out)
	}
}
