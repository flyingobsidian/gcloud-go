package cmd

import "github.com/spf13/cobra"

// --- gcloud source-manager (#386) ---

var sourceManagerCmd = &cobra.Command{Use: "source-manager", Short: "Manage Secure Source Manager"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(sourceManagerCmd, "instances", "Manage instances", crud...)
	registerStubGroup(sourceManagerCmd, "locations", "Manage locations", "describe", "list")
	registerStubGroup(sourceManagerCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(sourceManagerCmd, "repos", "Manage repositories", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	rootCmd.AddCommand(sourceManagerCmd)
}
