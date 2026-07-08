package cmd

import "github.com/spf13/cobra"

// --- gcloud composer (#319) ---

var composerCmd = &cobra.Command{Use: "composer", Short: "Manage Cloud Composer (stubbed)"}

func init() {
	registerStubGroup(composerCmd, "environments", "Manage Composer environments",
		"create", "delete", "describe", "list", "update", "run",
		"restart-web-server", "storage", "check-upgrade", "save-snapshot", "load-snapshot")
	registerStubGroup(composerCmd, "operations", "Manage operations", "delete", "describe", "list")
	rootCmd.AddCommand(composerCmd)
}
