package cmd

import "github.com/spf13/cobra"

// --- gcloud apihub (#298) ---

var apihubCmd = &cobra.Command{Use: "apihub", Short: "Manage API Hub"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(apihubCmd, "apis", "Manage APIs", crud...)
	registerStubGroup(apihubCmd, "attributes", "Manage attributes", crud...)
	registerStubGroup(apihubCmd, "curations", "Manage curations", crud...)
	registerStubGroup(apihubCmd, "dependencies", "Manage dependencies", crud...)
	registerStubGroup(apihubCmd, "deployments", "Manage deployments", crud...)
	registerStubGroup(apihubCmd, "discovered-api-observations", "Manage discovered API observations", "describe", "list")
	registerStubGroup(apihubCmd, "external-apis", "Manage external APIs", crud...)
	registerStubGroup(apihubCmd, "host-project-registrations", "Manage host project registrations", "create", "describe", "list")
	registerStubGroup(apihubCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(apihubCmd, "plugins", "Manage plugins", append(crud, "enable", "disable")...)
	registerStubGroup(apihubCmd, "runtime-project-attachments", "Manage runtime project attachments", "create", "describe", "list", "delete")
		registerStubGroup(apihubCmd, "addons", "Manage addons", "list", "describe")
	registerStubGroup(apihubCmd, "api-hub-instances", "Manage api-hub-instances", "list", "describe")
	rootCmd.AddCommand(apihubCmd)
}
