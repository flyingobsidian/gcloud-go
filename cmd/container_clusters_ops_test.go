package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func containerSubgroup(name string) *cobra.Command {
	for _, c := range containerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestContainerClustersSubcommands(t *testing.T) {
	g := containerSubgroup("clusters")
	if g == nil {
		t.Fatal("container clusters missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestContainerOperationsSubcommands(t *testing.T) {
	g := containerSubgroup("operations")
	if g == nil {
		t.Fatal("container operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list"})
}
