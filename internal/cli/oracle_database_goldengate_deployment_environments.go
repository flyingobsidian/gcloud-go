package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-deployment-environments (#1272) ---

func (s *oracleDatabaseCmds) initOracleDatabaseGoldengateDeploymentEnvironments() {
	all := []*cobra.Command{s.oracleGgDeCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbGgDeLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbGgDeFormat, "format", "", "Output format")
	}
	s.oracleGgDeCmd.Flags().Int64Var(&s.flagOdbGgDePageSize, "page-size", 0, "Maximum results per page")

	s.oracleGgDeployEnvsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleGgDeployEnvsCmd)
}

func (s *oracleDatabaseCmds) runOracleGgDeList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgDeLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateDeploymentEnvironment
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateDeploymentEnvironments.List(parent).Context(ctx)
		if s.flagOdbGgDePageSize > 0 {
			call = call.PageSize(s.flagOdbGgDePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate deployment environments: %w", err)
		}
		all = append(all, resp.GoldengateDeploymentEnvironments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbGgDeFormat)
}
