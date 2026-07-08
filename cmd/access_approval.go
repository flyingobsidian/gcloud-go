package cmd

import "github.com/spf13/cobra"

// --- gcloud access-approval (#287) ---

var accessApprovalCmd = &cobra.Command{Use: "access-approval", Short: "Manage Access Approval (stubbed)"}

func init() {
	registerStubGroup(accessApprovalCmd, "requests", "Manage access approval requests",
		"approve", "dismiss", "get", "invalidate", "list")
	registerStubGroup(accessApprovalCmd, "service-account", "Manage service account", "get")
	registerStubGroup(accessApprovalCmd, "settings", "Manage settings", "delete", "describe", "update")
	rootCmd.AddCommand(accessApprovalCmd)
}
