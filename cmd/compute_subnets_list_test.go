package cmd

import (
	"bytes"
	"strings"
	"testing"

	compute "google.golang.org/api/compute/v1"
)

func TestFilterSubnetsByNetwork(t *testing.T) {
	base := "https://www.googleapis.com/compute/v1/projects/p/global/networks/"
	subnets := []*compute.Subnetwork{
		{Name: "a", Network: base + "default"},
		{Name: "b", Network: base + "prod"},
		{Name: "c"},
	}
	names := func(s []*compute.Subnetwork) string {
		var n []string
		for _, x := range s {
			n = append(n, x.Name)
		}
		return strings.Join(n, ",")
	}
	if got := names(filterSubnetsByNetwork(subnets, "")); got != "a,b,c" {
		t.Errorf("no --network: %s", got)
	}
	if got := names(filterSubnetsByNetwork(subnets, "prod")); got != "b" {
		t.Errorf("--network prod: %s", got)
	}
	if got := filterSubnetsByNetwork(subnets, "none"); got == nil || len(got) != 0 {
		t.Errorf("--network none: got %v, want empty non-nil", got)
	}
}

func TestSubnetsDefaultTable(t *testing.T) {
	subnets := []*compute.Subnetwork{{
		Name: "default", Region: "https://x/projects/p/regions/europe-west2",
		Network: "https://x/projects/p/global/networks/default", IpCidrRange: "10.154.0.0/20", StackType: "IPV4_ONLY",
	}}
	var out bytes.Buffer
	if err := subnetsLister.emit(&out, &bytes.Buffer{}, subnets, "", false); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines:\n%s", len(lines), out.String())
	}
	if got := strings.Join(strings.Fields(lines[0]), " "); got != "NAME REGION NETWORK RANGE STACK_TYPE IPV6_ACCESS_TYPE INTERNAL_IPV6_PREFIX EXTERNAL_IPV6_PREFIX UTILIZATION_DETAILS" {
		t.Errorf("heading = %q", got)
	}
	if got := strings.Join(strings.Fields(lines[1]), " "); got != "default europe-west2 default 10.154.0.0/20 IPV4_ONLY" {
		t.Errorf("row = %q", got)
	}
}

func TestSubnetsListRegistered(t *testing.T) {
	list, _, err := computeCmd.Find([]string{"networks", "subnets", "list"})
	if err != nil || list.RunE == nil || list.Name() != "list" {
		t.Fatalf("compute networks subnets list not registered: %v", err)
	}
	for _, f := range []string{"regions", "network", "uri"} {
		if list.Flags().Lookup(f) == nil {
			t.Errorf("missing --%s", f)
		}
	}
}
