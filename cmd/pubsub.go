package cmd

import "github.com/spf13/cobra"

// --- gcloud pubsub (#377) ---

var pubsubCmd = &cobra.Command{Use: "pubsub", Short: "Manage Cloud Pub/Sub (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(pubsubCmd, "lite-operations", "Manage Pub/Sub Lite operations", "cancel", "describe", "list")
	registerStubGroup(pubsubCmd, "lite-reservations", "Manage Pub/Sub Lite reservations", crud...)
	registerStubGroup(pubsubCmd, "lite-subscriptions", "Manage Pub/Sub Lite subscriptions", append(crud, "seek")...)
	registerStubGroup(pubsubCmd, "lite-topics", "Manage Pub/Sub Lite topics", append(crud, "list-subscriptions")...)
	registerStubGroup(pubsubCmd, "message-transforms", "Manage Cloud Pub/Sub message transforms", "test")
	registerStubGroup(pubsubCmd, "schemas", "Manage Pub/Sub schemas", append(crud, "commit", "revisions", "validate-schema", "validate-message")...)
	registerStubGroup(pubsubCmd, "snapshots", "Manage Pub/Sub snapshots", crud...)
	registerStubGroup(pubsubCmd, "subscriptions", "Manage Pub/Sub subscriptions", append(crud, "ack", "pull", "seek", "modify-message-ack-deadline", "modify-push-config")...)
	registerStubGroup(pubsubCmd, "topics", "Manage Pub/Sub topics", append(crud, "publish", "list-subscriptions", "list-snapshots", "detach-subscription")...)
	rootCmd.AddCommand(pubsubCmd)
}
