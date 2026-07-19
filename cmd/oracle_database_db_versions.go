package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database db-versions (#1285, #1565) ---

var oracleDbVersionsCmd = &cobra.Command{
	Use:   "db-versions",
	Short: "Manage Oracle DB versions",
}

var (
	flagOdbDbvLocation string
	flagOdbDbvFormat   string
	flagOdbDbvFilter   string
	flagOdbDbvPageSize int64
)

var oracleDbvListCmd = &cobra.Command{
	Use: "list", Short: "List Oracle DB versions in a location",
	Args: cobra.NoArgs, RunE: runOracleDbvList,
}

func init() {
	all := []*cobra.Command{oracleDbvListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbDbvLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbDbvFormat, "format", "", "Output format")
	}
	oracleDbvListCmd.Flags().StringVar(&flagOdbDbvFilter, "filter", "", "Server-side filter expression")
	oracleDbvListCmd.Flags().Int64Var(&flagOdbDbvPageSize, "page-size", 0, "Maximum results per page")

	oracleDbVersionsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleDbVersionsCmd)
}

func runOracleDbvList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbDbvLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DbVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.DbVersions.List(parent).Context(ctx)
		if flagOdbDbvFilter != "" {
			call = call.Filter(flagOdbDbvFilter)
		}
		if flagOdbDbvPageSize > 0 {
			call = call.PageSize(flagOdbDbvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing db versions: %w", err)
		}
		all = append(all, resp.DbVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbDbvFormat)
}
