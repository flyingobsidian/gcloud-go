package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy releases (#1530) ---

var deployRelCmd = &cobra.Command{Use: "releases", Short: "Manage Cloud Deploy releases"}

var (
	flagDeployRelRegion       string
	flagDeployRelPipeline     string
	flagDeployRelFormat       string
	flagDeployRelConfigFile   string
	flagDeployRelPageSize     int64
	flagDeployRelOverride     []string
	flagDeployRelValidateOnly bool
	flagDeployRelPromoteTgt   string
	flagDeployRelPromoteID    string
)

var (
	deployRelAbandonCmd = &cobra.Command{
		Use: "abandon RELEASE", Short: "Abandon a release",
		Args: cobra.ExactArgs(1), RunE: runDeployRelAbandon,
	}
	deployRelCreateCmd = &cobra.Command{
		Use: "create RELEASE", Short: "Create a release (body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runDeployRelCreate,
	}
	deployRelDescribeCmd = &cobra.Command{
		Use: "describe RELEASE", Short: "Describe a release",
		Args: cobra.ExactArgs(1), RunE: runDeployRelDescribe,
	}
	deployRelListCmd = &cobra.Command{
		Use: "list", Short: "List releases",
		Args: cobra.NoArgs, RunE: runDeployRelList,
	}
	deployRelPromoteCmd = &cobra.Command{
		Use: "promote RELEASE", Short: "Promote a release (create a rollout targeting a new target)",
		Args: cobra.ExactArgs(1), RunE: runDeployRelPromote,
	}
)

func init() {
	all := []*cobra.Command{
		deployRelAbandonCmd, deployRelCreateCmd, deployRelDescribeCmd, deployRelListCmd, deployRelPromoteCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployRelRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployRelPipeline, "delivery-pipeline", "",
			"Delivery pipeline that owns the release (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&flagDeployRelFormat, "format", "", "Output format")
	}
	deployRelCreateCmd.Flags().StringVar(&flagDeployRelConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Release body (required)")
	_ = deployRelCreateCmd.MarkFlagRequired("config-file")
	deployRelCreateCmd.Flags().StringSliceVar(&flagDeployRelOverride, "override-deploy-policies", nil,
		"Deploy policies to override (repeatable)")
	deployRelCreateCmd.Flags().BoolVar(&flagDeployRelValidateOnly, "validate-only", false,
		"Only validate the request; do not create the release")
	deployRelListCmd.Flags().Int64Var(&flagDeployRelPageSize, "page-size", 0, "Maximum results per page")
	deployRelPromoteCmd.Flags().StringVar(&flagDeployRelPromoteTgt, "to-target", "",
		"Target ID to promote the release to (required)")
	_ = deployRelPromoteCmd.MarkFlagRequired("to-target")
	deployRelPromoteCmd.Flags().StringVar(&flagDeployRelPromoteID, "rollout-id", "",
		"Rollout ID to assign to the new rollout (required)")
	_ = deployRelPromoteCmd.MarkFlagRequired("rollout-id")

	deployRelCmd.AddCommand(all...)
	deployCmd.AddCommand(deployRelCmd)
}

func deployRelParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return deployChild("deliveryPipelines", flagDeployRelPipeline, deployLocationParent(project, flagDeployRelRegion)), nil
}

func deployRelName(id string) (string, error) {
	parent, err := deployRelParent()
	if err != nil {
		return "", err
	}
	return deployChild("releases", id, parent), nil
}

func runDeployRelAbandon(cmd *cobra.Command, args []string) error {
	name, err := deployRelName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Abandon(name, &clouddeploy.AbandonReleaseRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("abandoning release: %w", err)
	}
	fmt.Printf("Abandoned release [%s].\n", args[0])
	return nil
}

func runDeployRelCreate(cmd *cobra.Command, args []string) error {
	parent, err := deployRelParent()
	if err != nil {
		return err
	}
	body := &clouddeploy.Release{}
	if err := loadYAMLOrJSONInto(flagDeployRelConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DeliveryPipelines.Releases.Create(parent, body).ReleaseId(args[0]).Context(ctx)
	if len(flagDeployRelOverride) > 0 {
		call = call.OverrideDeployPolicy(flagDeployRelOverride...)
	}
	if flagDeployRelValidateOnly {
		call = call.ValidateOnly(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating release: %w", err)
	}
	fmt.Printf("Create request issued for release [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDeployRelFormat)
}

func runDeployRelDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployRelName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Releases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing release: %w", err)
	}
	return emitFormatted(got, flagDeployRelFormat)
}

func runDeployRelList(cmd *cobra.Command, args []string) error {
	parent, err := deployRelParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*clouddeploy.Release
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.Releases.List(parent).Context(ctx)
		if flagDeployRelPageSize > 0 {
			call = call.PageSize(flagDeployRelPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing releases: %w", err)
		}
		all = append(all, resp.Releases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDeployRelFormat)
}

func runDeployRelPromote(cmd *cobra.Command, args []string) error {
	release, err := deployRelName(args[0])
	if err != nil {
		return err
	}
	rollout := &clouddeploy.Rollout{
		TargetId: flagDeployRelPromoteTgt,
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Create(release, rollout).
		RolloutId(flagDeployRelPromoteID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("promoting release: %w", err)
	}
	fmt.Printf("Promote issued for release [%s] to target [%s] (operation: %s).\n",
		args[0], flagDeployRelPromoteTgt, op.Name)
	return emitFormatted(op, flagDeployRelFormat)
}
