package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-connection-types (#1270) ---

var oracleGgConnTypesCmd = &cobra.Command{
	Use:   "goldengate-connection-types",
	Short: "List Oracle GoldenGate connection types",
}

var (
	flagOdbGgCtLocation string
	flagOdbGgCtFormat   string
	flagOdbGgCtPageSize int64
	flagOdbGgCtFilter   string
)

var (
	oracleGgCtListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate connection types in a location",
		Args: cobra.NoArgs, RunE: runOracleGgCtList,
	}
)

func init() {
	all := []*cobra.Command{oracleGgCtListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbGgCtLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbGgCtFormat, "format", "", "Output format")
	}
	oracleGgCtListCmd.Flags().Int64Var(&flagOdbGgCtPageSize, "page-size", 0, "Maximum results per page")
	oracleGgCtListCmd.Flags().StringVar(&flagOdbGgCtFilter, "filter", "", "Server-side filter expression")

	oracleGgConnTypesCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleGgConnTypesCmd)
}

func runOracleGgCtList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgCtLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateConnectionType
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateConnectionTypes.List(parent).Context(ctx)
		if flagOdbGgCtPageSize > 0 {
			call = call.PageSize(flagOdbGgCtPageSize)
		}
		if flagOdbGgCtFilter != "" {
			call = call.Filter(flagOdbGgCtFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate connection types: %w", err)
		}
		all = append(all, resp.GoldengateConnectionTypes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbGgCtFormat)
}
