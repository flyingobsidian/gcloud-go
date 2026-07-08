package cmd

import "github.com/spf13/cobra"

// --- gcloud billing (#309) ---

var billingCmd = &cobra.Command{Use: "billing", Short: "Manage billing (stubbed)"}

func init() {
	registerStubGroup(billingCmd, "accounts", "Manage billing accounts",
		"describe", "list", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")
	registerStubGroup(billingCmd, "budgets", "Manage budgets",
		"create", "delete", "describe", "list", "update")
	registerStubGroup(billingCmd, "projects", "Manage project billing configuration",
		"describe", "link", "unlink", "list")
	rootCmd.AddCommand(billingCmd)
}
