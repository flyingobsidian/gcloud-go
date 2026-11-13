package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-deployment-types (#1273) ---

func (s *oracleDatabaseCmds) initOracleDatabaseGoldengateDeploymentTypes() {
	all := []*cobra.Command{s.oracleGgDtListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbGgDtLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbGgDtFormat, "format", "", "Output format")
	}
	s.oracleGgDtListCmd.Flags().Int64Var(&s.flagOdbGgDtPageSize, "page-size", 0, "Maximum results per page")
	s.oracleGgDtListCmd.Flags().StringVar(&s.flagOdbGgDtFilter, "filter", "", "Server-side filter expression")

	s.oracleGgDeployTypesCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleGgDeployTypesCmd)
}

func (s *oracleDatabaseCmds) runOracleGgDtList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgDtLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateDeploymentType
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateDeploymentTypes.List(parent).Context(ctx)
		if s.flagOdbGgDtPageSize > 0 {
			call = call.PageSize(s.flagOdbGgDtPageSize)
		}
		if s.flagOdbGgDtFilter != "" {
			call = call.Filter(s.flagOdbGgDtFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate deployment types: %w", err)
		}
		all = append(all, resp.GoldengateDeploymentTypes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbGgDtFormat)
}
