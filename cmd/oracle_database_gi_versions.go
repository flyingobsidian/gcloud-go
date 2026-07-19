package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database gi-versions (#1268) ---

var oracleGiVersionsCmd = &cobra.Command{
	Use:   "gi-versions",
	Short: "Manage Oracle GI versions",
}

var (
	flagOdbGiLocation string
	flagOdbGiFormat   string
	flagOdbGiFilter   string
	flagOdbGiPageSize int64
)

var oracleGiListCmd = &cobra.Command{
	Use: "list", Short: "List Oracle GI versions in a location",
	Args: cobra.NoArgs, RunE: runOracleGiList,
}

func init() {
	all := []*cobra.Command{oracleGiListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbGiLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbGiFormat, "format", "", "Output format")
	}
	oracleGiListCmd.Flags().StringVar(&flagOdbGiFilter, "filter", "", "Server-side filter expression")
	oracleGiListCmd.Flags().Int64Var(&flagOdbGiPageSize, "page-size", 0, "Maximum results per page")

	oracleGiVersionsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleGiVersionsCmd)
}

func runOracleGiList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGiLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GiVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.GiVersions.List(parent).Context(ctx)
		if flagOdbGiFilter != "" {
			call = call.Filter(flagOdbGiFilter)
		}
		if flagOdbGiPageSize > 0 {
			call = call.PageSize(flagOdbGiPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing GI versions: %w", err)
		}
		all = append(all, resp.GiVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbGiFormat)
}
