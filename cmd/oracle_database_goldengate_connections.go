package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-connections (#1271) ---

var oracleGgConnectionsCmd = &cobra.Command{
	Use:   "goldengate-connections",
	Short: "Manage Oracle GoldenGate connections",
}

var (
	flagOdbGgCnLocation   string
	flagOdbGgCnFormat     string
	flagOdbGgCnConfigFile string
	flagOdbGgCnPageSize   int64
	flagOdbGgCnFilter     string
)

var (
	oracleGgCnCreateCmd = &cobra.Command{
		Use: "create CONNECTION", Short: "Create a GoldenGate connection",
		Args: cobra.ExactArgs(1), RunE: runOracleGgCnCreate,
	}
	oracleGgCnDeleteCmd = &cobra.Command{
		Use: "delete CONNECTION", Short: "Delete a GoldenGate connection",
		Args: cobra.ExactArgs(1), RunE: runOracleGgCnDelete,
	}
	oracleGgCnDescribeCmd = &cobra.Command{
		Use: "describe CONNECTION", Short: "Describe a GoldenGate connection",
		Args: cobra.ExactArgs(1), RunE: runOracleGgCnDescribe,
	}
	oracleGgCnListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate connections in a location",
		Args: cobra.NoArgs, RunE: runOracleGgCnList,
	}
)

func init() {
	all := []*cobra.Command{
		oracleGgCnCreateCmd, oracleGgCnDeleteCmd,
		oracleGgCnDescribeCmd, oracleGgCnListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbGgCnLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbGgCnFormat, "format", "", "Output format")
	}
	oracleGgCnCreateCmd.Flags().StringVar(&flagOdbGgCnConfigFile, "config-file", "", "YAML/JSON file with GoldengateConnection body (required)")
	_ = oracleGgCnCreateCmd.MarkFlagRequired("config-file")
	oracleGgCnListCmd.Flags().Int64Var(&flagOdbGgCnPageSize, "page-size", 0, "Maximum results per page")
	oracleGgCnListCmd.Flags().StringVar(&flagOdbGgCnFilter, "filter", "", "Server-side filter expression")

	oracleGgConnectionsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleGgConnectionsCmd)
}

func oracleGgCnName(id string) (string, error) {
	return odbResource(flagOdbGgCnLocation, "goldengateConnections", id)
}

func runOracleGgCnCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgCnLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.GoldengateConnection{}
	if err := loadYAMLOrJSONInto(flagOdbGgCnConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateConnections.Create(parent, body).GoldengateConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating goldengate connection: %w", err)
	}
	fmt.Printf("Create goldengate connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbGgCnFormat)
}

func runOracleGgCnDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleGgCnName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateConnections.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting goldengate connection: %w", err)
	}
	fmt.Printf("Delete goldengate connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbGgCnFormat)
}

func runOracleGgCnDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleGgCnName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GoldengateConnections.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing goldengate connection: %w", err)
	}
	return emitFormatted(got, flagOdbGgCnFormat)
}

func runOracleGgCnList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgCnLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateConnection
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateConnections.List(parent).Context(ctx)
		if flagOdbGgCnPageSize > 0 {
			call = call.PageSize(flagOdbGgCnPageSize)
		}
		if flagOdbGgCnFilter != "" {
			call = call.Filter(flagOdbGgCnFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate connections: %w", err)
		}
		all = append(all, resp.GoldengateConnections...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbGgCnFormat)
}
