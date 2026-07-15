package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func loggingSubgroup(name string) *cobra.Command {
	for _, c := range loggingCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

type loggingGroupCase struct {
	name string
	subs []string
}

var loggingCRUDSubcommands = []string{"create", "delete", "describe", "list", "update"}

func TestLoggingSubgroups(t *testing.T) {
	cases := []loggingGroupCase{
		{"buckets", append([]string{"undelete"}, loggingCRUDSubcommands...)},
		{"links", []string{"create", "delete", "describe", "list"}},
		{"locations", []string{"describe", "list"}},
		{"logs", []string{"delete", "list"}},
		{"metrics", loggingCRUDSubcommands},
		{"operations", []string{"cancel", "describe", "list"}},
		{"recent-queries", []string{"list"}},
		{"resource-descriptors", []string{"describe", "list"}},
		{"saved-queries", loggingCRUDSubcommands},
		{"scopes", loggingCRUDSubcommands},
		{"settings", []string{"describe", "update"}},
		{"sinks", loggingCRUDSubcommands},
		{"views", loggingCRUDSubcommands},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			g := loggingSubgroup(tc.name)
			if g == nil {
				t.Fatalf("logging %s missing", tc.name)
			}
			assertSubcommands(t, g, tc.subs)
		})
	}
}

func TestLoggingDataPlaneCommands(t *testing.T) {
	for _, name := range []string{"copy", "read", "write", "tail"} {
		if findSub(loggingCmd, name) == nil {
			t.Errorf("logging %s missing", name)
		}
	}
}
