package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy custom-target-types (#1526) ---

func (s *deployCmds) initDeployCustomTargetTypes() {
	all := []*cobra.Command{
		s.deployCTTDeleteCmd, s.deployCTTDescribeCmd, s.deployCTTExportCmd, s.deployCTTListCmd,
		s.deployCTTGetIamCmd, s.deployCTTSetIamCmd, s.deployCTTAddIamCmd, s.deployCTTRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployCTTRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployCTTFormat, "format", "", "Output format")
	}
	s.deployCTTListCmd.Flags().Int64Var(&s.flagDeployCTTPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{s.deployCTTAddIamCmd, s.deployCTTRemoveIamCmd} {
		deployIamMemberFlags(c, &s.flagDeployCTTIamMember, &s.flagDeployCTTIamRole,
			&s.flagDeployCTTIamCondExpr, &s.flagDeployCTTIamCondT, &s.flagDeployCTTIamCondD)
	}
	s.deployCTTRemoveIamCmd.Flags().BoolVar(&s.flagDeployCTTIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.deployCTTCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployCTTCmd)
}

func (s *deployCmds) deployCTTParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return deployLocationParent(project, s.flagDeployCTTRegion), nil
}

func (s *deployCmds) deployCTTName(id string) (string, error) {
	parent, err := s.deployCTTParent()
	if err != nil {
		return "", err
	}
	return deployChild("customTargetTypes", id, parent), nil
}

func (s *deployCmds) runDeployCTTDelete(cmd *cobra.Command, args []string) error {
	name, err := s.deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CustomTargetTypes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting custom target type: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for custom target type [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDeployCTTFormat)
}

func (s *deployCmds) runDeployCTTDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CustomTargetTypes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing custom target type: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployCTTFormat)
}

func (s *deployCmds) runDeployCTTExport(cmd *cobra.Command, args []string) error {
	name, err := s.deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CustomTargetTypes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting custom target type: %w", err)
	}
	format := s.flagDeployCTTFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, got, format)
}

func (s *deployCmds) runDeployCTTList(cmd *cobra.Command, args []string) error {
	parent, err := s.deployCTTParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	var all []*clouddeploy.CustomTargetType
	pageToken := ""
	for {
		call := svc.Projects.Locations.CustomTargetTypes.List(parent).Context(ctx)
		if s.flagDeployCTTPageSize > 0 {
			call = call.PageSize(s.flagDeployCTTPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDeployCTTFormat)
}

func (s *deployCmds) runDeployCTTGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CustomTargetTypes.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDeployCTTFormat)
}

func (s *deployCmds) runDeployCTTSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployCTTName(args[0])
	if err != nil {
		return err
	}
	policy := &clouddeploy.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.CustomTargetTypes.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("custom target type [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployCTTFormat)
}

func (s *deployCmds) runDeployCTTAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CustomTargetTypes.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	deployAddBinding(policy, s.flagDeployCTTIamRole, s.flagDeployCTTIamMember,
		deployBuildCondition(s.flagDeployCTTIamCondExpr, s.flagDeployCTTIamCondT, s.flagDeployCTTIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.CustomTargetTypes.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("custom target type [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployCTTFormat)
}

func (s *deployCmds) runDeployCTTRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployCTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CustomTargetTypes.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !deployRemoveBinding(policy, s.flagDeployCTTIamRole, s.flagDeployCTTIamMember,
		deployBuildCondition(s.flagDeployCTTIamCondExpr, s.flagDeployCTTIamCondT, s.flagDeployCTTIamCondD), s.flagDeployCTTIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDeployCTTIamRole, s.flagDeployCTTIamMember)
	}
	updated, err := svc.Projects.Locations.CustomTargetTypes.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("custom target type [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployCTTFormat)
}
