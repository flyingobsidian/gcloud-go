package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy delivery-pipelines (#1527) ---

var deployDPCmd = &cobra.Command{Use: "delivery-pipelines", Short: "Manage Cloud Deploy delivery pipelines"}

var (
	flagDeployDPRegion   string
	flagDeployDPFormat   string
	flagDeployDPForce    bool
	flagDeployDPPageSize int64

	flagDeployDPIamMember   string
	flagDeployDPIamRole     string
	flagDeployDPIamCondExpr string
	flagDeployDPIamCondT    string
	flagDeployDPIamCondD    string
	flagDeployDPIamAllCond  bool
)

var (
	deployDPDeleteCmd = &cobra.Command{
		Use: "delete PIPELINE", Short: "Delete a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: runDeployDPDelete,
	}
	deployDPDescribeCmd = &cobra.Command{
		Use: "describe PIPELINE", Short: "Describe a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: runDeployDPDescribe,
	}
	deployDPExportCmd = &cobra.Command{
		Use: "export PIPELINE", Short: "Export a delivery pipeline (YAML/JSON)",
		Args: cobra.ExactArgs(1), RunE: runDeployDPExport,
	}
	deployDPListCmd = &cobra.Command{
		Use: "list", Short: "List delivery pipelines",
		Args: cobra.NoArgs, RunE: runDeployDPList,
	}
	deployDPGetIamCmd = &cobra.Command{
		Use: "get-iam-policy PIPELINE", Short: "Get the IAM policy for a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: runDeployDPGetIam,
	}
	deployDPSetIamCmd = &cobra.Command{
		Use: "set-iam-policy PIPELINE POLICY_FILE", Short: "Set the IAM policy for a delivery pipeline",
		Args: cobra.ExactArgs(2), RunE: runDeployDPSetIam,
	}
	deployDPAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding PIPELINE", Short: "Add an IAM policy binding to a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: runDeployDPAddIam,
	}
	deployDPRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding PIPELINE", Short: "Remove an IAM policy binding from a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: runDeployDPRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		deployDPDeleteCmd, deployDPDescribeCmd, deployDPExportCmd, deployDPListCmd,
		deployDPGetIamCmd, deployDPSetIamCmd, deployDPAddIamCmd, deployDPRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployDPRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployDPFormat, "format", "", "Output format")
	}
	deployDPDeleteCmd.Flags().BoolVar(&flagDeployDPForce, "force", false, "Force delete of a pipeline with children")
	deployDPListCmd.Flags().Int64Var(&flagDeployDPPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{deployDPAddIamCmd, deployDPRemoveIamCmd} {
		deployIamMemberFlags(c, &flagDeployDPIamMember, &flagDeployDPIamRole,
			&flagDeployDPIamCondExpr, &flagDeployDPIamCondT, &flagDeployDPIamCondD)
	}
	deployDPRemoveIamCmd.Flags().BoolVar(&flagDeployDPIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	deployDPCmd.AddCommand(all...)
	deployCmd.AddCommand(deployDPCmd)
}

func deployDPParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return deployLocationParent(project, flagDeployDPRegion), nil
}

func deployDPName(id string) (string, error) {
	parent, err := deployDPParent()
	if err != nil {
		return "", err
	}
	return deployChild("deliveryPipelines", id, parent), nil
}

func runDeployDPDelete(cmd *cobra.Command, args []string) error {
	name, err := deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DeliveryPipelines.Delete(name).Context(ctx)
	if flagDeployDPForce {
		call = call.Force(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting delivery pipeline: %w", err)
	}
	fmt.Printf("Delete request issued for delivery pipeline [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDeployDPFormat)
}

func runDeployDPDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing delivery pipeline: %w", err)
	}
	return emitFormatted(got, flagDeployDPFormat)
}

func runDeployDPExport(cmd *cobra.Command, args []string) error {
	name, err := deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting delivery pipeline: %w", err)
	}
	format := flagDeployDPFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormatted(got, format)
}

func runDeployDPList(cmd *cobra.Command, args []string) error {
	parent, err := deployDPParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*clouddeploy.DeliveryPipeline
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.List(parent).Context(ctx)
		if flagDeployDPPageSize > 0 {
			call = call.PageSize(flagDeployDPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing delivery pipelines: %w", err)
		}
		all = append(all, resp.DeliveryPipelines...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDeployDPFormat)
}

func runDeployDPGetIam(cmd *cobra.Command, args []string) error {
	name, err := deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeliveryPipelines.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDeployDPFormat)
}

func runDeployDPSetIam(cmd *cobra.Command, args []string) error {
	name, err := deployDPName(args[0])
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
	updated, err := svc.Projects.Locations.DeliveryPipelines.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("delivery pipeline [%s]", args[0]))
	return emitFormatted(updated, flagDeployDPFormat)
}

func runDeployDPAddIam(cmd *cobra.Command, args []string) error {
	name, err := deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeliveryPipelines.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	deployAddBinding(policy, flagDeployDPIamRole, flagDeployDPIamMember,
		deployBuildCondition(flagDeployDPIamCondExpr, flagDeployDPIamCondT, flagDeployDPIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.DeliveryPipelines.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("delivery pipeline [%s]", args[0]))
	return emitFormatted(updated, flagDeployDPFormat)
}

func runDeployDPRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeliveryPipelines.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !deployRemoveBinding(policy, flagDeployDPIamRole, flagDeployDPIamMember,
		deployBuildCondition(flagDeployDPIamCondExpr, flagDeployDPIamCondT, flagDeployDPIamCondD), flagDeployDPIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDeployDPIamRole, flagDeployDPIamMember)
	}
	updated, err := svc.Projects.Locations.DeliveryPipelines.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	deployUpdatedIam(fmt.Sprintf("delivery pipeline [%s]", args[0]))
	return emitFormatted(updated, flagDeployDPFormat)
}
