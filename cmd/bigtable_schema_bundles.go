package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable schema-bundles (#1489) ---

var bigtableSBCmd = &cobra.Command{Use: "schema-bundles", Short: "Manage Cloud Bigtable schema bundles"}

var (
	flagBTSBInstance   string
	flagBTSBTable      string
	flagBTSBFormat     string
	flagBTSBConfigFile string
	flagBTSBUpdateMask string
	flagBTSBPageSize   int64

	flagBTSBIamMember   string
	flagBTSBIamRole     string
	flagBTSBIamCondExpr string
	flagBTSBIamCondT    string
	flagBTSBIamCondD    string
	flagBTSBIamAllCond  bool
)

var (
	bigtableSBCreateCmd = &cobra.Command{
		Use: "create SCHEMA_BUNDLE", Short: "Create a schema bundle",
		Args: cobra.ExactArgs(1), RunE: runBTSBCreate,
	}
	bigtableSBDeleteCmd = &cobra.Command{
		Use: "delete SCHEMA_BUNDLE", Short: "Delete a schema bundle",
		Args: cobra.ExactArgs(1), RunE: runBTSBDelete,
	}
	bigtableSBDescribeCmd = &cobra.Command{
		Use: "describe SCHEMA_BUNDLE", Short: "Describe a schema bundle",
		Args: cobra.ExactArgs(1), RunE: runBTSBDescribe,
	}
	bigtableSBListCmd = &cobra.Command{
		Use: "list", Short: "List schema bundles on a table",
		Args: cobra.NoArgs, RunE: runBTSBList,
	}
	bigtableSBUpdateCmd = &cobra.Command{
		Use: "update SCHEMA_BUNDLE", Short: "Update a schema bundle",
		Args: cobra.ExactArgs(1), RunE: runBTSBUpdate,
	}
	bigtableSBGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SCHEMA_BUNDLE", Short: "Get the IAM policy for a schema bundle",
		Args: cobra.ExactArgs(1), RunE: runBTSBGetIam,
	}
	bigtableSBSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SCHEMA_BUNDLE POLICY_FILE", Short: "Set the IAM policy for a schema bundle",
		Args: cobra.ExactArgs(2), RunE: runBTSBSetIam,
	}
	bigtableSBAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SCHEMA_BUNDLE", Short: "Add an IAM policy binding to a schema bundle",
		Args: cobra.ExactArgs(1), RunE: runBTSBAddIam,
	}
	bigtableSBRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SCHEMA_BUNDLE", Short: "Remove an IAM policy binding from a schema bundle",
		Args: cobra.ExactArgs(1), RunE: runBTSBRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		bigtableSBCreateCmd, bigtableSBDeleteCmd, bigtableSBDescribeCmd, bigtableSBListCmd, bigtableSBUpdateCmd,
		bigtableSBGetIamCmd, bigtableSBSetIamCmd, bigtableSBAddIamCmd, bigtableSBRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBTSBInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagBTSBTable, "table", "", "Bigtable table ID (required)")
		_ = c.MarkFlagRequired("table")
		c.Flags().StringVar(&flagBTSBFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableSBCreateCmd, bigtableSBUpdateCmd} {
		c.Flags().StringVar(&flagBTSBConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the SchemaBundle body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	bigtableSBUpdateCmd.Flags().StringVar(&flagBTSBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	bigtableSBListCmd.Flags().Int64Var(&flagBTSBPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{bigtableSBAddIamCmd, bigtableSBRemoveIamCmd} {
		c.Flags().StringVar(&flagBTSBIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagBTSBIamRole, "role", "", "IAM role to bind (required)")
		c.Flags().StringVar(&flagBTSBIamCondExpr, "condition-expression", "", "CEL condition expression")
		c.Flags().StringVar(&flagBTSBIamCondT, "condition-title", "", "Condition title")
		c.Flags().StringVar(&flagBTSBIamCondD, "condition-description", "", "Condition description")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	bigtableSBRemoveIamCmd.Flags().BoolVar(&flagBTSBIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	bigtableSBCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableSBCmd)
}

func btSBParent() (string, error) {
	table, err := bigtableTableName(flagBTSBInstance, flagBTSBTable)
	if err != nil {
		return "", err
	}
	return table, nil
}

func btSBName(id string) (string, error) {
	parent, err := btSBParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/schemaBundles/%s", parent, id), nil
}

func runBTSBCreate(cmd *cobra.Command, args []string) error {
	parent, err := btSBParent()
	if err != nil {
		return err
	}
	body := &bigtableadmin.SchemaBundle{}
	if err := loadYAMLOrJSONInto(flagBTSBConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.SchemaBundles.Create(parent, body).SchemaBundleId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating schema bundle: %w", err)
	}
	fmt.Printf("Create request issued for schema bundle [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBTSBFormat)
}

func runBTSBDelete(cmd *cobra.Command, args []string) error {
	name, err := btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Tables.SchemaBundles.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting schema bundle: %w", err)
	}
	fmt.Printf("Deleted schema bundle [%s].\n", args[0])
	return nil
}

func runBTSBDescribe(cmd *cobra.Command, args []string) error {
	name, err := btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Tables.SchemaBundles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing schema bundle: %w", err)
	}
	return emitFormatted(got, flagBTSBFormat)
}

func runBTSBList(cmd *cobra.Command, args []string) error {
	parent, err := btSBParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.SchemaBundle
	pageToken := ""
	for {
		call := svc.Projects.Instances.Tables.SchemaBundles.List(parent).Context(ctx)
		if flagBTSBPageSize > 0 {
			call = call.PageSize(flagBTSBPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing schema bundles: %w", err)
		}
		all = append(all, resp.SchemaBundles...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBTSBFormat)
}

func runBTSBUpdate(cmd *cobra.Command, args []string) error {
	name, err := btSBName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.SchemaBundle{}
	if err := loadYAMLOrJSONInto(flagBTSBConfigFile, body); err != nil {
		return err
	}
	mask := flagBTSBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Tables.SchemaBundles.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating schema bundle: %w", err)
	}
	fmt.Printf("Update request issued for schema bundle [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBTSBFormat)
}

func runBTSBGetIam(cmd *cobra.Command, args []string) error {
	name, err := btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.SchemaBundles.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagBTSBFormat)
}

func runBTSBSetIam(cmd *cobra.Command, args []string) error {
	name, err := btSBName(args[0])
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
	updated, err := svc.Projects.Instances.Tables.SchemaBundles.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for schema bundle [%s].\n", args[0])
	return emitFormatted(updated, flagBTSBFormat)
}

func runBTSBAddIam(cmd *cobra.Command, args []string) error {
	name, err := btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.SchemaBundles.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	// Reuse bigtable IAM helpers from bigtable_tables.go.
	bigtableAddBindingToPolicy(policy, flagBTSBIamRole, flagBTSBIamMember, &bigtableadmin.Expr{
		Expression: flagBTSBIamCondExpr, Title: flagBTSBIamCondT, Description: flagBTSBIamCondD,
	})
	policy.Version = 3
	updated, err := svc.Projects.Instances.Tables.SchemaBundles.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for schema bundle [%s].\n", args[0])
	return emitFormatted(updated, flagBTSBFormat)
}

func runBTSBRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.SchemaBundles.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := &bigtableadmin.Expr{
		Expression: flagBTSBIamCondExpr, Title: flagBTSBIamCondT, Description: flagBTSBIamCondD,
	}
	if flagBTSBIamCondExpr == "" && flagBTSBIamCondT == "" && flagBTSBIamCondD == "" {
		cond = nil
	}
	if !bigtableRemoveBindingFromPolicy(policy, flagBTSBIamRole, flagBTSBIamMember, cond, flagBTSBIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagBTSBIamRole, flagBTSBIamMember)
	}
	updated, err := svc.Projects.Instances.Tables.SchemaBundles.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for schema bundle [%s].\n", args[0])
	return emitFormatted(updated, flagBTSBFormat)
}
