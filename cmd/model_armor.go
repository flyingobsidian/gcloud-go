package cmd

import "github.com/spf13/cobra"

// --- gcloud model-armor (#359) ---

var modelArmorCmd = &cobra.Command{Use: "model-armor", Short: "Manage Model Armor (stubbed)"}

func init() {
	registerStubGroup(modelArmorCmd, "floorsettings", "Manage floor settings", "describe", "update")
	registerStubGroup(modelArmorCmd, "templates", "Manage templates", "create", "delete", "describe", "list", "update", "sanitize-model-response", "sanitize-user-prompt")
	rootCmd.AddCommand(modelArmorCmd)
}
