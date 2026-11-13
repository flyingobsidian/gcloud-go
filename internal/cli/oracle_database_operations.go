package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database operations (#1277) ---

func (s *oracleDatabaseCmds) initOracleDatabaseOperations() {
	all := []*cobra.Command{s.oracleOpCancelCmd, s.oracleOpDeleteCmd, s.oracleOpDescribeCmd, s.oracleOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbOpFormat, "format", "", "Output format")
	}
	s.oracleOpListCmd.Flags().StringVar(&s.flagOdbOpFilter, "filter", "", "Server-side filter expression")
	s.oracleOpListCmd.Flags().Int64Var(&s.flagOdbOpPageSize, "page-size", 0, "Maximum results per page")

	s.oracleOperationsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleOperationsCmd)
}

func (s *oracleDatabaseCmds) oracleOpName(id string) (string, error) {
	return s.odbResource(s.flagOdbOpLocation, "operations", id)
}

func (s *oracleDatabaseCmds) runOracleOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.oracleOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &oracledatabase.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *oracleDatabaseCmds) runOracleOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *oracleDatabaseCmds) runOracleOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbOpFormat)
}

func (s *oracleDatabaseCmds) runOracleOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbOpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if s.flagOdbOpFilter != "" {
			call = call.Filter(s.flagOdbOpFilter)
		}
		if s.flagOdbOpPageSize > 0 {
			call = call.PageSize(s.flagOdbOpPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbOpFormat)
}
