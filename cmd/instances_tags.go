package cmd

import (
	"context"
	"fmt"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

var instancesAddTagsCmd = &cobra.Command{
	Use:   "add-tags INSTANCE_NAME",
	Short: "Add network tags to an instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesAddTags,
}

var instancesRemoveTagsCmd = &cobra.Command{
	Use:   "remove-tags INSTANCE_NAME",
	Short: "Remove network tags from an instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesRemoveTags,
}

var (
	flagInstanceAddTags    []string
	flagInstanceRemoveTags []string
	flagRemoveAllTags      bool
)

func init() {
	instancesAddTagsCmd.Flags().StringSliceVar(&flagInstanceAddTags, "tags", nil, "Tags to add")
	instancesAddTagsCmd.MarkFlagRequired("tags")

	instancesRemoveTagsCmd.Flags().StringSliceVar(&flagInstanceRemoveTags, "tags", nil, "Tags to remove")
	instancesRemoveTagsCmd.Flags().BoolVar(&flagRemoveAllTags, "all", false, "Remove all tags")

	instancesCmd.AddCommand(instancesAddTagsCmd)
	instancesCmd.AddCommand(instancesRemoveTagsCmd)
}

func runInstancesAddTags(cmd *cobra.Command, args []string) error {
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

	tags := inst.Tags
	if tags == nil {
		tags = &compute.Tags{}
	}

	existing := make(map[string]bool)
	for _, t := range tags.Items {
		existing[t] = true
	}
	for _, t := range flagInstanceAddTags {
		t = strings.TrimSpace(t)
		if !existing[t] {
			tags.Items = append(tags.Items, t)
		}
	}

	op, err := svc.Instances.SetTags(project, zone, instance, tags).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Updated tags for instance [%s].\n", instance)
	return nil
}

func runInstancesRemoveTags(cmd *cobra.Command, args []string) error {
	if !flagRemoveAllTags && len(flagInstanceRemoveTags) == 0 {
		return fmt.Errorf("one of --tags or --all must be specified")
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

	tags := inst.Tags
	if tags == nil {
		tags = &compute.Tags{}
	}

	if flagRemoveAllTags {
		tags.Items = nil
	} else {
		toRemove := make(map[string]bool)
		for _, t := range flagInstanceRemoveTags {
			toRemove[strings.TrimSpace(t)] = true
		}
		var filtered []string
		for _, t := range tags.Items {
			if !toRemove[t] {
				filtered = append(filtered, t)
			}
		}
		tags.Items = filtered
	}

	op, err := svc.Instances.SetTags(project, zone, instance, tags).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Updated tags for instance [%s].\n", instance)
	return nil
}
