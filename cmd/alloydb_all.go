package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
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

func adbFinishOp(ctx context.Context, svc *alloydb.Service, op *alloydb.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := adbWaitOp(ctx, svc, op)
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
	flagADBRegion     string
	flagADBConfigFile string
	flagADBUpdateMask string
	flagADBFormat     string
	flagADBAsync      bool
	flagADBCluster    string
)

// --- backups ---

var alloydbBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage AlloyDB backups"}

var (
	adbBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a backup from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBBackupCreate,
	}
	adbBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup",
		Args: cobra.ExactArgs(1), RunE: runADBBackupDelete,
	}
	adbBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup",
		Args: cobra.ExactArgs(1), RunE: runADBBackupDescribe,
	}
	adbBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups in a region",
		Args: cobra.NoArgs, RunE: runADBBackupList,
	}
	adbBackupUpdateCmd = &cobra.Command{
		Use: "update BACKUP", Short: "Update a backup from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBBackupUpdate,
	}
)

func init() {
	all := []*cobra.Command{adbBackupCreateCmd, adbBackupDeleteCmd, adbBackupDescribeCmd, adbBackupListCmd, adbBackupUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagADBRegion, "region", "", "Region containing the backup (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{adbBackupCreateCmd, adbBackupUpdateCmd} {
		c.Flags().StringVar(&flagADBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Backup message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	adbBackupUpdateCmd.Flags().StringVar(&flagADBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{adbBackupCreateCmd, adbBackupDeleteCmd, adbBackupUpdateCmd} {
		c.Flags().BoolVar(&flagADBAsync, "async", false, "Return the long-running operation without waiting")
	}
	adbBackupDescribeCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")
	adbBackupListCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")

	alloydbBackupsCmd.AddCommand(all...)
	alloydbCmd.AddCommand(alloydbBackupsCmd)
}

func adbBackupName(id, project, region string) string {
	return adbChild("backups", id, adbLocationParent(project, region))
}

func runADBBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	b := &alloydb.Backup{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, b); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Create(adbLocationParent(project, flagADBRegion), b).
		BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Create backup", args[0], flagADBAsync)
}

func runADBBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Delete(adbBackupName(args[0], project, flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Delete backup", args[0], flagADBAsync)
}

func runADBBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Backups.Get(adbBackupName(args[0], project, flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagADBFormat)
}

func runADBBackupList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Backups.List(adbLocationParent(project, flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if flagADBFormat != "" {
		return emitFormatted(resp.Backups, flagADBFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, b := range resp.Backups {
		fmt.Printf("%-40s %s\n", path.Base(b.Name), b.State)
	}
	return nil
}

func runADBBackupUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	b := &alloydb.Backup{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, b); err != nil {
		return err
	}
	mask := flagADBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(b))
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Patch(adbBackupName(args[0], project, flagADBRegion), b).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Update backup", args[0], flagADBAsync)
}

// --- clusters ---

var alloydbClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage AlloyDB clusters"}

var (
	adbClusterCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBClusterCreate,
	}
	adbClusterCreateSecondaryCmd = &cobra.Command{
		Use: "create-secondary CLUSTER", Short: "Create a secondary cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBClusterCreateSecondary,
	}
	adbClusterDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a cluster",
		Args: cobra.ExactArgs(1), RunE: runADBClusterDelete,
	}
	adbClusterDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a cluster",
		Args: cobra.ExactArgs(1), RunE: runADBClusterDescribe,
	}
	adbClusterListCmd = &cobra.Command{
		Use: "list", Short: "List clusters in a region",
		Args: cobra.NoArgs, RunE: runADBClusterList,
	}
	adbClusterUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBClusterUpdate,
	}
	adbClusterPromoteCmd = &cobra.Command{
		Use: "promote CLUSTER", Short: "Promote a secondary cluster to primary",
		Args: cobra.ExactArgs(1), RunE: runADBClusterPromote,
	}
	adbClusterRestoreCmd = &cobra.Command{
		Use: "restore CLUSTER", Short: "Restore a cluster from a backup/PITR (--config-file)",
		Args: cobra.ExactArgs(1), RunE: runADBClusterRestore,
	}
	adbClusterSwitchoverCmd = &cobra.Command{
		Use: "switchover CLUSTER", Short: "Switchover to a secondary cluster",
		Args: cobra.ExactArgs(1), RunE: runADBClusterSwitchover,
	}
	adbClusterUpgradeCmd = &cobra.Command{
		Use: "upgrade CLUSTER", Short: "Upgrade a cluster (--config-file with UpgradeClusterRequest)",
		Args: cobra.ExactArgs(1), RunE: runADBClusterUpgrade,
	}
	adbClusterMigrateCmd = &cobra.Command{
		Use: "migrate-cloud-sql CLUSTER", Short: "Migrate from Cloud SQL (--config-file with RestoreFromCloudSQLRequest)",
		Args: cobra.ExactArgs(1), RunE: runADBClusterMigrate,
	}
	adbClusterExportCmd = &cobra.Command{
		Use: "export CLUSTER", Short: "Export a cluster (--config-file with ExportClusterRequest)",
		Args: cobra.ExactArgs(1), RunE: runADBClusterExport,
	}
	adbClusterImportCmd = &cobra.Command{
		Use: "import CLUSTER", Short: "Import into a cluster (--config-file with ImportClusterRequest)",
		Args: cobra.ExactArgs(1), RunE: runADBClusterImport,
	}
)

func init() {
	all := []*cobra.Command{
		adbClusterCreateCmd, adbClusterCreateSecondaryCmd, adbClusterDeleteCmd,
		adbClusterDescribeCmd, adbClusterListCmd, adbClusterUpdateCmd,
		adbClusterPromoteCmd, adbClusterRestoreCmd, adbClusterSwitchoverCmd,
		adbClusterUpgradeCmd, adbClusterMigrateCmd, adbClusterExportCmd, adbClusterImportCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagADBRegion, "region", "", "Region containing the cluster (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{
		adbClusterCreateCmd, adbClusterCreateSecondaryCmd, adbClusterUpdateCmd,
		adbClusterRestoreCmd, adbClusterUpgradeCmd, adbClusterMigrateCmd,
		adbClusterExportCmd, adbClusterImportCmd,
	} {
		c.Flags().StringVar(&flagADBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	adbClusterUpdateCmd.Flags().StringVar(&flagADBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{
		adbClusterCreateCmd, adbClusterCreateSecondaryCmd, adbClusterDeleteCmd,
		adbClusterUpdateCmd, adbClusterPromoteCmd, adbClusterRestoreCmd,
		adbClusterSwitchoverCmd, adbClusterUpgradeCmd, adbClusterMigrateCmd,
		adbClusterExportCmd, adbClusterImportCmd,
	} {
		c.Flags().BoolVar(&flagADBAsync, "async", false, "Return the long-running operation without waiting")
	}
	adbClusterDescribeCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")
	adbClusterListCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")

	alloydbClustersCmd.AddCommand(all...)
	alloydbCmd.AddCommand(alloydbClustersCmd)
}

func adbClusterName(id, project, region string) string {
	return adbChild("clusters", id, adbLocationParent(project, region))
}

func runADBClusterCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &alloydb.Cluster{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, c); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Create(adbLocationParent(project, flagADBRegion), c).
		ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Create cluster", args[0], flagADBAsync)
}

func runADBClusterCreateSecondary(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &alloydb.Cluster{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, c); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Createsecondary(adbLocationParent(project, flagADBRegion), c).
		ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating secondary cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Create secondary cluster", args[0], flagADBAsync)
}

func runADBClusterDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(adbClusterName(args[0], project, flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Delete cluster", args[0], flagADBAsync)
}

func runADBClusterDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Get(adbClusterName(args[0], project, flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormatted(got, flagADBFormat)
}

func runADBClusterList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(adbLocationParent(project, flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if flagADBFormat != "" {
		return emitFormatted(resp.Clusters, flagADBFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "CLUSTER_TYPE")
	for _, c := range resp.Clusters {
		fmt.Printf("%-40s %-15s %s\n", path.Base(c.Name), c.State, c.ClusterType)
	}
	return nil
}

func runADBClusterUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &alloydb.Cluster{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, c); err != nil {
		return err
	}
	mask := flagADBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Patch(adbClusterName(args[0], project, flagADBRegion), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Update cluster", args[0], flagADBAsync)
}

func runADBClusterPromote(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Promote(adbClusterName(args[0], project, flagADBRegion), &alloydb.PromoteClusterRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("promoting cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Promote cluster", args[0], flagADBAsync)
}

func runADBClusterRestore(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &alloydb.RestoreClusterRequest{ClusterId: args[0]}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, req); err != nil {
		return err
	}
	if req.ClusterId == "" {
		req.ClusterId = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Restore(adbLocationParent(project, flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Restore cluster", args[0], flagADBAsync)
}

func runADBClusterSwitchover(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Switchover(adbClusterName(args[0], project, flagADBRegion), &alloydb.SwitchoverClusterRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("switching over cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Switchover cluster", args[0], flagADBAsync)
}

func runADBClusterUpgrade(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &alloydb.UpgradeClusterRequest{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Upgrade(adbClusterName(args[0], project, flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upgrading cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Upgrade cluster", args[0], flagADBAsync)
}

func runADBClusterMigrate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &alloydb.RestoreFromCloudSQLRequest{ClusterId: args[0]}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, req); err != nil {
		return err
	}
	if req.ClusterId == "" {
		req.ClusterId = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.RestoreFromCloudSQL(adbLocationParent(project, flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating from Cloud SQL: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Migrate cluster", args[0], flagADBAsync)
}

func runADBClusterExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &alloydb.ExportClusterRequest{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Export(adbClusterName(args[0], project, flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Export cluster", args[0], flagADBAsync)
}

func runADBClusterImport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &alloydb.ImportClusterRequest{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Import(adbClusterName(args[0], project, flagADBRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing cluster: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Import cluster", args[0], flagADBAsync)
}

// --- instances ---

var alloydbInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage AlloyDB instances"}

var (
	adbInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBInstCreate,
	}
	adbInstCreateSecondaryCmd = &cobra.Command{
		Use: "create-secondary INSTANCE", Short: "Create a secondary instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBInstCreateSecondary,
	}
	adbInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an instance",
		Args: cobra.ExactArgs(1), RunE: runADBInstDelete,
	}
	adbInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an instance",
		Args: cobra.ExactArgs(1), RunE: runADBInstDescribe,
	}
	adbInstListCmd = &cobra.Command{
		Use: "list", Short: "List instances in a cluster",
		Args: cobra.NoArgs, RunE: runADBInstList,
	}
	adbInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBInstUpdate,
	}
	adbInstRestartCmd = &cobra.Command{
		Use: "restart INSTANCE", Short: "Restart an instance",
		Args: cobra.ExactArgs(1), RunE: runADBInstRestart,
	}
	adbInstFailoverCmd = &cobra.Command{
		Use: "failover INSTANCE", Short: "Fail over an instance to a secondary",
		Args: cobra.ExactArgs(1), RunE: runADBInstFailover,
	}
	adbInstInjectFaultCmd = &cobra.Command{
		Use: "inject-fault INSTANCE", Short: "Inject a fault into an instance (--config-file with InjectFaultRequest)",
		Args: cobra.ExactArgs(1), RunE: runADBInstInjectFault,
	}
	adbInstGetConnInfoCmd = &cobra.Command{
		Use: "get-connection-info", Short: "Get instance connection info",
		Args: cobra.NoArgs, RunE: runADBInstGetConnInfo,
	}
)

func init() {
	all := []*cobra.Command{
		adbInstCreateCmd, adbInstCreateSecondaryCmd, adbInstDeleteCmd, adbInstDescribeCmd,
		adbInstListCmd, adbInstUpdateCmd, adbInstRestartCmd, adbInstFailoverCmd,
		adbInstInjectFaultCmd, adbInstGetConnInfoCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagADBRegion, "region", "", "Region containing the cluster (required)")
		c.Flags().StringVar(&flagADBCluster, "cluster", "", "Cluster containing the instance (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("cluster")
	}
	for _, c := range []*cobra.Command{adbInstCreateCmd, adbInstCreateSecondaryCmd, adbInstUpdateCmd, adbInstInjectFaultCmd} {
		c.Flags().StringVar(&flagADBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Instance / InjectFaultRequest body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	adbInstUpdateCmd.Flags().StringVar(&flagADBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{
		adbInstCreateCmd, adbInstCreateSecondaryCmd, adbInstDeleteCmd, adbInstUpdateCmd,
		adbInstRestartCmd, adbInstFailoverCmd, adbInstInjectFaultCmd,
	} {
		c.Flags().BoolVar(&flagADBAsync, "async", false, "Return the long-running operation without waiting")
	}
	adbInstDescribeCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")
	adbInstListCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")
	adbInstGetConnInfoCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")

	alloydbInstancesCmd.AddCommand(all...)
	alloydbCmd.AddCommand(alloydbInstancesCmd)
}

func adbClusterParent(project, region, cluster string) string {
	return adbClusterName(cluster, project, region)
}

func adbInstName(id, project, region, cluster string) string {
	return adbChild("instances", id, adbClusterParent(project, region, cluster))
}

func runADBInstCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &alloydb.Instance{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, inst); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Create(adbClusterParent(project, flagADBRegion, flagADBCluster), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Create instance", args[0], flagADBAsync)
}

func runADBInstCreateSecondary(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &alloydb.Instance{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, inst); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Createsecondary(adbClusterParent(project, flagADBRegion, flagADBCluster), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating secondary instance: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Create secondary instance", args[0], flagADBAsync)
}

func runADBInstDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Delete(adbInstName(args[0], project, flagADBRegion, flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Delete instance", args[0], flagADBAsync)
}

func runADBInstDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Instances.Get(adbInstName(args[0], project, flagADBRegion, flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagADBFormat)
}

func runADBInstList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.Instances.List(adbClusterParent(project, flagADBRegion, flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if flagADBFormat != "" {
		return emitFormatted(resp.Instances, flagADBFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "INSTANCE_TYPE")
	for _, i := range resp.Instances {
		fmt.Printf("%-40s %-15s %s\n", path.Base(i.Name), i.State, i.InstanceType)
	}
	return nil
}

func runADBInstUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &alloydb.Instance{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, inst); err != nil {
		return err
	}
	mask := flagADBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Patch(adbInstName(args[0], project, flagADBRegion, flagADBCluster), inst).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Update instance", args[0], flagADBAsync)
}

func runADBInstRestart(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Restart(adbInstName(args[0], project, flagADBRegion, flagADBCluster), &alloydb.RestartInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting instance: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Restart instance", args[0], flagADBAsync)
}

func runADBInstFailover(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.Failover(adbInstName(args[0], project, flagADBRegion, flagADBCluster), &alloydb.FailoverInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failing over instance: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Fail over instance", args[0], flagADBAsync)
}

func runADBInstInjectFault(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &alloydb.InjectFaultRequest{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Instances.InjectFault(adbInstName(args[0], project, flagADBRegion, flagADBCluster), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("injecting fault: %w", err)
	}
	return adbFinishOp(ctx, svc, op, "Inject fault", args[0], flagADBAsync)
}

func runADBInstGetConnInfo(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Instances.GetConnectionInfo(adbClusterParent(project, flagADBRegion, flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting connection info: %w", err)
	}
	return emitFormatted(got, flagADBFormat)
}

// --- users ---

var alloydbUsersCmd = &cobra.Command{Use: "users", Short: "Manage AlloyDB users"}

var (
	adbUserCreateCmd = &cobra.Command{
		Use: "create USER", Short: "Create a user from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBUserCreate,
	}
	adbUserDeleteCmd = &cobra.Command{
		Use: "delete USER", Short: "Delete a user",
		Args: cobra.ExactArgs(1), RunE: runADBUserDelete,
	}
	adbUserDescribeCmd = &cobra.Command{
		Use: "describe USER", Short: "Describe a user",
		Args: cobra.ExactArgs(1), RunE: runADBUserDescribe,
	}
	adbUserListCmd = &cobra.Command{
		Use: "list", Short: "List users in a cluster",
		Args: cobra.NoArgs, RunE: runADBUserList,
	}
	adbUserUpdateCmd = &cobra.Command{
		Use: "update USER", Short: "Update a user from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runADBUserUpdate,
	}
)

func init() {
	all := []*cobra.Command{adbUserCreateCmd, adbUserDeleteCmd, adbUserDescribeCmd, adbUserListCmd, adbUserUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagADBRegion, "region", "", "Region containing the cluster (required)")
		c.Flags().StringVar(&flagADBCluster, "cluster", "", "Cluster containing the user (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("cluster")
	}
	for _, c := range []*cobra.Command{adbUserCreateCmd, adbUserUpdateCmd} {
		c.Flags().StringVar(&flagADBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the User message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	adbUserUpdateCmd.Flags().StringVar(&flagADBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	adbUserDescribeCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")
	adbUserListCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")

	alloydbUsersCmd.AddCommand(all...)
	alloydbCmd.AddCommand(alloydbUsersCmd)
}

func adbUserName(id, project, region, cluster string) string {
	return adbChild("users", id, adbClusterParent(project, region, cluster))
}

func runADBUserCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	u := &alloydb.User{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, u); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Users.Create(adbClusterParent(project, flagADBRegion, flagADBCluster), u).
		UserId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return emitFormatted(got, "")
}

func runADBUserDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Clusters.Users.Delete(adbUserName(args[0], project, flagADBRegion, flagADBCluster)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	fmt.Printf("Deleted user [%s].\n", args[0])
	return nil
}

func runADBUserDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Users.Get(adbUserName(args[0], project, flagADBRegion, flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing user: %w", err)
	}
	return emitFormatted(got, flagADBFormat)
}

func runADBUserList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.Users.List(adbClusterParent(project, flagADBRegion, flagADBCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}
	if flagADBFormat != "" {
		return emitFormatted(resp.Users, flagADBFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "USER_TYPE")
	for _, u := range resp.Users {
		fmt.Printf("%-40s %s\n", path.Base(u.Name), u.UserType)
	}
	return nil
}

func runADBUserUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	u := &alloydb.User{}
	if err := loadYAMLOrJSONInto(flagADBConfigFile, u); err != nil {
		return err
	}
	mask := flagADBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(u))
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Users.Patch(adbUserName(args[0], project, flagADBRegion, flagADBCluster), u).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return emitFormatted(got, "")
}

// --- operations ---

var alloydbOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage AlloyDB operations"}

var (
	adbOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: runADBOpCancel,
	}
	adbOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: runADBOpDelete,
	}
	adbOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runADBOpDescribe,
	}
	adbOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a region",
		Args: cobra.NoArgs, RunE: runADBOpList,
	}
)

func init() {
	for _, c := range []*cobra.Command{adbOpCancelCmd, adbOpDeleteCmd, adbOpDescribeCmd, adbOpListCmd} {
		c.Flags().StringVar(&flagADBRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	adbOpDescribeCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")
	adbOpListCmd.Flags().StringVar(&flagADBFormat, "format", "", "Output format")

	alloydbOperationsCmd.AddCommand(adbOpCancelCmd, adbOpDeleteCmd, adbOpDescribeCmd, adbOpListCmd)
	alloydbCmd.AddCommand(alloydbOperationsCmd)
}

func adbOpName(id, project, region string) string {
	return adbChild("operations", id, adbLocationParent(project, region))
}

func runADBOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(adbOpName(args[0], project, flagADBRegion), &alloydb.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runADBOpDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(adbOpName(args[0], project, flagADBRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runADBOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(adbOpName(args[0], project, flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagADBFormat)
}

func runADBOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AlloyDBService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(adbLocationParent(project, flagADBRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagADBFormat != "" {
		return emitFormatted(resp.Operations, flagADBFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
