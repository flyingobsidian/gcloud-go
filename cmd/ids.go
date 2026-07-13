package cmd

import "github.com/spf13/cobra"

// --- gcloud ids (#347) ---

var idsCmd = &cobra.Command{Use: "ids", Short: "Manage Cloud IDS"}

func init() {
	registerStubGroup(idsCmd, "endpoints", "Manage IDS endpoints", "create", "delete", "describe", "list", "update")
	rootCmd.AddCommand(idsCmd)
}
