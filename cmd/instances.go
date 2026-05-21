package cmd

import (
	"context"
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
)

var instancesStopCmd = &cobra.Command{
	Use:   "stop INSTANCE_NAME [INSTANCE_NAME ...]",
	Short: "Stop a Compute Engine instance",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runInstancesStop,
}

var instancesStartCmd = &cobra.Command{
	Use:   "start INSTANCE_NAME [INSTANCE_NAME ...]",
	Short: "Start a Compute Engine instance",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runInstancesStart,
}

var (
	flagAsync               bool
	flagStopDiscardLocalSSD bool
)

func init() {
	instancesStopCmd.Flags().BoolVar(&flagAsync, "async", false, "Return immediately without waiting for operation to complete")
	instancesStopCmd.Flags().BoolVar(&flagStopDiscardLocalSSD, "discard-local-ssd", false, "Discard data on local SSDs when stopping")
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
		if !IsInteractive() {
			return "", "", fmt.Errorf("zone is required; set via --zone flag, CLOUDSDK_COMPUTE_ZONE env, or config")
		}
		fmt.Print("Enter zone (e.g. us-central1-a): ")
		fmt.Scanln(&zone)
		if zone == "" {
			return "", "", fmt.Errorf("zone is required")
		}
	}
	return project, zone, nil
}

func runInstancesStop(cmd *cobra.Command, args []string) error {
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	for _, instance := range args {
		fmt.Printf("Stopping instance [%s] in zone [%s]...\n", instance, zone)
		call := svc.Instances.Stop(project, zone, instance).Context(ctx)
		if flagStopDiscardLocalSSD {
			call = call.DiscardLocalSsd(true)
		}
		op, err := call.Do()
		if err != nil {
			return fmt.Errorf("stopping instance %s: %w", instance, err)
		}

		if flagAsync {
			fmt.Printf("Stop operation started: %s\n", op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Printf("Stopped instance [%s].\n", instance)
	}
	return nil
}

func runInstancesStart(cmd *cobra.Command, args []string) error {
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	for _, instance := range args {
		fmt.Printf("Starting instance [%s] in zone [%s]...\n", instance, zone)
		op, err := svc.Instances.Start(project, zone, instance).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("starting instance %s: %w", instance, err)
		}

		if flagAsync {
			fmt.Printf("Start operation started: %s\n", op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}

		// Print resulting IPs after start.
		inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
		if err == nil && len(inst.NetworkInterfaces) > 0 {
			ni := inst.NetworkInterfaces[0]
			fmt.Printf("  Internal IP: %s\n", ni.NetworkIP)
			if len(ni.AccessConfigs) > 0 && ni.AccessConfigs[0].NatIP != "" {
				fmt.Printf("  External IP: %s\n", ni.AccessConfigs[0].NatIP)
			}
		}

		fmt.Printf("Started instance [%s].\n", instance)
	}
	return nil
}
