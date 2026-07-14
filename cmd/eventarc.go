package cmd

import "github.com/spf13/cobra"

// --- gcloud eventarc (#338) ---

var eventarcCmd = &cobra.Command{Use: "eventarc", Short: "Manage Eventarc"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(eventarcCmd, "attributes", "Manage attributes", "list")
	registerStubGroup(eventarcCmd, "channel-connections", "Manage channel connections", "create", "delete", "describe", "list")
	registerStubGroup(eventarcCmd, "channels", "Manage channels", crud...)
	registerStubGroup(eventarcCmd, "enrollments", "Manage enrollments", crud...)
	registerStubGroup(eventarcCmd, "google-api-sources", "Manage Google API sources", crud...)
	registerStubGroup(eventarcCmd, "google-channels", "Manage Google channels", crud...)
	registerStubGroup(eventarcCmd, "locations", "Explore locations", "describe", "list")
	registerStubGroup(eventarcCmd, "message-buses", "Manage message buses", crud...)
	registerStubGroup(eventarcCmd, "pipelines", "Manage pipelines", crud...)
	registerStubGroup(eventarcCmd, "providers", "Explore event providers", "describe", "list")
	registerStubGroup(eventarcCmd, "triggers", "Manage triggers", crud...)
	rootCmd.AddCommand(eventarcCmd)
}
