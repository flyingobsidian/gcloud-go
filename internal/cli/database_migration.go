package cli

import "github.com/spf13/cobra"

// --- gcloud database-migration (#322) ---

func (s *databaseMigrationCmds) initDatabaseMigration() {
	// Subgroups are implemented in dedicated files:
	//   - connection-profiles: database_migration_connection_profiles.go (#782)
	//   - conversion-workspaces: database_migration_conversion_workspaces.go (#783)
	//   - migration-jobs: database_migration_migration_jobs.go (#784)
	//   - objects: database_migration_objects.go (#785)
	//   - operations: database_migration_operations.go (#786)
	//   - private-connections: database_migration_private_connections.go (#787)

}

// databaseMigrationCmds holds the commands and flag values of one invocation's "databaseMigration"
// command group, so invocations don't share state (#2070).
type databaseMigrationCmds struct {
	inv *invocation

	databaseMigrationCmd        *cobra.Command
	dmConnProfilesCmd           *cobra.Command
	dmCPDescribeCmd             *cobra.Command
	dmCPListCmd                 *cobra.Command
	dmCPDeleteCmd               *cobra.Command
	dmCPCreateCmd               *cobra.Command
	dmCPUpdateCmd               *cobra.Command
	dmCPTestCmd                 *cobra.Command
	dmCPFetchStaticIPsCmd       *cobra.Command
	flagDMCPRegion              string
	flagDMCPFormat              string
	flagDMCPListPageSize        int64
	flagDMCPListLimit           int64
	flagDMCPListFilter          string
	flagDMCPListURI             bool
	flagDMCPConfigFile          string
	flagDMCPUpdateMask          string
	flagDMCPSkipValidate        bool
	flagDMCPAsync               bool
	flagDMCPFetchReservedPublic bool
	dmCWCmd                     *cobra.Command
	dmCWCreateCmd               *cobra.Command
	dmCWDeleteCmd               *cobra.Command
	dmCWDescribeCmd             *cobra.Command
	dmCWUpdateCmd               *cobra.Command
	dmCWListCmd                 *cobra.Command
	dmCWApplyCmd                *cobra.Command
	dmCWCommitCmd               *cobra.Command
	dmCWConvertCmd              *cobra.Command
	dmCWRollbackCmd             *cobra.Command
	dmCWSeedCmd                 *cobra.Command
	dmCWDescribeDDLsCmd         *cobra.Command
	dmCWDescribeEntitiesCmd     *cobra.Command
	dmCWDescribeIssuesCmd       *cobra.Command
	dmCWImportRulesCmd          *cobra.Command
	dmCWListBackgroundJobsCmd   *cobra.Command
	dmCWMappingRulesCmd         *cobra.Command
	dmCWMRListCmd               *cobra.Command
	dmCWMRDescribeCmd           *cobra.Command
	dmCWMRDeleteCmd             *cobra.Command
	dmCWMRCreateCmd             *cobra.Command
	flagDMCWRegion              string
	flagDMCWFormat              string
	flagDMCWConfigFile          string
	flagDMCWUpdateMask          string
	flagDMCWListPageSize        int64
	flagDMCWListLimit           int64
	flagDMCWListFilter          string
	flagDMCWListURI             bool
	flagDMCWAsync               bool
	flagDMCWApplyAutoCommit     bool
	flagDMCWAutoCommit          bool
	flagDMCWNoAutoCommit        bool
	flagDMCWDryRun              bool
	flagDMCWFilter              string
	flagDMCWDestConnProfile     string
	flagDMCWSourceConnProfile   string
	flagDMCWConnProfile         string
	flagDMCWCommitName          string
	flagDMCWConvertFullPath     bool
	flagDMCWDescribeCommitID    string
	flagDMCWDescribeTree        string
	flagDMCWDescribeView        string
	flagDMCWDescribeUncomm      bool
	flagDMCWDescribeDBName      string
	flagDMCWRulesFormat         string
	flagDMCWRulesFiles          []string
	flagDMCWWorkspace           string
	dmMJCmd                     *cobra.Command
	dmMJCreateCmd               *cobra.Command
	dmMJDeleteCmd               *cobra.Command
	dmMJDescribeCmd             *cobra.Command
	dmMJUpdateCmd               *cobra.Command
	dmMJListCmd                 *cobra.Command
	dmMJStartCmd                *cobra.Command
	dmMJStopCmd                 *cobra.Command
	dmMJResumeCmd               *cobra.Command
	dmMJRestartCmd              *cobra.Command
	dmMJPromoteCmd              *cobra.Command
	dmMJVerifyCmd               *cobra.Command
	dmMJDemoteDestCmd           *cobra.Command
	dmMJGenerateSSHCmd          *cobra.Command
	dmMJFetchObjectsCmd         *cobra.Command
	flagDMMJRegion              string
	flagDMMJFormat              string
	flagDMMJConfigFile          string
	flagDMMJUpdateMask          string
	flagDMMJListPageSize        int64
	flagDMMJListLimit           int64
	flagDMMJListFilter          string
	flagDMMJListURI             bool
	flagDMMJAsync               bool
	flagDMMJSkipValidate        bool
	flagDMMJRestartFailed       bool
	flagDMMJDryRun              bool
	flagDMMJSSHVM               string
	flagDMMJSSHVMPort           int64
	dmObjectsCmd                *cobra.Command
	dmObjListCmd                *cobra.Command
	dmObjLookupCmd              *cobra.Command
	flagDMObjRegion             string
	flagDMObjMigrationJob       string
	flagDMObjFormat             string
	flagDMObjListPageSize       int64
	flagDMObjListLimit          int64
	flagDMObjListFilter         string
	flagDMObjListURI            bool
	flagDMObjLookupDatabase     string
	flagDMObjLookupSchema       string
	flagDMObjLookupTable        string
	flagDMObjLookupType         string
	dmOperationsCmd             *cobra.Command
	dmOpDescribeCmd             *cobra.Command
	dmOpDeleteCmd               *cobra.Command
	dmOpListCmd                 *cobra.Command
	flagDMOpRegion              string
	flagDMOpFormat              string
	flagDMOpListFilter          string
	flagDMOpListPageSize        int64
	flagDMOpListLimit           int64
	flagDMOpListURI             bool
	dmPrivateConnCmd            *cobra.Command
	dmPCDescribeCmd             *cobra.Command
	dmPCListCmd                 *cobra.Command
	dmPCDeleteCmd               *cobra.Command
	dmPCCreateCmd               *cobra.Command
	flagDMPCRegion              string
	flagDMPCFormat              string
	flagDMPCListPageSize        int64
	flagDMPCListLimit           int64
	flagDMPCListFilter          string
	flagDMPCListURI             bool
	flagDMPCConfigFile          string
	flagDMPCSkipValidate        bool
	flagDMPCAsync               bool
}

// newDatabaseMigrationCmd returns the "databaseMigration" command group for inv.
func newDatabaseMigrationCmd(inv *invocation) *cobra.Command {
	return newDatabaseMigrationCmds(inv).databaseMigrationCmd
}

// newDatabaseMigrationCmds builds the "databaseMigration" command group for inv and returns its state.
func newDatabaseMigrationCmds(inv *invocation) *databaseMigrationCmds {
	s := &databaseMigrationCmds{inv: inv}
	s.databaseMigrationCmd = &cobra.Command{Use: "database-migration", Short: "Manage Database Migration Service"}
	s.dmConnProfilesCmd = &cobra.Command{
		Use:   "connection-profiles",
		Short: "Manage Database Migration Service connection profiles",
	}
	s.dmCPDescribeCmd = &cobra.Command{
		Use:   "describe PROFILE",
		Short: "Show details about a connection profile",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMCPDescribe,
	}
	s.dmCPListCmd = &cobra.Command{
		Use:   "list",
		Short: "List connection profiles in a region",
		Args:  cobra.NoArgs,
		RunE:  s.runDMCPList,
	}
	s.dmCPDeleteCmd = &cobra.Command{
		Use:   "delete PROFILE",
		Short: "Delete a connection profile",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMCPDelete,
	}
	s.dmCPCreateCmd = &cobra.Command{
		Use:   "create PROFILE",
		Short: "Create a connection profile from a --config-file",
		Long: `Create a Database Migration Service connection profile.

The --config-file argument points at a JSON or YAML file containing a
ConnectionProfile message body. gcloud-python exposes this command as a
per-engine group (create mysql, create postgresql, etc.); the generic form
here accepts any engine by loading its message body from disk. See:
https://cloud.google.com/database-migration/docs/reference/rest/v1/projects.locations.connectionProfiles#ConnectionProfile
`,
		Args: cobra.ExactArgs(1),
		RunE: s.runDMCPCreate,
	}
	s.dmCPUpdateCmd = &cobra.Command{
		Use:   "update PROFILE",
		Short: "Update a connection profile from a --config-file",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMCPUpdate,
	}
	s.dmCPTestCmd = &cobra.Command{
		Use:   "test PROFILE",
		Short: "Test a connection profile without persisting changes",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMCPTest,
	}
	s.dmCPFetchStaticIPsCmd = &cobra.Command{
		Use:   "fetch-static-ips",
		Short: "Fetch static IPs the service uses to establish outbound connections",
		Args:  cobra.NoArgs,
		RunE:  s.runDMCPFetchStaticIPs,
	}
	s.dmCWCmd = &cobra.Command{
		Use:   "conversion-workspaces",
		Short: "Manage Database Migration Service conversion workspaces",
	}
	s.dmCWCreateCmd = &cobra.Command{
		Use: "create WORKSPACE", Short: "Create a conversion workspace from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWCreate,
	}
	s.dmCWDeleteCmd = &cobra.Command{
		Use: "delete WORKSPACE", Short: "Delete a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWDelete,
	}
	s.dmCWDescribeCmd = &cobra.Command{
		Use: "describe WORKSPACE", Short: "Show details about a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWDescribe,
	}
	s.dmCWUpdateCmd = &cobra.Command{
		Use: "update WORKSPACE", Short: "Update a conversion workspace from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWUpdate,
	}
	s.dmCWListCmd = &cobra.Command{
		Use: "list", Short: "List conversion workspaces in a region",
		Args: cobra.NoArgs, RunE: s.runDMCWList,
	}
	s.dmCWApplyCmd = &cobra.Command{
		Use: "apply WORKSPACE", Short: "Apply the conversion workspace to the destination",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWApply,
	}
	s.dmCWCommitCmd = &cobra.Command{
		Use: "commit WORKSPACE", Short: "Commit a conversion workspace revision",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWCommit,
	}
	s.dmCWConvertCmd = &cobra.Command{
		Use: "convert WORKSPACE", Short: "Convert a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWConvert,
	}
	s.dmCWRollbackCmd = &cobra.Command{
		Use: "rollback WORKSPACE", Short: "Rollback the conversion workspace to the previous commit",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWRollback,
	}
	s.dmCWSeedCmd = &cobra.Command{
		Use: "seed WORKSPACE", Short: "Seed a conversion workspace from a connection profile",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWSeed,
	}
	s.dmCWDescribeDDLsCmd = &cobra.Command{
		Use: "describe-ddls WORKSPACE", Short: "Describe the DDLs for a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWDescribeDDLs,
	}
	s.dmCWDescribeEntitiesCmd = &cobra.Command{
		Use: "describe-entities WORKSPACE", Short: "Describe entities in a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWDescribeEntities,
	}
	s.dmCWDescribeIssuesCmd = &cobra.Command{
		Use: "describe-issues WORKSPACE", Short: "Describe conversion issues in a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWDescribeIssues,
	}
	s.dmCWImportRulesCmd = &cobra.Command{
		Use: "import-rules WORKSPACE", Short: "Import mapping rules into a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWImportRules,
	}
	s.dmCWListBackgroundJobsCmd = &cobra.Command{
		Use: "list-background-jobs WORKSPACE", Short: "List background jobs for a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWListBackgroundJobs,
	}
	s.dmCWMappingRulesCmd = &cobra.Command{
		Use:   "mapping-rules",
		Short: "Manage mapping rules for a conversion workspace",
	}
	s.dmCWMRListCmd = &cobra.Command{
		Use: "list", Short: "List mapping rules in a conversion workspace",
		Args: cobra.NoArgs, RunE: s.runDMCWMRList,
	}
	s.dmCWMRDescribeCmd = &cobra.Command{
		Use: "describe RULE", Short: "Show a mapping rule",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWMRDescribe,
	}
	s.dmCWMRDeleteCmd = &cobra.Command{
		Use: "delete RULE", Short: "Delete a mapping rule",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWMRDelete,
	}
	s.dmCWMRCreateCmd = &cobra.Command{
		Use: "create RULE", Short: "Create a mapping rule from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDMCWMRCreate,
	}
	s.dmMJCmd = &cobra.Command{
		Use:   "migration-jobs",
		Short: "Manage Database Migration Service migration jobs",
	}
	s.dmMJCreateCmd = &cobra.Command{
		Use: "create JOB", Short: "Create a migration job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJCreate,
	}
	s.dmMJDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJDelete,
	}
	s.dmMJDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Show details about a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJDescribe,
	}
	s.dmMJUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update a migration job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJUpdate,
	}
	s.dmMJListCmd = &cobra.Command{
		Use: "list", Short: "List migration jobs in a region",
		Args: cobra.NoArgs, RunE: s.runDMMJList,
	}
	s.dmMJStartCmd = &cobra.Command{
		Use: "start JOB", Short: "Start a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJStart,
	}
	s.dmMJStopCmd = &cobra.Command{
		Use: "stop JOB", Short: "Stop a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJStop,
	}
	s.dmMJResumeCmd = &cobra.Command{
		Use: "resume JOB", Short: "Resume a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJResume,
	}
	s.dmMJRestartCmd = &cobra.Command{
		Use: "restart JOB", Short: "Restart a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJRestart,
	}
	s.dmMJPromoteCmd = &cobra.Command{
		Use: "promote JOB", Short: "Promote a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJPromote,
	}
	s.dmMJVerifyCmd = &cobra.Command{
		Use: "verify JOB", Short: "Verify a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJVerify,
	}
	s.dmMJDemoteDestCmd = &cobra.Command{
		Use: "demote-destination JOB", Short: "Demote the destination of a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJDemoteDestination,
	}
	s.dmMJGenerateSSHCmd = &cobra.Command{
		Use: "generate-ssh-script JOB", Short: "Generate the SSH script for a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJGenerateSSHScript,
	}
	s.dmMJFetchObjectsCmd = &cobra.Command{
		Use: "fetch-source-objects JOB", Short: "Fetch the source database objects of a migration job",
		Args: cobra.ExactArgs(1), RunE: s.runDMMJFetchSourceObjects,
	}
	s.dmObjectsCmd = &cobra.Command{
		Use:   "objects",
		Short: "Manage migration job objects",
	}
	s.dmObjListCmd = &cobra.Command{
		Use:   "list",
		Short: "List objects belonging to a migration job",
		Args:  cobra.NoArgs,
		RunE:  s.runDMObjList,
	}
	s.dmObjLookupCmd = &cobra.Command{
		Use:   "lookup",
		Short: "Look up a migration job object by its source identifier",
		Args:  cobra.NoArgs,
		RunE:  s.runDMObjLookup,
	}
	s.dmOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Database Migration Service operations",
	}
	s.dmOpDescribeCmd = &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a Database Migration operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMOpDescribe,
	}
	s.dmOpDeleteCmd = &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a Database Migration operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMOpDelete,
	}
	s.dmOpListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Database Migration operations in a region",
		Args:  cobra.NoArgs,
		RunE:  s.runDMOpList,
	}
	s.dmPrivateConnCmd = &cobra.Command{
		Use:   "private-connections",
		Short: "Manage Database Migration Service private connections",
	}
	s.dmPCDescribeCmd = &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details about a private connection",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMPCDescribe,
	}
	s.dmPCListCmd = &cobra.Command{
		Use:   "list",
		Short: "List private connections in a region",
		Args:  cobra.NoArgs,
		RunE:  s.runDMPCList,
	}
	s.dmPCDeleteCmd = &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a private connection",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMPCDelete,
	}
	s.dmPCCreateCmd = &cobra.Command{
		Use:   "create NAME",
		Short: "Create a private connection from a --config-file",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDMPCCreate,
	}
	s.initDatabaseMigration()
	s.initDatabaseMigrationConnectionProfiles()
	s.initDatabaseMigrationConversionWorkspaces()
	s.initDatabaseMigrationMigrationJobs()
	s.initDatabaseMigrationObjects()
	s.initDatabaseMigrationOperations()
	s.initDatabaseMigrationPrivateConnections()
	return s
}
