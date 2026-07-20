package cmd

import "github.com/spf13/cobra"

// --- gcloud oracle-database (#367) ---

var oracleDatabaseCmd = &cobra.Command{Use: "oracle-database", Short: "Manage Oracle Database"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(oracleDatabaseCmd, "backups", "Manage backups", crud...)
	registerStubGroup(oracleDatabaseCmd, "db-nodes", "Manage DB nodes", "describe", "list", "action")
	registerStubGroup(oracleDatabaseCmd, "db-servers", "Manage DB servers", "describe", "list")
	rootCmd.AddCommand(oracleDatabaseCmd)
}
