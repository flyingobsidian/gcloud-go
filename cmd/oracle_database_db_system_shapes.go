package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database db-system-shapes (#1266) ---

var oracleDbSystemShapesCmd = &cobra.Command{
	Use:   "db-system-shapes",
	Short: "Manage Oracle DB system shapes",
}

var (
	flagOdbShapeLocation string
	flagOdbShapeFormat   string
	flagOdbShapeFilter   string
	flagOdbShapePageSize int64
)

var oracleShapeListCmd = &cobra.Command{
	Use: "list", Short: "List Oracle DB system shapes in a location",
	Args: cobra.NoArgs, RunE: runOracleShapeList,
}

func init() {
	all := []*cobra.Command{oracleShapeListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbShapeLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbShapeFormat, "format", "", "Output format")
	}
	oracleShapeListCmd.Flags().StringVar(&flagOdbShapeFilter, "filter", "", "Server-side filter expression")
	oracleShapeListCmd.Flags().Int64Var(&flagOdbShapePageSize, "page-size", 0, "Maximum results per page")

	oracleDbSystemShapesCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleDbSystemShapesCmd)
}

func runOracleShapeList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbShapeLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DbSystemShape
	pageToken := ""
	for {
		call := svc.Projects.Locations.DbSystemShapes.List(parent).Context(ctx)
		if flagOdbShapeFilter != "" {
			call = call.Filter(flagOdbShapeFilter)
		}
		if flagOdbShapePageSize > 0 {
			call = call.PageSize(flagOdbShapePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing db system shapes: %w", err)
		}
		all = append(all, resp.DbSystemShapes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbShapeFormat)
}
