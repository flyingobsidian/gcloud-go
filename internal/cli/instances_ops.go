package cli

import (
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- instances reset ---

// --- instances set-machine-type ---

// --- instances get-serial-port-output ---

func (s *computeCmds) initInstancesOps() {
	s.instancesSetMachineTypeCmd.Flags().StringVar(&s.flagSetMachineType, "machine-type", "", "New machine type")
	s.instancesSetMachineTypeCmd.MarkFlagRequired("machine-type")

	s.instancesGetSerialPortOutputCmd.Flags().Int64Var(&s.flagSerialPort, "port", 1, "Serial port number (1-4)")
	s.instancesGetSerialPortOutputCmd.Flags().Int64Var(&s.flagSerialStart, "start", 0, "Byte offset to start reading from")

	s.instancesResetCmd.Flags().BoolVar(&s.flagResetAsync, "async", false, "Return immediately without waiting")
	s.instancesCmd.AddCommand(s.instancesResetCmd)
	s.instancesCmd.AddCommand(s.instancesSetMachineTypeCmd)
	s.instancesCmd.AddCommand(s.instancesGetSerialPortOutputCmd)
}

func (s *computeCmds) runInstancesReset(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(s.inv.stdout, "Resetting instance [%s] in zone [%s]...\n", instance, zone)
		op, err := svc.Instances.Reset(project, zone, instance).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("resetting instance %s: %w", instance, err)
		}

		if s.flagResetAsync {
			fmt.Fprintf(s.inv.stdout, "Reset operation started for [%s]: %s\n", instance, op.Name)
			continue
		}

		if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
			return err
		}
		fmt.Fprintf(s.inv.stdout, "Reset instance [%s].\n", instance)
	}
	return nil
}

func (s *computeCmds) runInstancesSetMachineType(cmd *cobra.Command, args []string) error {
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

	machineTypeURL := fmt.Sprintf("zones/%s/machineTypes/%s", zone, s.flagSetMachineType)
	req := &compute.InstancesSetMachineTypeRequest{
		MachineType: machineTypeURL,
	}

	fmt.Fprintf(s.inv.stdout, "Setting machine type of instance [%s] to [%s]...\n", instance, s.flagSetMachineType)
	op, err := svc.Instances.SetMachineType(project, zone, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting machine type: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Set machine type of instance [%s] to [%s].\n", instance, s.flagSetMachineType)
	return nil
}

func (s *computeCmds) runInstancesGetSerialPortOutput(cmd *cobra.Command, args []string) error {
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

	call := svc.Instances.GetSerialPortOutput(project, zone, instance).Context(ctx).Port(s.flagSerialPort)
	if s.flagSerialStart > 0 {
		call = call.Start(s.flagSerialStart)
	}

	output, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting serial port output: %w", err)
	}

	fmt.Fprint(s.inv.stdout, output.Contents)
	return nil
}
