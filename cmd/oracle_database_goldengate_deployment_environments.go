package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-deployment-environments (#1272) ---

var oracleGgDeployEnvsCmd = &cobra.Command{
	Use:   "goldengate-deployment-environments",
	Short: "List Oracle GoldenGate deployment environments",
}

var (
	flagOdbGgDeLocation string
	flagOdbGgDeFormat   string
	flagOdbGgDePageSize int64
)

var (
	oracleGgDeCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate deployment environments in a location",
		Args: cobra.NoArgs, RunE: runOracleGgDeList,
	}
)

func init() {
	all := []*cobra.Command{oracleGgDeCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbGgDeLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbGgDeFormat, "format", "", "Output format")
	}
	oracleGgDeCmd.Flags().Int64Var(&flagOdbGgDePageSize, "page-size", 0, "Maximum results per page")

	oracleGgDeployEnvsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleGgDeployEnvsCmd)
}

func runOracleGgDeList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgDeLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateDeploymentEnvironment
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateDeploymentEnvironments.List(parent).Context(ctx)
		if flagOdbGgDePageSize > 0 {
			call = call.PageSize(flagOdbGgDePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate deployment environments: %w", err)
		}
		all = append(all, resp.GoldengateDeploymentEnvironments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbGgDeFormat)
}
