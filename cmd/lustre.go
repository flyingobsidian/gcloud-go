package cmd

import "github.com/spf13/cobra"

// --- gcloud lustre (#352) ---

var lustreCmd = &cobra.Command{Use: "lustre", Short: "Manage Lustre (stubbed)"}

func init() {
	registerStubGroup(lustreCmd, "instances", "Manage Lustre instances", "create", "delete", "describe", "list", "update", "import-data", "export-data")
	registerStubGroup(lustreCmd, "operations", "Manage operations", "cancel", "describe", "list")
	rootCmd.AddCommand(lustreCmd)
}
