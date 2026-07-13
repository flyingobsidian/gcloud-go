package cmd

import "github.com/spf13/cobra"

// --- gcloud publicca (#376) ---

var publiccaCmd = &cobra.Command{Use: "publicca", Short: "Manage Google Trust Services PublicCA"}

func init() {
	registerStubGroup(publiccaCmd, "external-account-keys", "Manage ACME external account keys", "create")
	rootCmd.AddCommand(publiccaCmd)
}
