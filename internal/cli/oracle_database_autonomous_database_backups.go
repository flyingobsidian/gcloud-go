package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database autonomous-database-backups (#1261) ---
//
// The v1 client only exposes List on this service, so the subgroup is
// list-only. describe/create/delete are not available.

func (s *oracleDatabaseCmds) initOracleDatabaseAutonomousDatabaseBackups() {
	s.oracleAdbBkListCmd.Flags().StringVar(&s.flagOdbAdbBkLocation, "location", "", "Location (required)")
	_ = s.oracleAdbBkListCmd.MarkFlagRequired("location")
	s.oracleAdbBkListCmd.Flags().StringVar(&s.flagOdbAdbBkFormat, "format", "", "Output format")
	s.oracleAdbBkListCmd.Flags().StringVar(&s.flagOdbAdbBkFilter, "filter", "", "Server-side filter expression")
	s.oracleAdbBkListCmd.Flags().Int64Var(&s.flagOdbAdbBkPageSize, "page-size", 0, "Maximum results per page")

	s.oracleAutonomousDatabaseBackupsCmd.AddCommand(s.oracleAdbBkListCmd)
	s.oracleDatabaseCmd.AddCommand(s.oracleAutonomousDatabaseBackupsCmd)
}

func (s *oracleDatabaseCmds) runOracleAdbBkList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbAdbBkLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.AutonomousDatabaseBackup
	pageToken := ""
	for {
		call := svc.Projects.Locations.AutonomousDatabaseBackups.List(parent).Context(ctx)
		if s.flagOdbAdbBkFilter != "" {
			call = call.Filter(s.flagOdbAdbBkFilter)
		}
		if s.flagOdbAdbBkPageSize > 0 {
			call = call.PageSize(s.flagOdbAdbBkPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing autonomous database backups: %w", err)
		}
		all = append(all, resp.AutonomousDatabaseBackups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbAdbBkFormat)
}
