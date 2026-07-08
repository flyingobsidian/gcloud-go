package cmd

import "github.com/spf13/cobra"

// --- gcloud policy-troubleshoot (#372) ---

var policyTroubleshootCmd = &cobra.Command{Use: "policy-troubleshoot", Short: "Policy troubleshoot (stubbed)"}

func init() {
	registerStubCommand(policyTroubleshootCmd, "iam", "Troubleshoot the IAM Policy")
	rootCmd.AddCommand(policyTroubleshootCmd)
}
