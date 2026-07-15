package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v3"
	storagetransfer "google.golang.org/api/storagetransfer/v1"
)

// --- Shared helpers for the Storage Transfer surface (#887-#891) ---

var (
	flagXferFormat        string
	flagXferFilter        string
	flagXferConfigFile    string
	flagXferAsync         bool
	flagXferPageSize      int64
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
)

// xferAgentPoolName returns the fully-qualified agent-pool resource name.
// The Storage Transfer API expects "projects/PROJECT_ID/agentPools/POOL_ID".
func xferAgentPoolName(project, pool string) string {
	if strings.HasPrefix(pool, "projects/") {
		return pool
	}
	return fmt.Sprintf("projects/%s/agentPools/%s", project, pool)
}

// xferJobName returns the fully-qualified transfer-job resource name.
// The Storage Transfer API expects "transferJobs/JOB_NAME".
func xferJobName(name string) string {
	if strings.HasPrefix(name, "transferJobs/") {
		return name
	}
	return "transferJobs/" + name
}

// xferOpName returns the fully-qualified operation resource name.
func xferOpName(name string) string {
	if strings.HasPrefix(name, "transferOperations/") {
		return name
	}
	return "transferOperations/" + name
}

func xferService(ctx context.Context) (*storagetransfer.Service, error) {
	return gcp.StorageTransferService(ctx, flagAccount)
}

// --- Subgroup command objects ---

var (
	xferAgentPoolsCmd = &cobra.Command{Use: "agent-pools", Short: "Manage on-premise transfer agent pools"}
	xferAgentsCmd     = &cobra.Command{Use: "agents", Short: "Manage transfer agents"}
	xferJobsCmd       = &cobra.Command{Use: "jobs", Short: "Manage transfer jobs"}
	xferOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage transfer operations"}
)

// --- agent-pools ---

var (
	xferPoolCreateCmd = &cobra.Command{
		Use: "create POOL", Short: "Create a Transfer Service agent pool",
		Args: cobra.ExactArgs(1), RunE: runXferPoolCreate,
	}
	xferPoolDeleteCmd = &cobra.Command{
		Use: "delete POOL", Short: "Delete a Transfer Service agent pool",
		Args: cobra.ExactArgs(1), RunE: runXferPoolDelete,
	}
	xferPoolDescribeCmd = &cobra.Command{
		Use: "describe POOL", Short: "Describe a Transfer Service agent pool",
		Args: cobra.ExactArgs(1), RunE: runXferPoolDescribe,
	}
	xferPoolListCmd = &cobra.Command{
		Use: "list", Short: "List Transfer Service agent pools",
		Args: cobra.NoArgs, RunE: runXferPoolList,
	}
	xferPoolUpdateCmd = &cobra.Command{
		Use: "update POOL", Short: "Update a Transfer Service agent pool",
		Args: cobra.ExactArgs(1), RunE: runXferPoolUpdate,
	}
)

// --- agents ---

var (
	xferAgentInstallCmd = &cobra.Command{
		Use: "install", Short: "Install Transfer Service agents locally (Docker)",
		Args: cobra.NoArgs, RunE: runXferAgentInstall,
	}
	xferAgentDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Stop and remove locally installed transfer agent containers",
		Args: cobra.NoArgs, RunE: runXferAgentDelete,
	}
)

// --- jobs ---

var (
	xferJobCreateCmd = &cobra.Command{
		Use: "create SOURCE DESTINATION", Short: "Create a Transfer Service transfer job",
		Args: cobra.ExactArgs(2), RunE: runXferJobCreate,
	}
	xferJobDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Transfer Service transfer job",
		Args: cobra.ExactArgs(1), RunE: runXferJobDelete,
	}
	xferJobDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Transfer Service transfer job",
		Args: cobra.ExactArgs(1), RunE: runXferJobDescribe,
	}
	xferJobListCmd = &cobra.Command{
		Use: "list", Short: "List Transfer Service transfer jobs",
		Args: cobra.NoArgs, RunE: runXferJobList,
	}
	xferJobUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update a Transfer Service transfer job",
		Args: cobra.ExactArgs(1), RunE: runXferJobUpdate,
	}
	xferJobRunCmd = &cobra.Command{
		Use: "run JOB", Short: "Run a Transfer Service transfer job now",
		Args: cobra.ExactArgs(1), RunE: runXferJobRun,
	}
	xferJobMonitorCmd = &cobra.Command{
		Use: "monitor JOB", Short: "Monitor the latest operation for a transfer job",
		Args: cobra.ExactArgs(1), RunE: runXferJobMonitor,
	}
)

// --- operations ---

var (
	xferOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Transfer Service transfer operation",
		Args: cobra.ExactArgs(1), RunE: runXferOpCancel,
	}
	xferOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Transfer Service transfer operation",
		Args: cobra.ExactArgs(1), RunE: runXferOpDescribe,
	}
	xferOpListCmd = &cobra.Command{
		Use: "list", Short: "List Transfer Service transfer operations",
		Args: cobra.NoArgs, RunE: runXferOpList,
	}
	xferOpPauseCmd = &cobra.Command{
		Use: "pause OPERATION", Short: "Pause a Transfer Service transfer operation",
		Args: cobra.ExactArgs(1), RunE: runXferOpPause,
	}
	xferOpResumeCmd = &cobra.Command{
		Use: "resume OPERATION", Short: "Resume a Transfer Service transfer operation",
		Args: cobra.ExactArgs(1), RunE: runXferOpResume,
	}
)

// --- authorize ---

var xferAuthorizeCmd = &cobra.Command{
	Use: "authorize", Short: "Authorize an account for all Transfer Service features",
	Args: cobra.NoArgs, RunE: runXferAuthorize,
}

func init() {
	// Common flags.
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagXferFormat, "format", "", "Output format")
		}
	}
	addFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagXferFilter, "filter", "", "Server-side list filter")
		}
	}

	// agent-pools flags
	xferPoolCreateCmd.Flags().StringVar(&flagXferDisplayName, "display-name", "", "Human-readable display name")
	xferPoolCreateCmd.Flags().Int64Var(&flagXferBandwidthMbps, "bandwidth-limit", 0, "Bandwidth limit in MB/s (0 for none)")
	xferPoolCreateCmd.Flags().BoolVar(&flagXferAsync, "no-async", false, "Block until pool is created")
	xferPoolCreateCmd.Flags().StringVar(&flagXferConfigFile, "config-file", "", "Optional JSON/YAML AgentPool body override")
	xferPoolUpdateCmd.Flags().StringVar(&flagXferDisplayName, "display-name", "", "Update the display name")
	xferPoolUpdateCmd.Flags().Int64Var(&flagXferBandwidthMbps, "bandwidth-limit", 0, "Update the bandwidth limit in MB/s")
	xferPoolUpdateCmd.Flags().BoolVar(&flagXferClearName, "clear-display-name", false, "Remove the display name")
	xferPoolUpdateCmd.Flags().BoolVar(&flagXferClearBW, "clear-bandwidth-limit", false, "Remove the bandwidth limit")
	addFmt(xferPoolCreateCmd, xferPoolDescribeCmd, xferPoolListCmd, xferPoolUpdateCmd)
	addFilter(xferPoolListCmd)
	xferAgentPoolsCmd.AddCommand(xferPoolCreateCmd, xferPoolDeleteCmd, xferPoolDescribeCmd, xferPoolListCmd, xferPoolUpdateCmd)
	transferCmd.AddCommand(xferAgentPoolsCmd)

	// agents flags
	xferAgentInstallCmd.Flags().StringVar(&flagXferPool, "pool", "", "Agent pool to associate with the new agent(s) (required)")
	_ = xferAgentInstallCmd.MarkFlagRequired("pool")
	xferAgentInstallCmd.Flags().StringVar(&flagXferCredsFile, "creds-file", "", "Path to a service-account credentials file to mount into the container")
	xferAgentInstallCmd.Flags().IntVar(&flagXferCount, "count", 1, "Number of agent containers to start")
	xferAgentInstallCmd.Flags().StringVar(&flagXferIDPrefix, "id-prefix", "", "Optional prefix for the agent ID")
	xferAgentInstallCmd.Flags().StringVar(&flagXferLogsDirectory, "logs-directory", "/tmp", "Directory to mount for agent logs")
	xferAgentInstallCmd.Flags().Int64Var(&flagXferMemlockLimit, "memlock-limit", 64000000, "Container memlock ulimit")
	xferAgentInstallCmd.Flags().StringSliceVar(&flagXferMountDirs, "mount-directories", nil, "Directories to bind-mount into the container")
	xferAgentInstallCmd.Flags().StringVar(&flagXferProxy, "proxy", "", "HTTPS proxy URL")
	xferAgentInstallCmd.Flags().BoolVar(&flagXferS3Compatible, "s3-compatible-mode", false, "Enable S3-compatible source mode")
	xferAgentDeleteCmd.Flags().StringVar(&flagXferAgentID, "id", "", "Container ID to stop (either --id or --all is required)")
	xferAgentDeleteCmd.Flags().BoolVar(&flagXferAgentAll, "all", false, "Stop all local tsop-agent containers")
	xferAgentDeleteCmd.Flags().BoolVar(&flagXferUninstall, "uninstall", false, "Also remove the container image after stopping")
	xferAgentsCmd.AddCommand(xferAgentInstallCmd, xferAgentDeleteCmd)
	transferCmd.AddCommand(xferAgentsCmd)

	// jobs flags
	xferJobCreateCmd.Flags().StringVar(&flagXferJobName, "name", "", "Unique job identifier (auto-generated when empty)")
	xferJobCreateCmd.Flags().StringVar(&flagXferJobDesc, "description", "", "Optional job description")
	xferJobCreateCmd.Flags().StringVar(&flagXferSrcAgentPool, "source-agent-pool", "", "Source POSIX/HDFS agent pool")
	xferJobCreateCmd.Flags().StringVar(&flagXferDstAgentPool, "destination-agent-pool", "", "Destination POSIX agent pool")
	xferJobCreateCmd.Flags().StringVar(&flagXferSrcCredsFile, "source-creds-file", "", "Path to source credentials file")
	xferJobCreateCmd.Flags().StringVar(&flagXferSchedRepeats, "schedule-repeats-every", "", "Recurrence interval (e.g. 1d, 12h); omit for one-time transfer")
	xferJobCreateCmd.Flags().BoolVar(&flagXferAsync, "no-async", false, "Wait for the first operation to finish")
	xferJobCreateCmd.Flags().StringVar(&flagXferConfigFile, "config-file", "", "JSON/YAML TransferJob body override")
	xferJobUpdateCmd.Flags().StringVar(&flagXferJobStatus, "status", "", "Job status (enabled, disabled, deleted)")
	xferJobUpdateCmd.Flags().StringVar(&flagXferJobDesc, "description", "", "Update job description")
	xferJobUpdateCmd.Flags().StringVar(&flagXferConfigFile, "config-file", "", "JSON/YAML TransferJob body override")
	xferJobRunCmd.Flags().BoolVar(&flagXferAsync, "no-async", false, "Wait for the operation to finish")
	addFmt(xferJobCreateCmd, xferJobDescribeCmd, xferJobListCmd, xferJobUpdateCmd, xferJobRunCmd, xferJobMonitorCmd)
	addFilter(xferJobListCmd)
	xferJobsCmd.AddCommand(xferJobCreateCmd, xferJobDeleteCmd, xferJobDescribeCmd, xferJobListCmd,
		xferJobUpdateCmd, xferJobRunCmd, xferJobMonitorCmd)
	transferCmd.AddCommand(xferJobsCmd)

	// operations flags
	addFmt(xferOpDescribeCmd, xferOpListCmd)
	addFilter(xferOpListCmd)
	xferOperationsCmd.AddCommand(xferOpCancelCmd, xferOpDescribeCmd, xferOpListCmd, xferOpPauseCmd, xferOpResumeCmd)
	transferCmd.AddCommand(xferOperationsCmd)

	// authorize
	xferAuthorizeCmd.Flags().StringVar(&flagXferCredsFile, "creds-file", "", "Path to a service-account credentials file to authorize (defaults to logged-in gcloud account)")
	xferAuthorizeCmd.Flags().BoolVar(&flagXferAddMissing, "add-missing", false, "Grant any missing IAM roles instead of just reporting them")
	transferCmd.AddCommand(xferAuthorizeCmd)
}

// --- agent-pools impl ---

func runXferPoolCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	body := &storagetransfer.AgentPool{Name: xferAgentPoolName(project, args[0])}
	if flagXferConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagXferConfigFile, body); err != nil {
			return err
		}
	}
	if flagXferDisplayName != "" {
		body.DisplayName = flagXferDisplayName
	}
	if flagXferBandwidthMbps > 0 {
		body.BandwidthLimit = &storagetransfer.BandwidthLimit{LimitMbps: flagXferBandwidthMbps}
	}
	call := svc.Projects.AgentPools.Create(project, body).AgentPoolId(args[0]).Context(ctx)
	created, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating agent pool: %w", err)
	}
	if flagXferAsync {
		fmt.Fprintf(os.Stderr, "Waiting for agent pool %q to be CREATED...\n", args[0])
		final, err := xferWaitPool(ctx, svc, xferAgentPoolName(project, args[0]))
		if err != nil {
			return err
		}
		return emitFormatted(final, flagXferFormat)
	}
	return emitFormatted(created, flagXferFormat)
}

func xferWaitPool(ctx context.Context, svc *storagetransfer.Service, name string) (*storagetransfer.AgentPool, error) {
	for {
		got, err := svc.Projects.AgentPools.Get(name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling agent pool: %w", err)
		}
		if got.State == "CREATED" {
			return got, nil
		}
		if got.State == "DELETED" {
			return got, fmt.Errorf("agent pool %s is DELETED", name)
		}
		time.Sleep(2 * time.Second)
	}
}

func runXferPoolDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	_, err = svc.Projects.AgentPools.Delete(xferAgentPoolName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting agent pool: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deleted agent pool [%s].\n", args[0])
	return nil
}

func runXferPoolDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.AgentPools.Get(xferAgentPoolName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing agent pool: %w", err)
	}
	return emitFormatted(got, flagXferFormat)
}

func runXferPoolList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	var all []*storagetransfer.AgentPool
	pageToken := ""
	for {
		call := svc.Projects.AgentPools.List(project).Context(ctx)
		if flagXferFilter != "" {
			call = call.Filter(flagXferFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing agent pools: %w", err)
		}
		all = append(all, resp.AgentPools...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagXferFormat != "" {
		return emitFormatted(all, flagXferFormat)
	}
	fmt.Printf("%-40s %-16s %s\n", "NAME", "STATE", "DISPLAY_NAME")
	for _, p := range all {
		fmt.Printf("%-40s %-16s %s\n", path.Base(p.Name), p.State, p.DisplayName)
	}
	return nil
}

func runXferPoolUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	name := xferAgentPoolName(project, args[0])
	body := &storagetransfer.AgentPool{}
	var mask []string
	if flagXferBandwidthMbps > 0 || flagXferClearBW {
		mask = append(mask, "bandwidth_limit")
		if flagXferBandwidthMbps > 0 {
			body.BandwidthLimit = &storagetransfer.BandwidthLimit{LimitMbps: flagXferBandwidthMbps}
		}
	}
	if flagXferDisplayName != "" || flagXferClearName {
		mask = append(mask, "display_name")
		body.DisplayName = flagXferDisplayName
	}
	call := svc.Projects.AgentPools.Patch(name, body).Context(ctx)
	if len(mask) > 0 {
		call = call.UpdateMask(strings.Join(mask, ","))
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating agent pool: %w", err)
	}
	return emitFormatted(got, flagXferFormat)
}

// --- agents impl ---

const xferAgentImage = "gcr.io/cloud-ingest/tsop-agent:latest"

func runXferAgentInstall(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagXferCount < 1 {
		return fmt.Errorf("--count must be >= 1")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found on PATH: install Docker before running this command (see https://docs.docker.com/engine/install/)")
	}
	// Verify the agent pool exists (mirrors python behavior).
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	pool, err := svc.Projects.AgentPools.Get(xferAgentPoolName(project, flagXferPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up agent pool %q: %w", flagXferPool, err)
	}
	if pool.State != "CREATED" {
		return fmt.Errorf("agent pool %q is not in state CREATED (state=%s)", flagXferPool, pool.State)
	}
	credsPath := flagXferCredsFile
	if credsPath != "" {
		if abs, err := filepath.Abs(credsPath); err == nil {
			credsPath = abs
		}
		if _, err := os.Stat(credsPath); err != nil {
			return fmt.Errorf("credentials file not found: %w", err)
		}
	}
	logsDir := flagXferLogsDirectory
	if abs, err := filepath.Abs(logsDir); err == nil {
		logsDir = abs
	}
	fmt.Fprintln(os.Stderr, "[1/3] Credentials found")
	fmt.Fprintln(os.Stderr, "[2/3] Docker found")
	for i := 0; i < flagXferCount; i++ {
		containerCmd := xferBuildAgentDockerCmd(project, credsPath, logsDir, i)
		fmt.Fprintf(os.Stderr, "Running: %s\n", strings.Join(containerCmd, " "))
		run := exec.Command(containerCmd[0], containerCmd[1:]...)
		run.Stdout = os.Stderr
		run.Stderr = os.Stderr
		if err := run.Run(); err != nil {
			return fmt.Errorf("docker run failed: %w", err)
		}
	}
	fmt.Fprintln(os.Stderr, "[3/3] Agent installation complete")
	fmt.Fprintf(os.Stderr, "Check status at https://console.cloud.google.com/transfer/on-premises/agent-pools/pool/%s/agents?project=%s\n",
		flagXferPool, project)
	return nil
}

func xferBuildAgentDockerCmd(project, credsPath, logsDir string, idx int) []string {
	c := []string{"docker", "run",
		"--ulimit", fmt.Sprintf("memlock=%d", flagXferMemlockLimit),
		"--rm", "-d",
	}
	if flagXferProxy != "" {
		c = append(c, "--env", "HTTPS_PROXY="+flagXferProxy)
	}
	if len(flagXferMountDirs) == 0 {
		c = append(c, "-v=/:/transfer_root")
	} else {
		c = append(c, "-v="+logsDir+":/tmp")
		if credsPath != "" {
			c = append(c, "-v="+credsPath+":"+credsPath)
		}
		for _, d := range flagXferMountDirs {
			c = append(c, "-v="+d+":"+d)
		}
	}
	c = append(c, xferAgentImage,
		"--agent-pool="+flagXferPool,
		"--log-dir="+logsDir,
		"--project-id="+project,
	)
	if credsPath != "" {
		c = append(c, "--creds-file="+credsPath)
	}
	if len(flagXferMountDirs) == 0 {
		c = append(c, "--enable-mount-directory")
	}
	if flagXferIDPrefix != "" {
		suffix := ""
		if flagXferCount > 1 {
			suffix = fmt.Sprintf("%d", idx)
		}
		c = append(c, "--agent-id-prefix="+flagXferIDPrefix+suffix)
	}
	if flagXferS3Compatible {
		c = append(c, "--enable-s3")
	}
	return c
}

func runXferAgentDelete(cmd *cobra.Command, args []string) error {
	if !flagXferAgentAll && flagXferAgentID == "" {
		return fmt.Errorf("either --id or --all is required")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found on PATH: %w", err)
	}
	targets := []string{}
	if flagXferAgentAll {
		out, err := exec.Command("docker", "ps", "-q", "--filter", "ancestor="+xferAgentImage).CombinedOutput()
		if err != nil {
			return fmt.Errorf("listing agent containers: %w: %s", err, string(out))
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				targets = append(targets, line)
			}
		}
		if len(targets) == 0 {
			fmt.Fprintln(os.Stderr, "No local transfer-agent containers found.")
			return nil
		}
	} else {
		targets = append(targets, flagXferAgentID)
	}
	for _, id := range targets {
		fmt.Fprintf(os.Stderr, "Stopping container %s...\n", id)
		out, err := exec.Command("docker", "stop", id).CombinedOutput()
		if err != nil {
			return fmt.Errorf("stopping container %s: %w: %s", id, err, string(out))
		}
	}
	if flagXferUninstall {
		out, err := exec.Command("docker", "rmi", xferAgentImage).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not remove image %s: %v: %s\n", xferAgentImage, err, string(out))
		}
	}
	return nil
}

// --- jobs impl ---

func runXferJobCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	job := &storagetransfer.TransferJob{ProjectId: project}
	if flagXferConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagXferConfigFile, job); err != nil {
			return err
		}
	}
	if job.TransferSpec == nil {
		job.TransferSpec = &storagetransfer.TransferSpec{}
	}
	if err := xferPopulateSpec(job.TransferSpec, args[0], args[1], project); err != nil {
		return err
	}
	if flagXferJobDesc != "" {
		job.Description = flagXferJobDesc
	}
	if flagXferJobName != "" {
		job.Name = xferJobName(flagXferJobName)
	}
	if flagXferSchedRepeats != "" {
		d, err := time.ParseDuration(strings.Replace(flagXferSchedRepeats, "d", "*24h", 1))
		if err != nil {
			return fmt.Errorf("invalid --schedule-repeats-every: %w", err)
		}
		if job.Schedule == nil {
			job.Schedule = &storagetransfer.Schedule{}
		}
		job.Schedule.RepeatInterval = fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if job.Status == "" {
		job.Status = "ENABLED"
	}
	created, err := svc.TransferJobs.Create(job).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating transfer job: %w", err)
	}
	if flagXferAsync {
		fmt.Fprintf(os.Stderr, "Created job: %s\n", created.Name)
	}
	return emitFormatted(created, flagXferFormat)
}

// xferPopulateSpec fills common source/destination fields into spec based on
// URL-shaped positional arguments (gs://, s3://, http(s)://, posix:///,
// hdfs://). Existing values from --config-file are preserved.
func xferPopulateSpec(spec *storagetransfer.TransferSpec, source, dest, project string) error {
	if err := xferApplyEndpoint(spec, source, true); err != nil {
		return err
	}
	if err := xferApplyEndpoint(spec, dest, false); err != nil {
		return err
	}
	if flagXferSrcAgentPool != "" {
		spec.SourceAgentPoolName = xferAgentPoolName(project, flagXferSrcAgentPool)
	}
	if flagXferDstAgentPool != "" {
		spec.SinkAgentPoolName = xferAgentPoolName(project, flagXferDstAgentPool)
	}
	return nil
}

func xferApplyEndpoint(spec *storagetransfer.TransferSpec, url string, isSource bool) error {
	switch {
	case strings.HasPrefix(url, "gs://"):
		bucket, prefix := splitBucketPrefix(url[len("gs://"):])
		data := &storagetransfer.GcsData{BucketName: bucket, Path: prefix}
		if isSource {
			spec.GcsDataSource = data
		} else {
			spec.GcsDataSink = data
		}
	case strings.HasPrefix(url, "s3://"):
		if !isSource {
			return fmt.Errorf("s3:// only supported as source")
		}
		bucket, prefix := splitBucketPrefix(url[len("s3://"):])
		spec.AwsS3DataSource = &storagetransfer.AwsS3Data{BucketName: bucket, Path: prefix}
	case strings.HasPrefix(url, "http://"), strings.HasPrefix(url, "https://"):
		if !isSource {
			return fmt.Errorf("http(s):// only supported as source")
		}
		spec.HttpDataSource = &storagetransfer.HttpData{ListUrl: url}
	case strings.HasPrefix(url, "posix:///"):
		p := &storagetransfer.PosixFilesystem{RootDirectory: url[len("posix://"):]}
		if isSource {
			spec.PosixDataSource = p
		} else {
			spec.PosixDataSink = p
		}
	case strings.HasPrefix(url, "hdfs://"):
		if !isSource {
			return fmt.Errorf("hdfs:// only supported as source")
		}
		spec.HdfsDataSource = &storagetransfer.HdfsData{Path: url[len("hdfs://"):]}
	default:
		return fmt.Errorf("unsupported %s scheme: %s (use gs://, s3://, http(s)://, posix:///, or hdfs://)",
			endpointRole(isSource), url)
	}
	return nil
}

func endpointRole(isSource bool) string {
	if isSource {
		return "source"
	}
	return "destination"
}

func splitBucketPrefix(s string) (string, string) {
	s = strings.TrimSuffix(s, "/")
	i := strings.Index(s, "/")
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+1:] + "/"
}

func runXferJobDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.TransferJobs.Delete(xferJobName(args[0]), project).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting transfer job: %w", err)
	}
	got, err := svc.TransferJobs.Get(xferJobName(args[0]), project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("re-fetching transfer job after delete: %w", err)
	}
	return emitFormatted(got, flagXferFormat)
}

func runXferJobDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.TransferJobs.Get(xferJobName(args[0]), project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing transfer job: %w", err)
	}
	return emitFormatted(got, flagXferFormat)
}

func runXferJobList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	// The Storage Transfer API requires a projectId filter (JSON string).
	filter := fmt.Sprintf(`{"projectId":"%s"}`, project)
	if flagXferFilter != "" {
		filter = flagXferFilter
	}
	var all []*storagetransfer.TransferJob
	pageToken := ""
	for {
		call := svc.TransferJobs.List(filter).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing transfer jobs: %w", err)
		}
		all = append(all, resp.TransferJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagXferFormat != "" {
		return emitFormatted(all, flagXferFormat)
	}
	fmt.Printf("%-40s %-10s %s\n", "NAME", "STATUS", "DESCRIPTION")
	for _, j := range all {
		fmt.Printf("%-40s %-10s %s\n", strings.TrimPrefix(j.Name, "transferJobs/"), j.Status, j.Description)
	}
	return nil
}

func runXferJobUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	job := &storagetransfer.TransferJob{ProjectId: project}
	var mask []string
	if flagXferConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagXferConfigFile, job); err != nil {
			return err
		}
		mask = nonEmptyJSONFields(job)
	}
	if flagXferJobStatus != "" {
		job.Status = strings.ToUpper(flagXferJobStatus)
		mask = append(mask, "status")
	}
	if flagXferJobDesc != "" {
		job.Description = flagXferJobDesc
		mask = append(mask, "description")
	}
	req := &storagetransfer.UpdateTransferJobRequest{
		ProjectId:      project,
		TransferJob:    job,
		UpdateTransferJobFieldMask: strings.Join(dedupe(mask), ","),
	}
	got, err := svc.TransferJobs.Patch(xferJobName(args[0]), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating transfer job: %w", err)
	}
	return emitFormatted(got, flagXferFormat)
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func runXferJobRun(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.TransferJobs.Run(xferJobName(args[0]),
		&storagetransfer.RunTransferJobRequest{ProjectId: project}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running transfer job: %w", err)
	}
	if flagXferAsync {
		final, err := xferWaitOp(ctx, svc, op.Name)
		if err != nil {
			return err
		}
		return emitFormatted(final, flagXferFormat)
	}
	return emitFormatted(op, flagXferFormat)
}

func xferWaitOp(ctx context.Context, svc *storagetransfer.Service, name string) (*storagetransfer.Operation, error) {
	for {
		got, err := svc.TransferOperations.Get(name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", name, err)
		}
		if got.Done {
			if got.Error != nil {
				return got, fmt.Errorf("operation %s failed: %s", name, got.Error.Message)
			}
			return got, nil
		}
		time.Sleep(3 * time.Second)
	}
}

func runXferJobMonitor(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	// Wait until the job produces an operation, then stream progress until done.
	var opName string
	for opName == "" {
		got, err := svc.TransferJobs.Get(xferJobName(args[0]), project).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("fetching transfer job: %w", err)
		}
		opName = got.LatestOperationName
		if opName == "" {
			time.Sleep(3 * time.Second)
		}
	}
	fmt.Fprintf(os.Stderr, "Monitoring operation: %s\n", opName)
	for {
		got, err := svc.TransferOperations.Get(opName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if err := emitFormatted(got, flagXferFormat); err != nil {
			return err
		}
		if got.Done {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

// --- operations impl ---

func runXferOpCancel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	_, err = svc.TransferOperations.Cancel(xferOpName(args[0]),
		&storagetransfer.CancelOperationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Cancel requested for operation [%s].\n", args[0])
	return nil
}

func runXferOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.TransferOperations.Get(xferOpName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagXferFormat)
}

func runXferOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	filter := fmt.Sprintf(`{"projectId":"%s"}`, project)
	if flagXferFilter != "" {
		filter = flagXferFilter
	}
	var all []*storagetransfer.Operation
	pageToken := ""
	for {
		call := svc.TransferOperations.List("transferOperations", filter).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagXferFormat != "" {
		return emitFormatted(all, flagXferFormat)
	}
	fmt.Printf("%-60s %-6s\n", "NAME", "DONE")
	for _, op := range all {
		fmt.Printf("%-60s %-6t\n", op.Name, op.Done)
	}
	return nil
}

func runXferOpPause(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	_, err = svc.TransferOperations.Pause(xferOpName(args[0]),
		&storagetransfer.PauseTransferOperationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("pausing operation: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Paused operation [%s].\n", args[0])
	return nil
}

func runXferOpResume(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := xferService(ctx)
	if err != nil {
		return err
	}
	_, err = svc.TransferOperations.Resume(xferOpName(args[0]),
		&storagetransfer.ResumeTransferOperationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming operation: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Resumed operation [%s].\n", args[0])
	return nil
}

// --- authorize impl ---

var (
	xferExpectedUserRoles = []string{
		"roles/owner",
		"roles/storagetransfer.admin",
		"roles/storagetransfer.transferAgent",
		"roles/storage.objectAdmin",
		"roles/pubsub.editor",
	}
	xferExpectedP4SARoles = []string{
		"roles/storage.admin",
		"roles/storagetransfer.serviceAgent",
	}
)

func runXferAuthorize(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	xfer, err := xferService(ctx)
	if err != nil {
		return err
	}
	crm, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	accountEmail, isSA, err := xferResolveAuthorizeAccount(ctx)
	if err != nil {
		return err
	}
	userMember := xferIamMemberFor(accountEmail, isSA)

	policy, err := crm.Projects.GetIamPolicy("projects/"+project,
		&cloudresourcemanager.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching project IAM policy: %w", err)
	}

	missingUser := xferMissingRoles(policy, userMember, xferExpectedUserRoles)
	fmt.Fprintf(os.Stderr, "User %s missing roles: %v\n", accountEmail, missingUser)

	p4sa, err := xfer.GoogleServiceAccounts.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching transfer service account: %w", err)
	}
	p4saMember := xferIamMemberFor(p4sa.AccountEmail, true)
	missingP4SA := xferMissingRoles(policy, p4saMember, xferExpectedP4SARoles)
	fmt.Fprintf(os.Stderr, "Transfer service account %s missing roles: %v\n", p4sa.AccountEmail, missingP4SA)

	toAdd := map[string][]string{
		userMember: missingUser,
		p4saMember: missingP4SA,
	}
	total := len(missingUser) + len(missingP4SA)
	if total == 0 {
		fmt.Fprintln(os.Stderr, "No missing roles.")
		return nil
	}
	if !flagXferAddMissing {
		fmt.Fprintln(os.Stderr, "Re-run with --add-missing to grant the missing roles.")
		return nil
	}
	for member, roles := range toAdd {
		for _, role := range roles {
			xferAddBinding(policy, member, role)
		}
	}
	if _, err := crm.Projects.SetIamPolicy("projects/"+project,
		&cloudresourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("granting missing roles: %w", err)
	}
	fmt.Fprintln(os.Stderr, "Missing roles granted. Permissions can take up to 7 minutes to propagate.")
	return nil
}

func xferResolveAuthorizeAccount(ctx context.Context) (string, bool, error) {
	// If --creds-file supplied, parse client_email/type from JSON body.
	if flagXferCredsFile != "" {
		data, err := os.ReadFile(flagXferCredsFile)
		if err != nil {
			return "", false, fmt.Errorf("reading --creds-file: %w", err)
		}
		var parsed struct {
			Type        string `json:"type"`
			ClientEmail string `json:"client_email"`
		}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return "", false, fmt.Errorf("parsing --creds-file: %w", err)
		}
		return parsed.ClientEmail, parsed.Type == "service_account", nil
	}
	if flagAccount != "" {
		return flagAccount, strings.HasSuffix(flagAccount, ".gserviceaccount.com"), nil
	}
	return "", false, fmt.Errorf("no --creds-file supplied and --account is empty: specify --creds-file or --account")
}

func xferIamMemberFor(email string, isSA bool) string {
	if email == "" {
		return ""
	}
	if strings.Contains(email, ":") {
		return email
	}
	prefix := "user"
	if isSA || strings.HasSuffix(email, ".gserviceaccount.com") {
		prefix = "serviceAccount"
	}
	return prefix + ":" + email
}

func xferMissingRoles(policy *cloudresourcemanager.Policy, member string, expected []string) []string {
	have := map[string]bool{}
	for _, b := range policy.Bindings {
		for _, m := range b.Members {
			if m == member {
				have[b.Role] = true
				break
			}
		}
	}
	var missing []string
	for _, r := range expected {
		if !have[r] {
			missing = append(missing, r)
		}
	}
	return missing
}

func xferAddBinding(policy *cloudresourcemanager.Policy, member, role string) {
	for _, b := range policy.Bindings {
		if b.Role == role {
			for _, m := range b.Members {
				if m == member {
					return
				}
			}
			b.Members = append(b.Members, member)
			return
		}
	}
	policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
		Role:    role,
		Members: []string{member},
	})
}
