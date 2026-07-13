package cmd

import "github.com/spf13/cobra"

// --- gcloud design-center (#329) ---

var designCenterCmd = &cobra.Command{Use: "design-center", Short: "Manage Design Center"}

func init() {
	registerStubGroup(designCenterCmd, "locations", "Manage locations", "describe", "list")
	registerStubGroup(designCenterCmd, "operations", "Manage operations", "describe", "list", "cancel")
	registerStubGroup(designCenterCmd, "spaces", "Manage spaces", "create", "delete", "describe", "list", "update")
	rootCmd.AddCommand(designCenterCmd)
}
