package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud netapp (#360) ---

func (s *netappCmds) initNetapp() {

}

func netappLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func netappChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// netappCmds holds the commands and flag values of one invocation's "netapp"
// command group, so invocations don't share state (#2070).
type netappCmds struct {
	inv *invocation

	netappCmd                  *cobra.Command
	netappADCmd                *cobra.Command
	flagNetAppADLocation       string
	flagNetAppADConfigFile     string
	flagNetAppADUpdateMask     string
	flagNetAppADFormat         string
	flagNetAppADFilter         string
	flagNetAppADPageSize       int64
	netappADCreateCmd          *cobra.Command
	netappADDeleteCmd          *cobra.Command
	netappADDescribeCmd        *cobra.Command
	netappADListCmd            *cobra.Command
	netappADUpdateCmd          *cobra.Command
	netappBPCmd                *cobra.Command
	flagNetAppBPLocation       string
	flagNetAppBPConfigFile     string
	flagNetAppBPUpdateMask     string
	flagNetAppBPFormat         string
	flagNetAppBPFilter         string
	flagNetAppBPPageSize       int64
	netappBPCreateCmd          *cobra.Command
	netappBPDeleteCmd          *cobra.Command
	netappBPDescribeCmd        *cobra.Command
	netappBPListCmd            *cobra.Command
	netappBPUpdateCmd          *cobra.Command
	netappBVCmd                *cobra.Command
	flagNetAppBVLocation       string
	flagNetAppBVConfigFile     string
	flagNetAppBVUpdateMask     string
	flagNetAppBVFormat         string
	flagNetAppBVFilter         string
	flagNetAppBVPageSize       int64
	netappBVCreateCmd          *cobra.Command
	netappBVDeleteCmd          *cobra.Command
	netappBVDescribeCmd        *cobra.Command
	netappBVListCmd            *cobra.Command
	netappBVUpdateCmd          *cobra.Command
	netappBackupsCmd           *cobra.Command
	flagNetAppBackupVault      string
	flagNetAppBackupConfigFile string
	flagNetAppBackupUpdateMask string
	flagNetAppBackupFilter     string
	flagNetAppBackupPageSize   int64
	flagNetAppBackupOntapSrc   string
	netappBackupCreateCmd      *cobra.Command
	netappBackupDeleteCmd      *cobra.Command
	netappBackupDescribeCmd    *cobra.Command
	netappBackupListCmd        *cobra.Command
	netappBackupUpdateCmd      *cobra.Command
	netappHGCmd                *cobra.Command
	flagNetAppHGLocation       string
	flagNetAppHGConfigFile     string
	flagNetAppHGUpdateMask     string
	flagNetAppHGFormat         string
	flagNetAppHGFilter         string
	flagNetAppHGPageSize       int64
	netappHGCreateCmd          *cobra.Command
	netappHGDeleteCmd          *cobra.Command
	netappHGDescribeCmd        *cobra.Command
	netappHGListCmd            *cobra.Command
	netappHGUpdateCmd          *cobra.Command
	netappKMSCmd               *cobra.Command
	flagNetAppKMSLocation      string
	flagNetAppKMSConfigFile    string
	flagNetAppKMSUpdateMask    string
	flagNetAppKMSFormat        string
	flagNetAppKMSFilter        string
	flagNetAppKMSPageSize      int64
	netappKMSCreateCmd         *cobra.Command
	netappKMSDeleteCmd         *cobra.Command
	netappKMSDescribeCmd       *cobra.Command
	netappKMSListCmd           *cobra.Command
	netappKMSUpdateCmd         *cobra.Command
	netappKMSEncryptCmd        *cobra.Command
	netappKMSVerifyCmd         *cobra.Command
	netappLocCmd               *cobra.Command
	flagNetAppLocFormat        string
	flagNetAppLocFilter        string
	flagNetAppLocPageSize      int64
	netappLocListCmd           *cobra.Command
	netappLocDescribeCmd       *cobra.Command
	netappOpCmd                *cobra.Command
	flagNetAppOpLocation       string
	flagNetAppOpFormat         string
	flagNetAppOpFilter         string
	flagNetAppOpPageSize       int64
	netappOpCancelCmd          *cobra.Command
	netappOpDeleteCmd          *cobra.Command
	netappOpDescribeCmd        *cobra.Command
	netappOpListCmd            *cobra.Command
	netappSPCmd                *cobra.Command
	flagNetAppSPLocation       string
	flagNetAppSPConfigFile     string
	flagNetAppSPUpdateMask     string
	flagNetAppSPFormat         string
	flagNetAppSPFilter         string
	flagNetAppSPPageSize       int64
	// update-backup-config
	flagNetAppSPBackupCfgVolumeUUID string
	flagNetAppSPBackupCfgVault      string
	flagNetAppSPBackupCfgPolicies   []string
	flagNetAppSPBackupCfgSchedule   string
	// restore-volume
	flagNetAppSPRestoreBackup     string
	flagNetAppSPRestoreVolumeUUID string
	flagNetAppSPRestoreDestPath   string
	flagNetAppSPRestoreFileList   []string
	netappSPCreateCmd             *cobra.Command
	netappSPDeleteCmd             *cobra.Command
	netappSPDescribeCmd           *cobra.Command
	netappSPListCmd               *cobra.Command
	netappSPUpdateCmd             *cobra.Command
	netappSPSwitchCmd             *cobra.Command
	netappSPUpdateBackupCfgCmd    *cobra.Command
	netappSPListBackupCfgsCmd     *cobra.Command
	netappSPRestoreVolumeCmd      *cobra.Command
	netappVolCmd                  *cobra.Command
	flagNetAppVolLocation         string
	flagNetAppVolConfigFile       string
	flagNetAppVolUpdateMask       string
	flagNetAppVolFormat           string
	flagNetAppVolFilter           string
	flagNetAppVolPageSize         int64
	flagNetAppVolForce            bool
	flagNetAppVolSnapshotID       string
	netappVolCreateCmd            *cobra.Command
	netappVolDeleteCmd            *cobra.Command
	netappVolDescribeCmd          *cobra.Command
	netappVolListCmd              *cobra.Command
	netappVolUpdateCmd            *cobra.Command
	netappVolRevertCmd            *cobra.Command
	netappVolStartSplitCmd        *cobra.Command
	netappVolGetSplitStatusCmd    *cobra.Command
	netappReplCmd                 *cobra.Command
	flagNetAppReplVolume          string
	flagNetAppReplConfigFile      string
	flagNetAppReplUpdateMask      string
	flagNetAppReplFilter          string
	flagNetAppReplPageSize        int64
	netappReplCreateCmd           *cobra.Command
	netappReplDeleteCmd           *cobra.Command
	netappReplDescribeCmd         *cobra.Command
	netappReplListCmd             *cobra.Command
	netappReplUpdateCmd           *cobra.Command
	netappReplResumeCmd           *cobra.Command
	netappReplReverseCmd          *cobra.Command
	netappReplStopCmd             *cobra.Command
	netappReplSyncCmd             *cobra.Command
	netappSnapCmd                 *cobra.Command
	flagNetAppSnapVolume          string
	flagNetAppSnapConfigFile      string
	flagNetAppSnapUpdateMask      string
	flagNetAppSnapFilter          string
	flagNetAppSnapPageSize        int64
	netappSnapCreateCmd           *cobra.Command
	netappSnapDeleteCmd           *cobra.Command
	netappSnapDescribeCmd         *cobra.Command
	netappSnapListCmd             *cobra.Command
	netappSnapUpdateCmd           *cobra.Command
}

// newNetappCmd returns the "netapp" command group for inv.
func newNetappCmd(inv *invocation) *cobra.Command {
	return newNetappCmds(inv).netappCmd
}

// newNetappCmds builds the "netapp" command group for inv and returns its state.
func newNetappCmds(inv *invocation) *netappCmds {
	s := &netappCmds{inv: inv}
	s.netappCmd = &cobra.Command{Use: "netapp", Short: "Manage Cloud NetApp Files"}
	s.netappADCmd = &cobra.Command{
		Use:   "active-directories",
		Short: "Manage NetApp Active Directories",
	}
	s.netappADCreateCmd = &cobra.Command{
		Use: "create ACTIVE_DIRECTORY", Short: "Create an Active Directory",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppADCreate,
	}
	s.netappADDeleteCmd = &cobra.Command{
		Use: "delete ACTIVE_DIRECTORY", Short: "Delete an Active Directory",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppADDelete,
	}
	s.netappADDescribeCmd = &cobra.Command{
		Use: "describe ACTIVE_DIRECTORY", Short: "Describe an Active Directory",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppADDescribe,
	}
	s.netappADListCmd = &cobra.Command{
		Use: "list", Short: "List Active Directories",
		Args: cobra.NoArgs, RunE: s.runNetAppADList,
	}
	s.netappADUpdateCmd = &cobra.Command{
		Use: "update ACTIVE_DIRECTORY", Short: "Update an Active Directory",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppADUpdate,
	}
	s.netappBPCmd = &cobra.Command{Use: "backup-policies", Short: "Manage NetApp backup policies"}
	s.netappBPCreateCmd = &cobra.Command{
		Use: "create BACKUP_POLICY", Short: "Create a backup policy",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBPCreate,
	}
	s.netappBPDeleteCmd = &cobra.Command{
		Use: "delete BACKUP_POLICY", Short: "Delete a backup policy",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBPDelete,
	}
	s.netappBPDescribeCmd = &cobra.Command{
		Use: "describe BACKUP_POLICY", Short: "Describe a backup policy",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBPDescribe,
	}
	s.netappBPListCmd = &cobra.Command{
		Use: "list", Short: "List backup policies",
		Args: cobra.NoArgs, RunE: s.runNetAppBPList,
	}
	s.netappBPUpdateCmd = &cobra.Command{
		Use: "update BACKUP_POLICY", Short: "Update a backup policy",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBPUpdate,
	}
	s.netappBVCmd = &cobra.Command{Use: "backup-vaults", Short: "Manage NetApp backup vaults"}
	s.netappBVCreateCmd = &cobra.Command{
		Use: "create BACKUP_VAULT", Short: "Create a backup vault",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBVCreate,
	}
	s.netappBVDeleteCmd = &cobra.Command{
		Use: "delete BACKUP_VAULT", Short: "Delete a backup vault",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBVDelete,
	}
	s.netappBVDescribeCmd = &cobra.Command{
		Use: "describe BACKUP_VAULT", Short: "Describe a backup vault",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBVDescribe,
	}
	s.netappBVListCmd = &cobra.Command{
		Use: "list", Short: "List backup vaults",
		Args: cobra.NoArgs, RunE: s.runNetAppBVList,
	}
	s.netappBVUpdateCmd = &cobra.Command{
		Use: "update BACKUP_VAULT", Short: "Update a backup vault",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBVUpdate,
	}
	s.netappBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage NetApp backups within a backup vault"}
	s.netappBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a backup",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBackupCreate,
	}
	s.netappBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBackupDelete,
	}
	s.netappBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBackupDescribe,
	}
	s.netappBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups within a backup vault",
		Args: cobra.NoArgs, RunE: s.runNetAppBackupList,
	}
	s.netappBackupUpdateCmd = &cobra.Command{
		Use: "update BACKUP", Short: "Update a backup",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppBackupUpdate,
	}
	s.netappHGCmd = &cobra.Command{Use: "host-groups", Short: "Manage NetApp host groups"}
	s.netappHGCreateCmd = &cobra.Command{
		Use: "create HOST_GROUP", Short: "Create a host group",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppHGCreate,
	}
	s.netappHGDeleteCmd = &cobra.Command{
		Use: "delete HOST_GROUP", Short: "Delete a host group",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppHGDelete,
	}
	s.netappHGDescribeCmd = &cobra.Command{
		Use: "describe HOST_GROUP", Short: "Describe a host group",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppHGDescribe,
	}
	s.netappHGListCmd = &cobra.Command{
		Use: "list", Short: "List host groups",
		Args: cobra.NoArgs, RunE: s.runNetAppHGList,
	}
	s.netappHGUpdateCmd = &cobra.Command{
		Use: "update HOST_GROUP", Short: "Update a host group",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppHGUpdate,
	}
	s.netappKMSCmd = &cobra.Command{Use: "kms-configs", Short: "Manage NetApp KMS configs"}
	s.netappKMSCreateCmd = &cobra.Command{
		Use: "create KMS_CONFIG", Short: "Create a KMS config",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppKMSCreate,
	}
	s.netappKMSDeleteCmd = &cobra.Command{
		Use: "delete KMS_CONFIG", Short: "Delete a KMS config",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppKMSDelete,
	}
	s.netappKMSDescribeCmd = &cobra.Command{
		Use: "describe KMS_CONFIG", Short: "Describe a KMS config",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppKMSDescribe,
	}
	s.netappKMSListCmd = &cobra.Command{
		Use: "list", Short: "List KMS configs",
		Args: cobra.NoArgs, RunE: s.runNetAppKMSList,
	}
	s.netappKMSUpdateCmd = &cobra.Command{
		Use: "update KMS_CONFIG", Short: "Update a KMS config",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppKMSUpdate,
	}
	s.netappKMSEncryptCmd = &cobra.Command{
		Use: "encrypt KMS_CONFIG", Short: "Encrypt all existing volumes and storage pools with the KMS config",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppKMSEncrypt,
	}
	s.netappKMSVerifyCmd = &cobra.Command{
		Use: "verify KMS_CONFIG", Short: "Verify that the KMS config is reachable",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppKMSVerify,
	}
	s.netappLocCmd = &cobra.Command{Use: "locations", Short: "Get and list Cloud NetApp Files locations"}
	s.netappLocListCmd = &cobra.Command{
		Use: "list", Short: "List NetApp locations",
		Args: cobra.NoArgs, RunE: s.runNetAppLocList,
	}
	s.netappLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a NetApp location",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppLocDescribe,
	}
	s.netappOpCmd = &cobra.Command{Use: "operations", Short: "Manage NetApp Files operations"}
	s.netappOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a NetApp operation",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppOpCancel,
	}
	s.netappOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a NetApp operation record",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppOpDelete,
	}
	s.netappOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a NetApp operation",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppOpDescribe,
	}
	s.netappOpListCmd = &cobra.Command{
		Use: "list", Short: "List NetApp operations",
		Args: cobra.NoArgs, RunE: s.runNetAppOpList,
	}
	s.netappSPCmd = &cobra.Command{Use: "storage-pools", Short: "Manage NetApp storage pools"}
	s.netappSPCreateCmd = &cobra.Command{
		Use: "create STORAGE_POOL", Short: "Create a storage pool",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSPCreate,
	}
	s.netappSPDeleteCmd = &cobra.Command{
		Use: "delete STORAGE_POOL", Short: "Delete a storage pool",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSPDelete,
	}
	s.netappSPDescribeCmd = &cobra.Command{
		Use: "describe STORAGE_POOL", Short: "Describe a storage pool",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSPDescribe,
	}
	s.netappSPListCmd = &cobra.Command{
		Use: "list", Short: "List storage pools",
		Args: cobra.NoArgs, RunE: s.runNetAppSPList,
	}
	s.netappSPUpdateCmd = &cobra.Command{
		Use: "update STORAGE_POOL", Short: "Update a storage pool",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSPUpdate,
	}
	s.netappSPSwitchCmd = &cobra.Command{
		Use: "switch STORAGE_POOL", Short: "Switch the active zone of a Regional Flex storage pool",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSPSwitch,
	}
	s.netappSPUpdateBackupCfgCmd = &cobra.Command{
		Use: "update-backup-config STORAGE_POOL", Short: "Update the backup configuration for a volume in an ONTAP-mode storage pool",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSPUpdateBackupConfig,
	}
	s.netappSPListBackupCfgsCmd = &cobra.Command{
		Use: "list-backup-configs STORAGE_POOL", Short: "List backup configurations for all volumes in an ONTAP-mode storage pool",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSPListBackupConfigs,
	}
	s.netappSPRestoreVolumeCmd = &cobra.Command{
		Use: "restore-volume STORAGE_POOL", Short: "Restore a backup to a volume in an ONTAP-mode storage pool",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSPRestoreVolume,
	}
	s.netappVolCmd = &cobra.Command{Use: "volumes", Short: "Manage NetApp volumes"}
	s.netappVolCreateCmd = &cobra.Command{
		Use: "create VOLUME", Short: "Create a volume",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppVolCreate,
	}
	s.netappVolDeleteCmd = &cobra.Command{
		Use: "delete VOLUME", Short: "Delete a volume",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppVolDelete,
	}
	s.netappVolDescribeCmd = &cobra.Command{
		Use: "describe VOLUME", Short: "Describe a volume",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppVolDescribe,
	}
	s.netappVolListCmd = &cobra.Command{
		Use: "list", Short: "List volumes",
		Args: cobra.NoArgs, RunE: s.runNetAppVolList,
	}
	s.netappVolUpdateCmd = &cobra.Command{
		Use: "update VOLUME", Short: "Update a volume",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppVolUpdate,
	}
	s.netappVolRevertCmd = &cobra.Command{
		Use: "revert VOLUME", Short: "Revert a volume to a snapshot",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppVolRevert,
	}
	s.netappVolStartSplitCmd = &cobra.Command{
		Use: "start-split VOLUME", Short: "Start splitting a clone volume from its source volume",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppVolStartSplit,
	}
	s.netappVolGetSplitStatusCmd = &cobra.Command{
		Use: "get-split-status VOLUME", Short: "Retrieve the split status of a clone volume",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppVolGetSplitStatus,
	}
	s.netappReplCmd = &cobra.Command{Use: "replications", Short: "Manage NetApp volume replications"}
	s.netappReplCreateCmd = &cobra.Command{
		Use: "create REPLICATION", Short: "Create a replication",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppReplCreate,
	}
	s.netappReplDeleteCmd = &cobra.Command{
		Use: "delete REPLICATION", Short: "Delete a replication",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppReplDelete,
	}
	s.netappReplDescribeCmd = &cobra.Command{
		Use: "describe REPLICATION", Short: "Describe a replication",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppReplDescribe,
	}
	s.netappReplListCmd = &cobra.Command{
		Use: "list", Short: "List replications for a volume",
		Args: cobra.NoArgs, RunE: s.runNetAppReplList,
	}
	s.netappReplUpdateCmd = &cobra.Command{
		Use: "update REPLICATION", Short: "Update a replication",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppReplUpdate,
	}
	s.netappReplResumeCmd = &cobra.Command{
		Use: "resume REPLICATION", Short: "Resume a replication",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppReplResume,
	}
	s.netappReplReverseCmd = &cobra.Command{
		Use: "reverse-direction REPLICATION", Short: "Reverse a replication's direction",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppReplReverse,
	}
	s.netappReplStopCmd = &cobra.Command{
		Use: "stop REPLICATION", Short: "Stop a replication",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppReplStop,
	}
	s.netappReplSyncCmd = &cobra.Command{
		Use: "sync REPLICATION", Short: "Sync a replication",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppReplSync,
	}
	s.netappSnapCmd = &cobra.Command{Use: "snapshots", Short: "Manage NetApp volume snapshots"}
	s.netappSnapCreateCmd = &cobra.Command{
		Use: "create SNAPSHOT", Short: "Create a snapshot",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSnapCreate,
	}
	s.netappSnapDeleteCmd = &cobra.Command{
		Use: "delete SNAPSHOT", Short: "Delete a snapshot",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSnapDelete,
	}
	s.netappSnapDescribeCmd = &cobra.Command{
		Use: "describe SNAPSHOT", Short: "Describe a snapshot",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSnapDescribe,
	}
	s.netappSnapListCmd = &cobra.Command{
		Use: "list", Short: "List snapshots for a volume",
		Args: cobra.NoArgs, RunE: s.runNetAppSnapList,
	}
	s.netappSnapUpdateCmd = &cobra.Command{
		Use: "update SNAPSHOT", Short: "Update a snapshot",
		Args: cobra.ExactArgs(1), RunE: s.runNetAppSnapUpdate,
	}
	s.initNetapp()
	s.initNetappActiveDirectories()
	s.initNetappBackupPolicies()
	s.initNetappBackupVaults()
	s.initNetappHostGroups()
	s.initNetappKmsConfigs()
	s.initNetappLocations()
	s.initNetappOperations()
	s.initNetappStoragePools()
	s.initNetappVolumes()
	return s
}
