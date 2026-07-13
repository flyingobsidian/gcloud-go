package cmd

import "github.com/spf13/cobra"

// --- gcloud filestore (#339) ---

var filestoreCmd = &cobra.Command{Use: "filestore", Short: "Manage Filestore"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(filestoreCmd, "backups", "Manage backups", crud...)
	registerStubGroup(filestoreCmd, "instances", "Manage instances", append(crud, "restore", "snapshots")...)
	registerStubGroup(filestoreCmd, "locations", "List locations", "list", "describe")
	registerStubGroup(filestoreCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(filestoreCmd, "regions", "List regions", "list", "describe")
	registerStubGroup(filestoreCmd, "zones", "List zones", "list", "describe")
	rootCmd.AddCommand(filestoreCmd)
}
