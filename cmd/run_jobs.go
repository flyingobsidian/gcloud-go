package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	googleapi "google.golang.org/api/googleapi"
	logging "google.golang.org/api/logging/v2"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run jobs (#1050) ---
//
// Backed by run/v2 Projects.Locations.Jobs and .Jobs.Executions.

var runJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage Cloud Run jobs"}
var runJobsExecutionsCmd = &cobra.Command{Use: "executions", Short: "Manage Cloud Run job executions"}

var (
	flagRunJobsRegion     string
	flagRunJobsFormat     string
	flagRunJobsConfigFile string
	flagRunJobsUpdateMask string
	flagRunJobsPageSize   int64
	flagRunJobsShowDel    bool
	flagRunJobsLimit      int64

	flagRunJobsImage        string
	flagRunJobsCommand      []string
	flagRunJobsArgs         []string
	flagRunJobsEnvVars      map[string]string
	flagRunJobsServiceAcct  string
	flagRunJobsMemory       string
	flagRunJobsCPU          string
	flagRunJobsTaskCount    int64
	flagRunJobsMaxRetries   int64
	flagRunJobsParallelism  int64
	flagRunJobsTaskTimeout  string
	flagRunJobsVpcConnector string

	flagRunJobsIamMember   string
	flagRunJobsIamRole     string
	flagRunJobsIamCondExpr string
	flagRunJobsIamCondT    string
	flagRunJobsIamCondD    string
	flagRunJobsIamAllCond  bool

	// executions subgroup
	flagRunJobsExecJob      string
	flagRunJobsExecPageSize int64
	flagRunJobsExecShowDel  bool
)

var (
	runJobsCreateCmd = &cobra.Command{
		Use: "create JOB", Short: "Create a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: runJobsCreate,
	}
	runJobsDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: runJobsDelete,
	}
	runJobsDeployCmd = &cobra.Command{
		Use: "deploy JOB", Short: "Deploy a Cloud Run job (create-or-update)",
		Args: cobra.ExactArgs(1), RunE: runJobsDeploy,
	}
	runJobsDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: runJobsDescribe,
	}
	runJobsExecuteCmd = &cobra.Command{
		Use: "execute JOB", Short: "Execute a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: runJobsExecute,
	}
	runJobsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run jobs",
		Args: cobra.NoArgs, RunE: runJobsList,
	}
	runJobsLogsCmd = &cobra.Command{
		Use: "logs JOB", Short: "Read Cloud Logging entries for a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: runJobsLogs,
	}
	runJobsReplaceCmd = &cobra.Command{
		Use: "replace JOB", Short: "Replace a Cloud Run job from a config file",
		Args: cobra.ExactArgs(1), RunE: runJobsReplace,
	}
	runJobsUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update a Cloud Run job from a config file",
		Args: cobra.ExactArgs(1), RunE: runJobsUpdate,
	}
	runJobsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy JOB", Short: "Get the IAM policy for a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: runJobsGetIam,
	}
	runJobsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy JOB POLICY_FILE", Short: "Set the IAM policy for a Cloud Run job",
		Args: cobra.ExactArgs(2), RunE: runJobsSetIam,
	}
	runJobsAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding JOB", Short: "Add an IAM policy binding to a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: runJobsAddIam,
	}
	runJobsRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding JOB", Short: "Remove an IAM policy binding from a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: runJobsRemoveIam,
	}

	runJobsExecDescribeCmd = &cobra.Command{
		Use: "describe EXECUTION", Short: "Describe a Cloud Run job execution",
		Args: cobra.ExactArgs(1), RunE: runJobsExecDescribe,
	}
	runJobsExecListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run job executions",
		Args: cobra.NoArgs, RunE: runJobsExecList,
	}
	runJobsExecDeleteCmd = &cobra.Command{
		Use: "delete EXECUTION", Short: "Delete a Cloud Run job execution",
		Args: cobra.ExactArgs(1), RunE: runJobsExecDelete,
	}
	runJobsExecCancelCmd = &cobra.Command{
		Use: "cancel EXECUTION", Short: "Cancel a running Cloud Run job execution",
		Args: cobra.ExactArgs(1), RunE: runJobsExecCancel,
	}
)

func init() {
	jobsAll := []*cobra.Command{
		runJobsCreateCmd, runJobsDeleteCmd, runJobsDeployCmd, runJobsDescribeCmd,
		runJobsExecuteCmd, runJobsListCmd, runJobsLogsCmd, runJobsReplaceCmd,
		runJobsUpdateCmd, runJobsGetIamCmd, runJobsSetIamCmd, runJobsAddIamCmd,
		runJobsRemoveIamCmd,
	}
	for _, c := range jobsAll {
		c.Flags().StringVar(&flagRunJobsRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&flagRunJobsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{runJobsCreateCmd, runJobsReplaceCmd, runJobsUpdateCmd} {
		c.Flags().StringVar(&flagRunJobsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Job body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	runJobsUpdateCmd.Flags().StringVar(&flagRunJobsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; ignored (Cloud Run v2 merges patches automatically)")

	runJobsListCmd.Flags().Int64Var(&flagRunJobsPageSize, "page-size", 0, "Maximum results per page")
	runJobsListCmd.Flags().BoolVar(&flagRunJobsShowDel, "show-deleted", false, "Include deleted jobs")
	runJobsLogsCmd.Flags().Int64Var(&flagRunJobsLimit, "limit", 100, "Maximum number of log entries to return")

	// Deploy flags
	runJobsDeployCmd.Flags().StringVar(&flagRunJobsImage, "image", "", "Container image (required)")
	_ = runJobsDeployCmd.MarkFlagRequired("image")
	runJobsDeployCmd.Flags().StringSliceVar(&flagRunJobsCommand, "command", nil, "Container entrypoint override")
	runJobsDeployCmd.Flags().StringSliceVar(&flagRunJobsArgs, "args", nil, "Container arguments override")
	runJobsDeployCmd.Flags().StringToStringVar(&flagRunJobsEnvVars, "env-vars", nil, "Container environment variables (KEY=VALUE)")
	runJobsDeployCmd.Flags().StringVar(&flagRunJobsServiceAcct, "service-account", "", "Service account for the job")
	runJobsDeployCmd.Flags().StringVar(&flagRunJobsMemory, "memory", "", "Memory limit (e.g. 512Mi)")
	runJobsDeployCmd.Flags().StringVar(&flagRunJobsCPU, "cpu", "", "CPU limit (e.g. 1, 2, 4)")
	runJobsDeployCmd.Flags().Int64Var(&flagRunJobsTaskCount, "task-count", 0, "Number of tasks per execution")
	runJobsDeployCmd.Flags().Int64Var(&flagRunJobsMaxRetries, "max-retries", 0, "Maximum task retries")
	runJobsDeployCmd.Flags().Int64Var(&flagRunJobsParallelism, "parallelism", 0, "Maximum parallel tasks")
	runJobsDeployCmd.Flags().StringVar(&flagRunJobsTaskTimeout, "task-timeout", "", "Per-task timeout (e.g. 600s)")
	runJobsDeployCmd.Flags().StringVar(&flagRunJobsVpcConnector, "vpc-connector", "", "VPC connector to use")

	for _, c := range []*cobra.Command{runJobsAddIamCmd, runJobsRemoveIamCmd} {
		runIamFlags(c, &flagRunJobsIamMember, &flagRunJobsIamRole,
			&flagRunJobsIamCondExpr, &flagRunJobsIamCondT, &flagRunJobsIamCondD)
	}
	runJobsRemoveIamCmd.Flags().BoolVar(&flagRunJobsIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	// executions subgroup
	execAll := []*cobra.Command{
		runJobsExecDescribeCmd, runJobsExecListCmd, runJobsExecDeleteCmd, runJobsExecCancelCmd,
	}
	for _, c := range execAll {
		c.Flags().StringVar(&flagRunJobsRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&flagRunJobsExecJob, "job", "", "Parent Cloud Run job (required)")
		c.Flags().StringVar(&flagRunJobsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("job")
	}
	runJobsExecListCmd.Flags().Int64Var(&flagRunJobsExecPageSize, "page-size", 0, "Maximum results per page")
	runJobsExecListCmd.Flags().BoolVar(&flagRunJobsExecShowDel, "show-deleted", false, "Include deleted executions")

	runJobsExecutionsCmd.AddCommand(execAll...)
	runJobsCmd.AddCommand(jobsAll...)
	runJobsCmd.AddCommand(runJobsExecutionsCmd)
	runCmd.AddCommand(runJobsCmd)
}

func runJobsName(project, job string) string {
	return runResourceName(project, flagRunJobsRegion, "jobs", job)
}

func runJobsExecName(project, job, exec string) string {
	if hasProjectsPrefix(exec) {
		return exec
	}
	return fmt.Sprintf("%s/executions/%s", runJobsName(project, job), exec)
}

func runJobsCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Job{}
	if err := loadYAMLOrJSONInto(flagRunJobsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Create(runParent(project, flagRunJobsRegion), body).
		JobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	fmt.Printf("Create request issued for job [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagRunJobsFormat)
}

func runJobsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Delete(runJobsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Printf("Delete request issued for job [%s].\n", args[0])
	return emitFormatted(op, flagRunJobsFormat)
}

func runJobsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Get(runJobsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormatted(got, flagRunJobsFormat)
}

func runJobsExecute(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Run(runJobsName(project, args[0]),
		&runv2.GoogleCloudRunV2RunJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("executing job: %w", err)
	}
	fmt.Printf("Execute request issued for job [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagRunJobsFormat)
}

func runJobsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	var all []*runv2.GoogleCloudRunV2Job
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.List(runParent(project, flagRunJobsRegion)).Context(ctx)
		if flagRunJobsPageSize > 0 {
			call = call.PageSize(flagRunJobsPageSize)
		}
		if flagRunJobsShowDel {
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
	return emitFormatted(all, flagRunJobsFormat)
}

func runJobsLogs(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LoggingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	filter := fmt.Sprintf(
		`resource.type="cloud_run_job" AND resource.labels.job_name=%q AND resource.labels.location=%q`,
		args[0], flagRunJobsRegion,
	)
	limit := flagRunJobsLimit
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
	if flagRunJobsFormat != "" {
		return emitFormatted(resp.Entries, flagRunJobsFormat)
	}
	printLogEntries(resp.Entries)
	return nil
}

func runJobsReplace(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Job{}
	if err := loadYAMLOrJSONInto(flagRunJobsConfigFile, body); err != nil {
		return err
	}
	body.Name = runJobsName(project, args[0])
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Patch(body.Name, body).
		AllowMissing(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing job: %w", err)
	}
	fmt.Printf("Replace request issued for job [%s].\n", args[0])
	return emitFormatted(op, flagRunJobsFormat)
}

func runJobsUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Job{}
	if err := loadYAMLOrJSONInto(flagRunJobsConfigFile, body); err != nil {
		return err
	}
	body.Name = runJobsName(project, args[0])
	// The Jobs.Patch RPC in run/v2 (v0.279.0) does not accept an
	// updateMask parameter; the server derives it from the populated
	// fields on the request body. --update-mask is accepted for
	// parity with `gcloud run services update` but not forwarded.
	_ = flagRunJobsUpdateMask
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Patch(body.Name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	fmt.Printf("Update request issued for job [%s].\n", args[0])
	return emitFormatted(op, flagRunJobsFormat)
}

// runJobsDeploy performs a conditional Get -> Create/Patch flow, mirroring
// the `gcloud run jobs deploy` command in the Python CLI.
func runJobsDeploy(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	name := runJobsName(project, args[0])
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
	container.Image = flagRunJobsImage
	if flagRunJobsCommand != nil {
		container.Command = flagRunJobsCommand
	}
	if flagRunJobsArgs != nil {
		container.Args = flagRunJobsArgs
	}
	if len(flagRunJobsEnvVars) > 0 {
		container.Env = envVarsFromMap(flagRunJobsEnvVars)
	}
	applyResourceLimits(&container.Resources, flagRunJobsMemory, flagRunJobsCPU)
	if flagRunJobsServiceAcct != "" {
		tt.ServiceAccount = flagRunJobsServiceAcct
	}
	if flagRunJobsMaxRetries > 0 {
		tt.MaxRetries = flagRunJobsMaxRetries
	}
	if flagRunJobsTaskTimeout != "" {
		tt.Timeout = flagRunJobsTaskTimeout
	}
	if flagRunJobsVpcConnector != "" {
		if tt.VpcAccess == nil {
			tt.VpcAccess = &runv2.GoogleCloudRunV2VpcAccess{}
		}
		tt.VpcAccess.Connector = flagRunJobsVpcConnector
	}
	if flagRunJobsTaskCount > 0 {
		body.Template.TaskCount = flagRunJobsTaskCount
	}
	if flagRunJobsParallelism > 0 {
		body.Template.Parallelism = flagRunJobsParallelism
	}

	if existing == nil {
		op, err := svc.Projects.Locations.Jobs.Create(runParent(project, flagRunJobsRegion), body).
			JobId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating job: %w", err)
		}
		fmt.Printf("Created job [%s] (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagRunJobsFormat)
	}
	op, err := svc.Projects.Locations.Jobs.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	fmt.Printf("Updated job [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagRunJobsFormat)
}

func runJobsGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Jobs.GetIamPolicy(runJobsName(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagRunJobsFormat)
}

func runJobsSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &runv2.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Jobs.SetIamPolicy(runJobsName(project, args[0]),
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("job [%s]", args[0]))
	return emitFormatted(updated, flagRunJobsFormat)
}

func runJobsAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	resource := runJobsName(project, args[0])
	policy, err := svc.Projects.Locations.Jobs.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	runIamAddBinding(policy, flagRunJobsIamRole, flagRunJobsIamMember,
		runIamBuildCondition(flagRunJobsIamCondExpr, flagRunJobsIamCondT, flagRunJobsIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Jobs.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("job [%s]", args[0]))
	return emitFormatted(updated, flagRunJobsFormat)
}

func runJobsRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	resource := runJobsName(project, args[0])
	policy, err := svc.Projects.Locations.Jobs.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !runIamRemoveBinding(policy, flagRunJobsIamRole, flagRunJobsIamMember,
		runIamBuildCondition(flagRunJobsIamCondExpr, flagRunJobsIamCondT, flagRunJobsIamCondD),
		flagRunJobsIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			flagRunJobsIamRole, flagRunJobsIamMember)
	}
	updated, err := svc.Projects.Locations.Jobs.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("job [%s]", args[0]))
	return emitFormatted(updated, flagRunJobsFormat)
}

func runJobsExecDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Executions.Get(
		runJobsExecName(project, flagRunJobsExecJob, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing execution: %w", err)
	}
	return emitFormatted(got, flagRunJobsFormat)
}

func runJobsExecList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	parent := runJobsName(project, flagRunJobsExecJob)
	var all []*runv2.GoogleCloudRunV2Execution
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.Executions.List(parent).Context(ctx)
		if flagRunJobsExecPageSize > 0 {
			call = call.PageSize(flagRunJobsExecPageSize)
		}
		if flagRunJobsExecShowDel {
			call = call.ShowDeleted(true)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing executions: %w", err)
		}
		all = append(all, resp.Executions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagRunJobsFormat)
}

func runJobsExecDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Executions.Delete(
		runJobsExecName(project, flagRunJobsExecJob, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting execution: %w", err)
	}
	fmt.Printf("Delete request issued for execution [%s].\n", args[0])
	return emitFormatted(op, flagRunJobsFormat)
}

func runJobsExecCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Executions.Cancel(
		runJobsExecName(project, flagRunJobsExecJob, args[0]),
		&runv2.GoogleCloudRunV2CancelExecutionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling execution: %w", err)
	}
	fmt.Printf("Cancel request issued for execution [%s].\n", args[0])
	return emitFormatted(op, flagRunJobsFormat)
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
