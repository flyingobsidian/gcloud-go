package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	securityposture "google.golang.org/api/securityposture/v1"
)

// --- gcloud scc postures / posture-deployments / posture-templates /
//     posture-operations / iac-validation-reports (#812-#817) ---
//
// Backed by the Security Posture v1 API (securityposture.googleapis.com),
// which only exposes organization-scoped endpoints.

func (g *sccCmds) postureParent() (string, error) {
	if g.flagPostureOrg == "" {
		return "", fmt.Errorf("--organization is required")
	}
	loc := g.flagPostureLocation
	if loc == "" {
		loc = "global"
	}
	org := strings.TrimPrefix(g.flagPostureOrg, "organizations/")
	return fmt.Sprintf("organizations/%s/locations/%s", org, loc), nil
}

func postureQualify(parent, collection, id string) string {
	if strings.HasPrefix(id, "organizations/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func postureWaitOp(ctx context.Context, svc *securityposture.Service, op *securityposture.Operation) (*securityposture.Operation, error) {
	for !op.Done {
		got, err := svc.Organizations.Locations.Operations.Get(op.Name).Context(ctx).Do()
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

func (g *sccCmds) postureFinishOp(ctx context.Context, svc *securityposture.Service, op *securityposture.Operation, verb, name string) error {
	if g.flagPostureAsync {
		fmt.Fprintf(g.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(g.inv.stdout, op, "")
	}
	final, err := postureWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(g.inv.stdout, final.Response, "")
	}
	return nil
}

func (g *sccCmds) postureAddCommonFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagPostureOrg, "organization", "",
			"Organization ID or fully-qualified organizations/{id} (required)")
		c.Flags().StringVar(&g.flagPostureLocation, "location", "global", "Location (defaults to global)")
		_ = c.MarkFlagRequired("organization")
	}
}

// --- postures ---

// --- posture-deployments ---

// --- posture-templates ---

// --- posture-operations ---

// --- iac-validation-reports ---

func (g *sccCmds) initSccPosture() {
	// postures
	postAll := []*cobra.Command{g.postureCreateCmd, g.postureDeleteCmd, g.postureDescribeCmd,
		g.postureExtractCmd, g.postureListCmd, g.postureListRevisionsCmd, g.postureUpdateCmd}
	g.postureAddCommonFlags(postAll...)
	for _, c := range []*cobra.Command{g.postureDescribeCmd, g.postureListCmd, g.postureListRevisionsCmd} {
		c.Flags().StringVar(&g.flagPostureFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.postureListCmd, g.postureListRevisionsCmd} {
		c.Flags().Int64Var(&g.flagPosturePageSize, "page-size", 0, "Page size for list requests")
	}
	g.postureDescribeCmd.Flags().StringVar(&g.flagPostureRevision, "revision-id", "", "Optional revision ID")
	g.postureCreateCmd.Flags().StringVar(&g.flagPostureConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Posture body (required)")
	_ = g.postureCreateCmd.MarkFlagRequired("config-file")
	g.postureCreateCmd.Flags().BoolVar(&g.flagPostureAsync, "async", false, "Return the LRO without waiting")
	g.postureDeleteCmd.Flags().StringVar(&g.flagPostureRevision, "etag", "", "Optional etag for conditional delete")
	g.postureDeleteCmd.Flags().BoolVar(&g.flagPostureAsync, "async", false, "Return the LRO without waiting")
	g.postureUpdateCmd.Flags().StringVar(&g.flagPostureConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Posture body (required)")
	_ = g.postureUpdateCmd.MarkFlagRequired("config-file")
	g.postureUpdateCmd.Flags().StringVar(&g.flagPostureUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.postureUpdateCmd.Flags().StringVar(&g.flagPostureRevision, "revision-id", "", "Optional revision ID to update")
	g.postureUpdateCmd.Flags().BoolVar(&g.flagPostureAsync, "async", false, "Return the LRO without waiting")
	g.postureExtractCmd.Flags().StringVar(&g.flagPostureExtractWL, "workload", "",
		"Workload resource to extract from (required)")
	_ = g.postureExtractCmd.MarkFlagRequired("workload")
	g.postureExtractCmd.Flags().BoolVar(&g.flagPostureAsync, "async", false, "Return the LRO without waiting")
	g.sccPosturesCmd.AddCommand(postAll...)
	g.sccCmd.AddCommand(g.sccPosturesCmd)

	// posture-deployments
	pdAll := []*cobra.Command{g.pdCreateCmd, g.pdDeleteCmd, g.pdDescribeCmd, g.pdListCmd, g.pdUpdateCmd}
	g.postureAddCommonFlags(pdAll...)
	for _, c := range []*cobra.Command{g.pdDescribeCmd, g.pdListCmd} {
		c.Flags().StringVar(&g.flagPostureFormat, "format", "", "Output format")
	}
	g.pdListCmd.Flags().Int64Var(&g.flagPosturePageSize, "page-size", 0, "Page size for list requests")
	g.pdCreateCmd.Flags().StringVar(&g.flagPostureConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the PostureDeployment body (required)")
	_ = g.pdCreateCmd.MarkFlagRequired("config-file")
	g.pdCreateCmd.Flags().BoolVar(&g.flagPostureAsync, "async", false, "Return the LRO without waiting")
	g.pdDeleteCmd.Flags().BoolVar(&g.flagPostureAsync, "async", false, "Return the LRO without waiting")
	g.pdUpdateCmd.Flags().StringVar(&g.flagPostureConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the PostureDeployment body (required)")
	_ = g.pdUpdateCmd.MarkFlagRequired("config-file")
	g.pdUpdateCmd.Flags().StringVar(&g.flagPostureUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.pdUpdateCmd.Flags().BoolVar(&g.flagPostureAsync, "async", false, "Return the LRO without waiting")
	g.sccPostureDeploymentsCmd.AddCommand(pdAll...)
	g.sccCmd.AddCommand(g.sccPostureDeploymentsCmd)

	// posture-templates
	ptAll := []*cobra.Command{g.ptDescribeCmd, g.ptListCmd}
	g.postureAddCommonFlags(ptAll...)
	for _, c := range ptAll {
		c.Flags().StringVar(&g.flagPostureFormat, "format", "", "Output format")
	}
	g.ptListCmd.Flags().Int64Var(&g.flagPosturePageSize, "page-size", 0, "Page size for list requests")
	g.ptListCmd.Flags().StringVar(&g.flagPostureFilter, "filter", "", "Server-side list filter")
	g.ptDescribeCmd.Flags().StringVar(&g.flagPostureRevision, "revision-id", "", "Optional revision ID")
	g.sccPostureTemplatesCmd.AddCommand(ptAll...)
	g.sccCmd.AddCommand(g.sccPostureTemplatesCmd)

	// posture-operations
	poAll := []*cobra.Command{g.poCancelCmd, g.poDeleteCmd, g.poDescribeCmd, g.poListCmd}
	g.postureAddCommonFlags(poAll...)
	for _, c := range []*cobra.Command{g.poDescribeCmd, g.poListCmd} {
		c.Flags().StringVar(&g.flagPostureFormat, "format", "", "Output format")
	}
	g.poListCmd.Flags().StringVar(&g.flagPostureFilter, "filter", "", "Server-side list filter")
	g.poListCmd.Flags().Int64Var(&g.flagPosturePageSize, "page-size", 0, "Page size for list requests")
	g.sccPostureOperationsCmd.AddCommand(poAll...)
	g.sccCmd.AddCommand(g.sccPostureOperationsCmd)

	// iac-validation-reports
	iacAll := []*cobra.Command{g.iacCreateCmd, g.iacDescribeCmd, g.iacListCmd}
	g.postureAddCommonFlags(iacAll...)
	for _, c := range []*cobra.Command{g.iacDescribeCmd, g.iacListCmd} {
		c.Flags().StringVar(&g.flagPostureFormat, "format", "", "Output format")
	}
	g.iacListCmd.Flags().StringVar(&g.flagPostureFilter, "filter", "", "Server-side list filter")
	g.iacListCmd.Flags().Int64Var(&g.flagPosturePageSize, "page-size", 0, "Page size for list requests")
	g.iacCreateCmd.Flags().StringVar(&g.flagIaCConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the CreateIaCValidationReportRequest body (required)")
	_ = g.iacCreateCmd.MarkFlagRequired("config-file")
	g.sccIaCReportsCmd.AddCommand(iacAll...)
	g.sccCmd.AddCommand(g.sccIaCReportsCmd)
}

// --- postures impl ---

func (g *sccCmds) runPostureCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	body := &securityposture.Posture{}
	if err := loadYAMLOrJSONInto(g.flagPostureConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.Postures.Create(parent, body).PostureId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating posture: %w", err)
	}
	return g.postureFinishOp(ctx, svc, op, "Create posture", args[0])
}

func (g *sccCmds) runPostureDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postures", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.Delete(name).Context(ctx)
	if g.flagPostureRevision != "" {
		call = call.Etag(g.flagPostureRevision)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting posture: %w", err)
	}
	return g.postureFinishOp(ctx, svc, op, "Delete posture", args[0])
}

func (g *sccCmds) runPostureDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postures", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.Get(name).Context(ctx)
	if g.flagPostureRevision != "" {
		call = call.RevisionId(g.flagPostureRevision)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing posture: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPostureFormat)
}

func (g *sccCmds) runPostureExtract(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	req := &securityposture.ExtractPostureRequest{
		PostureId: args[0],
		Workload:  g.flagPostureExtractWL,
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.Postures.Extract(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("extracting posture: %w", err)
	}
	return g.postureFinishOp(ctx, svc, op, "Extract posture", args[0])
}

func (g *sccCmds) runPostureList(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.List(parent).Context(ctx)
	if g.flagPosturePageSize > 0 {
		call = call.PageSize(g.flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing postures: %w", err)
	}
	if g.flagPostureFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Postures, g.flagPostureFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Postures), "%-40s %s\n", "NAME", "STATE")
	for _, p := range resp.Postures {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

func (g *sccCmds) runPostureListRevisions(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postures", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.ListRevisions(name).Context(ctx)
	if g.flagPosturePageSize > 0 {
		call = call.PageSize(g.flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing posture revisions: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagPostureFormat)
}

func (g *sccCmds) runPostureUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postures", args[0])
	body := &securityposture.Posture{}
	if err := loadYAMLOrJSONInto(g.flagPostureConfigFile, body); err != nil {
		return err
	}
	mask := g.flagPostureUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.Patch(name, body).UpdateMask(mask).Context(ctx)
	if g.flagPostureRevision != "" {
		call = call.RevisionId(g.flagPostureRevision)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating posture: %w", err)
	}
	return g.postureFinishOp(ctx, svc, op, "Update posture", args[0])
}

// --- posture-deployments impl ---

func (g *sccCmds) runPDCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	body := &securityposture.PostureDeployment{}
	if err := loadYAMLOrJSONInto(g.flagPostureConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.PostureDeployments.Create(parent, body).PostureDeploymentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating posture deployment: %w", err)
	}
	return g.postureFinishOp(ctx, svc, op, "Create posture deployment", args[0])
}

func (g *sccCmds) runPDDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postureDeployments", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.PostureDeployments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting posture deployment: %w", err)
	}
	return g.postureFinishOp(ctx, svc, op, "Delete posture deployment", args[0])
}

func (g *sccCmds) runPDDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postureDeployments", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Locations.PostureDeployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing posture deployment: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPostureFormat)
}

func (g *sccCmds) runPDList(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.PostureDeployments.List(parent).Context(ctx)
	if g.flagPosturePageSize > 0 {
		call = call.PageSize(g.flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing posture deployments: %w", err)
	}
	if g.flagPostureFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.PostureDeployments, g.flagPostureFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.PostureDeployments), "%-40s %s\n", "NAME", "STATE")
	for _, d := range resp.PostureDeployments {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(d.Name), d.State)
	}
	return nil
}

func (g *sccCmds) runPDUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postureDeployments", args[0])
	body := &securityposture.PostureDeployment{}
	if err := loadYAMLOrJSONInto(g.flagPostureConfigFile, body); err != nil {
		return err
	}
	mask := g.flagPostureUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.PostureDeployments.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating posture deployment: %w", err)
	}
	return g.postureFinishOp(ctx, svc, op, "Update posture deployment", args[0])
}

// --- posture-templates impl ---

func (g *sccCmds) runPTDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postureTemplates", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.PostureTemplates.Get(name).Context(ctx)
	if g.flagPostureRevision != "" {
		call = call.RevisionId(g.flagPostureRevision)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing posture template: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPostureFormat)
}

func (g *sccCmds) runPTList(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.PostureTemplates.List(parent).Context(ctx)
	if g.flagPostureFilter != "" {
		call = call.Filter(g.flagPostureFilter)
	}
	if g.flagPosturePageSize > 0 {
		call = call.PageSize(g.flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing posture templates: %w", err)
	}
	if g.flagPostureFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.PostureTemplates, g.flagPostureFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.PostureTemplates), "%-40s %s\n", "NAME", "DESCRIPTION")
	for _, t := range resp.PostureTemplates {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(t.Name), t.Description)
	}
	return nil
}

// --- posture-operations impl ---

func (g *sccCmds) runPOCancel(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "operations", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Locations.Operations.Cancel(name, &securityposture.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling posture operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Cancel request issued for posture operation [%s].\n", args[0])
	return nil
}

func (g *sccCmds) runPODelete(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "operations", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting posture operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted posture operation [%s].\n", args[0])
	return nil
}

func (g *sccCmds) runPODescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "operations", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing posture operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPostureFormat)
}

func (g *sccCmds) runPOList(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := parent + "/operations"
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Operations.List(name).Context(ctx)
	if g.flagPostureFilter != "" {
		call = call.Filter(g.flagPostureFilter)
	}
	if g.flagPosturePageSize > 0 {
		call = call.PageSize(g.flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing posture operations: %w", err)
	}
	if g.flagPostureFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Operations, g.flagPostureFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(g.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- iac-validation-reports impl ---

func (g *sccCmds) runIaCCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	body := &securityposture.CreateIaCValidationReportRequest{}
	if err := loadYAMLOrJSONInto(g.flagIaCConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Locations.Reports.CreateIaCValidationReport(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating IaC validation report: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPostureFormat)
}

func (g *sccCmds) runIaCDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "reports", args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Locations.Reports.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing IaC validation report: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPostureFormat)
}

func (g *sccCmds) runIaCList(cmd *cobra.Command, args []string) error {
	parent, err := g.postureParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecurityPosture(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Reports.List(parent).Context(ctx)
	if g.flagPostureFilter != "" {
		call = call.Filter(g.flagPostureFilter)
	}
	if g.flagPosturePageSize > 0 {
		call = call.PageSize(g.flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing IaC validation reports: %w", err)
	}
	if g.flagPostureFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Reports, g.flagPostureFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Reports), "%-40s %s\n", "NAME", "CREATE_TIME")
	for _, r := range resp.Reports {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(r.Name), r.CreateTime)
	}
	return nil
}
