package cmd

import "github.com/spf13/cobra"

// --- gcloud source (#385) ---

var sourceCmd = &cobra.Command{Use: "source", Short: "Manage Cloud Source Repositories (stubbed)"}

func init() {
	registerStubGroup(sourceCmd, "project-configs", "Manage project configuration", "describe", "update")
	registerStubGroup(sourceCmd, "repos", "Manage source repositories",
		"clone", "create", "delete", "describe", "list", "update", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")
	rootCmd.AddCommand(sourceCmd)
}
