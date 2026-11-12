package cmd

import (
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- instances suspend ---

// --- instances resume ---

// --- instances set-scheduling ---

func (s *computeCmds) initInstancesScheduling() {
	s.instancesSuspendCmd.Flags().BoolVar(&s.flagAsync, "async", false, "Return immediately without waiting")
	s.instancesResumeCmd.Flags().BoolVar(&s.flagAsync, "async", false, "Return immediately without waiting")

	s.instancesSetSchedulingCmd.Flags().BoolVar(&s.flagRestartOnFailure, "restart-on-failure", false, "Restart on failure")
	s.instancesSetSchedulingCmd.Flags().BoolVar(&s.flagNoRestartOnFailure, "no-restart-on-failure", false, "Do not restart on failure")
	s.instancesSetSchedulingCmd.Flags().StringVar(&s.flagMaintenancePolicy, "maintenance-policy", "", "Maintenance policy: MIGRATE or TERMINATE")

	s.instancesCmd.AddCommand(s.instancesSuspendCmd)
	s.instancesCmd.AddCommand(s.instancesResumeCmd)
	s.instancesCmd.AddCommand(s.instancesSetSchedulingCmd)
}

func (s *computeCmds) runInstancesSuspend(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(s.inv.stdout, "Suspending instance [%s] in zone [%s]...\n", instance, zone)
		op, err := svc.Instances.Suspend(project, zone, instance).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("suspending instance %s: %w", instance, err)
		}

		if s.flagAsync {
			fmt.Fprintf(s.inv.stdout, "Suspend operation started: %s\n", op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Fprintf(s.inv.stdout, "Suspended instance [%s].\n", instance)
	}
	return nil
}

func (s *computeCmds) runInstancesResume(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(s.inv.stdout, "Resuming instance [%s] in zone [%s]...\n", instance, zone)
		op, err := svc.Instances.Resume(project, zone, instance).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("resuming instance %s: %w", instance, err)
		}

		if s.flagAsync {
			fmt.Fprintf(s.inv.stdout, "Resume operation started: %s\n", op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Fprintf(s.inv.stdout, "Resumed instance [%s].\n", instance)
	}
	return nil
}

func (s *computeCmds) runInstancesSetScheduling(cmd *cobra.Command, args []string) error {
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

	scheduling := &compute.Scheduling{}
	if s.flagRestartOnFailure {
		t := true
		scheduling.AutomaticRestart = &t
	}
	if s.flagNoRestartOnFailure {
		f := false
		scheduling.AutomaticRestart = &f
	}
	if s.flagMaintenancePolicy != "" {
		scheduling.OnHostMaintenance = s.flagMaintenancePolicy
	}

	op, err := svc.Instances.SetScheduling(project, zone, instance, scheduling).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting scheduling: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Updated scheduling for instance [%s].\n", instance)
	return nil
}
