package cmd

import "testing"

func TestContainerFleetSubgroups(t *testing.T) {
	g := containerSubgroup("fleet")
	if g == nil {
		t.Fatal("container fleet missing")
	}
	assertSubcommands(t, g, []string{"features", "fleets", "memberships"})
	for _, sub := range []string{"features", "fleets", "memberships"} {
		child := findSub(g, sub)
		if child == nil {
			t.Fatalf("container fleet %s missing", sub)
		}
		assertSubcommands(t, child, []string{"create", "delete", "describe", "list", "update"})
	}
}

func TestContainerHubSubgroups(t *testing.T) {
	g := containerSubgroup("hub")
	if g == nil {
		t.Fatal("container hub missing")
	}
	assertSubcommands(t, g, []string{"features", "fleets", "memberships"})
}
