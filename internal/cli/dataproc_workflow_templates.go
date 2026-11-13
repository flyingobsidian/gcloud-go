package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc workflow-templates (#1516) ---

func (s *dataprocCmds) initDataprocWorkflowTemplates() {
	all := []*cobra.Command{
		s.dpWFCreateCmd, s.dpWFDeleteCmd, s.dpWFDescribeCmd, s.dpWFExportCmd, s.dpWFImportCmd,
		s.dpWFListCmd, s.dpWFGetIamCmd, s.dpWFSetIamCmd, s.dpWFInstantiateCmd, s.dpWFInstantiateFromFileCmd,
		s.dpWFAddJobCmd, s.dpWFRemoveJobCmd, s.dpWFSetClusterSelectorCmd, s.dpWFSetManagedClusterCmd,
		s.dpWFSetDagTimeoutCmd, s.dpWFRemoveDagTimeoutCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDPWFRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDPWFFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dpWFCreateCmd, s.dpWFImportCmd, s.dpWFInstantiateFromFileCmd} {
		c.Flags().StringVar(&s.flagDPWFConfigFile, "source", "",
			"Path to a YAML/JSON file with the WorkflowTemplate body (required)")
		_ = c.MarkFlagRequired("source")
	}
	s.dpWFListCmd.Flags().Int64Var(&s.flagDPWFPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{s.dpWFInstantiateCmd, s.dpWFInstantiateFromFileCmd} {
		c.Flags().StringToStringVar(&s.flagDPWFParams, "parameters", nil,
			"KEY=VALUE parameters supplied to the workflow")
		c.Flags().StringVar(&s.flagDPWFRequestID, "request-id", "", "Optional idempotency ID")
	}
	s.dpWFInstantiateCmd.Flags().Int64Var(&s.flagDPWFVersion, "version", 0,
		"Optional template version to instantiate (defaults to latest)")

	s.dpWFAddJobCmd.Flags().StringVar(&s.flagDPWFAddJobFile, "source", "",
		"Path to a YAML/JSON file with the OrderedJob body to append (required)")
	_ = s.dpWFAddJobCmd.MarkFlagRequired("source")
	s.dpWFAddJobCmd.Flags().StringVar(&s.flagDPWFStepID, "step-id", "",
		"Step ID to assign to the job (overrides the file's stepId)")
	s.dpWFRemoveJobCmd.Flags().StringVar(&s.flagDPWFStepID, "step-id", "",
		"Step ID of the job to remove (required)")
	_ = s.dpWFRemoveJobCmd.MarkFlagRequired("step-id")

	s.dpWFSetClusterSelectorCmd.Flags().StringVar(&s.flagDPWFSelectorFile, "source", "",
		"Path to a YAML/JSON file with the ClusterSelector body (required)")
	_ = s.dpWFSetClusterSelectorCmd.MarkFlagRequired("source")

	s.dpWFSetManagedClusterCmd.Flags().StringVar(&s.flagDPWFManagedFile, "source", "",
		"Path to a YAML/JSON file with the ManagedCluster body (required)")
	_ = s.dpWFSetManagedClusterCmd.MarkFlagRequired("source")

	s.dpWFSetDagTimeoutCmd.Flags().StringVar(&s.flagDPWFDagTimeout, "dag-timeout", "",
		"DAG timeout in duration form (e.g. 10m, 1h) (required)")
	_ = s.dpWFSetDagTimeoutCmd.MarkFlagRequired("dag-timeout")

	s.dpWFCmd.AddCommand(all...)
	s.dataprocCmd.AddCommand(s.dpWFCmd)
}

func (s *dataprocCmds) dpWFParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return dpRegionParent(project, s.flagDPWFRegion), nil
}

func (s *dataprocCmds) dpWFName(id string) (string, error) {
	parent, err := s.dpWFParent()
	if err != nil {
		return "", err
	}
	return dpChild("workflowTemplates", id, parent), nil
}

func (s *dataprocCmds) runDPWFCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.dpWFParent()
	if err != nil {
		return err
	}
	body := &dataproc.WorkflowTemplate{}
	if err := loadYAMLOrJSONInto(s.flagDPWFConfigFile, body); err != nil {
		return err
	}
	body.Id = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workflow template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created workflow template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.WorkflowTemplates.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting workflow template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted workflow template [%s].\n", args[0])
	return nil
}

func (s *dataprocCmds) runDPWFDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workflow template: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFExport(cmd *cobra.Command, args []string) error {
	name, err := s.dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting workflow template: %w", err)
	}
	format := s.flagDPWFFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, got, format)
}

func (s *dataprocCmds) runDPWFImport(cmd *cobra.Command, args []string) error {
	parent, err := s.dpWFParent()
	if err != nil {
		return err
	}
	body := &dataproc.WorkflowTemplate{}
	if err := loadYAMLOrJSONInto(s.flagDPWFConfigFile, body); err != nil {
		return err
	}
	body.Id = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	name := dpChild("workflowTemplates", args[0], parent)
	if existing, err := svc.Projects.Regions.WorkflowTemplates.Get(name).Context(ctx).Do(); err == nil {
		body.Name = existing.Name
		body.Version = existing.Version
		got, err := svc.Projects.Regions.WorkflowTemplates.Update(name, body).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating workflow template: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Updated workflow template [%s].\n", args[0])
		return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workflow template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created workflow template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFList(cmd *cobra.Command, args []string) error {
	parent, err := s.dpWFParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.WorkflowTemplate
	pageToken := ""
	for {
		call := svc.Projects.Regions.WorkflowTemplates.List(parent).Context(ctx)
		if s.flagDPWFPageSize > 0 {
			call = call.PageSize(s.flagDPWFPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workflow templates: %w", err)
		}
		all = append(all, resp.Templates...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.WorkflowTemplates.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dpWFName(args[0])
	if err != nil {
		return err
	}
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.WorkflowTemplates.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dpUpdatedIam(fmt.Sprintf("workflow template [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFInstantiate(cmd *cobra.Command, args []string) error {
	name, err := s.dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	req := &dataproc.InstantiateWorkflowTemplateRequest{
		Parameters: s.flagDPWFParams,
		RequestId:  s.flagDPWFRequestID,
	}
	if s.flagDPWFVersion > 0 {
		req.Version = s.flagDPWFVersion
	}
	op, err := svc.Projects.Regions.WorkflowTemplates.Instantiate(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("instantiating workflow template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Instantiate request issued for workflow [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFInstantiateFromFile(cmd *cobra.Command, args []string) error {
	parent, err := s.dpWFParent()
	if err != nil {
		return err
	}
	body := &dataproc.WorkflowTemplate{}
	if err := loadYAMLOrJSONInto(s.flagDPWFConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.WorkflowTemplates.InstantiateInline(parent, body).Context(ctx)
	if s.flagDPWFRequestID != "" {
		call = call.RequestId(s.flagDPWFRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("instantiating inline workflow: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "InstantiateInline request issued (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPWFFormat)
}

func (s *dataprocCmds) loadAndPatchWF(id string, mutate func(*dataproc.WorkflowTemplate) error) (*dataproc.WorkflowTemplate, error) {
	name, err := s.dpWFName(id)
	if err != nil {
		return nil, err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPWFRegion)
	if err != nil {
		return nil, err
	}
	tmpl, err := svc.Projects.Regions.WorkflowTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("describing workflow template: %w", err)
	}
	if err := mutate(tmpl); err != nil {
		return nil, err
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Update(name, tmpl).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("updating workflow template: %w", err)
	}
	return got, nil
}

func (s *dataprocCmds) runDPWFAddJob(cmd *cobra.Command, args []string) error {
	job := &dataproc.OrderedJob{}
	if err := loadYAMLOrJSONInto(s.flagDPWFAddJobFile, job); err != nil {
		return err
	}
	if s.flagDPWFStepID != "" {
		job.StepId = s.flagDPWFStepID
	}
	if job.StepId == "" {
		return fmt.Errorf("job must have a stepId (via file or --step-id)")
	}
	got, err := s.loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		for _, j := range t.Jobs {
			if j.StepId == job.StepId {
				return fmt.Errorf("job with step ID [%s] already exists", job.StepId)
			}
		}
		t.Jobs = append(t.Jobs, job)
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Added job [%s] to workflow template [%s].\n", job.StepId, args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFRemoveJob(cmd *cobra.Command, args []string) error {
	got, err := s.loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		kept := t.Jobs[:0]
		found := false
		for _, j := range t.Jobs {
			if j.StepId == s.flagDPWFStepID {
				found = true
				continue
			}
			kept = append(kept, j)
		}
		if !found {
			return fmt.Errorf("no job with step ID [%s] in workflow template", s.flagDPWFStepID)
		}
		t.Jobs = kept
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Removed job [%s] from workflow template [%s].\n", s.flagDPWFStepID, args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFSetClusterSelector(cmd *cobra.Command, args []string) error {
	sel := &dataproc.ClusterSelector{}
	if err := loadYAMLOrJSONInto(s.flagDPWFSelectorFile, sel); err != nil {
		return err
	}
	got, err := s.loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		if t.Placement == nil {
			t.Placement = &dataproc.WorkflowTemplatePlacement{}
		}
		t.Placement.ClusterSelector = sel
		t.Placement.ManagedCluster = nil
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Set cluster-selector on workflow template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFSetManagedCluster(cmd *cobra.Command, args []string) error {
	mc := &dataproc.ManagedCluster{}
	if err := loadYAMLOrJSONInto(s.flagDPWFManagedFile, mc); err != nil {
		return err
	}
	got, err := s.loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		if t.Placement == nil {
			t.Placement = &dataproc.WorkflowTemplatePlacement{}
		}
		t.Placement.ManagedCluster = mc
		t.Placement.ClusterSelector = nil
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Set managed-cluster on workflow template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFSetDagTimeout(cmd *cobra.Command, args []string) error {
	got, err := s.loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		t.DagTimeout = s.flagDPWFDagTimeout
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Set DAG timeout on workflow template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}

func (s *dataprocCmds) runDPWFRemoveDagTimeout(cmd *cobra.Command, args []string) error {
	got, err := s.loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		t.DagTimeout = ""
		t.NullFields = append(t.NullFields, "DagTimeout")
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Removed DAG timeout from workflow template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPWFFormat)
}
