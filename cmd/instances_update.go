package cmd

import (
	"context"
	"fmt"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

var instancesUpdateCmd = &cobra.Command{
	Use:   "update INSTANCE_NAME",
	Short: "Update a Compute Engine instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesUpdate,
}

var (
	flagUpdateLabels      map[string]string
	flagUpdateRemoveLabels string
	flagDeletionProtection bool
	flagNoDeletionProtection bool
)

func init() {
	instancesUpdateCmd.Flags().StringToStringVar(&flagUpdateLabels, "update-labels", nil, "Labels to add or update")
	instancesUpdateCmd.Flags().StringVar(&flagUpdateRemoveLabels, "remove-labels", "", "Comma-separated label keys to remove")
	instancesUpdateCmd.Flags().BoolVar(&flagDeletionProtection, "deletion-protection", false, "Enable deletion protection")
	instancesUpdateCmd.Flags().BoolVar(&flagNoDeletionProtection, "no-deletion-protection", false, "Disable deletion protection")

	instancesCmd.AddCommand(instancesUpdateCmd)
}

func runInstancesUpdate(cmd *cobra.Command, args []string) error {
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

	// Handle labels.
	if len(flagUpdateLabels) > 0 || flagUpdateRemoveLabels != "" {
		labels := inst.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		for k, v := range flagUpdateLabels {
			labels[k] = v
		}
		if flagUpdateRemoveLabels != "" {
			for _, k := range strings.Split(flagUpdateRemoveLabels, ",") {
				delete(labels, strings.TrimSpace(k))
			}
		}
		req := &compute.InstancesSetLabelsRequest{
			LabelFingerprint: inst.LabelFingerprint,
			Labels:           labels,
		}
		op, err := svc.Instances.SetLabels(project, zone, instance, req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating labels: %w", err)
		}
		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
	}

	// Handle deletion protection.
	if flagDeletionProtection || flagNoDeletionProtection {
		inst.DeletionProtection = flagDeletionProtection
		op, err := svc.Instances.SetDeletionProtection(project, zone, instance).DeletionProtection(flagDeletionProtection).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("setting deletion protection: %w", err)
		}
		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
	}

	fmt.Printf("Updated instance [%s].\n", instance)
	return nil
}
