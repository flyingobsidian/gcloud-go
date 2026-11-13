package cli

import "github.com/spf13/cobra"

// --- gcloud migration (#357) ---
//
// vms target-projects, image-imports, machine-image-imports, and disk-migrations
// are implemented against the real VM Migration API in migration_vms.go and
// migration_vms_extra.go. gcloud-python's `migration vms` surface is limited to
// exactly these four subgroups; earlier fake stubs (sources, groups, migrations,
// cutover-jobs, replication-cycles, utilization-reports, operations, machine-images)
// have been removed as they don't exist in gcloud-python.

func (s *migrationCmds) initMigration() {
	s.migrationCmd.AddCommand(s.migrationVMsCmd)

}

// migrationCmds holds the commands and flag values of one invocation's "migration"
// command group, so invocations don't share state (#2070).
type migrationCmds struct {
	inv *invocation

	migrationCmd              *cobra.Command
	migrationVMsCmd           *cobra.Command
	flagMVMLocation           string
	flagMVMFormat             string
	flagMVMFilter             string
	flagMVMAsync              bool
	flagMVMTargetProj         string
	flagMVMDescription        string
	flagMVMUpdateMask         string
	flagMVMConfigFile         string
	flagMVMSourceFile         string
	flagMVMImageName          string
	flagMVMImportTP           string
	flagMVMFamily             string
	flagMVMLabels             []string
	flagMVMAddLicenses        []string
	flagMVMSingleRegion       bool
	flagMVMKmsKey             string
	flagMVMSkipOSAdapt        bool
	flagMVMGeneralize         bool
	flagMVMLicenseType        string
	flagMVMBootConv           string
	mvmTargetProjectsCmd      *cobra.Command
	mvmTPAddCmd               *cobra.Command
	mvmTPDeleteCmd            *cobra.Command
	mvmTPDescribeCmd          *cobra.Command
	mvmTPListCmd              *cobra.Command
	mvmTPUpdateCmd            *cobra.Command
	mvmImageImportsCmd        *cobra.Command
	mvmIICreateCmd            *cobra.Command
	mvmIIDeleteCmd            *cobra.Command
	mvmIIDescribeCmd          *cobra.Command
	mvmIIListCmd              *cobra.Command
	flagMVMMIISourceFile      string
	flagMVMMIIName            string
	flagMVMMIITarget          string
	flagMVMDMSource           string
	flagMVMDMDescription      string
	mvmMachineImageImportsCmd *cobra.Command
	mvmMIICreateCmd           *cobra.Command
	mvmMIIDeleteCmd           *cobra.Command
	mvmMIIDescribeCmd         *cobra.Command
	mvmMIIListCmd             *cobra.Command
	mvmDiskMigrationsCmd      *cobra.Command
	mvmDMCreateCmd            *cobra.Command
	mvmDMDeleteCmd            *cobra.Command
	mvmDMDescribeCmd          *cobra.Command
	mvmDMListCmd              *cobra.Command
	mvmDMCancelCmd            *cobra.Command
	mvmDMRunCmd               *cobra.Command
	mvmDMUpdateCmd            *cobra.Command
}

// newMigrationCmd returns the "migration" command group for inv.
func newMigrationCmd(inv *invocation) *cobra.Command {
	return newMigrationCmds(inv).migrationCmd
}

// newMigrationCmds builds the "migration" command group for inv and returns its state.
func newMigrationCmds(inv *invocation) *migrationCmds {
	s := &migrationCmds{inv: inv}
	s.migrationCmd = &cobra.Command{Use: "migration", Short: "Migrate to Virtual Machines"}
	s.migrationVMsCmd = &cobra.Command{Use: "vms", Short: "Migrate to VMs"}
	s.mvmTargetProjectsCmd = &cobra.Command{
		Use: "target-projects", Short: "Manage VM Migration target projects",
	}
	s.mvmTPAddCmd = &cobra.Command{
		Use: "add TARGET_PROJECT", Short: "Create a target project",
		Args: cobra.ExactArgs(1), RunE: s.runMVMTPCreate,
	}
	s.mvmTPDeleteCmd = &cobra.Command{
		Use: "delete TARGET_PROJECT", Short: "Delete a target project",
		Args: cobra.ExactArgs(1), RunE: s.runMVMTPDelete,
	}
	s.mvmTPDescribeCmd = &cobra.Command{
		Use: "describe TARGET_PROJECT", Short: "Describe a target project",
		Args: cobra.ExactArgs(1), RunE: s.runMVMTPDescribe,
	}
	s.mvmTPListCmd = &cobra.Command{
		Use: "list", Short: "List target projects",
		Args: cobra.NoArgs, RunE: s.runMVMTPList,
	}
	s.mvmTPUpdateCmd = &cobra.Command{
		Use: "update TARGET_PROJECT", Short: "Update a target project",
		Args: cobra.ExactArgs(1), RunE: s.runMVMTPUpdate,
	}
	s.mvmImageImportsCmd = &cobra.Command{
		Use: "image-imports", Short: "Manage VM Migration image imports",
	}
	s.mvmIICreateCmd = &cobra.Command{
		Use: "create IMAGE_IMPORT", Short: "Create an image import",
		Args: cobra.ExactArgs(1), RunE: s.runMVMIICreate,
	}
	s.mvmIIDeleteCmd = &cobra.Command{
		Use: "delete IMAGE_IMPORT", Short: "Delete an image import",
		Args: cobra.ExactArgs(1), RunE: s.runMVMIIDelete,
	}
	s.mvmIIDescribeCmd = &cobra.Command{
		Use: "describe IMAGE_IMPORT", Short: "Describe an image import",
		Args: cobra.ExactArgs(1), RunE: s.runMVMIIDescribe,
	}
	s.mvmIIListCmd = &cobra.Command{
		Use: "list", Short: "List image imports",
		Args: cobra.NoArgs, RunE: s.runMVMIIList,
	}
	s.mvmMachineImageImportsCmd = &cobra.Command{
		Use: "machine-image-imports", Short: "Manage VM Migration machine image imports",
	}
	s.mvmMIICreateCmd = &cobra.Command{
		Use: "create MACHINE_IMAGE_IMPORT", Short: "Create a machine image import",
		Args: cobra.ExactArgs(1), RunE: s.runMVMMIICreate,
	}
	s.mvmMIIDeleteCmd = &cobra.Command{
		Use: "delete MACHINE_IMAGE_IMPORT", Short: "Delete a machine image import",
		Args: cobra.ExactArgs(1), RunE: s.runMVMMIIDelete,
	}
	s.mvmMIIDescribeCmd = &cobra.Command{
		Use: "describe MACHINE_IMAGE_IMPORT", Short: "Describe a machine image import",
		Args: cobra.ExactArgs(1), RunE: s.runMVMMIIDescribe,
	}
	s.mvmMIIListCmd = &cobra.Command{
		Use: "list", Short: "List machine image imports",
		Args: cobra.NoArgs, RunE: s.runMVMMIIList,
	}
	s.mvmDiskMigrationsCmd = &cobra.Command{
		Use: "disk-migrations", Short: "Migrate disks to Compute Engine",
	}
	s.mvmDMCreateCmd = &cobra.Command{
		Use: "create DISK_MIGRATION_JOB", Short: "Create a disk migration job",
		Args: cobra.ExactArgs(1), RunE: s.runMVMDMCreate,
	}
	s.mvmDMDeleteCmd = &cobra.Command{
		Use: "delete DISK_MIGRATION_JOB", Short: "Delete a disk migration job",
		Args: cobra.ExactArgs(1), RunE: s.runMVMDMDelete,
	}
	s.mvmDMDescribeCmd = &cobra.Command{
		Use: "describe DISK_MIGRATION_JOB", Short: "Describe a disk migration job",
		Args: cobra.ExactArgs(1), RunE: s.runMVMDMDescribe,
	}
	s.mvmDMListCmd = &cobra.Command{
		Use: "list", Short: "List disk migration jobs for a source",
		Args: cobra.NoArgs, RunE: s.runMVMDMList,
	}
	s.mvmDMCancelCmd = &cobra.Command{
		Use: "cancel DISK_MIGRATION_JOB", Short: "Cancel a running disk migration job",
		Args: cobra.ExactArgs(1), RunE: s.runMVMDMCancel,
	}
	s.mvmDMRunCmd = &cobra.Command{
		Use: "run DISK_MIGRATION_JOB", Short: "Run a disk migration job",
		Args: cobra.ExactArgs(1), RunE: s.runMVMDMRun,
	}
	s.mvmDMUpdateCmd = &cobra.Command{
		Use: "update DISK_MIGRATION_JOB", Short: "Update a disk migration job",
		Args: cobra.ExactArgs(1), RunE: s.runMVMDMUpdate,
	}
	s.initMigration()
	s.initMigrationVms()
	s.initMigrationVmsExtra()
	return s
}
