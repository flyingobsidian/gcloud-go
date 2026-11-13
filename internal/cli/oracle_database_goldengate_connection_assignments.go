package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-connection-assignments (#1269) ---

func (s *oracleDatabaseCmds) initOracleDatabaseGoldengateConnectionAssignments() {
	all := []*cobra.Command{
		s.oracleGgCaCreateCmd, s.oracleGgCaDeleteCmd,
		s.oracleGgCaDescribeCmd, s.oracleGgCaListCmd, s.oracleGgCaTestCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbGgCaLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbGgCaFormat, "format", "", "Output format")
	}
	s.oracleGgCaCreateCmd.Flags().StringVar(&s.flagOdbGgCaConfigFile, "config-file", "", "YAML/JSON file with GoldengateConnectionAssignment body (required)")
	_ = s.oracleGgCaCreateCmd.MarkFlagRequired("config-file")
	s.oracleGgCaListCmd.Flags().Int64Var(&s.flagOdbGgCaPageSize, "page-size", 0, "Maximum results per page")
	s.oracleGgCaListCmd.Flags().StringVar(&s.flagOdbGgCaFilter, "filter", "", "Server-side filter expression")

	s.oracleGgConnAssignmentsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleGgConnAssignmentsCmd)
}

func (s *oracleDatabaseCmds) oracleGgCaName(id string) (string, error) {
	return s.odbResource(s.flagOdbGgCaLocation, "goldengateConnectionAssignments", id)
}

func (s *oracleDatabaseCmds) runOracleGgCaCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgCaLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.GoldengateConnectionAssignment{}
	if err := loadYAMLOrJSONInto(s.flagOdbGgCaConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateConnectionAssignments.Create(parent, body).GoldengateConnectionAssignmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating goldengate connection assignment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create goldengate connection assignment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbGgCaFormat)
}

func (s *oracleDatabaseCmds) runOracleGgCaDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgCaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateConnectionAssignments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting goldengate connection assignment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete goldengate connection assignment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbGgCaFormat)
}

func (s *oracleDatabaseCmds) runOracleGgCaDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgCaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GoldengateConnectionAssignments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing goldengate connection assignment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbGgCaFormat)
}

func (s *oracleDatabaseCmds) runOracleGgCaList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgCaLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateConnectionAssignment
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateConnectionAssignments.List(parent).Context(ctx)
		if s.flagOdbGgCaPageSize > 0 {
			call = call.PageSize(s.flagOdbGgCaPageSize)
		}
		if s.flagOdbGgCaFilter != "" {
			call = call.Filter(s.flagOdbGgCaFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate connection assignments: %w", err)
		}
		all = append(all, resp.GoldengateConnectionAssignments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbGgCaFormat)
}

func (s *oracleDatabaseCmds) runOracleGgCaTest(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgCaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.GoldengateConnectionAssignments.Test(name, &oracledatabase.TestGoldengateConnectionAssignmentRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("testing goldengate connection assignment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Test goldengate connection assignment [%s] completed.\n", args[0])
	return emitFormattedTo(s.inv.stdout, resp, s.flagOdbGgCaFormat)
}
