package cmd

import "github.com/spf13/cobra"

// --- gcloud batch (#304) ---
//
// Stubs the batch command surface pending SDK integration
// (google.golang.org/api/batch/v1).

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Manage Batch jobs and tasks",
}

func init() {
	registerStubGroup(batchCmd, "jobs", "Manage Batch job resources",
		"cancel", "delete", "describe", "list", "submit")
	registerStubGroup(batchCmd, "tasks", "Manage Batch task resources",
		"describe", "list")
	rootCmd.AddCommand(batchCmd)
}
