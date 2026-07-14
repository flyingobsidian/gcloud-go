package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func lookerSubgroup(name string) *cobra.Command {
	for _, c := range lookerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestLookerRegionsSubcommands(t *testing.T) {
	g := lookerSubgroup("regions")
	if g == nil {
		t.Fatal("looker regions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestLookerOperationsSubcommands(t *testing.T) {
	g := lookerSubgroup("operations")
	if g == nil {
		t.Fatal("looker operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list"})
}

func TestLookerInstancesSubcommands(t *testing.T) {
	g := lookerSubgroup("instances")
	if g == nil {
		t.Fatal("looker instances missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "export", "import", "list", "restart", "restore", "update",
	})
}

func TestLookerBackupsSubcommands(t *testing.T) {
	g := lookerSubgroup("backups")
	if g == nil {
		t.Fatal("looker backups missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}
