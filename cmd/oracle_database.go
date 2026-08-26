package cmd

import "github.com/spf13/cobra"

// --- gcloud oracle-database (#367) ---
//
// The real subgroups (autonomous-databases, autonomous-database-backups,
// cloud-exadata-infrastructures, cloud-vm-clusters, databases, db-systems,
// exadb-vm-clusters, exascale-db-storage-vaults, odb-networks,
// pluggable-databases, goldengate-* etc.) are registered from their own
// oracle_database_*.go files.
//
// The earlier `backups`, `db-nodes`, and `db-servers` stubs were invented --
// gcloud-python's `oracle-database` surface has no such subgroups (the real
// backup group is `autonomous-database-backups`) -- and have been removed
// (#1721).

var oracleDatabaseCmd = &cobra.Command{Use: "oracle-database", Short: "Manage Oracle Database"}

func init() {
	rootCmd.AddCommand(oracleDatabaseCmd)
}
