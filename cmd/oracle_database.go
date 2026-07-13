package cmd

import "github.com/spf13/cobra"

// --- gcloud oracle-database (#367) ---

var oracleDatabaseCmd = &cobra.Command{Use: "oracle-database", Short: "Manage Oracle Database"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(oracleDatabaseCmd, "autonomous-database-backups", "Manage autonomous database backups", crud...)
	registerStubGroup(oracleDatabaseCmd, "autonomous-database-character-sets", "Manage autonomous database character sets", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "autonomous-database-versions", "Manage autonomous database versions", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "autonomous-databases", "Manage autonomous databases", append(crud, "generate-wallet", "restore", "restart", "stop", "start")...)
	registerStubGroup(oracleDatabaseCmd, "backups", "Manage backups", crud...)
	registerStubGroup(oracleDatabaseCmd, "cloud-exadata-infrastructures", "Manage Exadata infrastructures", crud...)
	registerStubGroup(oracleDatabaseCmd, "cloud-vm-clusters", "Manage cloud VM clusters", crud...)
	registerStubGroup(oracleDatabaseCmd, "db-nodes", "Manage DB nodes", "describe", "list", "action")
	registerStubGroup(oracleDatabaseCmd, "db-servers", "Manage DB servers", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "db-system-shapes", "Manage DB system shapes", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "entitlements", "Manage entitlements", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "gi-versions", "Manage GI versions", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-connection-assignments", "Manage GoldenGate connection assignments", crud...)
	registerStubGroup(oracleDatabaseCmd, "goldengate-connection-types", "Manage GoldenGate connection types", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-connections", "Manage GoldenGate connections", crud...)
	registerStubGroup(oracleDatabaseCmd, "goldengate-deployment-environments", "Manage GoldenGate deployment environments", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-deployment-types", "Manage GoldenGate deployment types", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-deployment-versions", "Manage GoldenGate deployment versions", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "goldengate-deployments", "Manage GoldenGate deployments", crud...)
	registerStubGroup(oracleDatabaseCmd, "odb-networks", "Manage ODB networks", crud...)
	registerStubGroup(oracleDatabaseCmd, "operations", "Manage operations", "cancel", "describe", "list")
	registerStubGroup(oracleDatabaseCmd, "pluggable-databases", "Manage pluggable databases", crud...)
		registerStubGroup(oracleDatabaseCmd, "autonomous-db-versions", "Manage autonomous-db-versions", "list", "describe")
	registerStubGroup(oracleDatabaseCmd, "database-character-sets", "Manage database-character-sets", "list", "describe")
	registerStubGroup(oracleDatabaseCmd, "databases", "Manage databases", "list", "describe")
	registerStubGroup(oracleDatabaseCmd, "db-system-initial-storage-sizes", "Manage db-system-initial-storage-sizes", "list", "describe")
	registerStubGroup(oracleDatabaseCmd, "db-systems", "Manage db-systems", "list", "describe")
	registerStubGroup(oracleDatabaseCmd, "db-versions", "Manage db-versions", "list", "describe")
	registerStubGroup(oracleDatabaseCmd, "exadb-vm-clusters", "Manage exadb-vm-clusters", "list", "describe")
	registerStubGroup(oracleDatabaseCmd, "exascale-db-storage-vaults", "Manage exascale-db-storage-vaults", "list", "describe")
	rootCmd.AddCommand(oracleDatabaseCmd)
}
