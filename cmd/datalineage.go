package cmd

import "github.com/spf13/cobra"

// --- gcloud datalineage (#323) ---

var datalineageCmd = &cobra.Command{Use: "datalineage", Short: "Manage Data Lineage"}

func init() {
	registerStubGroup(datalineageCmd, "config", "Manage Data Lineage config", "describe", "update")
	registerStubGroup(datalineageCmd, "processes", "Manage Data Lineage processes", "create", "delete", "describe", "list", "update")
	rootCmd.AddCommand(datalineageCmd)
}
