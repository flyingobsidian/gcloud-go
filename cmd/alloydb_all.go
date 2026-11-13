package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	alloydb "google.golang.org/api/alloydb/v1"
)

func adbLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func adbChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func adbWaitOp(ctx context.Context, svc *alloydb.Service, op *alloydb.Operation) (*alloydb.Operation, error) {
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

func (s *alloydbCmds) adbFinishOp(ctx context.Context, svc *alloydb.Service, op *alloydb.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := adbWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// adbApplyCreateLabels merges the --labels flag onto an Instance body loaded
// from --config-file. Flag values take precedence for overlapping keys.
func (s *alloydbCmds) adbApplyCreateLabels(inst *alloydb.Instance) {
	if len(s.flagADBLabels) == 0 {
		return
	}
	if inst.Labels == nil {
		inst.Labels = make(map[string]string, len(s.flagADBLabels))
	}
	for k, v := range s.flagADBLabels {
		inst.Labels[k] = v
	}
}

// adbApplyUpdateLabels applies --update-labels / --remove-labels /
// --clear-labels onto an Instance body and returns the field-mask entry that
// callers should include ("labels") if any label mutation happened.
func (s *alloydbCmds) adbApplyUpdateLabels(inst *alloydb.Instance) (labelsMaskEntry string) {
	if !s.flagADBClearLabels && len(s.flagADBUpdateLabels) == 0 && len(s.flagADBRemoveLabels) == 0 {
		return ""
	}
	if s.flagADBClearLabels {
		inst.Labels = map[string]string{}
	}
	if inst.Labels == nil {
		inst.Labels = map[string]string{}
	}
	// --update-labels first, then --remove-labels so that removal wins on
	// conflicting keys (matches gcloud-python's label helper).
	for k, v := range s.flagADBUpdateLabels {
		inst.Labels[k] = v
	}
	for _, k := range s.flagADBRemoveLabels {
		delete(inst.Labels, k)
	}
	if len(inst.Labels) == 0 {
		inst.ForceSendFields = append(inst.ForceSendFields, "Labels")
	}
	return "labels"
}

// adbAddInstanceLabelFlags binds the create/create-secondary label flag.
func (s *alloydbCmds) adbAddInstanceLabelFlags(c *cobra.Command) {
	c.Flags().StringToStringVar(&s.flagADBLabels, "labels", nil,
		"Labels to assign to the instance, as KEY=VALUE pairs (repeatable)")
}

// adbAddInstanceUpdateLabelFlags binds the update label-mutation flags.
func (s *alloydbCmds) adbAddInstanceUpdateLabelFlags(c *cobra.Command) {
	c.Flags().StringToStringVar(&s.flagADBUpdateLabels, "update-labels", nil,
		"Labels to add or overwrite on the instance, as KEY=VALUE pairs (repeatable)")
	c.Flags().StringSliceVar(&s.flagADBRemoveLabels, "remove-labels", nil,
		"Comma-separated list of label keys to remove from the instance")
	c.Flags().BoolVar(&s.flagADBClearLabels, "clear-labels", false,
		"Clear all labels on the instance before applying --update-labels")
}

// adbApplyClusterRestoreBackupDR wires --backupdr-backup / --backupdr-data-source
// onto a RestoreClusterRequest. The two flags are mutually exclusive: --backupdr-backup
// populates BackupdrBackupSource (restore-from-backup) and --backupdr-data-source
// populates BackupdrPitrSource (point-in-time recovery). Values loaded from
// --config-file are preserved when the corresponding flag is empty.
func (s *alloydbCmds) adbApplyClusterRestoreBackupDR(req *alloydb.RestoreClusterRequest) error {
	if s.flagADBBackupDRBackup != "" && s.flagADBBackupDRDataSource != "" {
		return fmt.Errorf("--backupdr-backup and --backupdr-data-source are mutually exclusive")
	}
	if s.flagADBBackupDRBackup != "" {
		req.BackupdrBackupSource = &alloydb.BackupDrBackupSource{Backup: s.flagADBBackupDRBackup}
	}
	if s.flagADBBackupDRDataSource != "" {
		req.BackupdrPitrSource = &alloydb.BackupDrPitrSource{DataSource: s.flagADBBackupDRDataSource}
	}
	return nil
}

// --- backups ---

func (s *alloydbCmds) initAlloydbAll() {
	all := []*cobra.Command{s.adbBackupCreateCmd, s.adbBackupDeleteCmd, s.adbBackupDescribeCmd, s.adbBackupListCmd, s.adbBackupUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagADBRegion, "region", "", "Region containing the backup (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.adbBackupCreateCmd, s.adbBackupUpdateCmd} {
		c.Flags().StringVar(&s.flagADBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Backup message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.adbBackupUpdateCmd.Flags().StringVar(&s.flagADBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.adbBackupCreateCmd, s.adbBackupDeleteCmd, s.adbBackupUpdateCmd} {
		c.Flags().BoolVar(&s.flagADBAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.adbBackupDescribeCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")
	s.adbBackupListCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")

	s.alloydbBackupsCmd.AddCommand(all...)
	s.alloydbCmd.AddCommand(s.alloydbBackupsCmd)
}

func adbBackupName(id, project, region string) string {
	return adbChild("backups", id, adbLocationParent(project, region))
}

func (s *alloydbCmds) runADBBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	b := &alloydb.Backup{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, b); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Create(adbLocationParent(project, s.flagADBRegion), b).
		BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Create backup", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Delete(adbBackupName(args[0], project, s.flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Delete backup", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Backups.Get(adbBackupName(args[0], project, s.flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADBFormat)
}

func (s *alloydbCmds) runADBBackupList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Backups.List(adbLocationParent(project, s.flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if s.flagADBFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Backups, s.flagADBFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Backups), "%-40s %s\n", "NAME", "STATE")
	for _, b := range resp.Backups {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(b.Name), b.State)
	}
	return nil
}

func (s *alloydbCmds) runADBBackupUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	b := &alloydb.Backup{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, b); err != nil {
		return err
	}
	mask := s.flagADBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(b))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Patch(adbBackupName(args[0], project, s.flagADBRegion), b).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Update backup", args[0], s.flagADBAsync)
}

// --- clusters ---

func (s *alloydbCmds) initAlloydbAll2() {
	all := []*cobra.Command{
		s.adbClusterCreateCmd, s.adbClusterCreateSecondaryCmd, s.adbClusterDeleteCmd,
		s.adbClusterDescribeCmd, s.adbClusterListCmd, s.adbClusterUpdateCmd,
		s.adbClusterPromoteCmd, s.adbClusterRestoreCmd, s.adbClusterSwitchoverCmd,
		s.adbClusterUpgradeCmd, s.adbClusterMigrateCmd, s.adbClusterExportCmd, s.adbClusterImportCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagADBRegion, "region", "", "Region containing the cluster (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{
		s.adbClusterCreateCmd, s.adbClusterCreateSecondaryCmd, s.adbClusterUpdateCmd,
		s.adbClusterRestoreCmd, s.adbClusterUpgradeCmd, s.adbClusterMigrateCmd,
		s.adbClusterExportCmd, s.adbClusterImportCmd,
	} {
		c.Flags().StringVar(&s.flagADBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the request body")
		if c != s.adbClusterRestoreCmd {
			_ = c.MarkFlagRequired("config-file")
		}
	}
	// `restore` also accepts BackupDR sources via scalar flags (added GA in
	// gcloud-python 578.0.0), so --config-file is optional if either BackupDR
	// flag is present. runADBClusterRestore enforces that at least one source
	// is supplied.
	s.adbClusterRestoreCmd.Flags().StringVar(&s.flagADBBackupDRBackup, "backupdr-backup", "",
		"Full resource name of a BackupDR backup to restore from "+
			"(projects/.../backupVaults/.../dataSources/.../backups/...)")
	s.adbClusterRestoreCmd.Flags().StringVar(&s.flagADBBackupDRDataSource, "backupdr-data-source", "",
		"Full resource name of a BackupDR data source to point-in-time recover from "+
			"(projects/.../backupVaults/.../dataSources/...)")
	s.adbClusterUpdateCmd.Flags().StringVar(&s.flagADBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{
		s.adbClusterCreateCmd, s.adbClusterCreateSecondaryCmd, s.adbClusterDeleteCmd,
		s.adbClusterUpdateCmd, s.adbClusterPromoteCmd, s.adbClusterRestoreCmd,
		s.adbClusterSwitchoverCmd, s.adbClusterUpgradeCmd, s.adbClusterMigrateCmd,
		s.adbClusterExportCmd, s.adbClusterImportCmd,
	} {
		c.Flags().BoolVar(&s.flagADBAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.adbClusterDescribeCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")
	s.adbClusterListCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")

	s.alloydbClustersCmd.AddCommand(all...)
	s.alloydbCmd.AddCommand(s.alloydbClustersCmd)
}

func adbClusterName(id, project, region string) string {
	return adbChild("clusters", id, adbLocationParent(project, region))
}

func (s *alloydbCmds) runADBClusterCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &alloydb.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, c); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Create(adbLocationParent(project, s.flagADBRegion), c).
		ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Create cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterCreateSecondary(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &alloydb.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, c); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Createsecondary(adbLocationParent(project, s.flagADBRegion), c).
		ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating secondary cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Create secondary cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(adbClusterName(args[0], project, s.flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Delete cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Get(adbClusterName(args[0], project, s.flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADBFormat)
}

func (s *alloydbCmds) runADBClusterList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(adbLocationParent(project, s.flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if s.flagADBFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Clusters, s.flagADBFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Clusters), "%-40s %-15s %s\n", "NAME", "STATE", "CLUSTER_TYPE")
	for _, c := range resp.Clusters {
		fmt.Fprintf(s.inv.stdout, "%-40s %-15s %s\n", path.Base(c.Name), c.State, c.ClusterType)
	}
	return nil
}

func (s *alloydbCmds) runADBClusterUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &alloydb.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, c); err != nil {
		return err
	}
	mask := s.flagADBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Patch(adbClusterName(args[0], project, s.flagADBRegion), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Update cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterPromote(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Promote(adbClusterName(args[0], project, s.flagADBRegion), &alloydb.PromoteClusterRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("promoting cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Promote cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterRestore(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &alloydb.RestoreClusterRequest{ClusterId: args[0]}
	if s.flagADBConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagADBConfigFile, req); err != nil {
			return err
		}
	}
	if req.ClusterId == "" {
		req.ClusterId = args[0]
	}
	if err := s.adbApplyClusterRestoreBackupDR(req); err != nil {
		return err
	}
	if req.BackupSource == nil && req.BackupdrBackupSource == nil && req.BackupdrPitrSource == nil && req.ContinuousBackupSource == nil {
		return fmt.Errorf("no restore source: supply --config-file (with backupSource / continuousBackupSource) or --backupdr-backup / --backupdr-data-source")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Restore(adbLocationParent(project, s.flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Restore cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterSwitchover(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Switchover(adbClusterName(args[0], project, s.flagADBRegion), &alloydb.SwitchoverClusterRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("switching over cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Switchover cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterUpgrade(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &alloydb.UpgradeClusterRequest{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Upgrade(adbClusterName(args[0], project, s.flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upgrading cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Upgrade cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterMigrate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &alloydb.RestoreFromCloudSQLRequest{ClusterId: args[0]}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, req); err != nil {
		return err
	}
	if req.ClusterId == "" {
		req.ClusterId = args[0]
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.RestoreFromCloudSQL(adbLocationParent(project, s.flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating from Cloud SQL: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Migrate cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &alloydb.ExportClusterRequest{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Export(adbClusterName(args[0], project, s.flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Export cluster", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBClusterImport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &alloydb.ImportClusterRequest{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Import(adbClusterName(args[0], project, s.flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing cluster: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Import cluster", args[0], s.flagADBAsync)
}

// --- instances ---

func (s *alloydbCmds) initAlloydbAll3() {
	all := []*cobra.Command{
		s.adbInstCreateCmd, s.adbInstCreateSecondaryCmd, s.adbInstDeleteCmd, s.adbInstDescribeCmd,
		s.adbInstListCmd, s.adbInstUpdateCmd, s.adbInstRestartCmd, s.adbInstFailoverCmd,
		s.adbInstInjectFaultCmd, s.adbInstGetConnInfoCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagADBRegion, "region", "", "Region containing the cluster (required)")
		c.Flags().StringVar(&s.flagADBCluster, "cluster", "", "Cluster containing the instance (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("cluster")
	}
	for _, c := range []*cobra.Command{s.adbInstCreateCmd, s.adbInstCreateSecondaryCmd, s.adbInstUpdateCmd, s.adbInstInjectFaultCmd} {
		c.Flags().StringVar(&s.flagADBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Instance / InjectFaultRequest body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.adbInstUpdateCmd.Flags().StringVar(&s.flagADBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	// Label flags: --labels for create/create-secondary; --update-labels /
	// --remove-labels / --clear-labels for update (added in gcloud-python
	// 570.0.0/571.0.0 across the beta and GA tracks).
	s.adbAddInstanceLabelFlags(s.adbInstCreateCmd)
	s.adbAddInstanceLabelFlags(s.adbInstCreateSecondaryCmd)
	s.adbAddInstanceUpdateLabelFlags(s.adbInstUpdateCmd)
	for _, c := range []*cobra.Command{
		s.adbInstCreateCmd, s.adbInstCreateSecondaryCmd, s.adbInstDeleteCmd, s.adbInstUpdateCmd,
		s.adbInstRestartCmd, s.adbInstFailoverCmd, s.adbInstInjectFaultCmd,
	} {
		c.Flags().BoolVar(&s.flagADBAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.adbInstDescribeCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")
	s.adbInstListCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")
	s.adbInstGetConnInfoCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")

	s.alloydbInstancesCmd.AddCommand(all...)
	s.alloydbCmd.AddCommand(s.alloydbInstancesCmd)
}

func adbClusterParent(project, region, cluster string) string {
	return adbClusterName(cluster, project, region)
}

func adbInstName(id, project, region, cluster string) string {
	return adbChild("instances", id, adbClusterParent(project, region, cluster))
}

func (s *alloydbCmds) runADBInstCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	inst := &alloydb.Instance{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, inst); err != nil {
		return err
	}
	s.adbApplyCreateLabels(inst)
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Create(adbClusterParent(project, s.flagADBRegion, s.flagADBCluster), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Create instance", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBInstCreateSecondary(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	inst := &alloydb.Instance{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, inst); err != nil {
		return err
	}
	s.adbApplyCreateLabels(inst)
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Createsecondary(adbClusterParent(project, s.flagADBRegion, s.flagADBCluster), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating secondary instance: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Create secondary instance", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBInstDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Delete(adbInstName(args[0], project, s.flagADBRegion, s.flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Delete instance", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBInstDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Instances.Get(adbInstName(args[0], project, s.flagADBRegion, s.flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADBFormat)
}

func (s *alloydbCmds) runADBInstList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.Instances.List(adbClusterParent(project, s.flagADBRegion, s.flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if s.flagADBFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Instances, s.flagADBFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Instances), "%-40s %-15s %s\n", "NAME", "STATE", "INSTANCE_TYPE")
	for _, i := range resp.Instances {
		fmt.Fprintf(s.inv.stdout, "%-40s %-15s %s\n", path.Base(i.Name), i.State, i.InstanceType)
	}
	return nil
}

func (s *alloydbCmds) runADBInstUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	inst := &alloydb.Instance{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, inst); err != nil {
		return err
	}
	labelsMaskEntry := s.adbApplyUpdateLabels(inst)
	mask := s.flagADBUpdateMask
	if mask == "" {
		fields := nonEmptyJSONFields(inst)
		if labelsMaskEntry != "" {
			seen := false
			for _, f := range fields {
				if f == labelsMaskEntry {
					seen = true
					break
				}
			}
			if !seen {
				fields = append(fields, labelsMaskEntry)
			}
		}
		mask = joinMask(fields)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Patch(adbInstName(args[0], project, s.flagADBRegion, s.flagADBCluster), inst).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Update instance", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBInstRestart(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Restart(adbInstName(args[0], project, s.flagADBRegion, s.flagADBCluster), &alloydb.RestartInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting instance: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Restart instance", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBInstFailover(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Failover(adbInstName(args[0], project, s.flagADBRegion, s.flagADBCluster), &alloydb.FailoverInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failing over instance: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Fail over instance", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBInstInjectFault(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &alloydb.InjectFaultRequest{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.InjectFault(adbInstName(args[0], project, s.flagADBRegion, s.flagADBCluster), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("injecting fault: %w", err)
	}
	return s.adbFinishOp(ctx, svc, op, "Inject fault", args[0], s.flagADBAsync)
}

func (s *alloydbCmds) runADBInstGetConnInfo(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Instances.GetConnectionInfo(adbClusterParent(project, s.flagADBRegion, s.flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting connection info: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADBFormat)
}

// --- users ---

func (s *alloydbCmds) initAlloydbAll4() {
	all := []*cobra.Command{s.adbUserCreateCmd, s.adbUserDeleteCmd, s.adbUserDescribeCmd, s.adbUserListCmd, s.adbUserUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagADBRegion, "region", "", "Region containing the cluster (required)")
		c.Flags().StringVar(&s.flagADBCluster, "cluster", "", "Cluster containing the user (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("cluster")
	}
	for _, c := range []*cobra.Command{s.adbUserCreateCmd, s.adbUserUpdateCmd} {
		c.Flags().StringVar(&s.flagADBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the User message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.adbUserUpdateCmd.Flags().StringVar(&s.flagADBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.adbUserDescribeCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")
	s.adbUserListCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")

	s.alloydbUsersCmd.AddCommand(all...)
	s.alloydbCmd.AddCommand(s.alloydbUsersCmd)
}

func adbUserName(id, project, region, cluster string) string {
	return adbChild("users", id, adbClusterParent(project, region, cluster))
}

func (s *alloydbCmds) runADBUserCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	u := &alloydb.User{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, u); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Users.Create(adbClusterParent(project, s.flagADBRegion, s.flagADBCluster), u).
		UserId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *alloydbCmds) runADBUserDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Clusters.Users.Delete(adbUserName(args[0], project, s.flagADBRegion, s.flagADBCluster)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted user [%s].\n", args[0])
	return nil
}

func (s *alloydbCmds) runADBUserDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Users.Get(adbUserName(args[0], project, s.flagADBRegion, s.flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing user: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADBFormat)
}

func (s *alloydbCmds) runADBUserList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.Users.List(adbClusterParent(project, s.flagADBRegion, s.flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}
	if s.flagADBFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Users, s.flagADBFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Users), "%-40s %s\n", "NAME", "USER_TYPE")
	for _, u := range resp.Users {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(u.Name), u.UserType)
	}
	return nil
}

func (s *alloydbCmds) runADBUserUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	u := &alloydb.User{}
	if err := loadYAMLOrJSONInto(s.flagADBConfigFile, u); err != nil {
		return err
	}
	mask := s.flagADBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(u))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Users.Patch(adbUserName(args[0], project, s.flagADBRegion, s.flagADBCluster), u).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- operations ---

func (s *alloydbCmds) initAlloydbAll5() {
	for _, c := range []*cobra.Command{s.adbOpCancelCmd, s.adbOpDeleteCmd, s.adbOpDescribeCmd, s.adbOpListCmd} {
		c.Flags().StringVar(&s.flagADBRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	s.adbOpDescribeCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")
	s.adbOpListCmd.Flags().StringVar(&s.flagADBFormat, "format", "", "Output format")

	s.alloydbOperationsCmd.AddCommand(s.adbOpCancelCmd, s.adbOpDeleteCmd, s.adbOpDescribeCmd, s.adbOpListCmd)
	s.alloydbCmd.AddCommand(s.alloydbOperationsCmd)
}

func adbOpName(id, project, region string) string {
	return adbChild("operations", id, adbLocationParent(project, region))
}

func (s *alloydbCmds) runADBOpCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(adbOpName(args[0], project, s.flagADBRegion), &alloydb.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *alloydbCmds) runADBOpDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(adbOpName(args[0], project, s.flagADBRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *alloydbCmds) runADBOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(adbOpName(args[0], project, s.flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagADBFormat)
}

func (s *alloydbCmds) runADBOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AlloyDB(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(adbLocationParent(project, s.flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagADBFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagADBFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
