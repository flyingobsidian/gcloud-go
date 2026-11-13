package cli

import (
	"fmt"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

func (s *computeCmds) initInstancesTags() {
	s.instancesAddTagsCmd.Flags().StringSliceVar(&s.flagInstanceAddTags, "tags", nil, "Tags to add")
	s.instancesAddTagsCmd.MarkFlagRequired("tags")

	s.instancesRemoveTagsCmd.Flags().StringSliceVar(&s.flagInstanceRemoveTags, "tags", nil, "Tags to remove")
	s.instancesRemoveTagsCmd.Flags().BoolVar(&s.flagRemoveAllTags, "all", false, "Remove all tags")

	s.instancesCmd.AddCommand(s.instancesAddTagsCmd)
	s.instancesCmd.AddCommand(s.instancesRemoveTagsCmd)
}

func (s *computeCmds) runInstancesAddTags(cmd *cobra.Command, args []string) error {
	instance := args[0]
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
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
	for _, t := range s.flagInstanceAddTags {
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
	fmt.Fprintf(s.inv.stdout, "Updated tags for instance [%s].\n", instance)
	return nil
}

func (s *computeCmds) runInstancesRemoveTags(cmd *cobra.Command, args []string) error {
	if !s.flagRemoveAllTags && len(s.flagInstanceRemoveTags) == 0 {
		return fmt.Errorf("one of --tags or --all must be specified")
	}

	instance := args[0]
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
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

	if s.flagRemoveAllTags {
		tags.Items = nil
	} else {
		toRemove := make(map[string]bool)
		for _, t := range s.flagInstanceRemoveTags {
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
	fmt.Fprintf(s.inv.stdout, "Updated tags for instance [%s].\n", instance)
	return nil
}
