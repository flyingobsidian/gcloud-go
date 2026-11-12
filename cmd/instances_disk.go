package cmd

import (
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- attach-disk ---

// --- detach-disk ---

// --- set-disk-auto-delete ---

func (s *computeCmds) initInstancesDisk() {
	s.instancesAttachDiskCmd.Flags().StringVar(&s.flagAttachDisk, "disk", "", "Name of the disk to attach")
	s.instancesAttachDiskCmd.MarkFlagRequired("disk")
	s.instancesAttachDiskCmd.Flags().StringVar(&s.flagAttachDiskMode, "mode", "rw", "Attach mode: rw or ro")
	s.instancesAttachDiskCmd.Flags().StringVar(&s.flagAttachDeviceName, "device-name", "", "Device name for the disk")

	s.instancesDetachDiskCmd.Flags().StringVar(&s.flagDetachDisk, "disk", "", "Name of the disk to detach")
	s.instancesDetachDiskCmd.MarkFlagRequired("disk")

	s.instancesSetDiskAutoDeleteCmd.Flags().StringVar(&s.flagAutoDeleteDisk, "disk", "", "Name of the disk")
	s.instancesSetDiskAutoDeleteCmd.MarkFlagRequired("disk")
	s.instancesSetDiskAutoDeleteCmd.Flags().BoolVar(&s.flagAutoDelete, "auto-delete", false, "Enable auto-delete")
	s.instancesSetDiskAutoDeleteCmd.Flags().BoolVar(&s.flagNoAutoDelete, "no-auto-delete", false, "Disable auto-delete")

	s.instancesCmd.AddCommand(s.instancesAttachDiskCmd)
	s.instancesCmd.AddCommand(s.instancesDetachDiskCmd)
	s.instancesCmd.AddCommand(s.instancesSetDiskAutoDeleteCmd)
}

func (s *computeCmds) runInstancesAttachDisk(cmd *cobra.Command, args []string) error {
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

	diskURL := fmt.Sprintf("projects/%s/zones/%s/disks/%s", project, zone, s.flagAttachDisk)
	mode := "READ_WRITE"
	if s.flagAttachDiskMode == "ro" {
		mode = "READ_ONLY"
	}

	disk := &compute.AttachedDisk{
		Source:     diskURL,
		Mode:       mode,
		DeviceName: s.flagAttachDeviceName,
	}

	fmt.Fprintf(s.inv.stdout, "Attaching disk [%s] to instance [%s]...\n", s.flagAttachDisk, instance)
	op, err := svc.Instances.AttachDisk(project, zone, instance, disk).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("attaching disk: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Attached disk [%s] to instance [%s].\n", s.flagAttachDisk, instance)
	return nil
}

func (s *computeCmds) runInstancesDetachDisk(cmd *cobra.Command, args []string) error {
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

	fmt.Fprintf(s.inv.stdout, "Detaching disk [%s] from instance [%s]...\n", s.flagDetachDisk, instance)
	op, err := svc.Instances.DetachDisk(project, zone, instance, s.flagDetachDisk).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("detaching disk: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Detached disk [%s] from instance [%s].\n", s.flagDetachDisk, instance)
	return nil
}

func (s *computeCmds) runInstancesSetDiskAutoDelete(cmd *cobra.Command, args []string) error {
	if !s.flagAutoDelete && !s.flagNoAutoDelete {
		return fmt.Errorf("one of --auto-delete or --no-auto-delete must be specified")
	}

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

	autoDelete := s.flagAutoDelete
	op, err := svc.Instances.SetDiskAutoDelete(project, zone, instance, autoDelete, s.flagAutoDeleteDisk).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting disk auto-delete: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Updated auto-delete for disk [%s] on instance [%s].\n", s.flagAutoDeleteDisk, instance)
	return nil
}
