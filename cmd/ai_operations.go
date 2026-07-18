package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud ai operations (#1459) ---

var aiOpsCmd = &cobra.Command{Use: "operations", Short: "Manage Vertex AI long-running operations"}

var (
	flagAIOpsRegion string
	flagAIOpsFormat string
)

var aiOpsDescribeCmd = &cobra.Command{
	Use: "describe OPERATION", Short: "Describe a Vertex AI long-running operation",
	Args: cobra.ExactArgs(1), RunE: runAIOpsDescribe,
}

func init() {
	aiOpsDescribeCmd.Flags().StringVar(&flagAIOpsRegion, "region", "", "Region where the operation lives (required)")
	_ = aiOpsDescribeCmd.MarkFlagRequired("region")
	aiOpsDescribeCmd.Flags().StringVar(&flagAIOpsFormat, "format", "", "Output format")
	aiOpsCmd.AddCommand(aiOpsDescribeCmd)
	aiCmd.AddCommand(aiOpsCmd)
}

func aiOpsName(id string) (string, error) {
	parent, err := aiParent(flagAIOpsRegion)
	if err != nil {
		return "", err
	}
	return aiChild("operations", id, parent), nil
}

func runAIOpsDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiOpsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIOpsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagAIOpsFormat)
}
