package cmd

import "github.com/spf13/cobra"

// --- gcloud workbench (#396) ---

var workbenchCmd = &cobra.Command{Use: "workbench", Short: "Manage Vertex AI Workbench"}

func init() {
	// Subgroups instances, executions, and schedules are implemented in
	// workbench_instances.go and workbench_executions_schedules.go.
	rootCmd.AddCommand(workbenchCmd)
}
