package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database autonomous-database-character-sets (#1262) ---

func (s *oracleDatabaseCmds) initOracleDatabaseAutonomousDatabaseCharacterSets() {
	all := []*cobra.Command{s.oracleAdCsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbAdCsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbAdCsFormat, "format", "", "Output format")
	}
	s.oracleAdCsListCmd.Flags().StringVar(&s.flagOdbAdCsFilter, "filter", "", "Server-side filter expression")
	s.oracleAdCsListCmd.Flags().Int64Var(&s.flagOdbAdCsPageSize, "page-size", 0, "Maximum results per page")

	s.oracleAutonomousDatabaseCharacterSetsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleAutonomousDatabaseCharacterSetsCmd)
}

func (s *oracleDatabaseCmds) runOracleAdCsList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbAdCsLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.AutonomousDatabaseCharacterSet
	pageToken := ""
	for {
		call := svc.Projects.Locations.AutonomousDatabaseCharacterSets.List(parent).Context(ctx)
		if s.flagOdbAdCsFilter != "" {
			call = call.Filter(s.flagOdbAdCsFilter)
		}
		if s.flagOdbAdCsPageSize > 0 {
			call = call.PageSize(s.flagOdbAdCsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing autonomous database character sets: %w", err)
		}
		all = append(all, resp.AutonomousDatabaseCharacterSets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbAdCsFormat)
}
