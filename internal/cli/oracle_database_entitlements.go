package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database entitlements (#1267) ---

func (s *oracleDatabaseCmds) initOracleDatabaseEntitlements() {
	all := []*cobra.Command{s.oracleEntListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbEntLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbEntFormat, "format", "", "Output format")
	}
	s.oracleEntListCmd.Flags().Int64Var(&s.flagOdbEntPageSize, "page-size", 0, "Maximum results per page")

	s.oracleEntitlementsCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleEntitlementsCmd)
}

func (s *oracleDatabaseCmds) runOracleEntList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbEntLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.Entitlement
	pageToken := ""
	for {
		call := svc.Projects.Locations.Entitlements.List(parent).Context(ctx)
		if s.flagOdbEntPageSize > 0 {
			call = call.PageSize(s.flagOdbEntPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing entitlements: %w", err)
		}
		all = append(all, resp.Entitlements...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbEntFormat)
}
