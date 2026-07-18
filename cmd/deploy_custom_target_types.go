package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy custom-target-types (#1526) ---

var deployCTTCmd = &cobra.Command{Use: "custom-target-types", Short: "Manage Cloud Deploy custom target types"}

var (
	flagDeployCTTRegion   string
	flagDeployCTTFormat   string
	flagDeployCTTPageSize int64

	flagDeployCTTIamMember   string
	flagDeployCTTIamRole     string
	flagDeployCTTIamCondExpr string
	flagDeployCTTIamCondT    string
	flagDeployCTTIamCondD    string
	flagDeployCTTIamAllCond  bool
)

var (
	deployCTTDeleteCmd = &cobra.Command{
		Use: "delete CUSTOM_TARGET_TYPE", Short: "Delete a custom target type",
		Args: cobra.ExactArgs(1), RunE: runDeployCTTDelete,
	}
	deployCTTDescribeCmd = &cobra.Command{
		Use: "describe CUSTOM_TARGET_TYPE", Short: "Describe a custom target type",
		Args: cobra.ExactArgs(1), RunE: runDeployCTTDescribe,
	}
	deployCTTExportCmd = &cobra.Command{
		Use: "export CUSTOM_TARGET_TYPE", Short: "Export a custom target type",
		Args: cobra.ExactArgs(1), RunE: runDeployCTTExport,
	}
	deployCTTListCmd = &cobra.Command{
		Use: "list", Short: "List custom target types",
		Args: cobra.NoArgs, RunE: runDeployCTTList,
	}
	deployCTTGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CUSTOM_TARGET_TYPE", Short: "Get the IAM policy",
		Args: cobra.ExactArgs(1), RunE: runDeployCTTGetIam,
	}
	deployCTTSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CUSTOM_TARGET_TYPE POLICY_FILE", Short: "Set the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runDeployCTTSetIam,
	}
	deployCTTAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding CUSTOM_TARGET_TYPE", Short: "Add an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: runDeployCTTAddIam,
	}
	deployCTTRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding CUSTOM_TARGET_TYPE", Short: "Remove an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: runDeployCTTRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		deployCTTDeleteCmd, deployCTTDescribeCmd, deployCTTExportCmd, deployCTTListCmd,
		deployCTTGetIamCmd, deployCTTSetIamCmd, deployCTTAddIamCmd, deployCTTRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployCTTRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployCTTFormat, "format", "", "Output format")
	}
	deployCTTListCmd.Flags().Int64Var(&flagDeployCTTPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{deployCTTAddIamCmd, deployCTTRemoveIamCmd} {
		deployIamMemberFlags(c, &flagDeployCTTIamMember, &flagDeployCTTIamRole,
			&flagDeployCTTIamCondExpr, &flagDeployCTTIamCondT, &flagDeployCTTIamCondD)
	}
	deployCTTRemoveIamCmd.Flags().BoolVar(&flagDeployCTTIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	deployCTTCmd.AddCommand(all...)
	deployCmd.AddCommand(deployCTTCmd)
}

func deployCTTParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return deployLocationParent(project, flagDeployCTTRegion), nil
}

func deployCTTName(id string) (string, error) {
	parent, err := deployCTTParent()
	if err != nil {
		return "", err
	}
	return deployChild("customTargetTypes", id, parent), nil
}

func runDeployCTTDelete(cmd *cobra.Command, args []string) error {
	name, err := deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CustomTargetTypes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting custom target type: %w", err)
	}
	fmt.Printf("Delete request issued for custom target type [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDeployCTTFormat)
}

func runDeployCTTDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CustomTargetTypes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing custom target type: %w", err)
	}
	return emitFormatted(got, flagDeployCTTFormat)
}

func runDeployCTTExport(cmd *cobra.Command, args []string) error {
	name, err := deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CustomTargetTypes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting custom target type: %w", err)
	}
	format := flagDeployCTTFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormatted(got, format)
}

func runDeployCTTList(cmd *cobra.Command, args []string) error {
	parent, err := deployCTTParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*clouddeploy.CustomTargetType
	pageToken := ""
	for {
		call := svc.Projects.Locations.CustomTargetTypes.List(parent).Context(ctx)
		if flagDeployCTTPageSize > 0 {
			call = call.PageSize(flagDeployCTTPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing custom target types: %w", err)
		}
		all = append(all, resp.CustomTargetTypes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDeployCTTFormat)
}

func runDeployCTTGetIam(cmd *cobra.Command, args []string) error {
	name, err := deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CustomTargetTypes.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDeployCTTFormat)
}

func runDeployCTTSetIam(cmd *cobra.Command, args []string) error {
	name, err := deployCTTName(args[0])
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
	updated, err := svc.Projects.Locations.CustomTargetTypes.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("custom target type [%s]", args[0]))
	return emitFormatted(updated, flagDeployCTTFormat)
}

func runDeployCTTAddIam(cmd *cobra.Command, args []string) error {
	name, err := deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CustomTargetTypes.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	deployAddBinding(policy, flagDeployCTTIamRole, flagDeployCTTIamMember,
		deployBuildCondition(flagDeployCTTIamCondExpr, flagDeployCTTIamCondT, flagDeployCTTIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.CustomTargetTypes.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("custom target type [%s]", args[0]))
	return emitFormatted(updated, flagDeployCTTFormat)
}

func runDeployCTTRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CustomTargetTypes.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !deployRemoveBinding(policy, flagDeployCTTIamRole, flagDeployCTTIamMember,
		deployBuildCondition(flagDeployCTTIamCondExpr, flagDeployCTTIamCondT, flagDeployCTTIamCondD), flagDeployCTTIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDeployCTTIamRole, flagDeployCTTIamMember)
	}
	updated, err := svc.Projects.Locations.CustomTargetTypes.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("custom target type [%s]", args[0]))
	return emitFormatted(updated, flagDeployCTTFormat)
}
