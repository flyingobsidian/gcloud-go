package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database autonomous-db-versions (#1279) ---

func (s *oracleDatabaseCmds) initOracleDatabaseAutonomousDbVersions() {
	all := []*cobra.Command{s.oracleAdvListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbAdvLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbAdvFormat, "format", "", "Output format")
	}
	s.oracleAdvListCmd.Flags().Int64Var(&s.flagOdbAdvPageSize, "page-size", 0, "Maximum results per page")

	s.oracleAutonomousDbVersionsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleAutonomousDbVersionsCmd)
}

func (s *oracleDatabaseCmds) runOracleAdvList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbAdvLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.AutonomousDbVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.AutonomousDbVersions.List(parent).Context(ctx)
		if s.flagOdbAdvPageSize > 0 {
			call = call.PageSize(s.flagOdbAdvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing autonomous db versions: %w", err)
		}
		all = append(all, resp.AutonomousDbVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbAdvFormat)
}
