package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database gi-versions (#1268) ---

func (s *oracleDatabaseCmds) initOracleDatabaseGiVersions() {
	all := []*cobra.Command{s.oracleGiListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbGiLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbGiFormat, "format", "", "Output format")
	}
	s.oracleGiListCmd.Flags().StringVar(&s.flagOdbGiFilter, "filter", "", "Server-side filter expression")
	s.oracleGiListCmd.Flags().Int64Var(&s.flagOdbGiPageSize, "page-size", 0, "Maximum results per page")

	s.oracleGiVersionsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleGiVersionsCmd)
}

func (s *oracleDatabaseCmds) runOracleGiList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGiLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GiVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.GiVersions.List(parent).Context(ctx)
		if s.flagOdbGiFilter != "" {
			call = call.Filter(s.flagOdbGiFilter)
		}
		if s.flagOdbGiPageSize > 0 {
			call = call.PageSize(s.flagOdbGiPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing GI versions: %w", err)
		}
		all = append(all, resp.GiVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbGiFormat)
}
