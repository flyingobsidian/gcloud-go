package cmd

import "github.com/spf13/cobra"

// --- gcloud database-migration (#322) ---

var databaseMigrationCmd = &cobra.Command{Use: "database-migration", Short: "Manage Database Migration Service"}

func init() {
	// Subgroups are implemented in dedicated files:
	//   - connection-profiles: database_migration_connection_profiles.go (#782)
	//   - conversion-workspaces: database_migration_conversion_workspaces.go (#783)
	//   - migration-jobs: database_migration_migration_jobs.go (#784)
	//   - objects: database_migration_objects.go (#785)
	//   - operations: database_migration_operations.go (#786)
	//   - private-connections: database_migration_private_connections.go (#787)
	rootCmd.AddCommand(databaseMigrationCmd)
}
