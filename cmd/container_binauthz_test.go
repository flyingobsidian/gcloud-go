package cmd

import "testing"

func TestContainerBinauthzSubgroups(t *testing.T) {
	g := containerSubgroup("binauthz")
	if g == nil {
		t.Fatal("container binauthz missing")
	}
	assertSubcommands(t, g, []string{"attestors", "policy"})

	policy := findSub(g, "policy")
	if policy == nil {
		t.Fatal("container binauthz policy missing")
	}
	assertSubcommands(t, policy, []string{"describe", "update"})

	attestors := findSub(g, "attestors")
	if attestors == nil {
		t.Fatal("container binauthz attestors missing")
	}
	assertSubcommands(t, attestors, []string{"create", "delete", "describe", "list", "update"})
}
