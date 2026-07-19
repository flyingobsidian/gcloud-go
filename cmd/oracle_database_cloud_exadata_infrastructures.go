package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database cloud-exadata-infrastructures (#1264) ---

var oracleCloudExadataInfrastructuresCmd = &cobra.Command{
	Use:   "cloud-exadata-infrastructures",
	Short: "Manage Oracle Cloud Exadata infrastructures",
}

var (
	flagOdbExInfLocation   string
	flagOdbExInfFormat     string
	flagOdbExInfConfigFile string
	flagOdbExInfPageSize   int64
	flagOdbExInfFilter     string
)

var (
	oracleExInfCreateCmd = &cobra.Command{
		Use: "create INFRASTRUCTURE", Short: "Create a cloud Exadata infrastructure",
		Args: cobra.ExactArgs(1), RunE: runOracleExInfCreate,
	}
	oracleExInfDeleteCmd = &cobra.Command{
		Use: "delete INFRASTRUCTURE", Short: "Delete a cloud Exadata infrastructure",
		Args: cobra.ExactArgs(1), RunE: runOracleExInfDelete,
	}
	oracleExInfDescribeCmd = &cobra.Command{
		Use: "describe INFRASTRUCTURE", Short: "Describe a cloud Exadata infrastructure",
		Args: cobra.ExactArgs(1), RunE: runOracleExInfDescribe,
	}
	oracleExInfListCmd = &cobra.Command{
		Use: "list", Short: "List cloud Exadata infrastructures in a location",
		Args: cobra.NoArgs, RunE: runOracleExInfList,
	}
)

func init() {
	all := []*cobra.Command{
		oracleExInfCreateCmd, oracleExInfDeleteCmd,
		oracleExInfDescribeCmd, oracleExInfListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbExInfLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbExInfFormat, "format", "", "Output format")
	}
	oracleExInfCreateCmd.Flags().StringVar(&flagOdbExInfConfigFile, "config-file", "", "YAML/JSON file with cloud Exadata infrastructure body (required)")
	_ = oracleExInfCreateCmd.MarkFlagRequired("config-file")
	oracleExInfListCmd.Flags().Int64Var(&flagOdbExInfPageSize, "page-size", 0, "Maximum results per page")
	oracleExInfListCmd.Flags().StringVar(&flagOdbExInfFilter, "filter", "", "Server-side filter expression")

	oracleCloudExadataInfrastructuresCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleCloudExadataInfrastructuresCmd)
}

func oracleExInfName(id string) (string, error) {
	return odbResource(flagOdbExInfLocation, "cloudExadataInfrastructures", id)
}

func runOracleExInfCreate(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbExInfLocation)
	if err != nil {
		return err
	}
	body := &oracledatabase.CloudExadataInfrastructure{}
	if err := loadYAMLOrJSONInto(flagOdbExInfConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CloudExadataInfrastructures.Create(parent, body).CloudExadataInfrastructureId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cloud Exadata infrastructure: %w", err)
	}
	fmt.Printf("Create cloud Exadata infrastructure [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbExInfFormat)
}

func runOracleExInfDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleExInfName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CloudExadataInfrastructures.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cloud Exadata infrastructure: %w", err)
	}
	fmt.Printf("Delete cloud Exadata infrastructure [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagOdbExInfFormat)
}

func runOracleExInfDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleExInfName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CloudExadataInfrastructures.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cloud Exadata infrastructure: %w", err)
	}
	return emitFormatted(got, flagOdbExInfFormat)
}

func runOracleExInfList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbExInfLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.CloudExadataInfrastructure
	pageToken := ""
	for {
		call := svc.Projects.Locations.CloudExadataInfrastructures.List(parent).Context(ctx)
		if flagOdbExInfPageSize > 0 {
			call = call.PageSize(flagOdbExInfPageSize)
		}
		if flagOdbExInfFilter != "" {
			call = call.Filter(flagOdbExInfFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing cloud Exadata infrastructures: %w", err)
		}
		all = append(all, resp.CloudExadataInfrastructures...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagOdbExInfFormat)
}
