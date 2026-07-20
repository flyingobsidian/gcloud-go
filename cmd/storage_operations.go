package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage operations (#1244) ---
//
// Storage bucket-scoped long-running operations.

var storageOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage bucket-scoped storage long-running operations"}

var (
	flagStOpBucket   string
	flagStOpFormat   string
	flagStOpFilter   string
	flagStOpPageSize int64
)

var (
	storageOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a bucket-scoped storage operation",
		Args: cobra.ExactArgs(1), RunE: runStOpCancel,
	}
	storageOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a bucket-scoped storage operation",
		Args: cobra.ExactArgs(1), RunE: runStOpDescribe,
	}
	storageOpListCmd = &cobra.Command{
		Use: "list", Short: "List bucket-scoped storage operations",
		Args: cobra.NoArgs, RunE: runStOpList,
	}
)

func init() {
	all := []*cobra.Command{storageOpCancelCmd, storageOpDescribeCmd, storageOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagStOpBucket, "bucket", "", "Bucket that owns the operation (required)")
		_ = c.MarkFlagRequired("bucket")
		c.Flags().StringVar(&flagStOpFormat, "format", "", "Output format")
	}
	storageOpListCmd.Flags().StringVar(&flagStOpFilter, "filter", "", "Server-side filter expression")
	storageOpListCmd.Flags().Int64Var(&flagStOpPageSize, "page-size", 0, "Maximum results per page")

	storageOperationsCmd.AddCommand(all...)
	storageCmd.AddCommand(storageOperationsCmd)
}

func runStOpCancel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if err := svc.Operations.Cancel(flagStOpBucket, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s] in bucket [%s].\n", args[0], flagStOpBucket)
	return nil
}

func runStOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Operations.Get(flagStOpBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagStOpFormat)
}

func runStOpList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*storage.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Operations.List(flagStOpBucket).Context(ctx)
		if flagStOpFilter != "" {
			call = call.Filter(flagStOpFilter)
		}
		if flagStOpPageSize > 0 {
			call = call.PageSize(flagStOpPageSize)
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
	return emitFormatted(all, flagStOpFormat)
}
