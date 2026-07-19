package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	developerconnect "google.golang.org/api/developerconnect/v1"
)

// --- gcloud developer-connect operations (#1025) ---

var developerConnectOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Developer Connect long-running operations"}

var (
	flagDcOpLocation string
	flagDcOpFormat   string
	flagDcOpFilter   string
	flagDcOpPageSize int64
	flagDcOpTimeout  time.Duration
)

var (
	developerConnectOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Developer Connect operation",
		Args: cobra.ExactArgs(1), RunE: runDcOpCancel,
	}
	developerConnectOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Developer Connect operation",
		Args: cobra.ExactArgs(1), RunE: runDcOpDelete,
	}
	developerConnectOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Developer Connect operation",
		Args: cobra.ExactArgs(1), RunE: runDcOpDescribe,
	}
	developerConnectOpListCmd = &cobra.Command{
		Use: "list", Short: "List Developer Connect operations in a location",
		Args: cobra.NoArgs, RunE: runDcOpList,
	}
	developerConnectOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Developer Connect operation to complete",
		Args: cobra.ExactArgs(1), RunE: runDcOpWait,
	}
)

func init() {
	all := []*cobra.Command{
		developerConnectOpCancelCmd, developerConnectOpDeleteCmd,
		developerConnectOpDescribeCmd, developerConnectOpListCmd,
		developerConnectOpWaitCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDcOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDcOpFormat, "format", "", "Output format")
	}
	developerConnectOpListCmd.Flags().StringVar(&flagDcOpFilter, "filter", "", "Server-side filter expression")
	developerConnectOpListCmd.Flags().Int64Var(&flagDcOpPageSize, "page-size", 0, "Maximum results per page")
	developerConnectOpWaitCmd.Flags().DurationVar(&flagDcOpTimeout, "timeout", 30*time.Minute, "Maximum time to wait for the operation to finish")

	developerConnectOperationsCmd.AddCommand(all...)
	developerConnectCmd.AddCommand(developerConnectOperationsCmd)
}

func dcOpName(id string) (string, error) {
	return devConnResourceName(flagDcOpLocation, "operations", id)
}

func runDcOpCancel(cmd *cobra.Command, args []string) error {
	name, err := dcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &developerconnect.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runDcOpDelete(cmd *cobra.Command, args []string) error {
	name, err := dcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runDcOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagDcOpFormat)
}

func runDcOpList(cmd *cobra.Command, args []string) error {
	parent, err := devConnLocationParent(flagDcOpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*developerconnect.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagDcOpFilter != "" {
			call = call.Filter(flagDcOpFilter)
		}
		if flagDcOpPageSize > 0 {
			call = call.PageSize(flagDcOpPageSize)
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
	return emitFormatted(all, flagDcOpFormat)
}

func runDcOpWait(cmd *cobra.Command, args []string) error {
	name, err := dcOpName(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), flagDcOpTimeout)
	defer cancel()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
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
			return emitFormatted(op, flagDcOpFormat)
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
