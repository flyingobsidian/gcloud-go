package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database pluggable-databases (#1278) ---

var oraclePluggableDatabasesCmd = &cobra.Command{
	Use:   "pluggable-databases",
	Short: "Manage Oracle pluggable databases",
}

var (
	flagOdbPdbLocation string
	flagOdbPdbFormat   string
	flagOdbPdbPageSize int64
	flagOdbPdbFilter   string
)

var (
	oraclePdbDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Describe a pluggable database",
		Args: cobra.ExactArgs(1), RunE: runOraclePdbDescribe,
	}
	oraclePdbListCmd = &cobra.Command{
		Use: "list", Short: "List pluggable databases in a location",
		Args: cobra.NoArgs, RunE: runOraclePdbList,
	}
)

func init() {
	all := []*cobra.Command{oraclePdbDescribeCmd, oraclePdbListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbPdbLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbPdbFormat, "format", "", "Output format")
	}
	oraclePdbListCmd.Flags().Int64Var(&flagOdbPdbPageSize, "page-size", 0, "Maximum results per page")
	oraclePdbListCmd.Flags().StringVar(&flagOdbPdbFilter, "filter", "", "Server-side filter expression")

	oraclePluggableDatabasesCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oraclePluggableDatabasesCmd)
}

func oraclePdbName(id string) (string, error) {
	return odbResource(flagOdbPdbLocation, "pluggableDatabases", id)
}

func runOraclePdbDescribe(cmd *cobra.Command, args []string) error {
	name, err := oraclePdbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PluggableDatabases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing pluggable database: %w", err)
	}
	return emitFormatted(got, flagOdbPdbFormat)
}

func runOraclePdbList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbPdbLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.PluggableDatabase
	pageToken := ""
	for {
		call := svc.Projects.Locations.PluggableDatabases.List(parent).Context(ctx)
		if flagOdbPdbPageSize > 0 {
			call = call.PageSize(flagOdbPdbPageSize)
		}
		if flagOdbPdbFilter != "" {
			call = call.Filter(flagOdbPdbFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing pluggable databases: %w", err)
		}
		all = append(all, resp.PluggableDatabases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbPdbFormat)
}
