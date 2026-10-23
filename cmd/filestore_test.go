package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func filestoreSubgroup(name string) *cobra.Command {
	for _, c := range filestoreCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestFilestoreHasSubgroups(t *testing.T) {
	for _, want := range []string{"backups", "instances", "locations", "operations", "regions", "zones"} {
		if filestoreSubgroup(want) == nil {
			t.Errorf("filestore missing subgroup %q", want)
		}
	}
}

func TestFilestoreBackupsSubcommands(t *testing.T) {
	g := filestoreSubgroup("backups")
	if g == nil {
		t.Fatal("backups missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestFilestoreInstancesSubcommands(t *testing.T) {
	g := filestoreSubgroup("instances")
	if g == nil {
		t.Fatal("instances missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "restore", "revert", "update"})
}

func TestFilestoreLocationsSubcommands(t *testing.T) {
	g := filestoreSubgroup("locations")
	if g == nil {
		t.Fatal("locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestFilestoreOperationsSubcommands(t *testing.T) {
	g := filestoreSubgroup("operations")
	if g == nil {
		t.Fatal("operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list", "wait"})
}

func TestFilestoreRegionsSubcommands(t *testing.T) {
	g := filestoreSubgroup("regions")
	if g == nil {
		t.Fatal("regions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestFilestoreZonesSubcommands(t *testing.T) {
	g := filestoreSubgroup("zones")
	if g == nil {
		t.Fatal("zones missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestFilestoreResourceName(t *testing.T) {
	got := fsResourceName("inst1", "my-proj", "us-central1", "instances")
	want := "projects/my-proj/locations/us-central1/instances/inst1"
	if got != want {
		t.Errorf("fsResourceName = %q, want %q", got, want)
	}
	pass := "projects/x/locations/y/instances/z"
	if fsResourceName(pass, "ignored", "ignored", "instances") != pass {
		t.Errorf("fsResourceName should pass through fully-qualified names")
	}
}

func TestIsZoneLocationID(t *testing.T) {
	for id, want := range map[string]bool{
		"us-central1":   false,
		"us-central1-a": true,
		"europe-west4":  false,
		"asia-south1-b": true,
		"":              false,
	} {
		if got := isZoneLocationID(id); got != want {
			t.Errorf("isZoneLocationID(%q) = %v, want %v", id, got, want)
		}
	}
}

func TestFSParseNetwork(t *testing.T) {
	net, err := fsParseNetwork("name=default,connect-mode=PRIVATE_SERVICE_CONNECT,psc-requested-ip-address=10.0.0.5,address-mode=MODE_IPV4")
	if err != nil {
		t.Fatal(err)
	}
	if net.Network != "default" || net.ConnectMode != "PRIVATE_SERVICE_CONNECT" {
		t.Errorf("got network=%q connectMode=%q", net.Network, net.ConnectMode)
	}
	if net.PscConfig == nil || net.PscConfig.RequestedIpAddress != "10.0.0.5" || net.PscConfig.EndpointProject != "" {
		t.Errorf("pscConfig = %+v", net.PscConfig)
	}
	if len(net.Modes) != 1 || net.Modes[0] != "MODE_IPV4" {
		t.Errorf("modes = %v", net.Modes)
	}

	net, err = fsParseNetwork("name=vpc,reserved-ip-range=10.0.0.0/29")
	if err != nil {
		t.Fatal(err)
	}
	if net.ReservedIpRange != "10.0.0.0/29" || net.ConnectMode != "DIRECT_PEERING" || net.PscConfig != nil {
		t.Errorf("got %+v", net)
	}

	net, err = fsParseNetwork("default")
	if err != nil || net.Network != "default" {
		t.Errorf("bare name: %+v, %v", net, err)
	}

	for _, bad := range []string{"name=a,bogus=1", "name=a,connect-mode=NOPE", "name=a,address-mode=X", "name=a,junk"} {
		if _, err := fsParseNetwork(bad); err == nil {
			t.Errorf("fsParseNetwork(%q): expected error", bad)
		}
	}
}
