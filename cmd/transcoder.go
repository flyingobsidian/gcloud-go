package cmd

import "github.com/spf13/cobra"

// --- gcloud transcoder (#392) ---

var transcoderCmd = &cobra.Command{Use: "transcoder", Short: "Manage Transcoder (stubbed)"}

func init() {
	registerStubGroup(transcoderCmd, "jobs", "Manage transcoder jobs", "create", "delete", "describe", "list")
	registerStubGroup(transcoderCmd, "templates", "Manage transcoder job templates", "create", "delete", "describe", "list")
	rootCmd.AddCommand(transcoderCmd)
}
