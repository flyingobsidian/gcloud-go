package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services operations (#1000) ---

var networkServicesOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Network Services long-running operations"}

var (
	flagNsOpLocation string
	flagNsOpFormat   string
	flagNsOpFilter   string
	flagNsOpPageSize int64
	flagNsOpTimeout  time.Duration
)

var (
	networkServicesOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Network Services operation",
		Args: cobra.ExactArgs(1), RunE: runNsOpCancel,
	}
	networkServicesOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Network Services operation",
		Args: cobra.ExactArgs(1), RunE: runNsOpDescribe,
	}
	networkServicesOpListCmd = &cobra.Command{
		Use: "list", Short: "List Network Services operations in a location",
		Args: cobra.NoArgs, RunE: runNsOpList,
	}
	networkServicesOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Network Services operation to complete",
		Args: cobra.ExactArgs(1), RunE: runNsOpWait,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesOpCancelCmd, networkServicesOpDescribeCmd,
		networkServicesOpListCmd, networkServicesOpWaitCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsOpFormat, "format", "", "Output format")
	}
	networkServicesOpListCmd.Flags().StringVar(&flagNsOpFilter, "filter", "", "Server-side filter expression")
	networkServicesOpListCmd.Flags().Int64Var(&flagNsOpPageSize, "page-size", 0, "Maximum results per page")
	networkServicesOpWaitCmd.Flags().DurationVar(&flagNsOpTimeout, "timeout", 30*time.Minute, "Maximum time to wait for the operation to finish")

	networkServicesOperationsCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesOperationsCmd)
}

func nsOpName(id string) (string, error) {
	return nsResourceName(flagNsOpLocation, "operations", id)
}

func runNsOpCancel(cmd *cobra.Command, args []string) error {
	name, err := nsOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &networkservices.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runNsOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagNsOpFormat)
}

func runNsOpList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsOpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagNsOpFilter != "" {
			call = call.Filter(flagNsOpFilter)
		}
		if flagNsOpPageSize > 0 {
			call = call.PageSize(flagNsOpPageSize)
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
	return emitFormatted(all, flagNsOpFormat)
}

func runNsOpWait(cmd *cobra.Command, args []string) error {
	name, err := nsOpName(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), flagNsOpTimeout)
	defer cancel()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	backoff := time.Second
	for {
		op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation failed: %s (code %d)", op.Error.Message, op.Error.Code)
			}
			return emitFormatted(op, flagNsOpFormat)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for operation [%s]", args[0])
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}
