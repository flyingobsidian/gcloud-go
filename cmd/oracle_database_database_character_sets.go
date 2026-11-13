package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database database-character-sets (#1280) ---

func (s *oracleDatabaseCmds) initOracleDatabaseDatabaseCharacterSets() {
	all := []*cobra.Command{s.oracleDbCsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbDbCsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbDbCsFormat, "format", "", "Output format")
	}
	s.oracleDbCsListCmd.Flags().StringVar(&s.flagOdbDbCsFilter, "filter", "", "Server-side filter expression")
	s.oracleDbCsListCmd.Flags().Int64Var(&s.flagOdbDbCsPageSize, "page-size", 0, "Maximum results per page")

	s.oracleDatabaseCharacterSetsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleDatabaseCharacterSetsCmd)
}

func (s *oracleDatabaseCmds) runOracleDbCsList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbDbCsLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DatabaseCharacterSet
	pageToken := ""
	for {
		call := svc.Projects.Locations.DatabaseCharacterSets.List(parent).Context(ctx)
		if s.flagOdbDbCsFilter != "" {
			call = call.Filter(s.flagOdbDbCsFilter)
		}
		if s.flagOdbDbCsPageSize > 0 {
			call = call.PageSize(s.flagOdbDbCsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing database character sets: %w", err)
		}
		all = append(all, resp.DatabaseCharacterSets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbDbCsFormat)
}
