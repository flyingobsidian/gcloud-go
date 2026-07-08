package cmd

import "github.com/spf13/cobra"

// --- gcloud agent-registry (#290) ---

var agentRegistryCmd = &cobra.Command{Use: "agent-registry", Short: "Manage Agent Registry (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(agentRegistryCmd, "agents", "Manage agents", crud...)
	registerStubGroup(agentRegistryCmd, "bindings", "Manage bindings", crud...)
	registerStubGroup(agentRegistryCmd, "endpoints", "Manage endpoints", crud...)
	registerStubGroup(agentRegistryCmd, "mcp-servers", "Manage MCP servers", crud...)
	registerStubGroup(agentRegistryCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(agentRegistryCmd, "services", "Manage services", crud...)
	rootCmd.AddCommand(agentRegistryCmd)
}
