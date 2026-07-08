package cmd

import "github.com/spf13/cobra"

// --- gcloud notebooks (#365) ---

var notebooksCmd = &cobra.Command{Use: "notebooks", Short: "Manage AI Platform Notebooks (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(notebooksCmd, "environments", "Manage notebook environments", "create", "delete", "describe", "list")
	registerStubGroup(notebooksCmd, "instances", "Manage notebook instances", append(crud, "start", "stop", "reset", "register", "diagnose", "upgrade", "check-upgradability", "get-health", "set-accelerator", "set-labels", "set-machine-type", "add-metadata", "rollback", "migrate")...)
	registerStubGroup(notebooksCmd, "locations", "View locations", "list", "describe")
	registerStubGroup(notebooksCmd, "runtimes", "Manage runtimes", append(crud, "start", "stop", "reset", "migrate", "diagnose", "refresh-runtime-token-internal", "upgrade")...)
	rootCmd.AddCommand(notebooksCmd)
}
