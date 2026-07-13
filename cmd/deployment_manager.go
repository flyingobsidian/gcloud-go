package cmd

import "github.com/spf13/cobra"

// --- gcloud deployment-manager (#328) ---

var deploymentManagerCmd = &cobra.Command{Use: "deployment-manager", Short: "Manage Deployment Manager"}

func init() {
	registerStubGroup(deploymentManagerCmd, "deployments", "Manage deployments",
		"create", "delete", "describe", "list", "update", "stop", "cancel-preview")
	registerStubGroup(deploymentManagerCmd, "manifests", "Manage manifests", "describe", "list")
	registerStubGroup(deploymentManagerCmd, "operations", "Manage operations", "describe", "list", "wait")
	registerStubGroup(deploymentManagerCmd, "resources", "Manage resources", "describe", "list")
	registerStubGroup(deploymentManagerCmd, "types", "Manage types",
		"list", "providers")
	rootCmd.AddCommand(deploymentManagerCmd)
}
