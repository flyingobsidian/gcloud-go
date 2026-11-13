package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database databases (#1281) ---

func (s *oracleDatabaseCmds) initOracleDatabaseDatabases() {
	all := []*cobra.Command{s.oracleDbDescribeCmd, s.oracleDbListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbDbLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbDbFormat, "format", "", "Output format")
	}
	s.oracleDbListCmd.Flags().Int64Var(&s.flagOdbDbPageSize, "page-size", 0, "Maximum results per page")
	s.oracleDbListCmd.Flags().StringVar(&s.flagOdbDbFilter, "filter", "", "Server-side filter expression")

	s.oracleDatabasesCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleDatabasesCmd)
}

func (s *oracleDatabaseCmds) oracleDbName(id string) (string, error) {
	return s.odbResource(s.flagOdbDbLocation, "databases", id)
}

func (s *oracleDatabaseCmds) runOracleDbDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleDbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Databases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing database: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbDbFormat)
}

func (s *oracleDatabaseCmds) runOracleDbList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbDbLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.Database
	pageToken := ""
	for {
		call := svc.Projects.Locations.Databases.List(parent).Context(ctx)
		if s.flagOdbDbPageSize > 0 {
			call = call.PageSize(s.flagOdbDbPageSize)
		}
		if s.flagOdbDbFilter != "" {
			call = call.Filter(s.flagOdbDbFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing databases: %w", err)
		}
		all = append(all, resp.Databases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbDbFormat)
}
