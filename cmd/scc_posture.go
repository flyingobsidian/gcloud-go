package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	securityposture "google.golang.org/api/securityposture/v1"
)

// --- gcloud scc postures / posture-deployments / posture-templates /
//     posture-operations / iac-validation-reports (#812-#817) ---
//
// Backed by the Security Posture v1 API (securityposture.googleapis.com),
// which only exposes organization-scoped endpoints.

var (
	flagPostureOrg        string
	flagPostureLocation   string
	flagPostureFormat     string
	flagPostureFilter     string
	flagPostureConfigFile string
	flagPostureUpdateMask string
	flagPostureAsync      bool
	flagPostureRevision   string
	flagPostureExtractID  string
	flagPostureExtractWL  string
	flagPosturePageSize   int64

	// iac-validation-reports create
	flagIaCConfigFile string
)

func postureParent() (string, error) {
	if flagPostureOrg == "" {
		return "", fmt.Errorf("--organization is required")
	}
	loc := flagPostureLocation
	if loc == "" {
		loc = "global"
	}
	org := strings.TrimPrefix(flagPostureOrg, "organizations/")
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

func postureFinishOp(ctx context.Context, svc *securityposture.Service, op *securityposture.Operation, verb, name string) error {
	if flagPostureAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := postureWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

func postureAddCommonFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagPostureOrg, "organization", "",
			"Organization ID or fully-qualified organizations/{id} (required)")
		c.Flags().StringVar(&flagPostureLocation, "location", "global", "Location (defaults to global)")
		_ = c.MarkFlagRequired("organization")
	}
}

// --- postures ---

var sccPosturesCmd = &cobra.Command{Use: "postures", Short: "Manage Security Command Center postures"}

var (
	postureCreateCmd = &cobra.Command{
		Use: "create POSTURE", Short: "Create a posture from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPostureCreate,
	}
	postureDeleteCmd = &cobra.Command{
		Use: "delete POSTURE", Short: "Delete a posture",
		Args: cobra.ExactArgs(1), RunE: runPostureDelete,
	}
	postureDescribeCmd = &cobra.Command{
		Use: "describe POSTURE", Short: "Describe a posture",
		Args: cobra.ExactArgs(1), RunE: runPostureDescribe,
	}
	postureExtractCmd = &cobra.Command{
		Use: "extract POSTURE_ID", Short: "Extract a posture from a workload",
		Args: cobra.ExactArgs(1), RunE: runPostureExtract,
	}
	postureListCmd = &cobra.Command{
		Use: "list", Short: "List postures",
		Args: cobra.NoArgs, RunE: runPostureList,
	}
	postureListRevisionsCmd = &cobra.Command{
		Use: "list-revisions POSTURE", Short: "List revisions of a posture",
		Args: cobra.ExactArgs(1), RunE: runPostureListRevisions,
	}
	postureUpdateCmd = &cobra.Command{
		Use: "update POSTURE", Short: "Update a posture from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPostureUpdate,
	}
)

// --- posture-deployments ---

var sccPostureDeploymentsCmd = &cobra.Command{Use: "posture-deployments", Short: "Manage posture deployments"}

var (
	pdCreateCmd = &cobra.Command{
		Use: "create DEPLOYMENT", Short: "Create a posture deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPDCreate,
	}
	pdDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a posture deployment",
		Args: cobra.ExactArgs(1), RunE: runPDDelete,
	}
	pdDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a posture deployment",
		Args: cobra.ExactArgs(1), RunE: runPDDescribe,
	}
	pdListCmd = &cobra.Command{
		Use: "list", Short: "List posture deployments",
		Args: cobra.NoArgs, RunE: runPDList,
	}
	pdUpdateCmd = &cobra.Command{
		Use: "update DEPLOYMENT", Short: "Update a posture deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPDUpdate,
	}
)

// --- posture-templates ---

var sccPostureTemplatesCmd = &cobra.Command{Use: "posture-templates", Short: "Manage posture templates"}

var (
	ptDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a posture template",
		Args: cobra.ExactArgs(1), RunE: runPTDescribe,
	}
	ptListCmd = &cobra.Command{
		Use: "list", Short: "List posture templates",
		Args: cobra.NoArgs, RunE: runPTList,
	}
)

// --- posture-operations ---

var sccPostureOperationsCmd = &cobra.Command{Use: "posture-operations", Short: "Manage posture operations"}

var (
	poCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a posture operation",
		Args: cobra.ExactArgs(1), RunE: runPOCancel,
	}
	poDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a posture operation",
		Args: cobra.ExactArgs(1), RunE: runPODelete,
	}
	poDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a posture operation",
		Args: cobra.ExactArgs(1), RunE: runPODescribe,
	}
	poListCmd = &cobra.Command{
		Use: "list", Short: "List posture operations",
		Args: cobra.NoArgs, RunE: runPOList,
	}
)

// --- iac-validation-reports ---

var sccIaCReportsCmd = &cobra.Command{Use: "iac-validation-reports", Short: "Manage IaC validation reports"}

var (
	iacCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an IaC validation report from a --config-file",
		Args: cobra.NoArgs, RunE: runIaCCreate,
	}
	iacDescribeCmd = &cobra.Command{
		Use: "describe REPORT", Short: "Describe an IaC validation report",
		Args: cobra.ExactArgs(1), RunE: runIaCDescribe,
	}
	iacListCmd = &cobra.Command{
		Use: "list", Short: "List IaC validation reports",
		Args: cobra.NoArgs, RunE: runIaCList,
	}
)

func init() {
	// postures
	postAll := []*cobra.Command{postureCreateCmd, postureDeleteCmd, postureDescribeCmd,
		postureExtractCmd, postureListCmd, postureListRevisionsCmd, postureUpdateCmd}
	postureAddCommonFlags(postAll...)
	for _, c := range []*cobra.Command{postureDescribeCmd, postureListCmd, postureListRevisionsCmd} {
		c.Flags().StringVar(&flagPostureFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{postureListCmd, postureListRevisionsCmd} {
		c.Flags().Int64Var(&flagPosturePageSize, "page-size", 0, "Page size for list requests")
	}
	postureDescribeCmd.Flags().StringVar(&flagPostureRevision, "revision-id", "", "Optional revision ID")
	postureCreateCmd.Flags().StringVar(&flagPostureConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Posture body (required)")
	_ = postureCreateCmd.MarkFlagRequired("config-file")
	postureCreateCmd.Flags().BoolVar(&flagPostureAsync, "async", false, "Return the LRO without waiting")
	postureDeleteCmd.Flags().StringVar(&flagPostureRevision, "etag", "", "Optional etag for conditional delete")
	postureDeleteCmd.Flags().BoolVar(&flagPostureAsync, "async", false, "Return the LRO without waiting")
	postureUpdateCmd.Flags().StringVar(&flagPostureConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Posture body (required)")
	_ = postureUpdateCmd.MarkFlagRequired("config-file")
	postureUpdateCmd.Flags().StringVar(&flagPostureUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	postureUpdateCmd.Flags().StringVar(&flagPostureRevision, "revision-id", "", "Optional revision ID to update")
	postureUpdateCmd.Flags().BoolVar(&flagPostureAsync, "async", false, "Return the LRO without waiting")
	postureExtractCmd.Flags().StringVar(&flagPostureExtractWL, "workload", "",
		"Workload resource to extract from (required)")
	_ = postureExtractCmd.MarkFlagRequired("workload")
	postureExtractCmd.Flags().BoolVar(&flagPostureAsync, "async", false, "Return the LRO without waiting")
	sccPosturesCmd.AddCommand(postAll...)
	sccCmd.AddCommand(sccPosturesCmd)

	// posture-deployments
	pdAll := []*cobra.Command{pdCreateCmd, pdDeleteCmd, pdDescribeCmd, pdListCmd, pdUpdateCmd}
	postureAddCommonFlags(pdAll...)
	for _, c := range []*cobra.Command{pdDescribeCmd, pdListCmd} {
		c.Flags().StringVar(&flagPostureFormat, "format", "", "Output format")
	}
	pdListCmd.Flags().Int64Var(&flagPosturePageSize, "page-size", 0, "Page size for list requests")
	pdCreateCmd.Flags().StringVar(&flagPostureConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the PostureDeployment body (required)")
	_ = pdCreateCmd.MarkFlagRequired("config-file")
	pdCreateCmd.Flags().BoolVar(&flagPostureAsync, "async", false, "Return the LRO without waiting")
	pdDeleteCmd.Flags().BoolVar(&flagPostureAsync, "async", false, "Return the LRO without waiting")
	pdUpdateCmd.Flags().StringVar(&flagPostureConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the PostureDeployment body (required)")
	_ = pdUpdateCmd.MarkFlagRequired("config-file")
	pdUpdateCmd.Flags().StringVar(&flagPostureUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	pdUpdateCmd.Flags().BoolVar(&flagPostureAsync, "async", false, "Return the LRO without waiting")
	sccPostureDeploymentsCmd.AddCommand(pdAll...)
	sccCmd.AddCommand(sccPostureDeploymentsCmd)

	// posture-templates
	ptAll := []*cobra.Command{ptDescribeCmd, ptListCmd}
	postureAddCommonFlags(ptAll...)
	for _, c := range ptAll {
		c.Flags().StringVar(&flagPostureFormat, "format", "", "Output format")
	}
	ptListCmd.Flags().Int64Var(&flagPosturePageSize, "page-size", 0, "Page size for list requests")
	ptListCmd.Flags().StringVar(&flagPostureFilter, "filter", "", "Server-side list filter")
	ptDescribeCmd.Flags().StringVar(&flagPostureRevision, "revision-id", "", "Optional revision ID")
	sccPostureTemplatesCmd.AddCommand(ptAll...)
	sccCmd.AddCommand(sccPostureTemplatesCmd)

	// posture-operations
	poAll := []*cobra.Command{poCancelCmd, poDeleteCmd, poDescribeCmd, poListCmd}
	postureAddCommonFlags(poAll...)
	for _, c := range []*cobra.Command{poDescribeCmd, poListCmd} {
		c.Flags().StringVar(&flagPostureFormat, "format", "", "Output format")
	}
	poListCmd.Flags().StringVar(&flagPostureFilter, "filter", "", "Server-side list filter")
	poListCmd.Flags().Int64Var(&flagPosturePageSize, "page-size", 0, "Page size for list requests")
	sccPostureOperationsCmd.AddCommand(poAll...)
	sccCmd.AddCommand(sccPostureOperationsCmd)

	// iac-validation-reports
	iacAll := []*cobra.Command{iacCreateCmd, iacDescribeCmd, iacListCmd}
	postureAddCommonFlags(iacAll...)
	for _, c := range []*cobra.Command{iacDescribeCmd, iacListCmd} {
		c.Flags().StringVar(&flagPostureFormat, "format", "", "Output format")
	}
	iacListCmd.Flags().StringVar(&flagPostureFilter, "filter", "", "Server-side list filter")
	iacListCmd.Flags().Int64Var(&flagPosturePageSize, "page-size", 0, "Page size for list requests")
	iacCreateCmd.Flags().StringVar(&flagIaCConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the CreateIaCValidationReportRequest body (required)")
	_ = iacCreateCmd.MarkFlagRequired("config-file")
	sccIaCReportsCmd.AddCommand(iacAll...)
	sccCmd.AddCommand(sccIaCReportsCmd)
}

// --- postures impl ---

func runPostureCreate(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	body := &securityposture.Posture{}
	if err := loadYAMLOrJSONInto(flagPostureConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.Postures.Create(parent, body).PostureId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating posture: %w", err)
	}
	return postureFinishOp(ctx, svc, op, "Create posture", args[0])
}

func runPostureDelete(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postures", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.Delete(name).Context(ctx)
	if flagPostureRevision != "" {
		call = call.Etag(flagPostureRevision)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting posture: %w", err)
	}
	return postureFinishOp(ctx, svc, op, "Delete posture", args[0])
}

func runPostureDescribe(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postures", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.Get(name).Context(ctx)
	if flagPostureRevision != "" {
		call = call.RevisionId(flagPostureRevision)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing posture: %w", err)
	}
	return emitFormatted(got, flagPostureFormat)
}

func runPostureExtract(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	req := &securityposture.ExtractPostureRequest{
		PostureId: args[0],
		Workload:  flagPostureExtractWL,
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.Postures.Extract(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("extracting posture: %w", err)
	}
	return postureFinishOp(ctx, svc, op, "Extract posture", args[0])
}

func runPostureList(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.List(parent).Context(ctx)
	if flagPosturePageSize > 0 {
		call = call.PageSize(flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing postures: %w", err)
	}
	if flagPostureFormat != "" {
		return emitFormatted(resp.Postures, flagPostureFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, p := range resp.Postures {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

func runPostureListRevisions(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postures", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.ListRevisions(name).Context(ctx)
	if flagPosturePageSize > 0 {
		call = call.PageSize(flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing posture revisions: %w", err)
	}
	return emitFormatted(resp, flagPostureFormat)
}

func runPostureUpdate(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postures", args[0])
	body := &securityposture.Posture{}
	if err := loadYAMLOrJSONInto(flagPostureConfigFile, body); err != nil {
		return err
	}
	mask := flagPostureUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Postures.Patch(name, body).UpdateMask(mask).Context(ctx)
	if flagPostureRevision != "" {
		call = call.RevisionId(flagPostureRevision)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating posture: %w", err)
	}
	return postureFinishOp(ctx, svc, op, "Update posture", args[0])
}

// --- posture-deployments impl ---

func runPDCreate(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	body := &securityposture.PostureDeployment{}
	if err := loadYAMLOrJSONInto(flagPostureConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.PostureDeployments.Create(parent, body).PostureDeploymentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating posture deployment: %w", err)
	}
	return postureFinishOp(ctx, svc, op, "Create posture deployment", args[0])
}

func runPDDelete(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postureDeployments", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.PostureDeployments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting posture deployment: %w", err)
	}
	return postureFinishOp(ctx, svc, op, "Delete posture deployment", args[0])
}

func runPDDescribe(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postureDeployments", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Locations.PostureDeployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing posture deployment: %w", err)
	}
	return emitFormatted(got, flagPostureFormat)
}

func runPDList(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.PostureDeployments.List(parent).Context(ctx)
	if flagPosturePageSize > 0 {
		call = call.PageSize(flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing posture deployments: %w", err)
	}
	if flagPostureFormat != "" {
		return emitFormatted(resp.PostureDeployments, flagPostureFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, d := range resp.PostureDeployments {
		fmt.Printf("%-40s %s\n", path.Base(d.Name), d.State)
	}
	return nil
}

func runPDUpdate(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postureDeployments", args[0])
	body := &securityposture.PostureDeployment{}
	if err := loadYAMLOrJSONInto(flagPostureConfigFile, body); err != nil {
		return err
	}
	mask := flagPostureUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.PostureDeployments.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating posture deployment: %w", err)
	}
	return postureFinishOp(ctx, svc, op, "Update posture deployment", args[0])
}

// --- posture-templates impl ---

func runPTDescribe(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "postureTemplates", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.PostureTemplates.Get(name).Context(ctx)
	if flagPostureRevision != "" {
		call = call.RevisionId(flagPostureRevision)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing posture template: %w", err)
	}
	return emitFormatted(got, flagPostureFormat)
}

func runPTList(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.PostureTemplates.List(parent).Context(ctx)
	if flagPostureFilter != "" {
		call = call.Filter(flagPostureFilter)
	}
	if flagPosturePageSize > 0 {
		call = call.PageSize(flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing posture templates: %w", err)
	}
	if flagPostureFormat != "" {
		return emitFormatted(resp.PostureTemplates, flagPostureFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DESCRIPTION")
	for _, t := range resp.PostureTemplates {
		fmt.Printf("%-40s %s\n", path.Base(t.Name), t.Description)
	}
	return nil
}

// --- posture-operations impl ---

func runPOCancel(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "operations", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Locations.Operations.Cancel(name, &securityposture.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling posture operation: %w", err)
	}
	fmt.Printf("Cancel request issued for posture operation [%s].\n", args[0])
	return nil
}

func runPODelete(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "operations", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting posture operation: %w", err)
	}
	fmt.Printf("Deleted posture operation [%s].\n", args[0])
	return nil
}

func runPODescribe(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "operations", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing posture operation: %w", err)
	}
	return emitFormatted(got, flagPostureFormat)
}

func runPOList(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := parent + "/operations"
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Operations.List(name).Context(ctx)
	if flagPostureFilter != "" {
		call = call.Filter(flagPostureFilter)
	}
	if flagPosturePageSize > 0 {
		call = call.PageSize(flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing posture operations: %w", err)
	}
	if flagPostureFormat != "" {
		return emitFormatted(resp.Operations, flagPostureFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- iac-validation-reports impl ---

func runIaCCreate(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	body := &securityposture.CreateIaCValidationReportRequest{}
	if err := loadYAMLOrJSONInto(flagIaCConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Locations.Reports.CreateIaCValidationReport(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating IaC validation report: %w", err)
	}
	return emitFormatted(got, flagPostureFormat)
}

func runIaCDescribe(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	name := postureQualify(parent, "reports", args[0])
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Locations.Reports.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing IaC validation report: %w", err)
	}
	return emitFormatted(got, flagPostureFormat)
}

func runIaCList(cmd *cobra.Command, args []string) error {
	parent, err := postureParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecurityPostureService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Reports.List(parent).Context(ctx)
	if flagPostureFilter != "" {
		call = call.Filter(flagPostureFilter)
	}
	if flagPosturePageSize > 0 {
		call = call.PageSize(flagPosturePageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing IaC validation reports: %w", err)
	}
	if flagPostureFormat != "" {
		return emitFormatted(resp.Reports, flagPostureFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "CREATE_TIME")
	for _, r := range resp.Reports {
		fmt.Printf("%-40s %s\n", path.Base(r.Name), r.CreateTime)
	}
	return nil
}
