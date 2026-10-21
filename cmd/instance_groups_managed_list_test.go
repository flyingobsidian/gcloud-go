package cmd

import (
	"bytes"
	"strings"
	"testing"

	compute "google.golang.org/api/compute/v1"
)

const migTestBase = "https://www.googleapis.com/compute/v1/projects/p/"

func migTestFixtures() ([]*compute.InstanceGroupManager, []*compute.InstanceGroup, []*compute.Autoscaler) {
	migs := []*compute.InstanceGroupManager{
		{
			Name:             "zonal",
			Zone:             migTestBase + "zones/europe-west2-a",
			SelfLink:         migTestBase + "zones/europe-west2-a/instanceGroupManagers/zonal",
			BaseInstanceName: "zonal",
			TargetSize:       2,
			InstanceTemplate: migTestBase + "global/instanceTemplates/tmpl-1",
		},
		{
			Name:             "regional",
			Region:           migTestBase + "regions/europe-west2",
			SelfLink:         migTestBase + "regions/europe-west2/instanceGroupManagers/regional",
			BaseInstanceName: "regional",
			InstanceTemplate: migTestBase + "regions/europe-west2/instanceTemplates/tmpl-2",
		},
		{
			Name:     "broken",
			Zone:     migTestBase + "zones/us-central1-a",
			SelfLink: migTestBase + "zones/us-central1-a/instanceGroupManagers/broken",
		},
	}
	groups := []*compute.InstanceGroup{
		{SelfLink: migTestBase + "zones/europe-west2-a/instanceGroups/zonal", Size: 2},
		{SelfLink: migTestBase + "regions/europe-west2/instanceGroups/regional", Size: 0},
	}
	autoscalers := []*compute.Autoscaler{
		{Name: "as-regional", Target: "https://compute.googleapis.com/compute/beta/projects/p/regions/europe-west2/instanceGroupManagers/regional", Status: "ACTIVE"},
		{Name: "as-broken", Target: migTestBase + "zones/us-central1-a/instanceGroupManagers/broken", Status: "ERROR"},
	}
	return migs, groups, autoscalers
}

func TestManagedListRecords(t *testing.T) {
	migs, groups, autoscalers := migTestFixtures()
	records, hadErrors, err := managedListRecords(migs, groups, autoscalers)
	if err != nil {
		t.Fatal(err)
	}
	if !hadErrors {
		t.Error("hadErrors = false, want true for an autoscaler in ERROR")
	}
	want := []struct {
		size, autoscaled string
		hasAutoscaler    bool
	}{
		{"2", "no", false},
		{"0", "yes", true},
		{"", "yes (*)", true},
	}
	if len(records) != len(want) {
		t.Fatalf("got %d records, want %d", len(records), len(want))
	}
	for i, w := range want {
		rec := records[i]
		if rec["size"] != w.size || rec["autoscaled"] != w.autoscaled {
			t.Errorf("%s: size=%v autoscaled=%v, want %q %q", rec["name"], rec["size"], rec["autoscaled"], w.size, w.autoscaled)
		}
		if _, ok := rec["autoscaler"]; ok != w.hasAutoscaler {
			t.Errorf("%s: has autoscaler = %v, want %v", rec["name"], ok, w.hasAutoscaler)
		}
	}
}

func TestManagedListRecordsEmpty(t *testing.T) {
	records, hadErrors, err := managedListRecords(nil, nil, nil)
	if err != nil || hadErrors || records == nil || len(records) != 0 {
		t.Errorf("got %v, %v, %v; want empty non-nil records", records, hadErrors, err)
	}
}

func TestManagedListDefaultTable(t *testing.T) {
	migs, groups, autoscalers := migTestFixtures()
	records, _, err := managedListRecords(migs, groups, autoscalers)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := emitFormattedTo(&buf, managedListTableRows(records), managedListFormat); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	want := [][]string{
		{"NAME", "LOCATION", "SCOPE", "BASE_INSTANCE_NAME", "SIZE", "TARGET_SIZE", "INSTANCE_TEMPLATE", "AUTOSCALED"},
		{"zonal", "europe-west2-a", "zone", "zonal", "2", "2", "tmpl-1", "no"},
		{"regional", "europe-west2", "region", "regional", "0", "0", "tmpl-2", "yes"},
		{"broken", "us-central1-a", "zone", "0", "yes", "(*)"},
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(lines), len(want), buf.String())
	}
	for i, w := range want {
		if got := strings.Fields(lines[i]); strings.Join(got, " ") != strings.Join(w, " ") {
			t.Errorf("line %d = %q, want fields %q", i, lines[i], w)
		}
	}
}

func TestManagedListScope(t *testing.T) {
	zonal := &compute.InstanceGroupManager{Zone: "z"}
	regional := &compute.InstanceGroupManager{Region: "r"}
	cases := []struct {
		name               string
		scope              managedListScope
		keepZonal, keepReg bool
	}{
		{"all", managedListScope{}, true, true},
		{"zones only", managedListScope{zonalOnly: true}, true, false},
		{"regions only", managedListScope{regionalOnly: true}, false, true},
	}
	for _, tc := range cases {
		if got := tc.scope.keep(zonal); got != tc.keepZonal {
			t.Errorf("%s: keep(zonal) = %v", tc.name, got)
		}
		if got := tc.scope.keep(regional); got != tc.keepReg {
			t.Errorf("%s: keep(regional) = %v", tc.name, got)
		}
	}
}

func TestManagedListScopeFromFlags(t *testing.T) {
	defer func() {
		flagManagedListZones, flagManagedListRegions, flagZone = nil, nil, ""
		for _, name := range []string{"zones", "regions"} {
			managedListCmd.Flags().Lookup(name).Changed = false
		}
	}()

	if got := managedListScopeFromFlags(managedListCmd); got.zonalOnly || got.regionalOnly || len(got.zones)+len(got.regions) != 0 {
		t.Errorf("no flags: got %+v, want aggregate everything", got)
	}

	flagZone = "europe-west2-b"
	if got := managedListScopeFromFlags(managedListCmd); !got.zonalOnly || len(got.zones) != 1 || got.zones[0] != "europe-west2-b" {
		t.Errorf("--zone: got %+v", got)
	}
	flagZone = ""

	if err := managedListCmd.Flags().Set("regions", "europe-west2,us-central1"); err != nil {
		t.Fatal(err)
	}
	if got := managedListScopeFromFlags(managedListCmd); !got.regionalOnly || len(got.regions) != 2 {
		t.Errorf("--regions: got %+v", got)
	}
}

func TestComputeResourcePath(t *testing.T) {
	got := computeResourcePath("https://compute.googleapis.com/compute/beta/projects/p/zones/z/instanceGroupManagers/m")
	if want := "projects/p/zones/z/instanceGroupManagers/m"; got != want {
		t.Errorf("computeResourcePath = %q, want %q", got, want)
	}
}
