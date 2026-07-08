package cmd

import "github.com/spf13/cobra"

// --- gcloud database-migration (#322) ---

var databaseMigrationCmd = &cobra.Command{Use: "database-migration", Short: "Manage Database Migration Service (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(databaseMigrationCmd, "connection-profiles", "Manage connection profiles", crud...)
	registerStubGroup(databaseMigrationCmd, "conversion-workspaces", "Manage conversion workspaces", append(crud, "convert", "commit", "rollback", "apply", "seed")...)
	registerStubGroup(databaseMigrationCmd, "migration-jobs", "Manage migration jobs", append(crud, "start", "stop", "resume", "promote", "verify", "restart")...)
	registerStubGroup(databaseMigrationCmd, "objects", "Manage migration job objects", "describe", "list", "lookup")
	registerStubGroup(databaseMigrationCmd, "operations", "Manage operations", "cancel", "describe", "list")
	registerStubGroup(databaseMigrationCmd, "private-connections", "Manage private connections", crud...)
	rootCmd.AddCommand(databaseMigrationCmd)
}
