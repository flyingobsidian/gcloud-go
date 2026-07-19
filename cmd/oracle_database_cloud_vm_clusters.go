package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database cloud-vm-clusters (#1265) ---

var oracleCloudVmClustersCmd = &cobra.Command{
	Use:   "cloud-vm-clusters",
	Short: "Manage Oracle Cloud VM clusters",
}

var (
	flagOdbVmClLocation   string
	flagOdbVmClFormat     string
	flagOdbVmClConfigFile string
	flagOdbVmClPageSize   int64
	flagOdbVmClFilter     string
)

var (
	oracleVmClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a cloud VM cluster",
		Args: cobra.ExactArgs(1), RunE: runOracleVmClCreate,
	}
	oracleVmClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a cloud VM cluster",
		Args: cobra.ExactArgs(1), RunE: runOracleVmClDelete,
	}
	oracleVmClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a cloud VM cluster",
		Args: cobra.ExactArgs(1), RunE: runOracleVmClDescribe,
	}
	oracleVmClListCmd = &cobra.Command{
		Use: "list", Short: "List cloud VM clusters in a location",
		Args: cobra.NoArgs, RunE: runOracleVmClList,
	}
)

func init() {
	all := []*cobra.Command{
		oracleVmClCreateCmd, oracleVmClDeleteCmd,
		oracleVmClDescribeCmd, oracleVmClListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbVmClLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbVmClFormat, "format", "", "Output format")
	}
	oracleVmClCreateCmd.Flags().StringVar(&flagOdbVmClConfigFile, "config-file", "", "YAML/JSON file with cloud VM cluster body (required)")
	_ = oracleVmClCreateCmd.MarkFlagRequired("config-file")
	oracleVmClListCmd.Flags().Int64Var(&flagOdbVmClPageSize, "page-size", 0, "Maximum results per page")
	oracleVmClListCmd.Flags().StringVar(&flagOdbVmClFilter, "filter", "", "Server-side filter expression")

	oracleCloudVmClustersCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleCloudVmClustersCmd)
}

func oracleVmClName(id string) (string, error) {
	return odbResource(flagOdbVmClLocation, "cloudVmClusters", id)
}

func runOracleVmClCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbVmClLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.CloudVmCluster{}
	if err := loadYAMLOrJSONInto(flagOdbVmClConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CloudVmClusters.Create(parent, body).CloudVmClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cloud VM cluster: %w", err)
	}
	fmt.Printf("Create cloud VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbVmClFormat)
}

func runOracleVmClDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleVmClName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CloudVmClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cloud VM cluster: %w", err)
	}
	fmt.Printf("Delete cloud VM cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbVmClFormat)
}

func runOracleVmClDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleVmClName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CloudVmClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cloud VM cluster: %w", err)
	}
	return emitFormatted(got, flagOdbVmClFormat)
}

func runOracleVmClList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbVmClLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.CloudVmCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.CloudVmClusters.List(parent).Context(ctx)
		if flagOdbVmClPageSize > 0 {
			call = call.PageSize(flagOdbVmClPageSize)
		}
		if flagOdbVmClFilter != "" {
			call = call.Filter(flagOdbVmClFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing cloud VM clusters: %w", err)
		}
		all = append(all, resp.CloudVmClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbVmClFormat)
}
