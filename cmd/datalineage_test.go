package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func datalineageSubgroup(name string) *cobra.Command {
	for _, c := range datalineageCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestDatalineageConfigSubcommands(t *testing.T) {
	g := datalineageSubgroup("config")
	if g == nil {
		t.Fatal("datalineage config missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestDatalineageProcessesSubcommands(t *testing.T) {
	g := datalineageSubgroup("processes")
	if g == nil {
		t.Fatal("datalineage processes missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestDatalineageRunsSubcommands(t *testing.T) {
	g := datalineageSubgroup("runs")
	if g == nil {
		t.Fatal("datalineage runs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestDatalineageLineageEventsSubcommands(t *testing.T) {
	g := datalineageSubgroup("lineage-events")
	if g == nil {
		t.Fatal("datalineage lineage-events missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}
