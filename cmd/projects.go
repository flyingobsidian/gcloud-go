package cmd

import "github.com/spf13/cobra"

// --- gcloud projects (#375) ---

var projectsCmd = &cobra.Command{Use: "projects", Short: "Manage projects (stubbed)"}

func init() {
	for _, name := range []string{
		"add-iam-policy-binding", "create", "delete", "describe", "get-ancestors",
		"get-ancestors-iam-policy", "get-iam-policy", "list", "move",
		"remove-iam-policy-binding", "set-iam-policy", "undelete", "update",
	} {
		registerStubCommand(projectsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(projectsCmd)
}
