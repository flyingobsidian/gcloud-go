package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func memcacheSubgroup(name string) *cobra.Command {
	for _, c := range memcacheCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestMemcacheInstancesSubcommands(t *testing.T) {
	g := memcacheSubgroup("instances")
	if g == nil {
		t.Fatal("memcache instances missing")
	}
	assertSubcommands(t, g, []string{
		"apply-parameters", "create", "delete", "describe", "list", "update", "upgrade",
	})
}

func TestMemcacheRegionsSubcommands(t *testing.T) {
	g := memcacheSubgroup("regions")
	if g == nil {
		t.Fatal("memcache regions missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestMemcacheOperationsSubcommands(t *testing.T) {
	g := memcacheSubgroup("operations")
	if g == nil {
		t.Fatal("memcache operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list"})
}
