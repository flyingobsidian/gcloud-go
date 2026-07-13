package cmd

import "github.com/spf13/cobra"

// --- gcloud infra-manager (#348) ---

var infraManagerCmd = &cobra.Command{Use: "infra-manager", Short: "Manage Infrastructure Manager"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(infraManagerCmd, "automigrationconfig", "Manage auto migration config", "describe", "update")
	registerStubGroup(infraManagerCmd, "deployments", "Manage deployments", append(crud, "export-lock", "export-state", "import-state", "lock", "unlock")...)
	registerStubGroup(infraManagerCmd, "previews", "Manage previews", "create", "delete", "describe", "list", "export")
	registerStubGroup(infraManagerCmd, "resource-changes", "Manage resource changes", "describe", "list")
	registerStubGroup(infraManagerCmd, "resource-drifts", "Manage resource drifts", "describe", "list")
	registerStubGroup(infraManagerCmd, "resources", "List revision resources", "describe", "list")
	registerStubGroup(infraManagerCmd, "revisions", "Manage revisions", "describe", "list")
	registerStubGroup(infraManagerCmd, "terraform-versions", "Manage Terraform versions", "describe", "list")
	rootCmd.AddCommand(infraManagerCmd)
}
