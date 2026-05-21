package cmd

import (
	"context"
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- instances suspend ---

var instancesSuspendCmd = &cobra.Command{
	Use:   "suspend INSTANCE_NAME",
	Short: "Suspend a Compute Engine instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesSuspend,
}

// --- instances resume ---

var instancesResumeCmd = &cobra.Command{
	Use:   "resume INSTANCE_NAME",
	Short: "Resume a suspended Compute Engine instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesResume,
}

// --- instances set-scheduling ---

var instancesSetSchedulingCmd = &cobra.Command{
	Use:   "set-scheduling INSTANCE_NAME",
	Short: "Set scheduling options for an instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesSetScheduling,
}

var (
	flagRestartOnFailure   bool
	flagNoRestartOnFailure bool
	flagMaintenancePolicy  string
)

func init() {
	instancesSuspendCmd.Flags().BoolVar(&flagAsync, "async", false, "Return immediately without waiting")
	instancesResumeCmd.Flags().BoolVar(&flagAsync, "async", false, "Return immediately without waiting")

	instancesSetSchedulingCmd.Flags().BoolVar(&flagRestartOnFailure, "restart-on-failure", false, "Restart on failure")
	instancesSetSchedulingCmd.Flags().BoolVar(&flagNoRestartOnFailure, "no-restart-on-failure", false, "Do not restart on failure")
	instancesSetSchedulingCmd.Flags().StringVar(&flagMaintenancePolicy, "maintenance-policy", "", "Maintenance policy: MIGRATE or TERMINATE")

	instancesCmd.AddCommand(instancesSuspendCmd)
	instancesCmd.AddCommand(instancesResumeCmd)
	instancesCmd.AddCommand(instancesSetSchedulingCmd)
}

func runInstancesSuspend(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Suspending instance [%s] in zone [%s]...\n", instance, zone)
	op, err := svc.Instances.Suspend(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("suspending instance: %w", err)
	}

	if flagAsync {
		fmt.Printf("Suspend operation started: %s\n", op.Name)
		return nil
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Suspended instance [%s].\n", instance)
	return nil
}

func runInstancesResume(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Resuming instance [%s] in zone [%s]...\n", instance, zone)
	op, err := svc.Instances.Resume(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming instance: %w", err)
	}

	if flagAsync {
		fmt.Printf("Resume operation started: %s\n", op.Name)
		return nil
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Resumed instance [%s].\n", instance)
	return nil
}

func runInstancesSetScheduling(cmd *cobra.Command, args []string) error {
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

	scheduling := &compute.Scheduling{}
	if flagRestartOnFailure {
		t := true
		scheduling.AutomaticRestart = &t
	}
	if flagNoRestartOnFailure {
		f := false
		scheduling.AutomaticRestart = &f
	}
	if flagMaintenancePolicy != "" {
		scheduling.OnHostMaintenance = flagMaintenancePolicy
	}

	op, err := svc.Instances.SetScheduling(project, zone, instance, scheduling).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting scheduling: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Updated scheduling for instance [%s].\n", instance)
	return nil
}
