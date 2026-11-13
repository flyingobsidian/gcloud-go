package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	config1 "google.golang.org/api/config/v1"
)

// --- gcloud infra-manager (#348, #795-#802) ---

func imParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func imChild(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func (s *infraManagerCmds) imResolveParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagIMLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return imParent(project, s.flagIMLocation), nil
}

func imWaitOp(ctx context.Context, svc *config1.Service, op *config1.Operation) (*config1.Operation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (s *infraManagerCmds) imFinishOp(ctx context.Context, svc *config1.Service, op *config1.Operation, verb, name string) error {
	if s.flagIMAsync {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := imWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// --- automigrationconfig ---

// --- deployments ---

// --- previews ---

// --- resource-changes ---

// --- resource-drifts ---

// --- resources ---

// --- revisions ---

// --- terraform-versions ---

func (s *infraManagerCmds) initInfraManager() {
	addLoc := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagIMLocation, "location", "", "Location (required)")
		}
	}
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagIMFormat, "format", "", "Output format")
		}
	}
	addFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&s.flagIMFilter, "filter", "", "Server-side list filter")
		}
	}
	addAsync := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().BoolVar(&s.flagIMAsync, "async", false, "Do not wait for the operation to complete")
			c.Flags().StringVar(&s.flagIMRequestID, "request-id", "", "Optional idempotency request ID")
		}
	}

	// automigrationconfig
	addLoc(s.imAmcDescribeCmd, s.imAmcEnableCmd, s.imAmcDisableCmd)
	addFmt(s.imAmcDescribeCmd)
	s.imAutoMigrationCfgCmd.AddCommand(s.imAmcDescribeCmd, s.imAmcEnableCmd, s.imAmcDisableCmd)
	s.infraManagerCmd.AddCommand(s.imAutoMigrationCfgCmd)

	// deployments
	addLoc(s.imDepApplyCmd, s.imDepDeleteCmd, s.imDepDescribeCmd, s.imDepListCmd, s.imDepExportLockCmd,
		s.imDepExportStateCmd, s.imDepImportStateCmd, s.imDepLockCmd, s.imDepUnlockCmd)
	addFmt(s.imDepDescribeCmd, s.imDepListCmd, s.imDepExportLockCmd)
	addFilter(s.imDepListCmd)
	addAsync(s.imDepApplyCmd, s.imDepDeleteCmd, s.imDepImportStateCmd, s.imDepLockCmd, s.imDepUnlockCmd)
	s.imDepApplyCmd.Flags().StringVar(&s.flagIMConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Deployment body (required)")
	_ = s.imDepApplyCmd.MarkFlagRequired("config-file")
	s.imDepApplyCmd.Flags().StringVar(&s.flagIMUpdateMask, "update-mask", "",
		"For update, comma-separated list of fields to update")
	s.imDepDeleteCmd.Flags().StringVar(&s.flagIMDeletePolicy, "delete-policy", "",
		"Deletion policy (DELETE, ABANDON)")
	s.imDepDeleteCmd.Flags().BoolVar(&s.flagIMForceDelete, "force", false, "Force deletion")
	s.imDepImportStateCmd.Flags().StringVar(&s.flagIMExportPath, "statefile", "",
		"Path to a JSON/YAML file containing the statefile body (required)")
	_ = s.imDepImportStateCmd.MarkFlagRequired("statefile")
	s.imDepLockCmd.Flags().StringVar(&s.flagIMDraftID, "draft-id", "", "Optional draft ID for the lock request")
	s.imDepUnlockCmd.Flags().Int64Var(&s.flagIMLockID, "lock-id", 0, "Lock ID to release (required)")
	_ = s.imDepUnlockCmd.MarkFlagRequired("lock-id")
	s.imDepExportStateCmd.Flags().StringVar(&s.flagIMExportBucket, "draft-id", "", "Optional draft ID for the export request")
	s.imDeploymentsCmd.AddCommand(s.imDepApplyCmd, s.imDepDeleteCmd, s.imDepDescribeCmd, s.imDepListCmd,
		s.imDepExportLockCmd, s.imDepExportStateCmd, s.imDepImportStateCmd, s.imDepLockCmd, s.imDepUnlockCmd)
	s.infraManagerCmd.AddCommand(s.imDeploymentsCmd)

	// previews
	addLoc(s.imPrevCreateCmd, s.imPrevDeleteCmd, s.imPrevDescribeCmd, s.imPrevListCmd, s.imPrevExportCmd)
	addFmt(s.imPrevDescribeCmd, s.imPrevListCmd, s.imPrevExportCmd)
	addFilter(s.imPrevListCmd)
	addAsync(s.imPrevCreateCmd, s.imPrevDeleteCmd)
	s.imPrevCreateCmd.Flags().StringVar(&s.flagIMConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Preview body (required)")
	_ = s.imPrevCreateCmd.MarkFlagRequired("config-file")
	s.imPreviewsCmd.AddCommand(s.imPrevCreateCmd, s.imPrevDeleteCmd, s.imPrevDescribeCmd, s.imPrevListCmd, s.imPrevExportCmd)
	s.infraManagerCmd.AddCommand(s.imPreviewsCmd)

	// resource-changes
	addLoc(s.imRcDescribeCmd, s.imRcListCmd)
	addFmt(s.imRcDescribeCmd, s.imRcListCmd)
	addFilter(s.imRcListCmd)
	s.imRcDescribeCmd.Flags().StringVar(&s.flagIMPreview, "preview", "", "Owning preview (required)")
	_ = s.imRcDescribeCmd.MarkFlagRequired("preview")
	s.imRcListCmd.Flags().StringVar(&s.flagIMPreview, "preview", "", "Owning preview (required)")
	_ = s.imRcListCmd.MarkFlagRequired("preview")
	s.imResourceChangesCmd.AddCommand(s.imRcDescribeCmd, s.imRcListCmd)
	s.infraManagerCmd.AddCommand(s.imResourceChangesCmd)

	// resource-drifts
	addLoc(s.imRdDescribeCmd, s.imRdListCmd)
	addFmt(s.imRdDescribeCmd, s.imRdListCmd)
	addFilter(s.imRdListCmd)
	s.imRdDescribeCmd.Flags().StringVar(&s.flagIMPreview, "preview", "", "Owning preview (required)")
	_ = s.imRdDescribeCmd.MarkFlagRequired("preview")
	s.imRdListCmd.Flags().StringVar(&s.flagIMPreview, "preview", "", "Owning preview (required)")
	_ = s.imRdListCmd.MarkFlagRequired("preview")
	s.imResourceDriftsCmd.AddCommand(s.imRdDescribeCmd, s.imRdListCmd)
	s.infraManagerCmd.AddCommand(s.imResourceDriftsCmd)

	// resources
	addLoc(s.imResDescribeCmd, s.imResListCmd)
	addFmt(s.imResDescribeCmd, s.imResListCmd)
	addFilter(s.imResListCmd)
	s.imResDescribeCmd.Flags().StringVar(&s.flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = s.imResDescribeCmd.MarkFlagRequired("deployment")
	s.imResDescribeCmd.Flags().StringVar(&s.flagIMRevision, "revision", "", "Owning revision (required)")
	_ = s.imResDescribeCmd.MarkFlagRequired("revision")
	s.imResListCmd.Flags().StringVar(&s.flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = s.imResListCmd.MarkFlagRequired("deployment")
	s.imResListCmd.Flags().StringVar(&s.flagIMRevision, "revision", "", "Owning revision (required)")
	_ = s.imResListCmd.MarkFlagRequired("revision")
	s.imResourcesCmd.AddCommand(s.imResDescribeCmd, s.imResListCmd)
	s.infraManagerCmd.AddCommand(s.imResourcesCmd)

	// revisions
	addLoc(s.imRevDescribeCmd, s.imRevListCmd, s.imRevExportStateCmd)
	addFmt(s.imRevDescribeCmd, s.imRevListCmd, s.imRevExportStateCmd)
	addFilter(s.imRevListCmd)
	s.imRevDescribeCmd.Flags().StringVar(&s.flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = s.imRevDescribeCmd.MarkFlagRequired("deployment")
	s.imRevListCmd.Flags().StringVar(&s.flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = s.imRevListCmd.MarkFlagRequired("deployment")
	s.imRevExportStateCmd.Flags().StringVar(&s.flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = s.imRevExportStateCmd.MarkFlagRequired("deployment")
	s.imRevisionsCmd.AddCommand(s.imRevDescribeCmd, s.imRevListCmd, s.imRevExportStateCmd)
	s.infraManagerCmd.AddCommand(s.imRevisionsCmd)

	// terraform-versions
	addLoc(s.imTfDescribeCmd, s.imTfListCmd)
	addFmt(s.imTfDescribeCmd, s.imTfListCmd)
	addFilter(s.imTfListCmd)
	s.imTerraformVersionsCmd.AddCommand(s.imTfDescribeCmd, s.imTfListCmd)
	s.infraManagerCmd.AddCommand(s.imTerraformVersionsCmd)

}

// --- automigrationconfig impl ---

func imAmcName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/autoMigrationConfig", project, location)
}

func (s *infraManagerCmds) runIMAmcDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if s.flagIMLocation == "" {
		return fmt.Errorf("--location is required")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imAmcName(project, s.flagIMLocation)
	got, err := svc.Projects.Locations.GetAutoMigrationConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing auto migration config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMAmcEnable(cmd *cobra.Command, args []string) error {
	return s.imAmcUpdateEnabled(true)
}

func (s *infraManagerCmds) runIMAmcDisable(cmd *cobra.Command, args []string) error {
	return s.imAmcUpdateEnabled(false)
}

func (s *infraManagerCmds) imAmcUpdateEnabled(enabled bool) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if s.flagIMLocation == "" {
		return fmt.Errorf("--location is required")
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imAmcName(project, s.flagIMLocation)
	body := &config1.AutoMigrationConfig{AutoMigrationEnabled: enabled, ForceSendFields: []string{"AutoMigrationEnabled"}}
	got, err := svc.Projects.Locations.UpdateAutoMigrationConfig(name, body).UpdateMask("autoMigrationEnabled").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating auto migration config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- deployments impl ---

func (s *infraManagerCmds) runIMDepApply(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	body := &config1.Deployment{}
	if err := loadYAMLOrJSONInto(s.flagIMConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	existing, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do()
	if err == nil && existing != nil {
		mask := s.flagIMUpdateMask
		if mask == "" {
			mask = joinMask(nonEmptyJSONFields(body))
		}
		call := svc.Projects.Locations.Deployments.Patch(name, body).UpdateMask(mask).Context(ctx)
		if s.flagIMRequestID != "" {
			call = call.RequestId(s.flagIMRequestID)
		}
		op, err := call.Do()
		if err != nil {
			return fmt.Errorf("updating deployment: %w", err)
		}
		return s.imFinishOp(ctx, svc, op, "Update deployment", args[0])
	}
	call := svc.Projects.Locations.Deployments.Create(parent, body).DeploymentId(args[0]).Context(ctx)
	if s.flagIMRequestID != "" {
		call = call.RequestId(s.flagIMRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}
	return s.imFinishOp(ctx, svc, op, "Create deployment", args[0])
}

func (s *infraManagerCmds) runIMDepDelete(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	call := svc.Projects.Locations.Deployments.Delete(name).Context(ctx)
	if s.flagIMRequestID != "" {
		call = call.RequestId(s.flagIMRequestID)
	}
	if s.flagIMDeletePolicy != "" {
		call = call.DeletePolicy(s.flagIMDeletePolicy)
	}
	if s.flagIMForceDelete {
		call = call.Force(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	return s.imFinishOp(ctx, svc, op, "Delete deployment", args[0])
}

func (s *infraManagerCmds) runIMDepDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	got, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing deployment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMDepList(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	var all []*config1.Deployment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Deployments.List(parent).Context(ctx)
		if s.flagIMFilter != "" {
			call = call.Filter(s.flagIMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing deployments: %w", err)
		}
		all = append(all, resp.Deployments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagIMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "STATE")
	for _, d := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(d.Name), d.State)
	}
	return nil
}

func (s *infraManagerCmds) runIMDepExportLock(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	got, err := svc.Projects.Locations.Deployments.ExportLock(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting lock: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMDepExportState(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	req := &config1.ExportDeploymentStatefileRequest{}
	got, err := svc.Projects.Locations.Deployments.ExportState(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting statefile: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMDepImportState(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	req := &config1.ImportStatefileRequest{}
	if err := loadYAMLOrJSONInto(s.flagIMExportPath, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	got, err := svc.Projects.Locations.Deployments.ImportState(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing statefile: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *infraManagerCmds) runIMDepLock(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	req := &config1.LockDeploymentRequest{}
	op, err := svc.Projects.Locations.Deployments.Lock(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("locking deployment: %w", err)
	}
	return s.imFinishOp(ctx, svc, op, "Lock deployment", args[0])
}

func (s *infraManagerCmds) runIMDepUnlock(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	req := &config1.UnlockDeploymentRequest{LockId: s.flagIMLockID}
	op, err := svc.Projects.Locations.Deployments.Unlock(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unlocking deployment: %w", err)
	}
	return s.imFinishOp(ctx, svc, op, "Unlock deployment", args[0])
}

// --- previews impl ---

func (s *infraManagerCmds) runIMPrevCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	body := &config1.Preview{}
	if err := loadYAMLOrJSONInto(s.flagIMConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Previews.Create(parent, body).PreviewId(args[0]).Context(ctx)
	if s.flagIMRequestID != "" {
		call = call.RequestId(s.flagIMRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating preview: %w", err)
	}
	return s.imFinishOp(ctx, svc, op, "Create preview", args[0])
}

func (s *infraManagerCmds) runIMPrevDelete(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "previews", args[0])
	call := svc.Projects.Locations.Previews.Delete(name).Context(ctx)
	if s.flagIMRequestID != "" {
		call = call.RequestId(s.flagIMRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting preview: %w", err)
	}
	return s.imFinishOp(ctx, svc, op, "Delete preview", args[0])
}

func (s *infraManagerCmds) runIMPrevDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "previews", args[0])
	got, err := svc.Projects.Locations.Previews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing preview: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMPrevList(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	var all []*config1.Preview
	pageToken := ""
	for {
		call := svc.Projects.Locations.Previews.List(parent).Context(ctx)
		if s.flagIMFilter != "" {
			call = call.Filter(s.flagIMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing previews: %w", err)
		}
		all = append(all, resp.Previews...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagIMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "STATE")
	for _, p := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

func (s *infraManagerCmds) runIMPrevExport(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := imChild(parent, "previews", args[0])
	req := &config1.ExportPreviewResultRequest{}
	got, err := svc.Projects.Locations.Previews.Export(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting preview: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

// --- resource-changes impl ---

func (s *infraManagerCmds) runIMRcDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	previewName := imChild(parent, "previews", s.flagIMPreview)
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/resourceChanges/%s", previewName, args[0])
	got, err := svc.Projects.Locations.Previews.ResourceChanges.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing resource change: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMRcList(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	previewName := imChild(parent, "previews", s.flagIMPreview)
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	var all []*config1.ResourceChange
	pageToken := ""
	for {
		call := svc.Projects.Locations.Previews.ResourceChanges.List(previewName).Context(ctx)
		if s.flagIMFilter != "" {
			call = call.Filter(s.flagIMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing resource changes: %w", err)
		}
		all = append(all, resp.ResourceChanges...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagIMFormat)
}

// --- resource-drifts impl ---

func (s *infraManagerCmds) runIMRdDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	previewName := imChild(parent, "previews", s.flagIMPreview)
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/resourceDrifts/%s", previewName, args[0])
	got, err := svc.Projects.Locations.Previews.ResourceDrifts.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing resource drift: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMRdList(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	previewName := imChild(parent, "previews", s.flagIMPreview)
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	var all []*config1.ResourceDrift
	pageToken := ""
	for {
		call := svc.Projects.Locations.Previews.ResourceDrifts.List(previewName).Context(ctx)
		if s.flagIMFilter != "" {
			call = call.Filter(s.flagIMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing resource drifts: %w", err)
		}
		all = append(all, resp.ResourceDrifts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagIMFormat)
}

// --- resources impl ---

func (s *infraManagerCmds) runIMResDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", s.flagIMDeployment)
	revName := fmt.Sprintf("%s/revisions/%s", depName, s.flagIMRevision)
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/resources/%s", revName, args[0])
	got, err := svc.Projects.Locations.Deployments.Revisions.Resources.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing resource: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMResList(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", s.flagIMDeployment)
	revName := fmt.Sprintf("%s/revisions/%s", depName, s.flagIMRevision)
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	var all []*config1.Resource
	pageToken := ""
	for {
		call := svc.Projects.Locations.Deployments.Revisions.Resources.List(revName).Context(ctx)
		if s.flagIMFilter != "" {
			call = call.Filter(s.flagIMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing resources: %w", err)
		}
		all = append(all, resp.Resources...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagIMFormat)
}

// --- revisions impl ---

func (s *infraManagerCmds) runIMRevDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", s.flagIMDeployment)
	name := fmt.Sprintf("%s/revisions/%s", depName, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Deployments.Revisions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing revision: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMRevList(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", s.flagIMDeployment)
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	var all []*config1.Revision
	pageToken := ""
	for {
		call := svc.Projects.Locations.Deployments.Revisions.List(depName).Context(ctx)
		if s.flagIMFilter != "" {
			call = call.Filter(s.flagIMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing revisions: %w", err)
		}
		all = append(all, resp.Revisions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagIMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "ACTION")
	for _, r := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(r.Name), r.Action)
	}
	return nil
}

func (s *infraManagerCmds) runIMRevExportState(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", s.flagIMDeployment)
	name := fmt.Sprintf("%s/revisions/%s", depName, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	req := &config1.ExportRevisionStatefileRequest{}
	got, err := svc.Projects.Locations.Deployments.Revisions.ExportState(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting revision statefile: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

// --- terraform-versions impl ---

func (s *infraManagerCmds) runIMTfDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	name := imChild(parent, "terraformVersions", args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TerraformVersions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Terraform version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIMFormat)
}

func (s *infraManagerCmds) runIMTfList(cmd *cobra.Command, args []string) error {
	parent, err := s.imResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().InfraManager(ctx)
	if err != nil {
		return err
	}
	var all []*config1.TerraformVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.TerraformVersions.List(parent).Context(ctx)
		if s.flagIMFilter != "" {
			call = call.Filter(s.flagIMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing Terraform versions: %w", err)
		}
		all = append(all, resp.TerraformVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagIMFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "STATE")
	for _, v := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(v.Name), v.State)
	}
	return nil
}

// infraManagerCmds holds the commands and flag values of one invocation's "infraManager"
// command group, so invocations don't share state (#2070).
type infraManagerCmds struct {
	inv *invocation

	infraManagerCmd        *cobra.Command
	flagIMLocation         string
	flagIMFormat           string
	flagIMFilter           string
	flagIMConfigFile       string
	flagIMUpdateMask       string
	flagIMRequestID        string
	flagIMAsync            bool
	flagIMDeployment       string
	flagIMRevision         string
	flagIMPreview          string
	flagIMLockID           int64
	flagIMDeletePolicy     string
	flagIMForceDelete      bool
	flagIMDraftID          string
	flagIMExportBucket     string
	flagIMExportPath       string
	imAutoMigrationCfgCmd  *cobra.Command
	imAmcDescribeCmd       *cobra.Command
	imAmcEnableCmd         *cobra.Command
	imAmcDisableCmd        *cobra.Command
	imDeploymentsCmd       *cobra.Command
	imDepApplyCmd          *cobra.Command
	imDepDeleteCmd         *cobra.Command
	imDepDescribeCmd       *cobra.Command
	imDepListCmd           *cobra.Command
	imDepExportLockCmd     *cobra.Command
	imDepExportStateCmd    *cobra.Command
	imDepImportStateCmd    *cobra.Command
	imDepLockCmd           *cobra.Command
	imDepUnlockCmd         *cobra.Command
	imPreviewsCmd          *cobra.Command
	imPrevCreateCmd        *cobra.Command
	imPrevDeleteCmd        *cobra.Command
	imPrevDescribeCmd      *cobra.Command
	imPrevListCmd          *cobra.Command
	imPrevExportCmd        *cobra.Command
	imResourceChangesCmd   *cobra.Command
	imRcDescribeCmd        *cobra.Command
	imRcListCmd            *cobra.Command
	imResourceDriftsCmd    *cobra.Command
	imRdDescribeCmd        *cobra.Command
	imRdListCmd            *cobra.Command
	imResourcesCmd         *cobra.Command
	imResDescribeCmd       *cobra.Command
	imResListCmd           *cobra.Command
	imRevisionsCmd         *cobra.Command
	imRevDescribeCmd       *cobra.Command
	imRevListCmd           *cobra.Command
	imRevExportStateCmd    *cobra.Command
	imTerraformVersionsCmd *cobra.Command
	imTfDescribeCmd        *cobra.Command
	imTfListCmd            *cobra.Command
}

// newInfraManagerCmd returns the "infraManager" command group for inv.
func newInfraManagerCmd(inv *invocation) *cobra.Command {
	return newInfraManagerCmds(inv).infraManagerCmd
}

// newInfraManagerCmds builds the "infraManager" command group for inv and returns its state.
func newInfraManagerCmds(inv *invocation) *infraManagerCmds {
	s := &infraManagerCmds{inv: inv}
	s.infraManagerCmd = &cobra.Command{Use: "infra-manager", Short: "Manage Infrastructure Manager"}
	s.imAutoMigrationCfgCmd = &cobra.Command{Use: "automigrationconfig", Short: "Manage auto migration config"}
	s.imAmcDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the auto migration config for a location",
		Args: cobra.NoArgs, RunE: s.runIMAmcDescribe,
	}
	s.imAmcEnableCmd = &cobra.Command{
		Use: "enable-auto-migration", Short: "Enable auto migration",
		Args: cobra.NoArgs, RunE: s.runIMAmcEnable,
	}
	s.imAmcDisableCmd = &cobra.Command{
		Use: "disable-auto-migration", Short: "Disable auto migration",
		Args: cobra.NoArgs, RunE: s.runIMAmcDisable,
	}
	s.imDeploymentsCmd = &cobra.Command{Use: "deployments", Short: "Manage deployments"}
	s.imDepApplyCmd = &cobra.Command{
		Use: "apply DEPLOYMENT", Short: "Create or update a deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runIMDepApply,
	}
	s.imDepDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runIMDepDelete,
	}
	s.imDepDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runIMDepDescribe,
	}
	s.imDepListCmd = &cobra.Command{
		Use: "list", Short: "List deployments in a location",
		Args: cobra.NoArgs, RunE: s.runIMDepList,
	}
	s.imDepExportLockCmd = &cobra.Command{
		Use: "export-lock DEPLOYMENT", Short: "Export the deployment lock information",
		Args: cobra.ExactArgs(1), RunE: s.runIMDepExportLock,
	}
	s.imDepExportStateCmd = &cobra.Command{
		Use: "export-statefile DEPLOYMENT", Short: "Export the deployment statefile",
		Args: cobra.ExactArgs(1), RunE: s.runIMDepExportState,
	}
	s.imDepImportStateCmd = &cobra.Command{
		Use: "import-statefile DEPLOYMENT", Short: "Import a statefile into a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runIMDepImportState,
	}
	s.imDepLockCmd = &cobra.Command{
		Use: "lock DEPLOYMENT", Short: "Lock a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runIMDepLock,
	}
	s.imDepUnlockCmd = &cobra.Command{
		Use: "unlock DEPLOYMENT", Short: "Unlock a deployment",
		Args: cobra.ExactArgs(1), RunE: s.runIMDepUnlock,
	}
	s.imPreviewsCmd = &cobra.Command{Use: "previews", Short: "Manage previews"}
	s.imPrevCreateCmd = &cobra.Command{
		Use: "create PREVIEW", Short: "Create a preview from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runIMPrevCreate,
	}
	s.imPrevDeleteCmd = &cobra.Command{
		Use: "delete PREVIEW", Short: "Delete a preview",
		Args: cobra.ExactArgs(1), RunE: s.runIMPrevDelete,
	}
	s.imPrevDescribeCmd = &cobra.Command{
		Use: "describe PREVIEW", Short: "Describe a preview",
		Args: cobra.ExactArgs(1), RunE: s.runIMPrevDescribe,
	}
	s.imPrevListCmd = &cobra.Command{
		Use: "list", Short: "List previews in a location",
		Args: cobra.NoArgs, RunE: s.runIMPrevList,
	}
	s.imPrevExportCmd = &cobra.Command{
		Use: "export PREVIEW", Short: "Export a preview result",
		Args: cobra.ExactArgs(1), RunE: s.runIMPrevExport,
	}
	s.imResourceChangesCmd = &cobra.Command{Use: "resource-changes", Short: "Manage preview resource changes"}
	s.imRcDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE_CHANGE", Short: "Describe a preview resource change",
		Args: cobra.ExactArgs(1), RunE: s.runIMRcDescribe,
	}
	s.imRcListCmd = &cobra.Command{
		Use: "list", Short: "List preview resource changes",
		Args: cobra.NoArgs, RunE: s.runIMRcList,
	}
	s.imResourceDriftsCmd = &cobra.Command{Use: "resource-drifts", Short: "Manage preview resource drifts"}
	s.imRdDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE_DRIFT", Short: "Describe a preview resource drift",
		Args: cobra.ExactArgs(1), RunE: s.runIMRdDescribe,
	}
	s.imRdListCmd = &cobra.Command{
		Use: "list", Short: "List preview resource drifts",
		Args: cobra.NoArgs, RunE: s.runIMRdList,
	}
	s.imResourcesCmd = &cobra.Command{Use: "resources", Short: "Manage revision resources"}
	s.imResDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a revision resource",
		Args: cobra.ExactArgs(1), RunE: s.runIMResDescribe,
	}
	s.imResListCmd = &cobra.Command{
		Use: "list", Short: "List revision resources",
		Args: cobra.NoArgs, RunE: s.runIMResList,
	}
	s.imRevisionsCmd = &cobra.Command{Use: "revisions", Short: "Manage deployment revisions"}
	s.imRevDescribeCmd = &cobra.Command{
		Use: "describe REVISION", Short: "Describe a deployment revision",
		Args: cobra.ExactArgs(1), RunE: s.runIMRevDescribe,
	}
	s.imRevListCmd = &cobra.Command{
		Use: "list", Short: "List deployment revisions",
		Args: cobra.NoArgs, RunE: s.runIMRevList,
	}
	s.imRevExportStateCmd = &cobra.Command{
		Use: "export-statefile REVISION", Short: "Export a revision statefile",
		Args: cobra.ExactArgs(1), RunE: s.runIMRevExportState,
	}
	s.imTerraformVersionsCmd = &cobra.Command{Use: "terraform-versions", Short: "Manage Terraform versions"}
	s.imTfDescribeCmd = &cobra.Command{
		Use: "describe VERSION", Short: "Describe a Terraform version",
		Args: cobra.ExactArgs(1), RunE: s.runIMTfDescribe,
	}
	s.imTfListCmd = &cobra.Command{
		Use: "list", Short: "List Terraform versions",
		Args: cobra.NoArgs, RunE: s.runIMTfList,
	}
	s.initInfraManager()
	return s
}
