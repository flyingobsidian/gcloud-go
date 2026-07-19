package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database db-system-initial-storage-sizes (#1283, #1563) ---

var oracleDbSystemInitialStorageSizesCmd = &cobra.Command{
	Use:   "db-system-initial-storage-sizes",
	Short: "Manage Oracle DB system initial storage sizes",
}

var (
	flagOdbSizLocation string
	flagOdbSizFormat   string
	flagOdbSizPageSize int64
)

var oracleSizListCmd = &cobra.Command{
	Use: "list", Short: "List Oracle DB system initial storage sizes in a location",
	Args: cobra.NoArgs, RunE: runOracleSizList,
}

func init() {
	all := []*cobra.Command{oracleSizListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbSizLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbSizFormat, "format", "", "Output format")
	}
	oracleSizListCmd.Flags().Int64Var(&flagOdbSizPageSize, "page-size", 0, "Maximum results per page")

	oracleDbSystemInitialStorageSizesCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleDbSystemInitialStorageSizesCmd)
}

func runOracleSizList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbSizLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DbSystemInitialStorageSize
	pageToken := ""
	for {
		call := svc.Projects.Locations.DbSystemInitialStorageSizes.List(parent).Context(ctx)
		if flagOdbSizPageSize > 0 {
			call = call.PageSize(flagOdbSizPageSize)
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
	return emitFormatted(all, flagOdbSizFormat)
}
