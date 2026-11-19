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
	s.instancesGetSerialPortOutputCmd.Flags().Int64Var(&s.flagSerialStart, "start", 0, "Byte index (zero-based) of the first byte to return; negative values count back from the end")

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

	startSet := cmd.Flags().Changed("start")
	call := svc.Instances.GetSerialPortOutput(project, zone, instance).Context(ctx).Port(s.flagSerialPort)
	if startSet {
		call = call.Start(s.flagSerialStart)
	}

	output, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting serial port output: %w", err)
	}

	if s.inv.flags.format != "" {
		if err := emitFormattedTo(s.inv.stdout, output, s.inv.flags.format); err != nil {
			return err
		}
	} else {
		// gcloud-python's default format is value[no-quote](contents).
		fmt.Fprintln(s.inv.stdout, output.Contents)
	}
	if startSet && serialOutputLost(s.flagSerialStart, output) {
		fmt.Fprintf(s.inv.stderr, "WARNING: Some serial port output was lost due to a limited buffer. The oldest byte of output returned was at offset %d.\n", output.Start)
	}
	fmt.Fprintf(s.inv.stderr, "\nSpecify --start=%d in the next get-serial-port-output invocation to get only the new output starting from here.\n", output.Next)
	return nil
}

// serialOutputLost reports whether output starts after the requested start,
// as gcloud-python 589.0.0's Epilog does: a negative start is relative to
// output.Next, and output is lost if it begins after max(0, next+start).
func serialOutputLost(start int64, output *compute.SerialPortOutput) bool {
	if start < 0 {
		return output.Start > max(0, output.Next+start)
	}
	return output.Start != start
}
