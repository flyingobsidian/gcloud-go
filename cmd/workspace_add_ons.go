package cmd

import "github.com/spf13/cobra"

// --- gcloud workspace-add-ons (#399) ---

var workspaceAddOnsCmd = &cobra.Command{Use: "workspace-add-ons", Short: "Manage Google Workspace Add-ons (stubbed)"}

func init() {
	registerStubGroup(workspaceAddOnsCmd, "deployments", "Manage deployments",
		"create", "delete", "describe", "list", "replace", "install", "uninstall")
	registerStubCommand(workspaceAddOnsCmd, "get-authorization", "Get authorization info for deployments in a project")
	rootCmd.AddCommand(workspaceAddOnsCmd)
}
