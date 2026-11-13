package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	redis "google.golang.org/api/redis/v1"
)

// --- redis instances create (#191) ---

// --- redis instances delete (#192) ---

// --- redis instances update (#193) ---

// --- redis instances failover (#194) ---

// --- redis instances export/import (#195) ---

func (s *redisCmds) initRedis() {
	s.redisInstancesDescribeCmd.Flags().StringVar(&s.flagRedisRegion, "region", "", "Region of the Redis instance")
	s.redisInstancesListCmd.Flags().StringVar(&s.flagRedisRegion, "region", "", "Region of the Redis instances (uses '-' for all regions)")
	s.redisInstancesListCmd.Flags().StringVar(&s.flagRedisListFormat, "format", "", "Output format (json)")
	s.redisInstancesListCmd.Flags().StringVar(&s.flagRedisListFilter, "filter", "", "Filter expression")
	s.redisInstancesListCmd.Flags().Int64Var(&s.flagRedisListLimit, "limit", 0, "Maximum number of results")
	s.redisInstancesListCmd.Flags().BoolVar(&s.flagRedisListURI, "uri", false, "Print resource names")

	s.redisInstancesCreateCmd.Flags().StringVar(&s.flagRedisRegion, "region", "", "Region")
	s.redisInstancesCreateCmd.Flags().Int64Var(&s.flagRedisSize, "size", 1, "Memory size in GiB")
	s.redisInstancesCreateCmd.Flags().StringVar(&s.flagRedisTier, "tier", "BASIC", "Tier (BASIC or STANDARD)")
	s.redisInstancesCreateCmd.Flags().StringVar(&s.flagRedisZone, "zone", "", "Zone")
	s.redisInstancesCreateCmd.Flags().StringVar(&s.flagRedisVersion, "redis-version", "", "Redis version")
	s.redisInstancesCreateCmd.Flags().StringVar(&s.flagRedisNetwork, "network", "", "Network")
	s.redisInstancesCreateCmd.Flags().StringToStringVar(&s.flagRedisConfig, "redis-config", nil, "Redis config (key=value)")
	// GA in gcloud-python 584.0.0 (#1935).
	s.redisInstancesCreateCmd.Flags().StringToStringVar(&s.flagRedisTags, "tags", nil, "List of tag KEY=VALUE pairs to add")
	s.redisInstancesCreateCmd.Flags().BoolVar(&s.flagRedisEnableAuth, "enable-auth", false, "Enable AUTH")
	s.redisInstancesCreateCmd.Flags().StringVar(&s.flagRedisTransitEncrypt, "transit-encryption-mode", "", "Transit encryption mode")
	s.redisInstancesCreateCmd.Flags().StringToStringVar(&s.flagRedisLabels, "labels", nil, "Labels (key=value)")

	s.redisInstancesDeleteCmd.Flags().StringVar(&s.flagRedisRegion, "region", "", "Region")
	s.redisInstancesUpdateCmd.Flags().StringVar(&s.flagRedisRegion, "region", "", "Region")
	s.redisInstancesUpdateCmd.Flags().Int64Var(&s.flagRedisUpdateSize, "size", 0, "New memory size in GiB")
	s.redisInstancesUpdateCmd.Flags().StringToStringVar(&s.flagRedisUpdateConfig, "redis-config", nil, "Redis config")
	s.redisInstancesUpdateCmd.Flags().StringToStringVar(&s.flagRedisUpdateLabels, "update-labels", nil, "Labels to update")
	s.redisInstancesUpdateCmd.Flags().StringSliceVar(&s.flagRedisUpdateRemoveLabels, "remove-labels", nil, "Labels to remove")

	s.redisInstancesFailoverCmd.Flags().StringVar(&s.flagRedisRegion, "region", "", "Region")

	s.redisInstancesExportCmd.Flags().StringVar(&s.flagRedisRegion, "region", "", "Region")
	s.redisInstancesExportCmd.Flags().StringVar(&s.flagRedisExportURI, "output-config", "", "GCS URI for export (gs://bucket/path)")
	s.redisInstancesExportCmd.MarkFlagRequired("output-config")

	s.redisInstancesImportCmd.Flags().StringVar(&s.flagRedisRegion, "region", "", "Region")
	s.redisInstancesImportCmd.Flags().StringVar(&s.flagRedisImportURI, "input-config", "", "GCS URI for import (gs://bucket/path)")
	s.redisInstancesImportCmd.MarkFlagRequired("input-config")

	s.redisInstancesCmd.AddCommand(s.redisInstancesDescribeCmd)
	s.redisInstancesCmd.AddCommand(s.redisInstancesListCmd)
	s.redisInstancesCmd.AddCommand(s.redisInstancesCreateCmd)
	s.redisInstancesCmd.AddCommand(s.redisInstancesDeleteCmd)
	s.redisInstancesCmd.AddCommand(s.redisInstancesUpdateCmd)
	s.redisInstancesCmd.AddCommand(s.redisInstancesFailoverCmd)
	s.redisInstancesCmd.AddCommand(s.redisInstancesExportCmd)
	s.redisInstancesCmd.AddCommand(s.redisInstancesImportCmd)
	s.redisCmd.AddCommand(s.redisInstancesCmd)

	// acl-policies, clusters (+ backup-collections, backups), operations,
	// regions, and zones subgroups are implemented in redis_all.go.

}

func (s *redisCmds) runRedisInstancesDescribe(cmd *cobra.Command, args []string) error {
	instanceName := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	region := s.flagRedisRegion
	if region == "" {
		_, r, err := s.inv.region()
		if err != nil || r == "" {
			return fmt.Errorf("--region is required")
		}
		region = r
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, instanceName)
	instance, err := svc.Projects.Locations.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing redis instance: %w", err)
	}

	return s.inv.formatOutput(instance, "")
}

func (s *redisCmds) runRedisInstancesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	region := s.flagRedisRegion
	if region == "" {
		region = "-" // all regions
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	var allInstances []*redis.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing redis instances: %w", err)
		}
		allInstances = append(allInstances, resp.Instances...)
		if resp.NextPageToken == "" {
			break
		}
		if s.flagRedisListLimit > 0 && int64(len(allInstances)) >= s.flagRedisListLimit {
			break
		}
		pageToken = resp.NextPageToken
	}

	// Client-side filter.
	if s.flagRedisListFilter != "" {
		var filtered []*redis.Instance
		for _, inst := range allInstances {
			if strings.Contains(inst.Name, s.flagRedisListFilter) || strings.Contains(inst.DisplayName, s.flagRedisListFilter) {
				filtered = append(filtered, inst)
			}
		}
		allInstances = filtered
	}
	if s.flagRedisListLimit > 0 && int64(len(allInstances)) > s.flagRedisListLimit {
		allInstances = allInstances[:s.flagRedisListLimit]
	}

	if s.flagRedisListURI {
		for _, inst := range allInstances {
			fmt.Fprintln(s.inv.stdout, inst.Name)
		}
		return nil
	}

	if s.flagRedisListFormat == "json" {
		return s.inv.printJSON(allInstances)
	}

	printTableHeadingTo(s.inv.stdout, len(allInstances), "%-30s %-10s %-5s %-15s %-15s %-20s\n", "INSTANCE_NAME", "TIER", "SIZE", "REGION", "VERSION", "HOST:PORT")
	for _, inst := range allInstances {
		name := path.Base(inst.Name)
		hostPort := fmt.Sprintf("%s:%d", inst.Host, inst.Port)
		fmt.Fprintf(s.inv.stdout, "%-30s %-10s %-5d %-15s %-15s %-20s\n", name, inst.Tier, inst.MemorySizeGb, inst.LocationId, inst.RedisVersion, hostPort)
	}
	return nil
}

func (s *redisCmds) resolveRedisRegion() (string, string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", "", err
	}
	region := s.flagRedisRegion
	if region == "" {
		_, r, err := s.inv.region()
		if err != nil || r == "" {
			return "", "", fmt.Errorf("--region is required")
		}
		region = r
	}
	return project, region, nil
}

func waitForRedisOp(ctx context.Context, svc *redis.Service, opName string) error {
	deadline := time.Now().Add(30 * time.Minute)
	for {
		op, err := svc.Projects.Locations.Operations.Get(opName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation failed: %s", op.Error.Message)
			}
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for operation %s", opName)
		}
		time.Sleep(5 * time.Second)
	}
}

// --- redis instances create (#191) ---

func (s *redisCmds) runRedisInstancesCreate(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}

	inst := &redis.Instance{
		MemorySizeGb: s.flagRedisSize,
		Tier:         strings.ToUpper(s.flagRedisTier),
	}
	if s.flagRedisZone != "" {
		inst.LocationId = s.flagRedisZone
	}
	if s.flagRedisVersion != "" {
		inst.RedisVersion = s.flagRedisVersion
	}
	if s.flagRedisNetwork != "" {
		inst.AuthorizedNetwork = s.flagRedisNetwork
	}
	if len(s.flagRedisConfig) > 0 {
		inst.RedisConfigs = s.flagRedisConfig
	}
	if s.flagRedisEnableAuth {
		inst.AuthEnabled = true
	}
	if s.flagRedisTransitEncrypt != "" {
		inst.TransitEncryptionMode = s.flagRedisTransitEncrypt
	}
	if len(s.flagRedisTags) > 0 {
		inst.Tags = s.flagRedisTags
	}
	if len(s.flagRedisLabels) > 0 {
		inst.Labels = s.flagRedisLabels
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	op, err := svc.Projects.Locations.Instances.Create(parent, inst).InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating redis instance: %w", err)
	}

	fmt.Fprintf(s.inv.stderr, "Creating Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Created Redis instance [%s].\n", args[0])
	return nil
}

// --- redis instances delete (#192) ---

func (s *redisCmds) runRedisInstancesDelete(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveRedisRegion()
	if err != nil {
		return err
	}

	if !s.inv.flags.quiet {
		fmt.Fprintf(s.inv.stdout, "You are about to delete Redis instance [%s]. This action cannot be undone.\n", args[0])
		fmt.Fprint(s.inv.stdout, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(s.inv.stdout, "Aborted.")
			return nil
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])
	op, err := svc.Projects.Locations.Instances.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting redis instance: %w", err)
	}

	fmt.Fprintf(s.inv.stderr, "Deleting Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Deleted Redis instance [%s].\n", args[0])
	return nil
}

// --- redis instances update (#193) ---

func (s *redisCmds) runRedisInstancesUpdate(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])

	// Get current instance.
	current, err := svc.Projects.Locations.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting redis instance: %w", err)
	}

	var updateMask []string
	if s.flagRedisUpdateSize > 0 {
		current.MemorySizeGb = s.flagRedisUpdateSize
		updateMask = append(updateMask, "memory_size_gb")
	}
	if len(s.flagRedisUpdateConfig) > 0 {
		if current.RedisConfigs == nil {
			current.RedisConfigs = make(map[string]string)
		}
		for k, v := range s.flagRedisUpdateConfig {
			current.RedisConfigs[k] = v
		}
		updateMask = append(updateMask, "redis_configs")
	}
	if len(s.flagRedisUpdateLabels) > 0 || len(s.flagRedisUpdateRemoveLabels) > 0 {
		if current.Labels == nil {
			current.Labels = make(map[string]string)
		}
		for _, k := range s.flagRedisUpdateRemoveLabels {
			delete(current.Labels, k)
		}
		for k, v := range s.flagRedisUpdateLabels {
			current.Labels[k] = v
		}
		updateMask = append(updateMask, "labels")
	}

	if len(updateMask) == 0 {
		return fmt.Errorf("no update flags specified")
	}

	op, err := svc.Projects.Locations.Instances.Patch(name, current).UpdateMask(strings.Join(updateMask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating redis instance: %w", err)
	}

	fmt.Fprintf(s.inv.stderr, "Updating Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Updated Redis instance [%s].\n", args[0])
	return nil
}

// --- redis instances failover (#194) ---

func (s *redisCmds) runRedisInstancesFailover(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])
	req := &redis.FailoverInstanceRequest{}
	op, err := svc.Projects.Locations.Instances.Failover(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failing over redis instance: %w", err)
	}

	fmt.Fprintf(s.inv.stderr, "Failing over Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Failover completed for Redis instance [%s].\n", args[0])
	return nil
}

// --- redis instances export (#195) ---

func (s *redisCmds) runRedisInstancesExport(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])
	req := &redis.ExportInstanceRequest{
		OutputConfig: &redis.OutputConfig{
			GcsDestination: &redis.GcsDestination{
				Uri: s.flagRedisExportURI,
			},
		},
	}
	op, err := svc.Projects.Locations.Instances.Export(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting redis instance: %w", err)
	}

	fmt.Fprintf(s.inv.stderr, "Exporting Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Exported Redis instance [%s] to %s.\n", args[0], s.flagRedisExportURI)
	return nil
}

// --- redis instances import (#195) ---

func (s *redisCmds) runRedisInstancesImport(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Redis(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])
	req := &redis.ImportInstanceRequest{
		InputConfig: &redis.InputConfig{
			GcsSource: &redis.GcsSource{
				Uri: s.flagRedisImportURI,
			},
		},
	}
	op, err := svc.Projects.Locations.Instances.Import(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing to redis instance: %w", err)
	}

	fmt.Fprintf(s.inv.stderr, "Importing to Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Imported data to Redis instance [%s] from %s.\n", args[0], s.flagRedisImportURI)
	return nil
}

// redisCmds holds the commands and flag values of one invocation's "redis"
// command group, so invocations don't share state (#2070).
type redisCmds struct {
	inv *invocation

	redisCmd                       *cobra.Command
	redisInstancesCmd              *cobra.Command
	redisInstancesDescribeCmd      *cobra.Command
	redisInstancesListCmd          *cobra.Command
	redisInstancesCreateCmd        *cobra.Command
	flagRedisSize                  int64
	flagRedisTier                  string
	flagRedisZone                  string
	flagRedisVersion               string
	flagRedisNetwork               string
	flagRedisConfig                map[string]string
	flagRedisTags                  map[string]string
	flagRedisEnableAuth            bool
	flagRedisTransitEncrypt        string
	flagRedisLabels                map[string]string
	redisInstancesDeleteCmd        *cobra.Command
	redisInstancesUpdateCmd        *cobra.Command
	flagRedisUpdateSize            int64
	flagRedisUpdateConfig          map[string]string
	flagRedisUpdateLabels          map[string]string
	flagRedisUpdateRemoveLabels    []string
	redisInstancesFailoverCmd      *cobra.Command
	redisInstancesExportCmd        *cobra.Command
	flagRedisExportURI             string
	redisInstancesImportCmd        *cobra.Command
	flagRedisImportURI             string
	flagRedisRegion                string
	flagRedisListFormat            string
	flagRedisListFilter            string
	flagRedisListLimit             int64
	flagRedisListURI               bool
	redisAclPoliciesCmd            *cobra.Command
	redisACLCreateCmd              *cobra.Command
	redisACLDeleteCmd              *cobra.Command
	redisACLDescribeCmd            *cobra.Command
	redisACLListCmd                *cobra.Command
	redisACLUpdateCmd              *cobra.Command
	flagRedisACLRegion             string
	flagRedisACLConfigFile         string
	flagRedisACLUpdateMask         string
	flagRedisACLFormat             string
	flagRedisACLAsync              bool
	redisClustersCmd               *cobra.Command
	redisClusterCreateCmd          *cobra.Command
	redisClusterDeleteCmd          *cobra.Command
	redisClusterDescribeCmd        *cobra.Command
	redisClusterListCmd            *cobra.Command
	redisClusterUpdateCmd          *cobra.Command
	redisClusterCreateBackupCmd    *cobra.Command
	redisClusterGetCACmd           *cobra.Command
	redisClusterGetSharedCACmd     *cobra.Command
	redisClusterRescheduleMaintCmd *cobra.Command
	redisClusterAddTokenAuthCmd    *cobra.Command
	flagRedisClusterRegion         string
	flagRedisClusterConfigFile     string
	flagRedisClusterUpdateMask     string
	flagRedisClusterFormat         string
	flagRedisClusterAsync          bool
	flagRedisClusterBackupID       string
	flagRedisClusterBackupTTL      string
	flagRedisClusterMaintTime      string
	flagRedisClusterMaintType      string
	flagRedisClusterTokenAuthUser  string
	flagRedisClusterZDCZones       []string
	redisBackupCollectionsCmd      *cobra.Command
	redisBCDescribeCmd             *cobra.Command
	redisBCListCmd                 *cobra.Command
	redisClusterBackupsCmd         *cobra.Command
	redisBackupsDescribeCmd        *cobra.Command
	redisBackupsDeleteCmd          *cobra.Command
	redisBackupsListCmd            *cobra.Command
	redisBackupsExportCmd          *cobra.Command
	flagRedisBCRegion              string
	flagRedisBCCollection          string
	flagRedisBCFormat              string
	flagRedisBackupsGcs            string
	flagRedisBackupsAsync          bool
	redisOperationsCmd             *cobra.Command
	redisOpCancelCmd               *cobra.Command
	redisOpDescribeCmd             *cobra.Command
	redisOpListCmd                 *cobra.Command
	flagRedisOpRegion              string
	flagRedisOpFormat              string
	flagRedisOpFilter              string
	redisRegionsCmd                *cobra.Command
	redisRegionsListCmd            *cobra.Command
	redisRegionsDescribeCmd        *cobra.Command
	redisZonesCmd                  *cobra.Command
	redisZonesListCmd              *cobra.Command
	flagRedisRZFormat              string
}

// newRedisCmd returns the "redis" command group for inv.
func newRedisCmd(inv *invocation) *cobra.Command {
	return newRedisCmds(inv).redisCmd
}

// newRedisCmds builds the "redis" command group for inv and returns its state.
func newRedisCmds(inv *invocation) *redisCmds {
	s := &redisCmds{inv: inv}
	s.redisCmd = &cobra.Command{
		Use:   "redis",
		Short: "Manage Cloud Memorystore for Redis",
	}
	s.redisInstancesCmd = &cobra.Command{
		Use:   "instances",
		Short: "Manage Redis instances",
	}
	s.redisInstancesDescribeCmd = &cobra.Command{
		Use:   "describe INSTANCE_NAME",
		Short: "Describe a Redis instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRedisInstancesDescribe,
	}
	s.redisInstancesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Redis instances",
		Args:  cobra.NoArgs,
		RunE:  s.runRedisInstancesList,
	}
	s.redisInstancesCreateCmd = &cobra.Command{
		Use:   "create INSTANCE_NAME",
		Short: "Create a Redis instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRedisInstancesCreate,
	}
	s.redisInstancesDeleteCmd = &cobra.Command{
		Use:   "delete INSTANCE_NAME",
		Short: "Delete a Redis instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRedisInstancesDelete,
	}
	s.redisInstancesUpdateCmd = &cobra.Command{
		Use:   "update INSTANCE_NAME",
		Short: "Update a Redis instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRedisInstancesUpdate,
	}
	s.redisInstancesFailoverCmd = &cobra.Command{
		Use:   "failover INSTANCE_NAME",
		Short: "Failover a Redis instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRedisInstancesFailover,
	}
	s.redisInstancesExportCmd = &cobra.Command{
		Use:   "export INSTANCE_NAME",
		Short: "Export Redis instance data to GCS",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRedisInstancesExport,
	}
	s.redisInstancesImportCmd = &cobra.Command{
		Use:   "import INSTANCE_NAME",
		Short: "Import data to a Redis instance from GCS",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRedisInstancesImport,
	}
	s.redisAclPoliciesCmd = &cobra.Command{
		Use:   "acl-policies",
		Short: "Manage Redis Cluster ACL policies",
	}
	s.redisACLCreateCmd = &cobra.Command{
		Use: "create POLICY", Short: "Create an ACL policy from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runRedisACLCreate,
	}
	s.redisACLDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete an ACL policy",
		Args: cobra.ExactArgs(1), RunE: s.runRedisACLDelete,
	}
	s.redisACLDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe an ACL policy",
		Args: cobra.ExactArgs(1), RunE: s.runRedisACLDescribe,
	}
	s.redisACLListCmd = &cobra.Command{
		Use: "list", Short: "List ACL policies in a region",
		Args: cobra.NoArgs, RunE: s.runRedisACLList,
	}
	s.redisACLUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update an ACL policy from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runRedisACLUpdate,
	}
	s.redisClustersCmd = &cobra.Command{
		Use:   "clusters",
		Short: "Manage Memorystore for Redis Cluster instances",
	}
	s.redisClusterCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a Redis cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runRedisClusterCreate,
	}
	s.redisClusterDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: s.runRedisClusterDelete,
	}
	s.redisClusterDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: s.runRedisClusterDescribe,
	}
	s.redisClusterListCmd = &cobra.Command{
		Use: "list", Short: "List Redis clusters in a region",
		Args: cobra.NoArgs, RunE: s.runRedisClusterList,
	}
	s.redisClusterUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a Redis cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runRedisClusterUpdate,
	}
	s.redisClusterCreateBackupCmd = &cobra.Command{
		Use: "create-backup CLUSTER", Short: "Create an on-demand backup of a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: s.runRedisClusterBackup,
	}
	s.redisClusterGetCACmd = &cobra.Command{
		Use: "get-cluster-certificate-authority CLUSTER", Short: "Get the CA certificates for a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: s.runRedisClusterGetCA,
	}
	s.redisClusterGetSharedCACmd = &cobra.Command{
		Use: "get-shared-regional-certificate-authority", Short: "Get the shared regional CA for Redis clusters",
		Args: cobra.NoArgs, RunE: s.runRedisClusterGetSharedCA,
	}
	s.redisClusterRescheduleMaintCmd = &cobra.Command{
		Use: "reschedule-maintenance CLUSTER", Short: "Reschedule an upcoming maintenance window",
		Args: cobra.ExactArgs(1), RunE: s.runRedisClusterRescheduleMaintenance,
	}
	s.redisClusterAddTokenAuthCmd = &cobra.Command{
		Use: "add-token-auth-user CLUSTER", Short: "Add a token-auth user to a Redis cluster",
		Args: cobra.ExactArgs(1), RunE: s.runRedisClusterAddTokenAuthUser,
	}
	s.redisBackupCollectionsCmd = &cobra.Command{
		Use:   "backup-collections",
		Short: "Manage backup collections for Redis clusters",
	}
	s.redisBCDescribeCmd = &cobra.Command{
		Use: "describe COLLECTION", Short: "Describe a backup collection",
		Args: cobra.ExactArgs(1), RunE: s.runRedisBCDescribe,
	}
	s.redisBCListCmd = &cobra.Command{
		Use: "list", Short: "List backup collections in a region",
		Args: cobra.NoArgs, RunE: s.runRedisBCList,
	}
	s.redisClusterBackupsCmd = &cobra.Command{
		Use:   "backups",
		Short: "Manage backups for a Redis cluster",
	}
	s.redisBackupsDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a Redis cluster backup",
		Args: cobra.ExactArgs(1), RunE: s.runRedisBackupsDescribe,
	}
	s.redisBackupsDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a Redis cluster backup",
		Args: cobra.ExactArgs(1), RunE: s.runRedisBackupsDelete,
	}
	s.redisBackupsListCmd = &cobra.Command{
		Use: "list", Short: "List backups in a collection",
		Args: cobra.NoArgs, RunE: s.runRedisBackupsList,
	}
	s.redisBackupsExportCmd = &cobra.Command{
		Use: "export BACKUP", Short: "Export a Redis cluster backup to Cloud Storage",
		Args: cobra.ExactArgs(1), RunE: s.runRedisBackupsExport,
	}
	s.redisOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Redis operations",
	}
	s.redisOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Redis operation",
		Args: cobra.ExactArgs(1), RunE: s.runRedisOpCancel,
	}
	s.redisOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Redis operation",
		Args: cobra.ExactArgs(1), RunE: s.runRedisOpDescribe,
	}
	s.redisOpListCmd = &cobra.Command{
		Use: "list", Short: "List Redis operations in a region",
		Args: cobra.NoArgs, RunE: s.runRedisOpList,
	}
	s.redisRegionsCmd = &cobra.Command{
		Use:   "regions",
		Short: "Explore Cloud Memorystore Redis regions",
	}
	s.redisRegionsListCmd = &cobra.Command{
		Use: "list", Short: "List Redis regions available to the project",
		Args: cobra.NoArgs, RunE: s.runRedisRegionsList,
	}
	s.redisRegionsDescribeCmd = &cobra.Command{
		Use: "describe REGION", Short: "Describe a Redis region",
		Args: cobra.ExactArgs(1), RunE: s.runRedisRegionsDescribe,
	}
	s.redisZonesCmd = &cobra.Command{
		Use:   "zones",
		Short: "Explore Cloud Memorystore Redis zones",
	}
	s.redisZonesListCmd = &cobra.Command{
		Use: "list", Short: "List Redis zones available across all regions",
		Args: cobra.NoArgs, RunE: s.runRedisZonesList,
	}
	s.initRedis()
	s.initRedisAll()
	s.initRedisAll2()
	s.initRedisAll3()
	s.initRedisAll4()
	return s
}
