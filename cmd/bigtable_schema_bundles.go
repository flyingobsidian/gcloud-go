package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable schema-bundles (#1489) ---

func (s *bigtableCmds) initBigtableSchemaBundles() {
	all := []*cobra.Command{
		s.bigtableSBCreateCmd, s.bigtableSBDeleteCmd, s.bigtableSBDescribeCmd, s.bigtableSBListCmd, s.bigtableSBUpdateCmd,
		s.bigtableSBGetIamCmd, s.bigtableSBSetIamCmd, s.bigtableSBAddIamCmd, s.bigtableSBRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBTSBInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagBTSBTable, "table", "", "Bigtable table ID (required)")
		_ = c.MarkFlagRequired("table")
		c.Flags().StringVar(&s.flagBTSBFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableSBCreateCmd, s.bigtableSBUpdateCmd} {
		c.Flags().StringVar(&s.flagBTSBConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the SchemaBundle body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bigtableSBUpdateCmd.Flags().StringVar(&s.flagBTSBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.bigtableSBListCmd.Flags().Int64Var(&s.flagBTSBPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{s.bigtableSBAddIamCmd, s.bigtableSBRemoveIamCmd} {
		c.Flags().StringVar(&s.flagBTSBIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&s.flagBTSBIamRole, "role", "", "IAM role to bind (required)")
		c.Flags().StringVar(&s.flagBTSBIamCondExpr, "condition-expression", "", "CEL condition expression")
		c.Flags().StringVar(&s.flagBTSBIamCondT, "condition-title", "", "Condition title")
		c.Flags().StringVar(&s.flagBTSBIamCondD, "condition-description", "", "Condition description")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	s.bigtableSBRemoveIamCmd.Flags().BoolVar(&s.flagBTSBIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.bigtableSBCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableSBCmd)
}

func (s *bigtableCmds) btSBParent() (string, error) {
	table, err := s.bigtableTableName(s.flagBTSBInstance, s.flagBTSBTable)
	if err != nil {
		return "", err
	}
	return table, nil
}

func (s *bigtableCmds) btSBName(id string) (string, error) {
	parent, err := s.btSBParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/schemaBundles/%s", parent, id), nil
}

func (s *bigtableCmds) runBTSBCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.btSBParent()
	if err != nil {
		return err
	}
	body := &bigtableadmin.SchemaBundle{}
	if err := loadYAMLOrJSONInto(s.flagBTSBConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.SchemaBundles.Create(parent, body).SchemaBundleId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating schema bundle: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for schema bundle [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBTSBFormat)
}

func (s *bigtableCmds) runBTSBDelete(cmd *cobra.Command, args []string) error {
	name, err := s.btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Tables.SchemaBundles.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting schema bundle: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted schema bundle [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBTSBDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Tables.SchemaBundles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing schema bundle: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBTSBFormat)
}

func (s *bigtableCmds) runBTSBList(cmd *cobra.Command, args []string) error {
	parent, err := s.btSBParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.SchemaBundle
	pageToken := ""
	for {
		call := svc.Projects.Instances.Tables.SchemaBundles.List(parent).Context(ctx)
		if s.flagBTSBPageSize > 0 {
			call = call.PageSize(s.flagBTSBPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagBTSBFormat)
}

func (s *bigtableCmds) runBTSBUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.btSBName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.SchemaBundle{}
	if err := loadYAMLOrJSONInto(s.flagBTSBConfigFile, body); err != nil {
		return err
	}
	mask := s.flagBTSBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Update request issued for schema bundle [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBTSBFormat)
}

func (s *bigtableCmds) runBTSBGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Tables.SchemaBundles.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagBTSBFormat)
}

func (s *bigtableCmds) runBTSBSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.btSBName(args[0])
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
	updated, err := svc.Projects.Instances.Tables.SchemaBundles.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for schema bundle [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBTSBFormat)
}

func (s *bigtableCmds) runBTSBAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
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
	bigtableAddBindingToPolicy(policy, s.flagBTSBIamRole, s.flagBTSBIamMember, &bigtableadmin.Expr{
		Expression: s.flagBTSBIamCondExpr, Title: s.flagBTSBIamCondT, Description: s.flagBTSBIamCondD,
	})
	policy.Version = 3
	updated, err := svc.Projects.Instances.Tables.SchemaBundles.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for schema bundle [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBTSBFormat)
}

func (s *bigtableCmds) runBTSBRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.btSBName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
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
		Expression: s.flagBTSBIamCondExpr, Title: s.flagBTSBIamCondT, Description: s.flagBTSBIamCondD,
	}
	if s.flagBTSBIamCondExpr == "" && s.flagBTSBIamCondT == "" && s.flagBTSBIamCondD == "" {
		cond = nil
	}
	if !bigtableRemoveBindingFromPolicy(policy, s.flagBTSBIamRole, s.flagBTSBIamMember, cond, s.flagBTSBIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagBTSBIamRole, s.flagBTSBIamMember)
	}
	updated, err := svc.Projects.Instances.Tables.SchemaBundles.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for schema bundle [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBTSBFormat)
}
