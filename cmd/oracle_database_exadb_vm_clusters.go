package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database exadb-vm-clusters (#1286, dupe #1566) ---

var oracleExadbVmClustersCmd = &cobra.Command{
	Use:   "exadb-vm-clusters",
	Short: "Manage Oracle ExaDB VM clusters",
}

var (
	flagOdbExaVmLocation   string
	flagOdbExaVmFormat     string
	flagOdbExaVmConfigFile string
	flagOdbExaVmUpdateMask string
	flagOdbExaVmPageSize   int64
	flagOdbExaVmFilter     string
)

var (
	oracleExaVmCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create an ExaDB VM cluster",
		Args: cobra.ExactArgs(1), RunE: runOracleExaVmCreate,
	}
	oracleExaVmDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete an ExaDB VM cluster",
		Args: cobra.ExactArgs(1), RunE: runOracleExaVmDelete,
	}
	oracleExaVmDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe an ExaDB VM cluster",
		Args: cobra.ExactArgs(1), RunE: runOracleExaVmDescribe,
	}
	oracleExaVmListCmd = &cobra.Command{
		Use: "list", Short: "List ExaDB VM clusters in a location",
		Args: cobra.NoArgs, RunE: runOracleExaVmList,
	}
	oracleExaVmUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update an ExaDB VM cluster",
		Args: cobra.ExactArgs(1), RunE: runOracleExaVmUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		oracleExaVmCreateCmd, oracleExaVmDeleteCmd,
		oracleExaVmDescribeCmd, oracleExaVmListCmd, oracleExaVmUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbExaVmLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbExaVmFormat, "format", "", "Output format")
	}
	oracleExaVmCreateCmd.Flags().StringVar(&flagOdbExaVmConfigFile, "config-file", "", "YAML/JSON file with ExaDB VM cluster body (required)")
	_ = oracleExaVmCreateCmd.MarkFlagRequired("config-file")
	oracleExaVmUpdateCmd.Flags().StringVar(&flagOdbExaVmConfigFile, "config-file", "", "YAML/JSON file with ExaDB VM cluster body (required)")
	_ = oracleExaVmUpdateCmd.MarkFlagRequired("config-file")
	oracleExaVmUpdateCmd.Flags().StringVar(&flagOdbExaVmUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	oracleExaVmListCmd.Flags().Int64Var(&flagOdbExaVmPageSize, "page-size", 0, "Maximum results per page")
	oracleExaVmListCmd.Flags().StringVar(&flagOdbExaVmFilter, "filter", "", "Server-side filter expression")

	oracleExadbVmClustersCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleExadbVmClustersCmd)
}

func oracleExaVmName(id string) (string, error) {
	return odbResource(flagOdbExaVmLocation, "exadbVmClusters", id)
}

func runOracleExaVmCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbExaVmLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.ExadbVmCluster{}
	if err := loadYAMLOrJSONInto(flagOdbExaVmConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExadbVmClusters.Create(parent, body).ExadbVmClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ExaDB VM cluster: %w", err)
	}
	fmt.Printf("Create ExaDB VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbExaVmFormat)
}

func runOracleExaVmDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleExaVmName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExadbVmClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting ExaDB VM cluster: %w", err)
	}
	fmt.Printf("Delete ExaDB VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbExaVmFormat)
}

func runOracleExaVmDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleExaVmName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExadbVmClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ExaDB VM cluster: %w", err)
	}
	return emitFormatted(got, flagOdbExaVmFormat)
}

func runOracleExaVmList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbExaVmLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.ExadbVmCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.ExadbVmClusters.List(parent).Context(ctx)
		if flagOdbExaVmPageSize > 0 {
			call = call.PageSize(flagOdbExaVmPageSize)
		}
		if flagOdbExaVmFilter != "" {
			call = call.Filter(flagOdbExaVmFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ExaDB VM clusters: %w", err)
		}
		all = append(all, resp.ExadbVmClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbExaVmFormat)
}

func runOracleExaVmUpdate(cmd *cobra.Command, args []string) error {
	name, err := oracleExaVmName(args[0])
	if err != nil {
		return err
	}
	body := &oracledatabase.ExadbVmCluster{}
	if err := loadYAMLOrJSONInto(flagOdbExaVmConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagOdbExaVmUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ExadbVmClusters.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating ExaDB VM cluster: %w", err)
	}
	fmt.Printf("Update ExaDB VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbExaVmFormat)
}
