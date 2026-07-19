package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database databases (#1281) ---

var oracleDatabasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "Manage Oracle databases",
}

var (
	flagOdbDbLocation string
	flagOdbDbFormat   string
	flagOdbDbPageSize int64
	flagOdbDbFilter   string
)

var (
	oracleDbDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Describe an Oracle database",
		Args: cobra.ExactArgs(1), RunE: runOracleDbDescribe,
	}
	oracleDbListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle databases in a location",
		Args: cobra.NoArgs, RunE: runOracleDbList,
	}
)

func init() {
	all := []*cobra.Command{oracleDbDescribeCmd, oracleDbListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbDbLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbDbFormat, "format", "", "Output format")
	}
	oracleDbListCmd.Flags().Int64Var(&flagOdbDbPageSize, "page-size", 0, "Maximum results per page")
	oracleDbListCmd.Flags().StringVar(&flagOdbDbFilter, "filter", "", "Server-side filter expression")

	oracleDatabasesCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleDatabasesCmd)
}

func oracleDbName(id string) (string, error) {
	return odbResource(flagOdbDbLocation, "databases", id)
}

func runOracleDbDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleDbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Databases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing database: %w", err)
	}
	return emitFormatted(got, flagOdbDbFormat)
}

func runOracleDbList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbDbLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.Database
	pageToken := ""
	for {
		call := svc.Projects.Locations.Databases.List(parent).Context(ctx)
		if flagOdbDbPageSize > 0 {
			call = call.PageSize(flagOdbDbPageSize)
		}
		if flagOdbDbFilter != "" {
			call = call.Filter(flagOdbDbFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing databases: %w", err)
		}
		all = append(all, resp.Databases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbDbFormat)
}
