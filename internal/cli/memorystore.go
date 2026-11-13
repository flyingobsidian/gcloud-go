package cli

import "github.com/spf13/cobra"

// --- gcloud memorystore (#355) ---
//
// The memorystore v1beta API is not exposed through google.golang.org/api, so
// this file wires the top-level command up and hands its subcommands a shared
// REST client that talks directly to the memorystore endpoint.

func memorystoreRest() *restClient {
	return newRESTClient("https://memorystore.googleapis.com/v1beta")
}

func (s *memorystoreCmds) initMemorystore() {

}

// memorystoreCmds holds the commands and flag values of one invocation's "memorystore"
// command group, so invocations don't share state (#2070).
type memorystoreCmds struct {
	inv *invocation

	memorystoreCmd                 *cobra.Command
	memstoreACLCmd                 *cobra.Command
	flagMemstoreACLLocation        string
	flagMemstoreACLFormat          string
	flagMemstoreACLConfigFile      string
	flagMemstoreACLUpdateMask      string
	flagMemstoreACLPageSize        int64
	memstoreACLCreateCmd           *cobra.Command
	memstoreACLDeleteCmd           *cobra.Command
	memstoreACLDescribeCmd         *cobra.Command
	memstoreACLListCmd             *cobra.Command
	memstoreACLUpdateCmd           *cobra.Command
	memstoreBCCmd                  *cobra.Command
	memstoreBCBackupsCmd           *cobra.Command
	flagMemstoreBCLocation         string
	flagMemstoreBCFormat           string
	flagMemstoreBCPageSize         int64
	flagMemstoreBCBackupCollection string
	flagMemstoreBCBackupConfigFile string
	memstoreBCDescribeCmd          *cobra.Command
	memstoreBCListCmd              *cobra.Command
	memstoreBCBackupDescribeCmd    *cobra.Command
	memstoreBCBackupListCmd        *cobra.Command
	memstoreBCBackupDeleteCmd      *cobra.Command
	memstoreBCBackupExportCmd      *cobra.Command
	memstoreInstCmd                *cobra.Command
	flagMemstoreInstLocation       string
	flagMemstoreInstFormat         string
	flagMemstoreInstConfigFile     string
	flagMemstoreInstUpdateMask     string
	flagMemstoreInstPageSize       int64
	flagMemstoreInstZDCZones       []string
	memstoreInstCreateCmd          *cobra.Command
	memstoreInstDeleteCmd          *cobra.Command
	memstoreInstDescribeCmd        *cobra.Command
	memstoreInstListCmd            *cobra.Command
	memstoreInstUpdateCmd          *cobra.Command
	memstoreInstBackupCmd          *cobra.Command
	memstoreInstFinishMigrationCmd *cobra.Command
	memstoreInstStartMigrationCmd  *cobra.Command
	memstoreInstGetCACmd           *cobra.Command
	memstoreInstGetSharedCACmd     *cobra.Command
	memstoreInstRescheduleMaintCmd *cobra.Command
	memstoreLocCmd                 *cobra.Command
	flagMemstoreLocFormat          string
	flagMemstoreLocPageSize        int64
	memstoreLocDescribeCmd         *cobra.Command
	memstoreLocListCmd             *cobra.Command
	memstoreOpCmd                  *cobra.Command
	flagMemstoreOpLocation         string
	flagMemstoreOpFormat           string
	flagMemstoreOpPageSize         int64
	memstoreOpCancelCmd            *cobra.Command
	memstoreOpDeleteCmd            *cobra.Command
	memstoreOpDescribeCmd          *cobra.Command
	memstoreOpListCmd              *cobra.Command
	memstoreTAUCmd                 *cobra.Command
	memstoreAuthTokensCmd          *cobra.Command
	memstoreCreateTAUCmd           *cobra.Command
	memstoreTAUCreateTokenCmd      *cobra.Command
	memstoreTAUDeleteCmd           *cobra.Command
	memstoreTAUDescribeCmd         *cobra.Command
	memstoreTAUListCmd             *cobra.Command
	memstoreATDeleteCmd            *cobra.Command
	memstoreATDescribeCmd          *cobra.Command
	memstoreATListCmd              *cobra.Command
	flagMemstoreTAULocation        string
	flagMemstoreTAUInstance        string
	flagMemstoreTAUUser            string
	flagMemstoreTAUFormat          string
	flagMemstoreTAUAsync           bool
	flagMemstoreTAURequest         string
	flagMemstoreTAUForce           bool
	flagMemstoreTAUPageSize        int64
}

// newMemorystoreCmd returns the "memorystore" command group for inv.
func newMemorystoreCmd(inv *invocation) *cobra.Command {
	return newMemorystoreCmds(inv).memorystoreCmd
}

// newMemorystoreCmds builds the "memorystore" command group for inv and returns its state.
func newMemorystoreCmds(inv *invocation) *memorystoreCmds {
	s := &memorystoreCmds{inv: inv}
	s.memorystoreCmd = &cobra.Command{Use: "memorystore", Short: "Manage Memorystore"}
	s.memstoreACLCmd = &cobra.Command{Use: "acl-policies", Short: "Manage Memorystore ACL policies"}
	s.memstoreACLCreateCmd = &cobra.Command{
		Use: "create ACL_POLICY", Short: "Create a Memorystore ACL policy",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreACLCreate,
	}
	s.memstoreACLDeleteCmd = &cobra.Command{
		Use: "delete ACL_POLICY", Short: "Delete a Memorystore ACL policy",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreACLDelete,
	}
	s.memstoreACLDescribeCmd = &cobra.Command{
		Use: "describe ACL_POLICY", Short: "Describe a Memorystore ACL policy",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreACLDescribe,
	}
	s.memstoreACLListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore ACL policies",
		Args: cobra.NoArgs, RunE: s.runMemstoreACLList,
	}
	s.memstoreACLUpdateCmd = &cobra.Command{
		Use: "update ACL_POLICY", Short: "Update a Memorystore ACL policy",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreACLUpdate,
	}
	s.memstoreBCCmd = &cobra.Command{Use: "backup-collections", Short: "Manage Memorystore backup collections"}
	s.memstoreBCBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage backups within a Memorystore backup collection"}
	s.memstoreBCDescribeCmd = &cobra.Command{
		Use: "describe BACKUP_COLLECTION", Short: "Describe a Memorystore backup collection",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreBCDescribe,
	}
	s.memstoreBCListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore backup collections",
		Args: cobra.NoArgs, RunE: s.runMemstoreBCList,
	}
	s.memstoreBCBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup in a Memorystore backup collection",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreBCBackupDescribe,
	}
	s.memstoreBCBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups in a Memorystore backup collection",
		Args: cobra.NoArgs, RunE: s.runMemstoreBCBackupList,
	}
	s.memstoreBCBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup in a Memorystore backup collection",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreBCBackupDelete,
	}
	s.memstoreBCBackupExportCmd = &cobra.Command{
		Use: "export BACKUP", Short: "Export a Memorystore backup to Cloud Storage",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreBCBackupExport,
	}
	s.memstoreInstCmd = &cobra.Command{Use: "instances", Short: "Manage Memorystore instances"}
	s.memstoreInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstCreate,
	}
	s.memstoreInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstDelete,
	}
	s.memstoreInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstDescribe,
	}
	s.memstoreInstListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore instances",
		Args: cobra.NoArgs, RunE: s.runMemstoreInstList,
	}
	s.memstoreInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstUpdate,
	}
	s.memstoreInstBackupCmd = &cobra.Command{
		Use: "backup INSTANCE", Short: "Trigger a backup for a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstBackup,
	}
	s.memstoreInstFinishMigrationCmd = &cobra.Command{
		Use: "finish-migration INSTANCE", Short: "Finish an in-progress Memorystore migration",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstFinishMigration,
	}
	s.memstoreInstStartMigrationCmd = &cobra.Command{
		Use: "start-migration INSTANCE", Short: "Start a Memorystore migration",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstStartMigration,
	}
	s.memstoreInstGetCACmd = &cobra.Command{
		Use: "get-certificate-authority INSTANCE", Short: "Get the certificate authority for a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstGetCA,
	}
	s.memstoreInstGetSharedCACmd = &cobra.Command{
		Use:   "get-shared-regional-certificate-authority",
		Short: "Get the shared regional certificate authority for a Memorystore location",
		Args:  cobra.NoArgs, RunE: s.runMemstoreInstGetSharedCA,
	}
	s.memstoreInstRescheduleMaintCmd = &cobra.Command{
		Use: "reschedule-maintenance INSTANCE", Short: "Reschedule maintenance for a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreInstRescheduleMaint,
	}
	s.memstoreLocCmd = &cobra.Command{Use: "locations", Short: "Manage Memorystore locations"}
	s.memstoreLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Memorystore location",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreLocDescribe,
	}
	s.memstoreLocListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore locations",
		Args: cobra.NoArgs, RunE: s.runMemstoreLocList,
	}
	s.memstoreOpCmd = &cobra.Command{Use: "operations", Short: "Manage Memorystore operations"}
	s.memstoreOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Memorystore operation",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreOpCancel,
	}
	s.memstoreOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Memorystore operation record",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreOpDelete,
	}
	s.memstoreOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Memorystore operation",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreOpDescribe,
	}
	s.memstoreOpListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore operations",
		Args: cobra.NoArgs, RunE: s.runMemstoreOpList,
	}
	s.memstoreTAUCmd = &cobra.Command{
		Use: "token-auth-users", Short: "Manage token auth users of a Memorystore instance",
	}
	s.memstoreAuthTokensCmd = &cobra.Command{
		Use: "auth-tokens", Short: "Manage auth tokens of a Memorystore token auth user",
	}
	s.memstoreCreateTAUCmd = &cobra.Command{
		Use: "create-token-auth-user INSTANCE", Short: "Create token auth user for a Memorystore instance",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreCreateTAU,
	}
	s.memstoreTAUCreateTokenCmd = &cobra.Command{
		Use: "create-auth-token TOKEN_AUTH_USER", Short: "Create an auth token for a token auth user",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreTAUCreateToken,
	}
	s.memstoreTAUDeleteCmd = &cobra.Command{
		Use: "delete TOKEN_AUTH_USER", Short: "Delete a token auth user",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreTAUDelete,
	}
	s.memstoreTAUDescribeCmd = &cobra.Command{
		Use: "describe TOKEN_AUTH_USER", Short: "Describe a token auth user",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreTAUDescribe,
	}
	s.memstoreTAUListCmd = &cobra.Command{
		Use: "list", Short: "List token auth users of an instance",
		Args: cobra.NoArgs, RunE: s.runMemstoreTAUList,
	}
	s.memstoreATDeleteCmd = &cobra.Command{
		Use: "delete AUTH_TOKEN", Short: "Delete an auth token",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreATDelete,
	}
	s.memstoreATDescribeCmd = &cobra.Command{
		Use: "describe AUTH_TOKEN", Short: "Describe an auth token",
		Args: cobra.ExactArgs(1), RunE: s.runMemstoreATDescribe,
	}
	s.memstoreATListCmd = &cobra.Command{
		Use: "list", Short: "List auth tokens of a token auth user",
		Args: cobra.NoArgs, RunE: s.runMemstoreATList,
	}
	s.initMemorystore()
	s.initMemorystoreAclPolicies()
	s.initMemorystoreBackupCollections()
	s.initMemorystoreInstances()
	s.initMemorystoreLocations()
	s.initMemorystoreOperations()
	s.initMemorystoreTokenAuth()
	return s
}
