package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud artifacts operations (#1076) ---

var artifactsOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Artifact Registry long-running operations",
}

var (
	flagArtOpsLocation string
	flagArtOpsFormat   string
)

var artifactsOperationsDescribeCmd = &cobra.Command{
	Use:   "describe OPERATION",
	Short: "Describe an Artifact Registry operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsOperationsDescribe,
}

func init() {
	artifactsOperationsDescribeCmd.Flags().StringVar(&flagArtOpsLocation, "location", "", "Location of the operation")
	artifactsOperationsDescribeCmd.Flags().StringVar(&flagArtOpsFormat, "format", "", "Output format")

	artifactsOperationsCmd.AddCommand(artifactsOperationsDescribeCmd)
	artifactsCmd.AddCommand(artifactsOperationsCmd)
}

func runArtifactsOperationsDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	if flagArtOpsLocation == "" && !isFullOperationName(name) {
		return fmt.Errorf("--location is required (or pass a full projects/.../operations/... name)")
	}
	if !isFullOperationName(name) {
		project, err := resolveProject()
		if err != nil {
			return err
		}
		name = fmt.Sprintf("%s/operations/%s", artLocationParent(project, flagArtOpsLocation), name)
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagArtOpsFormat)
}

func isFullOperationName(s string) bool {
	return len(s) > 9 && s[:9] == "projects/"
}
