package cmd

import "testing"

func TestContainerSubnetsSubcommands(t *testing.T) {
	g := containerSubgroup("subnets")
	if g == nil {
		t.Fatal("container subnets missing")
	}
	assertSubcommands(t, g, []string{"list-usable"})
}
