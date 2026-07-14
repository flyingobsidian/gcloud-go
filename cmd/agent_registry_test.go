package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func arSubgroup(name string) *cobra.Command {
	for _, c := range agentRegistryCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestAgentRegistryAgentsSubcommands(t *testing.T) {
	g := arSubgroup("agents")
	if g == nil {
		t.Fatal("agent-registry agents missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "search"})
}

func TestAgentRegistryBindingsSubcommands(t *testing.T) {
	g := arSubgroup("bindings")
	if g == nil {
		t.Fatal("agent-registry bindings missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "fetch-available", "list", "update"})
}

func TestAgentRegistryEndpointsSubcommands(t *testing.T) {
	g := arSubgroup("endpoints")
	if g == nil {
		t.Fatal("agent-registry endpoints missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestAgentRegistryMcpServersSubcommands(t *testing.T) {
	g := arSubgroup("mcp-servers")
	if g == nil {
		t.Fatal("agent-registry mcp-servers missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "search"})
}

func TestAgentRegistryOperationsSubcommands(t *testing.T) {
	g := arSubgroup("operations")
	if g == nil {
		t.Fatal("agent-registry operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list", "wait"})
}

func TestAgentRegistryServicesSubcommands(t *testing.T) {
	g := arSubgroup("services")
	if g == nil {
		t.Fatal("agent-registry services missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
