package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database exascale-db-storage-vaults (#1287, dupe #1567) ---

func (s *oracleDatabaseCmds) initOracleDatabaseExascaleDbStorageVaults() {
	all := []*cobra.Command{
		s.oracleExaStCreateCmd, s.oracleExaStDeleteCmd,
		s.oracleExaStDescribeCmd, s.oracleExaStListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbExaStLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbExaStFormat, "format", "", "Output format")
	}
	s.oracleExaStCreateCmd.Flags().StringVar(&s.flagOdbExaStConfigFile, "config-file", "", "YAML/JSON file with Exascale DB storage vault body (required)")
	_ = s.oracleExaStCreateCmd.MarkFlagRequired("config-file")
	s.oracleExaStListCmd.Flags().Int64Var(&s.flagOdbExaStPageSize, "page-size", 0, "Maximum results per page")
	s.oracleExaStListCmd.Flags().StringVar(&s.flagOdbExaStFilter, "filter", "", "Server-side filter expression")

	s.oracleExascaleDbStorageVaultsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleExascaleDbStorageVaultsCmd)
}

func (s *oracleDatabaseCmds) oracleExaStName(id string) (string, error) {
	return s.odbResource(s.flagOdbExaStLocation, "exascaleDbStorageVaults", id)
}

func (s *oracleDatabaseCmds) runOracleExaStCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbExaStLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.ExascaleDbStorageVault{}
	if err := loadYAMLOrJSONInto(s.flagOdbExaStConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExascaleDbStorageVaults.Create(parent, body).ExascaleDbStorageVaultId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating Exascale DB storage vault: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create Exascale DB storage vault [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbExaStFormat)
}

func (s *oracleDatabaseCmds) runOracleExaStDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleExaStName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExascaleDbStorageVaults.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting Exascale DB storage vault: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete Exascale DB storage vault [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbExaStFormat)
}

func (s *oracleDatabaseCmds) runOracleExaStDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleExaStName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExascaleDbStorageVaults.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Exascale DB storage vault: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbExaStFormat)
}

func (s *oracleDatabaseCmds) runOracleExaStList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbExaStLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.ExascaleDbStorageVault
	pageToken := ""
	for {
		call := svc.Projects.Locations.ExascaleDbStorageVaults.List(parent).Context(ctx)
		if s.flagOdbExaStPageSize > 0 {
			call = call.PageSize(s.flagOdbExaStPageSize)
		}
		if s.flagOdbExaStFilter != "" {
			call = call.Filter(s.flagOdbExaStFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing Exascale DB storage vaults: %w", err)
		}
		all = append(all, resp.ExascaleDbStorageVaults...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbExaStFormat)
}
