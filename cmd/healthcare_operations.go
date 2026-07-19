package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare operations (#1225) ---

var healthcareOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Healthcare long-running operations"}

var (
	flagHcOpLocation string
	flagHcOpDataset  string
	flagHcOpFormat   string
	flagHcOpFilter   string
	flagHcOpPageSize int64
)

var (
	healthcareOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Cloud Healthcare operation",
		Args: cobra.ExactArgs(1), RunE: runHcOpCancel,
	}
	healthcareOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Cloud Healthcare operation",
		Args: cobra.ExactArgs(1), RunE: runHcOpDescribe,
	}
	healthcareOpListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Healthcare operations for a dataset",
		Args: cobra.NoArgs, RunE: runHcOpList,
	}
)

func init() {
	all := []*cobra.Command{healthcareOpCancelCmd, healthcareOpDescribeCmd, healthcareOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagHcOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagHcOpDataset, "dataset", "", "Dataset that owns the operation (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&flagHcOpFormat, "format", "", "Output format")
	}
	healthcareOpListCmd.Flags().StringVar(&flagHcOpFilter, "filter", "", "Server-side filter expression")
	healthcareOpListCmd.Flags().Int64Var(&flagHcOpPageSize, "page-size", 0, "Maximum results per page")

	healthcareOperationsCmd.AddCommand(all...)
	healthcareCmd.AddCommand(healthcareOperationsCmd)
}

func hcOpName(id string) (string, error) {
	dataset, err := hcDatasetName(flagHcOpLocation, flagHcOpDataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", dataset, id), nil
}

func runHcOpCancel(cmd *cobra.Command, args []string) error {
	name, err := hcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.Operations.Cancel(name, &healthcare.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runHcOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := hcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagHcOpFormat)
}

func runHcOpList(cmd *cobra.Command, args []string) error {
	dataset, err := hcDatasetName(flagHcOpLocation, flagHcOpDataset)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*healthcare.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.Operations.List(dataset).Context(ctx)
		if flagHcOpFilter != "" {
			call = call.Filter(flagHcOpFilter)
		}
		if flagHcOpPageSize > 0 {
			call = call.PageSize(flagHcOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagHcOpFormat)
}
