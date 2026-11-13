package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable tables (#1490) ---

func (s *bigtableCmds) initBigtableTables() {
	all := []*cobra.Command{
		s.bigtableTablesCreateCmd, s.bigtableTablesDeleteCmd, s.bigtableTablesDescribeCmd,
		s.bigtableTablesListCmd, s.bigtableTablesRestoreCmd, s.bigtableTablesUndeleteCmd,
		s.bigtableTablesUpdateCmd, s.bigtableTablesGetIamCmd, s.bigtableTablesSetIamCmd,
		s.bigtableTablesAddIamCmd, s.bigtableTablesRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBTTblInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagBTTblFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableTablesCreateCmd, s.bigtableTablesUpdateCmd} {
		c.Flags().StringVar(&s.flagBTTblConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Table body")
	}
	s.bigtableTablesUpdateCmd.Flags().StringVar(&s.flagBTTblUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.bigtableTablesUpdateCmd.Flags().BoolVar(&s.flagBTTblIgnoreWarn, "ignore-warnings", false,
		"If true, ignore safety checks when updating the table.")
	s.bigtableTablesListCmd.Flags().StringVar(&s.flagBTTblFilter, "filter", "", "Client-side filter")
	s.bigtableTablesListCmd.Flags().Int64Var(&s.flagBTTblPageSize, "page-size", 0, "Maximum number of results per page")
	s.bigtableTablesListCmd.Flags().StringVar(&s.flagBTTblView, "view", "", "Table view: NAME_ONLY | SCHEMA_VIEW | REPLICATION_VIEW | ENCRYPTION_VIEW | FULL")
	s.bigtableTablesDescribeCmd.Flags().StringVar(&s.flagBTTblView, "view", "", "Table view (see list --help)")
	s.bigtableTablesRestoreCmd.Flags().StringVar(&s.flagBTTblBackup, "backup", "",
		"Backup resource name (projects/.../backups/BACKUP) to restore from (required)")
	_ = s.bigtableTablesRestoreCmd.MarkFlagRequired("backup")

	for _, c := range []*cobra.Command{s.bigtableTablesAddIamCmd, s.bigtableTablesRemoveIamCmd} {
		c.Flags().StringVar(&s.flagBTTblIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&s.flagBTTblIamRole, "role", "", "IAM role to bind (e.g. roles/bigtable.reader) (required)")
		c.Flags().StringVar(&s.flagBTTblIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&s.flagBTTblIamCondT, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&s.flagBTTblIamCondD, "condition-description", "", "Description for a conditional binding")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	s.bigtableTablesRemoveIamCmd.Flags().BoolVar(&s.flagBTTblIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.bigtableTablesCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableTablesCmd)
}

func (s *bigtableCmds) bigtableInstanceParent(instance string) (string, error) {
	if strings.HasPrefix(instance, "projects/") {
		return instance, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/instances/%s", project, instance), nil
}

func (s *bigtableCmds) bigtableTableName(instance, table string) (string, error) {
	if strings.HasPrefix(table, "projects/") {
		return table, nil
	}
	parent, err := s.bigtableInstanceParent(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/tables/%s", parent, table), nil
}

func (s *bigtableCmds) runBTTblCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.bigtableInstanceParent(s.flagBTTblInstance)
	if err != nil {
		return err
	}
	req := &bigtableadmin.CreateTableRequest{TableId: args[0]}
	if s.flagBTTblConfigFile != "" {
		body := &bigtableadmin.Table{}
		if err := loadYAMLOrJSONInto(s.flagBTTblConfigFile, body); err != nil {
			return err
		}
		req.Table = body
	} else {
		req.Table = &bigtableadmin.Table{}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Tables.Create(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating table: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created table [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBTTblFormat)
}

func (s *bigtableCmds) runBTTblDelete(cmd *cobra.Command, args []string) error {
	name, err := s.bigtableTableName(s.flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Tables.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting table: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted table [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBTTblDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.bigtableTableName(s.flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Tables.Get(name).Context(ctx)
	if s.flagBTTblView != "" {
		call = call.View(s.flagBTTblView)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing table: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBTTblFormat)
}

func (s *bigtableCmds) runBTTblList(cmd *cobra.Command, args []string) error {
	parent, err := s.bigtableInstanceParent(s.flagBTTblInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.Table
	pageToken := ""
	for {
		call := svc.Projects.Instances.Tables.List(parent).Context(ctx)
		if s.flagBTTblPageSize > 0 {
			call = call.PageSize(s.flagBTTblPageSize)
		}
		if s.flagBTTblView != "" {
			call = call.View(s.flagBTTblView)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tables: %w", err)
		}
		for _, t := range resp.Tables {
			if s.flagBTTblFilter != "" && !strings.Contains(t.Name, s.flagBTTblFilter) {
				continue
			}
			all = append(all, t)
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBTTblFormat)
}

func (s *bigtableCmds) runBTTblRestore(cmd *cobra.Command, args []string) error {
	parent, err := s.bigtableInstanceParent(s.flagBTTblInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	req := &bigtableadmin.RestoreTableRequest{
		TableId: args[0],
		Backup:  s.flagBTTblBackup,
	}
	if err := s.applyBTRestoreTableProperties(cmd, req); err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.Restore(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring table: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restore request issued for table [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBTTblFormat)
}

func (s *bigtableCmds) runBTTblUndelete(cmd *cobra.Command, args []string) error {
	name, err := s.bigtableTableName(s.flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.Undelete(name, &bigtableadmin.UndeleteTableRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting table: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Undelete request issued for table [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBTTblFormat)
}

func (s *bigtableCmds) runBTTblUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.bigtableTableName(s.flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	if s.flagBTTblConfigFile == "" {
		return fmt.Errorf("--config-file is required for update")
	}
	body := &bigtableadmin.Table{}
	if err := loadYAMLOrJSONInto(s.flagBTTblConfigFile, body); err != nil {
		return err
	}
	mask := s.flagBTTblUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Tables.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	if s.flagBTTblIgnoreWarn {
		call = call.IgnoreWarnings(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating table: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for table [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBTTblFormat)
}

func (s *bigtableCmds) runBTTblGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.bigtableTableName(s.flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagBTTblFormat)
}

func (s *bigtableCmds) runBTTblSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.bigtableTableName(s.flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	policy := &bigtableadmin.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Instances.Tables.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for table [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBTTblFormat)
}

func (s *bigtableCmds) bigtableBuildCondition() *bigtableadmin.Expr {
	if s.flagBTTblIamCondExpr == "" && s.flagBTTblIamCondT == "" && s.flagBTTblIamCondD == "" {
		return nil
	}
	return &bigtableadmin.Expr{
		Expression:  s.flagBTTblIamCondExpr,
		Title:       s.flagBTTblIamCondT,
		Description: s.flagBTTblIamCondD,
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

func (s *bigtableCmds) runBTTblAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.bigtableTableName(s.flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	bigtableAddBindingToPolicy(policy, s.flagBTTblIamRole, s.flagBTTblIamMember, s.bigtableBuildCondition())
	policy.Version = 3
	updated, err := svc.Projects.Instances.Tables.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for table [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBTTblFormat)
}

func (s *bigtableCmds) runBTTblRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.bigtableTableName(s.flagBTTblInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !bigtableRemoveBindingFromPolicy(policy, s.flagBTTblIamRole, s.flagBTTblIamMember, s.bigtableBuildCondition(), s.flagBTTblIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagBTTblIamRole, s.flagBTTblIamMember)
	}
	updated, err := svc.Projects.Instances.Tables.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for table [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBTTblFormat)
}
