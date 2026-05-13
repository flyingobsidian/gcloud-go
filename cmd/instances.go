package cmd

import (
	"context"
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-golang-cli/internal/compute"
	"github.com/flyingobsidian/gcloud-golang-cli/internal/config"
	"github.com/spf13/cobra"
)

var instancesStopCmd = &cobra.Command{
	Use:   "stop INSTANCE_NAME",
	Short: "Stop a Compute Engine instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesStop,
}

var instancesStartCmd = &cobra.Command{
	Use:   "start INSTANCE_NAME",
	Short: "Start a Compute Engine instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesStart,
}

var flagAsync bool

func init() {
	instancesStopCmd.Flags().BoolVar(&flagAsync, "async", false, "Return immediately without waiting for operation to complete")
	instancesStartCmd.Flags().BoolVar(&flagAsync, "async", false, "Return immediately without waiting for operation to complete")

	instancesCmd.AddCommand(instancesStopCmd)
	instancesCmd.AddCommand(instancesStartCmd)
}

func resolveProjectZone() (string, string, error) {
	props, err := config.Load()
	if err != nil {
		return "", "", err
	}
	project := config.Resolve(flagProject, "CLOUDSDK_CORE_PROJECT", props.Core.Project)
	zone := config.Resolve(flagZone, "CLOUDSDK_COMPUTE_ZONE", props.Compute.Zone)
	if project == "" {
		return "", "", fmt.Errorf("project is required; set via --project flag, CLOUDSDK_CORE_PROJECT env, or config")
	}
	if zone == "" {
		return "", "", fmt.Errorf("zone is required; set via --zone flag, CLOUDSDK_COMPUTE_ZONE env, or config")
	}
	return project, zone, nil
}

func runInstancesStop(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Stopping instance [%s] in zone [%s]...\n", instance, zone)
	op, err := svc.Instances.Stop(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping instance: %w", err)
	}

	if flagAsync {
		fmt.Printf("Stop operation started: %s\n", op.Name)
		return nil
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Stopped instance [%s].\n", instance)
	return nil
}

func runInstancesStart(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Starting instance [%s] in zone [%s]...\n", instance, zone)
	op, err := svc.Instances.Start(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting instance: %w", err)
	}

	if flagAsync {
		fmt.Printf("Start operation started: %s\n", op.Name)
		return nil
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Started instance [%s].\n", instance)
	return nil
}
