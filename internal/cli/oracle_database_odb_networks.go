package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database odb-networks (#1276) ---

func (s *oracleDatabaseCmds) initOracleDatabaseOdbNetworks() {
	all := []*cobra.Command{
		s.oracleOdbNetCreateCmd, s.oracleOdbNetDeleteCmd,
		s.oracleOdbNetDescribeCmd, s.oracleOdbNetListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbNetLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbNetFormat, "format", "", "Output format")
	}
	s.oracleOdbNetCreateCmd.Flags().StringVar(&s.flagOdbNetConfigFile, "config-file", "", "YAML/JSON file with ODB network body (required)")
	_ = s.oracleOdbNetCreateCmd.MarkFlagRequired("config-file")
	s.oracleOdbNetListCmd.Flags().Int64Var(&s.flagOdbNetPageSize, "page-size", 0, "Maximum results per page")
	s.oracleOdbNetListCmd.Flags().StringVar(&s.flagOdbNetFilter, "filter", "", "Server-side filter expression")

	s.oracleOdbNetworksCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleOdbNetworksCmd)
}

func (s *oracleDatabaseCmds) oracleOdbNetName(id string) (string, error) {
	return s.odbResource(s.flagOdbNetLocation, "odbNetworks", id)
}

func (s *oracleDatabaseCmds) runOracleOdbNetCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbNetLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.OdbNetwork{}
	if err := loadYAMLOrJSONInto(s.flagOdbNetConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.OdbNetworks.Create(parent, body).OdbNetworkId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ODB network: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create ODB network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbNetFormat)
}

func (s *oracleDatabaseCmds) runOracleOdbNetDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleOdbNetName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.OdbNetworks.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting ODB network: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete ODB network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbNetFormat)
}

func (s *oracleDatabaseCmds) runOracleOdbNetDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleOdbNetName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OdbNetworks.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ODB network: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbNetFormat)
}

func (s *oracleDatabaseCmds) runOracleOdbNetList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbNetLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.OdbNetwork
	pageToken := ""
	for {
		call := svc.Projects.Locations.OdbNetworks.List(parent).Context(ctx)
		if s.flagOdbNetPageSize > 0 {
			call = call.PageSize(s.flagOdbNetPageSize)
		}
		if s.flagOdbNetFilter != "" {
			call = call.Filter(s.flagOdbNetFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ODB networks: %w", err)
		}
		all = append(all, resp.OdbNetworks...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbNetFormat)
}
