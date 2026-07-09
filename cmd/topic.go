package cmd

import "github.com/spf13/cobra"

// --- gcloud topic (#391) ---
//
// `topic` in gcloud-python is a supplementary help namespace: each sub-topic
// prints a reference document. gcloud-go does not currently ship these help
// pages; this stub exposes the topic surface so callers know it exists.

var topicCmd = &cobra.Command{
	Use:   "topic",
	Short: "gcloud supplementary help topics (stubbed)",
}

func init() {
	for _, name := range []string{
		"accessibility", "arg-files", "cli-trees", "client-certificate",
		"command-conventions", "configurations", "datetimes", "endpoint-override",
		"escaping", "filters", "flags-file", "formats", "gcloudignore",
		"offline-help", "projections", "resource-keys", "startup", "uninstall",
	} {
		registerStubCommand(topicCmd, name, "Supplementary help topic (not yet ported)")
	}
	rootCmd.AddCommand(topicCmd)
}
