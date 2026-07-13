package cmd

import "github.com/spf13/cobra"

// --- gcloud looker (#351) ---

var lookerCmd = &cobra.Command{Use: "looker", Short: "Manage Looker"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(lookerCmd, "backups", "Manage backups", crud...)
	registerStubGroup(lookerCmd, "instances", "Manage instances", append(crud, "import", "export", "restart", "restore")...)
	registerStubGroup(lookerCmd, "operations", "Manage operations", "cancel", "describe", "list")
	registerStubGroup(lookerCmd, "regions", "Manage regions", "list")
	rootCmd.AddCommand(lookerCmd)
}
