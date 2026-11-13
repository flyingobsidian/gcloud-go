package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc batches (#1511) ---

func (s *dataprocCmds) initDataprocBatches() {
	all := []*cobra.Command{
		s.dpBatchCancelCmd, s.dpBatchDeleteCmd, s.dpBatchDescribeCmd,
		s.dpBatchListCmd, s.dpBatchSubmitCmd, s.dpBatchWaitCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDPBatchRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDPBatchFormat, "format", "", "Output format")
	}
	s.dpBatchSubmitCmd.Flags().StringVar(&s.flagDPBatchConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Batch body (required)")
	_ = s.dpBatchSubmitCmd.MarkFlagRequired("config-file")
	s.dpBatchSubmitCmd.Flags().StringVar(&s.flagDPBatchRequestID, "request-id", "",
		"Optional client-supplied ID for idempotent submission")
	s.dpBatchListCmd.Flags().StringVar(&s.flagDPBatchFilter, "filter", "", "Server-side filter expression")
	s.dpBatchListCmd.Flags().StringVar(&s.flagDPBatchOrderBy, "order-by", "", "Order-by expression")
	s.dpBatchListCmd.Flags().Int64Var(&s.flagDPBatchPageSize, "page-size", 0, "Maximum results per page")
	s.dpBatchWaitCmd.Flags().DurationVar(&s.flagDPBatchWaitTO, "timeout", 30*time.Minute,
		"Maximum time to wait for the batch to reach a terminal state")

	s.dpBatchesCmd.AddCommand(all...)
	s.dataprocCmd.AddCommand(s.dpBatchesCmd)
}

func (s *dataprocCmds) dpBatchParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return dpLocationParent(project, s.flagDPBatchRegion), nil
}

func (s *dataprocCmds) dpBatchName(id string) (string, error) {
	parent, err := s.dpBatchParent()
	if err != nil {
		return "", err
	}
	return dpChild("batches", id, parent), nil
}

func (s *dataprocCmds) runDPBatchCancel(cmd *cobra.Command, args []string) error {
	name, err := s.dpBatchName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPBatchRegion)
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
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for batch [%s] (operation: %s).\n", args[0], batch.Operation)
	return nil
}

func (s *dataprocCmds) runDPBatchDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dpBatchName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPBatchRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Batches.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting batch: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted batch [%s].\n", args[0])
	return nil
}

func (s *dataprocCmds) runDPBatchDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dpBatchName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPBatchRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Batches.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing batch: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPBatchFormat)
}

func (s *dataprocCmds) runDPBatchList(cmd *cobra.Command, args []string) error {
	parent, err := s.dpBatchParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPBatchRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.Batch
	pageToken := ""
	for {
		call := svc.Projects.Locations.Batches.List(parent).Context(ctx)
		if s.flagDPBatchFilter != "" {
			call = call.Filter(s.flagDPBatchFilter)
		}
		if s.flagDPBatchOrderBy != "" {
			call = call.OrderBy(s.flagDPBatchOrderBy)
		}
		if s.flagDPBatchPageSize > 0 {
			call = call.PageSize(s.flagDPBatchPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDPBatchFormat)
}

func (s *dataprocCmds) runDPBatchSubmit(cmd *cobra.Command, args []string) error {
	parent, err := s.dpBatchParent()
	if err != nil {
		return err
	}
	body := &dataproc.Batch{}
	if err := loadYAMLOrJSONInto(s.flagDPBatchConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPBatchRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Batches.Create(parent, body).BatchId(args[0]).Context(ctx)
	if s.flagDPBatchRequestID != "" {
		call = call.RequestId(s.flagDPBatchRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("submitting batch: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Submit request issued for batch [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPBatchFormat)
}

func (s *dataprocCmds) runDPBatchWait(cmd *cobra.Command, args []string) error {
	name, err := s.dpBatchName(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), s.flagDPBatchWaitTO)
	defer cancel()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPBatchRegion)
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
			fmt.Fprintf(s.inv.stdout, "Batch [%s] succeeded.\n", args[0])
			return emitFormattedTo(s.inv.stdout, batch, s.flagDPBatchFormat)
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
