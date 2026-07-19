package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database database-character-sets (#1280) ---

var oracleDatabaseCharacterSetsCmd = &cobra.Command{
	Use:   "database-character-sets",
	Short: "Manage Oracle Database character sets",
}

var (
	flagOdbDbCsLocation string
	flagOdbDbCsFormat   string
	flagOdbDbCsFilter   string
	flagOdbDbCsPageSize int64
)

var oracleDbCsListCmd = &cobra.Command{
	Use: "list", Short: "List Oracle Database character sets in a location",
	Args: cobra.NoArgs, RunE: runOracleDbCsList,
}

func init() {
	all := []*cobra.Command{oracleDbCsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbDbCsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbDbCsFormat, "format", "", "Output format")
	}
	oracleDbCsListCmd.Flags().StringVar(&flagOdbDbCsFilter, "filter", "", "Server-side filter expression")
	oracleDbCsListCmd.Flags().Int64Var(&flagOdbDbCsPageSize, "page-size", 0, "Maximum results per page")

	oracleDatabaseCharacterSetsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleDatabaseCharacterSetsCmd)
}

func runOracleDbCsList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbDbCsLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DatabaseCharacterSet
	pageToken := ""
	for {
		call := svc.Projects.Locations.DatabaseCharacterSets.List(parent).Context(ctx)
		if flagOdbDbCsFilter != "" {
			call = call.Filter(flagOdbDbCsFilter)
		}
		if flagOdbDbCsPageSize > 0 {
			call = call.PageSize(flagOdbDbCsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing database character sets: %w", err)
		}
		all = append(all, resp.DatabaseCharacterSets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbDbCsFormat)
}
