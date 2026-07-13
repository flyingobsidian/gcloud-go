package cmd

import "github.com/spf13/cobra"

// --- gcloud endpoints (#336) ---

var endpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Endpoints services"}

func init() {
	registerStubGroup(endpointsCmd, "configs", "View service configurations", "describe", "list")
	registerStubGroup(endpointsCmd, "operations", "Manage operations", "describe", "list", "wait")
	registerStubGroup(endpointsCmd, "services", "Manage services", "add-iam-policy-binding", "check-iam-policy", "delete", "deploy", "describe", "get-iam-policy", "list", "remove-iam-policy-binding", "set-iam-policy", "undelete", "check-config", "generate-openapi-config", "update-metrics-report-metadata")
	rootCmd.AddCommand(endpointsCmd)
}
