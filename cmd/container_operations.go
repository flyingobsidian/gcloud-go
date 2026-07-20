package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
)

// --- gcloud container operations (#1141) ---

var containerOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage GKE long-running operations"}

var (
	flagCtnOpLocation string
	flagCtnOpFormat   string
)

var (
	containerOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a GKE operation",
		Args: cobra.ExactArgs(1), RunE: runCtnOpCancel,
	}
	containerOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a GKE operation",
		Args: cobra.ExactArgs(1), RunE: runCtnOpDescribe,
	}
	containerOpListCmd = &cobra.Command{
		Use: "list", Short: "List GKE operations in a location",
		Args: cobra.NoArgs, RunE: runCtnOpList,
	}
)

func init() {
	all := []*cobra.Command{containerOpCancelCmd, containerOpDescribeCmd, containerOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagCtnOpLocation, "location", "", "Location (region or zone) (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagCtnOpFormat, "format", "", "Output format")
	}

	containerOperationsCmd.AddCommand(all...)
	containerCmd.AddCommand(containerOperationsCmd)
}

func ctnOpParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagCtnOpLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagCtnOpLocation), nil
}

func ctnOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := ctnOpParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func runCtnOpCancel(cmd *cobra.Command, args []string) error {
	name, err := ctnOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &container.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runCtnOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := ctnOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagCtnOpFormat)
}

func runCtnOpList(cmd *cobra.Command, args []string) error {
	parent, err := ctnOpParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	return emitFormatted(resp.Operations, flagCtnOpFormat)
}
