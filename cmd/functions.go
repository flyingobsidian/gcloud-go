package cmd

import "github.com/spf13/cobra"

// --- gcloud functions (#342) ---

var functionsCmd = &cobra.Command{Use: "functions", Short: "Manage Cloud Functions"}

func init() {
	for _, name := range []string{
		"add-invoker-policy-binding", "call", "delete", "deploy", "describe", "detach",
		"get-iam-policy", "list", "remove-iam-policy-binding", "remove-invoker-policy-binding",
		"set-iam-policy",
	} {
		registerStubCommand(functionsCmd, name, "Not yet implemented")
	}
		registerStubGroup(functionsCmd, "add-iam-policy-binding", "Manage add-iam-policy-binding", "list", "describe")
	registerStubGroup(functionsCmd, "event-types", "Manage event-types", "list", "describe")
	registerStubGroup(functionsCmd, "logs", "Manage logs", "list", "describe")
	registerStubGroup(functionsCmd, "regions", "Manage regions", "list", "describe")
	registerStubGroup(functionsCmd, "runtimes", "Manage runtimes", "list", "describe")
	rootCmd.AddCommand(functionsCmd)
}
