package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func identitySubgroup(names ...string) *cobra.Command {
	var cur *cobra.Command = identityCmd
	for _, n := range names {
		next := findSub(cur, n)
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

func TestIdentityGroupsMembershipsSubcommands(t *testing.T) {
	g := identitySubgroup("groups", "memberships")
	if g == nil {
		t.Fatal("identity groups memberships missing")
	}
	assertSubcommands(t, g, []string{
		"add",
		"check-transitive-membership",
		"delete",
		"describe",
		"get-membership-graph",
		"list",
		"modify-membership-roles",
		"search-transitive-groups",
		"search-transitive-memberships",
	})
}

func TestIdentityGroupsMembershipsFlags(t *testing.T) {
	g := identitySubgroup("groups", "memberships")
	if g == nil {
		t.Fatal("identity groups memberships missing")
	}
	cases := map[string][]string{
		"add":                           {"group-email", "member-email", "roles"},
		"delete":                        {"group-email", "member-email"},
		"describe":                      {"group-email", "member-email"},
		"list":                          {"group-email"},
		"check-transitive-membership":   {"group-email", "member-email"},
		"get-membership-graph":          {"group-email", "query"},
		"modify-membership-roles":       {"group-email", "member-email", "add-roles", "remove-roles"},
		"search-transitive-groups":      {"member-email", "query"},
		"search-transitive-memberships": {"group-email"},
	}
	for name, flags := range cases {
		sub := findSub(g, name)
		if sub == nil {
			t.Errorf("subcommand %q missing", name)
			continue
		}
		for _, f := range flags {
			if sub.Flags().Lookup(f) == nil {
				t.Errorf("subcommand %q missing --%s", name, f)
			}
		}
	}
}
