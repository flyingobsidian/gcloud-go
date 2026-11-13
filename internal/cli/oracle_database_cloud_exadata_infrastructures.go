package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database cloud-exadata-infrastructures (#1264) ---

func (s *oracleDatabaseCmds) initOracleDatabaseCloudExadataInfrastructures() {
	all := []*cobra.Command{
		s.oracleExInfCreateCmd, s.oracleExInfDeleteCmd,
		s.oracleExInfDescribeCmd, s.oracleExInfListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbExInfLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbExInfFormat, "format", "", "Output format")
	}
	s.oracleExInfCreateCmd.Flags().StringVar(&s.flagOdbExInfConfigFile, "config-file", "", "YAML/JSON file with cloud Exadata infrastructure body (required)")
	_ = s.oracleExInfCreateCmd.MarkFlagRequired("config-file")
	s.oracleExInfListCmd.Flags().Int64Var(&s.flagOdbExInfPageSize, "page-size", 0, "Maximum results per page")
	s.oracleExInfListCmd.Flags().StringVar(&s.flagOdbExInfFilter, "filter", "", "Server-side filter expression")

	s.oracleCloudExadataInfrastructuresCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleCloudExadataInfrastructuresCmd)
}

func (s *oracleDatabaseCmds) oracleExInfName(id string) (string, error) {
	return s.odbResource(s.flagOdbExInfLocation, "cloudExadataInfrastructures", id)
}

func (s *oracleDatabaseCmds) runOracleExInfCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbExInfLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.CloudExadataInfrastructure{}
	if err := loadYAMLOrJSONInto(s.flagOdbExInfConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CloudExadataInfrastructures.Create(parent, body).CloudExadataInfrastructureId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cloud Exadata infrastructure: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create cloud Exadata infrastructure [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbExInfFormat)
}

func (s *oracleDatabaseCmds) runOracleExInfDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleExInfName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CloudExadataInfrastructures.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cloud Exadata infrastructure: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete cloud Exadata infrastructure [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbExInfFormat)
}

func (s *oracleDatabaseCmds) runOracleExInfDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleExInfName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CloudExadataInfrastructures.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cloud Exadata infrastructure: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbExInfFormat)
}

func (s *oracleDatabaseCmds) runOracleExInfList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbExInfLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.CloudExadataInfrastructure
	pageToken := ""
	for {
		call := svc.Projects.Locations.CloudExadataInfrastructures.List(parent).Context(ctx)
		if s.flagOdbExInfPageSize > 0 {
			call = call.PageSize(s.flagOdbExInfPageSize)
		}
		if s.flagOdbExInfFilter != "" {
			call = call.Filter(s.flagOdbExInfFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing cloud Exadata infrastructures: %w", err)
		}
		all = append(all, resp.CloudExadataInfrastructures...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbExInfFormat)
}
