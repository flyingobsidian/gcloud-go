package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-connection-types (#1270) ---

func (s *oracleDatabaseCmds) initOracleDatabaseGoldengateConnectionTypes() {
	all := []*cobra.Command{s.oracleGgCtListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbGgCtLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbGgCtFormat, "format", "", "Output format")
	}
	s.oracleGgCtListCmd.Flags().Int64Var(&s.flagOdbGgCtPageSize, "page-size", 0, "Maximum results per page")
	s.oracleGgCtListCmd.Flags().StringVar(&s.flagOdbGgCtFilter, "filter", "", "Server-side filter expression")

	s.oracleGgConnTypesCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleGgConnTypesCmd)
}

func (s *oracleDatabaseCmds) runOracleGgCtList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgCtLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateConnectionType
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateConnectionTypes.List(parent).Context(ctx)
		if s.flagOdbGgCtPageSize > 0 {
			call = call.PageSize(s.flagOdbGgCtPageSize)
		}
		if s.flagOdbGgCtFilter != "" {
			call = call.Filter(s.flagOdbGgCtFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate connection types: %w", err)
		}
		all = append(all, resp.GoldengateConnectionTypes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbGgCtFormat)
}
