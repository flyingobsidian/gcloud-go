package cmd

import "github.com/spf13/cobra"

// --- gcloud alloydb (#293) ---

func (s *alloydbCmds) initAlloydb() {
	// All subgroups (backups, clusters, instances, operations, users) are
	// implemented in alloydb_all.go.

}

// alloydbCmds holds the commands and flag values of one invocation's "alloydb"
// command group, so invocations don't share state (#2070).
type alloydbCmds struct {
	inv *invocation

	alloydbCmd        *cobra.Command
	flagADBRegion     string
	flagADBConfigFile string
	flagADBUpdateMask string
	flagADBFormat     string
	flagADBAsync      bool
	flagADBCluster    string
	// #1751 (571.0.0 / 570.0.0): --labels / --update-labels / --remove-labels /
	// --clear-labels on `alloydb instances (create|create-secondary|update)`.
	flagADBLabels       map[string]string
	flagADBUpdateLabels map[string]string
	flagADBRemoveLabels []string
	flagADBClearLabels  bool
	// #1751 (578.0.0): --backupdr-backup / --backupdr-data-source on
	// `alloydb clusters restore` (BackupDR restore flags promoted to GA).
	flagADBBackupDRBackup        string
	flagADBBackupDRDataSource    string
	alloydbBackupsCmd            *cobra.Command
	adbBackupCreateCmd           *cobra.Command
	adbBackupDeleteCmd           *cobra.Command
	adbBackupDescribeCmd         *cobra.Command
	adbBackupListCmd             *cobra.Command
	adbBackupUpdateCmd           *cobra.Command
	alloydbClustersCmd           *cobra.Command
	adbClusterCreateCmd          *cobra.Command
	adbClusterCreateSecondaryCmd *cobra.Command
	adbClusterDeleteCmd          *cobra.Command
	adbClusterDescribeCmd        *cobra.Command
	adbClusterListCmd            *cobra.Command
	adbClusterUpdateCmd          *cobra.Command
	adbClusterPromoteCmd         *cobra.Command
	adbClusterRestoreCmd         *cobra.Command
	adbClusterSwitchoverCmd      *cobra.Command
	adbClusterUpgradeCmd         *cobra.Command
	adbClusterMigrateCmd         *cobra.Command
	adbClusterExportCmd          *cobra.Command
	adbClusterImportCmd          *cobra.Command
	alloydbInstancesCmd          *cobra.Command
	adbInstCreateCmd             *cobra.Command
	adbInstCreateSecondaryCmd    *cobra.Command
	adbInstDeleteCmd             *cobra.Command
	adbInstDescribeCmd           *cobra.Command
	adbInstListCmd               *cobra.Command
	adbInstUpdateCmd             *cobra.Command
	adbInstRestartCmd            *cobra.Command
	adbInstFailoverCmd           *cobra.Command
	adbInstInjectFaultCmd        *cobra.Command
	adbInstGetConnInfoCmd        *cobra.Command
	alloydbUsersCmd              *cobra.Command
	adbUserCreateCmd             *cobra.Command
	adbUserDeleteCmd             *cobra.Command
	adbUserDescribeCmd           *cobra.Command
	adbUserListCmd               *cobra.Command
	adbUserUpdateCmd             *cobra.Command
	alloydbOperationsCmd         *cobra.Command
	adbOpCancelCmd               *cobra.Command
	adbOpDeleteCmd               *cobra.Command
	adbOpDescribeCmd             *cobra.Command
	adbOpListCmd                 *cobra.Command
}

// newAlloydbCmd returns the "alloydb" command group for inv.
func newAlloydbCmd(inv *invocation) *cobra.Command {
	return newAlloydbCmds(inv).alloydbCmd
}

// newAlloydbCmds builds the "alloydb" command group for inv and returns its state.
func newAlloydbCmds(inv *invocation) *alloydbCmds {
	s := &alloydbCmds{inv: inv}
	s.alloydbCmd = &cobra.Command{Use: "alloydb", Short: "Manage AlloyDB"}
	s.alloydbBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage AlloyDB backups"}
	s.adbBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a backup from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBBackupCreate,
	}
	s.adbBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup",
		Args: cobra.ExactArgs(1), RunE: s.runADBBackupDelete,
	}
	s.adbBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup",
		Args: cobra.ExactArgs(1), RunE: s.runADBBackupDescribe,
	}
	s.adbBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups in a region",
		Args: cobra.NoArgs, RunE: s.runADBBackupList,
	}
	s.adbBackupUpdateCmd = &cobra.Command{
		Use: "update BACKUP", Short: "Update a backup from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBBackupUpdate,
	}
	s.alloydbClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage AlloyDB clusters"}
	s.adbClusterCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterCreate,
	}
	s.adbClusterCreateSecondaryCmd = &cobra.Command{
		Use: "create-secondary CLUSTER", Short: "Create a secondary cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterCreateSecondary,
	}
	s.adbClusterDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a cluster",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterDelete,
	}
	s.adbClusterDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a cluster",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterDescribe,
	}
	s.adbClusterListCmd = &cobra.Command{
		Use: "list", Short: "List clusters in a region",
		Args: cobra.NoArgs, RunE: s.runADBClusterList,
	}
	s.adbClusterUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterUpdate,
	}
	s.adbClusterPromoteCmd = &cobra.Command{
		Use: "promote CLUSTER", Short: "Promote a secondary cluster to primary",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterPromote,
	}
	s.adbClusterRestoreCmd = &cobra.Command{
		Use: "restore CLUSTER", Short: "Restore a cluster from a backup/PITR (--config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterRestore,
	}
	s.adbClusterSwitchoverCmd = &cobra.Command{
		Use: "switchover CLUSTER", Short: "Switchover to a secondary cluster",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterSwitchover,
	}
	s.adbClusterUpgradeCmd = &cobra.Command{
		Use: "upgrade CLUSTER", Short: "Upgrade a cluster (--config-file with UpgradeClusterRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterUpgrade,
	}
	s.adbClusterMigrateCmd = &cobra.Command{
		Use: "migrate-cloud-sql CLUSTER", Short: "Migrate from Cloud SQL (--config-file with RestoreFromCloudSQLRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterMigrate,
	}
	s.adbClusterExportCmd = &cobra.Command{
		Use: "export CLUSTER", Short: "Export a cluster (--config-file with ExportClusterRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterExport,
	}
	s.adbClusterImportCmd = &cobra.Command{
		Use: "import CLUSTER", Short: "Import into a cluster (--config-file with ImportClusterRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runADBClusterImport,
	}
	s.alloydbInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage AlloyDB instances"}
	s.adbInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBInstCreate,
	}
	s.adbInstCreateSecondaryCmd = &cobra.Command{
		Use: "create-secondary INSTANCE", Short: "Create a secondary instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBInstCreateSecondary,
	}
	s.adbInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an instance",
		Args: cobra.ExactArgs(1), RunE: s.runADBInstDelete,
	}
	s.adbInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an instance",
		Args: cobra.ExactArgs(1), RunE: s.runADBInstDescribe,
	}
	s.adbInstListCmd = &cobra.Command{
		Use: "list", Short: "List instances in a cluster",
		Args: cobra.NoArgs, RunE: s.runADBInstList,
	}
	s.adbInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBInstUpdate,
	}
	s.adbInstRestartCmd = &cobra.Command{
		Use: "restart INSTANCE", Short: "Restart an instance",
		Args: cobra.ExactArgs(1), RunE: s.runADBInstRestart,
	}
	s.adbInstFailoverCmd = &cobra.Command{
		Use: "failover INSTANCE", Short: "Fail over an instance to a secondary",
		Args: cobra.ExactArgs(1), RunE: s.runADBInstFailover,
	}
	s.adbInstInjectFaultCmd = &cobra.Command{
		Use: "inject-fault INSTANCE", Short: "Inject a fault into an instance (--config-file with InjectFaultRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runADBInstInjectFault,
	}
	s.adbInstGetConnInfoCmd = &cobra.Command{
		Use: "get-connection-info", Short: "Get instance connection info",
		Args: cobra.NoArgs, RunE: s.runADBInstGetConnInfo,
	}
	s.alloydbUsersCmd = &cobra.Command{Use: "users", Short: "Manage AlloyDB users"}
	s.adbUserCreateCmd = &cobra.Command{
		Use: "create USER", Short: "Create a user from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBUserCreate,
	}
	s.adbUserDeleteCmd = &cobra.Command{
		Use: "delete USER", Short: "Delete a user",
		Args: cobra.ExactArgs(1), RunE: s.runADBUserDelete,
	}
	s.adbUserDescribeCmd = &cobra.Command{
		Use: "describe USER", Short: "Describe a user",
		Args: cobra.ExactArgs(1), RunE: s.runADBUserDescribe,
	}
	s.adbUserListCmd = &cobra.Command{
		Use: "list", Short: "List users in a cluster",
		Args: cobra.NoArgs, RunE: s.runADBUserList,
	}
	s.adbUserUpdateCmd = &cobra.Command{
		Use: "update USER", Short: "Update a user from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runADBUserUpdate,
	}
	s.alloydbOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage AlloyDB operations"}
	s.adbOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: s.runADBOpCancel,
	}
	s.adbOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: s.runADBOpDelete,
	}
	s.adbOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runADBOpDescribe,
	}
	s.adbOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a region",
		Args: cobra.NoArgs, RunE: s.runADBOpList,
	}
	s.initAlloydb()
	s.initAlloydbAll()
	s.initAlloydbAll2()
	s.initAlloydbAll3()
	s.initAlloydbAll4()
	s.initAlloydbAll5()
	return s
}
