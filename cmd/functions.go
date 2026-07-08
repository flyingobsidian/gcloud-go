package cmd

import "github.com/spf13/cobra"

// --- gcloud functions (#342) ---

var functionsCmd = &cobra.Command{Use: "functions", Short: "Manage Cloud Functions (stubbed)"}

func init() {
	for _, name := range []string{
		"add-invoker-policy-binding", "call", "delete", "deploy", "describe", "detach",
		"get-iam-policy", "list", "remove-iam-policy-binding", "remove-invoker-policy-binding",
		"set-iam-policy",
	} {
		registerStubCommand(functionsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(functionsCmd)
}
