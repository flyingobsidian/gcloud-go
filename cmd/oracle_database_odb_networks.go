package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database odb-networks (#1276) ---

var oracleOdbNetworksCmd = &cobra.Command{
	Use:   "odb-networks",
	Short: "Manage Oracle ODB networks",
}

var (
	flagOdbNetLocation   string
	flagOdbNetFormat     string
	flagOdbNetConfigFile string
	flagOdbNetPageSize   int64
	flagOdbNetFilter     string
)

var (
	oracleOdbNetCreateCmd = &cobra.Command{
		Use: "create NETWORK", Short: "Create an ODB network",
		Args: cobra.ExactArgs(1), RunE: runOracleOdbNetCreate,
	}
	oracleOdbNetDeleteCmd = &cobra.Command{
		Use: "delete NETWORK", Short: "Delete an ODB network",
		Args: cobra.ExactArgs(1), RunE: runOracleOdbNetDelete,
	}
	oracleOdbNetDescribeCmd = &cobra.Command{
		Use: "describe NETWORK", Short: "Describe an ODB network",
		Args: cobra.ExactArgs(1), RunE: runOracleOdbNetDescribe,
	}
	oracleOdbNetListCmd = &cobra.Command{
		Use: "list", Short: "List ODB networks in a location",
		Args: cobra.NoArgs, RunE: runOracleOdbNetList,
	}
)

func init() {
	all := []*cobra.Command{
		oracleOdbNetCreateCmd, oracleOdbNetDeleteCmd,
		oracleOdbNetDescribeCmd, oracleOdbNetListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbNetLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbNetFormat, "format", "", "Output format")
	}
	oracleOdbNetCreateCmd.Flags().StringVar(&flagOdbNetConfigFile, "config-file", "", "YAML/JSON file with ODB network body (required)")
	_ = oracleOdbNetCreateCmd.MarkFlagRequired("config-file")
	oracleOdbNetListCmd.Flags().Int64Var(&flagOdbNetPageSize, "page-size", 0, "Maximum results per page")
	oracleOdbNetListCmd.Flags().StringVar(&flagOdbNetFilter, "filter", "", "Server-side filter expression")

	oracleOdbNetworksCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleOdbNetworksCmd)
}

func oracleOdbNetName(id string) (string, error) {
	return odbResource(flagOdbNetLocation, "odbNetworks", id)
}

func runOracleOdbNetCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbNetLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.OdbNetwork{}
	if err := loadYAMLOrJSONInto(flagOdbNetConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.OdbNetworks.Create(parent, body).OdbNetworkId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ODB network: %w", err)
	}
	fmt.Printf("Create ODB network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbNetFormat)
}

func runOracleOdbNetDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleOdbNetName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.OdbNetworks.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting ODB network: %w", err)
	}
	fmt.Printf("Delete ODB network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbNetFormat)
}

func runOracleOdbNetDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleOdbNetName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OdbNetworks.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ODB network: %w", err)
	}
	return emitFormatted(got, flagOdbNetFormat)
}

func runOracleOdbNetList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbNetLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.OdbNetwork
	pageToken := ""
	for {
		call := svc.Projects.Locations.OdbNetworks.List(parent).Context(ctx)
		if flagOdbNetPageSize > 0 {
			call = call.PageSize(flagOdbNetPageSize)
		}
		if flagOdbNetFilter != "" {
			call = call.Filter(flagOdbNetFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ODB networks: %w", err)
		}
		all = append(all, resp.OdbNetworks...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbNetFormat)
}
