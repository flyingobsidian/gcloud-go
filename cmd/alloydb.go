package cmd

import "github.com/spf13/cobra"

// --- gcloud alloydb (#293) ---

var alloydbCmd = &cobra.Command{Use: "alloydb", Short: "Manage AlloyDB"}

func init() {
	// All subgroups (backups, clusters, instances, operations, users) are
	// implemented in alloydb_all.go.
	rootCmd.AddCommand(alloydbCmd)
}
