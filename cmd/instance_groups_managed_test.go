package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	compute "google.golang.org/api/compute/v1"
)

// captureStdout runs fn while capturing everything it writes to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return string(out)
}

// Verifies the csv(...) format prints a lowercase heading row followed by data.
func TestFormatManagedInstancesCSVHeadings(t *testing.T) {
	instances := []*compute.ManagedInstance{
		{
			Instance:       "https://www.googleapis.com/compute/v1/projects/p/zones/my-zone/instances/my-vm",
			InstanceStatus: "RUNNING",
			CurrentAction:  "NONE",
		},
	}

	out := captureStdout(t, func() {
		if err := formatManagedInstances(instances, "csv(NAME,ZONE,STATUS)", true); err != nil {
			t.Fatalf("formatManagedInstances: %v", err)
		}
	})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected heading + 1 row, got %d lines: %q", len(lines), out)
	}
	if lines[0] != "name,zone,status" {
		t.Errorf("heading = %q, want %q", lines[0], "name,zone,status")
	}
	if lines[1] != "my-vm,my-zone,RUNNING" {
		t.Errorf("row = %q, want %q", lines[1], "my-vm,my-zone,RUNNING")
	}
}

// Covers the column extraction behind --format for managed list-instances,
// including the issue's csv(NAME,ZONE,STATUS) example.
func TestManagedInstanceField(t *testing.T) {
	mi := &compute.ManagedInstance{
		Instance:       "https://www.googleapis.com/compute/v1/projects/p/zones/some-zone/instances/my-vm",
		InstanceStatus: "RUNNING",
		CurrentAction:  "NONE",
		InstanceHealth: []*compute.ManagedInstanceInstanceHealth{
			{DetailedHealthState: "HEALTHY"},
		},
	}

	cases := map[string]string{
		"NAME":   "my-vm",
		"name":   "my-vm", // case-insensitive
		"ZONE":   "some-zone",
		"STATUS": "RUNNING",
		"ACTION": "NONE",
		"HEALTH": "HEALTHY",
		"BOGUS":  "",
	}
	for field, want := range cases {
		if got := managedInstanceField(mi, field); got != want {
			t.Errorf("managedInstanceField(%q) = %q, want %q", field, got, want)
		}
	}
}
