package cmd

import "github.com/spf13/cobra"

// --- gcloud apphub (#300) ---

var apphubCmd = &cobra.Command{Use: "apphub", Short: "Manage App Hub"}

func init() {
	// All subgroups (applications with nested services + workloads, boundary,
	// discovered-services, discovered-workloads, locations, operations,
	// service-projects) are implemented in apphub_all.go.
	rootCmd.AddCommand(apphubCmd)
}
