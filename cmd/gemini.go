package cmd

import "github.com/spf13/cobra"

// --- gcloud gemini (#343) ---

var geminiCmd = &cobra.Command{Use: "gemini", Short: "Manage Gemini Code Assist / Cloud Assist (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(geminiCmd, "code-repository-indexes", "Manage code repository indexes", crud...)
	registerStubGroup(geminiCmd, "code-tools-settings", "Manage code tools settings", crud...)
	registerStubGroup(geminiCmd, "data-sharing-with-google-settings", "Manage data sharing settings", crud...)
	registerStubGroup(geminiCmd, "gda-observability-settings", "Manage GDA observability settings", crud...)
	registerStubGroup(geminiCmd, "gemini-gcp-enablement-settings", "Manage GCP enablement settings", crud...)
	registerStubGroup(geminiCmd, "gibq-observability-settings", "Manage GIBQ observability settings", crud...)
	registerStubGroup(geminiCmd, "logging-settings", "Manage logging settings", crud...)
	registerStubGroup(geminiCmd, "operations", "Manage operations", "describe", "list", "cancel", "delete")
	registerStubGroup(geminiCmd, "release-channel-settings", "Manage release channel settings", crud...)
	rootCmd.AddCommand(geminiCmd)
}
