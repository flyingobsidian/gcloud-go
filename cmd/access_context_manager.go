package cmd

import "github.com/spf13/cobra"

// --- gcloud access-context-manager (#288) ---

var accessContextManagerCmd = &cobra.Command{Use: "access-context-manager", Short: "Manage Access Context Manager (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(accessContextManagerCmd, "authorized-orgs", "Manage authorized org descriptions", crud...)
	registerStubGroup(accessContextManagerCmd, "cloud-bindings", "Manage cloud access bindings", crud...)
	registerStubGroup(accessContextManagerCmd, "levels", "Manage access levels", append(crud, "conditions", "replace-all")...)
	registerStubGroup(accessContextManagerCmd, "perimeters", "Manage service perimeters", append(crud, "dry-run", "commit", "replace-all")...)
	registerStubGroup(accessContextManagerCmd, "policies", "Manage policies", crud...)
	registerStubGroup(accessContextManagerCmd, "supported-permissions", "VPC-SC supported permissions", "list")
	registerStubGroup(accessContextManagerCmd, "supported-services", "VPC-SC supported services", "list", "describe")
	rootCmd.AddCommand(accessContextManagerCmd)
}
