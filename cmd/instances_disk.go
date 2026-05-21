package cmd

import (
	"context"
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- attach-disk ---

var instancesAttachDiskCmd = &cobra.Command{
	Use:   "attach-disk INSTANCE_NAME",
	Short: "Attach a disk to an instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesAttachDisk,
}

var (
	flagAttachDisk       string
	flagAttachDiskMode   string
	flagAttachDeviceName string
)

// --- detach-disk ---

var instancesDetachDiskCmd = &cobra.Command{
	Use:   "detach-disk INSTANCE_NAME",
	Short: "Detach a disk from an instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesDetachDisk,
}

var flagDetachDisk string

// --- set-disk-auto-delete ---

var instancesSetDiskAutoDeleteCmd = &cobra.Command{
	Use:   "set-disk-auto-delete INSTANCE_NAME",
	Short: "Set auto-delete policy for a disk on an instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstancesSetDiskAutoDelete,
}

var (
	flagAutoDeleteDisk   string
	flagAutoDelete       bool
	flagNoAutoDelete     bool
)

func init() {
	instancesAttachDiskCmd.Flags().StringVar(&flagAttachDisk, "disk", "", "Name of the disk to attach")
	instancesAttachDiskCmd.MarkFlagRequired("disk")
	instancesAttachDiskCmd.Flags().StringVar(&flagAttachDiskMode, "mode", "rw", "Attach mode: rw or ro")
	instancesAttachDiskCmd.Flags().StringVar(&flagAttachDeviceName, "device-name", "", "Device name for the disk")

	instancesDetachDiskCmd.Flags().StringVar(&flagDetachDisk, "disk", "", "Name of the disk to detach")
	instancesDetachDiskCmd.MarkFlagRequired("disk")

	instancesSetDiskAutoDeleteCmd.Flags().StringVar(&flagAutoDeleteDisk, "disk", "", "Name of the disk")
	instancesSetDiskAutoDeleteCmd.MarkFlagRequired("disk")
	instancesSetDiskAutoDeleteCmd.Flags().BoolVar(&flagAutoDelete, "auto-delete", false, "Enable auto-delete")
	instancesSetDiskAutoDeleteCmd.Flags().BoolVar(&flagNoAutoDelete, "no-auto-delete", false, "Disable auto-delete")

	instancesCmd.AddCommand(instancesAttachDiskCmd)
	instancesCmd.AddCommand(instancesDetachDiskCmd)
	instancesCmd.AddCommand(instancesSetDiskAutoDeleteCmd)
}

func runInstancesAttachDisk(cmd *cobra.Command, args []string) error {
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

	diskURL := fmt.Sprintf("projects/%s/zones/%s/disks/%s", project, zone, flagAttachDisk)
	mode := "READ_WRITE"
	if flagAttachDiskMode == "ro" {
		mode = "READ_ONLY"
	}

	disk := &compute.AttachedDisk{
		Source:     diskURL,
		Mode:       mode,
		DeviceName: flagAttachDeviceName,
	}

	fmt.Printf("Attaching disk [%s] to instance [%s]...\n", flagAttachDisk, instance)
	op, err := svc.Instances.AttachDisk(project, zone, instance, disk).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("attaching disk: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Attached disk [%s] to instance [%s].\n", flagAttachDisk, instance)
	return nil
}

func runInstancesDetachDisk(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Detaching disk [%s] from instance [%s]...\n", flagDetachDisk, instance)
	op, err := svc.Instances.DetachDisk(project, zone, instance, flagDetachDisk).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("detaching disk: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Detached disk [%s] from instance [%s].\n", flagDetachDisk, instance)
	return nil
}

func runInstancesSetDiskAutoDelete(cmd *cobra.Command, args []string) error {
	if !flagAutoDelete && !flagNoAutoDelete {
		return fmt.Errorf("one of --auto-delete or --no-auto-delete must be specified")
	}

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

	autoDelete := flagAutoDelete
	op, err := svc.Instances.SetDiskAutoDelete(project, zone, instance, autoDelete, flagAutoDeleteDisk).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting disk auto-delete: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Updated auto-delete for disk [%s] on instance [%s].\n", flagAutoDeleteDisk, instance)
	return nil
}
