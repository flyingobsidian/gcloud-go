package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func bigtableSubgroup(name string) *cobra.Command {
	for _, c := range bigtableCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestBigtableAppProfilesSubcommands(t *testing.T) {
	g := bigtableSubgroup("app-profiles")
	if g == nil {
		t.Fatal("bigtable app-profiles missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestBigtableAuthorizedViewsSubcommands(t *testing.T) {
	g := bigtableSubgroup("authorized-views")
	if g == nil {
		t.Fatal("bigtable authorized-views missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestBigtableBackupsSubcommands(t *testing.T) {
	g := bigtableSubgroup("backups")
	if g == nil {
		t.Fatal("bigtable backups missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update", "restore"})
}

func TestBigtableClustersSubcommands(t *testing.T) {
	g := bigtableSubgroup("clusters")
	if g == nil {
		t.Fatal("bigtable clusters missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestBigtableHotTabletsSubcommands(t *testing.T) {
	g := bigtableSubgroup("hot-tablets")
	if g == nil {
		t.Fatal("bigtable hot-tablets missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestBigtableInstancesSubcommands(t *testing.T) {
	g := bigtableSubgroup("instances")
	if g == nil {
		t.Fatal("bigtable instances missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding",
	})
}

func TestBigtableLogicalViewsSubcommands(t *testing.T) {
	g := bigtableSubgroup("logical-views")
	if g == nil {
		t.Fatal("bigtable logical-views missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
