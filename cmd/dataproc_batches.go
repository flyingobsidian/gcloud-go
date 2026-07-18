package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc batches (#1511) ---

var dpBatchesCmd = &cobra.Command{Use: "batches", Short: "Manage Dataproc serverless batches"}

var (
	flagDPBatchRegion     string
	flagDPBatchFormat     string
	flagDPBatchConfigFile string
	flagDPBatchRequestID  string
	flagDPBatchFilter     string
	flagDPBatchOrderBy    string
	flagDPBatchPageSize   int64
	flagDPBatchWaitTO     time.Duration
)

var (
	dpBatchCancelCmd = &cobra.Command{
		Use: "cancel BATCH", Short: "Cancel a batch (via its underlying operation)",
		Args: cobra.ExactArgs(1), RunE: runDPBatchCancel,
	}
	dpBatchDeleteCmd = &cobra.Command{
		Use: "delete BATCH", Short: "Delete a batch",
		Args: cobra.ExactArgs(1), RunE: runDPBatchDelete,
	}
	dpBatchDescribeCmd = &cobra.Command{
		Use: "describe BATCH", Short: "Describe a batch",
		Args: cobra.ExactArgs(1), RunE: runDPBatchDescribe,
	}
	dpBatchListCmd = &cobra.Command{
		Use: "list", Short: "List batches",
		Args: cobra.NoArgs, RunE: runDPBatchList,
	}
	dpBatchSubmitCmd = &cobra.Command{
		Use: "submit BATCH", Short: "Submit a batch (body loaded from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runDPBatchSubmit,
	}
	dpBatchWaitCmd = &cobra.Command{
		Use: "wait BATCH", Short: "Wait for a batch to reach a terminal state",
		Args: cobra.ExactArgs(1), RunE: runDPBatchWait,
	}
)

func init() {
	all := []*cobra.Command{
		dpBatchCancelCmd, dpBatchDeleteCmd, dpBatchDescribeCmd,
		dpBatchListCmd, dpBatchSubmitCmd, dpBatchWaitCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDPBatchRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDPBatchFormat, "format", "", "Output format")
	}
	dpBatchSubmitCmd.Flags().StringVar(&flagDPBatchConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Batch body (required)")
	_ = dpBatchSubmitCmd.MarkFlagRequired("config-file")
	dpBatchSubmitCmd.Flags().StringVar(&flagDPBatchRequestID, "request-id", "",
		"Optional client-supplied ID for idempotent submission")
	dpBatchListCmd.Flags().StringVar(&flagDPBatchFilter, "filter", "", "Server-side filter expression")
	dpBatchListCmd.Flags().StringVar(&flagDPBatchOrderBy, "order-by", "", "Order-by expression")
	dpBatchListCmd.Flags().Int64Var(&flagDPBatchPageSize, "page-size", 0, "Maximum results per page")
	dpBatchWaitCmd.Flags().DurationVar(&flagDPBatchWaitTO, "timeout", 30*time.Minute,
		"Maximum time to wait for the batch to reach a terminal state")

	dpBatchesCmd.AddCommand(all...)
	dataprocCmd.AddCommand(dpBatchesCmd)
}

func dpBatchParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return dpLocationParent(project, flagDPBatchRegion), nil
}

func dpBatchName(id string) (string, error) {
	parent, err := dpBatchParent()
	if err != nil {
		return "", err
	}
	return dpChild("batches", id, parent), nil
}

func runDPBatchCancel(cmd *cobra.Command, args []string) error {
	name, err := dpBatchName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPBatchRegion)
	if err != nil {
		return err
	}
	batch, err := svc.Projects.Locations.Batches.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing batch: %w", err)
	}
	if batch.Operation == "" {
		return fmt.Errorf("batch [%s] has no underlying operation to cancel (already terminal?)", args[0])
	}
	if _, err := svc.Projects.Regions.Operations.Cancel(batch.Operation).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling batch: %w", err)
	}
	fmt.Printf("Cancel request issued for batch [%s] (operation: %s).\n", args[0], batch.Operation)
	return nil
}

func runDPBatchDelete(cmd *cobra.Command, args []string) error {
	name, err := dpBatchName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPBatchRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Batches.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting batch: %w", err)
	}
	fmt.Printf("Deleted batch [%s].\n", args[0])
	return nil
}

func runDPBatchDescribe(cmd *cobra.Command, args []string) error {
	name, err := dpBatchName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPBatchRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Batches.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing batch: %w", err)
	}
	return emitFormatted(got, flagDPBatchFormat)
}

func runDPBatchList(cmd *cobra.Command, args []string) error {
	parent, err := dpBatchParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPBatchRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.Batch
	pageToken := ""
	for {
		call := svc.Projects.Locations.Batches.List(parent).Context(ctx)
		if flagDPBatchFilter != "" {
			call = call.Filter(flagDPBatchFilter)
		}
		if flagDPBatchOrderBy != "" {
			call = call.OrderBy(flagDPBatchOrderBy)
		}
		if flagDPBatchPageSize > 0 {
			call = call.PageSize(flagDPBatchPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing batches: %w", err)
		}
		all = append(all, resp.Batches...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDPBatchFormat)
}

func runDPBatchSubmit(cmd *cobra.Command, args []string) error {
	parent, err := dpBatchParent()
	if err != nil {
		return err
	}
	body := &dataproc.Batch{}
	if err := loadYAMLOrJSONInto(flagDPBatchConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPBatchRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Batches.Create(parent, body).BatchId(args[0]).Context(ctx)
	if flagDPBatchRequestID != "" {
		call = call.RequestId(flagDPBatchRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("submitting batch: %w", err)
	}
	fmt.Printf("Submit request issued for batch [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPBatchFormat)
}

func runDPBatchWait(cmd *cobra.Command, args []string) error {
	name, err := dpBatchName(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), flagDPBatchWaitTO)
	defer cancel()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPBatchRegion)
	if err != nil {
		return err
	}
	backoff := 5 * time.Second
	for {
		batch, err := svc.Projects.Locations.Batches.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling batch: %w", err)
		}
		switch batch.State {
		case "SUCCEEDED":
			fmt.Printf("Batch [%s] succeeded.\n", args[0])
			return emitFormatted(batch, flagDPBatchFormat)
		case "FAILED", "CANCELLED":
			return fmt.Errorf("batch [%s] ended in state %s", args[0], batch.State)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for batch %s: %w", args[0], ctx.Err())
		case <-time.After(backoff):
		}
		if backoff < 60*time.Second {
			backoff = time.Duration(float64(backoff) * 1.5)
		}
	}
}
