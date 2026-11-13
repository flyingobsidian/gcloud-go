package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage operations (#1244) ---
//
// Storage bucket-scoped long-running operations.

func (s *storageCmds) initStorageOperations() {
	all := []*cobra.Command{s.storageOpCancelCmd, s.storageOpDescribeCmd, s.storageOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagStOpBucket, "bucket", "", "Bucket that owns the operation (required)")
		_ = c.MarkFlagRequired("bucket")
		c.Flags().StringVar(&s.flagStOpFormat, "format", "", "Output format")
	}
	s.storageOpListCmd.Flags().StringVar(&s.flagStOpFilter, "filter", "", "Server-side filter expression")
	s.storageOpListCmd.Flags().Int64Var(&s.flagStOpPageSize, "page-size", 0, "Maximum results per page")

	s.storageOperationsCmd.AddCommand(all...)
	s.storageCmd.AddCommand(s.storageOperationsCmd)
}

func (s *storageCmds) runStOpCancel(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	if err := svc.Operations.Cancel(s.flagStOpBucket, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s] in bucket [%s].\n", args[0], s.flagStOpBucket)
	return nil
}

func (s *storageCmds) runStOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Operations.Get(s.flagStOpBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStOpFormat)
}

func (s *storageCmds) runStOpList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	var all []*storage.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Operations.List(s.flagStOpBucket).Context(ctx)
		if s.flagStOpFilter != "" {
			call = call.Filter(s.flagStOpFilter)
		}
		if s.flagStOpPageSize > 0 {
			call = call.PageSize(s.flagStOpPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagStOpFormat)
}
