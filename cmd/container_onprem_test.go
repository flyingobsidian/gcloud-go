package cmd

import "testing"

func TestContainerBareMetalSubgroups(t *testing.T) {
	g := containerSubgroup("bare-metal")
	if g == nil {
		t.Fatal("container bare-metal missing")
	}
	assertSubcommands(t, g, []string{"clusters"})
	clusters := findSub(g, "clusters")
	if clusters == nil {
		t.Fatal("container bare-metal clusters missing")
	}
	assertSubcommands(t, clusters, []string{"create", "delete", "describe", "list", "update"})
}

func TestContainerVmwareSubgroups(t *testing.T) {
	g := containerSubgroup("vmware")
	if g == nil {
		t.Fatal("container vmware missing")
	}
	assertSubcommands(t, g, []string{"clusters"})
	clusters := findSub(g, "clusters")
	if clusters == nil {
		t.Fatal("container vmware clusters missing")
	}
	assertSubcommands(t, clusters, []string{"create", "delete", "describe", "list", "update"})
}
