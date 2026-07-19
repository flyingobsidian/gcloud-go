package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func workstationsSubgroup(name string) *cobra.Command {
	for _, c := range workstationsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestWorkstationsClustersSubcommands(t *testing.T) {
	g := workstationsSubgroup("clusters")
	if g == nil {
		t.Fatal("workstations clusters missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestWorkstationsConfigsSubcommands(t *testing.T) {
	g := workstationsSubgroup("configs")
	if g == nil {
		t.Fatal("workstations configs missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy",
	})
}
