package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	redis "google.golang.org/api/redis/v1"
)

// --- shared helpers ---

func redisLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func redisWaitOp(ctx context.Context, svc *redis.Service, op *redis.Operation) (*redis.Operation, error) {
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

func redisFinishOp(ctx context.Context, svc *redis.Service, op *redis.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := redisWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// --- redis acl-policies ---

var redisAclPoliciesCmd = &cobra.Command{
	Use:   "acl-policies",
	Short: "Manage Redis Cluster ACL policies",
}

var (
	redisACLCreateCmd = &cobra.Command{
		Use: "create POLICY", Short: "Create an ACL policy from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runRedisACLCreate,
	}
	redisACLDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete an ACL policy",
		Args: cobra.ExactArgs(1), RunE: runRedisACLDelete,
	}
	redisACLDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe an ACL policy",
		Args: cobra.ExactArgs(1), RunE: runRedisACLDescribe,
	}
	redisACLListCmd = &cobra.Command{
		Use: "list", Short: "List ACL policies in a region",
		Args: cobra.NoArgs, RunE: runRedisACLList,
	}
	redisACLUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update an ACL policy from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runRedisACLUpdate,
	}
)

var (
	flagRedisACLRegion     string
	flagRedisACLConfigFile string
	flagRedisACLUpdateMask string
	flagRedisACLFormat     string
	flagRedisACLAsync      bool
)

func init() {
	for _, c := range []*cobra.Command{redisACLCreateCmd, redisACLDeleteCmd, redisACLDescribeCmd, redisACLListCmd, redisACLUpdateCmd} {
		c.Flags().StringVar(&flagRedisACLRegion, "region", "", "Region containing the ACL policy (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{redisACLCreateCmd, redisACLUpdateCmd} {
		c.Flags().StringVar(&flagRedisACLConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the AclPolicy message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	redisACLUpdateCmd.Flags().StringVar(&flagRedisACLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{redisACLCreateCmd, redisACLDeleteCmd, redisACLUpdateCmd} {
		c.Flags().BoolVar(&flagRedisACLAsync, "async", false, "Return the long-running operation without waiting")
	}
	redisACLDescribeCmd.Flags().StringVar(&flagRedisACLFormat, "format", "", "Output format")
	redisACLListCmd.Flags().StringVar(&flagRedisACLFormat, "format", "", "Output format")

	redisAclPoliciesCmd.AddCommand(redisACLCreateCmd, redisACLDeleteCmd, redisACLDescribeCmd, redisACLListCmd, redisACLUpdateCmd)
	redisCmd.AddCommand(redisAclPoliciesCmd)
}

func redisACLName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/aclPolicies/%s", redisLocationParent(project, region), id)
}

func runRedisACLCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	p := &redis.AclPolicy{}
	if err := loadYAMLOrJSONInto(flagRedisACLConfigFile, p); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	created, err := svc.Projects.Locations.AclPolicies.Create(redisLocationParent(project, flagRedisACLRegion), p).
		AclPolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ACL policy: %w", err)
	}
	return emitFormatted(created, "")
}

func runRedisACLDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AclPolicies.Delete(redisACLName(args[0], project, flagRedisACLRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting ACL policy: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Delete ACL policy", args[0], flagRedisACLAsync)
}

func runRedisACLDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Locations.AclPolicies.Get(redisACLName(args[0], project, flagRedisACLRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ACL policy: %w", err)
	}
	return emitFormatted(p, flagRedisACLFormat)
}

func runRedisACLList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.AclPolicies.List(redisLocationParent(project, flagRedisACLRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing ACL policies: %w", err)
	}
	if flagRedisACLFormat != "" {
		return emitFormatted(resp.AclPolicies, flagRedisACLFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, p := range resp.AclPolicies {
		fmt.Println(path.Base(p.Name))
	}
	return nil
}

func runRedisACLUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	p := &redis.AclPolicy{}
	if err := loadYAMLOrJSONInto(flagRedisACLConfigFile, p); err != nil {
		return err
	}
	mask := flagRedisACLUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(p))
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AclPolicies.Patch(redisACLName(args[0], project, flagRedisACLRegion), p).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating ACL policy: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Update ACL policy", args[0], flagRedisACLAsync)
}

// --- redis clusters ---

var redisClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Manage Memorystore for Redis Cluster instances",
}

var (
	redisClusterCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a Redis cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runRedisClusterCreate,
	}
	redisClusterDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: runRedisClusterDelete,
	}
	redisClusterDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: runRedisClusterDescribe,
	}
	redisClusterListCmd = &cobra.Command{
		Use: "list", Short: "List Redis clusters in a region",
		Args: cobra.NoArgs, RunE: runRedisClusterList,
	}
	redisClusterUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a Redis cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runRedisClusterUpdate,
	}
	redisClusterCreateBackupCmd = &cobra.Command{
		Use: "create-backup CLUSTER", Short: "Create an on-demand backup of a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: runRedisClusterBackup,
	}
	redisClusterGetCACmd = &cobra.Command{
		Use: "get-cluster-certificate-authority CLUSTER", Short: "Get the CA certificates for a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: runRedisClusterGetCA,
	}
	redisClusterGetSharedCACmd = &cobra.Command{
		Use: "get-shared-regional-certificate-authority", Short: "Get the shared regional CA for Redis clusters",
		Args: cobra.NoArgs, RunE: runRedisClusterGetSharedCA,
	}
	redisClusterRescheduleMaintCmd = &cobra.Command{
		Use: "reschedule-maintenance CLUSTER", Short: "Reschedule an upcoming maintenance window",
		Args: cobra.ExactArgs(1), RunE: runRedisClusterRescheduleMaintenance,
	}
	redisClusterAddTokenAuthCmd = &cobra.Command{
		Use: "add-token-auth-user CLUSTER", Short: "Add a token-auth user to a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: runRedisClusterAddTokenAuthUser,
	}
)

var (
	flagRedisClusterRegion         string
	flagRedisClusterConfigFile     string
	flagRedisClusterUpdateMask     string
	flagRedisClusterFormat         string
	flagRedisClusterAsync          bool
	flagRedisClusterBackupID       string
	flagRedisClusterBackupTTL      string
	flagRedisClusterMaintTime      string
	flagRedisClusterMaintType      string
	flagRedisClusterTokenAuthUser string
)

func init() {
	all := []*cobra.Command{
		redisClusterCreateCmd, redisClusterDeleteCmd, redisClusterDescribeCmd,
		redisClusterListCmd, redisClusterUpdateCmd, redisClusterCreateBackupCmd,
		redisClusterGetCACmd, redisClusterGetSharedCACmd, redisClusterRescheduleMaintCmd,
		redisClusterAddTokenAuthCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagRedisClusterRegion, "region", "", "Region containing the Redis cluster (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{redisClusterCreateCmd, redisClusterUpdateCmd} {
		c.Flags().StringVar(&flagRedisClusterConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Cluster message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	redisClusterUpdateCmd.Flags().StringVar(&flagRedisClusterUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{
		redisClusterCreateCmd, redisClusterDeleteCmd, redisClusterUpdateCmd,
		redisClusterCreateBackupCmd, redisClusterRescheduleMaintCmd, redisClusterAddTokenAuthCmd,
	} {
		c.Flags().BoolVar(&flagRedisClusterAsync, "async", false, "Return the long-running operation without waiting")
	}
	redisClusterDescribeCmd.Flags().StringVar(&flagRedisClusterFormat, "format", "", "Output format")
	redisClusterListCmd.Flags().StringVar(&flagRedisClusterFormat, "format", "", "Output format")
	redisClusterGetCACmd.Flags().StringVar(&flagRedisClusterFormat, "format", "", "Output format")
	redisClusterGetSharedCACmd.Flags().StringVar(&flagRedisClusterFormat, "format", "", "Output format")

	redisClusterCreateBackupCmd.Flags().StringVar(&flagRedisClusterBackupID, "backup-id", "",
		"ID for the new backup (required)")
	redisClusterCreateBackupCmd.Flags().StringVar(&flagRedisClusterBackupTTL, "ttl", "",
		"Retention duration for the backup, e.g. 24h")
	_ = redisClusterCreateBackupCmd.MarkFlagRequired("backup-id")

	redisClusterRescheduleMaintCmd.Flags().StringVar(&flagRedisClusterMaintType, "reschedule-type", "IMMEDIATE",
		"Reschedule type: IMMEDIATE or SPECIFIC_TIME")
	redisClusterRescheduleMaintCmd.Flags().StringVar(&flagRedisClusterMaintTime, "schedule-time", "",
		"RFC3339 time (for --reschedule-type=SPECIFIC_TIME)")

	redisClusterAddTokenAuthCmd.Flags().StringVar(&flagRedisClusterTokenAuthUser, "user", "",
		"IAM identity of the token-auth user (required)")
	_ = redisClusterAddTokenAuthCmd.MarkFlagRequired("user")

	redisClustersCmd.AddCommand(all...)
	registerRedisBackupCollections(redisClustersCmd)
	registerRedisClusterBackups(redisClustersCmd)
	redisCmd.AddCommand(redisClustersCmd)
}

func redisClusterName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/clusters/%s", redisLocationParent(project, region), id)
}

func runRedisClusterCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &redis.Cluster{}
	if err := loadYAMLOrJSONInto(flagRedisClusterConfigFile, c); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Create(redisLocationParent(project, flagRedisClusterRegion), c).
		ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Create cluster", args[0], flagRedisClusterAsync)
}

func runRedisClusterDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(redisClusterName(args[0], project, flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Delete cluster", args[0], flagRedisClusterAsync)
}

func runRedisClusterDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	c, err := svc.Projects.Locations.Clusters.Get(redisClusterName(args[0], project, flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormatted(c, flagRedisClusterFormat)
}

func runRedisClusterList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(redisLocationParent(project, flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if flagRedisClusterFormat != "" {
		return emitFormatted(resp.Clusters, flagRedisClusterFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "SIZE_GB")
	for _, c := range resp.Clusters {
		fmt.Printf("%-40s %-15s %d\n", path.Base(c.Name), c.State, c.SizeGb)
	}
	return nil
}

func runRedisClusterUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &redis.Cluster{}
	if err := loadYAMLOrJSONInto(flagRedisClusterConfigFile, c); err != nil {
		return err
	}
	mask := flagRedisClusterUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Patch(redisClusterName(args[0], project, flagRedisClusterRegion), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Update cluster", args[0], flagRedisClusterAsync)
}

func runRedisClusterBackup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &redis.BackupClusterRequest{BackupId: flagRedisClusterBackupID, Ttl: flagRedisClusterBackupTTL}
	op, err := svc.Projects.Locations.Clusters.Backup(redisClusterName(args[0], project, flagRedisClusterRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("backing up cluster: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Backup cluster", args[0], flagRedisClusterAsync)
}

func runRedisClusterGetCA(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	ca, err := svc.Projects.Locations.Clusters.GetCertificateAuthority(redisClusterName(args[0], project, flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting cluster CA: %w", err)
	}
	return emitFormatted(ca, flagRedisClusterFormat)
}

func runRedisClusterGetSharedCA(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	ca, err := svc.Projects.Locations.GetSharedRegionalCertificateAuthority(redisLocationParent(project, flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting shared regional CA: %w", err)
	}
	return emitFormatted(ca, flagRedisClusterFormat)
}

func runRedisClusterRescheduleMaintenance(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &redis.RescheduleClusterMaintenanceRequest{
		RescheduleType: flagRedisClusterMaintType,
		ScheduleTime:   flagRedisClusterMaintTime,
	}
	op, err := svc.Projects.Locations.Clusters.RescheduleClusterMaintenance(redisClusterName(args[0], project, flagRedisClusterRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rescheduling maintenance: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Reschedule maintenance", args[0], flagRedisClusterAsync)
}

func runRedisClusterAddTokenAuthUser(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &redis.AddTokenAuthUserRequest{TokenAuthUser: flagRedisClusterTokenAuthUser}
	op, err := svc.Projects.Locations.Clusters.AddTokenAuthUser(redisClusterName(args[0], project, flagRedisClusterRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding token auth user: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Add token auth user", args[0], flagRedisClusterAsync)
}

// --- redis clusters backup-collections + backups (nested) ---

var (
	redisBackupCollectionsCmd = &cobra.Command{
		Use:   "backup-collections",
		Short: "Manage backup collections for Redis clusters",
	}
	redisBCDescribeCmd = &cobra.Command{
		Use: "describe COLLECTION", Short: "Describe a backup collection",
		Args: cobra.ExactArgs(1), RunE: runRedisBCDescribe,
	}
	redisBCListCmd = &cobra.Command{
		Use: "list", Short: "List backup collections in a region",
		Args: cobra.NoArgs, RunE: runRedisBCList,
	}

	redisClusterBackupsCmd = &cobra.Command{
		Use:   "backups",
		Short: "Manage backups for a Redis cluster",
	}
	redisBackupsDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a Redis cluster backup",
		Args: cobra.ExactArgs(1), RunE: runRedisBackupsDescribe,
	}
	redisBackupsDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a Redis cluster backup",
		Args: cobra.ExactArgs(1), RunE: runRedisBackupsDelete,
	}
	redisBackupsListCmd = &cobra.Command{
		Use: "list", Short: "List backups in a collection",
		Args: cobra.NoArgs, RunE: runRedisBackupsList,
	}
	redisBackupsExportCmd = &cobra.Command{
		Use: "export BACKUP", Short: "Export a Redis cluster backup to Cloud Storage",
		Args: cobra.ExactArgs(1), RunE: runRedisBackupsExport,
	}
)

var (
	flagRedisBCRegion    string
	flagRedisBCCollection string
	flagRedisBCFormat    string
	flagRedisBackupsGcs  string
	flagRedisBackupsAsync bool
)

func registerRedisBackupCollections(parent *cobra.Command) {
	for _, c := range []*cobra.Command{redisBCDescribeCmd, redisBCListCmd} {
		c.Flags().StringVar(&flagRedisBCRegion, "region", "", "Region (required)")
		_ = c.MarkFlagRequired("region")
	}
	redisBCDescribeCmd.Flags().StringVar(&flagRedisBCFormat, "format", "", "Output format")
	redisBCListCmd.Flags().StringVar(&flagRedisBCFormat, "format", "", "Output format")

	redisBackupCollectionsCmd.AddCommand(redisBCDescribeCmd, redisBCListCmd)
	parent.AddCommand(redisBackupCollectionsCmd)
}

func registerRedisClusterBackups(parent *cobra.Command) {
	for _, c := range []*cobra.Command{redisBackupsDescribeCmd, redisBackupsDeleteCmd, redisBackupsListCmd, redisBackupsExportCmd} {
		c.Flags().StringVar(&flagRedisBCRegion, "region", "", "Region (required)")
		c.Flags().StringVar(&flagRedisBCCollection, "backup-collection", "",
			"Backup collection ID or fully qualified name (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("backup-collection")
	}
	redisBackupsExportCmd.Flags().StringVar(&flagRedisBackupsGcs, "gcs-bucket", "",
		"Destination Cloud Storage bucket for export (required)")
	_ = redisBackupsExportCmd.MarkFlagRequired("gcs-bucket")
	for _, c := range []*cobra.Command{redisBackupsDeleteCmd, redisBackupsExportCmd} {
		c.Flags().BoolVar(&flagRedisBackupsAsync, "async", false, "Return the long-running operation without waiting")
	}
	redisBackupsDescribeCmd.Flags().StringVar(&flagRedisBCFormat, "format", "", "Output format")
	redisBackupsListCmd.Flags().StringVar(&flagRedisBCFormat, "format", "", "Output format")

	redisClusterBackupsCmd.AddCommand(redisBackupsDeleteCmd, redisBackupsDescribeCmd, redisBackupsExportCmd, redisBackupsListCmd)
	parent.AddCommand(redisClusterBackupsCmd)
}

func redisBackupCollectionName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/backupCollections/%s", redisLocationParent(project, region), id)
}

func redisBackupName(id, project, region, collection string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/backups/%s", redisBackupCollectionName(collection, project, region), id)
}

func runRedisBCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	bc, err := svc.Projects.Locations.BackupCollections.Get(redisBackupCollectionName(args[0], project, flagRedisBCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup collection: %w", err)
	}
	return emitFormatted(bc, flagRedisBCFormat)
}

func runRedisBCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.BackupCollections.List(redisLocationParent(project, flagRedisBCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backup collections: %w", err)
	}
	if flagRedisBCFormat != "" {
		return emitFormatted(resp.BackupCollections, flagRedisBCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "CLUSTER")
	for _, bc := range resp.BackupCollections {
		fmt.Printf("%-40s %s\n", path.Base(bc.Name), bc.Cluster)
	}
	return nil
}

func runRedisBackupsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	b, err := svc.Projects.Locations.BackupCollections.Backups.Get(redisBackupName(args[0], project, flagRedisBCRegion, flagRedisBCCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(b, flagRedisBCFormat)
}

func runRedisBackupsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupCollections.Backups.Delete(redisBackupName(args[0], project, flagRedisBCRegion, flagRedisBCCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Delete backup", args[0], flagRedisBackupsAsync)
}

func runRedisBackupsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.BackupCollections.Backups.List(redisBackupCollectionName(flagRedisBCCollection, project, flagRedisBCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if flagRedisBCFormat != "" {
		return emitFormatted(resp.Backups, flagRedisBCFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "CREATE_TIME")
	for _, b := range resp.Backups {
		fmt.Printf("%-40s %-15s %s\n", path.Base(b.Name), b.State, b.CreateTime)
	}
	return nil
}

func runRedisBackupsExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &redis.ExportBackupRequest{GcsBucket: flagRedisBackupsGcs}
	op, err := svc.Projects.Locations.BackupCollections.Backups.Export(redisBackupName(args[0], project, flagRedisBCRegion, flagRedisBCCollection), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting backup: %w", err)
	}
	return redisFinishOp(ctx, svc, op, "Export backup", args[0], flagRedisBackupsAsync)
}

// --- redis operations ---

var redisOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Redis operations",
}

var (
	redisOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Redis operation",
		Args: cobra.ExactArgs(1), RunE: runRedisOpCancel,
	}
	redisOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Redis operation",
		Args: cobra.ExactArgs(1), RunE: runRedisOpDescribe,
	}
	redisOpListCmd = &cobra.Command{
		Use: "list", Short: "List Redis operations in a region",
		Args: cobra.NoArgs, RunE: runRedisOpList,
	}
)

var (
	flagRedisOpRegion string
	flagRedisOpFormat string
	flagRedisOpFilter string
)

func init() {
	for _, c := range []*cobra.Command{redisOpCancelCmd, redisOpDescribeCmd, redisOpListCmd} {
		c.Flags().StringVar(&flagRedisOpRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	redisOpDescribeCmd.Flags().StringVar(&flagRedisOpFormat, "format", "", "Output format")
	redisOpListCmd.Flags().StringVar(&flagRedisOpFormat, "format", "", "Output format")
	redisOpListCmd.Flags().StringVar(&flagRedisOpFilter, "filter", "", "Server-side filter expression")

	redisOperationsCmd.AddCommand(redisOpCancelCmd, redisOpDescribeCmd, redisOpListCmd)
	redisCmd.AddCommand(redisOperationsCmd)
}

func redisOpName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/operations/%s", redisLocationParent(project, region), id)
}

func runRedisOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(redisOpName(args[0], project, flagRedisOpRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runRedisOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(redisOpName(args[0], project, flagRedisOpRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagRedisOpFormat)
}

func runRedisOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Operations.List(redisLocationParent(project, flagRedisOpRegion)).Context(ctx)
	if flagRedisOpFilter != "" {
		call = call.Filter(flagRedisOpFilter)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagRedisOpFormat != "" {
		return emitFormatted(resp.Operations, flagRedisOpFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- redis regions + zones (both derived from Locations.List) ---

var redisRegionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "Explore Cloud Memorystore Redis regions",
}

var redisRegionsListCmd = &cobra.Command{
	Use: "list", Short: "List Redis regions available to the project",
	Args: cobra.NoArgs, RunE: runRedisRegionsList,
}

var redisRegionsDescribeCmd = &cobra.Command{
	Use: "describe REGION", Short: "Describe a Redis region",
	Args: cobra.ExactArgs(1), RunE: runRedisRegionsDescribe,
}

var redisZonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Explore Cloud Memorystore Redis zones",
}

var redisZonesListCmd = &cobra.Command{
	Use: "list", Short: "List Redis zones available across all regions",
	Args: cobra.NoArgs, RunE: runRedisZonesList,
}

var (
	flagRedisRZFormat string
)

func init() {
	redisRegionsListCmd.Flags().StringVar(&flagRedisRZFormat, "format", "", "Output format")
	redisRegionsDescribeCmd.Flags().StringVar(&flagRedisRZFormat, "format", "", "Output format")
	redisZonesListCmd.Flags().StringVar(&flagRedisRZFormat, "format", "", "Output format")

	redisRegionsCmd.AddCommand(redisRegionsDescribeCmd, redisRegionsListCmd)
	redisZonesCmd.AddCommand(redisZonesListCmd)
	redisCmd.AddCommand(redisRegionsCmd, redisZonesCmd)
}

func runRedisRegionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing regions: %w", err)
	}
	if flagRedisRZFormat != "" {
		return emitFormatted(resp.Locations, flagRedisRZFormat)
	}
	fmt.Printf("%-20s %s\n", "REGION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func runRedisRegionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(redisLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing region: %w", err)
	}
	return emitFormatted(loc, flagRedisRZFormat)
}

// runRedisZonesList extracts availableZones from every region's metadata
// (the redis Locations API embeds them in Location.Metadata). This matches
// gcloud-python's behaviour of aggregating zones across regions.
func runRedisZonesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing zones: %w", err)
	}
	type zone struct {
		Zone   string `json:"zone"`
		Region string `json:"region"`
	}
	var zones []zone
	for _, loc := range resp.Locations {
		if len(loc.Metadata) == 0 {
			continue
		}
		var m struct {
			AvailableZones map[string]json.RawMessage `json:"availableZones"`
		}
		if err := json.Unmarshal(loc.Metadata, &m); err != nil {
			continue
		}
		for z := range m.AvailableZones {
			zones = append(zones, zone{Zone: z, Region: loc.LocationId})
		}
	}
	if flagRedisRZFormat != "" {
		return emitFormatted(zones, flagRedisRZFormat)
	}
	fmt.Printf("%-30s %s\n", "ZONE", "REGION")
	for _, z := range zones {
		fmt.Printf("%-30s %s\n", z.Zone, z.Region)
	}
	return nil
}
