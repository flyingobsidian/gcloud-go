package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	looker "google.golang.org/api/looker/v1"
)

// --- gcloud looker (#351) ---

var lookerCmd = &cobra.Command{Use: "looker", Short: "Manage Looker"}

func lkLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func lkChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func lkWaitOp(ctx context.Context, svc *looker.Service, op *looker.Operation) (*looker.Operation, error) {
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

func lkFinishOp(ctx context.Context, svc *looker.Service, op *looker.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := lkWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

var (
	flagLKRegion     string
	flagLKInstance   string
	flagLKConfigFile string
	flagLKUpdateMask string
	flagLKFormat     string
	flagLKAsync      bool
)

// --- regions ---

var lookerRegionsCmd = &cobra.Command{Use: "regions", Short: "Explore Looker regions"}

var lookerRegionsListCmd = &cobra.Command{
	Use: "list", Short: "List Looker regions", Args: cobra.NoArgs, RunE: runLKRegionsList,
}

// --- operations ---

var lookerOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Looker operations"}

var (
	lkOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: runLKOpCancel,
	}
	lkOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runLKOpDescribe,
	}
	lkOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a region",
		Args: cobra.NoArgs, RunE: runLKOpList,
	}
)

// --- instances ---

var lookerInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Looker instances"}

var (
	lkInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runLKInstCreate,
	}
	lkInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an instance",
		Args: cobra.ExactArgs(1), RunE: runLKInstDelete,
	}
	lkInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an instance",
		Args: cobra.ExactArgs(1), RunE: runLKInstDescribe,
	}
	lkInstListCmd = &cobra.Command{
		Use: "list", Short: "List instances in a region",
		Args: cobra.NoArgs, RunE: runLKInstList,
	}
	lkInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runLKInstUpdate,
	}
	lkInstRestartCmd = &cobra.Command{
		Use: "restart INSTANCE", Short: "Restart an instance",
		Args: cobra.ExactArgs(1), RunE: runLKInstRestart,
	}
	lkInstRestoreCmd = &cobra.Command{
		Use: "restore INSTANCE", Short: "Restore an instance from a backup (--config-file with RestoreInstanceRequest)",
		Args: cobra.ExactArgs(1), RunE: runLKInstRestore,
	}
	lkInstExportCmd = &cobra.Command{
		Use: "export INSTANCE", Short: "Export an instance's data (--config-file with ExportInstanceRequest)",
		Args: cobra.ExactArgs(1), RunE: runLKInstExport,
	}
	lkInstImportCmd = &cobra.Command{
		Use: "import INSTANCE", Short: "Import data into an instance (--config-file with ImportInstanceRequest)",
		Args: cobra.ExactArgs(1), RunE: runLKInstImport,
	}
)

// --- backups (nested under instances) ---

var lookerBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Looker instance backups"}

var (
	lkBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a backup for an instance",
		Args: cobra.ExactArgs(1), RunE: runLKBackupCreate,
	}
	lkBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup",
		Args: cobra.ExactArgs(1), RunE: runLKBackupDelete,
	}
	lkBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup",
		Args: cobra.ExactArgs(1), RunE: runLKBackupDescribe,
	}
	lkBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups for an instance",
		Args: cobra.NoArgs, RunE: runLKBackupList,
	}
)

func init() {
	// regions
	lookerRegionsListCmd.Flags().StringVar(&flagLKFormat, "format", "", "Output format")
	lookerRegionsCmd.AddCommand(lookerRegionsListCmd)
	lookerCmd.AddCommand(lookerRegionsCmd)

	// operations
	for _, c := range []*cobra.Command{lkOpCancelCmd, lkOpDescribeCmd, lkOpListCmd} {
		c.Flags().StringVar(&flagLKRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	lkOpDescribeCmd.Flags().StringVar(&flagLKFormat, "format", "", "Output format")
	lkOpListCmd.Flags().StringVar(&flagLKFormat, "format", "", "Output format")
	lookerOperationsCmd.AddCommand(lkOpCancelCmd, lkOpDescribeCmd, lkOpListCmd)
	lookerCmd.AddCommand(lookerOperationsCmd)

	// instances
	instAll := []*cobra.Command{lkInstCreateCmd, lkInstDeleteCmd, lkInstDescribeCmd, lkInstListCmd, lkInstUpdateCmd,
		lkInstRestartCmd, lkInstRestoreCmd, lkInstExportCmd, lkInstImportCmd}
	for _, c := range instAll {
		c.Flags().StringVar(&flagLKRegion, "region", "", "Region containing the instance (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{lkInstCreateCmd, lkInstUpdateCmd, lkInstRestoreCmd, lkInstExportCmd, lkInstImportCmd} {
		c.Flags().StringVar(&flagLKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	lkInstUpdateCmd.Flags().StringVar(&flagLKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{lkInstCreateCmd, lkInstDeleteCmd, lkInstUpdateCmd, lkInstRestartCmd, lkInstRestoreCmd, lkInstExportCmd, lkInstImportCmd} {
		c.Flags().BoolVar(&flagLKAsync, "async", false, "Return the long-running operation without waiting")
	}
	lkInstDescribeCmd.Flags().StringVar(&flagLKFormat, "format", "", "Output format")
	lkInstListCmd.Flags().StringVar(&flagLKFormat, "format", "", "Output format")
	lookerInstancesCmd.AddCommand(instAll...)
	lookerCmd.AddCommand(lookerInstancesCmd)

	// backups
	backupAll := []*cobra.Command{lkBackupCreateCmd, lkBackupDeleteCmd, lkBackupDescribeCmd, lkBackupListCmd}
	for _, c := range backupAll {
		c.Flags().StringVar(&flagLKRegion, "region", "", "Region containing the instance (required)")
		c.Flags().StringVar(&flagLKInstance, "instance", "", "Instance containing the backup (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("instance")
	}
	for _, c := range []*cobra.Command{lkBackupCreateCmd, lkBackupDeleteCmd} {
		c.Flags().BoolVar(&flagLKAsync, "async", false, "Return the long-running operation without waiting")
	}
	lkBackupDescribeCmd.Flags().StringVar(&flagLKFormat, "format", "", "Output format")
	lkBackupListCmd.Flags().StringVar(&flagLKFormat, "format", "", "Output format")
	lookerBackupsCmd.AddCommand(backupAll...)
	lookerCmd.AddCommand(lookerBackupsCmd)

	rootCmd.AddCommand(lookerCmd)
}

// --- regions impl ---

func runLKRegionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing regions: %w", err)
	}
	if flagLKFormat != "" {
		return emitFormatted(resp.Locations, flagLKFormat)
	}
	fmt.Printf("%-20s %s\n", "REGION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- operations impl ---

func lkOpName(id, project, region string) string {
	return lkChild("operations", id, lkLocationParent(project, region))
}

func runLKOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(lkOpName(args[0], project, flagLKRegion), &looker.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runLKOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(lkOpName(args[0], project, flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagLKFormat)
}

func runLKOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(lkLocationParent(project, flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagLKFormat != "" {
		return emitFormatted(resp.Operations, flagLKFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- instances impl ---

func lkInstName(id, project, region string) string {
	return lkChild("instances", id, lkLocationParent(project, region))
}

func runLKInstCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &looker.Instance{}
	if err := loadYAMLOrJSONInto(flagLKConfigFile, inst); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(lkLocationParent(project, flagLKRegion), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Create instance", args[0], flagLKAsync)
}

func runLKInstDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(lkInstName(args[0], project, flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Delete instance", args[0], flagLKAsync)
}

func runLKInstDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Get(lkInstName(args[0], project, flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagLKFormat)
}

func runLKInstList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.List(lkLocationParent(project, flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if flagLKFormat != "" {
		return emitFormatted(resp.Instances, flagLKFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, i := range resp.Instances {
		fmt.Printf("%-40s %s\n", path.Base(i.Name), i.State)
	}
	return nil
}

func runLKInstUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &looker.Instance{}
	if err := loadYAMLOrJSONInto(flagLKConfigFile, inst); err != nil {
		return err
	}
	mask := flagLKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Patch(lkInstName(args[0], project, flagLKRegion), inst).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Update instance", args[0], flagLKAsync)
}

func runLKInstRestart(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Restart(lkInstName(args[0], project, flagLKRegion), &looker.RestartInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting instance: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Restart instance", args[0], flagLKAsync)
}

func runLKInstRestore(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &looker.RestoreInstanceRequest{}
	if err := loadYAMLOrJSONInto(flagLKConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Restore(lkInstName(args[0], project, flagLKRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring instance: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Restore instance", args[0], flagLKAsync)
}

func runLKInstExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &looker.ExportInstanceRequest{}
	if err := loadYAMLOrJSONInto(flagLKConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Export(lkInstName(args[0], project, flagLKRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting instance: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Export instance", args[0], flagLKAsync)
}

func runLKInstImport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &looker.ImportInstanceRequest{}
	if err := loadYAMLOrJSONInto(flagLKConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Import(lkInstName(args[0], project, flagLKRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing instance: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Import instance", args[0], flagLKAsync)
}

// --- backups impl ---

func lkBackupParent(project, region, instance string) string {
	return lkInstName(instance, project, region)
}

func lkBackupName(id, project, region, instance string) string {
	return lkChild("backups", id, lkBackupParent(project, region, instance))
}

func runLKBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// InstanceBackup.Create doesn't take an ID; the API assigns one. The
	// positional BACKUP arg is accepted for CLI symmetry.
	op, err := svc.Projects.Locations.Instances.Backups.Create(lkBackupParent(project, flagLKRegion, flagLKInstance), &looker.InstanceBackup{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Create backup", args[0], flagLKAsync)
}

func runLKBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Backups.Delete(lkBackupName(args[0], project, flagLKRegion, flagLKInstance)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return lkFinishOp(ctx, svc, op, "Delete backup", args[0], flagLKAsync)
}

func runLKBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Backups.Get(lkBackupName(args[0], project, flagLKRegion, flagLKInstance)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagLKFormat)
}

func runLKBackupList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LookerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.Backups.List(lkBackupParent(project, flagLKRegion, flagLKInstance)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if flagLKFormat != "" {
		return emitFormatted(resp.InstanceBackups, flagLKFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, b := range resp.InstanceBackups {
		fmt.Printf("%-40s %s\n", path.Base(b.Name), b.State)
	}
	return nil
}
