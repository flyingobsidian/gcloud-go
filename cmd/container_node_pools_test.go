package cmd

import "testing"

func TestContainerNodePoolsSubcommands(t *testing.T) {
	g := containerSubgroup("node-pools")
	if g == nil {
		t.Fatal("container node-pools missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update", "rollback",
	})
}
