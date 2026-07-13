package cmd

import "github.com/spf13/cobra"

// --- gcloud service-health (#384) ---

var serviceHealthCmd = &cobra.Command{Use: "service-health", Short: "Manage Service Health"}

func init() {
	registerStubGroup(serviceHealthCmd, "events", "Manage events", "describe", "list")
	registerStubGroup(serviceHealthCmd, "organization-events", "Manage organization events", "describe", "list")
	registerStubGroup(serviceHealthCmd, "organization-impacts", "Manage organization impacts", "describe", "list")
	rootCmd.AddCommand(serviceHealthCmd)
}
