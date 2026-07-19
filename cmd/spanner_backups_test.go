package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func spannerSubgroup(name string) *cobra.Command {
	for _, c := range spannerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestSpannerBackupsSubcommands(t *testing.T) {
	g := spannerSubgroup("backups")
	if g == nil {
		t.Fatal("spanner backups missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "copy", "create", "delete", "describe",
		"get-iam-policy", "list", "remove-iam-policy-binding",
		"set-iam-policy", "update-metadata",
	})
}
