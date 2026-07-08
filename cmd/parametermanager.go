package cmd

import "github.com/spf13/cobra"

// --- gcloud parametermanager (#370) ---

var parameterManagerCmd = &cobra.Command{Use: "parametermanager", Short: "Manage Parameter Manager (stubbed)"}

func init() {
	registerStubGroup(parameterManagerCmd, "parameters", "Manage parameters",
		"create", "delete", "describe", "list", "update", "versions", "render", "access")
	rootCmd.AddCommand(parameterManagerCmd)
}
