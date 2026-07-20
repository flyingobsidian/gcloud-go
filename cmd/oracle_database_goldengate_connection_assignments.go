package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database goldengate-connection-assignments (#1269) ---

var oracleGgConnAssignmentsCmd = &cobra.Command{
	Use:   "goldengate-connection-assignments",
	Short: "Manage Oracle GoldenGate connection assignments",
}

var (
	flagOdbGgCaLocation   string
	flagOdbGgCaFormat     string
	flagOdbGgCaConfigFile string
	flagOdbGgCaPageSize   int64
	flagOdbGgCaFilter     string
)

var (
	oracleGgCaCreateCmd = &cobra.Command{
		Use: "create ASSIGNMENT", Short: "Create a GoldenGate connection assignment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgCaCreate,
	}
	oracleGgCaDeleteCmd = &cobra.Command{
		Use: "delete ASSIGNMENT", Short: "Delete a GoldenGate connection assignment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgCaDelete,
	}
	oracleGgCaDescribeCmd = &cobra.Command{
		Use: "describe ASSIGNMENT", Short: "Describe a GoldenGate connection assignment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgCaDescribe,
	}
	oracleGgCaListCmd = &cobra.Command{
		Use: "list", Short: "List GoldenGate connection assignments in a location",
		Args: cobra.NoArgs, RunE: runOracleGgCaList,
	}
	oracleGgCaTestCmd = &cobra.Command{
		Use: "test ASSIGNMENT", Short: "Test a GoldenGate connection assignment",
		Args: cobra.ExactArgs(1), RunE: runOracleGgCaTest,
	}
)

func init() {
	all := []*cobra.Command{
		oracleGgCaCreateCmd, oracleGgCaDeleteCmd,
		oracleGgCaDescribeCmd, oracleGgCaListCmd, oracleGgCaTestCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbGgCaLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbGgCaFormat, "format", "", "Output format")
	}
	oracleGgCaCreateCmd.Flags().StringVar(&flagOdbGgCaConfigFile, "config-file", "", "YAML/JSON file with GoldengateConnectionAssignment body (required)")
	_ = oracleGgCaCreateCmd.MarkFlagRequired("config-file")
	oracleGgCaListCmd.Flags().Int64Var(&flagOdbGgCaPageSize, "page-size", 0, "Maximum results per page")
	oracleGgCaListCmd.Flags().StringVar(&flagOdbGgCaFilter, "filter", "", "Server-side filter expression")

	oracleGgConnAssignmentsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleGgConnAssignmentsCmd)
}

func oracleGgCaName(id string) (string, error) {
	return odbResource(flagOdbGgCaLocation, "goldengateConnectionAssignments", id)
}

func runOracleGgCaCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgCaLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.GoldengateConnectionAssignment{}
	if err := loadYAMLOrJSONInto(flagOdbGgCaConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateConnectionAssignments.Create(parent, body).GoldengateConnectionAssignmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating goldengate connection assignment: %w", err)
	}
	fmt.Printf("Create goldengate connection assignment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbGgCaFormat)
}

func runOracleGgCaDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleGgCaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoldengateConnectionAssignments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting goldengate connection assignment: %w", err)
	}
	fmt.Printf("Delete goldengate connection assignment [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbGgCaFormat)
}

func runOracleGgCaDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleGgCaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GoldengateConnectionAssignments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing goldengate connection assignment: %w", err)
	}
	return emitFormatted(got, flagOdbGgCaFormat)
}

func runOracleGgCaList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbGgCaLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.GoldengateConnectionAssignment
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoldengateConnectionAssignments.List(parent).Context(ctx)
		if flagOdbGgCaPageSize > 0 {
			call = call.PageSize(flagOdbGgCaPageSize)
		}
		if flagOdbGgCaFilter != "" {
			call = call.Filter(flagOdbGgCaFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing goldengate connection assignments: %w", err)
		}
		all = append(all, resp.GoldengateConnectionAssignments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbGgCaFormat)
}

func runOracleGgCaTest(cmd *cobra.Command, args []string) error {
	name, err := oracleGgCaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.GoldengateConnectionAssignments.Test(name, &oracledatabase.TestGoldengateConnectionAssignmentRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("testing goldengate connection assignment: %w", err)
	}
	fmt.Printf("Test goldengate connection assignment [%s] completed.\n", args[0])
	return emitFormatted(resp, flagOdbGgCaFormat)
}
