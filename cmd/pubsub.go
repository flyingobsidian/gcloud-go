package cmd

import "github.com/spf13/cobra"

// --- gcloud pubsub (#377) ---

var pubsubCmd = &cobra.Command{Use: "pubsub", Short: "Manage Cloud Pub/Sub"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(pubsubCmd, "schemas", "Manage Pub/Sub schemas", append(crud, "commit", "revisions", "validate-schema", "validate-message")...)
	registerStubGroup(pubsubCmd, "snapshots", "Manage Pub/Sub snapshots", crud...)
	rootCmd.AddCommand(pubsubCmd)
}
