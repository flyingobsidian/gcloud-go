package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func edgeCloudSubgroup(name string) *cobra.Command {
	for _, c := range edgeCloudCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// TestEdgeCloudContainerSurface guards #1715: the `container` subgroup must
// match gcloud-python's shape -- no node-pools, no server-config group;
// instead a get-server-config leaf command and regions/zones stub subgroups.
func TestEdgeCloudContainerSurface(t *testing.T) {
	container := edgeCloudSubgroup("container")
	if container == nil {
		t.Fatal("edge-cloud container missing")
	}
	names := map[string]*cobra.Command{}
	for _, c := range container.Commands() {
		names[c.Name()] = c
	}

	// Must NOT contain removed extras.
	for _, gone := range []string{"node-pools", "server-config"} {
		if _, ok := names[gone]; ok {
			t.Errorf("edge-cloud container should NOT contain %q (not in gcloud-python)", gone)
		}
	}

	// Must expose these that were previously missing.
	for _, want := range []string{"get-server-config", "regions", "zones"} {
		if _, ok := names[want]; !ok {
			t.Errorf("edge-cloud container missing %q", want)
		}
	}

	// get-server-config must be a leaf (no subcommands), not a group.
	leaf := names["get-server-config"]
	if leaf != nil && leaf.HasSubCommands() {
		t.Errorf("get-server-config should be a leaf command, but has subcommands: %v", childNames(leaf))
	}
}

func childNames(c *cobra.Command) []string {
	var out []string
	for _, x := range c.Commands() {
		out = append(out, x.Name())
	}
	return out
}
