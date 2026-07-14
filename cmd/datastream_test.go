package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func datastreamSubgroup(name string) *cobra.Command {
	for _, c := range datastreamCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestDatastreamConnectionProfilesSubcommands(t *testing.T) {
	g := datastreamSubgroup("connection-profiles")
	if g == nil {
		t.Fatal("datastream connection-profiles missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "discover", "list", "update"})
}

func TestDatastreamLocationsSubcommands(t *testing.T) {
	g := datastreamSubgroup("locations")
	if g == nil {
		t.Fatal("datastream locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestDatastreamObjectsSubcommands(t *testing.T) {
	g := datastreamSubgroup("objects")
	if g == nil {
		t.Fatal("datastream objects missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "lookup", "start-backfill", "stop-backfill"})
}

func TestDatastreamOperationsSubcommands(t *testing.T) {
	g := datastreamSubgroup("operations")
	if g == nil {
		t.Fatal("datastream operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}

func TestDatastreamPrivateConnectionsSubcommands(t *testing.T) {
	g := datastreamSubgroup("private-connections")
	if g == nil {
		t.Fatal("datastream private-connections missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestDatastreamRoutesSubcommands(t *testing.T) {
	g := datastreamSubgroup("routes")
	if g == nil {
		t.Fatal("datastream routes missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestDatastreamStreamsSubcommands(t *testing.T) {
	g := datastreamSubgroup("streams")
	if g == nil {
		t.Fatal("datastream streams missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
