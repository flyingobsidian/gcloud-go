package cmd

import "github.com/spf13/cobra"

// --- gcloud workload-identity (#398) ---
//
// Note: gcloud-go already ships workload identity federation commands under
// `gcloud iam workload-identity-pools` (see #201 / #202). This file adds the
// standalone `gcloud workload-identity` surface called out in gcloud-python
// as a stub pending consolidation with the iam implementations.

var workloadIdentityCmd = &cobra.Command{Use: "workload-identity", Short: "Manage Workload Identity (stubbed)"}

func init() {
	registerStubGroup(workloadIdentityCmd, "service-agents", "Manage Workload Identity service agents", "list", "describe")
	rootCmd.AddCommand(workloadIdentityCmd)
}
