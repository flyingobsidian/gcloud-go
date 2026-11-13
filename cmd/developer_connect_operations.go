package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	developerconnect "google.golang.org/api/developerconnect/v1"
)

// --- gcloud developer-connect operations (#1025) ---

func (s *developerConnectCmds) initDeveloperConnectOperations() {
	all := []*cobra.Command{
		s.developerConnectOpCancelCmd, s.developerConnectOpDeleteCmd,
		s.developerConnectOpDescribeCmd, s.developerConnectOpListCmd,
		s.developerConnectOpWaitCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDcOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDcOpFormat, "format", "", "Output format")
	}
	s.developerConnectOpListCmd.Flags().StringVar(&s.flagDcOpFilter, "filter", "", "Server-side filter expression")
	s.developerConnectOpListCmd.Flags().Int64Var(&s.flagDcOpPageSize, "page-size", 0, "Maximum results per page")
	s.developerConnectOpWaitCmd.Flags().DurationVar(&s.flagDcOpTimeout, "timeout", 30*time.Minute, "Maximum time to wait for the operation to finish")

	s.developerConnectOperationsCmd.AddCommand(all...)
	s.developerConnectCmd.AddCommand(s.developerConnectOperationsCmd)
}

func (s *developerConnectCmds) dcOpName(id string) (string, error) {
	return s.devConnResourceName(s.flagDcOpLocation, "operations", id)
}

func (s *developerConnectCmds) runDcOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.dcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &developerconnect.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *developerConnectCmds) runDcOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *developerConnectCmds) runDcOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDcOpFormat)
}

func (s *developerConnectCmds) runDcOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.devConnLocationParent(s.flagDcOpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	var all []*developerconnect.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if s.flagDcOpFilter != "" {
			call = call.Filter(s.flagDcOpFilter)
		}
		if s.flagDcOpPageSize > 0 {
			call = call.PageSize(s.flagDcOpPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDcOpFormat)
}

func (s *developerConnectCmds) runDcOpWait(cmd *cobra.Command, args []string) error {
	name, err := s.dcOpName(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), s.flagDcOpTimeout)
	defer cancel()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
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
			return emitFormattedTo(s.inv.stdout, op, s.flagDcOpFormat)
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
