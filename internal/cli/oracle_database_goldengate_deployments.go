package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-deployments (#1275) ---

func (s *oracleDatabaseCmds) initOracleDatabaseGoldengateDeployments() {
	all := []*cobra.Command{
		s.oracleGgDpCreateCmd, s.oracleGgDpDeleteCmd,
		s.oracleGgDpDescribeCmd, s.oracleGgDpListCmd,
		s.oracleGgDpStartCmd, s.oracleGgDpStopCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbGgDpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbGgDpFormat, "format", "", "Output format")
	}
	s.oracleGgDpCreateCmd.Flags().StringVar(&s.flagOdbGgDpConfigFile, "config-file", "", "YAML/JSON file with GoldengateDeployment body (required)")
	_ = s.oracleGgDpCreateCmd.MarkFlagRequired("config-file")
	s.oracleGgDpListCmd.Flags().Int64Var(&s.flagOdbGgDpPageSize, "page-size", 0, "Maximum results per page")
	s.oracleGgDpListCmd.Flags().StringVar(&s.flagOdbGgDpFilter, "filter", "", "Server-side filter expression")

	s.oracleGgDeploymentsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleGgDeploymentsCmd)
}

func (s *oracleDatabaseCmds) oracleGgDpName(id string) (string, error) {
	return s.odbResource(s.flagOdbGgDpLocation, "goldengateDeployments", id)
}

func (s *oracleDatabaseCmds) runOracleGgDpCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgDpLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.GoldengateDeployment{}
	if err := loadYAMLOrJSONInto(s.flagOdbGgDpConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateDeployments.Create(parent, body).GoldengateDeploymentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating goldengate deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create goldengate deployment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbGgDpFormat)
}

func (s *oracleDatabaseCmds) runOracleGgDpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgDpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateDeployments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting goldengate deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete goldengate deployment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbGgDpFormat)
}

func (s *oracleDatabaseCmds) runOracleGgDpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgDpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GoldengateDeployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing goldengate deployment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagOdbGgDpFormat)
}

func (s *oracleDatabaseCmds) runOracleGgDpList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgDpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateDeployment
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateDeployments.List(parent).Context(ctx)
		if s.flagOdbGgDpPageSize > 0 {
			call = call.PageSize(s.flagOdbGgDpPageSize)
		}
		if s.flagOdbGgDpFilter != "" {
			call = call.Filter(s.flagOdbGgDpFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate deployments: %w", err)
		}
		all = append(all, resp.GoldengateDeployments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbGgDpFormat)
}

func (s *oracleDatabaseCmds) runOracleGgDpStart(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgDpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateDeployments.Start(name, &oracledatabase.StartGoldengateDeploymentRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting goldengate deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start goldengate deployment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbGgDpFormat)
}

func (s *oracleDatabaseCmds) runOracleGgDpStop(cmd *cobra.Command, args []string) error {
	name, err := s.oracleGgDpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateDeployments.Stop(name, &oracledatabase.StopGoldengateDeploymentRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping goldengate deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Stop goldengate deployment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagOdbGgDpFormat)
}
