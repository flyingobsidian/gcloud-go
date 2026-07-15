package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func bqSubgroup(name string) *cobra.Command {
	for _, c := range bqCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestBQHasMigrationWorkflows(t *testing.T) {
	if bqSubgroup("migration-workflows") == nil {
		t.Fatal("bq migration-workflows missing")
	}
}

func TestBQMigrationWorkflowsSubcommands(t *testing.T) {
	g := bqSubgroup("migration-workflows")
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "start"})
}

func TestBQMWName(t *testing.T) {
	got := bqMWName("wf1", "p", "us")
	want := "projects/p/locations/us/workflows/wf1"
	if got != want {
		t.Errorf("bqMWName = %q, want %q", got, want)
	}
	pass := "projects/x/locations/y/workflows/z"
	if bqMWName(pass, "ignored", "ignored") != pass {
		t.Errorf("bqMWName should pass through fully-qualified names")
	}
}
