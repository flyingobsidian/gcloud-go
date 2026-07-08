package cmd

import "github.com/spf13/cobra"

// --- gcloud bq (#311) ---

var bqCmd = &cobra.Command{Use: "bq", Short: "Manage BigQuery Migration resources (stubbed)"}

func init() {
	registerStubGroup(bqCmd, "migration-workflows", "Manage BigQuery migration workflows",
		"create", "delete", "describe", "list", "start")
	rootCmd.AddCommand(bqCmd)
}
