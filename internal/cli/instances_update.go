package cli

import (
	"fmt"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

func (s *computeCmds) initInstancesUpdate() {
	s.instancesUpdateCmd.Flags().StringToStringVar(&s.flagUpdateLabels, "update-labels", nil, "Labels to add or update")
	s.instancesUpdateCmd.Flags().StringVar(&s.flagUpdateRemoveLabels, "remove-labels", "", "Comma-separated label keys to remove")
	s.instancesUpdateCmd.Flags().BoolVar(&s.flagDeletionProtection, "deletion-protection", false, "Enable deletion protection")
	s.instancesUpdateCmd.Flags().BoolVar(&s.flagNoDeletionProtection, "no-deletion-protection", false, "Disable deletion protection")

	s.instancesCmd.AddCommand(s.instancesUpdateCmd)
}

func (s *computeCmds) runInstancesUpdate(cmd *cobra.Command, args []string) error {
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

	// Handle labels.
	if len(s.flagUpdateLabels) > 0 || s.flagUpdateRemoveLabels != "" {
		labels := inst.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		for k, v := range s.flagUpdateLabels {
			labels[k] = v
		}
		if s.flagUpdateRemoveLabels != "" {
			for _, k := range strings.Split(s.flagUpdateRemoveLabels, ",") {
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
	if s.flagDeletionProtection || s.flagNoDeletionProtection {
		inst.DeletionProtection = s.flagDeletionProtection
		op, err := svc.Instances.SetDeletionProtection(project, zone, instance).DeletionProtection(s.flagDeletionProtection).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("setting deletion protection: %w", err)
		}
		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
	}

	fmt.Fprintf(s.inv.stdout, "Updated instance [%s].\n", instance)
	return nil
}
