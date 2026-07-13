package cmd

import "github.com/spf13/cobra"

// --- gcloud alloydb (#293) ---

var alloydbCmd = &cobra.Command{Use: "alloydb", Short: "Manage AlloyDB"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(alloydbCmd, "backups", "Manage backups", crud...)
	registerStubGroup(alloydbCmd, "clusters", "Manage clusters", append(crud, "promote", "restart", "restore", "switchover", "upgrade", "migrate-cloud-sql", "export", "import")...)
	registerStubGroup(alloydbCmd, "instances", "Manage instances", append(crud, "restart", "failover", "inject-fault", "execute-sql")...)
	registerStubGroup(alloydbCmd, "operations", "Manage operations", "cancel", "describe", "list")
	registerStubGroup(alloydbCmd, "users", "Manage users", crud...)
	rootCmd.AddCommand(alloydbCmd)
}
