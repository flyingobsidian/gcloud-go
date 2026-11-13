package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database db-systems (#1284, dupe #1564) ---

func (s *oracleDatabaseCmds) initOracleDatabaseDbSystems() {
	all := []*cobra.Command{
		s.oracleDbsCreateCmd, s.oracleDbsDeleteCmd,
		s.oracleDbsDescribeCmd, s.oracleDbsListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbDbsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbDbsFormat, "format", "", "Output format")
	}
	s.oracleDbsCreateCmd.Flags().StringVar(&s.flagOdbDbsConfigFile, "config-file", "", "YAML/JSON file with DB system body (required)")
	_ = s.oracleDbsCreateCmd.MarkFlagRequired("config-file")
	s.oracleDbsListCmd.Flags().Int64Var(&s.flagOdbDbsPageSize, "page-size", 0, "Maximum results per page")
	s.oracleDbsListCmd.Flags().StringVar(&s.flagOdbDbsFilter, "filter", "", "Server-side filter expression")

	s.oracleDbSystemsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleDbSystemsCmd)
}

func (s *oracleDatabaseCmds) oracleDbsName(id string) (string, error) {
	return s.odbResource(s.flagOdbDbsLocation, "dbSystems", id)
}

func (s *oracleDatabaseCmds) runOracleDbsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbDbsLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.DbSystem{}
	if err := loadYAMLOrJSONInto(s.flagOdbDbsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DbSystems.Create(parent, body).DbSystemId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DB system: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create DB system [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbDbsFormat)
}

func (s *oracleDatabaseCmds) runOracleDbsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleDbsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DbSystems.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting DB system: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete DB system [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbDbsFormat)
}

func (s *oracleDatabaseCmds) runOracleDbsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleDbsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DbSystems.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DB system: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbDbsFormat)
}

func (s *oracleDatabaseCmds) runOracleDbsList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbDbsLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DbSystem
	pageToken := ""
	for {
		call := svc.Projects.Locations.DbSystems.List(parent).Context(ctx)
		if s.flagOdbDbsPageSize > 0 {
			call = call.PageSize(s.flagOdbDbsPageSize)
		}
		if s.flagOdbDbsFilter != "" {
			call = call.Filter(s.flagOdbDbsFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing DB systems: %w", err)
		}
		all = append(all, resp.DbSystems...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbDbsFormat)
}
