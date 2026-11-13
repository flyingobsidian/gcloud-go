package cli

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

	"github.com/spf13/cobra"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v3"
	storagetransfer "google.golang.org/api/storagetransfer/v1"
)

// --- Shared helpers for the Storage Transfer surface (#887-#891) ---

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

func (s *transferCmds) xferService(ctx context.Context) (*storagetransfer.Service, error) {
	return s.inv.clients().StorageTransfer(ctx)
}

// --- Subgroup command objects ---

// --- agent-pools ---

// --- agents ---

// --- jobs ---

// --- operations ---

// --- authorize ---

func (s *transferCmds) initTransferAll() {
	// Common flags.
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagXferFormat, "format", "", "Output format")
		}
	}
	addFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagXferFilter, "filter", "", "Server-side list filter")
		}
	}

	// agent-pools flags
	s.xferPoolCreateCmd.Flags().StringVar(&s.flagXferDisplayName, "display-name", "", "Human-readable display name")
	s.xferPoolCreateCmd.Flags().Int64Var(&s.flagXferBandwidthMbps, "bandwidth-limit", 0, "Bandwidth limit in MB/s (0 for none)")
	s.xferPoolCreateCmd.Flags().BoolVar(&s.flagXferAsync, "no-async", false, "Block until pool is created")
	s.xferPoolCreateCmd.Flags().StringVar(&s.flagXferConfigFile, "config-file", "", "Optional JSON/YAML AgentPool body override")
	s.xferPoolUpdateCmd.Flags().StringVar(&s.flagXferDisplayName, "display-name", "", "Update the display name")
	s.xferPoolUpdateCmd.Flags().Int64Var(&s.flagXferBandwidthMbps, "bandwidth-limit", 0, "Update the bandwidth limit in MB/s")
	s.xferPoolUpdateCmd.Flags().BoolVar(&s.flagXferClearName, "clear-display-name", false, "Remove the display name")
	s.xferPoolUpdateCmd.Flags().BoolVar(&s.flagXferClearBW, "clear-bandwidth-limit", false, "Remove the bandwidth limit")
	addFmt(s.xferPoolCreateCmd, s.xferPoolDescribeCmd, s.xferPoolListCmd, s.xferPoolUpdateCmd)
	addFilter(s.xferPoolListCmd)
	s.xferAgentPoolsCmd.AddCommand(s.xferPoolCreateCmd, s.xferPoolDeleteCmd, s.xferPoolDescribeCmd, s.xferPoolListCmd, s.xferPoolUpdateCmd)
	s.transferCmd.AddCommand(s.xferAgentPoolsCmd)

	// agents flags
	s.xferAgentInstallCmd.Flags().StringVar(&s.flagXferPool, "pool", "", "Agent pool to associate with the new agent(s) (required)")
	_ = s.xferAgentInstallCmd.MarkFlagRequired("pool")
	s.xferAgentInstallCmd.Flags().StringVar(&s.flagXferCredsFile, "creds-file", "", "Path to a service-account credentials file to mount into the container")
	s.xferAgentInstallCmd.Flags().IntVar(&s.flagXferCount, "count", 1, "Number of agent containers to start")
	s.xferAgentInstallCmd.Flags().StringVar(&s.flagXferIDPrefix, "id-prefix", "", "Optional prefix for the agent ID")
	s.xferAgentInstallCmd.Flags().StringVar(&s.flagXferLogsDirectory, "logs-directory", "/tmp", "Directory to mount for agent logs")
	s.xferAgentInstallCmd.Flags().Int64Var(&s.flagXferMemlockLimit, "memlock-limit", 64000000, "Container memlock ulimit")
	s.xferAgentInstallCmd.Flags().StringSliceVar(&s.flagXferMountDirs, "mount-directories", nil, "Directories to bind-mount into the container")
	s.xferAgentInstallCmd.Flags().StringVar(&s.flagXferProxy, "proxy", "", "HTTPS proxy URL")
	s.xferAgentInstallCmd.Flags().BoolVar(&s.flagXferS3Compatible, "s3-compatible-mode", false, "Enable S3-compatible source mode (deprecated)")
	_ = s.xferAgentInstallCmd.Flags().MarkDeprecated("s3-compatible-mode",
		"Transfer Service now auto-detects S3-compatible sources; this flag is a no-op.")
	s.xferAgentDeleteCmd.Flags().StringVar(&s.flagXferAgentID, "id", "", "Container ID to stop (either --id or --all is required)")
	s.xferAgentDeleteCmd.Flags().BoolVar(&s.flagXferAgentAll, "all", false, "Stop all local tsop-agent containers")
	s.xferAgentDeleteCmd.Flags().BoolVar(&s.flagXferUninstall, "uninstall", false, "Also remove the container image after stopping")
	s.xferAgentsCmd.AddCommand(s.xferAgentInstallCmd, s.xferAgentDeleteCmd)
	s.transferCmd.AddCommand(s.xferAgentsCmd)

	// jobs flags
	s.xferJobCreateCmd.Flags().StringVar(&s.flagXferJobName, "name", "", "Unique job identifier (auto-generated when empty)")
	s.xferJobCreateCmd.Flags().StringVar(&s.flagXferJobDesc, "description", "", "Optional job description")
	s.xferJobCreateCmd.Flags().StringVar(&s.flagXferSrcAgentPool, "source-agent-pool", "", "Source POSIX/HDFS agent pool")
	s.xferJobCreateCmd.Flags().StringVar(&s.flagXferDstAgentPool, "destination-agent-pool", "", "Destination POSIX agent pool")
	s.xferJobCreateCmd.Flags().StringVar(&s.flagXferSrcCredsFile, "source-creds-file", "", "Path to source credentials file")
	s.xferJobCreateCmd.Flags().StringVar(&s.flagXferSchedRepeats, "schedule-repeats-every", "", "Recurrence interval (e.g. 1d, 12h); omit for one-time transfer")
	s.xferJobCreateCmd.Flags().BoolVar(&s.flagXferAsync, "no-async", false, "Wait for the first operation to finish")
	s.xferJobCreateCmd.Flags().StringVar(&s.flagXferConfigFile, "config-file", "", "JSON/YAML TransferJob body override")
	s.xferJobUpdateCmd.Flags().StringVar(&s.flagXferJobStatus, "status", "", "Job status (enabled, disabled, deleted)")
	s.xferJobUpdateCmd.Flags().StringVar(&s.flagXferJobDesc, "description", "", "Update job description")
	s.xferJobUpdateCmd.Flags().StringVar(&s.flagXferConfigFile, "config-file", "", "JSON/YAML TransferJob body override")
	s.xferJobRunCmd.Flags().BoolVar(&s.flagXferAsync, "no-async", false, "Wait for the operation to finish")
	addFmt(s.xferJobCreateCmd, s.xferJobDescribeCmd, s.xferJobListCmd, s.xferJobUpdateCmd, s.xferJobRunCmd, s.xferJobMonitorCmd)
	addFilter(s.xferJobListCmd)
	s.xferJobsCmd.AddCommand(s.xferJobCreateCmd, s.xferJobDeleteCmd, s.xferJobDescribeCmd, s.xferJobListCmd,
		s.xferJobUpdateCmd, s.xferJobRunCmd, s.xferJobMonitorCmd)
	s.transferCmd.AddCommand(s.xferJobsCmd)

	// operations flags
	addFmt(s.xferOpDescribeCmd, s.xferOpListCmd)
	addFilter(s.xferOpListCmd)
	s.xferOperationsCmd.AddCommand(s.xferOpCancelCmd, s.xferOpDescribeCmd, s.xferOpListCmd, s.xferOpPauseCmd, s.xferOpResumeCmd)
	s.transferCmd.AddCommand(s.xferOperationsCmd)

	// authorize
	s.xferAuthorizeCmd.Flags().StringVar(&s.flagXferCredsFile, "creds-file", "", "Path to a service-account credentials file to authorize (defaults to logged-in gcloud account)")
	s.xferAuthorizeCmd.Flags().BoolVar(&s.flagXferAddMissing, "add-missing", false, "Grant any missing IAM roles instead of just reporting them")
	s.transferCmd.AddCommand(s.xferAuthorizeCmd)
}

// --- agent-pools impl ---

func (s *transferCmds) runXferPoolCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	body := &storagetransfer.AgentPool{Name: xferAgentPoolName(project, args[0])}
	if s.flagXferConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagXferConfigFile, body); err != nil {
			return err
		}
	}
	if s.flagXferDisplayName != "" {
		body.DisplayName = s.flagXferDisplayName
	}
	if s.flagXferBandwidthMbps > 0 {
		body.BandwidthLimit = &storagetransfer.BandwidthLimit{LimitMbps: s.flagXferBandwidthMbps}
	}
	call := svc.Projects.AgentPools.Create(project, body).AgentPoolId(args[0]).Context(ctx)
	created, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating agent pool: %w", err)
	}
	if s.flagXferAsync {
		fmt.Fprintf(s.inv.stderr, "Waiting for agent pool %q to be CREATED...\n", args[0])
		final, err := xferWaitPool(ctx, svc, xferAgentPoolName(project, args[0]))
		if err != nil {
			return err
		}
		return emitFormattedTo(s.inv.stdout, final, s.flagXferFormat)
	}
	return emitFormattedTo(s.inv.stdout, created, s.flagXferFormat)
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

func (s *transferCmds) runXferPoolDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	_, err = svc.Projects.AgentPools.Delete(xferAgentPoolName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting agent pool: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Deleted agent pool [%s].\n", args[0])
	return nil
}

func (s *transferCmds) runXferPoolDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.AgentPools.Get(xferAgentPoolName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing agent pool: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagXferFormat)
}

func (s *transferCmds) runXferPoolList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	var all []*storagetransfer.AgentPool
	pageToken := ""
	for {
		call := svc.Projects.AgentPools.List(project).Context(ctx)
		if s.flagXferFilter != "" {
			call = call.Filter(s.flagXferFilter)
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
	if s.flagXferFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagXferFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-16s %s\n", "NAME", "STATE", "DISPLAY_NAME")
	for _, p := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-16s %s\n", path.Base(p.Name), p.State, p.DisplayName)
	}
	return nil
}

func (s *transferCmds) runXferPoolUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	name := xferAgentPoolName(project, args[0])
	body := &storagetransfer.AgentPool{}
	var mask []string
	if s.flagXferBandwidthMbps > 0 || s.flagXferClearBW {
		mask = append(mask, "bandwidth_limit")
		if s.flagXferBandwidthMbps > 0 {
			body.BandwidthLimit = &storagetransfer.BandwidthLimit{LimitMbps: s.flagXferBandwidthMbps}
		}
	}
	if s.flagXferDisplayName != "" || s.flagXferClearName {
		mask = append(mask, "display_name")
		body.DisplayName = s.flagXferDisplayName
	}
	call := svc.Projects.AgentPools.Patch(name, body).Context(ctx)
	if len(mask) > 0 {
		call = call.UpdateMask(strings.Join(mask, ","))
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating agent pool: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagXferFormat)
}

// --- agents impl ---

const xferAgentImage = "gcr.io/cloud-ingest/tsop-agent:latest"

func (s *transferCmds) runXferAgentInstall(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if s.flagXferCount < 1 {
		return fmt.Errorf("--count must be >= 1")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found on PATH: install Docker before running this command (see https://docs.docker.com/engine/install/)")
	}
	// Verify the agent pool exists (mirrors python behavior).
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	pool, err := svc.Projects.AgentPools.Get(xferAgentPoolName(project, s.flagXferPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up agent pool %q: %w", s.flagXferPool, err)
	}
	if pool.State != "CREATED" {
		return fmt.Errorf("agent pool %q is not in state CREATED (state=%s)", s.flagXferPool, pool.State)
	}
	credsPath := s.flagXferCredsFile
	if credsPath != "" {
		if abs, err := filepath.Abs(credsPath); err == nil {
			credsPath = abs
		}
		if _, err := os.Stat(credsPath); err != nil {
			return fmt.Errorf("credentials file not found: %w", err)
		}
	}
	logsDir := s.flagXferLogsDirectory
	if abs, err := filepath.Abs(logsDir); err == nil {
		logsDir = abs
	}
	fmt.Fprintln(s.inv.stderr, "[1/3] Credentials found")
	fmt.Fprintln(s.inv.stderr, "[2/3] Docker found")
	for i := 0; i < s.flagXferCount; i++ {
		containerCmd := s.xferBuildAgentDockerCmd(project, credsPath, logsDir, i)
		fmt.Fprintf(s.inv.stderr, "Running: %s\n", strings.Join(containerCmd, " "))
		run := exec.Command(containerCmd[0], containerCmd[1:]...)
		run.Stdout = s.inv.stderr
		run.Stderr = s.inv.stderr
		if err := run.Run(); err != nil {
			return fmt.Errorf("docker run failed: %w", err)
		}
	}
	fmt.Fprintln(s.inv.stderr, "[3/3] Agent installation complete")
	fmt.Fprintf(s.inv.stderr, "Check status at https://console.cloud.google.com/transfer/on-premises/agent-pools/pool/%s/agents?project=%s\n",
		s.flagXferPool, project)
	return nil
}

func (s *transferCmds) xferBuildAgentDockerCmd(project, credsPath, logsDir string, idx int) []string {
	c := []string{"docker", "run",
		"--ulimit", fmt.Sprintf("memlock=%d", s.flagXferMemlockLimit),
		"--rm", "-d",
	}
	if s.flagXferProxy != "" {
		c = append(c, "--env", "HTTPS_PROXY="+s.flagXferProxy)
	}
	if len(s.flagXferMountDirs) == 0 {
		c = append(c, "-v=/:/transfer_root")
	} else {
		c = append(c, "-v="+logsDir+":/tmp")
		if credsPath != "" {
			c = append(c, "-v="+credsPath+":"+credsPath)
		}
		for _, d := range s.flagXferMountDirs {
			c = append(c, "-v="+d+":"+d)
		}
	}
	c = append(c, xferAgentImage,
		"--agent-pool="+s.flagXferPool,
		"--log-dir="+logsDir,
		"--project-id="+project,
	)
	if credsPath != "" {
		c = append(c, "--creds-file="+credsPath)
	}
	if len(s.flagXferMountDirs) == 0 {
		c = append(c, "--enable-mount-directory")
	}
	if s.flagXferIDPrefix != "" {
		suffix := ""
		if s.flagXferCount > 1 {
			suffix = fmt.Sprintf("%d", idx)
		}
		c = append(c, "--agent-id-prefix="+s.flagXferIDPrefix+suffix)
	}
	if s.flagXferS3Compatible {
		c = append(c, "--enable-s3")
	}
	return c
}

func (s *transferCmds) runXferAgentDelete(cmd *cobra.Command, args []string) error {
	if !s.flagXferAgentAll && s.flagXferAgentID == "" {
		return fmt.Errorf("either --id or --all is required")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found on PATH: %w", err)
	}
	targets := []string{}
	if s.flagXferAgentAll {
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
			fmt.Fprintln(s.inv.stderr, "No local transfer-agent containers found.")
			return nil
		}
	} else {
		targets = append(targets, s.flagXferAgentID)
	}
	for _, id := range targets {
		fmt.Fprintf(s.inv.stderr, "Stopping container %s...\n", id)
		out, err := exec.Command("docker", "stop", id).CombinedOutput()
		if err != nil {
			return fmt.Errorf("stopping container %s: %w: %s", id, err, string(out))
		}
	}
	if s.flagXferUninstall {
		out, err := exec.Command("docker", "rmi", xferAgentImage).CombinedOutput()
		if err != nil {
			fmt.Fprintf(s.inv.stderr, "warning: could not remove image %s: %v: %s\n", xferAgentImage, err, string(out))
		}
	}
	return nil
}

// --- jobs impl ---

func (s *transferCmds) runXferJobCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	job := &storagetransfer.TransferJob{ProjectId: project}
	if s.flagXferConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagXferConfigFile, job); err != nil {
			return err
		}
	}
	if job.TransferSpec == nil {
		job.TransferSpec = &storagetransfer.TransferSpec{}
	}
	if err := s.xferPopulateSpec(job.TransferSpec, args[0], args[1], project); err != nil {
		return err
	}
	if s.flagXferJobDesc != "" {
		job.Description = s.flagXferJobDesc
	}
	if s.flagXferJobName != "" {
		job.Name = xferJobName(s.flagXferJobName)
	}
	if s.flagXferSchedRepeats != "" {
		d, err := time.ParseDuration(strings.Replace(s.flagXferSchedRepeats, "d", "*24h", 1))
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
	if s.flagXferAsync {
		fmt.Fprintf(s.inv.stderr, "Created job: %s\n", created.Name)
	}
	return emitFormattedTo(s.inv.stdout, created, s.flagXferFormat)
}

// xferPopulateSpec fills common source/destination fields into spec based on
// URL-shaped positional arguments (gs://, s3://, http(s)://, posix:///,
// hdfs://). Existing values from --config-file are preserved.
func (s *transferCmds) xferPopulateSpec(spec *storagetransfer.TransferSpec, source, dest, project string) error {
	if err := xferApplyEndpoint(spec, source, true); err != nil {
		return err
	}
	if err := xferApplyEndpoint(spec, dest, false); err != nil {
		return err
	}
	if s.flagXferSrcAgentPool != "" {
		spec.SourceAgentPoolName = xferAgentPoolName(project, s.flagXferSrcAgentPool)
	}
	if s.flagXferDstAgentPool != "" {
		spec.SinkAgentPoolName = xferAgentPoolName(project, s.flagXferDstAgentPool)
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

func (s *transferCmds) runXferJobDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
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
	return emitFormattedTo(s.inv.stdout, got, s.flagXferFormat)
}

func (s *transferCmds) runXferJobDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.TransferJobs.Get(xferJobName(args[0]), project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing transfer job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagXferFormat)
}

func (s *transferCmds) runXferJobList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	// The Storage Transfer API requires a projectId filter (JSON string).
	filter := fmt.Sprintf(`{"projectId":"%s"}`, project)
	if s.flagXferFilter != "" {
		filter = s.flagXferFilter
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
	if s.flagXferFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagXferFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-10s %s\n", "NAME", "STATUS", "DESCRIPTION")
	for _, j := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-10s %s\n", strings.TrimPrefix(j.Name, "transferJobs/"), j.Status, j.Description)
	}
	return nil
}

func (s *transferCmds) runXferJobUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	job := &storagetransfer.TransferJob{ProjectId: project}
	var mask []string
	if s.flagXferConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagXferConfigFile, job); err != nil {
			return err
		}
		mask = nonEmptyJSONFields(job)
	}
	if s.flagXferJobStatus != "" {
		job.Status = strings.ToUpper(s.flagXferJobStatus)
		mask = append(mask, "status")
	}
	if s.flagXferJobDesc != "" {
		job.Description = s.flagXferJobDesc
		mask = append(mask, "description")
	}
	req := &storagetransfer.UpdateTransferJobRequest{
		ProjectId:                  project,
		TransferJob:                job,
		UpdateTransferJobFieldMask: strings.Join(dedupe(mask), ","),
	}
	got, err := svc.TransferJobs.Patch(xferJobName(args[0]), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating transfer job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagXferFormat)
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

func (s *transferCmds) runXferJobRun(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.TransferJobs.Run(xferJobName(args[0]),
		&storagetransfer.RunTransferJobRequest{ProjectId: project}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running transfer job: %w", err)
	}
	if s.flagXferAsync {
		final, err := xferWaitOp(ctx, svc, op.Name)
		if err != nil {
			return err
		}
		return emitFormattedTo(s.inv.stdout, final, s.flagXferFormat)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagXferFormat)
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

func (s *transferCmds) runXferJobMonitor(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
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
	fmt.Fprintf(s.inv.stderr, "Monitoring operation: %s\n", opName)
	for {
		got, err := svc.TransferOperations.Get(opName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if err := emitFormattedTo(s.inv.stdout, got, s.flagXferFormat); err != nil {
			return err
		}
		if got.Done {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

// --- operations impl ---

func (s *transferCmds) runXferOpCancel(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	_, err = svc.TransferOperations.Cancel(xferOpName(args[0]),
		&storagetransfer.CancelOperationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Cancel requested for operation [%s].\n", args[0])
	return nil
}

func (s *transferCmds) runXferOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.TransferOperations.Get(xferOpName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagXferFormat)
}

func (s *transferCmds) runXferOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	filter := fmt.Sprintf(`{"projectId":"%s"}`, project)
	if s.flagXferFilter != "" {
		filter = s.flagXferFilter
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
	if s.flagXferFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagXferFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-60s %-6s\n", "NAME", "DONE")
	for _, op := range all {
		fmt.Fprintf(s.inv.stdout, "%-60s %-6t\n", op.Name, op.Done)
	}
	return nil
}

func (s *transferCmds) runXferOpPause(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	_, err = svc.TransferOperations.Pause(xferOpName(args[0]),
		&storagetransfer.PauseTransferOperationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("pausing operation: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Paused operation [%s].\n", args[0])
	return nil
}

func (s *transferCmds) runXferOpResume(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	_, err = svc.TransferOperations.Resume(xferOpName(args[0]),
		&storagetransfer.ResumeTransferOperationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming operation: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Resumed operation [%s].\n", args[0])
	return nil
}

// --- authorize impl ---

func xferExpectedUserRoles() []string {
	return []string{
		"roles/owner",
		"roles/storagetransfer.admin",
		"roles/storagetransfer.transferAgent",
		"roles/storage.objectAdmin",
		"roles/pubsub.editor",
	}
}

func xferExpectedP4SARoles() []string {
	return []string{
		"roles/storage.admin",
		"roles/storagetransfer.serviceAgent",
	}
}

func (s *transferCmds) runXferAuthorize(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	xfer, err := s.xferService(ctx)
	if err != nil {
		return err
	}
	crm, err := s.inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	accountEmail, isSA, err := s.xferResolveAuthorizeAccount(ctx)
	if err != nil {
		return err
	}
	userMember := xferIamMemberFor(accountEmail, isSA)

	policy, err := crm.Projects.GetIamPolicy("projects/"+project,
		&cloudresourcemanager.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching project IAM policy: %w", err)
	}

	missingUser := xferMissingRoles(policy, userMember, xferExpectedUserRoles())
	fmt.Fprintf(s.inv.stderr, "User %s missing roles: %v\n", accountEmail, missingUser)

	p4sa, err := xfer.GoogleServiceAccounts.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching transfer service account: %w", err)
	}
	p4saMember := xferIamMemberFor(p4sa.AccountEmail, true)
	missingP4SA := xferMissingRoles(policy, p4saMember, xferExpectedP4SARoles())
	fmt.Fprintf(s.inv.stderr, "Transfer service account %s missing roles: %v\n", p4sa.AccountEmail, missingP4SA)

	toAdd := map[string][]string{
		userMember: missingUser,
		p4saMember: missingP4SA,
	}
	total := len(missingUser) + len(missingP4SA)
	if total == 0 {
		fmt.Fprintln(s.inv.stderr, "No missing roles.")
		return nil
	}
	if !s.flagXferAddMissing {
		fmt.Fprintln(s.inv.stderr, "Re-run with --add-missing to grant the missing roles.")
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
	fmt.Fprintln(s.inv.stderr, "Missing roles granted. Permissions can take up to 7 minutes to propagate.")
	return nil
}

func (s *transferCmds) xferResolveAuthorizeAccount(ctx context.Context) (string, bool, error) {
	// If --creds-file supplied, parse client_email/type from JSON body.
	if s.flagXferCredsFile != "" {
		data, err := os.ReadFile(s.flagXferCredsFile)
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
	if s.inv.flags.account != "" {
		return s.inv.flags.account, strings.HasSuffix(s.inv.flags.account, ".gserviceaccount.com"), nil
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
