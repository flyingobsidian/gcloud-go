package cmd

import (
	"context"
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- instances reset ---

var instancesResetCmd = &cobra.Command{
	Use:   "reset INSTANCE_NAME",
	Short: "Reset a Compute Engine instance",
	Long:  "Perform a hard reboot of an instance (distinct from stop+start).",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesReset,
}

// --- instances set-machine-type ---

var instancesSetMachineTypeCmd = &cobra.Command{
	Use:   "set-machine-type INSTANCE_NAME",
	Short: "Change the machine type of a stopped instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesSetMachineType,
}

var flagSetMachineType string

// --- instances get-serial-port-output ---

var instancesGetSerialPortOutputCmd = &cobra.Command{
	Use:   "get-serial-port-output INSTANCE_NAME",
	Short: "Read serial console output from an instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesGetSerialPortOutput,
}

var (
	flagSerialPort  int64
	flagSerialStart int64
)

func init() {
	instancesSetMachineTypeCmd.Flags().StringVar(&flagSetMachineType, "machine-type", "", "New machine type")
	instancesSetMachineTypeCmd.MarkFlagRequired("machine-type")

	instancesGetSerialPortOutputCmd.Flags().Int64Var(&flagSerialPort, "port", 1, "Serial port number (1-4)")
	instancesGetSerialPortOutputCmd.Flags().Int64Var(&flagSerialStart, "start", 0, "Byte offset to start reading from")

	instancesCmd.AddCommand(instancesResetCmd)
	instancesCmd.AddCommand(instancesSetMachineTypeCmd)
	instancesCmd.AddCommand(instancesGetSerialPortOutputCmd)
}

func runInstancesReset(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Resetting instance [%s] in zone [%s]...\n", instance, zone)
	op, err := svc.Instances.Reset(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting instance: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Reset instance [%s].\n", instance)
	return nil
}

func runInstancesSetMachineType(cmd *cobra.Command, args []string) error {
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

	machineTypeURL := fmt.Sprintf("zones/%s/machineTypes/%s", zone, flagSetMachineType)
	req := &compute.InstancesSetMachineTypeRequest{
		MachineType: machineTypeURL,
	}

	fmt.Printf("Setting machine type of instance [%s] to [%s]...\n", instance, flagSetMachineType)
	op, err := svc.Instances.SetMachineType(project, zone, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting machine type: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Set machine type of instance [%s] to [%s].\n", instance, flagSetMachineType)
	return nil
}

func runInstancesGetSerialPortOutput(cmd *cobra.Command, args []string) error {
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

	call := svc.Instances.GetSerialPortOutput(project, zone, instance).Context(ctx).Port(flagSerialPort)
	if flagSerialStart > 0 {
		call = call.Start(flagSerialStart)
	}

	output, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting serial port output: %w", err)
	}

	fmt.Print(output.Contents)
	return nil
}
