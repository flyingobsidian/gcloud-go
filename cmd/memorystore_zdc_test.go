package cmd

import (
	"reflect"
	"testing"
)

func TestApplyZoneDistributionZonesToMap(t *testing.T) {
	// Empty zones -> unchanged.
	b := map[string]any{}
	applyZoneDistributionZonesToMap(b, nil)
	if len(b) != 0 {
		t.Errorf("expected unchanged empty body, got %v", b)
	}
	// Fresh set.
	b = map[string]any{}
	applyZoneDistributionZonesToMap(b, []string{"us-central1-a", "us-central1-b"})
	zdc, ok := b["zoneDistributionConfig"].(map[string]any)
	if !ok {
		t.Fatalf("zoneDistributionConfig missing: %v", b)
	}
	got, ok := zdc["zones"].([]any)
	if !ok || len(got) != 2 || got[0] != "us-central1-a" || got[1] != "us-central1-b" {
		t.Errorf("zones = %v", got)
	}
	// Preserve existing mode.
	b = map[string]any{
		"zoneDistributionConfig": map[string]any{"mode": "MULTI_ZONE"},
	}
	applyZoneDistributionZonesToMap(b, []string{"us-east1-b"})
	zdc = b["zoneDistributionConfig"].(map[string]any)
	if zdc["mode"] != "MULTI_ZONE" {
		t.Errorf("mode was clobbered: %v", zdc)
	}
	if !reflect.DeepEqual(zdc["zones"], []any{"us-east1-b"}) {
		t.Errorf("zones = %v", zdc["zones"])
	}
}

func TestMemstoreInstCreateHasZoneDistributionFlag(t *testing.T) {
	c := findSub(findRootSub("memorystore"), "instances")
	if c == nil {
		t.Fatal("memorystore instances missing")
	}
	create := findSub(c, "create")
	if create == nil {
		t.Fatal("memorystore instances create missing")
	}
	if create.Flags().Lookup("zone-distribution-config-zones") == nil {
		t.Error("--zone-distribution-config-zones missing on memorystore instances create")
	}
}

func TestRedisClustersCreateHasZoneDistributionFlag(t *testing.T) {
	c := findSub(findRootSub("redis"), "clusters")
	if c == nil {
		t.Fatal("redis clusters missing")
	}
	create := findSub(c, "create")
	if create == nil {
		t.Fatal("redis clusters create missing")
	}
	if create.Flags().Lookup("zone-distribution-config-zones") == nil {
		t.Error("--zone-distribution-config-zones missing on redis clusters create")
	}
}
