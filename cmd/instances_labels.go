package cmd

import (
	"context"
	"fmt"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

var instancesAddLabelsCmd = &cobra.Command{
	Use:   "add-labels INSTANCE_NAME",
	Short: "Add labels to an instance",
	Long: `Add or update labels on an instance.

Examples:
  gcloud compute instances add-labels my-instance --labels=env=prod,team=backend --zone=us-central1-a`,
	Args: cobra.ExactArgs(1),
	RunE: runInstancesAddLabels,
}

var instancesRemoveLabelsCmd = &cobra.Command{
	Use:   "remove-labels INSTANCE_NAME",
	Short: "Remove labels from an instance",
	Long: `Remove labels from an instance by key.

Examples:
  gcloud compute instances remove-labels my-instance --labels=env,team --zone=us-central1-a
  gcloud compute instances remove-labels my-instance --all --zone=us-central1-a`,
	Args: cobra.ExactArgs(1),
	RunE: runInstancesRemoveLabels,
}

var (
	flagAddLabels       map[string]string
	flagRemoveLabelsKeys string
	flagRemoveLabelsAll  bool
)

func init() {
	instancesAddLabelsCmd.Flags().StringToStringVar(&flagAddLabels, "labels", nil, "Label key=value pairs to add")
	instancesAddLabelsCmd.MarkFlagRequired("labels")

	instancesRemoveLabelsCmd.Flags().StringVar(&flagRemoveLabelsKeys, "labels", "", "Comma-separated list of label keys to remove")
	instancesRemoveLabelsCmd.Flags().BoolVar(&flagRemoveLabelsAll, "all", false, "Remove all labels")

	instancesCmd.AddCommand(instancesAddLabelsCmd)
	instancesCmd.AddCommand(instancesRemoveLabelsCmd)
}

func runInstancesAddLabels(cmd *cobra.Command, args []string) error {
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

	labels := inst.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	for k, v := range flagAddLabels {
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
	fmt.Printf("Updated labels for instance [%s].\n", instance)
	return nil
}

func runInstancesRemoveLabels(cmd *cobra.Command, args []string) error {
	if !flagRemoveLabelsAll && flagRemoveLabelsKeys == "" {
		return fmt.Errorf("one of --labels or --all must be specified")
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

	labels := inst.Labels
	if flagRemoveLabelsAll {
		labels = map[string]string{}
	} else {
		if labels == nil {
			labels = map[string]string{}
		}
		for _, k := range strings.Split(flagRemoveLabelsKeys, ",") {
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
	fmt.Printf("Updated labels for instance [%s].\n", instance)
	return nil
}
