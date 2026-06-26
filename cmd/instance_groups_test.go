package cmd

import (
	"testing"

	compute "google.golang.org/api/compute/v1"
)

func unmanagedItem(name, status string) *compute.InstanceWithNamedPorts {
	return &compute.InstanceWithNamedPorts{
		Instance: "https://www.googleapis.com/compute/v1/projects/p/zones/some-zone/instances/" + name,
		Status:   status,
	}
}

func TestUnmanagedInstanceField(t *testing.T) {
	it := unmanagedItem("my-vm", "RUNNING")
	cases := map[string]string{
		"NAME":   "my-vm",
		"name":   "my-vm",
		"STATUS": "RUNNING",
		"ZONE":   "some-zone",
		"BOGUS":  "",
	}
	for field, want := range cases {
		if got := unmanagedInstanceField(it, field); got != want {
			t.Errorf("unmanagedInstanceField(%q) = %q, want %q", field, got, want)
		}
	}
}

func TestFilterUnmanagedInstances(t *testing.T) {
	items := []*compute.InstanceWithNamedPorts{
		unmanagedItem("web-1", "RUNNING"),
		unmanagedItem("db-1", "RUNNING"),
	}

	if got := filterUnmanagedInstances(items, ""); len(got) != 2 {
		t.Errorf("empty filter should keep all, got %d", len(got))
	}

	got := filterUnmanagedInstances(items, "WEB") // case-insensitive substring
	if len(got) != 1 || unmanagedInstanceField(got[0], "NAME") != "web-1" {
		t.Errorf("filter WEB = %v, want [web-1]", got)
	}
}

// Reproduces the issue: --filter SOME_VM_NAME --format 'csv(NAME)'.
func TestUnmanagedListInstancesFilterAndCSV(t *testing.T) {
	items := []*compute.InstanceWithNamedPorts{
		unmanagedItem("SOME_VM_NAME", "RUNNING"),
		unmanagedItem("OTHER_VM", "RUNNING"),
	}

	filtered := filterUnmanagedInstances(items, "SOME_VM_NAME")
	out := captureStdout(t, func() {
		if err := formatUnmanagedInstances(filtered, "csv(NAME)"); err != nil {
			t.Fatalf("formatUnmanagedInstances: %v", err)
		}
	})

	if out != "name\nSOME_VM_NAME\n" {
		t.Errorf("output = %q, want %q", out, "name\nSOME_VM_NAME\n")
	}
}
