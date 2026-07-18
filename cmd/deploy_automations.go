package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy automations (#1525) ---

var deployAutoCmd = &cobra.Command{Use: "automations", Short: "Manage Cloud Deploy automations"}

var (
	flagDeployAutoRegion   string
	flagDeployAutoPipeline string
	flagDeployAutoFormat   string
	flagDeployAutoPageSize int64
)

var (
	deployAutoDeleteCmd = &cobra.Command{
		Use: "delete AUTOMATION", Short: "Delete an automation",
		Args: cobra.ExactArgs(1), RunE: runDeployAutoDelete,
	}
	deployAutoDescribeCmd = &cobra.Command{
		Use: "describe AUTOMATION", Short: "Describe an automation",
		Args: cobra.ExactArgs(1), RunE: runDeployAutoDescribe,
	}
	deployAutoExportCmd = &cobra.Command{
		Use: "export AUTOMATION", Short: "Export an automation",
		Args: cobra.ExactArgs(1), RunE: runDeployAutoExport,
	}
	deployAutoListCmd = &cobra.Command{
		Use: "list", Short: "List automations for a delivery pipeline",
		Args: cobra.NoArgs, RunE: runDeployAutoList,
	}
)

func init() {
	all := []*cobra.Command{deployAutoDeleteCmd, deployAutoDescribeCmd, deployAutoExportCmd, deployAutoListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployAutoRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployAutoPipeline, "delivery-pipeline", "",
			"Delivery pipeline (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&flagDeployAutoFormat, "format", "", "Output format")
	}
	deployAutoListCmd.Flags().Int64Var(&flagDeployAutoPageSize, "page-size", 0, "Maximum results per page")

	deployAutoCmd.AddCommand(all...)
	deployCmd.AddCommand(deployAutoCmd)
}

func deployAutoParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return deployChild("deliveryPipelines", flagDeployAutoPipeline, deployLocationParent(project, flagDeployAutoRegion)), nil
}

func deployAutoName(id string) (string, error) {
	parent, err := deployAutoParent()
	if err != nil {
		return "", err
	}
	return deployChild("automations", id, parent), nil
}

func runDeployAutoDelete(cmd *cobra.Command, args []string) error {
	name, err := deployAutoName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DeliveryPipelines.Automations.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting automation: %w", err)
	}
	fmt.Printf("Delete request issued for automation [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDeployAutoFormat)
}

func runDeployAutoDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployAutoName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Automations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing automation: %w", err)
	}
	return emitFormatted(got, flagDeployAutoFormat)
}

func runDeployAutoExport(cmd *cobra.Command, args []string) error {
	name, err := deployAutoName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Automations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting automation: %w", err)
	}
	format := flagDeployAutoFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormatted(got, format)
}

func runDeployAutoList(cmd *cobra.Command, args []string) error {
	parent, err := deployAutoParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*clouddeploy.Automation
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.Automations.List(parent).Context(ctx)
		if flagDeployAutoPageSize > 0 {
			call = call.PageSize(flagDeployAutoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing automations: %w", err)
		}
		all = append(all, resp.Automations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDeployAutoFormat)
}
