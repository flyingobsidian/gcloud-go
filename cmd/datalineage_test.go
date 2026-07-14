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
