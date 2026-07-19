package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database db-systems (#1284, dupe #1564) ---

var oracleDbSystemsCmd = &cobra.Command{
	Use:   "db-systems",
	Short: "Manage Oracle DB systems",
}

var (
	flagOdbDbsLocation   string
	flagOdbDbsFormat     string
	flagOdbDbsConfigFile string
	flagOdbDbsPageSize   int64
	flagOdbDbsFilter     string
)

var (
	oracleDbsCreateCmd = &cobra.Command{
		Use: "create DB_SYSTEM", Short: "Create an Oracle DB system",
		Args: cobra.ExactArgs(1), RunE: runOracleDbsCreate,
	}
	oracleDbsDeleteCmd = &cobra.Command{
		Use: "delete DB_SYSTEM", Short: "Delete an Oracle DB system",
		Args: cobra.ExactArgs(1), RunE: runOracleDbsDelete,
	}
	oracleDbsDescribeCmd = &cobra.Command{
		Use: "describe DB_SYSTEM", Short: "Describe an Oracle DB system",
		Args: cobra.ExactArgs(1), RunE: runOracleDbsDescribe,
	}
	oracleDbsListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle DB systems in a location",
		Args: cobra.NoArgs, RunE: runOracleDbsList,
	}
)

func init() {
	all := []*cobra.Command{
		oracleDbsCreateCmd, oracleDbsDeleteCmd,
		oracleDbsDescribeCmd, oracleDbsListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbDbsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbDbsFormat, "format", "", "Output format")
	}
	oracleDbsCreateCmd.Flags().StringVar(&flagOdbDbsConfigFile, "config-file", "", "YAML/JSON file with DB system body (required)")
	_ = oracleDbsCreateCmd.MarkFlagRequired("config-file")
	oracleDbsListCmd.Flags().Int64Var(&flagOdbDbsPageSize, "page-size", 0, "Maximum results per page")
	oracleDbsListCmd.Flags().StringVar(&flagOdbDbsFilter, "filter", "", "Server-side filter expression")

	oracleDbSystemsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleDbSystemsCmd)
}

func oracleDbsName(id string) (string, error) {
	return odbResource(flagOdbDbsLocation, "dbSystems", id)
}

func runOracleDbsCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbDbsLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.DbSystem{}
	if err := loadYAMLOrJSONInto(flagOdbDbsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DbSystems.Create(parent, body).DbSystemId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DB system: %w", err)
	}
	fmt.Printf("Create DB system [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbDbsFormat)
}

func runOracleDbsDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleDbsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DbSystems.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting DB system: %w", err)
	}
	fmt.Printf("Delete DB system [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbDbsFormat)
}

func runOracleDbsDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleDbsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DbSystems.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DB system: %w", err)
	}
	return emitFormatted(got, flagOdbDbsFormat)
}

func runOracleDbsList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbDbsLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.DbSystem
	pageToken := ""
	for {
		call := svc.Projects.Locations.DbSystems.List(parent).Context(ctx)
		if flagOdbDbsPageSize > 0 {
			call = call.PageSize(flagOdbDbsPageSize)
		}
		if flagOdbDbsFilter != "" {
			call = call.Filter(flagOdbDbsFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing DB systems: %w", err)
		}
		all = append(all, resp.DbSystems...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbDbsFormat)
}
