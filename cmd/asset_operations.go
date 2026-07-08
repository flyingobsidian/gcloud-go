package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

var assetOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Cloud Asset Inventory operations",
}

var assetOperationDescribeCmd = &cobra.Command{
	Use:   "describe OPERATION_NAME",
	Short: "Describe a Cloud Asset Inventory operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetOperationDescribe,
}

func init() {
	assetOperationsCmd.AddCommand(assetOperationDescribeCmd)
	assetCmd.AddCommand(assetOperationsCmd)
}

func runAssetOperationDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return yamlEncode(op)
}
