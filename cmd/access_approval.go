package cmd

import "github.com/spf13/cobra"

// --- gcloud access-approval (#287) ---

// accessApprovalCmd exposes the gcloud-python access-approval surface.
// All three direct child subgroups (requests, service-account, settings) are
// registered; leaf commands currently error with "not yet implemented" pending
// full API wiring. See #566.
var accessApprovalCmd = &cobra.Command{Use: "access-approval", Short: "Manage Access Approval"}

func init() {
	registerStubGroup(accessApprovalCmd, "requests", "Manage access approval requests",
		"approve", "dismiss", "get", "invalidate", "list")
	registerStubGroup(accessApprovalCmd, "service-account", "Manage service account", "get")
	registerStubGroup(accessApprovalCmd, "settings", "Manage settings", "delete", "describe", "update")
	rootCmd.AddCommand(accessApprovalCmd)
}
