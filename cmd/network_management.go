package cmd

import "github.com/spf13/cobra"

// --- gcloud network-management (#362) ---

var networkManagementCmd = &cobra.Command{Use: "network-management", Short: "Manage Network Management (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(networkManagementCmd, "connectivity-tests", "Manage connectivity tests", append(crud, "rerun")...)
	registerStubGroup(networkManagementCmd, "network-monitoring-providers", "Manage network monitoring providers", "describe", "list", "enable", "disable")
	registerStubGroup(networkManagementCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(networkManagementCmd, "vpc-flow-logs-configs", "Manage VPC flow logs configs", crud...)
	rootCmd.AddCommand(networkManagementCmd)
}
