package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database autonomous-db-versions (#1279) ---

var oracleAutonomousDbVersionsCmd = &cobra.Command{
	Use:   "autonomous-db-versions",
	Short: "Manage Autonomous DB versions",
}

var (
	flagOdbAdvLocation string
	flagOdbAdvFormat   string
	flagOdbAdvPageSize int64
)

var oracleAdvListCmd = &cobra.Command{
	Use: "list", Short: "List Autonomous DB versions in a location",
	Args: cobra.NoArgs, RunE: runOracleAdvList,
}

func init() {
	all := []*cobra.Command{oracleAdvListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbAdvLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbAdvFormat, "format", "", "Output format")
	}
	oracleAdvListCmd.Flags().Int64Var(&flagOdbAdvPageSize, "page-size", 0, "Maximum results per page")

	oracleAutonomousDbVersionsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleAutonomousDbVersionsCmd)
}

func runOracleAdvList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbAdvLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.AutonomousDbVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.AutonomousDbVersions.List(parent).Context(ctx)
		if flagOdbAdvPageSize > 0 {
			call = call.PageSize(flagOdbAdvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing autonomous db versions: %w", err)
		}
		all = append(all, resp.AutonomousDbVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbAdvFormat)
}
