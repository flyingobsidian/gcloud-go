package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database entitlements (#1267) ---

var oracleEntitlementsCmd = &cobra.Command{
	Use:   "entitlements",
	Short: "Manage Oracle Database entitlements",
}

var (
	flagOdbEntLocation string
	flagOdbEntFormat   string
	flagOdbEntPageSize int64
)

var oracleEntListCmd = &cobra.Command{
	Use: "list", Short: "List Oracle Database entitlements in a location",
	Args: cobra.NoArgs, RunE: runOracleEntList,
}

func init() {
	all := []*cobra.Command{oracleEntListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbEntLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbEntFormat, "format", "", "Output format")
	}
	oracleEntListCmd.Flags().Int64Var(&flagOdbEntPageSize, "page-size", 0, "Maximum results per page")

	oracleEntitlementsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleEntitlementsCmd)
}

func runOracleEntList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbEntLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.Entitlement
	pageToken := ""
	for {
		call := svc.Projects.Locations.Entitlements.List(parent).Context(ctx)
		if flagOdbEntPageSize > 0 {
			call = call.PageSize(flagOdbEntPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing entitlements: %w", err)
		}
		all = append(all, resp.Entitlements...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbEntFormat)
}
