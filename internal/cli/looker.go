package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	looker "google.golang.org/api/looker/v1"
)

// --- gcloud looker (#351) ---

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

func (s *lookerCmds) lkFinishOp(ctx context.Context, svc *looker.Service, op *looker.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := lkWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// --- regions ---

// --- operations ---

// --- instances ---

// --- backups (nested under instances) ---

func (s *lookerCmds) initLooker() {
	// regions
	s.lookerRegionsListCmd.Flags().StringVar(&s.flagLKFormat, "format", "", "Output format")
	s.lookerRegionsCmd.AddCommand(s.lookerRegionsListCmd)
	s.lookerCmd.AddCommand(s.lookerRegionsCmd)

	// operations
	for _, c := range []*cobra.Command{s.lkOpCancelCmd, s.lkOpDescribeCmd, s.lkOpListCmd} {
		c.Flags().StringVar(&s.flagLKRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	s.lkOpDescribeCmd.Flags().StringVar(&s.flagLKFormat, "format", "", "Output format")
	s.lkOpListCmd.Flags().StringVar(&s.flagLKFormat, "format", "", "Output format")
	s.lookerOperationsCmd.AddCommand(s.lkOpCancelCmd, s.lkOpDescribeCmd, s.lkOpListCmd)
	s.lookerCmd.AddCommand(s.lookerOperationsCmd)

	// instances
	instAll := []*cobra.Command{s.lkInstCreateCmd, s.lkInstDeleteCmd, s.lkInstDescribeCmd, s.lkInstListCmd, s.lkInstUpdateCmd,
		s.lkInstRestartCmd, s.lkInstRestoreCmd, s.lkInstExportCmd, s.lkInstImportCmd}
	for _, c := range instAll {
		c.Flags().StringVar(&s.flagLKRegion, "region", "", "Region containing the instance (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.lkInstCreateCmd, s.lkInstUpdateCmd, s.lkInstRestoreCmd, s.lkInstExportCmd, s.lkInstImportCmd} {
		c.Flags().StringVar(&s.flagLKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.lkInstUpdateCmd.Flags().StringVar(&s.flagLKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.lkInstCreateCmd, s.lkInstDeleteCmd, s.lkInstUpdateCmd, s.lkInstRestartCmd, s.lkInstRestoreCmd, s.lkInstExportCmd, s.lkInstImportCmd} {
		c.Flags().BoolVar(&s.flagLKAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.lkInstDescribeCmd.Flags().StringVar(&s.flagLKFormat, "format", "", "Output format")
	s.lkInstListCmd.Flags().StringVar(&s.flagLKFormat, "format", "", "Output format")
	s.lookerInstancesCmd.AddCommand(instAll...)
	s.lookerCmd.AddCommand(s.lookerInstancesCmd)

	// backups
	backupAll := []*cobra.Command{s.lkBackupCreateCmd, s.lkBackupDeleteCmd, s.lkBackupDescribeCmd, s.lkBackupListCmd}
	for _, c := range backupAll {
		c.Flags().StringVar(&s.flagLKRegion, "region", "", "Region containing the instance (required)")
		c.Flags().StringVar(&s.flagLKInstance, "instance", "", "Instance containing the backup (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("instance")
	}
	for _, c := range []*cobra.Command{s.lkBackupCreateCmd, s.lkBackupDeleteCmd} {
		c.Flags().BoolVar(&s.flagLKAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.lkBackupDescribeCmd.Flags().StringVar(&s.flagLKFormat, "format", "", "Output format")
	s.lkBackupListCmd.Flags().StringVar(&s.flagLKFormat, "format", "", "Output format")
	s.lookerBackupsCmd.AddCommand(backupAll...)
	s.lookerCmd.AddCommand(s.lookerBackupsCmd)

}

// --- regions impl ---

func (s *lookerCmds) runLKRegionsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing regions: %w", err)
	}
	if s.flagLKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Locations, s.flagLKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Locations), "%-20s %s\n", "REGION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(s.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- operations impl ---

func lkOpName(id, project, region string) string {
	return lkChild("operations", id, lkLocationParent(project, region))
}

func (s *lookerCmds) runLKOpCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(lkOpName(args[0], project, s.flagLKRegion), &looker.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *lookerCmds) runLKOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(lkOpName(args[0], project, s.flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagLKFormat)
}

func (s *lookerCmds) runLKOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(lkLocationParent(project, s.flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagLKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagLKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- instances impl ---

func lkInstName(id, project, region string) string {
	return lkChild("instances", id, lkLocationParent(project, region))
}

func (s *lookerCmds) runLKInstCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	inst := &looker.Instance{}
	if err := loadYAMLOrJSONInto(s.flagLKConfigFile, inst); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(lkLocationParent(project, s.flagLKRegion), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Create instance", args[0], s.flagLKAsync)
}

func (s *lookerCmds) runLKInstDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(lkInstName(args[0], project, s.flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Delete instance", args[0], s.flagLKAsync)
}

func (s *lookerCmds) runLKInstDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Get(lkInstName(args[0], project, s.flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagLKFormat)
}

func (s *lookerCmds) runLKInstList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.List(lkLocationParent(project, s.flagLKRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if s.flagLKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Instances, s.flagLKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Instances), "%-40s %s\n", "NAME", "STATE")
	for _, i := range resp.Instances {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(i.Name), i.State)
	}
	return nil
}

func (s *lookerCmds) runLKInstUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	inst := &looker.Instance{}
	if err := loadYAMLOrJSONInto(s.flagLKConfigFile, inst); err != nil {
		return err
	}
	mask := s.flagLKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Patch(lkInstName(args[0], project, s.flagLKRegion), inst).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Update instance", args[0], s.flagLKAsync)
}

func (s *lookerCmds) runLKInstRestart(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Restart(lkInstName(args[0], project, s.flagLKRegion), &looker.RestartInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting instance: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Restart instance", args[0], s.flagLKAsync)
}

func (s *lookerCmds) runLKInstRestore(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &looker.RestoreInstanceRequest{}
	if err := loadYAMLOrJSONInto(s.flagLKConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Restore(lkInstName(args[0], project, s.flagLKRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring instance: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Restore instance", args[0], s.flagLKAsync)
}

func (s *lookerCmds) runLKInstExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &looker.ExportInstanceRequest{}
	if err := loadYAMLOrJSONInto(s.flagLKConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Export(lkInstName(args[0], project, s.flagLKRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting instance: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Export instance", args[0], s.flagLKAsync)
}

func (s *lookerCmds) runLKInstImport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &looker.ImportInstanceRequest{}
	if err := loadYAMLOrJSONInto(s.flagLKConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Import(lkInstName(args[0], project, s.flagLKRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing instance: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Import instance", args[0], s.flagLKAsync)
}

// --- backups impl ---

func lkBackupParent(project, region, instance string) string {
	return lkInstName(instance, project, region)
}

func lkBackupName(id, project, region, instance string) string {
	return lkChild("backups", id, lkBackupParent(project, region, instance))
}

func (s *lookerCmds) runLKBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	// InstanceBackup.Create doesn't take an ID; the API assigns one. The
	// positional BACKUP arg is accepted for CLI symmetry.
	op, err := svc.Projects.Locations.Instances.Backups.Create(lkBackupParent(project, s.flagLKRegion, s.flagLKInstance), &looker.InstanceBackup{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Create backup", args[0], s.flagLKAsync)
}

func (s *lookerCmds) runLKBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Backups.Delete(lkBackupName(args[0], project, s.flagLKRegion, s.flagLKInstance)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return s.lkFinishOp(ctx, svc, op, "Delete backup", args[0], s.flagLKAsync)
}

func (s *lookerCmds) runLKBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Backups.Get(lkBackupName(args[0], project, s.flagLKRegion, s.flagLKInstance)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagLKFormat)
}

func (s *lookerCmds) runLKBackupList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Looker(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.Backups.List(lkBackupParent(project, s.flagLKRegion, s.flagLKInstance)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if s.flagLKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.InstanceBackups, s.flagLKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.InstanceBackups), "%-40s %s\n", "NAME", "STATE")
	for _, b := range resp.InstanceBackups {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(b.Name), b.State)
	}
	return nil
}

// lookerCmds holds the commands and flag values of one invocation's "looker"
// command group, so invocations don't share state (#2070).
type lookerCmds struct {
	inv *invocation

	lookerCmd            *cobra.Command
	flagLKRegion         string
	flagLKInstance       string
	flagLKConfigFile     string
	flagLKUpdateMask     string
	flagLKFormat         string
	flagLKAsync          bool
	lookerRegionsCmd     *cobra.Command
	lookerRegionsListCmd *cobra.Command
	lookerOperationsCmd  *cobra.Command
	lkOpCancelCmd        *cobra.Command
	lkOpDescribeCmd      *cobra.Command
	lkOpListCmd          *cobra.Command
	lookerInstancesCmd   *cobra.Command
	lkInstCreateCmd      *cobra.Command
	lkInstDeleteCmd      *cobra.Command
	lkInstDescribeCmd    *cobra.Command
	lkInstListCmd        *cobra.Command
	lkInstUpdateCmd      *cobra.Command
	lkInstRestartCmd     *cobra.Command
	lkInstRestoreCmd     *cobra.Command
	lkInstExportCmd      *cobra.Command
	lkInstImportCmd      *cobra.Command
	lookerBackupsCmd     *cobra.Command
	lkBackupCreateCmd    *cobra.Command
	lkBackupDeleteCmd    *cobra.Command
	lkBackupDescribeCmd  *cobra.Command
	lkBackupListCmd      *cobra.Command
}

// newLookerCmd returns the "looker" command group for inv.
func newLookerCmd(inv *invocation) *cobra.Command {
	return newLookerCmds(inv).lookerCmd
}

// newLookerCmds builds the "looker" command group for inv and returns its state.
func newLookerCmds(inv *invocation) *lookerCmds {
	s := &lookerCmds{inv: inv}
	s.lookerCmd = &cobra.Command{Use: "looker", Short: "Manage Looker"}
	s.lookerRegionsCmd = &cobra.Command{Use: "regions", Short: "Explore Looker regions"}
	s.lookerRegionsListCmd = &cobra.Command{
		Use: "list", Short: "List Looker regions", Args: cobra.NoArgs, RunE: s.runLKRegionsList,
	}
	s.lookerOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Looker operations"}
	s.lkOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: s.runLKOpCancel,
	}
	s.lkOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runLKOpDescribe,
	}
	s.lkOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a region",
		Args: cobra.NoArgs, RunE: s.runLKOpList,
	}
	s.lookerInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Looker instances"}
	s.lkInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runLKInstCreate,
	}
	s.lkInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an instance",
		Args: cobra.ExactArgs(1), RunE: s.runLKInstDelete,
	}
	s.lkInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an instance",
		Args: cobra.ExactArgs(1), RunE: s.runLKInstDescribe,
	}
	s.lkInstListCmd = &cobra.Command{
		Use: "list", Short: "List instances in a region",
		Args: cobra.NoArgs, RunE: s.runLKInstList,
	}
	s.lkInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runLKInstUpdate,
	}
	s.lkInstRestartCmd = &cobra.Command{
		Use: "restart INSTANCE", Short: "Restart an instance",
		Args: cobra.ExactArgs(1), RunE: s.runLKInstRestart,
	}
	s.lkInstRestoreCmd = &cobra.Command{
		Use: "restore INSTANCE", Short: "Restore an instance from a backup (--config-file with RestoreInstanceRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runLKInstRestore,
	}
	s.lkInstExportCmd = &cobra.Command{
		Use: "export INSTANCE", Short: "Export an instance's data (--config-file with ExportInstanceRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runLKInstExport,
	}
	s.lkInstImportCmd = &cobra.Command{
		Use: "import INSTANCE", Short: "Import data into an instance (--config-file with ImportInstanceRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runLKInstImport,
	}
	s.lookerBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Looker instance backups"}
	s.lkBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a backup for an instance",
		Args: cobra.ExactArgs(1), RunE: s.runLKBackupCreate,
	}
	s.lkBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup",
		Args: cobra.ExactArgs(1), RunE: s.runLKBackupDelete,
	}
	s.lkBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup",
		Args: cobra.ExactArgs(1), RunE: s.runLKBackupDescribe,
	}
	s.lkBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups for an instance",
		Args: cobra.NoArgs, RunE: s.runLKBackupList,
	}
	s.initLooker()
	return s
}
