package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func buildsSubgroup(name string) *cobra.Command {
	for _, c := range buildsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestBuildsConnectionsSubcommands(t *testing.T) {
	g := buildsSubgroup("connections")
	if g == nil {
		t.Fatal("builds connections missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "create", "delete", "describe",
		"get-iam-policy", "list", "remove-iam-policy-binding", "set-iam-policy", "update",
	})
}

func TestBuildsRepositoriesSubcommands(t *testing.T) {
	g := buildsSubgroup("repositories")
	if g == nil {
		t.Fatal("builds repositories missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestBuildsTriggersSubcommands(t *testing.T) {
	g := buildsSubgroup("triggers")
	if g == nil {
		t.Fatal("builds triggers missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "import", "list", "run", "update"})
}

func TestBuildsWorkerPoolsSubcommands(t *testing.T) {
	g := buildsSubgroup("worker-pools")
	if g == nil {
		t.Fatal("builds worker-pools missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
