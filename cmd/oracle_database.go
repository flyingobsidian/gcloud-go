package cmd

import "github.com/spf13/cobra"

// --- gcloud oracle-database (#367) ---

var oracleDatabaseCmd = &cobra.Command{Use: "oracle-database", Short: "Manage Oracle Database"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(oracleDatabaseCmd, "autonomous-database-backups", "Manage autonomous database backups", crud...)
	registerStubGroup(oracleDatabaseCmd, "autonomous-databases", "Manage autonomous databases", append(crud, "generate-wallet", "restore", "restart", "stop", "start")...)
	registerStubGroup(oracleDatabaseCmd, "backups", "Manage backups", crud...)
	registerStubGroup(oracleDatabaseCmd, "db-nodes", "Manage DB nodes", "describe", "list", "action")
	registerStubGroup(oracleDatabaseCmd, "db-servers", "Manage DB servers", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-connection-assignments", "Manage GoldenGate connection assignments", crud...)
	registerStubGroup(oracleDatabaseCmd, "goldengate-connection-types", "Manage GoldenGate connection types", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-connections", "Manage GoldenGate connections", crud...)
	registerStubGroup(oracleDatabaseCmd, "goldengate-deployment-environments", "Manage GoldenGate deployment environments", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-deployment-types", "Manage GoldenGate deployment types", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-deployment-versions", "Manage GoldenGate deployment versions", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-deployments", "Manage GoldenGate deployments", crud...)
	rootCmd.AddCommand(oracleDatabaseCmd)
}
