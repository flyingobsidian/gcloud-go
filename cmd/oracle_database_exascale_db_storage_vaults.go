package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database exascale-db-storage-vaults (#1287, dupe #1567) ---

var oracleExascaleDbStorageVaultsCmd = &cobra.Command{
	Use:   "exascale-db-storage-vaults",
	Short: "Manage Oracle Exascale DB storage vaults",
}

var (
	flagOdbExaStLocation   string
	flagOdbExaStFormat     string
	flagOdbExaStConfigFile string
	flagOdbExaStPageSize   int64
	flagOdbExaStFilter     string
)

var (
	oracleExaStCreateCmd = &cobra.Command{
		Use: "create VAULT", Short: "Create an Exascale DB storage vault",
		Args: cobra.ExactArgs(1), RunE: runOracleExaStCreate,
	}
	oracleExaStDeleteCmd = &cobra.Command{
		Use: "delete VAULT", Short: "Delete an Exascale DB storage vault",
		Args: cobra.ExactArgs(1), RunE: runOracleExaStDelete,
	}
	oracleExaStDescribeCmd = &cobra.Command{
		Use: "describe VAULT", Short: "Describe an Exascale DB storage vault",
		Args: cobra.ExactArgs(1), RunE: runOracleExaStDescribe,
	}
	oracleExaStListCmd = &cobra.Command{
		Use: "list", Short: "List Exascale DB storage vaults in a location",
		Args: cobra.NoArgs, RunE: runOracleExaStList,
	}
)

func init() {
	all := []*cobra.Command{
		oracleExaStCreateCmd, oracleExaStDeleteCmd,
		oracleExaStDescribeCmd, oracleExaStListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbExaStLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbExaStFormat, "format", "", "Output format")
	}
	oracleExaStCreateCmd.Flags().StringVar(&flagOdbExaStConfigFile, "config-file", "", "YAML/JSON file with Exascale DB storage vault body (required)")
	_ = oracleExaStCreateCmd.MarkFlagRequired("config-file")
	oracleExaStListCmd.Flags().Int64Var(&flagOdbExaStPageSize, "page-size", 0, "Maximum results per page")
	oracleExaStListCmd.Flags().StringVar(&flagOdbExaStFilter, "filter", "", "Server-side filter expression")

	oracleExascaleDbStorageVaultsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleExascaleDbStorageVaultsCmd)
}

func oracleExaStName(id string) (string, error) {
	return odbResource(flagOdbExaStLocation, "exascaleDbStorageVaults", id)
}

func runOracleExaStCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbExaStLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.ExascaleDbStorageVault{}
	if err := loadYAMLOrJSONInto(flagOdbExaStConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExascaleDbStorageVaults.Create(parent, body).ExascaleDbStorageVaultId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating Exascale DB storage vault: %w", err)
	}
	fmt.Printf("Create Exascale DB storage vault [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbExaStFormat)
}

func runOracleExaStDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleExaStName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExascaleDbStorageVaults.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting Exascale DB storage vault: %w", err)
	}
	fmt.Printf("Delete Exascale DB storage vault [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbExaStFormat)
}

func runOracleExaStDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleExaStName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExascaleDbStorageVaults.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Exascale DB storage vault: %w", err)
	}
	return emitFormatted(got, flagOdbExaStFormat)
}

func runOracleExaStList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbExaStLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.ExascaleDbStorageVault
	pageToken := ""
	for {
		call := svc.Projects.Locations.ExascaleDbStorageVaults.List(parent).Context(ctx)
		if flagOdbExaStPageSize > 0 {
			call = call.PageSize(flagOdbExaStPageSize)
		}
		if flagOdbExaStFilter != "" {
			call = call.Filter(flagOdbExaStFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing Exascale DB storage vaults: %w", err)
		}
		all = append(all, resp.ExascaleDbStorageVaults...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbExaStFormat)
}
