package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

var servicesOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Service Usage operations",
}

var servicesOperationDescribeCmd = &cobra.Command{
	Use:   "describe OPERATION_NAME",
	Short: "Describe a Service Usage operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runServicesOperationDescribe,
}

var servicesOperationWaitCmd = &cobra.Command{
	Use:   "wait OPERATION_NAME",
	Short: "Wait for a Service Usage operation to complete",
	Args:  cobra.ExactArgs(1),
	RunE:  runServicesOperationWait,
}

var flagServicesOpWaitInterval time.Duration

func init() {
	servicesOperationWaitCmd.Flags().DurationVar(&flagServicesOpWaitInterval, "poll-interval", 2*time.Second, "Interval between poll attempts")
	servicesOperationsCmd.AddCommand(servicesOperationDescribeCmd, servicesOperationWaitCmd)
	servicesCmd.AddCommand(servicesOperationsCmd)
}

func runServicesOperationDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.ServiceUsageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return yamlEncode(op)
}

func runServicesOperationWait(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.ServiceUsageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	for {
		op, err := svc.Operations.Get(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			return yamlEncode(op)
		}
		time.Sleep(flagServicesOpWaitInterval)
	}
}
