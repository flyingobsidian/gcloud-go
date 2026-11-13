package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable instances (#1306) ---

func (s *bigtableCmds) initBigtableInstances() {
	all := []*cobra.Command{
		s.bigtableInstCreateCmd, s.bigtableInstDeleteCmd, s.bigtableInstDescribeCmd,
		s.bigtableInstListCmd, s.bigtableInstUpdateCmd,
		s.bigtableInstGetIamCmd, s.bigtableInstSetIamCmd, s.bigtableInstAddIamCmd, s.bigtableInstRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBtInstFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableInstCreateCmd, s.bigtableInstUpdateCmd} {
		c.Flags().StringVar(&s.flagBtInstConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bigtableInstCreateCmd.Flags().StringToStringVar(&s.flagBtInstTags, "tags", nil,
		"Comma-separated list of KEY=VALUE resource-manager tags to bind to the instance on creation "+
			"(e.g. --tags=123/environment=production,123/costCenter=marketing).")

	for _, c := range []*cobra.Command{s.bigtableInstAddIamCmd, s.bigtableInstRemoveIamCmd} {
		c.Flags().StringVar(&s.flagBtInstIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&s.flagBtInstIamRole, "role", "", "IAM role to bind (required)")
		c.Flags().StringVar(&s.flagBtInstIamCondE, "condition-expression", "", "CEL condition expression")
		c.Flags().StringVar(&s.flagBtInstIamCondT, "condition-title", "", "Condition title")
		c.Flags().StringVar(&s.flagBtInstIamCondD, "condition-description", "", "Condition description")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	s.bigtableInstRemoveIamCmd.Flags().BoolVar(&s.flagBtInstIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.bigtableInstancesCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableInstancesCmd)
}

func (s *bigtableCmds) btInstBuildCondition() *bigtableadmin.Expr {
	if s.flagBtInstIamCondE == "" && s.flagBtInstIamCondT == "" && s.flagBtInstIamCondD == "" {
		return nil
	}
	return &bigtableadmin.Expr{
		Expression:  s.flagBtInstIamCondE,
		Title:       s.flagBtInstIamCondT,
		Description: s.flagBtInstIamCondD,
	}
}

func (s *bigtableCmds) runBtInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.btProjectName()
	if err != nil {
		return err
	}
	req := &bigtableadmin.CreateInstanceRequest{}
	if err := loadYAMLOrJSONInto(s.flagBtInstConfigFile, req); err != nil {
		return err
	}
	if req.InstanceId == "" {
		req.InstanceId = args[0]
	}
	if len(s.flagBtInstTags) > 0 {
		if req.Instance == nil {
			req.Instance = &bigtableadmin.Instance{}
		}
		if req.Instance.Tags == nil {
			req.Instance.Tags = make(map[string]string, len(s.flagBtInstTags))
		}
		for k, v := range s.flagBtInstTags {
			req.Instance.Tags[k] = v
		}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Create(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtInstFormat)
}

func (s *bigtableCmds) runBtInstDelete(cmd *cobra.Command, args []string) error {
	name, err := s.btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted instance [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBtInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBtInstFormat)
}

func (s *bigtableCmds) runBtInstList(cmd *cobra.Command, args []string) error {
	parent, err := s.btProjectName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagBtInstFormat)
}

func (s *bigtableCmds) runBtInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.btInstanceName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.Instance{}
	if err := loadYAMLOrJSONInto(s.flagBtInstConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBtInstFormat)
}

func (s *bigtableCmds) runBtInstGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagBtInstFormat)
}

func (s *bigtableCmds) runBtInstSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.btInstanceName(args[0])
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
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBtInstFormat)
}

func (s *bigtableCmds) runBtInstAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	bigtableAddBindingToPolicy(policy, s.flagBtInstIamRole, s.flagBtInstIamMember, s.btInstBuildCondition())
	policy.Version = 3
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBtInstFormat)
}

func (s *bigtableCmds) runBtInstRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.btInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &bigtableadmin.GetIamPolicyRequest{
		Options: &bigtableadmin.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !bigtableRemoveBindingFromPolicy(policy, s.flagBtInstIamRole, s.flagBtInstIamMember, s.btInstBuildCondition(), s.flagBtInstIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagBtInstIamRole, s.flagBtInstIamMember)
	}
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &bigtableadmin.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagBtInstFormat)
}
