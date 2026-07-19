package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database autonomous-database-character-sets (#1262) ---

var oracleAutonomousDatabaseCharacterSetsCmd = &cobra.Command{
	Use:   "autonomous-database-character-sets",
	Short: "Manage Autonomous Database character sets",
}

var (
	flagOdbAdCsLocation string
	flagOdbAdCsFormat   string
	flagOdbAdCsFilter   string
	flagOdbAdCsPageSize int64
)

var oracleAdCsListCmd = &cobra.Command{
	Use: "list", Short: "List Autonomous Database character sets in a location",
	Args: cobra.NoArgs, RunE: runOracleAdCsList,
}

func init() {
	all := []*cobra.Command{oracleAdCsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbAdCsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbAdCsFormat, "format", "", "Output format")
	}
	oracleAdCsListCmd.Flags().StringVar(&flagOdbAdCsFilter, "filter", "", "Server-side filter expression")
	oracleAdCsListCmd.Flags().Int64Var(&flagOdbAdCsPageSize, "page-size", 0, "Maximum results per page")

	oracleAutonomousDatabaseCharacterSetsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleAutonomousDatabaseCharacterSetsCmd)
}

func runOracleAdCsList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbAdCsLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.AutonomousDatabaseCharacterSet
	pageToken := ""
	for {
		call := svc.Projects.Locations.AutonomousDatabaseCharacterSets.List(parent).Context(ctx)
		if flagOdbAdCsFilter != "" {
			call = call.Filter(flagOdbAdCsFilter)
		}
		if flagOdbAdCsPageSize > 0 {
			call = call.PageSize(flagOdbAdCsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing autonomous database character sets: %w", err)
		}
		all = append(all, resp.AutonomousDatabaseCharacterSets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbAdCsFormat)
}
