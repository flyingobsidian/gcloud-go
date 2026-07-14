package cmd

import "github.com/spf13/cobra"

// --- gcloud datastream (#326) ---

var datastreamCmd = &cobra.Command{Use: "datastream", Short: "Manage Datastream"}

func init() {
	// All subgroups (connection-profiles, locations, objects, operations,
	// private-connections, routes, streams) are implemented in
	// datastream_all.go.
	rootCmd.AddCommand(datastreamCmd)
}
