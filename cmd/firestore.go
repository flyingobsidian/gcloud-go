package cmd

import "github.com/spf13/cobra"

// --- gcloud firestore (#341) ---

func (g *firestoreCmds) initFirestore() {
	// Subgroups (databases, backups, fields, indexes, locations, operations,
	// user-creds) are implemented in dedicated firestore_*.go files. The three
	// top-level bulk-delete / export / import commands remain stubs pending
	// their own implementations.
	for _, name := range []string{"bulk-delete", "export", "import"} {
		registerStubCommand(g.firestoreCmd, name, "Not yet implemented")
	}

}

// firestoreCmds holds the commands and flag values of one invocation's "firestore"
// command group, so invocations don't share state (#2070).
type firestoreCmds struct {
	inv *invocation

	firestoreCmd                 *cobra.Command
	firestoreBackupsCmd          *cobra.Command
	fsBackupDescribeCmd          *cobra.Command
	fsBackupListCmd              *cobra.Command
	fsBackupDeleteCmd            *cobra.Command
	flagFSBackupLocation         string
	flagFSBackupFormat           string
	flagFSBackupListLoc          string
	firestoreBackupsSchedulesCmd *cobra.Command
	fsBSCreateCmd                *cobra.Command
	fsBSDeleteCmd                *cobra.Command
	fsBSDescribeCmd              *cobra.Command
	fsBSListCmd                  *cobra.Command
	fsBSUpdateCmd                *cobra.Command
	flagFSBSDatabase             string
	flagFSBSConfigFile           string
	flagFSBSUpdateMask           string
	flagFSBSFormat               string
	firestoreChangeStreamsCmd    *cobra.Command
	fsCSCreateCmd                *cobra.Command
	fsCSDeleteCmd                *cobra.Command
	fsCSDescribeCmd              *cobra.Command
	fsCSListCmd                  *cobra.Command
	flagFSCSDatabase             string
	flagFSCSFormat               string
	flagFSCSRetention            string
	flagFSCSDatabaseScope        bool
	flagFSCSCollectionGroup      string
	flagFSCSEtag                 string
	firestoreDatabasesCmd        *cobra.Command
	fsDBCreateCmd                *cobra.Command
	fsDBDeleteCmd                *cobra.Command
	fsDBDescribeCmd              *cobra.Command
	fsDBListCmd                  *cobra.Command
	fsDBUpdateCmd                *cobra.Command
	fsDBRestoreCmd               *cobra.Command
	fsDBCloneCmd                 *cobra.Command
	fsDBConnStrCmd               *cobra.Command
	flagFSDBConfigFile           string
	flagFSDBUpdateMask           string
	flagFSDBFormat               string
	flagFSDBAsync                bool
	flagFSDBListPage             int64
	flagFSDBListLimit            int64
	flagFSDBListShowDeleted      bool
	flagFSDBRestoreDest          string
	flagFSDBRestoreBackup        string
	flagFSDBRestoreKMS           string
	flagFSDBCloneSource          string
	flagFSDBCloneDest            string
	flagFSDBCloneSnapshotTime    string
	flagFSDBCloneKMS             string
	firestoreFieldsCmd           *cobra.Command
	firestoreFieldsTTLsCmd       *cobra.Command
	fsFieldsTTLListCmd           *cobra.Command
	fsFieldsTTLUpdateCmd         *cobra.Command
	flagFSFieldsDatabase         string
	flagFSFieldsCollection       string
	flagFSFieldsConfigFile       string
	flagFSFieldsUpdateMask       string
	flagFSFieldsFormat           string
	flagFSFieldsAsync            bool
	firestoreIndexesCmd          *cobra.Command
	firestoreIndexesCompositeCmd *cobra.Command
	firestoreIndexesFieldsCmd    *cobra.Command
	fsIdxCompCreateCmd           *cobra.Command
	fsIdxCompDeleteCmd           *cobra.Command
	fsIdxCompDescribeCmd         *cobra.Command
	fsIdxCompListCmd             *cobra.Command
	fsIdxFieldsDescribeCmd       *cobra.Command
	fsIdxFieldsListCmd           *cobra.Command
	fsIdxFieldsUpdateCmd         *cobra.Command
	flagFSIdxDatabase            string
	flagFSIdxCollection          string
	flagFSIdxConfigFile          string
	flagFSIdxUpdateMask          string
	flagFSIdxFormat              string
	flagFSIdxAsync               bool
	firestoreLocationsCmd        *cobra.Command
	fsLocListCmd                 *cobra.Command
	fsLocDescribeCmd             *cobra.Command
	flagFSLocFormat              string
	flagFSLocPageSize            int64
	flagFSLocLimit               int64
	firestoreOperationsCmd       *cobra.Command
	fsOpDescribeCmd              *cobra.Command
	fsOpCancelCmd                *cobra.Command
	fsOpDeleteCmd                *cobra.Command
	fsOpListCmd                  *cobra.Command
	flagFSOpDatabase             string
	flagFSOpFormat               string
	flagFSOpFilter               string
	flagFSOpPageSize             int64
	flagFSOpLimit                int64
	firestoreUserCredsCmd        *cobra.Command
	fsUCCreateCmd                *cobra.Command
	fsUCDeleteCmd                *cobra.Command
	fsUCDescribeCmd              *cobra.Command
	fsUCListCmd                  *cobra.Command
	fsUCEnableCmd                *cobra.Command
	fsUCDisableCmd               *cobra.Command
	fsUCResetPasswordCmd         *cobra.Command
	flagFSUCDatabase             string
	flagFSUCFormat               string
}

// newFirestoreCmd returns the "firestore" command group for inv.
func newFirestoreCmd(inv *invocation) *cobra.Command {
	return newFirestoreCmds(inv).firestoreCmd
}

// newFirestoreCmds builds the "firestore" command group for inv and returns its state.
func newFirestoreCmds(inv *invocation) *firestoreCmds {
	g := &firestoreCmds{inv: inv}
	g.firestoreCmd = &cobra.Command{Use: "firestore", Short: "Manage Cloud Firestore"}
	g.firestoreBackupsCmd = &cobra.Command{
		Use:   "backups",
		Short: "Manage Cloud Firestore backups",
	}
	g.fsBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a Firestore backup",
		Args: cobra.ExactArgs(1), RunE: g.runFSBackupDescribe,
	}
	g.fsBackupListCmd = &cobra.Command{
		Use: "list", Short: "List Firestore backups",
		Args: cobra.NoArgs, RunE: g.runFSBackupList,
	}
	g.fsBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a Firestore backup",
		Args: cobra.ExactArgs(1), RunE: g.runFSBackupDelete,
	}
	g.firestoreBackupsSchedulesCmd = &cobra.Command{
		Use:   "schedules",
		Short: "Manage backup schedules for a Firestore database",
	}
	g.fsBSCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a backup schedule from a --config-file",
		Args: cobra.NoArgs, RunE: g.runFSBSCreate,
	}
	g.fsBSDeleteCmd = &cobra.Command{
		Use: "delete SCHEDULE", Short: "Delete a backup schedule",
		Args: cobra.ExactArgs(1), RunE: g.runFSBSDelete,
	}
	g.fsBSDescribeCmd = &cobra.Command{
		Use: "describe SCHEDULE", Short: "Describe a backup schedule",
		Args: cobra.ExactArgs(1), RunE: g.runFSBSDescribe,
	}
	g.fsBSListCmd = &cobra.Command{
		Use: "list", Short: "List backup schedules for a database",
		Args: cobra.NoArgs, RunE: g.runFSBSList,
	}
	g.fsBSUpdateCmd = &cobra.Command{
		Use: "update SCHEDULE", Short: "Update a backup schedule from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runFSBSUpdate,
	}
	g.firestoreChangeStreamsCmd = &cobra.Command{
		Use:   "change-streams",
		Short: "Manage Cloud Firestore ChangeStreams",
		Long: "Manage Cloud Firestore ChangeStreams.\n\n" +
			"Change streams are only available on Enterprise edition databases.",
	}
	g.fsCSCreateCmd = &cobra.Command{
		Use: "create CHANGE_STREAM", Short: "Create a Cloud Firestore change stream",
		Args: cobra.ExactArgs(1), RunE: g.runFSCSCreate,
	}
	g.fsCSDeleteCmd = &cobra.Command{
		Use: "delete CHANGE_STREAM", Short: "Delete a Cloud Firestore change stream",
		Args: cobra.ExactArgs(1), RunE: g.runFSCSDelete,
	}
	g.fsCSDescribeCmd = &cobra.Command{
		Use: "describe CHANGE_STREAM", Short: "Describe a Cloud Firestore change stream",
		Args: cobra.ExactArgs(1), RunE: g.runFSCSDescribe,
	}
	g.fsCSListCmd = &cobra.Command{
		Use: "list", Short: "List change streams under a Firestore database",
		Args: cobra.NoArgs, RunE: g.runFSCSList,
	}
	g.firestoreDatabasesCmd = &cobra.Command{
		Use:   "databases",
		Short: "Manage Cloud Firestore databases",
	}
	g.fsDBCreateCmd = &cobra.Command{
		Use: "create DATABASE", Short: "Create a Firestore database from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runFSDBCreate,
	}
	g.fsDBDeleteCmd = &cobra.Command{
		Use: "delete DATABASE", Short: "Delete a Firestore database",
		Args: cobra.ExactArgs(1), RunE: g.runFSDBDelete,
	}
	g.fsDBDescribeCmd = &cobra.Command{
		Use: "describe [DATABASE]", Short: "Describe a Firestore database",
		Args: cobra.MaximumNArgs(1), RunE: g.runFSDBDescribe,
	}
	g.fsDBListCmd = &cobra.Command{
		Use: "list", Short: "List Firestore databases in the project",
		Args: cobra.NoArgs, RunE: g.runFSDBList,
	}
	g.fsDBUpdateCmd = &cobra.Command{
		Use: "update DATABASE", Short: "Update a Firestore database from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runFSDBUpdate,
	}
	g.fsDBRestoreCmd = &cobra.Command{
		Use: "restore", Short: "Restore a Firestore database from a backup",
		Args: cobra.NoArgs, RunE: g.runFSDBRestore,
	}
	g.fsDBCloneCmd = &cobra.Command{
		Use: "clone", Short: "Clone a Firestore database from another",
		Args: cobra.NoArgs, RunE: g.runFSDBClone,
	}
	g.fsDBConnStrCmd = &cobra.Command{
		Use: "connection-string [DATABASE]", Short: "Print a mongo connection string for the database",
		Args: cobra.MaximumNArgs(1), RunE: g.runFSDBConnectionString,
	}
	g.firestoreFieldsCmd = &cobra.Command{
		Use:   "fields",
		Short: "Manage single-field indexes for Cloud Firestore",
	}
	g.firestoreFieldsTTLsCmd = &cobra.Command{
		Use:   "ttls",
		Short: "Manage Time-to-Live metadata for Cloud Firestore fields",
	}
	g.fsFieldsTTLListCmd = &cobra.Command{
		Use: "list", Short: "List TTL-configured fields for a collection group",
		Args: cobra.NoArgs, RunE: g.runFSFieldsTTLList,
	}
	g.fsFieldsTTLUpdateCmd = &cobra.Command{
		Use: "update FIELD", Short: "Update a field's TTL configuration from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runFSFieldsTTLUpdate,
	}
	g.firestoreIndexesCmd = &cobra.Command{
		Use:   "indexes",
		Short: "Manage Cloud Firestore indexes",
	}
	g.firestoreIndexesCompositeCmd = &cobra.Command{
		Use:   "composite",
		Short: "Manage composite indexes for Cloud Firestore",
	}
	g.firestoreIndexesFieldsCmd = &cobra.Command{
		Use:   "fields",
		Short: "Manage single-field indexes for Cloud Firestore",
	}
	g.fsIdxCompCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a composite index from a --config-file",
		Args: cobra.NoArgs, RunE: g.runFSIdxCompCreate,
	}
	g.fsIdxCompDeleteCmd = &cobra.Command{
		Use: "delete INDEX", Short: "Delete a composite index",
		Args: cobra.ExactArgs(1), RunE: g.runFSIdxCompDelete,
	}
	g.fsIdxCompDescribeCmd = &cobra.Command{
		Use: "describe INDEX", Short: "Describe a composite index",
		Args: cobra.ExactArgs(1), RunE: g.runFSIdxCompDescribe,
	}
	g.fsIdxCompListCmd = &cobra.Command{
		Use: "list", Short: "List composite indexes for a collection group",
		Args: cobra.NoArgs, RunE: g.runFSIdxCompList,
	}
	g.fsIdxFieldsDescribeCmd = &cobra.Command{
		Use: "describe FIELD", Short: "Describe a single-field index",
		Args: cobra.ExactArgs(1), RunE: g.runFSIdxFieldsDescribe,
	}
	g.fsIdxFieldsListCmd = &cobra.Command{
		Use: "list", Short: "List single-field indexes for a collection group",
		Args: cobra.NoArgs, RunE: g.runFSIdxFieldsList,
	}
	g.fsIdxFieldsUpdateCmd = &cobra.Command{
		Use: "update FIELD", Short: "Update a single-field index from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runFSIdxFieldsUpdate,
	}
	g.firestoreLocationsCmd = &cobra.Command{
		Use:   "locations",
		Short: "Explore Cloud Firestore locations",
	}
	g.fsLocListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Firestore-enabled locations for the project",
		Args:  cobra.NoArgs,
		RunE:  g.runFSLocList,
	}
	g.fsLocDescribeCmd = &cobra.Command{
		Use:   "describe LOCATION",
		Short: "Describe a Firestore location",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runFSLocDescribe,
	}
	g.firestoreOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Firestore admin operations",
	}
	g.fsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Firestore admin operation",
		Args: cobra.ExactArgs(1), RunE: g.runFSOpDescribe,
	}
	g.fsOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a running Firestore admin operation",
		Args: cobra.ExactArgs(1), RunE: g.runFSOpCancel,
	}
	g.fsOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a completed Firestore admin operation",
		Args: cobra.ExactArgs(1), RunE: g.runFSOpDelete,
	}
	g.fsOpListCmd = &cobra.Command{
		Use: "list", Short: "List Firestore admin operations for a database",
		Args: cobra.NoArgs, RunE: g.runFSOpList,
	}
	g.firestoreUserCredsCmd = &cobra.Command{
		Use:   "user-creds",
		Short: "Manage Cloud Firestore user credentials",
	}
	g.fsUCCreateCmd = &cobra.Command{
		Use: "create USER_CREDS", Short: "Create a new set of user credentials",
		Args: cobra.ExactArgs(1), RunE: g.runFSUCCreate,
	}
	g.fsUCDeleteCmd = &cobra.Command{
		Use: "delete USER_CREDS", Short: "Delete user credentials",
		Args: cobra.ExactArgs(1), RunE: g.runFSUCDelete,
	}
	g.fsUCDescribeCmd = &cobra.Command{
		Use: "describe USER_CREDS", Short: "Describe user credentials",
		Args: cobra.ExactArgs(1), RunE: g.runFSUCDescribe,
	}
	g.fsUCListCmd = &cobra.Command{
		Use: "list", Short: "List user credentials for a database",
		Args: cobra.NoArgs, RunE: g.runFSUCList,
	}
	g.fsUCEnableCmd = &cobra.Command{
		Use: "enable USER_CREDS", Short: "Enable disabled user credentials",
		Args: cobra.ExactArgs(1), RunE: g.runFSUCEnable,
	}
	g.fsUCDisableCmd = &cobra.Command{
		Use: "disable USER_CREDS", Short: "Disable user credentials",
		Args: cobra.ExactArgs(1), RunE: g.runFSUCDisable,
	}
	g.fsUCResetPasswordCmd = &cobra.Command{
		Use: "reset-password USER_CREDS", Short: "Reset the password on user credentials",
		Args: cobra.ExactArgs(1), RunE: g.runFSUCResetPassword,
	}
	g.initFirestore()
	g.initFirestoreBackups()
	g.initFirestoreChangeStreams()
	g.initFirestoreDatabases()
	g.initFirestoreFieldsIndexes()
	g.initFirestoreFieldsIndexes5()
	g.initFirestoreLocations()
	g.initFirestoreOperations()
	g.initFirestoreUserCreds()
	return g
}
