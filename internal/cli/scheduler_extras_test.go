package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func schedulerSubgroup(t *testing.T, name string) *cobra.Command {
	inv, _, _ := newTestInvocation(t, nil, "")
	for _, c := range newSchedulerCmd(inv).Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestSchedulerCmekConfigSubcommands(t *testing.T) {
	g := schedulerSubgroup(t, "cmek-config")
	if g == nil {
		t.Fatal("scheduler cmek-config missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestSchedulerLocationsSubcommands(t *testing.T) {
	g := schedulerSubgroup(t, "locations")
	if g == nil {
		t.Fatal("scheduler locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestSchedulerOperationsSubcommands(t *testing.T) {
	g := schedulerSubgroup(t, "operations")
	if g == nil {
		t.Fatal("scheduler operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}
