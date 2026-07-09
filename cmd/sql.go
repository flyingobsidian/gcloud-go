package cmd

import "github.com/spf13/cobra"

// --- gcloud sql (#388) ---

var sqlCmd = &cobra.Command{Use: "sql", Short: "Manage Cloud SQL (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(sqlCmd, "backups", "Manage backups", "create", "delete", "describe", "list", "restore")
	registerStubGroup(sqlCmd, "connect", "Connect to instances", "psql", "mysql", "sqlserver")
	registerStubGroup(sqlCmd, "databases", "Manage databases", crud...)
	registerStubGroup(sqlCmd, "export", "Export databases", "sql", "csv", "bak")
	registerStubGroup(sqlCmd, "flags", "List flags", "list")
	registerStubGroup(sqlCmd, "import", "Import databases", "sql", "csv", "bak")
	registerStubGroup(sqlCmd, "instances", "Manage instances", append(crud, "clone", "failover", "patch", "reset-async-replica-lag", "reset-ssl-config", "restart", "restore-backup", "start-replica", "stop-replica", "promote-replica", "reencrypt", "import", "export", "reschedule-maintenance", "release-ssrs-lease", "list-server-cas", "verify-external-sync-settings", "acquire-ssrs-lease")...)
	registerStubGroup(sqlCmd, "operations", "Manage operations", "cancel", "describe", "list", "wait")
	registerStubGroup(sqlCmd, "ssl", "Manage SSL certificates", "server-ca-certs")
	registerStubGroup(sqlCmd, "ssl-certs", "(DEPRECATED) Manage SSL certificates", "create", "delete", "describe", "list")
	registerStubGroup(sqlCmd, "tiers", "List tiers", "list")
	registerStubGroup(sqlCmd, "users", "Manage users", crud...)
	for _, name := range []string{"connect", "generate-login-token", "reschedule-maintenance"} {
		registerStubCommand(sqlCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(sqlCmd)
}
