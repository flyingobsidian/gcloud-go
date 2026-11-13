package cli

import "github.com/spf13/cobra"

// --- gcloud transfer (#887-#891) ---
//
// The top-level `transfer` command aggregates the Storage Transfer Service
// subgroups. All subgroups (agent-pools, agents, jobs, operations) and the
// authorize command are implemented in transfer_all.go.

func (s *transferCmds) initTransfer() {

}

// transferCmds holds the commands and flag values of one invocation's "transfer"
// command group, so invocations don't share state (#2070).
type transferCmds struct {
	inv *invocation

	transferCmd           *cobra.Command
	flagXferFormat        string
	flagXferFilter        string
	flagXferConfigFile    string
	flagXferAsync         bool
	flagXferDisplayName   string
	flagXferBandwidthMbps int64
	flagXferClearName     bool
	flagXferClearBW       bool
	flagXferJobName       string
	flagXferJobDesc       string
	flagXferSrcAgentPool  string
	flagXferDstAgentPool  string
	flagXferSrcCredsFile  string
	flagXferSchedRepeats  string
	flagXferJobStatus     string
	flagXferPool          string
	flagXferCredsFile     string
	flagXferCount         int
	flagXferIDPrefix      string
	flagXferLogsDirectory string
	flagXferMemlockLimit  int64
	flagXferMountDirs     []string
	flagXferProxy         string
	flagXferS3Compatible  bool
	flagXferAgentID       string
	flagXferAgentAll      bool
	flagXferUninstall     bool
	flagXferAddMissing    bool
	xferAgentPoolsCmd     *cobra.Command
	xferAgentsCmd         *cobra.Command
	xferJobsCmd           *cobra.Command
	xferOperationsCmd     *cobra.Command
	xferPoolCreateCmd     *cobra.Command
	xferPoolDeleteCmd     *cobra.Command
	xferPoolDescribeCmd   *cobra.Command
	xferPoolListCmd       *cobra.Command
	xferPoolUpdateCmd     *cobra.Command
	xferAgentInstallCmd   *cobra.Command
	xferAgentDeleteCmd    *cobra.Command
	xferJobCreateCmd      *cobra.Command
	xferJobDeleteCmd      *cobra.Command
	xferJobDescribeCmd    *cobra.Command
	xferJobListCmd        *cobra.Command
	xferJobUpdateCmd      *cobra.Command
	xferJobRunCmd         *cobra.Command
	xferJobMonitorCmd     *cobra.Command
	xferOpCancelCmd       *cobra.Command
	xferOpDescribeCmd     *cobra.Command
	xferOpListCmd         *cobra.Command
	xferOpPauseCmd        *cobra.Command
	xferOpResumeCmd       *cobra.Command
	xferAuthorizeCmd      *cobra.Command
}

// newTransferCmd returns the "transfer" command group for inv.
func newTransferCmd(inv *invocation) *cobra.Command {
	return newTransferCmds(inv).transferCmd
}

// newTransferCmds builds the "transfer" command group for inv and returns its state.
func newTransferCmds(inv *invocation) *transferCmds {
	s := &transferCmds{inv: inv}
	s.transferCmd = &cobra.Command{Use: "transfer", Short: "Manage Storage Transfer Service"}
	s.xferAgentPoolsCmd = &cobra.Command{Use: "agent-pools", Short: "Manage on-premise transfer agent pools"}
	s.xferAgentsCmd = &cobra.Command{Use: "agents", Short: "Manage transfer agents"}
	s.xferJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage transfer jobs"}
	s.xferOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage transfer operations"}
	s.xferPoolCreateCmd = &cobra.Command{
		Use: "create POOL", Short: "Create a Transfer Service agent pool",
		Args: cobra.ExactArgs(1), RunE: s.runXferPoolCreate,
	}
	s.xferPoolDeleteCmd = &cobra.Command{
		Use: "delete POOL", Short: "Delete a Transfer Service agent pool",
		Args: cobra.ExactArgs(1), RunE: s.runXferPoolDelete,
	}
	s.xferPoolDescribeCmd = &cobra.Command{
		Use: "describe POOL", Short: "Describe a Transfer Service agent pool",
		Args: cobra.ExactArgs(1), RunE: s.runXferPoolDescribe,
	}
	s.xferPoolListCmd = &cobra.Command{
		Use: "list", Short: "List Transfer Service agent pools",
		Args: cobra.NoArgs, RunE: s.runXferPoolList,
	}
	s.xferPoolUpdateCmd = &cobra.Command{
		Use: "update POOL", Short: "Update a Transfer Service agent pool",
		Args: cobra.ExactArgs(1), RunE: s.runXferPoolUpdate,
	}
	s.xferAgentInstallCmd = &cobra.Command{
		Use: "install", Short: "Install Transfer Service agents locally (Docker)",
		Args: cobra.NoArgs, RunE: s.runXferAgentInstall,
	}
	s.xferAgentDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Stop and remove locally installed transfer agent containers",
		Args: cobra.NoArgs, RunE: s.runXferAgentDelete,
	}
	s.xferJobCreateCmd = &cobra.Command{
		Use: "create SOURCE DESTINATION", Short: "Create a Transfer Service transfer job",
		Args: cobra.ExactArgs(2), RunE: s.runXferJobCreate,
	}
	s.xferJobDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Transfer Service transfer job",
		Args: cobra.ExactArgs(1), RunE: s.runXferJobDelete,
	}
	s.xferJobDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Transfer Service transfer job",
		Args: cobra.ExactArgs(1), RunE: s.runXferJobDescribe,
	}
	s.xferJobListCmd = &cobra.Command{
		Use: "list", Short: "List Transfer Service transfer jobs",
		Args: cobra.NoArgs, RunE: s.runXferJobList,
	}
	s.xferJobUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update a Transfer Service transfer job",
		Args: cobra.ExactArgs(1), RunE: s.runXferJobUpdate,
	}
	s.xferJobRunCmd = &cobra.Command{
		Use: "run JOB", Short: "Run a Transfer Service transfer job now",
		Args: cobra.ExactArgs(1), RunE: s.runXferJobRun,
	}
	s.xferJobMonitorCmd = &cobra.Command{
		Use: "monitor JOB", Short: "Monitor the latest operation for a transfer job",
		Args: cobra.ExactArgs(1), RunE: s.runXferJobMonitor,
	}
	s.xferOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Transfer Service transfer operation",
		Args: cobra.ExactArgs(1), RunE: s.runXferOpCancel,
	}
	s.xferOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Transfer Service transfer operation",
		Args: cobra.ExactArgs(1), RunE: s.runXferOpDescribe,
	}
	s.xferOpListCmd = &cobra.Command{
		Use: "list", Short: "List Transfer Service transfer operations",
		Args: cobra.NoArgs, RunE: s.runXferOpList,
	}
	s.xferOpPauseCmd = &cobra.Command{
		Use: "pause OPERATION", Short: "Pause a Transfer Service transfer operation",
		Args: cobra.ExactArgs(1), RunE: s.runXferOpPause,
	}
	s.xferOpResumeCmd = &cobra.Command{
		Use: "resume OPERATION", Short: "Resume a Transfer Service transfer operation",
		Args: cobra.ExactArgs(1), RunE: s.runXferOpResume,
	}
	s.xferAuthorizeCmd = &cobra.Command{
		Use: "authorize", Short: "Authorize an account for all Transfer Service features",
		Args: cobra.NoArgs, RunE: s.runXferAuthorize,
	}
	s.initTransfer()
	s.initTransferAll()
	return s
}
