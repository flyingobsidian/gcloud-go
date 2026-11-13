package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database db-versions (#1285, #1565) ---

func (s *oracleDatabaseCmds) initOracleDatabaseDbVersions() {
	all := []*cobra.Command{s.oracleDbvListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbDbvLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbDbvFormat, "format", "", "Output format")
	}
	s.oracleDbvListCmd.Flags().StringVar(&s.flagOdbDbvFilter, "filter", "", "Server-side filter expression")
	s.oracleDbvListCmd.Flags().Int64Var(&s.flagOdbDbvPageSize, "page-size", 0, "Maximum results per page")

	s.oracleDbVersionsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleDbVersionsCmd)
}

func (s *oracleDatabaseCmds) runOracleDbvList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbDbvLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DbVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.DbVersions.List(parent).Context(ctx)
		if s.flagOdbDbvFilter != "" {
			call = call.Filter(s.flagOdbDbvFilter)
		}
		if s.flagOdbDbvPageSize > 0 {
			call = call.PageSize(s.flagOdbDbvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing db versions: %w", err)
		}
		all = append(all, resp.DbVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbDbvFormat)
}
