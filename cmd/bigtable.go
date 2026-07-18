package cmd

import "github.com/spf13/cobra"

// --- gcloud bigtable (#308) ---

var bigtableCmd = &cobra.Command{Use: "bigtable", Short: "Manage Cloud Bigtable"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(bigtableCmd, "app-profiles", "Manage app profiles", crud...)
	registerStubGroup(bigtableCmd, "authorized-views", "Manage authorized views", crud...)
	registerStubGroup(bigtableCmd, "backups", "Manage backups", append(crud, "restore")...)
	registerStubGroup(bigtableCmd, "clusters", "Manage clusters", crud...)
	registerStubGroup(bigtableCmd, "hot-tablets", "Manage hot tablets", "list")
	registerStubGroup(bigtableCmd, "instances", "Manage instances", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(bigtableCmd, "logical-views", "Manage logical views", crud...)
	rootCmd.AddCommand(bigtableCmd)
}
