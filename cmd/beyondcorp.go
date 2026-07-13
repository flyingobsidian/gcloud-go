package cmd

import "github.com/spf13/cobra"

// --- gcloud beyondcorp (#306) ---

var beyondcorpCmd = &cobra.Command{
	Use:   "beyondcorp",
	Short: "Manage BeyondCorp resources",
}

func init() {
	registerStubGroup(beyondcorpCmd, "operations", "Manage BeyondCorp operations", "describe", "list", "cancel", "delete")
	registerStubGroup(beyondcorpCmd, "security-gateways", "Manage BeyondCorp security gateways", "create", "delete", "describe", "list", "update")
	rootCmd.AddCommand(beyondcorpCmd)
}
