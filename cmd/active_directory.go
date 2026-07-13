package cmd

import "github.com/spf13/cobra"

// --- gcloud active-directory (#289) ---

var activeDirectoryCmd = &cobra.Command{Use: "active-directory", Short: "Manage Managed Microsoft AD"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(activeDirectoryCmd, "domains", "Manage AD domains", append(crud, "attach-trust", "detach-trust", "reset-managed-identities-admin-password", "restore", "backup", "sudoers")...)
	registerStubGroup(activeDirectoryCmd, "operations", "Manage AD operations", "cancel", "describe", "list")
	registerStubGroup(activeDirectoryCmd, "peerings", "Manage AD peerings", crud...)
	rootCmd.AddCommand(activeDirectoryCmd)
}
