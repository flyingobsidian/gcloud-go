package cmd

import "github.com/spf13/cobra"

// --- gcloud compliance-manager (#317) ---

var complianceManagerCmd = &cobra.Command{Use: "compliance-manager", Short: "Manage Compliance Manager"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(complianceManagerCmd, "cloud-control-deployments", "Manage cloud control deployments", crud...)
	registerStubGroup(complianceManagerCmd, "cloud-controls", "Manage cloud controls", crud...)
	registerStubGroup(complianceManagerCmd, "framework-deployments", "Manage framework deployments", crud...)
	registerStubGroup(complianceManagerCmd, "frameworks", "Manage frameworks", crud...)
	registerStubGroup(complianceManagerCmd, "operations", "Manage operations", "describe", "list")
	rootCmd.AddCommand(complianceManagerCmd)
}
