package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run worker-pools (#1055) ---
//
// Backed by run/v2 Projects.Locations.WorkerPools and .Revisions.

var runWorkerPoolsCmd = &cobra.Command{Use: "worker-pools", Short: "Manage Cloud Run worker pools"}
var runWorkerPoolsRevisionsCmd = &cobra.Command{Use: "revisions", Short: "Manage Cloud Run worker pool revisions"}

var (
	flagRunWPRegion     string
	flagRunWPFormat     string
	flagRunWPConfigFile string
	flagRunWPUpdateMask string
	flagRunWPPageSize   int64
	flagRunWPShowDel    bool
	flagRunWPLimit      int64

	flagRunWPImage        string
	flagRunWPCommand      []string
	flagRunWPArgs         []string
	flagRunWPEnvVars      map[string]string
	flagRunWPServiceAcct  string
	flagRunWPMemory       string
	flagRunWPCPU          string
	flagRunWPVpcConnector string

	flagRunWPIamMember   string
	flagRunWPIamRole     string
	flagRunWPIamCondExpr string
	flagRunWPIamCondT    string
	flagRunWPIamCondD    string
	flagRunWPIamAllCond  bool

	flagRunWPSplitRevs   map[string]string
	flagRunWPSplitLatest bool

	// revisions subgroup
	flagRunWPRevPool     string
	flagRunWPRevPageSize int64
	flagRunWPRevShowDel  bool
)

var (
	runWPDeleteCmd = &cobra.Command{
		Use: "delete POOL", Short: "Delete a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: runWPDelete,
	}
	runWPDeployCmd = &cobra.Command{
		Use: "deploy POOL", Short: "Deploy a Cloud Run worker pool (create-or-update)",
		Args: cobra.ExactArgs(1), RunE: runWPDeploy,
	}
	runWPDescribeCmd = &cobra.Command{
		Use: "describe POOL", Short: "Describe a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: runWPDescribe,
	}
	runWPListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run worker pools",
		Args: cobra.NoArgs, RunE: runWPList,
	}
	runWPLogsCmd = &cobra.Command{
		Use: "logs POOL", Short: "Read Cloud Logging entries for a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: runWPLogs,
	}
	runWPReplaceCmd = &cobra.Command{
		Use: "replace POOL", Short: "Replace a Cloud Run worker pool from a config file",
		Args: cobra.ExactArgs(1), RunE: runWPReplace,
	}
	runWPUpdateCmd = &cobra.Command{
		Use: "update POOL", Short: "Update a Cloud Run worker pool from a config file",
		Args: cobra.ExactArgs(1), RunE: runWPUpdate,
	}
	runWPUpdateSplitCmd = &cobra.Command{
		Use: "update-instance-split POOL", Short: "Update the instance split for a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: runWPUpdateSplit,
	}
	runWPGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POOL", Short: "Get the IAM policy for a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: runWPGetIam,
	}
	runWPSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POOL POLICY_FILE", Short: "Set the IAM policy for a Cloud Run worker pool",
		Args: cobra.ExactArgs(2), RunE: runWPSetIam,
	}
	runWPAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding POOL", Short: "Add an IAM policy binding to a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: runWPAddIam,
	}
	runWPRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding POOL", Short: "Remove an IAM policy binding from a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: runWPRemoveIam,
	}

	runWPRevDescribeCmd = &cobra.Command{
		Use: "describe REVISION", Short: "Describe a Cloud Run worker pool revision",
		Args: cobra.ExactArgs(1), RunE: runWPRevDescribe,
	}
	runWPRevListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run worker pool revisions",
		Args: cobra.NoArgs, RunE: runWPRevList,
	}
	runWPRevDeleteCmd = &cobra.Command{
		Use: "delete REVISION", Short: "Delete a Cloud Run worker pool revision",
		Args: cobra.ExactArgs(1), RunE: runWPRevDelete,
	}
)

func init() {
	poolAll := []*cobra.Command{
		runWPDeleteCmd, runWPDeployCmd, runWPDescribeCmd, runWPListCmd, runWPLogsCmd,
		runWPReplaceCmd, runWPUpdateCmd, runWPUpdateSplitCmd, runWPGetIamCmd,
		runWPSetIamCmd, runWPAddIamCmd, runWPRemoveIamCmd,
	}
	for _, c := range poolAll {
		c.Flags().StringVar(&flagRunWPRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&flagRunWPFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{runWPReplaceCmd, runWPUpdateCmd} {
		c.Flags().StringVar(&flagRunWPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the WorkerPool body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	runWPUpdateCmd.Flags().StringVar(&flagRunWPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")

	runWPListCmd.Flags().Int64Var(&flagRunWPPageSize, "page-size", 0, "Maximum results per page")
	runWPListCmd.Flags().BoolVar(&flagRunWPShowDel, "show-deleted", false, "Include deleted worker pools")
	runWPLogsCmd.Flags().Int64Var(&flagRunWPLimit, "limit", 100, "Maximum number of log entries to return")

	runWPDeployCmd.Flags().StringVar(&flagRunWPImage, "image", "", "Container image (required)")
	_ = runWPDeployCmd.MarkFlagRequired("image")
	runWPDeployCmd.Flags().StringSliceVar(&flagRunWPCommand, "command", nil, "Container entrypoint override")
	runWPDeployCmd.Flags().StringSliceVar(&flagRunWPArgs, "args", nil, "Container arguments override")
	runWPDeployCmd.Flags().StringToStringVar(&flagRunWPEnvVars, "env-vars", nil, "Container environment variables (KEY=VALUE)")
	runWPDeployCmd.Flags().StringVar(&flagRunWPServiceAcct, "service-account", "", "Service account for the worker pool")
	runWPDeployCmd.Flags().StringVar(&flagRunWPMemory, "memory", "", "Memory limit (e.g. 512Mi)")
	runWPDeployCmd.Flags().StringVar(&flagRunWPCPU, "cpu", "", "CPU limit (e.g. 1, 2, 4)")
	runWPDeployCmd.Flags().StringVar(&flagRunWPVpcConnector, "vpc-connector", "", "VPC connector to use")

	runWPUpdateSplitCmd.Flags().StringToStringVar(&flagRunWPSplitRevs, "to-revisions", nil,
		"Revision-to-percent map, e.g. REV1=60,REV2=40")
	runWPUpdateSplitCmd.Flags().BoolVar(&flagRunWPSplitLatest, "to-latest", false,
		"Route 100% of instances to the latest ready revision")

	for _, c := range []*cobra.Command{runWPAddIamCmd, runWPRemoveIamCmd} {
		runIamFlags(c, &flagRunWPIamMember, &flagRunWPIamRole,
			&flagRunWPIamCondExpr, &flagRunWPIamCondT, &flagRunWPIamCondD)
	}
	runWPRemoveIamCmd.Flags().BoolVar(&flagRunWPIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	// Revisions subgroup
	revAll := []*cobra.Command{runWPRevDescribeCmd, runWPRevListCmd, runWPRevDeleteCmd}
	for _, c := range revAll {
		c.Flags().StringVar(&flagRunWPRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&flagRunWPRevPool, "worker-pool", "", "Parent Cloud Run worker pool (required)")
		c.Flags().StringVar(&flagRunWPFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("worker-pool")
	}
	runWPRevListCmd.Flags().Int64Var(&flagRunWPRevPageSize, "page-size", 0, "Maximum results per page")
	runWPRevListCmd.Flags().BoolVar(&flagRunWPRevShowDel, "show-deleted", false, "Include deleted revisions")

	runWorkerPoolsRevisionsCmd.AddCommand(revAll...)
	runWorkerPoolsCmd.AddCommand(poolAll...)
	runWorkerPoolsCmd.AddCommand(runWorkerPoolsRevisionsCmd)
	runCmd.AddCommand(runWorkerPoolsCmd)
}

func runWPName(project, pool string) string {
	return runResourceName(project, flagRunWPRegion, "workerPools", pool)
}

func runWPRevName(project, pool, rev string) string {
	if hasProjectsPrefix(rev) {
		return rev
	}
	return fmt.Sprintf("%s/revisions/%s", runWPName(project, pool), rev)
}

func runWPDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Delete(runWPName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting worker pool: %w", err)
	}
	fmt.Printf("Delete request issued for worker pool [%s].\n", args[0])
	return emitFormatted(op, flagRunWPFormat)
}

func runWPDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkerPools.Get(runWPName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing worker pool: %w", err)
	}
	return emitFormatted(got, flagRunWPFormat)
}

func runWPList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	var all []*runv2.GoogleCloudRunV2WorkerPool
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkerPools.List(runParent(project, flagRunWPRegion)).Context(ctx)
		if flagRunWPPageSize > 0 {
			call = call.PageSize(flagRunWPPageSize)
		}
		if flagRunWPShowDel {
			call = call.ShowDeleted(true)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing worker pools: %w", err)
		}
		all = append(all, resp.WorkerPools...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagRunWPFormat)
}

func runWPReplace(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2WorkerPool{}
	if err := loadYAMLOrJSONInto(flagRunWPConfigFile, body); err != nil {
		return err
	}
	body.Name = runWPName(project, args[0])
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Patch(body.Name, body).
		AllowMissing(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing worker pool: %w", err)
	}
	fmt.Printf("Replace request issued for worker pool [%s].\n", args[0])
	return emitFormatted(op, flagRunWPFormat)
}

func runWPUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2WorkerPool{}
	if err := loadYAMLOrJSONInto(flagRunWPConfigFile, body); err != nil {
		return err
	}
	mask := flagRunWPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	body.Name = runWPName(project, args[0])
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.WorkerPools.Patch(body.Name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating worker pool: %w", err)
	}
	fmt.Printf("Update request issued for worker pool [%s].\n", args[0])
	return emitFormatted(op, flagRunWPFormat)
}

func runWPUpdateSplit(cmd *cobra.Command, args []string) error {
	if len(flagRunWPSplitRevs) == 0 && !flagRunWPSplitLatest {
		return fmt.Errorf("one of --to-revisions or --to-latest is required")
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	name := runWPName(project, args[0])
	current, err := svc.Projects.Locations.WorkerPools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("loading worker pool: %w", err)
	}
	var splits []*runv2.GoogleCloudRunV2InstanceSplit
	if flagRunWPSplitLatest {
		splits = append(splits, &runv2.GoogleCloudRunV2InstanceSplit{
			Type: "INSTANCE_SPLIT_ALLOCATION_TYPE_LATEST", Percent: 100,
		})
	}
	for rev, pctStr := range flagRunWPSplitRevs {
		pct, err := parseTrafficPercent(pctStr)
		if err != nil {
			return err
		}
		splits = append(splits, &runv2.GoogleCloudRunV2InstanceSplit{
			Type:     "INSTANCE_SPLIT_ALLOCATION_TYPE_REVISION",
			Revision: rev,
			Percent:  pct,
		})
	}
	patch := &runv2.GoogleCloudRunV2WorkerPool{
		Name:           current.Name,
		InstanceSplits: splits,
	}
	op, err := svc.Projects.Locations.WorkerPools.Patch(name, patch).
		UpdateMask("instanceSplits").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance split: %w", err)
	}
	fmt.Printf("Instance split update issued for worker pool [%s].\n", args[0])
	return emitFormatted(op, flagRunWPFormat)
}

func runWPLogs(cmd *cobra.Command, args []string) error {
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
		`resource.type="cloud_run_worker_pool" AND resource.labels.worker_pool_name=%q AND resource.labels.location=%q`,
		args[0], flagRunWPRegion,
	)
	limit := flagRunWPLimit
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
	if flagRunWPFormat != "" {
		return emitFormatted(resp.Entries, flagRunWPFormat)
	}
	printLogEntries(resp.Entries)
	return nil
}

func runWPDeploy(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	name := runWPName(project, args[0])
	existing, err := svc.Projects.Locations.WorkerPools.Get(name).Context(ctx).Do()
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("checking existing worker pool: %w", err)
	}
	body := &runv2.GoogleCloudRunV2WorkerPool{}
	if existing != nil {
		body = existing
		body.Name = name
	}
	if body.Template == nil {
		body.Template = &runv2.GoogleCloudRunV2WorkerPoolRevisionTemplate{}
	}
	var container *runv2.GoogleCloudRunV2Container
	if len(body.Template.Containers) > 0 {
		container = body.Template.Containers[0]
	} else {
		container = &runv2.GoogleCloudRunV2Container{}
		body.Template.Containers = []*runv2.GoogleCloudRunV2Container{container}
	}
	container.Image = flagRunWPImage
	if flagRunWPCommand != nil {
		container.Command = flagRunWPCommand
	}
	if flagRunWPArgs != nil {
		container.Args = flagRunWPArgs
	}
	if len(flagRunWPEnvVars) > 0 {
		container.Env = envVarsFromMap(flagRunWPEnvVars)
	}
	applyResourceLimits(&container.Resources, flagRunWPMemory, flagRunWPCPU)
	if flagRunWPServiceAcct != "" {
		body.Template.ServiceAccount = flagRunWPServiceAcct
	}
	if flagRunWPVpcConnector != "" {
		if body.Template.VpcAccess == nil {
			body.Template.VpcAccess = &runv2.GoogleCloudRunV2VpcAccess{}
		}
		body.Template.VpcAccess.Connector = flagRunWPVpcConnector
	}

	if existing == nil {
		op, err := svc.Projects.Locations.WorkerPools.Create(runParent(project, flagRunWPRegion), body).
			WorkerPoolId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating worker pool: %w", err)
		}
		fmt.Printf("Created worker pool [%s] (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagRunWPFormat)
	}
	op, err := svc.Projects.Locations.WorkerPools.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating worker pool: %w", err)
	}
	fmt.Printf("Updated worker pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagRunWPFormat)
}

func runWPGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.WorkerPools.GetIamPolicy(runWPName(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagRunWPFormat)
}

func runWPSetIam(cmd *cobra.Command, args []string) error {
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
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.WorkerPools.SetIamPolicy(runWPName(project, args[0]),
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("worker pool [%s]", args[0]))
	return emitFormatted(updated, flagRunWPFormat)
}

func runWPAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	resource := runWPName(project, args[0])
	policy, err := svc.Projects.Locations.WorkerPools.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	runIamAddBinding(policy, flagRunWPIamRole, flagRunWPIamMember,
		runIamBuildCondition(flagRunWPIamCondExpr, flagRunWPIamCondT, flagRunWPIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.WorkerPools.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("worker pool [%s]", args[0]))
	return emitFormatted(updated, flagRunWPFormat)
}

func runWPRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	resource := runWPName(project, args[0])
	policy, err := svc.Projects.Locations.WorkerPools.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !runIamRemoveBinding(policy, flagRunWPIamRole, flagRunWPIamMember,
		runIamBuildCondition(flagRunWPIamCondExpr, flagRunWPIamCondT, flagRunWPIamCondD),
		flagRunWPIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			flagRunWPIamRole, flagRunWPIamMember)
	}
	updated, err := svc.Projects.Locations.WorkerPools.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("worker pool [%s]", args[0]))
	return emitFormatted(updated, flagRunWPFormat)
}

func runWPRevDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkerPools.Revisions.Get(
		runWPRevName(project, flagRunWPRevPool, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing worker pool revision: %w", err)
	}
	return emitFormatted(got, flagRunWPFormat)
}

func runWPRevList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	parent := runWPName(project, flagRunWPRevPool)
	var all []*runv2.GoogleCloudRunV2Revision
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkerPools.Revisions.List(parent).Context(ctx)
		if flagRunWPRevPageSize > 0 {
			call = call.PageSize(flagRunWPRevPageSize)
		}
		if flagRunWPRevShowDel {
			call = call.ShowDeleted(true)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing worker pool revisions: %w", err)
		}
		all = append(all, resp.Revisions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagRunWPFormat)
}

func runWPRevDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunWPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Revisions.Delete(
		runWPRevName(project, flagRunWPRevPool, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting worker pool revision: %w", err)
	}
	fmt.Printf("Delete request issued for worker pool revision [%s].\n", args[0])
	return emitFormatted(op, flagRunWPFormat)
}
