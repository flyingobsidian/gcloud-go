package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	redis "google.golang.org/api/redis/v1"
)

var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Manage Cloud Memorystore for Redis",
}

var redisInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage Redis instances",
}

var redisInstancesDescribeCmd = &cobra.Command{
	Use:   "describe INSTANCE_NAME",
	Short: "Describe a Redis instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runRedisInstancesDescribe,
}

var redisInstancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Redis instances",
	Args:  cobra.NoArgs,
	RunE:  runRedisInstancesList,
}

// --- redis instances create (#191) ---

var redisInstancesCreateCmd = &cobra.Command{
	Use:   "create INSTANCE_NAME",
	Short: "Create a Redis instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runRedisInstancesCreate,
}

var (
	flagRedisSize           int64
	flagRedisTier           string
	flagRedisZone           string
	flagRedisVersion        string
	flagRedisNetwork        string
	flagRedisConfig         map[string]string
	flagRedisEnableAuth     bool
	flagRedisTransitEncrypt string
	flagRedisLabels         map[string]string
)

// --- redis instances delete (#192) ---

var redisInstancesDeleteCmd = &cobra.Command{
	Use:   "delete INSTANCE_NAME",
	Short: "Delete a Redis instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runRedisInstancesDelete,
}


// --- redis instances update (#193) ---

var redisInstancesUpdateCmd = &cobra.Command{
	Use:   "update INSTANCE_NAME",
	Short: "Update a Redis instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runRedisInstancesUpdate,
}

var (
	flagRedisUpdateSize         int64
	flagRedisUpdateConfig       map[string]string
	flagRedisUpdateLabels       map[string]string
	flagRedisUpdateRemoveLabels []string
)

// --- redis instances failover (#194) ---

var redisInstancesFailoverCmd = &cobra.Command{
	Use:   "failover INSTANCE_NAME",
	Short: "Failover a Redis instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runRedisInstancesFailover,
}

// --- redis instances export/import (#195) ---

var redisInstancesExportCmd = &cobra.Command{
	Use:   "export INSTANCE_NAME",
	Short: "Export Redis instance data to GCS",
	Args:  cobra.ExactArgs(1),
	RunE:  runRedisInstancesExport,
}

var flagRedisExportURI string

var redisInstancesImportCmd = &cobra.Command{
	Use:   "import INSTANCE_NAME",
	Short: "Import data to a Redis instance from GCS",
	Args:  cobra.ExactArgs(1),
	RunE:  runRedisInstancesImport,
}

var flagRedisImportURI string

var (
	flagRedisRegion     string
	flagRedisListFormat string
	flagRedisListFilter string
	flagRedisListLimit  int64
	flagRedisListURI    bool
)

func init() {
	redisInstancesDescribeCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region of the Redis instance")
	redisInstancesListCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region of the Redis instances (uses '-' for all regions)")
	redisInstancesListCmd.Flags().StringVar(&flagRedisListFormat, "format", "", "Output format (json)")
	redisInstancesListCmd.Flags().StringVar(&flagRedisListFilter, "filter", "", "Filter expression")
	redisInstancesListCmd.Flags().Int64Var(&flagRedisListLimit, "limit", 0, "Maximum number of results")
	redisInstancesListCmd.Flags().BoolVar(&flagRedisListURI, "uri", false, "Print resource names")

	redisInstancesCreateCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region")
	redisInstancesCreateCmd.Flags().Int64Var(&flagRedisSize, "size", 1, "Memory size in GiB")
	redisInstancesCreateCmd.Flags().StringVar(&flagRedisTier, "tier", "BASIC", "Tier (BASIC or STANDARD)")
	redisInstancesCreateCmd.Flags().StringVar(&flagRedisZone, "zone", "", "Zone")
	redisInstancesCreateCmd.Flags().StringVar(&flagRedisVersion, "redis-version", "", "Redis version")
	redisInstancesCreateCmd.Flags().StringVar(&flagRedisNetwork, "network", "", "Network")
	redisInstancesCreateCmd.Flags().StringToStringVar(&flagRedisConfig, "redis-config", nil, "Redis config (key=value)")
	redisInstancesCreateCmd.Flags().BoolVar(&flagRedisEnableAuth, "enable-auth", false, "Enable AUTH")
	redisInstancesCreateCmd.Flags().StringVar(&flagRedisTransitEncrypt, "transit-encryption-mode", "", "Transit encryption mode")
	redisInstancesCreateCmd.Flags().StringToStringVar(&flagRedisLabels, "labels", nil, "Labels (key=value)")

	redisInstancesDeleteCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region")
	redisInstancesUpdateCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region")
	redisInstancesUpdateCmd.Flags().Int64Var(&flagRedisUpdateSize, "size", 0, "New memory size in GiB")
	redisInstancesUpdateCmd.Flags().StringToStringVar(&flagRedisUpdateConfig, "redis-config", nil, "Redis config")
	redisInstancesUpdateCmd.Flags().StringToStringVar(&flagRedisUpdateLabels, "update-labels", nil, "Labels to update")
	redisInstancesUpdateCmd.Flags().StringSliceVar(&flagRedisUpdateRemoveLabels, "remove-labels", nil, "Labels to remove")

	redisInstancesFailoverCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region")

	redisInstancesExportCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region")
	redisInstancesExportCmd.Flags().StringVar(&flagRedisExportURI, "output-config", "", "GCS URI for export (gs://bucket/path)")
	redisInstancesExportCmd.MarkFlagRequired("output-config")

	redisInstancesImportCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region")
	redisInstancesImportCmd.Flags().StringVar(&flagRedisImportURI, "input-config", "", "GCS URI for import (gs://bucket/path)")
	redisInstancesImportCmd.MarkFlagRequired("input-config")

	redisInstancesCmd.AddCommand(redisInstancesDescribeCmd)
	redisInstancesCmd.AddCommand(redisInstancesListCmd)
	redisInstancesCmd.AddCommand(redisInstancesCreateCmd)
	redisInstancesCmd.AddCommand(redisInstancesDeleteCmd)
	redisInstancesCmd.AddCommand(redisInstancesUpdateCmd)
	redisInstancesCmd.AddCommand(redisInstancesFailoverCmd)
	redisInstancesCmd.AddCommand(redisInstancesExportCmd)
	redisInstancesCmd.AddCommand(redisInstancesImportCmd)
	redisCmd.AddCommand(redisInstancesCmd)

	// acl-policies, clusters (+ backup-collections, backups), operations,
	// regions, and zones subgroups are implemented in redis_all.go.
	rootCmd.AddCommand(redisCmd)
}

func runRedisInstancesDescribe(cmd *cobra.Command, args []string) error {
	instanceName := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	region := flagRedisRegion
	if region == "" {
		_, r, err := resolveRegion()
		if err != nil || r == "" {
			return fmt.Errorf("--region is required")
		}
		region = r
	}

	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, instanceName)
	instance, err := svc.Projects.Locations.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing redis instance: %w", err)
	}

	return formatOutput(instance, "")
}

func runRedisInstancesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	region := flagRedisRegion
	if region == "" {
		region = "-" // all regions
	}

	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
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
		if flagRedisListLimit > 0 && int64(len(allInstances)) >= flagRedisListLimit {
			break
		}
		pageToken = resp.NextPageToken
	}

	// Client-side filter.
	if flagRedisListFilter != "" {
		var filtered []*redis.Instance
		for _, inst := range allInstances {
			if strings.Contains(inst.Name, flagRedisListFilter) || strings.Contains(inst.DisplayName, flagRedisListFilter) {
				filtered = append(filtered, inst)
			}
		}
		allInstances = filtered
	}
	if flagRedisListLimit > 0 && int64(len(allInstances)) > flagRedisListLimit {
		allInstances = allInstances[:flagRedisListLimit]
	}

	if flagRedisListURI {
		for _, inst := range allInstances {
			fmt.Println(inst.Name)
		}
		return nil
	}

	if flagRedisListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allInstances)
	}

	fmt.Printf("%-30s %-10s %-5s %-15s %-15s %-20s\n", "INSTANCE_NAME", "TIER", "SIZE", "REGION", "VERSION", "HOST:PORT")
	for _, inst := range allInstances {
		name := path.Base(inst.Name)
		hostPort := fmt.Sprintf("%s:%d", inst.Host, inst.Port)
		fmt.Printf("%-30s %-10s %-5d %-15s %-15s %-20s\n", name, inst.Tier, inst.MemorySizeGb, inst.LocationId, inst.RedisVersion, hostPort)
	}
	return nil
}

func resolveRedisRegion() (string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", err
	}
	region := flagRedisRegion
	if region == "" {
		_, r, err := resolveRegion()
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

func runRedisInstancesCreate(cmd *cobra.Command, args []string) error {
	project, region, err := resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}

	inst := &redis.Instance{
		MemorySizeGb: flagRedisSize,
		Tier:         strings.ToUpper(flagRedisTier),
	}
	if flagRedisZone != "" {
		inst.LocationId = flagRedisZone
	}
	if flagRedisVersion != "" {
		inst.RedisVersion = flagRedisVersion
	}
	if flagRedisNetwork != "" {
		inst.AuthorizedNetwork = flagRedisNetwork
	}
	if len(flagRedisConfig) > 0 {
		inst.RedisConfigs = flagRedisConfig
	}
	if flagRedisEnableAuth {
		inst.AuthEnabled = true
	}
	if flagRedisTransitEncrypt != "" {
		inst.TransitEncryptionMode = flagRedisTransitEncrypt
	}
	if len(flagRedisLabels) > 0 {
		inst.Labels = flagRedisLabels
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	op, err := svc.Projects.Locations.Instances.Create(parent, inst).InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating redis instance: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Creating Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Printf("Created Redis instance [%s].\n", args[0])
	return nil
}

// --- redis instances delete (#192) ---

func runRedisInstancesDelete(cmd *cobra.Command, args []string) error {
	project, region, err := resolveRedisRegion()
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Printf("You are about to delete Redis instance [%s]. This action cannot be undone.\n", args[0])
		fmt.Print("Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])
	op, err := svc.Projects.Locations.Instances.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting redis instance: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Deleting Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Printf("Deleted Redis instance [%s].\n", args[0])
	return nil
}

// --- redis instances update (#193) ---

func runRedisInstancesUpdate(cmd *cobra.Command, args []string) error {
	project, region, err := resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
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
	if flagRedisUpdateSize > 0 {
		current.MemorySizeGb = flagRedisUpdateSize
		updateMask = append(updateMask, "memory_size_gb")
	}
	if len(flagRedisUpdateConfig) > 0 {
		if current.RedisConfigs == nil {
			current.RedisConfigs = make(map[string]string)
		}
		for k, v := range flagRedisUpdateConfig {
			current.RedisConfigs[k] = v
		}
		updateMask = append(updateMask, "redis_configs")
	}
	if len(flagRedisUpdateLabels) > 0 || len(flagRedisUpdateRemoveLabels) > 0 {
		if current.Labels == nil {
			current.Labels = make(map[string]string)
		}
		for _, k := range flagRedisUpdateRemoveLabels {
			delete(current.Labels, k)
		}
		for k, v := range flagRedisUpdateLabels {
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

	fmt.Fprintf(os.Stderr, "Updating Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Printf("Updated Redis instance [%s].\n", args[0])
	return nil
}

// --- redis instances failover (#194) ---

func runRedisInstancesFailover(cmd *cobra.Command, args []string) error {
	project, region, err := resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])
	req := &redis.FailoverInstanceRequest{}
	op, err := svc.Projects.Locations.Instances.Failover(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failing over redis instance: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Failing over Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Printf("Failover completed for Redis instance [%s].\n", args[0])
	return nil
}

// --- redis instances export (#195) ---

func runRedisInstancesExport(cmd *cobra.Command, args []string) error {
	project, region, err := resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])
	req := &redis.ExportInstanceRequest{
		OutputConfig: &redis.OutputConfig{
			GcsDestination: &redis.GcsDestination{
				Uri: flagRedisExportURI,
			},
		},
	}
	op, err := svc.Projects.Locations.Instances.Export(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting redis instance: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Exporting Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Printf("Exported Redis instance [%s] to %s.\n", args[0], flagRedisExportURI)
	return nil
}

// --- redis instances import (#195) ---

func runRedisInstancesImport(cmd *cobra.Command, args []string) error {
	project, region, err := resolveRedisRegion()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.RedisService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, region, args[0])
	req := &redis.ImportInstanceRequest{
		InputConfig: &redis.InputConfig{
			GcsSource: &redis.GcsSource{
				Uri: flagRedisImportURI,
			},
		},
	}
	op, err := svc.Projects.Locations.Instances.Import(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing to redis instance: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Importing to Redis instance [%s]...\n", args[0])
	if err := waitForRedisOp(ctx, svc, op.Name); err != nil {
		return err
	}
	fmt.Printf("Imported data to Redis instance [%s] from %s.\n", args[0], flagRedisImportURI)
	return nil
}
