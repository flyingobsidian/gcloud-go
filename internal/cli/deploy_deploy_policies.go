package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy deploy-policies (#1528) ---

func (s *deployCmds) initDeployDeployPolicies() {
	all := []*cobra.Command{
		s.deployDPolDeleteCmd, s.deployDPolDescribeCmd, s.deployDPolExportCmd,
		s.deployDPolGetIamCmd, s.deployDPolSetIamCmd, s.deployDPolAddIamCmd, s.deployDPolRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployDPolRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployDPolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.deployDPolAddIamCmd, s.deployDPolRemoveIamCmd} {
		deployIamMemberFlags(c, &s.flagDeployDPolIamMember, &s.flagDeployDPolIamRole,
			&s.flagDeployDPolIamCondExpr, &s.flagDeployDPolIamCondT, &s.flagDeployDPolIamCondD)
	}
	s.deployDPolRemoveIamCmd.Flags().BoolVar(&s.flagDeployDPolIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.deployDPolCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployDPolCmd)
}

func (s *deployCmds) deployDPolName(id string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return deployChild("deployPolicies", id, deployLocationParent(project, s.flagDeployDPolRegion)), nil
}

func (s *deployCmds) runDeployDPolDelete(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DeployPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting deploy policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for deploy policy [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDeployDPolFormat)
}

func (s *deployCmds) runDeployDPolDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeployPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing deploy policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployDPolFormat)
}

func (s *deployCmds) runDeployDPolExport(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeployPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting deploy policy: %w", err)
	}
	format := s.flagDeployDPolFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, got, format)
}

func (s *deployCmds) runDeployDPolGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeployPolicies.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDeployDPolFormat)
}

func (s *deployCmds) runDeployDPolSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPolName(args[0])
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
	updated, err := svc.Projects.Locations.DeployPolicies.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("deploy policy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployDPolFormat)
}

func (s *deployCmds) runDeployDPolAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeployPolicies.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	deployAddBinding(policy, s.flagDeployDPolIamRole, s.flagDeployDPolIamMember,
		deployBuildCondition(s.flagDeployDPolIamCondExpr, s.flagDeployDPolIamCondT, s.flagDeployDPolIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.DeployPolicies.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("deploy policy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployDPolFormat)
}

func (s *deployCmds) runDeployDPolRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeployPolicies.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !deployRemoveBinding(policy, s.flagDeployDPolIamRole, s.flagDeployDPolIamMember,
		deployBuildCondition(s.flagDeployDPolIamCondExpr, s.flagDeployDPolIamCondT, s.flagDeployDPolIamCondD), s.flagDeployDPolIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDeployDPolIamRole, s.flagDeployDPolIamMember)
	}
	updated, err := svc.Projects.Locations.DeployPolicies.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("deploy policy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployDPolFormat)
}
