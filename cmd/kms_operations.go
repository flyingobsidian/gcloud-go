package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud kms operations (#1109) ---

var kmsOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Cloud KMS long-running operations",
}

var (
	flagKmsOpLocation string
	flagKmsOpFormat   string
)

var kmsOpDescribeCmd = &cobra.Command{
	Use:   "describe OPERATION",
	Short: "Describe a Cloud KMS long-running operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsOpDescribe,
}

func init() {
	kmsOpDescribeCmd.Flags().StringVar(&flagKmsOpLocation, "location", "", "Location (required)")
	kmsOpDescribeCmd.Flags().StringVar(&flagKmsOpFormat, "format", "", "Output format")
	_ = kmsOpDescribeCmd.MarkFlagRequired("location")
	kmsOperationsCmd.AddCommand(kmsOpDescribeCmd)
	kmsCmd.AddCommand(kmsOperationsCmd)
}

func runKmsOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsOpLocation) + "/operations"
	name := kmsFullName(parent, args[0])
	out, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(out, flagKmsOpFormat)
}
