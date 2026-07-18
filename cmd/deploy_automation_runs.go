package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy automation-runs (#1524) ---

var deployARCmd = &cobra.Command{Use: "automation-runs", Short: "Manage Cloud Deploy automation runs"}

var (
	flagDeployARRegion   string
	flagDeployARPipeline string
	flagDeployARFormat   string
	flagDeployARPageSize int64
)

var (
	deployARCancelCmd = &cobra.Command{
		Use: "cancel AUTOMATION_RUN", Short: "Cancel an automation run",
		Args: cobra.ExactArgs(1), RunE: runDeployARCancel,
	}
	deployARDescribeCmd = &cobra.Command{
		Use: "describe AUTOMATION_RUN", Short: "Describe an automation run",
		Args: cobra.ExactArgs(1), RunE: runDeployARDescribe,
	}
	deployARListCmd = &cobra.Command{
		Use: "list", Short: "List automation runs for a delivery pipeline",
		Args: cobra.NoArgs, RunE: runDeployARList,
	}
)

func init() {
	all := []*cobra.Command{deployARCancelCmd, deployARDescribeCmd, deployARListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployARRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployARPipeline, "delivery-pipeline", "",
			"Delivery pipeline (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&flagDeployARFormat, "format", "", "Output format")
	}
	deployARListCmd.Flags().Int64Var(&flagDeployARPageSize, "page-size", 0, "Maximum results per page")

	deployARCmd.AddCommand(all...)
	deployCmd.AddCommand(deployARCmd)
}

func deployARParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return deployChild("deliveryPipelines", flagDeployARPipeline, deployLocationParent(project, flagDeployARRegion)), nil
}

func deployARName(id string) (string, error) {
	parent, err := deployARParent()
	if err != nil {
		return "", err
	}
	return deployChild("automationRuns", id, parent), nil
}

func runDeployARCancel(cmd *cobra.Command, args []string) error {
	name, err := deployARName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.AutomationRuns.Cancel(name, &clouddeploy.CancelAutomationRunRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling automation run: %w", err)
	}
	fmt.Printf("Cancelled automation run [%s].\n", args[0])
	return nil
}

func runDeployARDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployARName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.AutomationRuns.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing automation run: %w", err)
	}
	return emitFormatted(got, flagDeployARFormat)
}

func runDeployARList(cmd *cobra.Command, args []string) error {
	parent, err := deployARParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*clouddeploy.AutomationRun
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.AutomationRuns.List(parent).Context(ctx)
		if flagDeployARPageSize > 0 {
			call = call.PageSize(flagDeployARPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing automation runs: %w", err)
		}
		all = append(all, resp.AutomationRuns...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDeployARFormat)
}
