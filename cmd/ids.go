package cmd

import "github.com/spf13/cobra"

// --- gcloud ids (#347) ---

var idsCmd = &cobra.Command{Use: "ids", Short: "Manage Cloud IDS"}

func init() {
	rootCmd.AddCommand(idsCmd)
}
