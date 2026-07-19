package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable instances (#1306) ---

var bigtableInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Bigtable instances"}

var (
	flagBtInstFormat     string
	flagBtInstConfigFile string
	flagBtInstIamMember  string
	flagBtInstIamRole    string
	flagBtInstIamCondE   string
	flagBtInstIamCondT   string
	flagBtInstIamCondD   string
	flagBtInstIamAllCond bool
)

var (
	bigtableInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: runBtInstCreate,
	}
	bigtableInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: runBtInstDelete,
	}
	bigtableInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: runBtInstDescribe,
	}
	bigtableInstListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable instances in the project",
		Args: cobra.NoArgs, RunE: runBtInstList,
	}
	bigtableInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: runBtInstUpdate,
	}
	bigtableInstGetIamCmd = &cobra.Command{
		Use: "get-iam-policy INSTANCE", Short: "Get the IAM policy for a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: runBtInstGetIam,
	}
	bigtableInstSetIamCmd = &cobra.Command{
		Use: "set-iam-policy INSTANCE POLICY_FILE", Short: "Set the IAM policy for a Bigtable instance",
		Args: cobra.ExactArgs(2), RunE: runBtInstSetIam,
	}
	bigtableInstAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding INSTANCE", Short: "Add an IAM policy binding to a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: runBtInstAddIam,
	}
	bigtableInstRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding INSTANCE", Short: "Remove an IAM policy binding from a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: runBtInstRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		bigtableInstCreateCmd, bigtableInstDeleteCmd, bigtableInstDescribeCmd,
		bigtableInstListCmd, bigtableInstUpdateCmd,
		bigtableInstGetIamCmd, bigtableInstSetIamCmd, bigtableInstAddIamCmd, bigtableInstRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBtInstFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableInstCreateCmd, bigtableInstUpdateCmd} {
		c.Flags().StringVar(&flagBtInstConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}

	for _, c := range []*cobra.Command{bigtableInstAddIamCmd, bigtableInstRemoveIamCmd} {
		c.Flags().StringVar(&flagBtInstIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagBtInstIamRole, "role", "", "IAM role to bind (required)")
		c.Flags().StringVar(&flagBtInstIamCondE, "condition-expression", "", "CEL condition expression")
		c.Flags().StringVar(&flagBtInstIamCondT, "condition-title", "", "Condition title")
		c.Flags().StringVar(&flagBtInstIamCondD, "condition-description", "", "Condition description")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	bigtableInstRemoveIamCmd.Flags().BoolVar(&flagBtInstIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	bigtableInstancesCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableInstancesCmd)
}

func btInstBuildCondition() *bigtableadmin.Expr {
	if flagBtInstIamCondE == "" && flagBtInstIamCondT == "" && flagBtInstIamCondD == "" {
		return nil
	}
	return &bigtableadmin.Expr{
		Expression:  flagBtInstIamCondE,
		Title:       flagBtInstIamCondT,
		Description: flagBtInstIamCondD,
	}
}

func runBtInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := btProjectName()
	if err != nil {
		return err
	}
	req := &bigtableadmin.CreateInstanceRequest{}
	if err := loadYAMLOrJSONInto(flagBtInstConfigFile, req); err != nil {
		return err
	}
	if req.InstanceId == "" {
		req.InstanceId = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Create(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Printf("Create request issued for instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtInstFormat)
}

func runBtInstDelete(cmd *cobra.Command, args []string) error {
	name, err := btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Printf("Deleted instance [%s].\n", args[0])
	return nil
}

func runBtInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagBtInstFormat)
}

func runBtInstList(cmd *cobra.Command, args []string) error {
	parent, err := btProjectName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.Instance
	pageToken := ""
	for {
		call := svc.Projects.Instances.List(parent).Context(ctx)
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
	return emitFormatted(all, flagBtInstFormat)
}

func runBtInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := btInstanceName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.Instance{}
	if err := loadYAMLOrJSONInto(flagBtInstConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	fmt.Printf("Updated instance [%s].\n", args[0])
	return emitFormatted(got, flagBtInstFormat)
}

func runBtInstGetIam(cmd *cobra.Command, args []string) error {
	name, err := btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagBtInstFormat)
}

func runBtInstSetIam(cmd *cobra.Command, args []string) error {
	name, err := btInstanceName(args[0])
	if err != nil {
		return err
	}
	policy := &bigtableadmin.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for instance [%s].\n", args[0])
	return emitFormatted(updated, flagBtInstFormat)
}

func runBtInstAddIam(cmd *cobra.Command, args []string) error {
	name, err := btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	bigtableAddBindingToPolicy(policy, flagBtInstIamRole, flagBtInstIamMember, btInstBuildCondition())
	policy.Version = 3
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for instance [%s].\n", args[0])
	return emitFormatted(updated, flagBtInstFormat)
}

func runBtInstRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !bigtableRemoveBindingFromPolicy(policy, flagBtInstIamRole, flagBtInstIamMember, btInstBuildCondition(), flagBtInstIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagBtInstIamRole, flagBtInstIamMember)
	}
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for instance [%s].\n", args[0])
	return emitFormatted(updated, flagBtInstFormat)
}
