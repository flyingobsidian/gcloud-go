package cli

import (
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
)

func (s *computeCmds) initInstances() {
	s.instancesStopCmd.Flags().BoolVar(&s.flagAsync, "async", false, "Return immediately without waiting for operation to complete")
	s.instancesStopCmd.Flags().BoolVar(&s.flagStopDiscardLocalSSD, "discard-local-ssd", false, "Discard data on local SSDs when stopping")
	s.instancesStartCmd.Flags().BoolVar(&s.flagAsync, "async", false, "Return immediately without waiting for operation to complete")

	s.instancesCmd.AddCommand(s.instancesStopCmd)
	s.instancesCmd.AddCommand(s.instancesStartCmd)
}

func (s *computeCmds) runInstancesStop(cmd *cobra.Command, args []string) error {
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	for _, instance := range args {
		fmt.Fprintf(s.inv.stdout, "Stopping instance [%s] in zone [%s]...\n", instance, zone)
		call := svc.Instances.Stop(project, zone, instance).Context(ctx)
		if s.flagStopDiscardLocalSSD {
			call = call.DiscardLocalSsd(true)
		}
		op, err := call.Do()
		if err != nil {
			return fmt.Errorf("stopping instance %s: %w", instance, err)
		}

		if s.flagAsync {
			fmt.Fprintf(s.inv.stdout, "Stop operation started: %s\n", op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Fprintf(s.inv.stdout, "Stopped instance [%s].\n", instance)
	}
	return nil
}

func (s *computeCmds) runInstancesStart(cmd *cobra.Command, args []string) error {
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	for _, instance := range args {
		fmt.Fprintf(s.inv.stdout, "Starting instance [%s] in zone [%s]...\n", instance, zone)
		op, err := svc.Instances.Start(project, zone, instance).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("starting instance %s: %w", instance, err)
		}

		if s.flagAsync {
			fmt.Fprintf(s.inv.stdout, "Start operation started: %s\n", op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}

		// Print resulting IPs after start.
		inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
		if err == nil && len(inst.NetworkInterfaces) > 0 {
			ni := inst.NetworkInterfaces[0]
			fmt.Fprintf(s.inv.stdout, "  Internal IP: %s\n", ni.NetworkIP)
			if len(ni.AccessConfigs) > 0 && ni.AccessConfigs[0].NatIP != "" {
				fmt.Fprintf(s.inv.stdout, "  External IP: %s\n", ni.AccessConfigs[0].NatIP)
			}
		}

		fmt.Fprintf(s.inv.stdout, "Started instance [%s].\n", instance)
	}
	return nil
}
