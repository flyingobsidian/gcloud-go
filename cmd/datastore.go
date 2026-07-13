package cmd

import "github.com/spf13/cobra"

// --- gcloud datastore (#325) ---

var datastoreCmd = &cobra.Command{Use: "datastore", Short: "Manage Cloud Datastore"}

func init() {
	registerStubGroup(datastoreCmd, "indexes", "Manage Datastore indexes", "create", "list", "cleanup")
	registerStubGroup(datastoreCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	for _, name := range []string{"export", "import"} {
		registerStubCommand(datastoreCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(datastoreCmd)
}
