package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-deployment-versions (#1274) ---

func (s *oracleDatabaseCmds) initOracleDatabaseGoldengateDeploymentVersions() {
	all := []*cobra.Command{s.oracleGgDvListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbGgDvLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbGgDvFormat, "format", "", "Output format")
	}
	s.oracleGgDvListCmd.Flags().Int64Var(&s.flagOdbGgDvPageSize, "page-size", 0, "Maximum results per page")
	s.oracleGgDvListCmd.Flags().StringVar(&s.flagOdbGgDvFilter, "filter", "", "Server-side filter expression")

	s.oracleGgDeployVersionsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleGgDeployVersionsCmd)
}

func (s *oracleDatabaseCmds) runOracleGgDvList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbGgDvLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateDeploymentVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateDeploymentVersions.List(parent).Context(ctx)
		if s.flagOdbGgDvPageSize > 0 {
			call = call.PageSize(s.flagOdbGgDvPageSize)
		}
		if s.flagOdbGgDvFilter != "" {
			call = call.Filter(s.flagOdbGgDvFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate deployment versions: %w", err)
		}
		all = append(all, resp.GoldengateDeploymentVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbGgDvFormat)
}
