package cmd

import (
	"testing"

	compute "google.golang.org/api/compute/v1"
)

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
