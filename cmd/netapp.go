package cmd

import "github.com/spf13/cobra"

// --- gcloud netapp (#360) ---

var netappCmd = &cobra.Command{Use: "netapp", Short: "Manage Cloud NetApp Files (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(netappCmd, "active-directories", "Manage Active Directories", crud...)
	registerStubGroup(netappCmd, "backup-policies", "Manage backup policies", crud...)
	registerStubGroup(netappCmd, "backup-vaults", "Manage backup vaults", append(crud, "backups")...)
	registerStubGroup(netappCmd, "host-groups", "Manage host groups", crud...)
	registerStubGroup(netappCmd, "kms-configs", "Manage KMS configs", append(crud, "encrypt", "verify")...)
	registerStubGroup(netappCmd, "locations", "Get locations", "list", "describe")
	registerStubGroup(netappCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(netappCmd, "storage-pools", "Manage storage pools", append(crud, "switch")...)
	registerStubGroup(netappCmd, "volumes", "Manage volumes", append(crud, "revert", "replications", "snapshots")...)
	rootCmd.AddCommand(netappCmd)
}
