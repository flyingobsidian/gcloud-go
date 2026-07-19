package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func sourceSubgroup(name string) *cobra.Command {
	for _, c := range sourceCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestSourceProjectConfigsSubcommands(t *testing.T) {
	g := sourceSubgroup("project-configs")
	if g == nil {
		t.Fatal("source project-configs missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestSourceReposSubcommands(t *testing.T) {
	g := sourceSubgroup("repos")
	if g == nil {
		t.Fatal("source repos missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy",
	})
}
