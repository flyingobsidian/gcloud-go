package cli

import (
	"fmt"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

func (s *computeCmds) initInstancesLabels() {
	s.instancesAddLabelsCmd.Flags().StringToStringVar(&s.flagAddLabels, "labels", nil, "Label key=value pairs to add")
	s.instancesAddLabelsCmd.MarkFlagRequired("labels")

	s.instancesRemoveLabelsCmd.Flags().StringVar(&s.flagRemoveLabelsKeys, "labels", "", "Comma-separated list of label keys to remove")
	s.instancesRemoveLabelsCmd.Flags().BoolVar(&s.flagRemoveLabelsAll, "all", false, "Remove all labels")

	s.instancesCmd.AddCommand(s.instancesAddLabelsCmd)
	s.instancesCmd.AddCommand(s.instancesRemoveLabelsCmd)
}

func (s *computeCmds) runInstancesAddLabels(cmd *cobra.Command, args []string) error {
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

	labels := inst.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	for k, v := range s.flagAddLabels {
		labels[k] = v
	}

	req := &compute.InstancesSetLabelsRequest{
		LabelFingerprint: inst.LabelFingerprint,
		Labels:           labels,
	}

	op, err := svc.Instances.SetLabels(project, zone, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting labels: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Updated labels for instance [%s].\n", instance)
	return nil
}

func (s *computeCmds) runInstancesRemoveLabels(cmd *cobra.Command, args []string) error {
	if !s.flagRemoveLabelsAll && s.flagRemoveLabelsKeys == "" {
		return fmt.Errorf("one of --labels or --all must be specified")
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

	labels := inst.Labels
	if s.flagRemoveLabelsAll {
		labels = map[string]string{}
	} else {
		if labels == nil {
			labels = map[string]string{}
		}
		for _, k := range strings.Split(s.flagRemoveLabelsKeys, ",") {
			delete(labels, strings.TrimSpace(k))
		}
	}

	req := &compute.InstancesSetLabelsRequest{
		LabelFingerprint: inst.LabelFingerprint,
		Labels:           labels,
	}

	op, err := svc.Instances.SetLabels(project, zone, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting labels: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Updated labels for instance [%s].\n", instance)
	return nil
}
