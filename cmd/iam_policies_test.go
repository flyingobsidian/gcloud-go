package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func iamSubgroup(name string) *cobra.Command {
	for _, c := range iamCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestIamPoliciesSubcommands(t *testing.T) {
	g := iamSubgroup("policies")
	if g == nil {
		t.Fatal("iam policies missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "get", "list", "update"})
}
