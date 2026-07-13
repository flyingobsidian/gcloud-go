package cmd

import "github.com/spf13/cobra"

// --- gcloud datastream (#326) ---

var datastreamCmd = &cobra.Command{Use: "datastream", Short: "Manage Datastream"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(datastreamCmd, "connection-profiles", "Manage connection profiles", crud...)
	registerStubGroup(datastreamCmd, "locations", "Manage locations", "describe", "list")
	registerStubGroup(datastreamCmd, "objects", "Manage stream objects", "describe", "list", "start-backfill", "stop-backfill", "lookup")
	registerStubGroup(datastreamCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(datastreamCmd, "private-connections", "Manage private connections", crud...)
	registerStubGroup(datastreamCmd, "routes", "Manage routes", crud...)
	registerStubGroup(datastreamCmd, "streams", "Manage streams", append(crud, "pause", "resume")...)
	rootCmd.AddCommand(datastreamCmd)
}
