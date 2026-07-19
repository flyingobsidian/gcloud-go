package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub operations (#1163) ---

var apihubOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage API Hub long-running operations"}

var (
	flagApihubOpLocation string
	flagApihubOpFormat   string
	flagApihubOpFilter   string
	flagApihubOpPageSize int64
)

var (
	apihubOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an API Hub operation",
		Args: cobra.ExactArgs(1), RunE: runApihubOpCancel,
	}
	apihubOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an API Hub operation",
		Args: cobra.ExactArgs(1), RunE: runApihubOpDelete,
	}
	apihubOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an API Hub operation",
		Args: cobra.ExactArgs(1), RunE: runApihubOpDescribe,
	}
	apihubOpListCmd = &cobra.Command{
		Use: "list", Short: "List API Hub operations in a location",
		Args: cobra.NoArgs, RunE: runApihubOpList,
	}
)

func init() {
	all := []*cobra.Command{apihubOpCancelCmd, apihubOpDeleteCmd, apihubOpDescribeCmd, apihubOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubOpFormat, "format", "", "Output format")
	}
	apihubOpListCmd.Flags().StringVar(&flagApihubOpFilter, "filter", "", "Server-side filter expression")
	apihubOpListCmd.Flags().Int64Var(&flagApihubOpPageSize, "page-size", 0, "Maximum results per page")

	apihubOperationsCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubOperationsCmd)
}

func apihubOpName(id string) (string, error) {
	return apihubResource(flagApihubOpLocation, "operations", id)
}

func runApihubOpCancel(cmd *cobra.Command, args []string) error {
	name, err := apihubOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &apihub.GoogleLongrunningCancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runApihubOpDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runApihubOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagApihubOpFormat)
}

func runApihubOpList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubOpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagApihubOpFilter != "" {
			call = call.Filter(flagApihubOpFilter)
		}
		if flagApihubOpPageSize > 0 {
			call = call.PageSize(flagApihubOpPageSize)
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
	return emitFormatted(all, flagApihubOpFormat)
}
