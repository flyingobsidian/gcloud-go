package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

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

var (
	flagRedisRegion     string
	flagRedisListFormat string
)

func init() {
	redisInstancesDescribeCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region of the Redis instance")
	redisInstancesListCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region of the Redis instances (uses '-' for all regions)")
	redisInstancesListCmd.Flags().StringVar(&flagRedisListFormat, "format", "", "Output format (json)")

	redisInstancesCmd.AddCommand(redisInstancesDescribeCmd)
	redisInstancesCmd.AddCommand(redisInstancesListCmd)
	redisCmd.AddCommand(redisInstancesCmd)
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
		pageToken = resp.NextPageToken
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
