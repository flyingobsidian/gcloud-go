package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud bms operations (#1229) ---

var bmsOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage bare metal long-running operations"}

var (
	flagBmsOpLocation string
	flagBmsOpFormat   string
)

var (
	bmsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a bare metal long-running operation",
		Args: cobra.ExactArgs(1), RunE: runBmsOpDescribe,
	}
)

func init() {
	all := []*cobra.Command{bmsOpDescribeCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagBmsOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagBmsOpFormat, "format", "", "Output format")
	}

	bmsOperationsCmd.AddCommand(all...)
	bmsCmd.AddCommand(bmsOperationsCmd)
}

func bmsOpName(id string) (string, error) {
	return bmsResource(flagBmsOpLocation, "operations", id)
}

func runBmsOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := bmsOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagBmsOpFormat)
}
