package cmd

import "github.com/spf13/cobra"

// --- gcloud endpoints (#892-#894) ---
//
// The top-level `endpoints` command aggregates the Endpoints/Service
// Management subgroups. All subgroups (configs, operations, services) are
// implemented in endpoints_all.go.

var endpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Endpoints services"}

func init() {
	rootCmd.AddCommand(endpointsCmd)
}
