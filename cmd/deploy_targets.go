package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy targets (#1532) ---

var deployTargetsCmd = &cobra.Command{Use: "targets", Short: "Manage Cloud Deploy targets"}

var (
	flagDeployTgtRegion   string
	flagDeployTgtFormat   string
	flagDeployTgtPageSize int64

	// rollback/redeploy shared flags
	flagDeployTgtPipeline    string
	flagDeployTgtRelease     string
	flagDeployTgtRolloutID   string
	flagDeployTgtDescription string
	flagDeployTgtStartPhase  string
	flagDeployTgtOverridePol []string

	flagDeployTgtIamMember   string
	flagDeployTgtIamRole     string
	flagDeployTgtIamCondExpr string
	flagDeployTgtIamCondT    string
	flagDeployTgtIamCondD    string
	flagDeployTgtIamAllCond  bool
)

var (
	deployTgtDeleteCmd = &cobra.Command{
		Use: "delete TARGET", Short: "Delete a target",
		Args: cobra.ExactArgs(1), RunE: runDeployTgtDelete,
	}
	deployTgtDescribeCmd = &cobra.Command{
		Use: "describe TARGET", Short: "Describe a target",
		Args: cobra.ExactArgs(1), RunE: runDeployTgtDescribe,
	}
	deployTgtExportCmd = &cobra.Command{
		Use: "export TARGET", Short: "Export a target",
		Args: cobra.ExactArgs(1), RunE: runDeployTgtExport,
	}
	deployTgtListCmd = &cobra.Command{
		Use: "list", Short: "List targets",
		Args: cobra.NoArgs, RunE: runDeployTgtList,
	}
	deployTgtRedeployCmd = &cobra.Command{
		Use: "redeploy TARGET",
		Short: "Redeploy the most recent release for a target by creating a new rollout",
		Args:  cobra.ExactArgs(1), RunE: runDeployTgtRedeploy,
	}
	deployTgtRollbackCmd = &cobra.Command{
		Use: "rollback TARGET",
		Short: "Roll a target back to a previous release",
		Args:  cobra.ExactArgs(1), RunE: runDeployTgtRollback,
	}
	deployTgtGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TARGET", Short: "Get the IAM policy",
		Args: cobra.ExactArgs(1), RunE: runDeployTgtGetIam,
	}
	deployTgtSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TARGET POLICY_FILE", Short: "Set the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runDeployTgtSetIam,
	}
	deployTgtAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TARGET", Short: "Add an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: runDeployTgtAddIam,
	}
	deployTgtRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TARGET", Short: "Remove an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: runDeployTgtRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		deployTgtDeleteCmd, deployTgtDescribeCmd, deployTgtExportCmd, deployTgtListCmd,
		deployTgtRedeployCmd, deployTgtRollbackCmd, deployTgtGetIamCmd, deployTgtSetIamCmd,
		deployTgtAddIamCmd, deployTgtRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployTgtRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployTgtFormat, "format", "", "Output format")
	}
	deployTgtListCmd.Flags().Int64Var(&flagDeployTgtPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{deployTgtRedeployCmd, deployTgtRollbackCmd} {
		c.Flags().StringVar(&flagDeployTgtPipeline, "delivery-pipeline", "",
			"Delivery pipeline that owns the target (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&flagDeployTgtRelease, "release", "",
			"Release to redeploy/roll back to (required for rollback; defaults to the target's latest for redeploy)")
		c.Flags().StringVar(&flagDeployTgtRolloutID, "rollout-id", "",
			"Optional rollout ID (auto-generated when omitted)")
		c.Flags().StringVar(&flagDeployTgtDescription, "description", "", "Rollout description")
		c.Flags().StringVar(&flagDeployTgtStartPhase, "starting-phase-id", "", "Starting phase for the rollout")
		c.Flags().StringSliceVar(&flagDeployTgtOverridePol, "override-deploy-policies", nil,
			"Deploy policies to override (repeatable)")
	}
	_ = deployTgtRollbackCmd.MarkFlagRequired("release")

	for _, c := range []*cobra.Command{deployTgtAddIamCmd, deployTgtRemoveIamCmd} {
		deployIamMemberFlags(c, &flagDeployTgtIamMember, &flagDeployTgtIamRole,
			&flagDeployTgtIamCondExpr, &flagDeployTgtIamCondT, &flagDeployTgtIamCondD)
	}
	deployTgtRemoveIamCmd.Flags().BoolVar(&flagDeployTgtIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	deployTargetsCmd.AddCommand(all...)
	deployCmd.AddCommand(deployTargetsCmd)
}

func deployTgtName(id string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return deployChild("targets", id, deployLocationParent(project, flagDeployTgtRegion)), nil
}

func runDeployTgtDelete(cmd *cobra.Command, args []string) error {
	name, err := deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Targets.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting target: %w", err)
	}
	fmt.Printf("Delete request issued for target [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDeployTgtFormat)
}

func runDeployTgtDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Targets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing target: %w", err)
	}
	return emitFormatted(got, flagDeployTgtFormat)
}

func runDeployTgtExport(cmd *cobra.Command, args []string) error {
	name, err := deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Targets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting target: %w", err)
	}
	format := flagDeployTgtFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormatted(got, format)
}

func runDeployTgtList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := deployLocationParent(project, flagDeployTgtRegion)
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*clouddeploy.Target
	pageToken := ""
	for {
		call := svc.Projects.Locations.Targets.List(parent).Context(ctx)
		if flagDeployTgtPageSize > 0 {
			call = call.PageSize(flagDeployTgtPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing targets: %w", err)
		}
		all = append(all, resp.Targets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDeployTgtFormat)
}

func runDeployTgtRollback(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := deployLocationParent(project, flagDeployTgtRegion)
	pipeline := deployChild("deliveryPipelines", flagDeployTgtPipeline, parent)
	req := &clouddeploy.RollbackTargetRequest{
		TargetId:             args[0],
		ReleaseId:            flagDeployTgtRelease,
		RolloutId:            flagDeployTgtRolloutID,
		OverrideDeployPolicy: flagDeployTgtOverridePol,
	}
	if flagDeployTgtDescription != "" || flagDeployTgtStartPhase != "" {
		req.RollbackConfig = &clouddeploy.RollbackTargetConfig{
			Rollout: &clouddeploy.Rollout{
				Description:     flagDeployTgtDescription,
			},
			StartingPhaseId: flagDeployTgtStartPhase,
		}
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DeliveryPipelines.RollbackTarget(pipeline, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back target: %w", err)
	}
	fmt.Printf("Rollback issued for target [%s].\n", args[0])
	return emitFormatted(resp, flagDeployTgtFormat)
}

func runDeployTgtRedeploy(cmd *cobra.Command, args []string) error {
	// Redeploy: create a new Rollout on the target's most-recent (or --release'd) release.
	if flagDeployTgtRelease == "" {
		return fmt.Errorf("--release is required (Cloud Deploy has no server-side 'latest' resolver here)")
	}
	if flagDeployTgtRolloutID == "" {
		return fmt.Errorf("--rollout-id is required (auto-generation not supported)")
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := deployLocationParent(project, flagDeployTgtRegion)
	release := deployChild("releases", flagDeployTgtRelease,
		deployChild("deliveryPipelines", flagDeployTgtPipeline, parent))
	rollout := &clouddeploy.Rollout{
		TargetId:    args[0],
		Description: flagDeployTgtDescription,
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Create(release, rollout).RolloutId(flagDeployTgtRolloutID).Context(ctx)
	if flagDeployTgtStartPhase != "" {
		call = call.StartingPhaseId(flagDeployTgtStartPhase)
	}
	_ = flagDeployTgtOverridePol
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("redeploying: %w", err)
	}
	fmt.Printf("Redeploy issued for target [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDeployTgtFormat)
}

func runDeployTgtGetIam(cmd *cobra.Command, args []string) error {
	name, err := deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Targets.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDeployTgtFormat)
}

func runDeployTgtSetIam(cmd *cobra.Command, args []string) error {
	name, err := deployTgtName(args[0])
	if err != nil {
		return err
	}
	policy := &clouddeploy.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Targets.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("target [%s]", args[0]))
	return emitFormatted(updated, flagDeployTgtFormat)
}

func runDeployTgtAddIam(cmd *cobra.Command, args []string) error {
	name, err := deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Targets.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	deployAddBinding(policy, flagDeployTgtIamRole, flagDeployTgtIamMember,
		deployBuildCondition(flagDeployTgtIamCondExpr, flagDeployTgtIamCondT, flagDeployTgtIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Targets.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("target [%s]", args[0]))
	return emitFormatted(updated, flagDeployTgtFormat)
}

func runDeployTgtRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Targets.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !deployRemoveBinding(policy, flagDeployTgtIamRole, flagDeployTgtIamMember,
		deployBuildCondition(flagDeployTgtIamCondExpr, flagDeployTgtIamCondT, flagDeployTgtIamCondD), flagDeployTgtIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDeployTgtIamRole, flagDeployTgtIamMember)
	}
	updated, err := svc.Projects.Locations.Targets.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("target [%s]", args[0]))
	return emitFormatted(updated, flagDeployTgtFormat)
}
