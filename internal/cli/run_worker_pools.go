package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run worker-pools (#1055) ---
//
// Backed by run/v2 Projects.Locations.WorkerPools and .Revisions.

func (s *runCmds) initRunWorkerPools() {
	poolAll := []*cobra.Command{
		s.runWPDeleteCmd, s.runWPDeployCmd, s.runWPDescribeCmd, s.runWPListCmd, s.runWPLogsCmd,
		s.runWPReplaceCmd, s.runWPUpdateCmd, s.runWPUpdateSplitCmd, s.runWPGetIamCmd,
		s.runWPSetIamCmd, s.runWPAddIamCmd, s.runWPRemoveIamCmd,
	}
	for _, c := range poolAll {
		c.Flags().StringVar(&s.flagRunWPRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&s.flagRunWPFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.runWPReplaceCmd, s.runWPUpdateCmd} {
		c.Flags().StringVar(&s.flagRunWPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the WorkerPool body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.runWPUpdateCmd.Flags().StringVar(&s.flagRunWPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")

	s.runWPListCmd.Flags().Int64Var(&s.flagRunWPPageSize, "page-size", 0, "Maximum results per page")
	s.runWPListCmd.Flags().BoolVar(&s.flagRunWPShowDel, "show-deleted", false, "Include deleted worker pools")
	s.runWPLogsCmd.Flags().Int64Var(&s.flagRunWPLimit, "limit", 100, "Maximum number of log entries to return")

	s.runWPDeployCmd.Flags().StringVar(&s.flagRunWPImage, "image", "", "Container image (required)")
	_ = s.runWPDeployCmd.MarkFlagRequired("image")
	s.runWPDeployCmd.Flags().StringSliceVar(&s.flagRunWPCommand, "command", nil, "Container entrypoint override")
	s.runWPDeployCmd.Flags().StringSliceVar(&s.flagRunWPArgs, "args", nil, "Container arguments override")
	s.runWPDeployCmd.Flags().StringToStringVar(&s.flagRunWPEnvVars, "env-vars", nil, "Container environment variables (KEY=VALUE)")
	s.runWPDeployCmd.Flags().StringVar(&s.flagRunWPServiceAcct, "service-account", "", "Service account for the worker pool")
	s.runWPDeployCmd.Flags().StringVar(&s.flagRunWPMemory, "memory", "", "Memory limit (e.g. 512Mi)")
	s.runWPDeployCmd.Flags().StringVar(&s.flagRunWPCPU, "cpu", "", "CPU limit (e.g. 1, 2, 4)")
	s.runWPDeployCmd.Flags().StringVar(&s.flagRunWPVpcConnector, "vpc-connector", "", "VPC connector to use")

	s.runWPUpdateSplitCmd.Flags().StringToStringVar(&s.flagRunWPSplitRevs, "to-revisions", nil,
		"Revision-to-percent map, e.g. REV1=60,REV2=40")
	s.runWPUpdateSplitCmd.Flags().BoolVar(&s.flagRunWPSplitLatest, "to-latest", false,
		"Route 100% of instances to the latest ready revision")

	for _, c := range []*cobra.Command{s.runWPAddIamCmd, s.runWPRemoveIamCmd} {
		runIamFlags(c, &s.flagRunWPIamMember, &s.flagRunWPIamRole,
			&s.flagRunWPIamCondExpr, &s.flagRunWPIamCondT, &s.flagRunWPIamCondD)
	}
	s.runWPRemoveIamCmd.Flags().BoolVar(&s.flagRunWPIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	// Revisions subgroup
	revAll := []*cobra.Command{s.runWPRevDescribeCmd, s.runWPRevListCmd, s.runWPRevDeleteCmd}
	for _, c := range revAll {
		c.Flags().StringVar(&s.flagRunWPRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&s.flagRunWPRevPool, "worker-pool", "", "Parent Cloud Run worker pool (required)")
		c.Flags().StringVar(&s.flagRunWPFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("worker-pool")
	}
	s.runWPRevListCmd.Flags().Int64Var(&s.flagRunWPRevPageSize, "page-size", 0, "Maximum results per page")
	s.runWPRevListCmd.Flags().BoolVar(&s.flagRunWPRevShowDel, "show-deleted", false, "Include deleted revisions")

	s.runWorkerPoolsRevisionsCmd.AddCommand(revAll...)
	s.runWorkerPoolsCmd.AddCommand(poolAll...)
	s.runWorkerPoolsCmd.AddCommand(s.runWorkerPoolsRevisionsCmd)
	s.runCmd.AddCommand(s.runWorkerPoolsCmd)
}

func (s *runCmds) runWPName(project, pool string) string {
	return runResourceName(project, s.flagRunWPRegion, "workerPools", pool)
}

func (s *runCmds) runWPRevName(project, pool, rev string) string {
	if hasProjectsPrefix(rev) {
		return rev
	}
	return fmt.Sprintf("%s/revisions/%s", s.runWPName(project, pool), rev)
}

func (s *runCmds) runWPDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Delete(s.runWPName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting worker pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for worker pool [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunWPFormat)
}

func (s *runCmds) runWPDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkerPools.Get(s.runWPName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing worker pool: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRunWPFormat)
}

func (s *runCmds) runWPList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	var all []*runv2.GoogleCloudRunV2WorkerPool
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkerPools.List(runParent(project, s.flagRunWPRegion)).Context(ctx)
		if s.flagRunWPPageSize > 0 {
			call = call.PageSize(s.flagRunWPPageSize)
		}
		if s.flagRunWPShowDel {
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
	return emitFormattedTo(s.inv.stdout, all, s.flagRunWPFormat)
}

func (s *runCmds) runWPReplace(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2WorkerPool{}
	if err := loadYAMLOrJSONInto(s.flagRunWPConfigFile, body); err != nil {
		return err
	}
	body.Name = s.runWPName(project, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Patch(body.Name, body).
		AllowMissing(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing worker pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Replace request issued for worker pool [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunWPFormat)
}

func (s *runCmds) runWPUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2WorkerPool{}
	if err := loadYAMLOrJSONInto(s.flagRunWPConfigFile, body); err != nil {
		return err
	}
	mask := s.flagRunWPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	body.Name = s.runWPName(project, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
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
	fmt.Fprintf(s.inv.stdout, "Update request issued for worker pool [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunWPFormat)
}

func (s *runCmds) runWPUpdateSplit(cmd *cobra.Command, args []string) error {
	if len(s.flagRunWPSplitRevs) == 0 && !s.flagRunWPSplitLatest {
		return fmt.Errorf("one of --to-revisions or --to-latest is required")
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	name := s.runWPName(project, args[0])
	current, err := svc.Projects.Locations.WorkerPools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("loading worker pool: %w", err)
	}
	var splits []*runv2.GoogleCloudRunV2InstanceSplit
	if s.flagRunWPSplitLatest {
		splits = append(splits, &runv2.GoogleCloudRunV2InstanceSplit{
			Type: "INSTANCE_SPLIT_ALLOCATION_TYPE_LATEST", Percent: 100,
		})
	}
	for rev, pctStr := range s.flagRunWPSplitRevs {
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
	fmt.Fprintf(s.inv.stdout, "Instance split update issued for worker pool [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunWPFormat)
}

func (s *runCmds) runWPLogs(cmd *cobra.Command, args []string) error {
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
		`resource.type="cloud_run_worker_pool" AND resource.labels.worker_pool_name=%q AND resource.labels.location=%q`,
		args[0], s.flagRunWPRegion,
	)
	limit := s.flagRunWPLimit
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
	if s.flagRunWPFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Entries, s.flagRunWPFormat)
	}
	s.printLogEntries(resp.Entries)
	return nil
}

func (s *runCmds) runWPDeploy(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	name := s.runWPName(project, args[0])
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
	container.Image = s.flagRunWPImage
	if s.flagRunWPCommand != nil {
		container.Command = s.flagRunWPCommand
	}
	if s.flagRunWPArgs != nil {
		container.Args = s.flagRunWPArgs
	}
	if len(s.flagRunWPEnvVars) > 0 {
		container.Env = envVarsFromMap(s.flagRunWPEnvVars)
	}
	applyResourceLimits(&container.Resources, s.flagRunWPMemory, s.flagRunWPCPU)
	if s.flagRunWPServiceAcct != "" {
		body.Template.ServiceAccount = s.flagRunWPServiceAcct
	}
	if s.flagRunWPVpcConnector != "" {
		if body.Template.VpcAccess == nil {
			body.Template.VpcAccess = &runv2.GoogleCloudRunV2VpcAccess{}
		}
		body.Template.VpcAccess.Connector = s.flagRunWPVpcConnector
	}

	if existing == nil {
		op, err := svc.Projects.Locations.WorkerPools.Create(runParent(project, s.flagRunWPRegion), body).
			WorkerPoolId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating worker pool: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Created worker pool [%s] (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagRunWPFormat)
	}
	op, err := svc.Projects.Locations.WorkerPools.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating worker pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated worker pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagRunWPFormat)
}

func (s *runCmds) runWPGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.WorkerPools.GetIamPolicy(s.runWPName(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagRunWPFormat)
}

func (s *runCmds) runWPSetIam(cmd *cobra.Command, args []string) error {
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
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.WorkerPools.SetIamPolicy(s.runWPName(project, args[0]),
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("worker pool [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunWPFormat)
}

func (s *runCmds) runWPAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	resource := s.runWPName(project, args[0])
	policy, err := svc.Projects.Locations.WorkerPools.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	runIamAddBinding(policy, s.flagRunWPIamRole, s.flagRunWPIamMember,
		runIamBuildCondition(s.flagRunWPIamCondExpr, s.flagRunWPIamCondT, s.flagRunWPIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.WorkerPools.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("worker pool [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunWPFormat)
}

func (s *runCmds) runWPRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	resource := s.runWPName(project, args[0])
	policy, err := svc.Projects.Locations.WorkerPools.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !runIamRemoveBinding(policy, s.flagRunWPIamRole, s.flagRunWPIamMember,
		runIamBuildCondition(s.flagRunWPIamCondExpr, s.flagRunWPIamCondT, s.flagRunWPIamCondD),
		s.flagRunWPIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			s.flagRunWPIamRole, s.flagRunWPIamMember)
	}
	updated, err := svc.Projects.Locations.WorkerPools.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("worker pool [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunWPFormat)
}

func (s *runCmds) runWPRevDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkerPools.Revisions.Get(
		s.runWPRevName(project, s.flagRunWPRevPool, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing worker pool revision: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRunWPFormat)
}

func (s *runCmds) runWPRevList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	parent := s.runWPName(project, s.flagRunWPRevPool)
	var all []*runv2.GoogleCloudRunV2Revision
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkerPools.Revisions.List(parent).Context(ctx)
		if s.flagRunWPRevPageSize > 0 {
			call = call.PageSize(s.flagRunWPRevPageSize)
		}
		if s.flagRunWPRevShowDel {
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
	return emitFormattedTo(s.inv.stdout, all, s.flagRunWPFormat)
}

func (s *runCmds) runWPRevDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunWPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Revisions.Delete(
		s.runWPRevName(project, s.flagRunWPRevPool, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting worker pool revision: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for worker pool revision [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunWPFormat)
}
