package cmd

import "github.com/spf13/cobra"

// --- gcloud policy-intelligence (#371) ---

var policyIntelligenceCmd = &cobra.Command{Use: "policy-intelligence", Short: "Policy Intelligence (stubbed)"}

func init() {
	simulate := &cobra.Command{Use: "simulate", Short: "Simulate policy changes"}
	registerStubGroup(simulate, "orgpolicy", "Simulate org policy changes", "create", "describe", "list")
	policyIntelligenceCmd.AddCommand(simulate)
	registerStubGroup(policyIntelligenceCmd, "troubleshoot-policy", "Troubleshoot IAM policies", "iam")
	registerStubCommand(policyIntelligenceCmd, "query-activity", "Query activities on cloud resource")
	rootCmd.AddCommand(policyIntelligenceCmd)
}
