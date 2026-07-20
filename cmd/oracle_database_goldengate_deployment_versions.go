package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-deployment-versions (#1274) ---

var oracleGgDeployVersionsCmd = &cobra.Command{
	Use:   "goldengate-deployment-versions",
	Short: "List Oracle GoldenGate deployment versions",
}

var (
	flagOdbGgDvLocation string
	flagOdbGgDvFormat   string
	flagOdbGgDvPageSize int64
	flagOdbGgDvFilter   string
)

var (
	oracleGgDvListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate deployment versions in a location",
		Args: cobra.NoArgs, RunE: runOracleGgDvList,
	}
)

func init() {
	all := []*cobra.Command{oracleGgDvListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbGgDvLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbGgDvFormat, "format", "", "Output format")
	}
	oracleGgDvListCmd.Flags().Int64Var(&flagOdbGgDvPageSize, "page-size", 0, "Maximum results per page")
	oracleGgDvListCmd.Flags().StringVar(&flagOdbGgDvFilter, "filter", "", "Server-side filter expression")

	oracleGgDeployVersionsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleGgDeployVersionsCmd)
}

func runOracleGgDvList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgDvLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateDeploymentVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateDeploymentVersions.List(parent).Context(ctx)
		if flagOdbGgDvPageSize > 0 {
			call = call.PageSize(flagOdbGgDvPageSize)
		}
		if flagOdbGgDvFilter != "" {
			call = call.Filter(flagOdbGgDvFilter)
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
	return emitFormatted(all, flagOdbGgDvFormat)
}
