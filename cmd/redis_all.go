package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

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

func (s *redisCmds) redisFinishOp(ctx context.Context, svc *redis.Service, op *redis.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := redisWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// --- redis acl-policies ---

func (s *redisCmds) initRedisAll() {
	for _, c := range []*cobra.Command{s.redisACLCreateCmd, s.redisACLDeleteCmd, s.redisACLDescribeCmd, s.redisACLListCmd, s.redisACLUpdateCmd} {
		c.Flags().StringVar(&s.flagRedisACLRegion, "region", "", "Region containing the ACL policy (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.redisACLCreateCmd, s.redisACLUpdateCmd} {
		c.Flags().StringVar(&s.flagRedisACLConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the AclPolicy message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.redisACLUpdateCmd.Flags().StringVar(&s.flagRedisACLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.redisACLCreateCmd, s.redisACLDeleteCmd, s.redisACLUpdateCmd} {
		c.Flags().BoolVar(&s.flagRedisACLAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.redisACLDescribeCmd.Flags().StringVar(&s.flagRedisACLFormat, "format", "", "Output format")
	s.redisACLListCmd.Flags().StringVar(&s.flagRedisACLFormat, "format", "", "Output format")

	s.redisAclPoliciesCmd.AddCommand(s.redisACLCreateCmd, s.redisACLDeleteCmd, s.redisACLDescribeCmd, s.redisACLListCmd, s.redisACLUpdateCmd)
	s.redisCmd.AddCommand(s.redisAclPoliciesCmd)
}

func redisACLName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/aclPolicies/%s", redisLocationParent(project, region), id)
}

func (s *redisCmds) runRedisACLCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	p := &redis.AclPolicy{}
	if err := loadYAMLOrJSONInto(s.flagRedisACLConfigFile, p); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	created, err := svc.Projects.Locations.AclPolicies.Create(redisLocationParent(project, s.flagRedisACLRegion), p).
		AclPolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ACL policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, created, "")
}

func (s *redisCmds) runRedisACLDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AclPolicies.Delete(redisACLName(args[0], project, s.flagRedisACLRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting ACL policy: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Delete ACL policy", args[0], s.flagRedisACLAsync)
}

func (s *redisCmds) runRedisACLDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Locations.AclPolicies.Get(redisACLName(args[0], project, s.flagRedisACLRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ACL policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, p, s.flagRedisACLFormat)
}

func (s *redisCmds) runRedisACLList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.AclPolicies.List(redisLocationParent(project, s.flagRedisACLRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing ACL policies: %w", err)
	}
	if s.flagRedisACLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.AclPolicies, s.flagRedisACLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.AclPolicies), "%-40s\n", "NAME")
	for _, p := range resp.AclPolicies {
		fmt.Fprintln(s.inv.stdout, path.Base(p.Name))
	}
	return nil
}

func (s *redisCmds) runRedisACLUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	p := &redis.AclPolicy{}
	if err := loadYAMLOrJSONInto(s.flagRedisACLConfigFile, p); err != nil {
		return err
	}
	mask := s.flagRedisACLUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(p))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AclPolicies.Patch(redisACLName(args[0], project, s.flagRedisACLRegion), p).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating ACL policy: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Update ACL policy", args[0], s.flagRedisACLAsync)
}

// --- redis clusters ---

func (s *redisCmds) initRedisAll2() {
	all := []*cobra.Command{
		s.redisClusterCreateCmd, s.redisClusterDeleteCmd, s.redisClusterDescribeCmd,
		s.redisClusterListCmd, s.redisClusterUpdateCmd, s.redisClusterCreateBackupCmd,
		s.redisClusterGetCACmd, s.redisClusterGetSharedCACmd, s.redisClusterRescheduleMaintCmd,
		s.redisClusterAddTokenAuthCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagRedisClusterRegion, "region", "", "Region containing the Redis cluster (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.redisClusterCreateCmd, s.redisClusterUpdateCmd} {
		c.Flags().StringVar(&s.flagRedisClusterConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Cluster message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.redisClusterUpdateCmd.Flags().StringVar(&s.flagRedisClusterUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{
		s.redisClusterCreateCmd, s.redisClusterDeleteCmd, s.redisClusterUpdateCmd,
		s.redisClusterCreateBackupCmd, s.redisClusterRescheduleMaintCmd, s.redisClusterAddTokenAuthCmd,
	} {
		c.Flags().BoolVar(&s.flagRedisClusterAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.redisClusterDescribeCmd.Flags().StringVar(&s.flagRedisClusterFormat, "format", "", "Output format")
	s.redisClusterListCmd.Flags().StringVar(&s.flagRedisClusterFormat, "format", "", "Output format")
	s.redisClusterGetCACmd.Flags().StringVar(&s.flagRedisClusterFormat, "format", "", "Output format")
	s.redisClusterGetSharedCACmd.Flags().StringVar(&s.flagRedisClusterFormat, "format", "", "Output format")

	s.redisClusterCreateBackupCmd.Flags().StringVar(&s.flagRedisClusterBackupID, "backup-id", "",
		"ID for the new backup (required)")
	s.redisClusterCreateBackupCmd.Flags().StringVar(&s.flagRedisClusterBackupTTL, "ttl", "",
		"Retention duration for the backup, e.g. 24h")
	_ = s.redisClusterCreateBackupCmd.MarkFlagRequired("backup-id")

	s.redisClusterRescheduleMaintCmd.Flags().StringVar(&s.flagRedisClusterMaintType, "reschedule-type", "IMMEDIATE",
		"Reschedule type: IMMEDIATE or SPECIFIC_TIME")
	s.redisClusterRescheduleMaintCmd.Flags().StringVar(&s.flagRedisClusterMaintTime, "schedule-time", "",
		"RFC3339 time (for --reschedule-type=SPECIFIC_TIME)")

	s.redisClusterAddTokenAuthCmd.Flags().StringVar(&s.flagRedisClusterTokenAuthUser, "user", "",
		"IAM identity of the token-auth user (required)")
	_ = s.redisClusterAddTokenAuthCmd.MarkFlagRequired("user")

	// --zone-distribution-config-zones added in gcloud-python 574.0.0.
	s.redisClusterCreateCmd.Flags().StringSliceVar(&s.flagRedisClusterZDCZones, "zone-distribution-config-zones", nil,
		"Zones for a MULTI_ZONE cluster (comma-separated); overrides zoneDistributionConfig.zones from --config-file")

	s.redisClustersCmd.AddCommand(all...)
	s.registerRedisBackupCollections(s.redisClustersCmd)
	s.registerRedisClusterBackups(s.redisClustersCmd)
	s.redisCmd.AddCommand(s.redisClustersCmd)
}

func redisClusterName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/clusters/%s", redisLocationParent(project, region), id)
}

func (s *redisCmds) runRedisClusterCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &redis.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagRedisClusterConfigFile, c); err != nil {
		return err
	}
	if len(s.flagRedisClusterZDCZones) > 0 {
		if c.ZoneDistributionConfig == nil {
			c.ZoneDistributionConfig = &redis.ZoneDistributionConfig{}
		}
		c.ZoneDistributionConfig.Zones = append([]string(nil), s.flagRedisClusterZDCZones...)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Create(redisLocationParent(project, s.flagRedisClusterRegion), c).
		ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Create cluster", args[0], s.flagRedisClusterAsync)
}

func (s *redisCmds) runRedisClusterDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(redisClusterName(args[0], project, s.flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Delete cluster", args[0], s.flagRedisClusterAsync)
}

func (s *redisCmds) runRedisClusterDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	c, err := svc.Projects.Locations.Clusters.Get(redisClusterName(args[0], project, s.flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, c, s.flagRedisClusterFormat)
}

func (s *redisCmds) runRedisClusterList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(redisLocationParent(project, s.flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if s.flagRedisClusterFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Clusters, s.flagRedisClusterFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Clusters), "%-40s %-15s %s\n", "NAME", "STATE", "SIZE_GB")
	for _, c := range resp.Clusters {
		fmt.Fprintf(s.inv.stdout, "%-40s %-15s %d\n", path.Base(c.Name), c.State, c.SizeGb)
	}
	return nil
}

func (s *redisCmds) runRedisClusterUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &redis.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagRedisClusterConfigFile, c); err != nil {
		return err
	}
	mask := s.flagRedisClusterUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Patch(redisClusterName(args[0], project, s.flagRedisClusterRegion), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Update cluster", args[0], s.flagRedisClusterAsync)
}

func (s *redisCmds) runRedisClusterBackup(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	req := &redis.BackupClusterRequest{BackupId: s.flagRedisClusterBackupID, Ttl: s.flagRedisClusterBackupTTL}
	op, err := svc.Projects.Locations.Clusters.Backup(redisClusterName(args[0], project, s.flagRedisClusterRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("backing up cluster: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Backup cluster", args[0], s.flagRedisClusterAsync)
}

func (s *redisCmds) runRedisClusterGetCA(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	ca, err := svc.Projects.Locations.Clusters.GetCertificateAuthority(redisClusterName(args[0], project, s.flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting cluster CA: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, ca, s.flagRedisClusterFormat)
}

func (s *redisCmds) runRedisClusterGetSharedCA(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	ca, err := svc.Projects.Locations.GetSharedRegionalCertificateAuthority(redisLocationParent(project, s.flagRedisClusterRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting shared regional CA: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, ca, s.flagRedisClusterFormat)
}

func (s *redisCmds) runRedisClusterRescheduleMaintenance(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	req := &redis.RescheduleClusterMaintenanceRequest{
		RescheduleType: s.flagRedisClusterMaintType,
		ScheduleTime:   s.flagRedisClusterMaintTime,
	}
	op, err := svc.Projects.Locations.Clusters.RescheduleClusterMaintenance(redisClusterName(args[0], project, s.flagRedisClusterRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rescheduling maintenance: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Reschedule maintenance", args[0], s.flagRedisClusterAsync)
}

func (s *redisCmds) runRedisClusterAddTokenAuthUser(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	req := &redis.AddTokenAuthUserRequest{TokenAuthUser: s.flagRedisClusterTokenAuthUser}
	op, err := svc.Projects.Locations.Clusters.AddTokenAuthUser(redisClusterName(args[0], project, s.flagRedisClusterRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding token auth user: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Add token auth user", args[0], s.flagRedisClusterAsync)
}

// --- redis clusters backup-collections + backups (nested) ---

func (s *redisCmds) registerRedisBackupCollections(parent *cobra.Command) {
	for _, c := range []*cobra.Command{s.redisBCDescribeCmd, s.redisBCListCmd} {
		c.Flags().StringVar(&s.flagRedisBCRegion, "region", "", "Region (required)")
		_ = c.MarkFlagRequired("region")
	}
	s.redisBCDescribeCmd.Flags().StringVar(&s.flagRedisBCFormat, "format", "", "Output format")
	s.redisBCListCmd.Flags().StringVar(&s.flagRedisBCFormat, "format", "", "Output format")

	s.redisBackupCollectionsCmd.AddCommand(s.redisBCDescribeCmd, s.redisBCListCmd)
	parent.AddCommand(s.redisBackupCollectionsCmd)
}

func (s *redisCmds) registerRedisClusterBackups(parent *cobra.Command) {
	for _, c := range []*cobra.Command{s.redisBackupsDescribeCmd, s.redisBackupsDeleteCmd, s.redisBackupsListCmd, s.redisBackupsExportCmd} {
		c.Flags().StringVar(&s.flagRedisBCRegion, "region", "", "Region (required)")
		c.Flags().StringVar(&s.flagRedisBCCollection, "backup-collection", "",
			"Backup collection ID or fully qualified name (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("backup-collection")
	}
	s.redisBackupsExportCmd.Flags().StringVar(&s.flagRedisBackupsGcs, "gcs-bucket", "",
		"Destination Cloud Storage bucket for export (required)")
	_ = s.redisBackupsExportCmd.MarkFlagRequired("gcs-bucket")
	for _, c := range []*cobra.Command{s.redisBackupsDeleteCmd, s.redisBackupsExportCmd} {
		c.Flags().BoolVar(&s.flagRedisBackupsAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.redisBackupsDescribeCmd.Flags().StringVar(&s.flagRedisBCFormat, "format", "", "Output format")
	s.redisBackupsListCmd.Flags().StringVar(&s.flagRedisBCFormat, "format", "", "Output format")

	s.redisClusterBackupsCmd.AddCommand(s.redisBackupsDeleteCmd, s.redisBackupsDescribeCmd, s.redisBackupsExportCmd, s.redisBackupsListCmd)
	parent.AddCommand(s.redisClusterBackupsCmd)
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

func (s *redisCmds) runRedisBCDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	bc, err := svc.Projects.Locations.BackupCollections.Get(redisBackupCollectionName(args[0], project, s.flagRedisBCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup collection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, bc, s.flagRedisBCFormat)
}

func (s *redisCmds) runRedisBCList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.BackupCollections.List(redisLocationParent(project, s.flagRedisBCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backup collections: %w", err)
	}
	if s.flagRedisBCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.BackupCollections, s.flagRedisBCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.BackupCollections), "%-40s %s\n", "NAME", "CLUSTER")
	for _, bc := range resp.BackupCollections {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(bc.Name), bc.Cluster)
	}
	return nil
}

func (s *redisCmds) runRedisBackupsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	b, err := svc.Projects.Locations.BackupCollections.Backups.Get(redisBackupName(args[0], project, s.flagRedisBCRegion, s.flagRedisBCCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, b, s.flagRedisBCFormat)
}

func (s *redisCmds) runRedisBackupsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupCollections.Backups.Delete(redisBackupName(args[0], project, s.flagRedisBCRegion, s.flagRedisBCCollection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Delete backup", args[0], s.flagRedisBackupsAsync)
}

func (s *redisCmds) runRedisBackupsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.BackupCollections.Backups.List(redisBackupCollectionName(s.flagRedisBCCollection, project, s.flagRedisBCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if s.flagRedisBCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Backups, s.flagRedisBCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Backups), "%-40s %-15s %s\n", "NAME", "STATE", "CREATE_TIME")
	for _, b := range resp.Backups {
		fmt.Fprintf(s.inv.stdout, "%-40s %-15s %s\n", path.Base(b.Name), b.State, b.CreateTime)
	}
	return nil
}

func (s *redisCmds) runRedisBackupsExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	req := &redis.ExportBackupRequest{GcsBucket: s.flagRedisBackupsGcs}
	op, err := svc.Projects.Locations.BackupCollections.Backups.Export(redisBackupName(args[0], project, s.flagRedisBCRegion, s.flagRedisBCCollection), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting backup: %w", err)
	}
	return s.redisFinishOp(ctx, svc, op, "Export backup", args[0], s.flagRedisBackupsAsync)
}

// --- redis operations ---

func (s *redisCmds) initRedisAll3() {
	for _, c := range []*cobra.Command{s.redisOpCancelCmd, s.redisOpDescribeCmd, s.redisOpListCmd} {
		c.Flags().StringVar(&s.flagRedisOpRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	s.redisOpDescribeCmd.Flags().StringVar(&s.flagRedisOpFormat, "format", "", "Output format")
	s.redisOpListCmd.Flags().StringVar(&s.flagRedisOpFormat, "format", "", "Output format")
	s.redisOpListCmd.Flags().StringVar(&s.flagRedisOpFilter, "filter", "", "Server-side filter expression")

	s.redisOperationsCmd.AddCommand(s.redisOpCancelCmd, s.redisOpDescribeCmd, s.redisOpListCmd)
	s.redisCmd.AddCommand(s.redisOperationsCmd)
}

func redisOpName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/operations/%s", redisLocationParent(project, region), id)
}

func (s *redisCmds) runRedisOpCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(redisOpName(args[0], project, s.flagRedisOpRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *redisCmds) runRedisOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(redisOpName(args[0], project, s.flagRedisOpRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagRedisOpFormat)
}

func (s *redisCmds) runRedisOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Operations.List(redisLocationParent(project, s.flagRedisOpRegion)).Context(ctx)
	if s.flagRedisOpFilter != "" {
		call = call.Filter(s.flagRedisOpFilter)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagRedisOpFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagRedisOpFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- redis regions + zones (both derived from Locations.List) ---

func (s *redisCmds) initRedisAll4() {
	s.redisRegionsListCmd.Flags().StringVar(&s.flagRedisRZFormat, "format", "", "Output format")
	s.redisRegionsDescribeCmd.Flags().StringVar(&s.flagRedisRZFormat, "format", "", "Output format")
	s.redisZonesListCmd.Flags().StringVar(&s.flagRedisRZFormat, "format", "", "Output format")

	s.redisRegionsCmd.AddCommand(s.redisRegionsDescribeCmd, s.redisRegionsListCmd)
	s.redisZonesCmd.AddCommand(s.redisZonesListCmd)
	s.redisCmd.AddCommand(s.redisRegionsCmd, s.redisZonesCmd)
}

func (s *redisCmds) runRedisRegionsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing regions: %w", err)
	}
	if s.flagRedisRZFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Locations, s.flagRedisRZFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Locations), "%-20s %s\n", "REGION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(s.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func (s *redisCmds) runRedisRegionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(redisLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing region: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, loc, s.flagRedisRZFormat)
}

// runRedisZonesList extracts availableZones from every region's metadata
// (the redis Locations API embeds them in Location.Metadata). This matches
// gcloud-python's behaviour of aggregating zones across regions.
func (s *redisCmds) runRedisZonesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
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
	if s.flagRedisRZFormat != "" {
		return emitFormattedTo(s.inv.stdout, zones, s.flagRedisRZFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(zones), "%-30s %s\n", "ZONE", "REGION")
	for _, z := range zones {
		fmt.Fprintf(s.inv.stdout, "%-30s %s\n", z.Zone, z.Region)
	}
	return nil
}
