package cmd

import "github.com/spf13/cobra"

// --- gcloud metastore (#356) ---

var metastoreCmd = &cobra.Command{Use: "metastore", Short: "Manage Dataproc Metastore"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(metastoreCmd, "federations", "Manage federations", crud...)
	registerStubGroup(metastoreCmd, "locations", "Manage locations", "list", "describe")
	registerStubGroup(metastoreCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list", "wait")
	registerStubGroup(metastoreCmd, "services", "Manage services", append(crud, "backups", "databases", "export-metadata", "import-metadata", "move-table-to-database", "query-metadata", "alter-location", "alter-table-properties", "restore")...)
	rootCmd.AddCommand(metastoreCmd)
}
