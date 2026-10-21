package cmd

import (
	"bytes"
	"strings"
	"testing"

	container "google.golang.org/api/container/v1"
)

func TestCtnClListLocation(t *testing.T) {
	defer func() { flagCtnClLocation, flagCtnClRegion, flagZone = "", "", "" }()
	cases := []struct {
		location, region, zone, want string
	}{
		{"", "", "", "-"},
		{"europe-west2", "", "", "europe-west2"},
		{"", "us-central1", "", "us-central1"},
		{"", "", "europe-west2-a", "europe-west2-a"},
	}
	for _, tc := range cases {
		flagCtnClLocation, flagCtnClRegion, flagZone = tc.location, tc.region, tc.zone
		if got := ctnClListLocation(); got != tc.want {
			t.Errorf("ctnClListLocation(%q, %q, %q) = %q, want %q", tc.location, tc.region, tc.zone, got, tc.want)
		}
	}
}

func TestContainerClustersListLocationOptional(t *testing.T) {
	f := containerClListCmd.Flags().Lookup("location")
	if f == nil {
		t.Fatal("clusters list missing --location")
	}
	if _, required := f.Annotations["cobra_annotation_bash_completion_one_required_flag"]; required {
		t.Error("clusters list --location should be optional")
	}
	if containerClListCmd.Flags().Lookup("region") == nil {
		t.Error("clusters list missing --region")
	}
}

func TestCtnSortClusters(t *testing.T) {
	got := ctnSortClusters([]*container.Cluster{
		{Name: "b", Zone: "europe-west2"},
		{Name: "a", Zone: "us-central1"},
		{Name: "a", Zone: "europe-west2"},
	})
	var order []string
	for _, c := range got {
		order = append(order, c.Zone+"/"+c.Name)
	}
	if want := "europe-west2/a europe-west2/b us-central1/a"; strings.Join(order, " ") != want {
		t.Errorf("order = %v, want %s", order, want)
	}
	if empty := ctnSortClusters(nil); empty == nil {
		t.Error("ctnSortClusters(nil) returned nil, want empty slice")
	}
}

func TestCtnClDefaultTable(t *testing.T) {
	clusters := []*container.Cluster{
		{
			Name: "prod", Zone: "europe-west2", CurrentMasterVersion: "1.33.1-gke.100", Endpoint: "10.0.0.1",
			CurrentNodeVersion: "1.33.1-gke.100", CurrentNodeCount: 3, Status: "RUNNING",
			NodePools: []*container.NodePool{{Config: &container.NodeConfig{MachineType: "e2-medium"}}},
			IpAllocationPolicy: &container.IPAllocationPolicy{StackType: "IPV4_IPV6"},
		},
		{
			Name: "empty", Zone: "us-central1-a", CurrentMasterVersion: "1.32.0", Endpoint: "10.0.0.2",
			CurrentNodeVersion: "1.32.0", InitialNodeCount: 1, Status: "PROVISIONING",
		},
	}
	var buf bytes.Buffer
	if err := emitFormattedTo(&buf, ctnClTableRows(clusters), ctnClListFormat); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	want := []string{
		"NAME LOCATION MASTER_VERSION MASTER_IP MACHINE_TYPE NODE_VERSION NUM_NODES STATUS STACK_TYPE",
		"prod europe-west2 1.33.1-gke.100 10.0.0.1 e2-medium 1.33.1-gke.100 3 RUNNING IPV4_IPV6",
		"empty us-central1-a 1.32.0 10.0.0.2 1 PROVISIONING IPV4",
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(lines), len(want), buf.String())
	}
	for i, w := range want {
		if got := strings.Join(strings.Fields(lines[i]), " "); got != w {
			t.Errorf("line %d = %q, want %q", i, got, w)
		}
	}
}
