package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func schedulerSubgroup(name string) *cobra.Command {
	for _, c := range schedulerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestSchedulerCmekConfigSubcommands(t *testing.T) {
	g := schedulerSubgroup("cmek-config")
	if g == nil {
		t.Fatal("scheduler cmek-config missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestSchedulerLocationsSubcommands(t *testing.T) {
	g := schedulerSubgroup("locations")
	if g == nil {
		t.Fatal("scheduler locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestSchedulerOperationsSubcommands(t *testing.T) {
	g := schedulerSubgroup("operations")
	if g == nil {
		t.Fatal("scheduler operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}
