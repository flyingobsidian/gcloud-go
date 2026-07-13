package cmd

import "github.com/spf13/cobra"

// --- gcloud apigee (#297) ---

var apigeeCmd = &cobra.Command{Use: "apigee", Short: "Manage Apigee"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(apigeeCmd, "apis", "Manage API proxies", "delete", "deploy", "describe", "list", "undeploy")
	registerStubGroup(apigeeCmd, "applications", "Manage applications", "describe", "list")
	registerStubGroup(apigeeCmd, "archives", "Manage archives", crud...)
	registerStubGroup(apigeeCmd, "deployments", "Manage deployments", "describe", "list")
	registerStubGroup(apigeeCmd, "developers", "Manage developers", crud...)
	registerStubGroup(apigeeCmd, "environments", "Manage environments", crud...)
	registerStubGroup(apigeeCmd, "operations", "Manage operations", "describe", "list")
	registerStubGroup(apigeeCmd, "organizations", "Manage organizations", "describe", "list", "provision")
	registerStubGroup(apigeeCmd, "products", "Manage API products", crud...)
	rootCmd.AddCommand(apigeeCmd)
}
