package cmd

import "github.com/spf13/cobra"

// --- gcloud transfer (#887-#891) ---
//
// The top-level `transfer` command aggregates the Storage Transfer Service
// subgroups. All subgroups (agent-pools, agents, jobs, operations) and the
// authorize command are implemented in transfer_all.go.

var transferCmd = &cobra.Command{Use: "transfer", Short: "Manage Storage Transfer Service"}

func init() {
	rootCmd.AddCommand(transferCmd)
}
