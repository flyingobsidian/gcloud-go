package cmd

import "github.com/spf13/cobra"

// --- gcloud service-directory (#382) ---

var serviceDirectoryCmd = &cobra.Command{Use: "service-directory", Short: "Manage Service Directory (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(serviceDirectoryCmd, "endpoints", "Manage endpoints", crud...)
	registerStubGroup(serviceDirectoryCmd, "locations", "Manage locations", "describe", "list")
	registerStubGroup(serviceDirectoryCmd, "namespaces", "Manage namespaces", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(serviceDirectoryCmd, "services", "Manage services", append(crud, "resolve")...)
	rootCmd.AddCommand(serviceDirectoryCmd)
}
