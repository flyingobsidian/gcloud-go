package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable tables (#1490) ---

var bigtableTablesCmd = &cobra.Command{Use: "tables", Short: "Manage Cloud Bigtable tables"}

var (
	flagBTTblInstance    string
	flagBTTblFormat      string
	flagBTTblConfigFile  string
	flagBTTblUpdateMask  string
	flagBTTblFilter      string
	flagBTTblPageSize    int64
	flagBTTblView        string
	flagBTTblBackup      string
	flagBTTblIamMember   string
	flagBTTblIamRole     string
	flagBTTblIamCondExpr string
	flagBTTblIamCondT    string
	flagBTTblIamCondD    string
	flagBTTblIamAllCond  bool
)

var (
	bigtableTablesCreateCmd = &cobra.Command{
		Use: "create TABLE", Short: "Create a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: runBTTblCreate,
	}
	bigtableTablesDeleteCmd = &cobra.Command{
		Use: "delete TABLE", Short: "Delete a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: runBTTblDelete,
	}
	bigtableTablesDescribeCmd = &cobra.Command{
		Use: "describe TABLE", Short: "Describe a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: runBTTblDescribe,
	}
	bigtableTablesListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable tables in an instance",
		Args: cobra.NoArgs, RunE: runBTTblList,
	}
	bigtableTablesRestoreCmd = &cobra.Command{
		Use: "restore TABLE", Short: "Restore a Bigtable table from a backup",
		Args: cobra.ExactArgs(1), RunE: runBTTblRestore,
	}
	bigtableTablesUndeleteCmd = &cobra.Command{
		Use: "undelete TABLE", Short: "Undelete a soft-deleted Bigtable table",
		Args: cobra.ExactArgs(1), RunE: runBTTblUndelete,
	}
	bigtableTablesUpdateCmd = &cobra.Command{
		Use: "update TABLE", Short: "Update a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: runBTTblUpdate,
	}
	bigtableTablesGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TABLE", Short: "Get the IAM policy for a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: runBTTblGetIam,
	}
	bigtableTablesSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TABLE POLICY_FILE", Short: "Set the IAM policy for a Bigtable table",
		Args: cobra.ExactArgs(2), RunE: runBTTblSetIam,
	}
	bigtableTablesAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TABLE", Short: "Add an IAM policy binding to a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: runBTTblAddIam,
	}
	bigtableTablesRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TABLE", Short: "Remove an IAM policy binding from a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: runBTTblRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		bigtableTablesCreateCmd, bigtableTablesDeleteCmd, bigtableTablesDescribeCmd,
		bigtableTablesListCmd, bigtableTablesRestoreCmd, bigtableTablesUndeleteCmd,
		bigtableTablesUpdateCmd, bigtableTablesGetIamCmd, bigtableTablesSetIamCmd,
		bigtableTablesAddIamCmd, bigtableTablesRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBTTblInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagBTTblFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableTablesCreateCmd, bigtableTablesUpdateCmd} {
		c.Flags().StringVar(&flagBTTblConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Table body")
	}
	bigtableTablesUpdateCmd.Flags().StringVar(&flagBTTblUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	bigtableTablesListCmd.Flags().StringVar(&flagBTTblFilter, "filter", "", "Client-side filter")
	bigtableTablesListCmd.Flags().Int64Var(&flagBTTblPageSize, "page-size", 0, "Maximum number of results per page")
	bigtableTablesListCmd.Flags().StringVar(&flagBTTblView, "view", "", "Table view: NAME_ONLY | SCHEMA_VIEW | REPLICATION_VIEW | ENCRYPTION_VIEW | FULL")
	bigtableTablesDescribeCmd.Flags().StringVar(&flagBTTblView, "view", "", "Table view (see list --help)")
	bigtableTablesRestoreCmd.Flags().StringVar(&flagBTTblBackup, "backup", "",
		"Backup resource name (projects/.../backups/BACKUP) to restore from (required)")
	_ = bigtableTablesRestoreCmd.MarkFlagRequired("backup")

	for _, c := range []*cobra.Command{bigtableTablesAddIamCmd, bigtableTablesRemoveIamCmd} {
		c.Flags().StringVar(&flagBTTblIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&flagBTTblIamRole, "role", "", "IAM role to bind (e.g. roles/bigtable.reader) (required)")
		c.Flags().StringVar(&flagBTTblIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&flagBTTblIamCondT, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&flagBTTblIamCondD, "condition-description", "", "Description for a conditional binding")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	bigtableTablesRemoveIamCmd.Flags().BoolVar(&flagBTTblIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	bigtableTablesCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableTablesCmd)
}

func bigtableInstanceParent(instance string) (string, error) {
	if strings.HasPrefix(instance, "projects/") {
		return instance, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/instances/%s", project, instance), nil
}

func bigtableTableName(instance, table string) (string, error) {
	if strings.HasPrefix(table, "projects/") {
		return table, nil
	}
	parent, err := bigtableInstanceParent(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/tables/%s", parent, table), nil
}

func runBTTblCreate(cmd *cobra.Command, args []string) error {
	parent, err := bigtableInstanceParent(flagBTTblInstance)
	if err != nil {
		return err
	}
	req := &bigtableadmin.CreateTableRequest{TableId: args[0]}
	if flagBTTblConfigFile != "" {
		body := &bigtableadmin.Table{}
		if err := loadYAMLOrJSONInto(flagBTTblConfigFile, body); err != nil {
			return err
		}
		req.Table = body
	} else {
		req.Table = &bigtableadmin.Table{}
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Tables.Create(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating table: %w", err)
	}
	fmt.Printf("Created table [%s].\n", args[0])
	return emitFormatted(got, flagBTTblFormat)
}

func runBTTblDelete(cmd *cobra.Command, args []string) error {
	name, err := bigtableTableName(flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Tables.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting table: %w", err)
	}
	fmt.Printf("Deleted table [%s].\n", args[0])
	return nil
}

func runBTTblDescribe(cmd *cobra.Command, args []string) error {
	name, err := bigtableTableName(flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Tables.Get(name).Context(ctx)
	if flagBTTblView != "" {
		call = call.View(flagBTTblView)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing table: %w", err)
	}
	return emitFormatted(got, flagBTTblFormat)
}

func runBTTblList(cmd *cobra.Command, args []string) error {
	parent, err := bigtableInstanceParent(flagBTTblInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.Table
	pageToken := ""
	for {
		call := svc.Projects.Instances.Tables.List(parent).Context(ctx)
		if flagBTTblPageSize > 0 {
			call = call.PageSize(flagBTTblPageSize)
		}
		if flagBTTblView != "" {
			call = call.View(flagBTTblView)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tables: %w", err)
		}
		for _, t := range resp.Tables {
			if flagBTTblFilter != "" && !strings.Contains(t.Name, flagBTTblFilter) {
				continue
			}
			all = append(all, t)
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBTTblFormat)
}

func runBTTblRestore(cmd *cobra.Command, args []string) error {
	parent, err := bigtableInstanceParent(flagBTTblInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.Restore(parent, &bigtableadmin.RestoreTableRequest{
		TableId: args[0],
		Backup:  flagBTTblBackup,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring table: %w", err)
	}
	fmt.Printf("Restore request issued for table [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBTTblFormat)
}

func runBTTblUndelete(cmd *cobra.Command, args []string) error {
	name, err := bigtableTableName(flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.Undelete(name, &bigtableadmin.UndeleteTableRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting table: %w", err)
	}
	fmt.Printf("Undelete request issued for table [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBTTblFormat)
}

func runBTTblUpdate(cmd *cobra.Command, args []string) error {
	name, err := bigtableTableName(flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	if flagBTTblConfigFile == "" {
		return fmt.Errorf("--config-file is required for update")
	}
	body := &bigtableadmin.Table{}
	if err := loadYAMLOrJSONInto(flagBTTblConfigFile, body); err != nil {
		return err
	}
	mask := flagBTTblUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Tables.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating table: %w", err)
	}
	fmt.Printf("Update request issued for table [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBTTblFormat)
}

func runBTTblGetIam(cmd *cobra.Command, args []string) error {
	name, err := bigtableTableName(flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagBTTblFormat)
}

func runBTTblSetIam(cmd *cobra.Command, args []string) error {
	name, err := bigtableTableName(flagBTTblInstance, args[0])
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
	updated, err := svc.Projects.Instances.Tables.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for table [%s].\n", args[0])
	return emitFormatted(updated, flagBTTblFormat)
}

func bigtableBuildCondition() *bigtableadmin.Expr {
	if flagBTTblIamCondExpr == "" && flagBTTblIamCondT == "" && flagBTTblIamCondD == "" {
		return nil
	}
	return &bigtableadmin.Expr{
		Expression:  flagBTTblIamCondExpr,
		Title:       flagBTTblIamCondT,
		Description: flagBTTblIamCondD,
	}
}

func bigtableConditionsEqual(a, b *bigtableadmin.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func bigtableAddBindingToPolicy(policy *bigtableadmin.Policy, role, member string, condition *bigtableadmin.Expr) bool {
	for _, b := range policy.Bindings {
		if b.Role != role || !bigtableConditionsEqual(b.Condition, condition) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return false
			}
		}
		b.Members = append(b.Members, member)
		return true
	}
	policy.Bindings = append(policy.Bindings, &bigtableadmin.Binding{
		Role:      role,
		Members:   []string{member},
		Condition: condition,
	})
	return true
}

func bigtableRemoveBindingFromPolicy(policy *bigtableadmin.Policy, role, member string, condition *bigtableadmin.Expr, allConditions bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConditions || bigtableConditionsEqual(b.Condition, condition))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				changed = true
				continue
			}
			newMembers = append(newMembers, m)
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

func runBTTblAddIam(cmd *cobra.Command, args []string) error {
	name, err := bigtableTableName(flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	bigtableAddBindingToPolicy(policy, flagBTTblIamRole, flagBTTblIamMember, bigtableBuildCondition())
	policy.Version = 3
	updated, err := svc.Projects.Instances.Tables.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for table [%s].\n", args[0])
	return emitFormatted(updated, flagBTTblFormat)
}

func runBTTblRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := bigtableTableName(flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !bigtableRemoveBindingFromPolicy(policy, flagBTTblIamRole, flagBTTblIamMember, bigtableBuildCondition(), flagBTTblIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagBTTblIamRole, flagBTTblIamMember)
	}
	updated, err := svc.Projects.Instances.Tables.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for table [%s].\n", args[0])
	return emitFormatted(updated, flagBTTblFormat)
}
