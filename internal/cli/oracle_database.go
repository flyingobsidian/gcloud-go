package cli

import "github.com/spf13/cobra"

// --- gcloud oracle-database (#367) ---
//
// The real subgroups (autonomous-databases, autonomous-database-backups,
// cloud-exadata-infrastructures, cloud-vm-clusters, databases, db-systems,
// exadb-vm-clusters, exascale-db-storage-vaults, odb-networks,
// pluggable-databases, goldengate-* etc.) are registered from their own
// oracle_database_*.go files.
//
// The earlier `backups`, `db-nodes`, and `db-servers` stubs were invented --
// gcloud-python's `oracle-database` surface has no such subgroups (the real
// backup group is `autonomous-database-backups`) -- and have been removed
// (#1721).

func (s *oracleDatabaseCmds) initOracleDatabase() {

}

// oracleDatabaseCmds holds the commands and flag values of one invocation's "oracleDatabase"
// command group, so invocations don't share state (#2070).
type oracleDatabaseCmds struct {
	inv *invocation

	oracleDatabaseCmd                        *cobra.Command
	oracleAutonomousDatabaseBackupsCmd       *cobra.Command
	flagOdbAdbBkLocation                     string
	flagOdbAdbBkFormat                       string
	flagOdbAdbBkFilter                       string
	flagOdbAdbBkPageSize                     int64
	oracleAdbBkListCmd                       *cobra.Command
	oracleAutonomousDatabaseCharacterSetsCmd *cobra.Command
	flagOdbAdCsLocation                      string
	flagOdbAdCsFormat                        string
	flagOdbAdCsFilter                        string
	flagOdbAdCsPageSize                      int64
	oracleAdCsListCmd                        *cobra.Command
	oracleAutonomousDatabasesCmd             *cobra.Command
	flagOdbAdbLocation                       string
	flagOdbAdbFormat                         string
	flagOdbAdbConfigFile                     string
	flagOdbAdbUpdateMask                     string
	flagOdbAdbPageSize                       int64
	flagOdbAdbFilter                         string
	oracleAdbCreateCmd                       *cobra.Command
	oracleAdbDeleteCmd                       *cobra.Command
	oracleAdbDescribeCmd                     *cobra.Command
	oracleAdbListCmd                         *cobra.Command
	oracleAdbUpdateCmd                       *cobra.Command
	oracleAdbFailoverCmd                     *cobra.Command
	oracleAdbGenerateWalletCmd               *cobra.Command
	oracleAdbRestartCmd                      *cobra.Command
	oracleAdbRestoreCmd                      *cobra.Command
	oracleAdbStartCmd                        *cobra.Command
	oracleAdbStopCmd                         *cobra.Command
	oracleAdbSwitchoverCmd                   *cobra.Command
	oracleAutonomousDbVersionsCmd            *cobra.Command
	flagOdbAdvLocation                       string
	flagOdbAdvFormat                         string
	flagOdbAdvPageSize                       int64
	oracleAdvListCmd                         *cobra.Command
	oracleCloudExadataInfrastructuresCmd     *cobra.Command
	flagOdbExInfLocation                     string
	flagOdbExInfFormat                       string
	flagOdbExInfConfigFile                   string
	flagOdbExInfPageSize                     int64
	flagOdbExInfFilter                       string
	oracleExInfCreateCmd                     *cobra.Command
	oracleExInfDeleteCmd                     *cobra.Command
	oracleExInfDescribeCmd                   *cobra.Command
	oracleExInfListCmd                       *cobra.Command
	oracleCloudVmClustersCmd                 *cobra.Command
	flagOdbVmClLocation                      string
	flagOdbVmClFormat                        string
	flagOdbVmClConfigFile                    string
	flagOdbVmClPageSize                      int64
	flagOdbVmClFilter                        string
	oracleVmClCreateCmd                      *cobra.Command
	oracleVmClDeleteCmd                      *cobra.Command
	oracleVmClDescribeCmd                    *cobra.Command
	oracleVmClListCmd                        *cobra.Command
	oracleDatabaseCharacterSetsCmd           *cobra.Command
	flagOdbDbCsLocation                      string
	flagOdbDbCsFormat                        string
	flagOdbDbCsFilter                        string
	flagOdbDbCsPageSize                      int64
	oracleDbCsListCmd                        *cobra.Command
	oracleDatabasesCmd                       *cobra.Command
	flagOdbDbLocation                        string
	flagOdbDbFormat                          string
	flagOdbDbPageSize                        int64
	flagOdbDbFilter                          string
	oracleDbDescribeCmd                      *cobra.Command
	oracleDbListCmd                          *cobra.Command
	oracleDbSystemInitialStorageSizesCmd     *cobra.Command
	flagOdbSizLocation                       string
	flagOdbSizFormat                         string
	flagOdbSizPageSize                       int64
	oracleSizListCmd                         *cobra.Command
	oracleDbSystemShapesCmd                  *cobra.Command
	flagOdbShapeLocation                     string
	flagOdbShapeFormat                       string
	flagOdbShapeFilter                       string
	flagOdbShapePageSize                     int64
	oracleShapeListCmd                       *cobra.Command
	oracleDbSystemsCmd                       *cobra.Command
	flagOdbDbsLocation                       string
	flagOdbDbsFormat                         string
	flagOdbDbsConfigFile                     string
	flagOdbDbsPageSize                       int64
	flagOdbDbsFilter                         string
	oracleDbsCreateCmd                       *cobra.Command
	oracleDbsDeleteCmd                       *cobra.Command
	oracleDbsDescribeCmd                     *cobra.Command
	oracleDbsListCmd                         *cobra.Command
	oracleDbVersionsCmd                      *cobra.Command
	flagOdbDbvLocation                       string
	flagOdbDbvFormat                         string
	flagOdbDbvFilter                         string
	flagOdbDbvPageSize                       int64
	oracleDbvListCmd                         *cobra.Command
	oracleEntitlementsCmd                    *cobra.Command
	flagOdbEntLocation                       string
	flagOdbEntFormat                         string
	flagOdbEntPageSize                       int64
	oracleEntListCmd                         *cobra.Command
	oracleExadbVmClustersCmd                 *cobra.Command
	flagOdbExaVmLocation                     string
	flagOdbExaVmFormat                       string
	flagOdbExaVmConfigFile                   string
	flagOdbExaVmUpdateMask                   string
	flagOdbExaVmPageSize                     int64
	flagOdbExaVmFilter                       string
	oracleExaVmCreateCmd                     *cobra.Command
	oracleExaVmDeleteCmd                     *cobra.Command
	oracleExaVmDescribeCmd                   *cobra.Command
	oracleExaVmListCmd                       *cobra.Command
	oracleExaVmUpdateCmd                     *cobra.Command
	oracleExascaleDbStorageVaultsCmd         *cobra.Command
	flagOdbExaStLocation                     string
	flagOdbExaStFormat                       string
	flagOdbExaStConfigFile                   string
	flagOdbExaStPageSize                     int64
	flagOdbExaStFilter                       string
	oracleExaStCreateCmd                     *cobra.Command
	oracleExaStDeleteCmd                     *cobra.Command
	oracleExaStDescribeCmd                   *cobra.Command
	oracleExaStListCmd                       *cobra.Command
	oracleGiVersionsCmd                      *cobra.Command
	flagOdbGiLocation                        string
	flagOdbGiFormat                          string
	flagOdbGiFilter                          string
	flagOdbGiPageSize                        int64
	oracleGiListCmd                          *cobra.Command
	oracleGgConnAssignmentsCmd               *cobra.Command
	flagOdbGgCaLocation                      string
	flagOdbGgCaFormat                        string
	flagOdbGgCaConfigFile                    string
	flagOdbGgCaPageSize                      int64
	flagOdbGgCaFilter                        string
	oracleGgCaCreateCmd                      *cobra.Command
	oracleGgCaDeleteCmd                      *cobra.Command
	oracleGgCaDescribeCmd                    *cobra.Command
	oracleGgCaListCmd                        *cobra.Command
	oracleGgCaTestCmd                        *cobra.Command
	oracleGgConnTypesCmd                     *cobra.Command
	flagOdbGgCtLocation                      string
	flagOdbGgCtFormat                        string
	flagOdbGgCtPageSize                      int64
	flagOdbGgCtFilter                        string
	oracleGgCtListCmd                        *cobra.Command
	oracleGgConnectionsCmd                   *cobra.Command
	flagOdbGgCnLocation                      string
	flagOdbGgCnFormat                        string
	flagOdbGgCnConfigFile                    string
	flagOdbGgCnPageSize                      int64
	flagOdbGgCnFilter                        string
	oracleGgCnCreateCmd                      *cobra.Command
	oracleGgCnDeleteCmd                      *cobra.Command
	oracleGgCnDescribeCmd                    *cobra.Command
	oracleGgCnListCmd                        *cobra.Command
	oracleGgDeployEnvsCmd                    *cobra.Command
	flagOdbGgDeLocation                      string
	flagOdbGgDeFormat                        string
	flagOdbGgDePageSize                      int64
	oracleGgDeCmd                            *cobra.Command
	oracleGgDeployTypesCmd                   *cobra.Command
	flagOdbGgDtLocation                      string
	flagOdbGgDtFormat                        string
	flagOdbGgDtPageSize                      int64
	flagOdbGgDtFilter                        string
	oracleGgDtListCmd                        *cobra.Command
	oracleGgDeployVersionsCmd                *cobra.Command
	flagOdbGgDvLocation                      string
	flagOdbGgDvFormat                        string
	flagOdbGgDvPageSize                      int64
	flagOdbGgDvFilter                        string
	oracleGgDvListCmd                        *cobra.Command
	oracleGgDeploymentsCmd                   *cobra.Command
	flagOdbGgDpLocation                      string
	flagOdbGgDpFormat                        string
	flagOdbGgDpConfigFile                    string
	flagOdbGgDpPageSize                      int64
	flagOdbGgDpFilter                        string
	oracleGgDpCreateCmd                      *cobra.Command
	oracleGgDpDeleteCmd                      *cobra.Command
	oracleGgDpDescribeCmd                    *cobra.Command
	oracleGgDpListCmd                        *cobra.Command
	oracleGgDpStartCmd                       *cobra.Command
	oracleGgDpStopCmd                        *cobra.Command
	oracleOdbNetworksCmd                     *cobra.Command
	flagOdbNetLocation                       string
	flagOdbNetFormat                         string
	flagOdbNetConfigFile                     string
	flagOdbNetPageSize                       int64
	flagOdbNetFilter                         string
	oracleOdbNetCreateCmd                    *cobra.Command
	oracleOdbNetDeleteCmd                    *cobra.Command
	oracleOdbNetDescribeCmd                  *cobra.Command
	oracleOdbNetListCmd                      *cobra.Command
	oracleOperationsCmd                      *cobra.Command
	flagOdbOpLocation                        string
	flagOdbOpFormat                          string
	flagOdbOpFilter                          string
	flagOdbOpPageSize                        int64
	oracleOpCancelCmd                        *cobra.Command
	oracleOpDeleteCmd                        *cobra.Command
	oracleOpDescribeCmd                      *cobra.Command
	oracleOpListCmd                          *cobra.Command
	oraclePluggableDatabasesCmd              *cobra.Command
	flagOdbPdbLocation                       string
	flagOdbPdbFormat                         string
	flagOdbPdbPageSize                       int64
	flagOdbPdbFilter                         string
	oraclePdbDescribeCmd                     *cobra.Command
	oraclePdbListCmd                         *cobra.Command
}

// newOracleDatabaseCmd returns the "oracleDatabase" command group for inv.
func newOracleDatabaseCmd(inv *invocation) *cobra.Command {
	return newOracleDatabaseCmds(inv).oracleDatabaseCmd
}

// newOracleDatabaseCmds builds the "oracleDatabase" command group for inv and returns its state.
func newOracleDatabaseCmds(inv *invocation) *oracleDatabaseCmds {
	s := &oracleDatabaseCmds{inv: inv}
	s.oracleDatabaseCmd = &cobra.Command{Use: "oracle-database", Short: "Manage Oracle Database"}
	s.oracleAutonomousDatabaseBackupsCmd = &cobra.Command{Use: "autonomous-database-backups", Short: "Manage Oracle Autonomous Database backups"}
	s.oracleAdbBkListCmd = &cobra.Command{
		Use: "list", Short: "List Autonomous Database backups in a location",
		Args: cobra.NoArgs, RunE: s.runOracleAdbBkList,
	}
	s.oracleAutonomousDatabaseCharacterSetsCmd = &cobra.Command{
		Use:   "autonomous-database-character-sets",
		Short: "Manage Autonomous Database character sets",
	}
	s.oracleAdCsListCmd = &cobra.Command{
		Use: "list", Short: "List Autonomous Database character sets in a location",
		Args: cobra.NoArgs, RunE: s.runOracleAdCsList,
	}
	s.oracleAutonomousDatabasesCmd = &cobra.Command{Use: "autonomous-databases", Short: "Manage Oracle Autonomous Databases"}
	s.oracleAdbCreateCmd = &cobra.Command{
		Use: "create AUTONOMOUS_DATABASE", Short: "Create an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbCreate,
	}
	s.oracleAdbDeleteCmd = &cobra.Command{
		Use: "delete AUTONOMOUS_DATABASE", Short: "Delete an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbDelete,
	}
	s.oracleAdbDescribeCmd = &cobra.Command{
		Use: "describe AUTONOMOUS_DATABASE", Short: "Describe an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbDescribe,
	}
	s.oracleAdbListCmd = &cobra.Command{
		Use: "list", Short: "List Autonomous Databases in a location",
		Args: cobra.NoArgs, RunE: s.runOracleAdbList,
	}
	s.oracleAdbUpdateCmd = &cobra.Command{
		Use: "update AUTONOMOUS_DATABASE", Short: "Update an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbUpdate,
	}
	s.oracleAdbFailoverCmd = &cobra.Command{
		Use: "failover AUTONOMOUS_DATABASE", Short: "Failover an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbFailover,
	}
	s.oracleAdbGenerateWalletCmd = &cobra.Command{
		Use: "generate-wallet AUTONOMOUS_DATABASE", Short: "Generate a wallet for an Autonomous Database (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbGenerateWallet,
	}
	s.oracleAdbRestartCmd = &cobra.Command{
		Use: "restart AUTONOMOUS_DATABASE", Short: "Restart an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbRestart,
	}
	s.oracleAdbRestoreCmd = &cobra.Command{
		Use: "restore AUTONOMOUS_DATABASE", Short: "Restore an Autonomous Database from a backup (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbRestore,
	}
	s.oracleAdbStartCmd = &cobra.Command{
		Use: "start AUTONOMOUS_DATABASE", Short: "Start an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbStart,
	}
	s.oracleAdbStopCmd = &cobra.Command{
		Use: "stop AUTONOMOUS_DATABASE", Short: "Stop an Autonomous Database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbStop,
	}
	s.oracleAdbSwitchoverCmd = &cobra.Command{
		Use: "switchover AUTONOMOUS_DATABASE", Short: "Switchover an Autonomous Database (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runOracleAdbSwitchover,
	}
	s.oracleAutonomousDbVersionsCmd = &cobra.Command{
		Use:   "autonomous-db-versions",
		Short: "Manage Autonomous DB versions",
	}
	s.oracleAdvListCmd = &cobra.Command{
		Use: "list", Short: "List Autonomous DB versions in a location",
		Args: cobra.NoArgs, RunE: s.runOracleAdvList,
	}
	s.oracleCloudExadataInfrastructuresCmd = &cobra.Command{
		Use:   "cloud-exadata-infrastructures",
		Short: "Manage Oracle Cloud Exadata infrastructures",
	}
	s.oracleExInfCreateCmd = &cobra.Command{
		Use: "create INFRASTRUCTURE", Short: "Create a cloud Exadata infrastructure",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExInfCreate,
	}
	s.oracleExInfDeleteCmd = &cobra.Command{
		Use: "delete INFRASTRUCTURE", Short: "Delete a cloud Exadata infrastructure",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExInfDelete,
	}
	s.oracleExInfDescribeCmd = &cobra.Command{
		Use: "describe INFRASTRUCTURE", Short: "Describe a cloud Exadata infrastructure",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExInfDescribe,
	}
	s.oracleExInfListCmd = &cobra.Command{
		Use: "list", Short: "List cloud Exadata infrastructures in a location",
		Args: cobra.NoArgs, RunE: s.runOracleExInfList,
	}
	s.oracleCloudVmClustersCmd = &cobra.Command{
		Use:   "cloud-vm-clusters",
		Short: "Manage Oracle Cloud VM clusters",
	}
	s.oracleVmClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a cloud VM cluster",
		Args: cobra.ExactArgs(1), RunE: s.runOracleVmClCreate,
	}
	s.oracleVmClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a cloud VM cluster",
		Args: cobra.ExactArgs(1), RunE: s.runOracleVmClDelete,
	}
	s.oracleVmClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a cloud VM cluster",
		Args: cobra.ExactArgs(1), RunE: s.runOracleVmClDescribe,
	}
	s.oracleVmClListCmd = &cobra.Command{
		Use: "list", Short: "List cloud VM clusters in a location",
		Args: cobra.NoArgs, RunE: s.runOracleVmClList,
	}
	s.oracleDatabaseCharacterSetsCmd = &cobra.Command{
		Use:   "database-character-sets",
		Short: "Manage Oracle Database character sets",
	}
	s.oracleDbCsListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle Database character sets in a location",
		Args: cobra.NoArgs, RunE: s.runOracleDbCsList,
	}
	s.oracleDatabasesCmd = &cobra.Command{
		Use:   "databases",
		Short: "Manage Oracle databases",
	}
	s.oracleDbDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Describe an Oracle database",
		Args: cobra.ExactArgs(1), RunE: s.runOracleDbDescribe,
	}
	s.oracleDbListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle databases in a location",
		Args: cobra.NoArgs, RunE: s.runOracleDbList,
	}
	s.oracleDbSystemInitialStorageSizesCmd = &cobra.Command{
		Use:   "db-system-initial-storage-sizes",
		Short: "Manage Oracle DB system initial storage sizes",
	}
	s.oracleSizListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle DB system initial storage sizes in a location",
		Args: cobra.NoArgs, RunE: s.runOracleSizList,
	}
	s.oracleDbSystemShapesCmd = &cobra.Command{
		Use:   "db-system-shapes",
		Short: "Manage Oracle DB system shapes",
	}
	s.oracleShapeListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle DB system shapes in a location",
		Args: cobra.NoArgs, RunE: s.runOracleShapeList,
	}
	s.oracleDbSystemsCmd = &cobra.Command{
		Use:   "db-systems",
		Short: "Manage Oracle DB systems",
	}
	s.oracleDbsCreateCmd = &cobra.Command{
		Use: "create DB_SYSTEM", Short: "Create an Oracle DB system",
		Args: cobra.ExactArgs(1), RunE: s.runOracleDbsCreate,
	}
	s.oracleDbsDeleteCmd = &cobra.Command{
		Use: "delete DB_SYSTEM", Short: "Delete an Oracle DB system",
		Args: cobra.ExactArgs(1), RunE: s.runOracleDbsDelete,
	}
	s.oracleDbsDescribeCmd = &cobra.Command{
		Use: "describe DB_SYSTEM", Short: "Describe an Oracle DB system",
		Args: cobra.ExactArgs(1), RunE: s.runOracleDbsDescribe,
	}
	s.oracleDbsListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle DB systems in a location",
		Args: cobra.NoArgs, RunE: s.runOracleDbsList,
	}
	s.oracleDbVersionsCmd = &cobra.Command{
		Use:   "db-versions",
		Short: "Manage Oracle DB versions",
	}
	s.oracleDbvListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle DB versions in a location",
		Args: cobra.NoArgs, RunE: s.runOracleDbvList,
	}
	s.oracleEntitlementsCmd = &cobra.Command{
		Use:   "entitlements",
		Short: "Manage Oracle Database entitlements",
	}
	s.oracleEntListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle Database entitlements in a location",
		Args: cobra.NoArgs, RunE: s.runOracleEntList,
	}
	s.oracleExadbVmClustersCmd = &cobra.Command{
		Use:   "exadb-vm-clusters",
		Short: "Manage Oracle ExaDB VM clusters",
	}
	s.oracleExaVmCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create an ExaDB VM cluster",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExaVmCreate,
	}
	s.oracleExaVmDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete an ExaDB VM cluster",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExaVmDelete,
	}
	s.oracleExaVmDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe an ExaDB VM cluster",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExaVmDescribe,
	}
	s.oracleExaVmListCmd = &cobra.Command{
		Use: "list", Short: "List ExaDB VM clusters in a location",
		Args: cobra.NoArgs, RunE: s.runOracleExaVmList,
	}
	s.oracleExaVmUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update an ExaDB VM cluster",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExaVmUpdate,
	}
	s.oracleExascaleDbStorageVaultsCmd = &cobra.Command{
		Use:   "exascale-db-storage-vaults",
		Short: "Manage Oracle Exascale DB storage vaults",
	}
	s.oracleExaStCreateCmd = &cobra.Command{
		Use: "create VAULT", Short: "Create an Exascale DB storage vault",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExaStCreate,
	}
	s.oracleExaStDeleteCmd = &cobra.Command{
		Use: "delete VAULT", Short: "Delete an Exascale DB storage vault",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExaStDelete,
	}
	s.oracleExaStDescribeCmd = &cobra.Command{
		Use: "describe VAULT", Short: "Describe an Exascale DB storage vault",
		Args: cobra.ExactArgs(1), RunE: s.runOracleExaStDescribe,
	}
	s.oracleExaStListCmd = &cobra.Command{
		Use: "list", Short: "List Exascale DB storage vaults in a location",
		Args: cobra.NoArgs, RunE: s.runOracleExaStList,
	}
	s.oracleGiVersionsCmd = &cobra.Command{
		Use:   "gi-versions",
		Short: "Manage Oracle GI versions",
	}
	s.oracleGiListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle GI versions in a location",
		Args: cobra.NoArgs, RunE: s.runOracleGiList,
	}
	s.oracleGgConnAssignmentsCmd = &cobra.Command{
		Use:   "goldengate-connection-assignments",
		Short: "Manage Oracle GoldenGate connection assignments",
	}
	s.oracleGgCaCreateCmd = &cobra.Command{
		Use: "create ASSIGNMENT", Short: "Create a GoldenGate connection assignment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgCaCreate,
	}
	s.oracleGgCaDeleteCmd = &cobra.Command{
		Use: "delete ASSIGNMENT", Short: "Delete a GoldenGate connection assignment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgCaDelete,
	}
	s.oracleGgCaDescribeCmd = &cobra.Command{
		Use: "describe ASSIGNMENT", Short: "Describe a GoldenGate connection assignment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgCaDescribe,
	}
	s.oracleGgCaListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate connection assignments in a location",
		Args: cobra.NoArgs, RunE: s.runOracleGgCaList,
	}
	s.oracleGgCaTestCmd = &cobra.Command{
		Use: "test ASSIGNMENT", Short: "Test a GoldenGate connection assignment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgCaTest,
	}
	s.oracleGgConnTypesCmd = &cobra.Command{
		Use:   "goldengate-connection-types",
		Short: "List Oracle GoldenGate connection types",
	}
	s.oracleGgCtListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate connection types in a location",
		Args: cobra.NoArgs, RunE: s.runOracleGgCtList,
	}
	s.oracleGgConnectionsCmd = &cobra.Command{
		Use:   "goldengate-connections",
		Short: "Manage Oracle GoldenGate connections",
	}
	s.oracleGgCnCreateCmd = &cobra.Command{
		Use: "create CONNECTION", Short: "Create a GoldenGate connection",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgCnCreate,
	}
	s.oracleGgCnDeleteCmd = &cobra.Command{
		Use: "delete CONNECTION", Short: "Delete a GoldenGate connection",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgCnDelete,
	}
	s.oracleGgCnDescribeCmd = &cobra.Command{
		Use: "describe CONNECTION", Short: "Describe a GoldenGate connection",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgCnDescribe,
	}
	s.oracleGgCnListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate connections in a location",
		Args: cobra.NoArgs, RunE: s.runOracleGgCnList,
	}
	s.oracleGgDeployEnvsCmd = &cobra.Command{
		Use:   "goldengate-deployment-environments",
		Short: "List Oracle GoldenGate deployment environments",
	}
	s.oracleGgDeCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate deployment environments in a location",
		Args: cobra.NoArgs, RunE: s.runOracleGgDeList,
	}
	s.oracleGgDeployTypesCmd = &cobra.Command{
		Use:   "goldengate-deployment-types",
		Short: "List Oracle GoldenGate deployment types",
	}
	s.oracleGgDtListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate deployment types in a location",
		Args: cobra.NoArgs, RunE: s.runOracleGgDtList,
	}
	s.oracleGgDeployVersionsCmd = &cobra.Command{
		Use:   "goldengate-deployment-versions",
		Short: "List Oracle GoldenGate deployment versions",
	}
	s.oracleGgDvListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate deployment versions in a location",
		Args: cobra.NoArgs, RunE: s.runOracleGgDvList,
	}
	s.oracleGgDeploymentsCmd = &cobra.Command{
		Use:   "goldengate-deployments",
		Short: "Manage Oracle GoldenGate deployments",
	}
	s.oracleGgDpCreateCmd = &cobra.Command{
		Use: "create DEPLOYMENT", Short: "Create a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgDpCreate,
	}
	s.oracleGgDpDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgDpDelete,
	}
	s.oracleGgDpDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgDpDescribe,
	}
	s.oracleGgDpListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate deployments in a location",
		Args: cobra.NoArgs, RunE: s.runOracleGgDpList,
	}
	s.oracleGgDpStartCmd = &cobra.Command{
		Use: "start DEPLOYMENT", Short: "Start a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgDpStart,
	}
	s.oracleGgDpStopCmd = &cobra.Command{
		Use: "stop DEPLOYMENT", Short: "Stop a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: s.runOracleGgDpStop,
	}
	s.oracleOdbNetworksCmd = &cobra.Command{
		Use:   "odb-networks",
		Short: "Manage Oracle ODB networks",
	}
	s.oracleOdbNetCreateCmd = &cobra.Command{
		Use: "create NETWORK", Short: "Create an ODB network",
		Args: cobra.ExactArgs(1), RunE: s.runOracleOdbNetCreate,
	}
	s.oracleOdbNetDeleteCmd = &cobra.Command{
		Use: "delete NETWORK", Short: "Delete an ODB network",
		Args: cobra.ExactArgs(1), RunE: s.runOracleOdbNetDelete,
	}
	s.oracleOdbNetDescribeCmd = &cobra.Command{
		Use: "describe NETWORK", Short: "Describe an ODB network",
		Args: cobra.ExactArgs(1), RunE: s.runOracleOdbNetDescribe,
	}
	s.oracleOdbNetListCmd = &cobra.Command{
		Use: "list", Short: "List ODB networks in a location",
		Args: cobra.NoArgs, RunE: s.runOracleOdbNetList,
	}
	s.oracleOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Oracle Database long-running operations",
	}
	s.oracleOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an Oracle Database operation",
		Args: cobra.ExactArgs(1), RunE: s.runOracleOpCancel,
	}
	s.oracleOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an Oracle Database operation",
		Args: cobra.ExactArgs(1), RunE: s.runOracleOpDelete,
	}
	s.oracleOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an Oracle Database operation",
		Args: cobra.ExactArgs(1), RunE: s.runOracleOpDescribe,
	}
	s.oracleOpListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle Database operations in a location",
		Args: cobra.NoArgs, RunE: s.runOracleOpList,
	}
	s.oraclePluggableDatabasesCmd = &cobra.Command{
		Use:   "pluggable-databases",
		Short: "Manage Oracle pluggable databases",
	}
	s.oraclePdbDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Describe a pluggable database",
		Args: cobra.ExactArgs(1), RunE: s.runOraclePdbDescribe,
	}
	s.oraclePdbListCmd = &cobra.Command{
		Use: "list", Short: "List pluggable databases in a location",
		Args: cobra.NoArgs, RunE: s.runOraclePdbList,
	}
	s.initOracleDatabase()
	s.initOracleDatabaseAutonomousDatabaseBackups()
	s.initOracleDatabaseAutonomousDatabaseCharacterSets()
	s.initOracleDatabaseAutonomousDatabases()
	s.initOracleDatabaseAutonomousDbVersions()
	s.initOracleDatabaseCloudExadataInfrastructures()
	s.initOracleDatabaseCloudVmClusters()
	s.initOracleDatabaseDatabaseCharacterSets()
	s.initOracleDatabaseDatabases()
	s.initOracleDatabaseDbSystemInitialStorageSizes()
	s.initOracleDatabaseDbSystemShapes()
	s.initOracleDatabaseDbSystems()
	s.initOracleDatabaseDbVersions()
	s.initOracleDatabaseEntitlements()
	s.initOracleDatabaseExadbVmClusters()
	s.initOracleDatabaseExascaleDbStorageVaults()
	s.initOracleDatabaseGiVersions()
	s.initOracleDatabaseGoldengateConnectionAssignments()
	s.initOracleDatabaseGoldengateConnectionTypes()
	s.initOracleDatabaseGoldengateConnections()
	s.initOracleDatabaseGoldengateDeploymentEnvironments()
	s.initOracleDatabaseGoldengateDeploymentTypes()
	s.initOracleDatabaseGoldengateDeploymentVersions()
	s.initOracleDatabaseGoldengateDeployments()
	s.initOracleDatabaseOdbNetworks()
	s.initOracleDatabaseOperations()
	s.initOracleDatabasePluggableDatabases()
	return s
}
