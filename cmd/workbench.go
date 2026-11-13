package cmd

import "github.com/spf13/cobra"

// --- gcloud workbench (#396) ---

func (g *workbenchCmds) initWorkbench() {
	// Subgroups instances, executions, and schedules are implemented in
	// workbench_instances.go and workbench_executions_schedules.go.

}

// workbenchCmds holds the commands and flag values of one invocation's "workbench"
// command group, so invocations don't share state (#2070).
type workbenchCmds struct {
	inv *invocation

	workbenchCmd           *cobra.Command
	workbenchExecutionsCmd *cobra.Command
	wbExecCreateCmd        *cobra.Command
	wbExecDeleteCmd        *cobra.Command
	wbExecDescribeCmd      *cobra.Command
	wbExecListCmd          *cobra.Command
	flagWBExecRegion       string
	flagWBExecConfigFile   string
	flagWBExecFormat       string
	workbenchSchedulesCmd  *cobra.Command
	wbSchedCreateCmd       *cobra.Command
	wbSchedDeleteCmd       *cobra.Command
	wbSchedDescribeCmd     *cobra.Command
	wbSchedListCmd         *cobra.Command
	wbSchedUpdateCmd       *cobra.Command
	wbSchedPauseCmd        *cobra.Command
	wbSchedResumeCmd       *cobra.Command
	flagWBSchedRegion      string
	flagWBSchedConfigFile  string
	flagWBSchedUpdateMask  string
	flagWBSchedFormat      string
	flagWBSchedCatchup     bool
	workbenchInstancesCmd  *cobra.Command
	wbiCreateCmd           *cobra.Command
	wbiDeleteCmd           *cobra.Command
	wbiDescribeCmd         *cobra.Command
	wbiListCmd             *cobra.Command
	wbiUpdateCmd           *cobra.Command
	wbiStartCmd            *cobra.Command
	wbiStopCmd             *cobra.Command
	wbiResetCmd            *cobra.Command
	wbiDiagnoseCmd         *cobra.Command
	wbiGetConfigCmd        *cobra.Command
	wbiCheckUpgCmd         *cobra.Command
	wbiUpgradeCmd          *cobra.Command
	wbiRollbackCmd         *cobra.Command
	wbiRestoreCmd          *cobra.Command
	wbiResizeDiskCmd       *cobra.Command
	wbiGetIamCmd           *cobra.Command
	wbiSetIamCmd           *cobra.Command
	wbiAddIamCmd           *cobra.Command
	wbiRemoveIamCmd        *cobra.Command
	flagWBILocation        string
	flagWBIConfigFile      string
	flagWBIUpdateMask      string
	flagWBIFormat          string
	flagWBIAsync           bool
	flagWBIDiskSize        int64
	flagWBIDiskType        string
	flagWBISnapshot        string
	flagWBIRestoreSource   string
	flagWBIRestoreTime     string
	flagWBIRevision        string
	flagWBIIamMember       string
	flagWBIIamRole         string
	flagWBIImageFamily     string
}

// newWorkbenchCmd returns the "workbench" command group for inv.
func newWorkbenchCmd(inv *invocation) *cobra.Command {
	return newWorkbenchCmds(inv).workbenchCmd
}

// newWorkbenchCmds builds the "workbench" command group for inv and returns its state.
func newWorkbenchCmds(inv *invocation) *workbenchCmds {
	g := &workbenchCmds{inv: inv}
	g.workbenchCmd = &cobra.Command{Use: "workbench", Short: "Manage Vertex AI Workbench"}
	g.workbenchExecutionsCmd = &cobra.Command{
		Use:   "executions",
		Short: "Manage Vertex AI Workbench notebook execution jobs",
	}
	g.wbExecCreateCmd = &cobra.Command{
		Use: "create EXECUTION", Short: "Create a notebook execution job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runWBExecCreate,
	}
	g.wbExecDeleteCmd = &cobra.Command{
		Use: "delete EXECUTION", Short: "Delete a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: g.runWBExecDelete,
	}
	g.wbExecDescribeCmd = &cobra.Command{
		Use: "describe EXECUTION", Short: "Describe a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: g.runWBExecDescribe,
	}
	g.wbExecListCmd = &cobra.Command{
		Use: "list", Short: "List notebook execution jobs in a location",
		Args: cobra.NoArgs, RunE: g.runWBExecList,
	}
	g.workbenchSchedulesCmd = &cobra.Command{
		Use:   "schedules",
		Short: "Manage Vertex AI Workbench schedules",
	}
	g.wbSchedCreateCmd = &cobra.Command{
		Use: "create SCHEDULE", Short: "Create a schedule from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runWBSchedCreate,
	}
	g.wbSchedDeleteCmd = &cobra.Command{
		Use: "delete SCHEDULE", Short: "Delete a schedule",
		Args: cobra.ExactArgs(1), RunE: g.runWBSchedDelete,
	}
	g.wbSchedDescribeCmd = &cobra.Command{
		Use: "describe SCHEDULE", Short: "Describe a schedule",
		Args: cobra.ExactArgs(1), RunE: g.runWBSchedDescribe,
	}
	g.wbSchedListCmd = &cobra.Command{
		Use: "list", Short: "List schedules in a location",
		Args: cobra.NoArgs, RunE: g.runWBSchedList,
	}
	g.wbSchedUpdateCmd = &cobra.Command{
		Use: "update SCHEDULE", Short: "Update a schedule from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runWBSchedUpdate,
	}
	g.wbSchedPauseCmd = &cobra.Command{
		Use: "pause SCHEDULE", Short: "Pause a schedule",
		Args: cobra.ExactArgs(1), RunE: g.runWBSchedPause,
	}
	g.wbSchedResumeCmd = &cobra.Command{
		Use: "resume SCHEDULE", Short: "Resume a paused schedule",
		Args: cobra.ExactArgs(1), RunE: g.runWBSchedResume,
	}
	g.workbenchInstancesCmd = &cobra.Command{
		Use:   "instances",
		Short: "Manage Vertex AI Workbench instances",
	}
	g.wbiCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Workbench instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runWBICreate,
	}
	g.wbiDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIDelete,
	}
	g.wbiDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIDescribe,
	}
	g.wbiListCmd = &cobra.Command{
		Use: "list", Short: "List Workbench instances in a location",
		Args: cobra.NoArgs, RunE: g.runWBIList,
	}
	g.wbiUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a Workbench instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runWBIUpdate,
	}
	g.wbiStartCmd = &cobra.Command{
		Use: "start INSTANCE", Short: "Start a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIStart,
	}
	g.wbiStopCmd = &cobra.Command{
		Use: "stop INSTANCE", Short: "Stop a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIStop,
	}
	g.wbiResetCmd = &cobra.Command{
		Use: "reset INSTANCE", Short: "Reset a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIReset,
	}
	g.wbiDiagnoseCmd = &cobra.Command{
		Use: "diagnose INSTANCE", Short: "Collect a diagnostic snapshot of a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIDiagnose,
	}
	g.wbiGetConfigCmd = &cobra.Command{
		Use: "get-config", Short: "Get the Workbench instance configuration for a location",
		Args: cobra.NoArgs, RunE: g.runWBIGetConfig,
	}
	g.wbiCheckUpgCmd = &cobra.Command{
		Use: "check-instance-upgradability INSTANCE", Short: "Check whether an instance can be upgraded",
		Args: cobra.ExactArgs(1), RunE: g.runWBICheckUpgrade,
	}
	g.wbiUpgradeCmd = &cobra.Command{
		Use: "upgrade INSTANCE", Short: "Upgrade a Workbench instance to the latest image",
		Args: cobra.ExactArgs(1), RunE: g.runWBIUpgrade,
	}
	g.wbiRollbackCmd = &cobra.Command{
		Use: "rollback INSTANCE", Short: "Roll back a Workbench instance to a previous snapshot",
		Args: cobra.ExactArgs(1), RunE: g.runWBIRollback,
	}
	g.wbiRestoreCmd = &cobra.Command{
		Use: "restore INSTANCE", Short: "Restore a Workbench instance from a snapshot",
		Args: cobra.ExactArgs(1), RunE: g.runWBIRestore,
	}
	g.wbiResizeDiskCmd = &cobra.Command{
		Use: "resize-disk INSTANCE", Short: "Resize a Workbench instance's disk",
		Args: cobra.ExactArgs(1), RunE: g.runWBIResizeDisk,
	}
	g.wbiGetIamCmd = &cobra.Command{
		Use: "get-iam-policy INSTANCE", Short: "Print the IAM policy for a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIGetIam,
	}
	g.wbiSetIamCmd = &cobra.Command{
		Use: "set-iam-policy INSTANCE POLICY_FILE", Short: "Replace the IAM policy for a Workbench instance",
		Args: cobra.ExactArgs(2), RunE: g.runWBISetIam,
	}
	g.wbiAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding INSTANCE", Short: "Add an IAM binding to a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIAddIam,
	}
	g.wbiRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding INSTANCE", Short: "Remove an IAM binding from a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: g.runWBIRemoveIam,
	}
	g.initWorkbench()
	g.initWorkbenchExecutionsSchedules()
	g.initWorkbenchExecutionsSchedules2()
	g.initWorkbenchInstances()
	return g
}
