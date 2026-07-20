package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-deployments (#1275) ---

var oracleGgDeploymentsCmd = &cobra.Command{
	Use:   "goldengate-deployments",
	Short: "Manage Oracle GoldenGate deployments",
}

var (
	flagOdbGgDpLocation   string
	flagOdbGgDpFormat     string
	flagOdbGgDpConfigFile string
	flagOdbGgDpPageSize   int64
	flagOdbGgDpFilter     string
)

var (
	oracleGgDpCreateCmd = &cobra.Command{
		Use: "create DEPLOYMENT", Short: "Create a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgDpCreate,
	}
	oracleGgDpDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgDpDelete,
	}
	oracleGgDpDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgDpDescribe,
	}
	oracleGgDpListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate deployments in a location",
		Args: cobra.NoArgs, RunE: runOracleGgDpList,
	}
	oracleGgDpStartCmd = &cobra.Command{
		Use: "start DEPLOYMENT", Short: "Start a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgDpStart,
	}
	oracleGgDpStopCmd = &cobra.Command{
		Use: "stop DEPLOYMENT", Short: "Stop a GoldenGate deployment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgDpStop,
	}
)

func init() {
	all := []*cobra.Command{
		oracleGgDpCreateCmd, oracleGgDpDeleteCmd,
		oracleGgDpDescribeCmd, oracleGgDpListCmd,
		oracleGgDpStartCmd, oracleGgDpStopCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbGgDpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbGgDpFormat, "format", "", "Output format")
	}
	oracleGgDpCreateCmd.Flags().StringVar(&flagOdbGgDpConfigFile, "config-file", "", "YAML/JSON file with GoldengateDeployment body (required)")
	_ = oracleGgDpCreateCmd.MarkFlagRequired("config-file")
	oracleGgDpListCmd.Flags().Int64Var(&flagOdbGgDpPageSize, "page-size", 0, "Maximum results per page")
	oracleGgDpListCmd.Flags().StringVar(&flagOdbGgDpFilter, "filter", "", "Server-side filter expression")

	oracleGgDeploymentsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleGgDeploymentsCmd)
}

func oracleGgDpName(id string) (string, error) {
	return odbResource(flagOdbGgDpLocation, "goldengateDeployments", id)
}

func runOracleGgDpCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgDpLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.GoldengateDeployment{}
	if err := loadYAMLOrJSONInto(flagOdbGgDpConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateDeployments.Create(parent, body).GoldengateDeploymentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating goldengate deployment: %w", err)
	}
	fmt.Printf("Create goldengate deployment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbGgDpFormat)
}

func runOracleGgDpDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleGgDpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateDeployments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting goldengate deployment: %w", err)
	}
	fmt.Printf("Delete goldengate deployment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbGgDpFormat)
}

func runOracleGgDpDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleGgDpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GoldengateDeployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing goldengate deployment: %w", err)
	}
	return emitFormatted(got, flagOdbGgDpFormat)
}

func runOracleGgDpList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgDpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateDeployment
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateDeployments.List(parent).Context(ctx)
		if flagOdbGgDpPageSize > 0 {
			call = call.PageSize(flagOdbGgDpPageSize)
		}
		if flagOdbGgDpFilter != "" {
			call = call.Filter(flagOdbGgDpFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate deployments: %w", err)
		}
		all = append(all, resp.GoldengateDeployments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbGgDpFormat)
}

func runOracleGgDpStart(cmd *cobra.Command, args []string) error {
	name, err := oracleGgDpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateDeployments.Start(name, &oracledatabase.StartGoldengateDeploymentRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting goldengate deployment: %w", err)
	}
	fmt.Printf("Start goldengate deployment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbGgDpFormat)
}

func runOracleGgDpStop(cmd *cobra.Command, args []string) error {
	name, err := oracleGgDpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateDeployments.Stop(name, &oracledatabase.StopGoldengateDeploymentRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping goldengate deployment: %w", err)
	}
	fmt.Printf("Stop goldengate deployment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbGgDpFormat)
}
