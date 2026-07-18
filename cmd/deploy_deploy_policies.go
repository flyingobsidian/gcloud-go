package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy deploy-policies (#1528) ---

var deployDPolCmd = &cobra.Command{Use: "deploy-policies", Short: "Manage Cloud Deploy deploy policies"}

var (
	flagDeployDPolRegion string
	flagDeployDPolFormat string

	flagDeployDPolIamMember   string
	flagDeployDPolIamRole     string
	flagDeployDPolIamCondExpr string
	flagDeployDPolIamCondT    string
	flagDeployDPolIamCondD    string
	flagDeployDPolIamAllCond  bool
)

var (
	deployDPolDeleteCmd = &cobra.Command{
		Use: "delete DEPLOY_POLICY", Short: "Delete a deploy policy",
		Args: cobra.ExactArgs(1), RunE: runDeployDPolDelete,
	}
	deployDPolDescribeCmd = &cobra.Command{
		Use: "describe DEPLOY_POLICY", Short: "Describe a deploy policy",
		Args: cobra.ExactArgs(1), RunE: runDeployDPolDescribe,
	}
	deployDPolExportCmd = &cobra.Command{
		Use: "export DEPLOY_POLICY", Short: "Export a deploy policy",
		Args: cobra.ExactArgs(1), RunE: runDeployDPolExport,
	}
	deployDPolGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DEPLOY_POLICY", Short: "Get the IAM policy",
		Args: cobra.ExactArgs(1), RunE: runDeployDPolGetIam,
	}
	deployDPolSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DEPLOY_POLICY POLICY_FILE", Short: "Set the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runDeployDPolSetIam,
	}
	deployDPolAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding DEPLOY_POLICY", Short: "Add an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: runDeployDPolAddIam,
	}
	deployDPolRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding DEPLOY_POLICY", Short: "Remove an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: runDeployDPolRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		deployDPolDeleteCmd, deployDPolDescribeCmd, deployDPolExportCmd,
		deployDPolGetIamCmd, deployDPolSetIamCmd, deployDPolAddIamCmd, deployDPolRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployDPolRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployDPolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{deployDPolAddIamCmd, deployDPolRemoveIamCmd} {
		deployIamMemberFlags(c, &flagDeployDPolIamMember, &flagDeployDPolIamRole,
			&flagDeployDPolIamCondExpr, &flagDeployDPolIamCondT, &flagDeployDPolIamCondD)
	}
	deployDPolRemoveIamCmd.Flags().BoolVar(&flagDeployDPolIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	deployDPolCmd.AddCommand(all...)
	deployCmd.AddCommand(deployDPolCmd)
}

func deployDPolName(id string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return deployChild("deployPolicies", id, deployLocationParent(project, flagDeployDPolRegion)), nil
}

func runDeployDPolDelete(cmd *cobra.Command, args []string) error {
	name, err := deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DeployPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting deploy policy: %w", err)
	}
	fmt.Printf("Delete request issued for deploy policy [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDeployDPolFormat)
}

func runDeployDPolDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeployPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing deploy policy: %w", err)
	}
	return emitFormatted(got, flagDeployDPolFormat)
}

func runDeployDPolExport(cmd *cobra.Command, args []string) error {
	name, err := deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeployPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting deploy policy: %w", err)
	}
	format := flagDeployDPolFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormatted(got, format)
}

func runDeployDPolGetIam(cmd *cobra.Command, args []string) error {
	name, err := deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeployPolicies.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDeployDPolFormat)
}

func runDeployDPolSetIam(cmd *cobra.Command, args []string) error {
	name, err := deployDPolName(args[0])
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
	updated, err := svc.Projects.Locations.DeployPolicies.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("deploy policy [%s]", args[0]))
	return emitFormatted(updated, flagDeployDPolFormat)
}

func runDeployDPolAddIam(cmd *cobra.Command, args []string) error {
	name, err := deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeployPolicies.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	deployAddBinding(policy, flagDeployDPolIamRole, flagDeployDPolIamMember,
		deployBuildCondition(flagDeployDPolIamCondExpr, flagDeployDPolIamCondT, flagDeployDPolIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.DeployPolicies.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("deploy policy [%s]", args[0]))
	return emitFormatted(updated, flagDeployDPolFormat)
}

func runDeployDPolRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeployPolicies.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !deployRemoveBinding(policy, flagDeployDPolIamRole, flagDeployDPolIamMember,
		deployBuildCondition(flagDeployDPolIamCondExpr, flagDeployDPolIamCondT, flagDeployDPolIamCondD), flagDeployDPolIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDeployDPolIamRole, flagDeployDPolIamMember)
	}
	updated, err := svc.Projects.Locations.DeployPolicies.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("deploy policy [%s]", args[0]))
	return emitFormatted(updated, flagDeployDPolFormat)
}
