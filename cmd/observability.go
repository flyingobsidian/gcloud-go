package cmd

import "github.com/spf13/cobra"

// --- gcloud observability (#366) ---

var observabilityCmd = &cobra.Command{Use: "observability", Short: "Manage Observability resources (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(observabilityCmd, "scopes", "Manage scopes", crud...)
	registerStubGroup(observabilityCmd, "trace-scopes", "Manage trace scopes", crud...)
	rootCmd.AddCommand(observabilityCmd)
}
