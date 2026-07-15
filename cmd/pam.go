package cmd

import "github.com/spf13/cobra"

// --- gcloud pam (#369) ---

var pamCmd = &cobra.Command{Use: "pam", Short: "Manage Privileged Access Manager"}

func init() {
	registerStubGroup(pamCmd, "grants", "Manage grants",
		"approve", "create", "deny", "describe", "list", "revoke", "search")
	registerStubGroup(pamCmd, "operations", "Manage PAM operations", "describe", "list")
	registerStubCommand(pamCmd, "check-onboarding-status", "Check PAM onboarding status for a resource")
	rootCmd.AddCommand(pamCmd)
}
