package cmd

import "testing"

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
