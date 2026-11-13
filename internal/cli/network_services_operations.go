package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services operations (#1000) ---

func (s *networkServicesCmds) initNetworkServicesOperations() {
	all := []*cobra.Command{
		s.networkServicesOpCancelCmd, s.networkServicesOpDescribeCmd,
		s.networkServicesOpListCmd, s.networkServicesOpWaitCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsOpFormat, "format", "", "Output format")
	}
	s.networkServicesOpListCmd.Flags().StringVar(&s.flagNsOpFilter, "filter", "", "Server-side filter expression")
	s.networkServicesOpListCmd.Flags().Int64Var(&s.flagNsOpPageSize, "page-size", 0, "Maximum results per page")
	s.networkServicesOpWaitCmd.Flags().DurationVar(&s.flagNsOpTimeout, "timeout", 30*time.Minute, "Maximum time to wait for the operation to finish")

	s.networkServicesOperationsCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesOperationsCmd)
}

func (s *networkServicesCmds) nsOpName(id string) (string, error) {
	return s.nsResourceName(s.flagNsOpLocation, "operations", id)
}

func (s *networkServicesCmds) runNsOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.nsOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &networkservices.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *networkServicesCmds) runNsOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsOpFormat)
}

func (s *networkServicesCmds) runNsOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsOpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if s.flagNsOpFilter != "" {
			call = call.Filter(s.flagNsOpFilter)
		}
		if s.flagNsOpPageSize > 0 {
			call = call.PageSize(s.flagNsOpPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagNsOpFormat)
}

func (s *networkServicesCmds) runNsOpWait(cmd *cobra.Command, args []string) error {
	name, err := s.nsOpName(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), s.flagNsOpTimeout)
	defer cancel()
	svc, err := s.inv.clients().NetworkServices(ctx)
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
			return emitFormattedTo(s.inv.stdout, op, s.flagNsOpFormat)
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
