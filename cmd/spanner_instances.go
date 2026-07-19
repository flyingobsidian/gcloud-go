package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner instances (#1210) ---

var spannerInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Cloud Spanner instances"}

var (
	flagSpInstFormat      string
	flagSpInstConfigFile  string
	flagSpInstUpdateMask  string
	flagSpInstConfig      string
	flagSpInstNodes       int64
	flagSpInstPU          int64
	flagSpInstDisplayName string
	flagSpInstFilter      string
	flagSpInstPageSize    int64
	flagSpInstTargetCfg   string
	flagSpInstIamMember   string
	flagSpInstIamRole     string
	flagSpInstIamCondExpr string
	flagSpInstIamCondT    string
	flagSpInstIamCondD    string
	flagSpInstIamAllCond  bool
)

var (
	spannerInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Cloud Spanner instance",
		Args: cobra.ExactArgs(1), RunE: runSpInstCreate,
	}
	spannerInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Cloud Spanner instance",
		Args: cobra.ExactArgs(1), RunE: runSpInstDelete,
	}
	spannerInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Cloud Spanner instance",
		Args: cobra.ExactArgs(1), RunE: runSpInstDescribe,
	}
	spannerInstListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Spanner instances",
		Args: cobra.NoArgs, RunE: runSpInstList,
	}
	spannerInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a Cloud Spanner instance",
		Args: cobra.ExactArgs(1), RunE: runSpInstUpdate,
	}
	spannerInstMoveCmd = &cobra.Command{
		Use: "move INSTANCE", Short: "Move a Cloud Spanner instance to a new instance configuration",
		Args: cobra.ExactArgs(1), RunE: runSpInstMove,
	}
	spannerInstGetLocsCmd = &cobra.Command{
		Use: "get-locations INSTANCE", Short: "Print the locations (replicas) of the instance's config",
		Args: cobra.ExactArgs(1), RunE: runSpInstGetLocations,
	}
	spannerInstGetIamCmd = &cobra.Command{
		Use: "get-iam-policy INSTANCE", Short: "Get the IAM policy for an instance",
		Args: cobra.ExactArgs(1), RunE: runSpInstGetIam,
	}
	spannerInstSetIamCmd = &cobra.Command{
		Use: "set-iam-policy INSTANCE POLICY_FILE", Short: "Set the IAM policy for an instance",
		Args: cobra.ExactArgs(2), RunE: runSpInstSetIam,
	}
	spannerInstAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding INSTANCE", Short: "Add an IAM policy binding on an instance",
		Args: cobra.ExactArgs(1), RunE: runSpInstAddIam,
	}
	spannerInstRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding INSTANCE", Short: "Remove an IAM policy binding from an instance",
		Args: cobra.ExactArgs(1), RunE: runSpInstRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		spannerInstCreateCmd, spannerInstDeleteCmd, spannerInstDescribeCmd, spannerInstListCmd,
		spannerInstUpdateCmd, spannerInstMoveCmd, spannerInstGetLocsCmd,
		spannerInstGetIamCmd, spannerInstSetIamCmd, spannerInstAddIamCmd, spannerInstRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSpInstFormat, "format", "", "Output format")
	}

	spannerInstCreateCmd.Flags().StringVar(&flagSpInstConfigFile, "config-file", "", "YAML/JSON file with the Instance body")
	spannerInstCreateCmd.Flags().StringVar(&flagSpInstConfig, "config", "", "Instance config (short name or full URI)")
	spannerInstCreateCmd.Flags().Int64Var(&flagSpInstNodes, "nodes", 0, "Number of nodes")
	spannerInstCreateCmd.Flags().Int64Var(&flagSpInstPU, "processing-units", 0, "Processing units")
	spannerInstCreateCmd.Flags().StringVar(&flagSpInstDisplayName, "display-name", "", "Display name")

	spannerInstUpdateCmd.Flags().StringVar(&flagSpInstConfigFile, "config-file", "", "YAML/JSON file with the Instance body")
	spannerInstUpdateCmd.Flags().StringVar(&flagSpInstUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	spannerInstUpdateCmd.Flags().Int64Var(&flagSpInstNodes, "nodes", 0, "Number of nodes")
	spannerInstUpdateCmd.Flags().Int64Var(&flagSpInstPU, "processing-units", 0, "Processing units")
	spannerInstUpdateCmd.Flags().StringVar(&flagSpInstDisplayName, "display-name", "", "Display name")

	spannerInstMoveCmd.Flags().StringVar(&flagSpInstTargetCfg, "target-config", "", "Target instance configuration (required)")
	_ = spannerInstMoveCmd.MarkFlagRequired("target-config")

	spannerInstListCmd.Flags().StringVar(&flagSpInstFilter, "filter", "", "Server-side filter expression")
	spannerInstListCmd.Flags().Int64Var(&flagSpInstPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{spannerInstAddIamCmd, spannerInstRemoveIamCmd} {
		spIamMemberFlags(c, &flagSpInstIamMember, &flagSpInstIamRole,
			&flagSpInstIamCondExpr, &flagSpInstIamCondT, &flagSpInstIamCondD)
	}
	spannerInstRemoveIamCmd.Flags().BoolVar(&flagSpInstIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	spannerInstancesCmd.AddCommand(all...)
	spannerCmd.AddCommand(spannerInstancesCmd)
}

// spannerInstBuild constructs an Instance from either the config file or the
// short flags; short flags override file fields when set.
func spannerInstBuild(existing *spanner.Instance) *spanner.Instance {
	if existing == nil {
		existing = &spanner.Instance{}
	}
	if flagSpInstConfig != "" {
		cfg, err := spannerInstanceConfig(flagSpInstConfig)
		if err == nil {
			existing.Config = cfg
		}
	}
	if flagSpInstNodes > 0 {
		existing.NodeCount = flagSpInstNodes
	}
	if flagSpInstPU > 0 {
		existing.ProcessingUnits = flagSpInstPU
	}
	if flagSpInstDisplayName != "" {
		existing.DisplayName = flagSpInstDisplayName
	}
	return existing
}

func runSpInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := spannerProject()
	if err != nil {
		return err
	}
	inst := &spanner.Instance{}
	if flagSpInstConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSpInstConfigFile, inst); err != nil {
			return err
		}
	}
	inst = spannerInstBuild(inst)
	body := &spanner.CreateInstanceRequest{Instance: inst, InstanceId: args[0]}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Printf("Create instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSpInstFormat)
}

func runSpInstDelete(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Printf("Deleted instance [%s].\n", args[0])
	return nil
}

func runSpInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagSpInstFormat)
}

func runSpInstList(cmd *cobra.Command, args []string) error {
	parent, err := spannerProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.Instance
	pageToken := ""
	for {
		call := svc.Projects.Instances.List(parent).Context(ctx)
		if flagSpInstFilter != "" {
			call = call.Filter(flagSpInstFilter)
		}
		if flagSpInstPageSize > 0 {
			call = call.PageSize(flagSpInstPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpInstFormat)
}

func runSpInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	inst := &spanner.Instance{}
	if flagSpInstConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSpInstConfigFile, inst); err != nil {
			return err
		}
	}
	inst = spannerInstBuild(inst)
	inst.Name = name
	mask := flagSpInstUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
	}
	body := &spanner.UpdateInstanceRequest{Instance: inst, FieldMask: mask}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return emitFormatted(op, flagSpInstFormat)
}

func runSpInstMove(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	target, err := spannerInstanceConfig(flagSpInstTargetCfg)
	if err != nil {
		return err
	}
	body := &spanner.MoveInstanceRequest{TargetConfig: target}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Move(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("moving instance: %w", err)
	}
	fmt.Printf("Move instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSpInstFormat)
}

func runSpInstGetLocations(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	inst, err := svc.Projects.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	if inst.Config == "" {
		return fmt.Errorf("instance [%s] has no config", args[0])
	}
	cfg, err := svc.Projects.InstanceConfigs.Get(inst.Config).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance config: %w", err)
	}
	return emitFormatted(cfg.Replicas, flagSpInstFormat)
}

func runSpInstGetIam(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagSpInstFormat)
}

func runSpInstSetIam(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	policy := &spanner.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("instance [%s]", args[0]))
	return emitFormatted(updated, flagSpInstFormat)
}

func runSpInstAddIam(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	spIamAddBinding(policy, flagSpInstIamRole, flagSpInstIamMember,
		spIamBuildCondition(flagSpInstIamCondExpr, flagSpInstIamCondT, flagSpInstIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("instance [%s]", args[0]))
	return emitFormatted(updated, flagSpInstFormat)
}

func runSpInstRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !spIamRemoveBinding(policy, flagSpInstIamRole, flagSpInstIamMember,
		spIamBuildCondition(flagSpInstIamCondExpr, flagSpInstIamCondT, flagSpInstIamCondD), flagSpInstIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagSpInstIamRole, flagSpInstIamMember)
	}
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("instance [%s]", args[0]))
	return emitFormatted(updated, flagSpInstFormat)
}
