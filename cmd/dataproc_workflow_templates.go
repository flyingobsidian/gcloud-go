package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc workflow-templates (#1516) ---

var dpWFCmd = &cobra.Command{Use: "workflow-templates", Short: "Manage Dataproc workflow templates"}

var (
	flagDPWFRegion        string
	flagDPWFFormat        string
	flagDPWFConfigFile    string
	flagDPWFPageSize      int64
	flagDPWFParams        map[string]string
	flagDPWFRequestID     string
	flagDPWFAddJobFile    string
	flagDPWFRemoveJob     string
	flagDPWFSelectorFile  string
	flagDPWFManagedFile   string
	flagDPWFDagTimeout    string
	flagDPWFStepID        string
	flagDPWFVersion       int64
)

var (
	dpWFCreateCmd = &cobra.Command{
		Use: "create TEMPLATE", Short: "Create a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFCreate,
	}
	dpWFDeleteCmd = &cobra.Command{
		Use: "delete TEMPLATE", Short: "Delete a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFDelete,
	}
	dpWFDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFDescribe,
	}
	dpWFExportCmd = &cobra.Command{
		Use: "export TEMPLATE", Short: "Export a workflow template (YAML/JSON)",
		Args: cobra.ExactArgs(1), RunE: runDPWFExport,
	}
	dpWFImportCmd = &cobra.Command{
		Use: "import TEMPLATE", Short: "Import (create or update) a workflow template from a file",
		Args: cobra.ExactArgs(1), RunE: runDPWFImport,
	}
	dpWFListCmd = &cobra.Command{
		Use: "list", Short: "List workflow templates",
		Args: cobra.NoArgs, RunE: runDPWFList,
	}
	dpWFGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TEMPLATE", Short: "Get the IAM policy for a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFGetIam,
	}
	dpWFSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TEMPLATE POLICY_FILE", Short: "Set the IAM policy for a workflow template",
		Args: cobra.ExactArgs(2), RunE: runDPWFSetIam,
	}
	dpWFInstantiateCmd = &cobra.Command{
		Use: "instantiate TEMPLATE", Short: "Instantiate a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFInstantiate,
	}
	dpWFInstantiateFromFileCmd = &cobra.Command{
		Use: "instantiate-from-file", Short: "Instantiate an inline workflow template loaded from a file",
		Args: cobra.NoArgs, RunE: runDPWFInstantiateFromFile,
	}
	dpWFAddJobCmd = &cobra.Command{
		Use: "add-job TEMPLATE", Short: "Add a job to a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFAddJob,
	}
	dpWFRemoveJobCmd = &cobra.Command{
		Use: "remove-job TEMPLATE", Short: "Remove a job from a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFRemoveJob,
	}
	dpWFSetClusterSelectorCmd = &cobra.Command{
		Use: "set-cluster-selector TEMPLATE", Short: "Set the ClusterSelector for a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFSetClusterSelector,
	}
	dpWFSetManagedClusterCmd = &cobra.Command{
		Use: "set-managed-cluster TEMPLATE", Short: "Set the ManagedCluster for a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFSetManagedCluster,
	}
	dpWFSetDagTimeoutCmd = &cobra.Command{
		Use: "set-dag-timeout TEMPLATE", Short: "Set the DAG timeout for a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFSetDagTimeout,
	}
	dpWFRemoveDagTimeoutCmd = &cobra.Command{
		Use: "remove-dag-timeout TEMPLATE", Short: "Remove the DAG timeout from a workflow template",
		Args: cobra.ExactArgs(1), RunE: runDPWFRemoveDagTimeout,
	}
)

func init() {
	all := []*cobra.Command{
		dpWFCreateCmd, dpWFDeleteCmd, dpWFDescribeCmd, dpWFExportCmd, dpWFImportCmd,
		dpWFListCmd, dpWFGetIamCmd, dpWFSetIamCmd, dpWFInstantiateCmd, dpWFInstantiateFromFileCmd,
		dpWFAddJobCmd, dpWFRemoveJobCmd, dpWFSetClusterSelectorCmd, dpWFSetManagedClusterCmd,
		dpWFSetDagTimeoutCmd, dpWFRemoveDagTimeoutCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDPWFRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDPWFFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dpWFCreateCmd, dpWFImportCmd, dpWFInstantiateFromFileCmd} {
		c.Flags().StringVar(&flagDPWFConfigFile, "source", "",
			"Path to a YAML/JSON file with the WorkflowTemplate body (required)")
		_ = c.MarkFlagRequired("source")
	}
	dpWFListCmd.Flags().Int64Var(&flagDPWFPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{dpWFInstantiateCmd, dpWFInstantiateFromFileCmd} {
		c.Flags().StringToStringVar(&flagDPWFParams, "parameters", nil,
			"KEY=VALUE parameters supplied to the workflow")
		c.Flags().StringVar(&flagDPWFRequestID, "request-id", "", "Optional idempotency ID")
	}
	dpWFInstantiateCmd.Flags().Int64Var(&flagDPWFVersion, "version", 0,
		"Optional template version to instantiate (defaults to latest)")

	dpWFAddJobCmd.Flags().StringVar(&flagDPWFAddJobFile, "source", "",
		"Path to a YAML/JSON file with the OrderedJob body to append (required)")
	_ = dpWFAddJobCmd.MarkFlagRequired("source")
	dpWFAddJobCmd.Flags().StringVar(&flagDPWFStepID, "step-id", "",
		"Step ID to assign to the job (overrides the file's stepId)")
	dpWFRemoveJobCmd.Flags().StringVar(&flagDPWFStepID, "step-id", "",
		"Step ID of the job to remove (required)")
	_ = dpWFRemoveJobCmd.MarkFlagRequired("step-id")

	dpWFSetClusterSelectorCmd.Flags().StringVar(&flagDPWFSelectorFile, "source", "",
		"Path to a YAML/JSON file with the ClusterSelector body (required)")
	_ = dpWFSetClusterSelectorCmd.MarkFlagRequired("source")

	dpWFSetManagedClusterCmd.Flags().StringVar(&flagDPWFManagedFile, "source", "",
		"Path to a YAML/JSON file with the ManagedCluster body (required)")
	_ = dpWFSetManagedClusterCmd.MarkFlagRequired("source")

	dpWFSetDagTimeoutCmd.Flags().StringVar(&flagDPWFDagTimeout, "dag-timeout", "",
		"DAG timeout in duration form (e.g. 10m, 1h) (required)")
	_ = dpWFSetDagTimeoutCmd.MarkFlagRequired("dag-timeout")

	dpWFCmd.AddCommand(all...)
	dataprocCmd.AddCommand(dpWFCmd)
}

func dpWFParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return dpRegionParent(project, flagDPWFRegion), nil
}

func dpWFName(id string) (string, error) {
	parent, err := dpWFParent()
	if err != nil {
		return "", err
	}
	return dpChild("workflowTemplates", id, parent), nil
}

func runDPWFCreate(cmd *cobra.Command, args []string) error {
	parent, err := dpWFParent()
	if err != nil {
		return err
	}
	body := &dataproc.WorkflowTemplate{}
	if err := loadYAMLOrJSONInto(flagDPWFConfigFile, body); err != nil {
		return err
	}
	body.Id = args[0]
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workflow template: %w", err)
	}
	fmt.Printf("Created workflow template [%s].\n", args[0])
	return emitFormatted(got, flagDPWFFormat)
}

func runDPWFDelete(cmd *cobra.Command, args []string) error {
	name, err := dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.WorkflowTemplates.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting workflow template: %w", err)
	}
	fmt.Printf("Deleted workflow template [%s].\n", args[0])
	return nil
}

func runDPWFDescribe(cmd *cobra.Command, args []string) error {
	name, err := dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workflow template: %w", err)
	}
	return emitFormatted(got, flagDPWFFormat)
}

func runDPWFExport(cmd *cobra.Command, args []string) error {
	name, err := dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting workflow template: %w", err)
	}
	format := flagDPWFFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormatted(got, format)
}

func runDPWFImport(cmd *cobra.Command, args []string) error {
	parent, err := dpWFParent()
	if err != nil {
		return err
	}
	body := &dataproc.WorkflowTemplate{}
	if err := loadYAMLOrJSONInto(flagDPWFConfigFile, body); err != nil {
		return err
	}
	body.Id = args[0]
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
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
		fmt.Printf("Updated workflow template [%s].\n", args[0])
		return emitFormatted(got, flagDPWFFormat)
	}
	got, err := svc.Projects.Regions.WorkflowTemplates.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workflow template: %w", err)
	}
	fmt.Printf("Created workflow template [%s].\n", args[0])
	return emitFormatted(got, flagDPWFFormat)
}

func runDPWFList(cmd *cobra.Command, args []string) error {
	parent, err := dpWFParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.WorkflowTemplate
	pageToken := ""
	for {
		call := svc.Projects.Regions.WorkflowTemplates.List(parent).Context(ctx)
		if flagDPWFPageSize > 0 {
			call = call.PageSize(flagDPWFPageSize)
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
	return emitFormatted(all, flagDPWFFormat)
}

func runDPWFGetIam(cmd *cobra.Command, args []string) error {
	name, err := dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.WorkflowTemplates.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDPWFFormat)
}

func runDPWFSetIam(cmd *cobra.Command, args []string) error {
	name, err := dpWFName(args[0])
	if err != nil {
		return err
	}
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.WorkflowTemplates.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dpUpdatedIam(fmt.Sprintf("workflow template [%s]", args[0]))
	return emitFormatted(updated, flagDPWFFormat)
}

func runDPWFInstantiate(cmd *cobra.Command, args []string) error {
	name, err := dpWFName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	req := &dataproc.InstantiateWorkflowTemplateRequest{
		Parameters: flagDPWFParams,
		RequestId:  flagDPWFRequestID,
	}
	if flagDPWFVersion > 0 {
		req.Version = flagDPWFVersion
	}
	op, err := svc.Projects.Regions.WorkflowTemplates.Instantiate(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("instantiating workflow template: %w", err)
	}
	fmt.Printf("Instantiate request issued for workflow [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPWFFormat)
}

func runDPWFInstantiateFromFile(cmd *cobra.Command, args []string) error {
	parent, err := dpWFParent()
	if err != nil {
		return err
	}
	body := &dataproc.WorkflowTemplate{}
	if err := loadYAMLOrJSONInto(flagDPWFConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.WorkflowTemplates.InstantiateInline(parent, body).Context(ctx)
	if flagDPWFRequestID != "" {
		call = call.RequestId(flagDPWFRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("instantiating inline workflow: %w", err)
	}
	fmt.Printf("InstantiateInline request issued (operation: %s).\n", op.Name)
	return emitFormatted(op, flagDPWFFormat)
}

func loadAndPatchWF(id string, mutate func(*dataproc.WorkflowTemplate) error) (*dataproc.WorkflowTemplate, error) {
	name, err := dpWFName(id)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPWFRegion)
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

func runDPWFAddJob(cmd *cobra.Command, args []string) error {
	job := &dataproc.OrderedJob{}
	if err := loadYAMLOrJSONInto(flagDPWFAddJobFile, job); err != nil {
		return err
	}
	if flagDPWFStepID != "" {
		job.StepId = flagDPWFStepID
	}
	if job.StepId == "" {
		return fmt.Errorf("job must have a stepId (via file or --step-id)")
	}
	got, err := loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
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
	fmt.Printf("Added job [%s] to workflow template [%s].\n", job.StepId, args[0])
	return emitFormatted(got, flagDPWFFormat)
}

func runDPWFRemoveJob(cmd *cobra.Command, args []string) error {
	got, err := loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		kept := t.Jobs[:0]
		found := false
		for _, j := range t.Jobs {
			if j.StepId == flagDPWFStepID {
				found = true
				continue
			}
			kept = append(kept, j)
		}
		if !found {
			return fmt.Errorf("no job with step ID [%s] in workflow template", flagDPWFStepID)
		}
		t.Jobs = kept
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Removed job [%s] from workflow template [%s].\n", flagDPWFStepID, args[0])
	return emitFormatted(got, flagDPWFFormat)
}

func runDPWFSetClusterSelector(cmd *cobra.Command, args []string) error {
	sel := &dataproc.ClusterSelector{}
	if err := loadYAMLOrJSONInto(flagDPWFSelectorFile, sel); err != nil {
		return err
	}
	got, err := loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
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
	fmt.Printf("Set cluster-selector on workflow template [%s].\n", args[0])
	return emitFormatted(got, flagDPWFFormat)
}

func runDPWFSetManagedCluster(cmd *cobra.Command, args []string) error {
	mc := &dataproc.ManagedCluster{}
	if err := loadYAMLOrJSONInto(flagDPWFManagedFile, mc); err != nil {
		return err
	}
	got, err := loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
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
	fmt.Printf("Set managed-cluster on workflow template [%s].\n", args[0])
	return emitFormatted(got, flagDPWFFormat)
}

func runDPWFSetDagTimeout(cmd *cobra.Command, args []string) error {
	got, err := loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		t.DagTimeout = flagDPWFDagTimeout
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Set DAG timeout on workflow template [%s].\n", args[0])
	return emitFormatted(got, flagDPWFFormat)
}

func runDPWFRemoveDagTimeout(cmd *cobra.Command, args []string) error {
	got, err := loadAndPatchWF(args[0], func(t *dataproc.WorkflowTemplate) error {
		t.DagTimeout = ""
		t.NullFields = append(t.NullFields, "DagTimeout")
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Removed DAG timeout from workflow template [%s].\n", args[0])
	return emitFormatted(got, flagDPWFFormat)
}
