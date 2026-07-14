package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	config1 "google.golang.org/api/config/v1"
)

// --- gcloud infra-manager (#348, #795-#802) ---

var infraManagerCmd = &cobra.Command{Use: "infra-manager", Short: "Manage Infrastructure Manager"}

var (
	flagIMLocation    string
	flagIMFormat      string
	flagIMFilter      string
	flagIMConfigFile  string
	flagIMUpdateMask  string
	flagIMRequestID   string
	flagIMAsync       bool
	flagIMDeployment  string
	flagIMRevision    string
	flagIMPreview     string
	flagIMLockID      int64
	flagIMDeletePolicy string
	flagIMForceDelete  bool
	flagIMDraftID     string
	flagIMExportBucket string
	flagIMExportPath   string
)

func imParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func imChild(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func imResolveParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagIMLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return imParent(project, flagIMLocation), nil
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

func imFinishOp(ctx context.Context, svc *config1.Service, op *config1.Operation, verb, name string) error {
	if flagIMAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := imWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// --- automigrationconfig ---

var imAutoMigrationCfgCmd = &cobra.Command{Use: "automigrationconfig", Short: "Manage auto migration config"}

var (
	imAmcDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the auto migration config for a location",
		Args: cobra.NoArgs, RunE: runIMAmcDescribe,
	}
	imAmcEnableCmd = &cobra.Command{
		Use: "enable-auto-migration", Short: "Enable auto migration",
		Args: cobra.NoArgs, RunE: runIMAmcEnable,
	}
	imAmcDisableCmd = &cobra.Command{
		Use: "disable-auto-migration", Short: "Disable auto migration",
		Args: cobra.NoArgs, RunE: runIMAmcDisable,
	}
)

// --- deployments ---

var imDeploymentsCmd = &cobra.Command{Use: "deployments", Short: "Manage deployments"}

var (
	imDepApplyCmd = &cobra.Command{
		Use: "apply DEPLOYMENT", Short: "Create or update a deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runIMDepApply,
	}
	imDepDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a deployment",
		Args: cobra.ExactArgs(1), RunE: runIMDepDelete,
	}
	imDepDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a deployment",
		Args: cobra.ExactArgs(1), RunE: runIMDepDescribe,
	}
	imDepListCmd = &cobra.Command{
		Use: "list", Short: "List deployments in a location",
		Args: cobra.NoArgs, RunE: runIMDepList,
	}
	imDepExportLockCmd = &cobra.Command{
		Use: "export-lock DEPLOYMENT", Short: "Export the deployment lock information",
		Args: cobra.ExactArgs(1), RunE: runIMDepExportLock,
	}
	imDepExportStateCmd = &cobra.Command{
		Use: "export-statefile DEPLOYMENT", Short: "Export the deployment statefile",
		Args: cobra.ExactArgs(1), RunE: runIMDepExportState,
	}
	imDepImportStateCmd = &cobra.Command{
		Use: "import-statefile DEPLOYMENT", Short: "Import a statefile into a deployment",
		Args: cobra.ExactArgs(1), RunE: runIMDepImportState,
	}
	imDepLockCmd = &cobra.Command{
		Use: "lock DEPLOYMENT", Short: "Lock a deployment",
		Args: cobra.ExactArgs(1), RunE: runIMDepLock,
	}
	imDepUnlockCmd = &cobra.Command{
		Use: "unlock DEPLOYMENT", Short: "Unlock a deployment",
		Args: cobra.ExactArgs(1), RunE: runIMDepUnlock,
	}
)

// --- previews ---

var imPreviewsCmd = &cobra.Command{Use: "previews", Short: "Manage previews"}

var (
	imPrevCreateCmd = &cobra.Command{
		Use: "create PREVIEW", Short: "Create a preview from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runIMPrevCreate,
	}
	imPrevDeleteCmd = &cobra.Command{
		Use: "delete PREVIEW", Short: "Delete a preview",
		Args: cobra.ExactArgs(1), RunE: runIMPrevDelete,
	}
	imPrevDescribeCmd = &cobra.Command{
		Use: "describe PREVIEW", Short: "Describe a preview",
		Args: cobra.ExactArgs(1), RunE: runIMPrevDescribe,
	}
	imPrevListCmd = &cobra.Command{
		Use: "list", Short: "List previews in a location",
		Args: cobra.NoArgs, RunE: runIMPrevList,
	}
	imPrevExportCmd = &cobra.Command{
		Use: "export PREVIEW", Short: "Export a preview result",
		Args: cobra.ExactArgs(1), RunE: runIMPrevExport,
	}
)

// --- resource-changes ---

var imResourceChangesCmd = &cobra.Command{Use: "resource-changes", Short: "Manage preview resource changes"}

var (
	imRcDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE_CHANGE", Short: "Describe a preview resource change",
		Args: cobra.ExactArgs(1), RunE: runIMRcDescribe,
	}
	imRcListCmd = &cobra.Command{
		Use: "list", Short: "List preview resource changes",
		Args: cobra.NoArgs, RunE: runIMRcList,
	}
)

// --- resource-drifts ---

var imResourceDriftsCmd = &cobra.Command{Use: "resource-drifts", Short: "Manage preview resource drifts"}

var (
	imRdDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE_DRIFT", Short: "Describe a preview resource drift",
		Args: cobra.ExactArgs(1), RunE: runIMRdDescribe,
	}
	imRdListCmd = &cobra.Command{
		Use: "list", Short: "List preview resource drifts",
		Args: cobra.NoArgs, RunE: runIMRdList,
	}
)

// --- resources ---

var imResourcesCmd = &cobra.Command{Use: "resources", Short: "Manage revision resources"}

var (
	imResDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a revision resource",
		Args: cobra.ExactArgs(1), RunE: runIMResDescribe,
	}
	imResListCmd = &cobra.Command{
		Use: "list", Short: "List revision resources",
		Args: cobra.NoArgs, RunE: runIMResList,
	}
)

// --- revisions ---

var imRevisionsCmd = &cobra.Command{Use: "revisions", Short: "Manage deployment revisions"}

var (
	imRevDescribeCmd = &cobra.Command{
		Use: "describe REVISION", Short: "Describe a deployment revision",
		Args: cobra.ExactArgs(1), RunE: runIMRevDescribe,
	}
	imRevListCmd = &cobra.Command{
		Use: "list", Short: "List deployment revisions",
		Args: cobra.NoArgs, RunE: runIMRevList,
	}
	imRevExportStateCmd = &cobra.Command{
		Use: "export-statefile REVISION", Short: "Export a revision statefile",
		Args: cobra.ExactArgs(1), RunE: runIMRevExportState,
	}
)

// --- terraform-versions ---

var imTerraformVersionsCmd = &cobra.Command{Use: "terraform-versions", Short: "Manage Terraform versions"}

var (
	imTfDescribeCmd = &cobra.Command{
		Use: "describe VERSION", Short: "Describe a Terraform version",
		Args: cobra.ExactArgs(1), RunE: runIMTfDescribe,
	}
	imTfListCmd = &cobra.Command{
		Use: "list", Short: "List Terraform versions",
		Args: cobra.NoArgs, RunE: runIMTfList,
	}
)

func init() {
	addLoc := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagIMLocation, "location", "", "Location (required)")
		}
	}
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagIMFormat, "format", "", "Output format")
		}
	}
	addFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagIMFilter, "filter", "", "Server-side list filter")
		}
	}
	addAsync := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().BoolVar(&flagIMAsync, "async", false, "Do not wait for the operation to complete")
			c.Flags().StringVar(&flagIMRequestID, "request-id", "", "Optional idempotency request ID")
		}
	}

	// automigrationconfig
	addLoc(imAmcDescribeCmd, imAmcEnableCmd, imAmcDisableCmd)
	addFmt(imAmcDescribeCmd)
	imAutoMigrationCfgCmd.AddCommand(imAmcDescribeCmd, imAmcEnableCmd, imAmcDisableCmd)
	infraManagerCmd.AddCommand(imAutoMigrationCfgCmd)

	// deployments
	addLoc(imDepApplyCmd, imDepDeleteCmd, imDepDescribeCmd, imDepListCmd, imDepExportLockCmd,
		imDepExportStateCmd, imDepImportStateCmd, imDepLockCmd, imDepUnlockCmd)
	addFmt(imDepDescribeCmd, imDepListCmd, imDepExportLockCmd)
	addFilter(imDepListCmd)
	addAsync(imDepApplyCmd, imDepDeleteCmd, imDepImportStateCmd, imDepLockCmd, imDepUnlockCmd)
	imDepApplyCmd.Flags().StringVar(&flagIMConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Deployment body (required)")
	_ = imDepApplyCmd.MarkFlagRequired("config-file")
	imDepApplyCmd.Flags().StringVar(&flagIMUpdateMask, "update-mask", "",
		"For update, comma-separated list of fields to update")
	imDepDeleteCmd.Flags().StringVar(&flagIMDeletePolicy, "delete-policy", "",
		"Deletion policy (DELETE, ABANDON)")
	imDepDeleteCmd.Flags().BoolVar(&flagIMForceDelete, "force", false, "Force deletion")
	imDepImportStateCmd.Flags().StringVar(&flagIMExportPath, "statefile", "",
		"Path to a JSON/YAML file containing the statefile body (required)")
	_ = imDepImportStateCmd.MarkFlagRequired("statefile")
	imDepLockCmd.Flags().StringVar(&flagIMDraftID, "draft-id", "", "Optional draft ID for the lock request")
	imDepUnlockCmd.Flags().Int64Var(&flagIMLockID, "lock-id", 0, "Lock ID to release (required)")
	_ = imDepUnlockCmd.MarkFlagRequired("lock-id")
	imDepExportStateCmd.Flags().StringVar(&flagIMExportBucket, "draft-id", "", "Optional draft ID for the export request")
	imDeploymentsCmd.AddCommand(imDepApplyCmd, imDepDeleteCmd, imDepDescribeCmd, imDepListCmd,
		imDepExportLockCmd, imDepExportStateCmd, imDepImportStateCmd, imDepLockCmd, imDepUnlockCmd)
	infraManagerCmd.AddCommand(imDeploymentsCmd)

	// previews
	addLoc(imPrevCreateCmd, imPrevDeleteCmd, imPrevDescribeCmd, imPrevListCmd, imPrevExportCmd)
	addFmt(imPrevDescribeCmd, imPrevListCmd, imPrevExportCmd)
	addFilter(imPrevListCmd)
	addAsync(imPrevCreateCmd, imPrevDeleteCmd)
	imPrevCreateCmd.Flags().StringVar(&flagIMConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Preview body (required)")
	_ = imPrevCreateCmd.MarkFlagRequired("config-file")
	imPreviewsCmd.AddCommand(imPrevCreateCmd, imPrevDeleteCmd, imPrevDescribeCmd, imPrevListCmd, imPrevExportCmd)
	infraManagerCmd.AddCommand(imPreviewsCmd)

	// resource-changes
	addLoc(imRcDescribeCmd, imRcListCmd)
	addFmt(imRcDescribeCmd, imRcListCmd)
	addFilter(imRcListCmd)
	imRcDescribeCmd.Flags().StringVar(&flagIMPreview, "preview", "", "Owning preview (required)")
	_ = imRcDescribeCmd.MarkFlagRequired("preview")
	imRcListCmd.Flags().StringVar(&flagIMPreview, "preview", "", "Owning preview (required)")
	_ = imRcListCmd.MarkFlagRequired("preview")
	imResourceChangesCmd.AddCommand(imRcDescribeCmd, imRcListCmd)
	infraManagerCmd.AddCommand(imResourceChangesCmd)

	// resource-drifts
	addLoc(imRdDescribeCmd, imRdListCmd)
	addFmt(imRdDescribeCmd, imRdListCmd)
	addFilter(imRdListCmd)
	imRdDescribeCmd.Flags().StringVar(&flagIMPreview, "preview", "", "Owning preview (required)")
	_ = imRdDescribeCmd.MarkFlagRequired("preview")
	imRdListCmd.Flags().StringVar(&flagIMPreview, "preview", "", "Owning preview (required)")
	_ = imRdListCmd.MarkFlagRequired("preview")
	imResourceDriftsCmd.AddCommand(imRdDescribeCmd, imRdListCmd)
	infraManagerCmd.AddCommand(imResourceDriftsCmd)

	// resources
	addLoc(imResDescribeCmd, imResListCmd)
	addFmt(imResDescribeCmd, imResListCmd)
	addFilter(imResListCmd)
	imResDescribeCmd.Flags().StringVar(&flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = imResDescribeCmd.MarkFlagRequired("deployment")
	imResDescribeCmd.Flags().StringVar(&flagIMRevision, "revision", "", "Owning revision (required)")
	_ = imResDescribeCmd.MarkFlagRequired("revision")
	imResListCmd.Flags().StringVar(&flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = imResListCmd.MarkFlagRequired("deployment")
	imResListCmd.Flags().StringVar(&flagIMRevision, "revision", "", "Owning revision (required)")
	_ = imResListCmd.MarkFlagRequired("revision")
	imResourcesCmd.AddCommand(imResDescribeCmd, imResListCmd)
	infraManagerCmd.AddCommand(imResourcesCmd)

	// revisions
	addLoc(imRevDescribeCmd, imRevListCmd, imRevExportStateCmd)
	addFmt(imRevDescribeCmd, imRevListCmd, imRevExportStateCmd)
	addFilter(imRevListCmd)
	imRevDescribeCmd.Flags().StringVar(&flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = imRevDescribeCmd.MarkFlagRequired("deployment")
	imRevListCmd.Flags().StringVar(&flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = imRevListCmd.MarkFlagRequired("deployment")
	imRevExportStateCmd.Flags().StringVar(&flagIMDeployment, "deployment", "", "Owning deployment (required)")
	_ = imRevExportStateCmd.MarkFlagRequired("deployment")
	imRevisionsCmd.AddCommand(imRevDescribeCmd, imRevListCmd, imRevExportStateCmd)
	infraManagerCmd.AddCommand(imRevisionsCmd)

	// terraform-versions
	addLoc(imTfDescribeCmd, imTfListCmd)
	addFmt(imTfDescribeCmd, imTfListCmd)
	addFilter(imTfListCmd)
	imTerraformVersionsCmd.AddCommand(imTfDescribeCmd, imTfListCmd)
	infraManagerCmd.AddCommand(imTerraformVersionsCmd)

	rootCmd.AddCommand(infraManagerCmd)
}

// --- automigrationconfig impl ---

func imAmcName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/autoMigrationConfig", project, location)
}

func runIMAmcDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagIMLocation == "" {
		return fmt.Errorf("--location is required")
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imAmcName(project, flagIMLocation)
	got, err := svc.Projects.Locations.GetAutoMigrationConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing auto migration config: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMAmcEnable(cmd *cobra.Command, args []string) error {
	return imAmcUpdateEnabled(true)
}

func runIMAmcDisable(cmd *cobra.Command, args []string) error {
	return imAmcUpdateEnabled(false)
}

func imAmcUpdateEnabled(enabled bool) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagIMLocation == "" {
		return fmt.Errorf("--location is required")
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imAmcName(project, flagIMLocation)
	body := &config1.AutoMigrationConfig{AutoMigrationEnabled: enabled, ForceSendFields: []string{"AutoMigrationEnabled"}}
	got, err := svc.Projects.Locations.UpdateAutoMigrationConfig(name, body).UpdateMask("autoMigrationEnabled").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating auto migration config: %w", err)
	}
	return emitFormatted(got, "")
}

// --- deployments impl ---

func runIMDepApply(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	body := &config1.Deployment{}
	if err := loadYAMLOrJSONInto(flagIMConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	existing, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do()
	if err == nil && existing != nil {
		mask := flagIMUpdateMask
		if mask == "" {
			mask = joinMask(nonEmptyJSONFields(body))
		}
		call := svc.Projects.Locations.Deployments.Patch(name, body).UpdateMask(mask).Context(ctx)
		if flagIMRequestID != "" {
			call = call.RequestId(flagIMRequestID)
		}
		op, err := call.Do()
		if err != nil {
			return fmt.Errorf("updating deployment: %w", err)
		}
		return imFinishOp(ctx, svc, op, "Update deployment", args[0])
	}
	call := svc.Projects.Locations.Deployments.Create(parent, body).DeploymentId(args[0]).Context(ctx)
	if flagIMRequestID != "" {
		call = call.RequestId(flagIMRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}
	return imFinishOp(ctx, svc, op, "Create deployment", args[0])
}

func runIMDepDelete(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	call := svc.Projects.Locations.Deployments.Delete(name).Context(ctx)
	if flagIMRequestID != "" {
		call = call.RequestId(flagIMRequestID)
	}
	if flagIMDeletePolicy != "" {
		call = call.DeletePolicy(flagIMDeletePolicy)
	}
	if flagIMForceDelete {
		call = call.Force(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	return imFinishOp(ctx, svc, op, "Delete deployment", args[0])
}

func runIMDepDescribe(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	got, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing deployment: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMDepList(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*config1.Deployment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Deployments.List(parent).Context(ctx)
		if flagIMFilter != "" {
			call = call.Filter(flagIMFilter)
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
	if flagIMFormat != "" {
		return emitFormatted(all, flagIMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, d := range all {
		fmt.Printf("%-40s %s\n", path.Base(d.Name), d.State)
	}
	return nil
}

func runIMDepExportLock(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	got, err := svc.Projects.Locations.Deployments.ExportLock(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting lock: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMDepExportState(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	req := &config1.ExportDeploymentStatefileRequest{}
	got, err := svc.Projects.Locations.Deployments.ExportState(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting statefile: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMDepImportState(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	req := &config1.ImportStatefileRequest{}
	if err := loadYAMLOrJSONInto(flagIMExportPath, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	got, err := svc.Projects.Locations.Deployments.ImportState(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing statefile: %w", err)
	}
	return emitFormatted(got, "")
}

func runIMDepLock(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	req := &config1.LockDeploymentRequest{}
	op, err := svc.Projects.Locations.Deployments.Lock(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("locking deployment: %w", err)
	}
	return imFinishOp(ctx, svc, op, "Lock deployment", args[0])
}

func runIMDepUnlock(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "deployments", args[0])
	req := &config1.UnlockDeploymentRequest{LockId: flagIMLockID}
	op, err := svc.Projects.Locations.Deployments.Unlock(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unlocking deployment: %w", err)
	}
	return imFinishOp(ctx, svc, op, "Unlock deployment", args[0])
}

// --- previews impl ---

func runIMPrevCreate(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	body := &config1.Preview{}
	if err := loadYAMLOrJSONInto(flagIMConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Previews.Create(parent, body).PreviewId(args[0]).Context(ctx)
	if flagIMRequestID != "" {
		call = call.RequestId(flagIMRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating preview: %w", err)
	}
	return imFinishOp(ctx, svc, op, "Create preview", args[0])
}

func runIMPrevDelete(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "previews", args[0])
	call := svc.Projects.Locations.Previews.Delete(name).Context(ctx)
	if flagIMRequestID != "" {
		call = call.RequestId(flagIMRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting preview: %w", err)
	}
	return imFinishOp(ctx, svc, op, "Delete preview", args[0])
}

func runIMPrevDescribe(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "previews", args[0])
	got, err := svc.Projects.Locations.Previews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing preview: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMPrevList(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*config1.Preview
	pageToken := ""
	for {
		call := svc.Projects.Locations.Previews.List(parent).Context(ctx)
		if flagIMFilter != "" {
			call = call.Filter(flagIMFilter)
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
	if flagIMFormat != "" {
		return emitFormatted(all, flagIMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, p := range all {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

func runIMPrevExport(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := imChild(parent, "previews", args[0])
	req := &config1.ExportPreviewResultRequest{}
	got, err := svc.Projects.Locations.Previews.Export(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting preview: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

// --- resource-changes impl ---

func runIMRcDescribe(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	previewName := imChild(parent, "previews", flagIMPreview)
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/resourceChanges/%s", previewName, args[0])
	got, err := svc.Projects.Locations.Previews.ResourceChanges.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing resource change: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMRcList(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	previewName := imChild(parent, "previews", flagIMPreview)
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*config1.ResourceChange
	pageToken := ""
	for {
		call := svc.Projects.Locations.Previews.ResourceChanges.List(previewName).Context(ctx)
		if flagIMFilter != "" {
			call = call.Filter(flagIMFilter)
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
	return emitFormatted(all, flagIMFormat)
}

// --- resource-drifts impl ---

func runIMRdDescribe(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	previewName := imChild(parent, "previews", flagIMPreview)
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/resourceDrifts/%s", previewName, args[0])
	got, err := svc.Projects.Locations.Previews.ResourceDrifts.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing resource drift: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMRdList(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	previewName := imChild(parent, "previews", flagIMPreview)
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*config1.ResourceDrift
	pageToken := ""
	for {
		call := svc.Projects.Locations.Previews.ResourceDrifts.List(previewName).Context(ctx)
		if flagIMFilter != "" {
			call = call.Filter(flagIMFilter)
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
	return emitFormatted(all, flagIMFormat)
}

// --- resources impl ---

func runIMResDescribe(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", flagIMDeployment)
	revName := fmt.Sprintf("%s/revisions/%s", depName, flagIMRevision)
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/resources/%s", revName, args[0])
	got, err := svc.Projects.Locations.Deployments.Revisions.Resources.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing resource: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMResList(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", flagIMDeployment)
	revName := fmt.Sprintf("%s/revisions/%s", depName, flagIMRevision)
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*config1.Resource
	pageToken := ""
	for {
		call := svc.Projects.Locations.Deployments.Revisions.Resources.List(revName).Context(ctx)
		if flagIMFilter != "" {
			call = call.Filter(flagIMFilter)
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
	return emitFormatted(all, flagIMFormat)
}

// --- revisions impl ---

func runIMRevDescribe(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", flagIMDeployment)
	name := fmt.Sprintf("%s/revisions/%s", depName, args[0])
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Deployments.Revisions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing revision: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMRevList(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", flagIMDeployment)
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*config1.Revision
	pageToken := ""
	for {
		call := svc.Projects.Locations.Deployments.Revisions.List(depName).Context(ctx)
		if flagIMFilter != "" {
			call = call.Filter(flagIMFilter)
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
	if flagIMFormat != "" {
		return emitFormatted(all, flagIMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "ACTION")
	for _, r := range all {
		fmt.Printf("%-40s %s\n", path.Base(r.Name), r.Action)
	}
	return nil
}

func runIMRevExportState(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	depName := imChild(parent, "deployments", flagIMDeployment)
	name := fmt.Sprintf("%s/revisions/%s", depName, args[0])
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &config1.ExportRevisionStatefileRequest{}
	got, err := svc.Projects.Locations.Deployments.Revisions.ExportState(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting revision statefile: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

// --- terraform-versions impl ---

func runIMTfDescribe(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	name := imChild(parent, "terraformVersions", args[0])
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TerraformVersions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Terraform version: %w", err)
	}
	return emitFormatted(got, flagIMFormat)
}

func runIMTfList(cmd *cobra.Command, args []string) error {
	parent, err := imResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.InfraManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*config1.TerraformVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.TerraformVersions.List(parent).Context(ctx)
		if flagIMFilter != "" {
			call = call.Filter(flagIMFilter)
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
	if flagIMFormat != "" {
		return emitFormatted(all, flagIMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, v := range all {
		fmt.Printf("%-40s %s\n", path.Base(v.Name), v.State)
	}
	return nil
}
