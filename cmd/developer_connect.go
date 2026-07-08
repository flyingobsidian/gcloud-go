package cmd

import "github.com/spf13/cobra"

// --- gcloud developer-connect (#330) ---

var developerConnectCmd = &cobra.Command{Use: "developer-connect", Short: "Manage Developer Connect (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(developerConnectCmd, "connections", "Manage connections", append(crud, "fetch-git-refs", "fetch-git-hashes", "fetch-linkable-git-repositories", "list-git-repositories")...)
	registerStubGroup(developerConnectCmd, "insights-configs", "Manage insights configs", crud...)
	registerStubGroup(developerConnectCmd, "operations", "Manage operations", "describe", "list", "cancel", "delete")
	rootCmd.AddCommand(developerConnectCmd)
}
