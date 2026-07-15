package cmd

import "github.com/spf13/cobra"

// --- gcloud sql (#388) ---
//
// The top-level `sql` command aggregates the Cloud SQL subgroups. All
// subgroups (backups, connect, databases, export, flags, import, instances,
// operations, ssl, ssl-certs, tiers, users) are implemented in sql_all.go.

var sqlCmd = &cobra.Command{Use: "sql", Short: "Manage Cloud SQL"}

func init() {
	rootCmd.AddCommand(sqlCmd)
}
