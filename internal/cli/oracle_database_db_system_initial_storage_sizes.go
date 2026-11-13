package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database db-system-initial-storage-sizes (#1283, #1563) ---

func (s *oracleDatabaseCmds) initOracleDatabaseDbSystemInitialStorageSizes() {
	all := []*cobra.Command{s.oracleSizListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagOdbSizLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagOdbSizFormat, "format", "", "Output format")
	}
	s.oracleSizListCmd.Flags().Int64Var(&s.flagOdbSizPageSize, "page-size", 0, "Maximum results per page")

	s.oracleDbSystemInitialStorageSizesCmd.AddCommand(all...)
	s.oracleDatabaseCmd.AddCommand(s.oracleDbSystemInitialStorageSizesCmd)
}

func (s *oracleDatabaseCmds) runOracleSizList(cmd *cobra.Command, args []string) error {
	parent, err := s.odbLocationParent(s.flagOdbSizLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().OracleDatabase(ctx)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DbSystemInitialStorageSize
	pageToken := ""
	for {
		call := svc.Projects.Locations.DbSystemInitialStorageSizes.List(parent).Context(ctx)
		if s.flagOdbSizPageSize > 0 {
			call = call.PageSize(s.flagOdbSizPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing db system initial storage sizes: %w", err)
		}
		all = append(all, resp.DbSystemInitialStorageSizes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagOdbSizFormat)
}
