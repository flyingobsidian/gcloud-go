package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database db-system-shapes (#1266) ---

func (s *oracleDatabaseCmds) initOracleDatabaseDbSystemShapes() {
	all := []*cobra.Command{s.oracleShapeListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbShapeLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbShapeFormat, "format", "", "Output format")
	}
	s.oracleShapeListCmd.Flags().StringVar(&s.flagOdbShapeFilter, "filter", "", "Server-side filter expression")
	s.oracleShapeListCmd.Flags().Int64Var(&s.flagOdbShapePageSize, "page-size", 0, "Maximum results per page")

	s.oracleDbSystemShapesCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleDbSystemShapesCmd)
}

func (s *oracleDatabaseCmds) runOracleShapeList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbShapeLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DbSystemShape
	pageToken := ""
	for {
		call := svc.Projects.Locations.DbSystemShapes.List(parent).Context(ctx)
		if s.flagOdbShapeFilter != "" {
			call = call.Filter(s.flagOdbShapeFilter)
		}
		if s.flagOdbShapePageSize > 0 {
			call = call.PageSize(s.flagOdbShapePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing db system shapes: %w", err)
		}
		all = append(all, resp.DbSystemShapes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbShapeFormat)
}
