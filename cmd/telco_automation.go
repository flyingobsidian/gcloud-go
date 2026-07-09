package cmd

import "github.com/spf13/cobra"

// --- gcloud telco-automation (#390) ---

var telcoAutomationCmd = &cobra.Command{Use: "telco-automation", Short: "Manage Telco Automation (stubbed)"}

func init() {
	registerStubGroup(telcoAutomationCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(telcoAutomationCmd, "orchestration-cluster", "Manage orchestration cluster instances",
		"create", "delete", "describe", "list", "update", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")
	rootCmd.AddCommand(telcoAutomationCmd)
}
