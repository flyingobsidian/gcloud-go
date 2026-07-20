package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database autonomous-database-backups (#1261) ---
//
// The v1 client only exposes List on this service, so the subgroup is
// list-only. describe/create/delete are not available.

var oracleAutonomousDatabaseBackupsCmd = &cobra.Command{Use: "autonomous-database-backups", Short: "Manage Oracle Autonomous Database backups"}

var (
	flagOdbAdbBkLocation string
	flagOdbAdbBkFormat   string
	flagOdbAdbBkFilter   string
	flagOdbAdbBkPageSize int64
)

var oracleAdbBkListCmd = &cobra.Command{
	Use: "list", Short: "List Autonomous Database backups in a location",
	Args: cobra.NoArgs, RunE: runOracleAdbBkList,
}

func init() {
	oracleAdbBkListCmd.Flags().StringVar(&flagOdbAdbBkLocation, "location", "", "Location (required)")
	_ = oracleAdbBkListCmd.MarkFlagRequired("location")
	oracleAdbBkListCmd.Flags().StringVar(&flagOdbAdbBkFormat, "format", "", "Output format")
	oracleAdbBkListCmd.Flags().StringVar(&flagOdbAdbBkFilter, "filter", "", "Server-side filter expression")
	oracleAdbBkListCmd.Flags().Int64Var(&flagOdbAdbBkPageSize, "page-size", 0, "Maximum results per page")

	oracleAutonomousDatabaseBackupsCmd.AddCommand(oracleAdbBkListCmd)
	oracleDatabaseCmd.AddCommand(oracleAutonomousDatabaseBackupsCmd)
}

func runOracleAdbBkList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbAdbBkLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.AutonomousDatabaseBackup
	pageToken := ""
	for {
		call := svc.Projects.Locations.AutonomousDatabaseBackups.List(parent).Context(ctx)
		if flagOdbAdbBkFilter != "" {
			call = call.Filter(flagOdbAdbBkFilter)
		}
		if flagOdbAdbBkPageSize > 0 {
			call = call.PageSize(flagOdbAdbBkPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing autonomous database backups: %w", err)
		}
		all = append(all, resp.AutonomousDatabaseBackups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbAdbBkFormat)
}
