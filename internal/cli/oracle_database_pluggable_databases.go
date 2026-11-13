package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database pluggable-databases (#1278) ---

func (s *oracleDatabaseCmds) initOracleDatabasePluggableDatabases() {
	all := []*cobra.Command{s.oraclePdbDescribeCmd, s.oraclePdbListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbPdbLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbPdbFormat, "format", "", "Output format")
	}
	s.oraclePdbListCmd.Flags().Int64Var(&s.flagOdbPdbPageSize, "page-size", 0, "Maximum results per page")
	s.oraclePdbListCmd.Flags().StringVar(&s.flagOdbPdbFilter, "filter", "", "Server-side filter expression")

	s.oraclePluggableDatabasesCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oraclePluggableDatabasesCmd)
}

func (s *oracleDatabaseCmds) oraclePdbName(id string) (string, error) {
	return s.odbResource(s.flagOdbPdbLocation, "pluggableDatabases", id)
}

func (s *oracleDatabaseCmds) runOraclePdbDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oraclePdbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PluggableDatabases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing pluggable database: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbPdbFormat)
}

func (s *oracleDatabaseCmds) runOraclePdbList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbPdbLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.PluggableDatabase
	pageToken := ""
	for {
		call := svc.Projects.Locations.PluggableDatabases.List(parent).Context(ctx)
		if s.flagOdbPdbPageSize > 0 {
			call = call.PageSize(s.flagOdbPdbPageSize)
		}
		if s.flagOdbPdbFilter != "" {
			call = call.Filter(s.flagOdbPdbFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing pluggable databases: %w", err)
		}
		all = append(all, resp.PluggableDatabases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbPdbFormat)
}
