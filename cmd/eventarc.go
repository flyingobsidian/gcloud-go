package cmd

import "github.com/spf13/cobra"

// --- gcloud eventarc (#338) ---

var eventarcCmd = &cobra.Command{Use: "eventarc", Short: "Manage Eventarc"}

func init() {
	// `attributes` remains as a stub group until its own implementation lands;
	// every other subgroup is implemented in a dedicated eventarc_*.go file.
	registerStubGroup(eventarcCmd, "attributes", "Manage attributes", "list")
	rootCmd.AddCommand(eventarcCmd)
}
