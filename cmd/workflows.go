package cmd

import "github.com/spf13/cobra"

// --- gcloud workflows (#397) ---

var workflowsCmd = &cobra.Command{Use: "workflows", Short: "Manage Cloud Workflows (stubbed)"}

func init() {
	registerStubGroup(workflowsCmd, "executions", "Manage workflow executions", "cancel", "describe", "list", "wait", "wait-last")
	for _, name := range []string{"delete", "deploy", "describe", "execute", "list", "run"} {
		registerStubCommand(workflowsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(workflowsCmd)
}
