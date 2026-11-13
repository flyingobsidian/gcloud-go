package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-connections (#1271) ---

func (s *oracleDatabaseCmds) initOracleDatabaseGoldengateConnections() {
	all := []*cobra.Command{
		s.oracleGgCnCreateCmd, s.oracleGgCnDeleteCmd,
		s.oracleGgCnDescribeCmd, s.oracleGgCnListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbGgCnLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbGgCnFormat, "format", "", "Output format")
	}
	s.oracleGgCnCreateCmd.Flags().StringVar(&s.flagOdbGgCnConfigFile, "config-file", "", "YAML/JSON file with GoldengateConnection body (required)")
	_ = s.oracleGgCnCreateCmd.MarkFlagRequired("config-file")
	s.oracleGgCnListCmd.Flags().Int64Var(&s.flagOdbGgCnPageSize, "page-size", 0, "Maximum results per page")
	s.oracleGgCnListCmd.Flags().StringVar(&s.flagOdbGgCnFilter, "filter", "", "Server-side filter expression")

	s.oracleGgConnectionsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleGgConnectionsCmd)
}

func (s *oracleDatabaseCmds) oracleGgCnName(id string) (string, error) {
	return s.odbResource(s.flagOdbGgCnLocation, "goldengateConnections", id)
}

func (s *oracleDatabaseCmds) runOracleGgCnCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgCnLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.GoldengateConnection{}
	if err := loadYAMLOrJSONInto(s.flagOdbGgCnConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateConnections.Create(parent, body).GoldengateConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating goldengate connection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create goldengate connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbGgCnFormat)
}

func (s *oracleDatabaseCmds) runOracleGgCnDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgCnName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateConnections.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting goldengate connection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete goldengate connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbGgCnFormat)
}

func (s *oracleDatabaseCmds) runOracleGgCnDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgCnName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GoldengateConnections.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing goldengate connection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbGgCnFormat)
}

func (s *oracleDatabaseCmds) runOracleGgCnList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgCnLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateConnection
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateConnections.List(parent).Context(ctx)
		if s.flagOdbGgCnPageSize > 0 {
			call = call.PageSize(s.flagOdbGgCnPageSize)
		}
		if s.flagOdbGgCnFilter != "" {
			call = call.Filter(s.flagOdbGgCnFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate connections: %w", err)
		}
		all = append(all, resp.GoldengateConnections...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbGgCnFormat)
}
