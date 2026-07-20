package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-deployment-types (#1273) ---

var oracleGgDeployTypesCmd = &cobra.Command{
	Use:   "goldengate-deployment-types",
	Short: "List Oracle GoldenGate deployment types",
}

var (
	flagOdbGgDtLocation string
	flagOdbGgDtFormat   string
	flagOdbGgDtPageSize int64
	flagOdbGgDtFilter   string
)

var (
	oracleGgDtListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate deployment types in a location",
		Args: cobra.NoArgs, RunE: runOracleGgDtList,
	}
)

func init() {
	all := []*cobra.Command{oracleGgDtListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbGgDtLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbGgDtFormat, "format", "", "Output format")
	}
	oracleGgDtListCmd.Flags().Int64Var(&flagOdbGgDtPageSize, "page-size", 0, "Maximum results per page")
	oracleGgDtListCmd.Flags().StringVar(&flagOdbGgDtFilter, "filter", "", "Server-side filter expression")

	oracleGgDeployTypesCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleGgDeployTypesCmd)
}

func runOracleGgDtList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgDtLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateDeploymentType
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateDeploymentTypes.List(parent).Context(ctx)
		if flagOdbGgDtPageSize > 0 {
			call = call.PageSize(flagOdbGgDtPageSize)
		}
		if flagOdbGgDtFilter != "" {
			call = call.Filter(flagOdbGgDtFilter)
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
	return emitFormatted(all, flagOdbGgDtFormat)
}
