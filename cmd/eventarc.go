package cmd

import "github.com/spf13/cobra"

// --- gcloud eventarc (#338) ---

var eventarcCmd = &cobra.Command{Use: "eventarc", Short: "Manage Eventarc"}

func init() {
	// gcloud-python's `eventarc --help` groups: audit-logs-provider,
	// channel-connections, channels, enrollments, gke-destinations,
	// google-api-sources, google-channels, kafka-sources, locations,
	// message-buses, pipelines, providers, triggers. No `attributes`; that
	// stub was invented and has been removed (#1720). Every other subgroup
	// is registered from its own eventarc_*.go file.
	rootCmd.AddCommand(eventarcCmd)
}
