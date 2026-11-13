package cli

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/spf13/cobra"
	googleapi "google.golang.org/api/googleapi"
	logging "google.golang.org/api/logging/v2"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run jobs (#1050) ---
//
// Backed by run/v2 Projects.Locations.Jobs and .Jobs.Executions.

func (s *runCmds) initRunJobs() {
	jobsAll := []*cobra.Command{
		s.runJobsCreateCmd, s.runJobsDeleteCmd, s.runJobsDeployCmd, s.runJobsDescribeCmd,
		s.runJobsExecuteCmd, s.runJobsListCmd, s.runJobsLogsCmd, s.runJobsReplaceCmd,
		s.runJobsUpdateCmd, s.runJobsGetIamCmd, s.runJobsSetIamCmd, s.runJobsAddIamCmd,
		s.runJobsRemoveIamCmd,
	}
	for _, c := range jobsAll {
		c.Flags().StringVar(&s.flagRunJobsFormat, "format", "", "Output format")
		if c == s.runJobsListCmd {
			// As in gcloud, list spans every region unless --region is given.
			c.Flags().StringVar(&s.flagRunJobsRegion, "region", "", "Cloud Run region; default all regions")
			continue
		}
		c.Flags().StringVar(&s.flagRunJobsRegion, "region", "", "Cloud Run region (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.runJobsCreateCmd, s.runJobsReplaceCmd, s.runJobsUpdateCmd} {
		c.Flags().StringVar(&s.flagRunJobsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Job body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.runJobsUpdateCmd.Flags().StringVar(&s.flagRunJobsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; ignored (Cloud Run v2 merges patches automatically)")

	s.runJobsListCmd.Flags().Int64Var(&s.flagRunJobsPageSize, "page-size", 0, "Maximum results per page")
	s.runJobsListCmd.Flags().BoolVar(&s.flagRunJobsShowDel, "show-deleted", false, "Include deleted jobs")
	s.runJobsLogsCmd.Flags().Int64Var(&s.flagRunJobsLimit, "limit", 100, "Maximum number of log entries to return")
	s.runJobsExecuteCmd.Flags().BoolVar(&s.flagRunJobsWait, "wait", false,
		"Wait until the execution has completed running before exiting")

	// Deploy flags
	s.runJobsDeployCmd.Flags().StringVar(&s.flagRunJobsImage, "image", "", "Container image (required)")
	_ = s.runJobsDeployCmd.MarkFlagRequired("image")
	s.runJobsDeployCmd.Flags().StringSliceVar(&s.flagRunJobsCommand, "command", nil, "Container entrypoint override")
	s.runJobsDeployCmd.Flags().StringSliceVar(&s.flagRunJobsArgs, "args", nil, "Container arguments override")
	s.runJobsDeployCmd.Flags().StringToStringVar(&s.flagRunJobsEnvVars, "env-vars", nil, "Container environment variables (KEY=VALUE)")
	s.runJobsDeployCmd.Flags().StringVar(&s.flagRunJobsServiceAcct, "service-account", "", "Service account for the job")
	s.runJobsDeployCmd.Flags().StringVar(&s.flagRunJobsMemory, "memory", "", "Memory limit (e.g. 512Mi)")
	s.runJobsDeployCmd.Flags().StringVar(&s.flagRunJobsCPU, "cpu", "", "CPU limit (e.g. 1, 2, 4)")
	s.runJobsDeployCmd.Flags().Int64Var(&s.flagRunJobsTaskCount, "task-count", 0, "Number of tasks per execution")
	s.runJobsDeployCmd.Flags().Int64Var(&s.flagRunJobsMaxRetries, "max-retries", 0, "Maximum task retries")
	s.runJobsDeployCmd.Flags().Int64Var(&s.flagRunJobsParallelism, "parallelism", 0, "Maximum parallel tasks")
	s.runJobsDeployCmd.Flags().StringVar(&s.flagRunJobsTaskTimeout, "task-timeout", "", "Per-task timeout (e.g. 600s)")
	s.runJobsDeployCmd.Flags().StringVar(&s.flagRunJobsVpcConnector, "vpc-connector", "", "VPC connector to use")

	for _, c := range []*cobra.Command{s.runJobsAddIamCmd, s.runJobsRemoveIamCmd} {
		runIamFlags(c, &s.flagRunJobsIamMember, &s.flagRunJobsIamRole,
			&s.flagRunJobsIamCondExpr, &s.flagRunJobsIamCondT, &s.flagRunJobsIamCondD)
	}
	s.runJobsRemoveIamCmd.Flags().BoolVar(&s.flagRunJobsIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	// executions subgroup
	execAll := []*cobra.Command{
		s.runJobsExecDescribeCmd, s.runJobsExecListCmd, s.runJobsExecDeleteCmd, s.runJobsExecCancelCmd,
	}
	for _, c := range execAll {
		c.Flags().StringVar(&s.flagRunJobsRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&s.flagRunJobsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
		if c == s.runJobsExecListCmd {
			// As in gcloud, list spans every job unless --job is given.
			c.Flags().StringVar(&s.flagRunJobsExecJob, "job", "", "Limit matched executions to the given job")
			continue
		}
		c.Flags().StringVar(&s.flagRunJobsExecJob, "job", "", "Parent Cloud Run job (required)")
		_ = c.MarkFlagRequired("job")
	}
	s.runJobsExecListCmd.Flags().Int64Var(&s.flagRunJobsExecPageSize, "page-size", 0, "Maximum results per page (default 100)")
	s.runJobsExecListCmd.Flags().Int64Var(&s.flagRunJobsExecLimit, "limit", 0, "Maximum number of executions to return, newest first (0 = no limit)")

	s.runJobsExecutionsCmd.AddCommand(execAll...)
	s.runJobsCmd.AddCommand(jobsAll...)
	s.runJobsCmd.AddCommand(s.runJobsExecutionsCmd)
	s.runCmd.AddCommand(s.runJobsCmd)
}

func (s *runCmds) runJobsName(project, job string) string {
	return runResourceName(project, s.flagRunJobsRegion, "jobs", job)
}

func (s *runCmds) runJobsExecName(project, job, exec string) string {
	if hasProjectsPrefix(exec) {
		return exec
	}
	return fmt.Sprintf("%s/executions/%s", s.runJobsName(project, job), exec)
}

func (s *runCmds) runJobsCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Job{}
	if err := loadYAMLOrJSONInto(s.flagRunJobsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Create(runParent(project, s.flagRunJobsRegion), body).
		JobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for job [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Delete(s.runJobsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for job [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Get(s.runJobsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsExecute(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Run(s.runJobsName(project, args[0]),
		&runv2.GoogleCloudRunV2RunJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("executing job: %w", err)
	}
	if !s.flagRunJobsWait {
		fmt.Fprintf(s.inv.stdout, "Execute request issued for job [%s] (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
	}
	created, err := runJobsOpExecution(op)
	if err != nil {
		return err
	}
	get := func(ctx context.Context) (*runv2.GoogleCloudRunV2Execution, error) {
		return svc.Projects.Locations.Jobs.Executions.Get(created.Name).Context(ctx).Do()
	}
	exec, err := runJobsWaitForExecution(ctx, s.inv.stderr, get, runJobsWaitInterval)
	if errors.Is(err, errRunExecutionFailed) {
		return fmt.Errorf("%w.%s", err, runJobsExecDetailsMessage(project, s.flagRunJobsRegion, exec))
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "Execution [%s] has successfully completed.\n", path.Base(exec.Name))
	fmt.Fprint(s.inv.stderr, runJobsExecDetailsMessage(project, s.flagRunJobsRegion, exec))
	if s.flagRunJobsFormat == "" {
		return nil
	}
	return emitFormattedTo(s.inv.stdout, exec, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if s.flagRunJobsRegion == "" {
		v1, err := s.inv.clients().RunV1(ctx, "")
		if err != nil {
			return err
		}
		all, err := runListAllJobs(ctx, s.inv.stderr, v1, project)
		if err != nil {
			return err
		}
		return emitFormattedTo(s.inv.stdout, all, s.flagRunJobsFormat)
	}
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	var all []*runv2.GoogleCloudRunV2Job
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.List(runParent(project, s.flagRunJobsRegion)).Context(ctx)
		if s.flagRunJobsPageSize > 0 {
			call = call.PageSize(s.flagRunJobsPageSize)
		}
		if s.flagRunJobsShowDel {
			call = call.ShowDeleted(true)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing jobs: %w", err)
		}
		all = append(all, resp.Jobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsLogs(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Logging(ctx)
	if err != nil {
		return err
	}
	filter := fmt.Sprintf(
		`resource.type="cloud_run_job" AND resource.labels.job_name=%q AND resource.labels.location=%q`,
		args[0], s.flagRunJobsRegion,
	)
	limit := s.flagRunJobsLimit
	if limit <= 0 {
		limit = 100
	}
	req := &logging.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + project},
		Filter:        filter,
		OrderBy:       "timestamp desc",
		PageSize:      limit,
	}
	resp, err := svc.Entries.List(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing log entries: %w", err)
	}
	if s.flagRunJobsFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Entries, s.flagRunJobsFormat)
	}
	s.printLogEntries(resp.Entries)
	return nil
}

func (s *runCmds) runJobsReplace(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Job{}
	if err := loadYAMLOrJSONInto(s.flagRunJobsConfigFile, body); err != nil {
		return err
	}
	body.Name = s.runJobsName(project, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Patch(body.Name, body).
		AllowMissing(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Replace request issued for job [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Job{}
	if err := loadYAMLOrJSONInto(s.flagRunJobsConfigFile, body); err != nil {
		return err
	}
	body.Name = s.runJobsName(project, args[0])
	// The Jobs.Patch RPC in run/v2 (v0.279.0) does not accept an
	// updateMask parameter; the server derives it from the populated
	// fields on the request body. --update-mask is accepted for
	// parity with `gcloud run services update` but not forwarded.
	_ = s.flagRunJobsUpdateMask
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Patch(body.Name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for job [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
}

// runJobsDeploy performs a conditional Get -> Create/Patch flow, mirroring
// the `gcloud run jobs deploy` command in the Python CLI.
func (s *runCmds) runJobsDeploy(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	name := s.runJobsName(project, args[0])
	existing, err := svc.Projects.Locations.Jobs.Get(name).Context(ctx).Do()
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("checking existing job: %w", err)
	}

	body := &runv2.GoogleCloudRunV2Job{}
	if existing != nil {
		// Preserve the existing template shape so we only mutate the
		// fields the caller specified.
		body = existing
		body.Name = name
	}
	if body.Template == nil {
		body.Template = &runv2.GoogleCloudRunV2ExecutionTemplate{}
	}
	if body.Template.Template == nil {
		body.Template.Template = &runv2.GoogleCloudRunV2TaskTemplate{}
	}
	tt := body.Template.Template
	var container *runv2.GoogleCloudRunV2Container
	if len(tt.Containers) > 0 {
		container = tt.Containers[0]
	} else {
		container = &runv2.GoogleCloudRunV2Container{}
		tt.Containers = []*runv2.GoogleCloudRunV2Container{container}
	}
	container.Image = s.flagRunJobsImage
	if s.flagRunJobsCommand != nil {
		container.Command = s.flagRunJobsCommand
	}
	if s.flagRunJobsArgs != nil {
		container.Args = s.flagRunJobsArgs
	}
	if len(s.flagRunJobsEnvVars) > 0 {
		container.Env = envVarsFromMap(s.flagRunJobsEnvVars)
	}
	applyResourceLimits(&container.Resources, s.flagRunJobsMemory, s.flagRunJobsCPU)
	if s.flagRunJobsServiceAcct != "" {
		tt.ServiceAccount = s.flagRunJobsServiceAcct
	}
	if s.flagRunJobsMaxRetries > 0 {
		tt.MaxRetries = s.flagRunJobsMaxRetries
	}
	if s.flagRunJobsTaskTimeout != "" {
		tt.Timeout = s.flagRunJobsTaskTimeout
	}
	if s.flagRunJobsVpcConnector != "" {
		if tt.VpcAccess == nil {
			tt.VpcAccess = &runv2.GoogleCloudRunV2VpcAccess{}
		}
		tt.VpcAccess.Connector = s.flagRunJobsVpcConnector
	}
	if s.flagRunJobsTaskCount > 0 {
		body.Template.TaskCount = s.flagRunJobsTaskCount
	}
	if s.flagRunJobsParallelism > 0 {
		body.Template.Parallelism = s.flagRunJobsParallelism
	}

	if existing == nil {
		op, err := svc.Projects.Locations.Jobs.Create(runParent(project, s.flagRunJobsRegion), body).
			JobId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating job: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Created job [%s] (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
	}
	op, err := svc.Projects.Locations.Jobs.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated job [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Jobs.GetIamPolicy(s.runJobsName(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &runv2.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Jobs.SetIamPolicy(s.runJobsName(project, args[0]),
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("job [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	resource := s.runJobsName(project, args[0])
	policy, err := svc.Projects.Locations.Jobs.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	runIamAddBinding(policy, s.flagRunJobsIamRole, s.flagRunJobsIamMember,
		runIamBuildCondition(s.flagRunJobsIamCondExpr, s.flagRunJobsIamCondT, s.flagRunJobsIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Jobs.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("job [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	resource := s.runJobsName(project, args[0])
	policy, err := svc.Projects.Locations.Jobs.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !runIamRemoveBinding(policy, s.flagRunJobsIamRole, s.flagRunJobsIamMember,
		runIamBuildCondition(s.flagRunJobsIamCondExpr, s.flagRunJobsIamCondT, s.flagRunJobsIamCondD),
		s.flagRunJobsIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			s.flagRunJobsIamRole, s.flagRunJobsIamMember)
	}
	updated, err := svc.Projects.Locations.Jobs.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("job [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsExecDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Executions.Get(
		s.runJobsExecName(project, s.flagRunJobsExecJob, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing execution: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsExecDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Executions.Delete(
		s.runJobsExecName(project, s.flagRunJobsExecJob, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting execution: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for execution [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
}

func (s *runCmds) runJobsExecCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Executions.Cancel(
		s.runJobsExecName(project, s.flagRunJobsExecJob, args[0]),
		&runv2.GoogleCloudRunV2CancelExecutionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling execution: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for execution [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunJobsFormat)
}

// --- Shared helpers reused by deploy/services/worker-pools. ---

// envVarsFromMap converts a StringToString flag value to a v2 EnvVar list.
// StringToString ordering isn't guaranteed but Cloud Run doesn't rely on it.
func envVarsFromMap(m map[string]string) []*runv2.GoogleCloudRunV2EnvVar {
	if len(m) == 0 {
		return nil
	}
	out := make([]*runv2.GoogleCloudRunV2EnvVar, 0, len(m))
	for k, v := range m {
		out = append(out, &runv2.GoogleCloudRunV2EnvVar{Name: k, Value: v})
	}
	return out
}

// applyResourceLimits sets memory/cpu on the container's Resources block,
// allocating the block on demand and preserving any existing limits.
func applyResourceLimits(res **runv2.GoogleCloudRunV2ResourceRequirements, memory, cpu string) {
	if memory == "" && cpu == "" {
		return
	}
	if *res == nil {
		*res = &runv2.GoogleCloudRunV2ResourceRequirements{}
	}
	if (*res).Limits == nil {
		(*res).Limits = map[string]string{}
	}
	if memory != "" {
		(*res).Limits["memory"] = memory
	}
	if cpu != "" {
		(*res).Limits["cpu"] = cpu
	}
}

// isNotFound reports whether err is a Google API HTTP 404.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && apiErr.Code == 404 {
		return true
	}
	return false
}
