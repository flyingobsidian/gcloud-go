package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database exadb-vm-clusters (#1286, dupe #1566) ---

func (s *oracleDatabaseCmds) initOracleDatabaseExadbVmClusters() {
	all := []*cobra.Command{
		s.oracleExaVmCreateCmd, s.oracleExaVmDeleteCmd,
		s.oracleExaVmDescribeCmd, s.oracleExaVmListCmd, s.oracleExaVmUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbExaVmLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbExaVmFormat, "format", "", "Output format")
	}
	s.oracleExaVmCreateCmd.Flags().StringVar(&s.flagOdbExaVmConfigFile, "config-file", "", "YAML/JSON file with ExaDB VM cluster body (required)")
	_ = s.oracleExaVmCreateCmd.MarkFlagRequired("config-file")
	s.oracleExaVmUpdateCmd.Flags().StringVar(&s.flagOdbExaVmConfigFile, "config-file", "", "YAML/JSON file with ExaDB VM cluster body (required)")
	_ = s.oracleExaVmUpdateCmd.MarkFlagRequired("config-file")
	s.oracleExaVmUpdateCmd.Flags().StringVar(&s.flagOdbExaVmUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	s.oracleExaVmListCmd.Flags().Int64Var(&s.flagOdbExaVmPageSize, "page-size", 0, "Maximum results per page")
	s.oracleExaVmListCmd.Flags().StringVar(&s.flagOdbExaVmFilter, "filter", "", "Server-side filter expression")

	s.oracleExadbVmClustersCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleExadbVmClustersCmd)
}

func (s *oracleDatabaseCmds) oracleExaVmName(id string) (string, error) {
	return s.odbResource(s.flagOdbExaVmLocation, "exadbVmClusters", id)
}

func (s *oracleDatabaseCmds) runOracleExaVmCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbExaVmLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.ExadbVmCluster{}
	if err := loadYAMLOrJSONInto(s.flagOdbExaVmConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExadbVmClusters.Create(parent, body).ExadbVmClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ExaDB VM cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create ExaDB VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbExaVmFormat)
}

func (s *oracleDatabaseCmds) runOracleExaVmDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleExaVmName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExadbVmClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting ExaDB VM cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete ExaDB VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbExaVmFormat)
}

func (s *oracleDatabaseCmds) runOracleExaVmDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleExaVmName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExadbVmClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ExaDB VM cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbExaVmFormat)
}

func (s *oracleDatabaseCmds) runOracleExaVmList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbExaVmLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.ExadbVmCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.ExadbVmClusters.List(parent).Context(ctx)
		if s.flagOdbExaVmPageSize > 0 {
			call = call.PageSize(s.flagOdbExaVmPageSize)
		}
		if s.flagOdbExaVmFilter != "" {
			call = call.Filter(s.flagOdbExaVmFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ExaDB VM clusters: %w", err)
		}
		all = append(all, resp.ExadbVmClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbExaVmFormat)
}

func (s *oracleDatabaseCmds) runOracleExaVmUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.oracleExaVmName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.ExadbVmCluster{}
	if err := loadYAMLOrJSONInto(s.flagOdbExaVmConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagOdbExaVmUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExadbVmClusters.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating ExaDB VM cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update ExaDB VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbExaVmFormat)
}
