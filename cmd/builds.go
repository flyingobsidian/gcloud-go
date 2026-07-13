package cmd

import "github.com/spf13/cobra"

// --- gcloud builds (#312) ---

var buildsCmd = &cobra.Command{Use: "builds", Short: "Manage Cloud Build"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(buildsCmd, "connections", "Manage connections", crud...)
	registerStubGroup(buildsCmd, "repositories", "Manage repositories", crud...)
	registerStubGroup(buildsCmd, "triggers", "Manage build triggers", append(crud, "run", "import", "export")...)
	registerStubGroup(buildsCmd, "worker-pools", "Manage worker pools", crud...)
	for _, name := range []string{"cancel", "describe", "get-default-service-account", "list", "log", "submit"} {
		registerStubCommand(buildsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(buildsCmd)
}
