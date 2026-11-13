package cmd

import "github.com/spf13/cobra"

// --- gcloud sql (#388) ---
//
// The top-level `sql` command aggregates the Cloud SQL subgroups. All
// subgroups (backups, connect, databases, export, flags, import, instances,
// operations, ssl, ssl-certs, tiers, users) are implemented in sql_all.go.

func (s *sqlCmds) initSql() {

}

// sqlCmds holds the commands and flag values of one invocation's "sql"
// command group, so invocations don't share state (#2070).
type sqlCmds struct {
	inv *invocation

	sqlCmd                                   *cobra.Command
	flagSQLInstance                          string
	flagSQLConfigFile                        string
	flagSQLFormat                            string
	flagSQLAsync                             bool
	flagSQLFilter                            string
	flagSQLLimit                             int64
	flagSQLDescription                       string
	flagSQLLocation                          string
	flagSQLBackup                            string
	flagSQLDatabase                          string
	flagSQLTable                             []string
	flagSQLColumns                           []string
	flagSQLQuery                             string
	flagSQLOffload                           bool
	flagSQLUser                              string
	flagSQLHost                              string
	flagSQLPassword                          string
	flagSQLType                              string
	flagSQLCharset                           string
	flagSQLCollation                         string
	flagSQLDBVersion                         string
	flagSQLTier                              string
	flagSQLRegion                            string
	flagSQLZone                              string
	flagSQLPort                              int
	flagSQLNextVersion                       string
	flagSQLReschedule                        string
	flagSQLScheduleTime                      string
	flagSQLPrivateIP                         bool
	flagSQLPSC                               bool
	flagSQLAutoIP                            bool
	flagSQLSettingsVer                       int64
	flagSQLBAKType                           string
	flagSQLExportDBs                         []string
	flagSQLImportUser                        string
	flagSQLEncryptionKey                     string
	flagSQLCertPath                          string
	flagSQLPvKPath                           string
	flagSQLStorageAutoIncreaseLimit          int64
	sqlBackupsCmd                            *cobra.Command
	sqlConnectCmd                            *cobra.Command
	sqlDatabasesCmd                          *cobra.Command
	sqlExportCmd                             *cobra.Command
	sqlFlagsCmd                              *cobra.Command
	sqlImportCmd                             *cobra.Command
	sqlInstancesCmd                          *cobra.Command
	sqlOperationsCmd                         *cobra.Command
	sqlSSLCmd                                *cobra.Command
	sqlSSLServerCACmd                        *cobra.Command
	sqlSSLCertsCmd                           *cobra.Command
	sqlTiersCmd                              *cobra.Command
	sqlUsersCmd                              *cobra.Command
	sqlBackupCreateCmd                       *cobra.Command
	sqlBackupDeleteCmd                       *cobra.Command
	sqlBackupDescribeCmd                     *cobra.Command
	sqlBackupListCmd                         *cobra.Command
	sqlBackupRestoreCmd                      *cobra.Command
	sqlConnectPsqlCmd                        *cobra.Command
	sqlConnectMysqlCmd                       *cobra.Command
	sqlConnectSqlserverCmd                   *cobra.Command
	sqlDatabaseCreateCmd                     *cobra.Command
	sqlDatabaseDeleteCmd                     *cobra.Command
	sqlDatabaseDescribeCmd                   *cobra.Command
	sqlDatabaseListCmd                       *cobra.Command
	sqlDatabasePatchCmd                      *cobra.Command
	sqlExportSQLCmd                          *cobra.Command
	sqlExportCSVCmd                          *cobra.Command
	sqlExportBAKCmd                          *cobra.Command
	sqlFlagsListCmd                          *cobra.Command
	sqlImportSQLCmd                          *cobra.Command
	sqlImportCSVCmd                          *cobra.Command
	sqlImportBAKCmd                          *cobra.Command
	sqlInstanceCreateCmd                     *cobra.Command
	sqlInstanceDeleteCmd                     *cobra.Command
	sqlInstanceDescribeCmd                   *cobra.Command
	sqlInstanceListCmd                       *cobra.Command
	sqlInstancePatchCmd                      *cobra.Command
	sqlInstanceCloneCmd                      *cobra.Command
	sqlInstanceFailoverCmd                   *cobra.Command
	sqlInstanceRestartCmd                    *cobra.Command
	sqlInstanceRestoreBackupCmd              *cobra.Command
	sqlInstancePromoteReplicaCmd             *cobra.Command
	sqlInstanceStartReplicaCmd               *cobra.Command
	sqlInstanceStopReplicaCmd                *cobra.Command
	sqlInstanceReencryptCmd                  *cobra.Command
	sqlInstanceResetSSLConfigCmd             *cobra.Command
	sqlInstanceImportCmd                     *cobra.Command
	sqlInstanceExportCmd                     *cobra.Command
	sqlInstanceListServerCasCmd              *cobra.Command
	sqlInstanceRescheduleMaintenanceCmd      *cobra.Command
	sqlInstanceAcquireSsrsLeaseCmd           *cobra.Command
	sqlInstanceReleaseSsrsLeaseCmd           *cobra.Command
	sqlInstanceResetAsyncReplicaLagCmd       *cobra.Command
	sqlInstanceVerifyExternalSyncSettingsCmd *cobra.Command
	sqlOpCancelCmd                           *cobra.Command
	sqlOpDescribeCmd                         *cobra.Command
	sqlOpListCmd                             *cobra.Command
	sqlOpWaitCmd                             *cobra.Command
	sqlSSLServerCAListCmd                    *cobra.Command
	sqlSSLServerCAAddCmd                     *cobra.Command
	sqlSSLServerCARemoveCmd                  *cobra.Command
	sqlSSLServerCARotateCmd                  *cobra.Command
	sqlSSLCertCreateCmd                      *cobra.Command
	sqlSSLCertDeleteCmd                      *cobra.Command
	sqlSSLCertDescribeCmd                    *cobra.Command
	sqlSSLCertListCmd                        *cobra.Command
	sqlTiersListCmd                          *cobra.Command
	sqlUserCreateCmd                         *cobra.Command
	sqlUserDeleteCmd                         *cobra.Command
	sqlUserDescribeCmd                       *cobra.Command
	sqlUserListCmd                           *cobra.Command
	sqlUserSetPasswordCmd                    *cobra.Command
	sqlUserPatchCmd                          *cobra.Command
}

// newSqlCmd returns the "sql" command group for inv.
func newSqlCmd(inv *invocation) *cobra.Command {
	return newSqlCmds(inv).sqlCmd
}

// newSqlCmds builds the "sql" command group for inv and returns its state.
func newSqlCmds(inv *invocation) *sqlCmds {
	s := &sqlCmds{inv: inv}
	s.sqlCmd = &cobra.Command{Use: "sql", Short: "Manage Cloud SQL"}
	s.sqlBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Cloud SQL backups"}
	s.sqlConnectCmd = &cobra.Command{Use: "connect", Short: "Connect to a Cloud SQL instance"}
	s.sqlDatabasesCmd = &cobra.Command{Use: "databases", Short: "Manage Cloud SQL databases"}
	s.sqlExportCmd = &cobra.Command{Use: "export", Short: "Export a Cloud SQL instance"}
	s.sqlFlagsCmd = &cobra.Command{Use: "flags", Short: "List Cloud SQL flags"}
	s.sqlImportCmd = &cobra.Command{Use: "import", Short: "Import into a Cloud SQL instance"}
	s.sqlInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Cloud SQL instances"}
	s.sqlOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud SQL operations"}
	s.sqlSSLCmd = &cobra.Command{Use: "ssl", Short: "Manage Cloud SQL SSL"}
	s.sqlSSLServerCACmd = &cobra.Command{Use: "server-ca-certs", Short: "Manage instance server CA certificates"}
	s.sqlSSLCertsCmd = &cobra.Command{Use: "ssl-certs", Short: "(DEPRECATED) Manage Cloud SQL client SSL certificates"}
	s.sqlTiersCmd = &cobra.Command{Use: "tiers", Short: "List Cloud SQL machine tiers"}
	s.sqlUsersCmd = &cobra.Command{Use: "users", Short: "Manage Cloud SQL users"}
	s.sqlBackupCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a Cloud SQL backup",
		Args: cobra.NoArgs, RunE: s.runSQLBackupCreate,
	}
	s.sqlBackupDeleteCmd = &cobra.Command{
		Use: "delete ID", Short: "Delete a Cloud SQL backup",
		Args: cobra.ExactArgs(1), RunE: s.runSQLBackupDelete,
	}
	s.sqlBackupDescribeCmd = &cobra.Command{
		Use: "describe ID", Short: "Describe a Cloud SQL backup",
		Args: cobra.ExactArgs(1), RunE: s.runSQLBackupDescribe,
	}
	s.sqlBackupListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud SQL backups",
		Args: cobra.NoArgs, RunE: s.runSQLBackupList,
	}
	s.sqlBackupRestoreCmd = &cobra.Command{
		Use: "restore BACKUP_ID", Short: "Restore a Cloud SQL instance from a backup",
		Args: cobra.ExactArgs(1), RunE: s.runSQLBackupRestore,
	}
	s.sqlConnectPsqlCmd = &cobra.Command{
		Use: "psql INSTANCE", Short: "Connect to a PostgreSQL Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLConnect("psql", cmd, args)
		},
	}
	s.sqlConnectMysqlCmd = &cobra.Command{
		Use: "mysql INSTANCE", Short: "Connect to a MySQL Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLConnect("mysql", cmd, args)
		},
	}
	s.sqlConnectSqlserverCmd = &cobra.Command{
		Use: "sqlserver INSTANCE", Short: "Connect to a SQL Server Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLConnect("mssql-cli", cmd, args)
		},
	}
	s.sqlDatabaseCreateCmd = &cobra.Command{
		Use: "create DATABASE", Short: "Create a database on a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLDatabaseCreate,
	}
	s.sqlDatabaseDeleteCmd = &cobra.Command{
		Use: "delete DATABASE", Short: "Delete a database from a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLDatabaseDelete,
	}
	s.sqlDatabaseDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Describe a database",
		Args: cobra.ExactArgs(1), RunE: s.runSQLDatabaseDescribe,
	}
	s.sqlDatabaseListCmd = &cobra.Command{
		Use: "list", Short: "List databases on a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: s.runSQLDatabaseList,
	}
	s.sqlDatabasePatchCmd = &cobra.Command{
		Use: "patch DATABASE", Short: "Patch a database on a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLDatabasePatch,
	}
	s.sqlExportSQLCmd = &cobra.Command{
		Use: "sql INSTANCE URI", Short: "Export a Cloud SQL instance as SQL statements to Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLExport("SQL", cmd, args)
		},
	}
	s.sqlExportCSVCmd = &cobra.Command{
		Use: "csv INSTANCE URI --query=QUERY --database=DB", Short: "Export a Cloud SQL instance as CSV to Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLExport("CSV", cmd, args)
		},
	}
	s.sqlExportBAKCmd = &cobra.Command{
		Use: "bak INSTANCE URI --database=DB", Short: "Export a SQL Server instance as a BAK to Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLExport("BAK", cmd, args)
		},
	}
	s.sqlFlagsListCmd = &cobra.Command{
		Use: "list", Short: "List available database flags",
		Args: cobra.NoArgs, RunE: s.runSQLFlagsList,
	}
	s.sqlImportSQLCmd = &cobra.Command{
		Use: "sql INSTANCE URI", Short: "Import a SQL dump from Cloud Storage into a Cloud SQL instance",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLImport("SQL", cmd, args)
		},
	}
	s.sqlImportCSVCmd = &cobra.Command{
		Use: "csv INSTANCE URI --database=DB --table=TABLE", Short: "Import CSV data from Cloud Storage into a Cloud SQL database",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLImport("CSV", cmd, args)
		},
	}
	s.sqlImportBAKCmd = &cobra.Command{
		Use: "bak INSTANCE URI --database=DB", Short: "Import a SQL Server .BAK from Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLImport("BAK", cmd, args)
		},
	}
	s.sqlInstanceCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceCreate,
	}
	s.sqlInstanceDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceDelete,
	}
	s.sqlInstanceDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceDescribe,
	}
	s.sqlInstanceListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud SQL instances in the project",
		Args: cobra.NoArgs, RunE: s.runSQLInstanceList,
	}
	s.sqlInstancePatchCmd = &cobra.Command{
		Use: "patch INSTANCE", Short: "Patch a Cloud SQL instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstancePatch,
	}
	s.sqlInstanceCloneCmd = &cobra.Command{
		Use: "clone SOURCE DESTINATION", Short: "Clone a Cloud SQL instance",
		Args: cobra.ExactArgs(2), RunE: s.runSQLInstanceClone,
	}
	s.sqlInstanceFailoverCmd = &cobra.Command{
		Use: "failover INSTANCE", Short: "Failover a Cloud SQL instance to its replica",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceFailover,
	}
	s.sqlInstanceRestartCmd = &cobra.Command{
		Use: "restart INSTANCE", Short: "Restart a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceRestart,
	}
	s.sqlInstanceRestoreBackupCmd = &cobra.Command{
		Use: "restore-backup INSTANCE", Short: "Restore a Cloud SQL instance from a backup",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceRestoreBackup,
	}
	s.sqlInstancePromoteReplicaCmd = &cobra.Command{
		Use: "promote-replica INSTANCE", Short: "Promote a Cloud SQL replica to a standalone instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstancePromoteReplica,
	}
	s.sqlInstanceStartReplicaCmd = &cobra.Command{
		Use: "start-replica INSTANCE", Short: "Start replication on a Cloud SQL read replica",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceStartReplica,
	}
	s.sqlInstanceStopReplicaCmd = &cobra.Command{
		Use: "stop-replica INSTANCE", Short: "Stop replication on a Cloud SQL read replica",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceStopReplica,
	}
	s.sqlInstanceReencryptCmd = &cobra.Command{
		Use: "reencrypt INSTANCE", Short: "Re-encrypt a CMEK-enabled Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceReencrypt,
	}
	s.sqlInstanceResetSSLConfigCmd = &cobra.Command{
		Use: "reset-ssl-config INSTANCE", Short: "Reset the SSL configuration of a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceResetSSLConfig,
	}
	s.sqlInstanceImportCmd = &cobra.Command{
		Use: "import INSTANCE URI", Short: "Import a SQL dump into a Cloud SQL instance",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLImport("SQL", cmd, args)
		},
	}
	s.sqlInstanceExportCmd = &cobra.Command{
		Use: "export INSTANCE URI", Short: "Export a Cloud SQL instance to Cloud Storage as SQL",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return s.runSQLExport("SQL", cmd, args)
		},
	}
	s.sqlInstanceListServerCasCmd = &cobra.Command{
		Use: "list-server-cas INSTANCE", Short: "List server CA certificates for a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceListServerCas,
	}
	s.sqlInstanceRescheduleMaintenanceCmd = &cobra.Command{
		Use: "reschedule-maintenance INSTANCE", Short: "Reschedule a Cloud SQL instance's maintenance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceRescheduleMaintenance,
	}
	s.sqlInstanceAcquireSsrsLeaseCmd = &cobra.Command{
		Use: "acquire-ssrs-lease INSTANCE", Short: "Acquire an SSRS lease on a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceAcquireSsrsLease,
	}
	s.sqlInstanceReleaseSsrsLeaseCmd = &cobra.Command{
		Use: "release-ssrs-lease INSTANCE", Short: "Release an SSRS lease on a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceReleaseSsrsLease,
	}
	s.sqlInstanceResetAsyncReplicaLagCmd = &cobra.Command{
		Use: "reset-async-replica-lag INSTANCE", Short: "Reset asynchronous replica lag (alias of reset-replica-size)",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceResetAsyncReplicaLag,
	}
	s.sqlInstanceVerifyExternalSyncSettingsCmd = &cobra.Command{
		Use: "verify-external-sync-settings INSTANCE", Short: "Verify external sync settings for a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: s.runSQLInstanceVerifyExternalSyncSettings,
	}
	s.sqlOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Cloud SQL operation",
		Args: cobra.ExactArgs(1), RunE: s.runSQLOpCancel,
	}
	s.sqlOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Cloud SQL operation",
		Args: cobra.ExactArgs(1), RunE: s.runSQLOpDescribe,
	}
	s.sqlOpListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud SQL operations",
		Args: cobra.NoArgs, RunE: s.runSQLOpList,
	}
	s.sqlOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Cloud SQL operation to finish",
		Args: cobra.ExactArgs(1), RunE: s.runSQLOpWait,
	}
	s.sqlSSLServerCAListCmd = &cobra.Command{
		Use: "list", Short: "List server CA certificates for a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: s.runSQLSSLServerCAList,
	}
	s.sqlSSLServerCAAddCmd = &cobra.Command{
		Use: "add", Short: "Add a new server CA certificate to a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: s.runSQLSSLServerCAAdd,
	}
	s.sqlSSLServerCARemoveCmd = &cobra.Command{
		Use: "remove", Short: "Remove a server CA certificate from a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: s.runSQLSSLServerCARemove,
	}
	s.sqlSSLServerCARotateCmd = &cobra.Command{
		Use: "rotate", Short: "Rotate the server CA certificate for a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: s.runSQLSSLServerCARotate,
	}
	s.sqlSSLCertCreateCmd = &cobra.Command{
		Use: "create COMMON_NAME CERT_FILE", Short: "(DEPRECATED) Create a Cloud SQL client SSL certificate",
		Args: cobra.ExactArgs(2), RunE: s.runSQLSSLCertCreate,
	}
	s.sqlSSLCertDeleteCmd = &cobra.Command{
		Use: "delete COMMON_NAME", Short: "(DEPRECATED) Delete a Cloud SQL client SSL certificate",
		Args: cobra.ExactArgs(1), RunE: s.runSQLSSLCertDelete,
	}
	s.sqlSSLCertDescribeCmd = &cobra.Command{
		Use: "describe COMMON_NAME", Short: "(DEPRECATED) Describe a Cloud SQL client SSL certificate",
		Args: cobra.ExactArgs(1), RunE: s.runSQLSSLCertDescribe,
	}
	s.sqlSSLCertListCmd = &cobra.Command{
		Use: "list", Short: "(DEPRECATED) List Cloud SQL client SSL certificates",
		Args: cobra.NoArgs, RunE: s.runSQLSSLCertList,
	}
	s.sqlTiersListCmd = &cobra.Command{
		Use: "list", Short: "List available Cloud SQL tiers",
		Args: cobra.NoArgs, RunE: s.runSQLTiersList,
	}
	s.sqlUserCreateCmd = &cobra.Command{
		Use: "create USERNAME", Short: "Create a Cloud SQL user",
		Args: cobra.ExactArgs(1), RunE: s.runSQLUserCreate,
	}
	s.sqlUserDeleteCmd = &cobra.Command{
		Use: "delete USERNAME", Short: "Delete a Cloud SQL user",
		Args: cobra.ExactArgs(1), RunE: s.runSQLUserDelete,
	}
	s.sqlUserDescribeCmd = &cobra.Command{
		Use: "describe USERNAME", Short: "Describe a Cloud SQL user",
		Args: cobra.ExactArgs(1), RunE: s.runSQLUserDescribe,
	}
	s.sqlUserListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud SQL users on an instance",
		Args: cobra.NoArgs, RunE: s.runSQLUserList,
	}
	s.sqlUserSetPasswordCmd = &cobra.Command{
		Use: "set-password USERNAME", Short: "Set the password for a Cloud SQL user",
		Args: cobra.ExactArgs(1), RunE: s.runSQLUserSetPassword,
	}
	s.sqlUserPatchCmd = &cobra.Command{
		Use: "patch USERNAME", Short: "Patch a Cloud SQL user from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runSQLUserPatch,
	}
	s.initSql()
	s.initSqlAll()
	return s
}
