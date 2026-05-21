package cmd

import (
	"context"
	"fmt"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

var instancesAddMetadataCmd = &cobra.Command{
	Use:   "add-metadata INSTANCE_NAME",
	Short: "Add or update metadata for an instance",
	Long: `Add or update instance metadata. Existing keys are overwritten.

Examples:
  gcloud compute instances add-metadata my-instance --metadata=key1=value1,key2=value2 --zone=us-central1-a`,
	Args: cobra.ExactArgs(1),
	RunE: runInstancesAddMetadata,
}

var instancesRemoveMetadataCmd = &cobra.Command{
	Use:   "remove-metadata INSTANCE_NAME",
	Short: "Remove metadata from an instance",
	Long: `Remove instance metadata by key.

Examples:
  gcloud compute instances remove-metadata my-instance --keys=key1,key2 --zone=us-central1-a
  gcloud compute instances remove-metadata my-instance --all --zone=us-central1-a`,
	Args: cobra.ExactArgs(1),
	RunE: runInstancesRemoveMetadata,
}

var (
	flagAddMetadataKV     map[string]string
	flagRemoveMetadataKeys string
	flagRemoveMetadataAll  bool
)

func init() {
	instancesAddMetadataCmd.Flags().StringToStringVar(&flagAddMetadataKV, "metadata", nil, "Metadata key=value pairs to add")
	instancesAddMetadataCmd.MarkFlagRequired("metadata")

	instancesRemoveMetadataCmd.Flags().StringVar(&flagRemoveMetadataKeys, "keys", "", "Comma-separated list of metadata keys to remove")
	instancesRemoveMetadataCmd.Flags().BoolVar(&flagRemoveMetadataAll, "all", false, "Remove all metadata")

	instancesCmd.AddCommand(instancesAddMetadataCmd)
	instancesCmd.AddCommand(instancesRemoveMetadataCmd)
}

func runInstancesAddMetadata(cmd *cobra.Command, args []string) error {
	instance := args[0]
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance: %w", err)
	}

	metadata := inst.Metadata
	if metadata == nil {
		metadata = &compute.Metadata{}
	}

	// Update or add keys.
	for k, v := range flagAddMetadataKV {
		val := v
		found := false
		for _, item := range metadata.Items {
			if item.Key == k {
				item.Value = &val
				found = true
				break
			}
		}
		if !found {
			metadata.Items = append(metadata.Items, &compute.MetadataItems{Key: k, Value: &val})
		}
	}

	op, err := svc.Instances.SetMetadata(project, zone, instance, metadata).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting metadata: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Updated metadata for instance [%s].\n", instance)
	return nil
}

func runInstancesRemoveMetadata(cmd *cobra.Command, args []string) error {
	if !flagRemoveMetadataAll && flagRemoveMetadataKeys == "" {
		return fmt.Errorf("one of --keys or --all must be specified")
	}

	instance := args[0]
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance: %w", err)
	}

	metadata := inst.Metadata
	if metadata == nil {
		metadata = &compute.Metadata{}
	}

	if flagRemoveMetadataAll {
		metadata.Items = nil
	} else {
		keysToRemove := make(map[string]bool)
		for _, k := range strings.Split(flagRemoveMetadataKeys, ",") {
			keysToRemove[strings.TrimSpace(k)] = true
		}
		var filtered []*compute.MetadataItems
		for _, item := range metadata.Items {
			if !keysToRemove[item.Key] {
				filtered = append(filtered, item)
			}
		}
		metadata.Items = filtered
	}

	op, err := svc.Instances.SetMetadata(project, zone, instance, metadata).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting metadata: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Updated metadata for instance [%s].\n", instance)
	return nil
}
