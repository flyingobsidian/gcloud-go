package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
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

var flagRedisRegion string

func init() {
	redisInstancesDescribeCmd.Flags().StringVar(&flagRedisRegion, "region", "", "Region of the Redis instance")

	redisInstancesCmd.AddCommand(redisInstancesDescribeCmd)
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
