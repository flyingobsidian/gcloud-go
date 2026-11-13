package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy delivery-pipelines (#1527) ---

func (s *deployCmds) initDeployDeliveryPipelines() {
	all := []*cobra.Command{
		s.deployDPDeleteCmd, s.deployDPDescribeCmd, s.deployDPExportCmd, s.deployDPListCmd,
		s.deployDPGetIamCmd, s.deployDPSetIamCmd, s.deployDPAddIamCmd, s.deployDPRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployDPRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployDPFormat, "format", "", "Output format")
	}
	s.deployDPDeleteCmd.Flags().BoolVar(&s.flagDeployDPForce, "force", false, "Force delete of a pipeline with children")
	s.deployDPListCmd.Flags().Int64Var(&s.flagDeployDPPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{s.deployDPAddIamCmd, s.deployDPRemoveIamCmd} {
		deployIamMemberFlags(c, &s.flagDeployDPIamMember, &s.flagDeployDPIamRole,
			&s.flagDeployDPIamCondExpr, &s.flagDeployDPIamCondT, &s.flagDeployDPIamCondD)
	}
	s.deployDPRemoveIamCmd.Flags().BoolVar(&s.flagDeployDPIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.deployDPCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployDPCmd)
}

func (s *deployCmds) deployDPParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return deployLocationParent(project, s.flagDeployDPRegion), nil
}

func (s *deployCmds) deployDPName(id string) (string, error) {
	parent, err := s.deployDPParent()
	if err != nil {
		return "", err
	}
	return deployChild("deliveryPipelines", id, parent), nil
}

func (s *deployCmds) runDeployDPDelete(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DeliveryPipelines.Delete(name).Context(ctx)
	if s.flagDeployDPForce {
		call = call.Force(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting delivery pipeline: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for delivery pipeline [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDeployDPFormat)
}

func (s *deployCmds) runDeployDPDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing delivery pipeline: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployDPFormat)
}

func (s *deployCmds) runDeployDPExport(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting delivery pipeline: %w", err)
	}
	format := s.flagDeployDPFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, got, format)
}

func (s *deployCmds) runDeployDPList(cmd *cobra.Command, args []string) error {
	parent, err := s.deployDPParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	var all []*clouddeploy.DeliveryPipeline
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.List(parent).Context(ctx)
		if s.flagDeployDPPageSize > 0 {
			call = call.PageSize(s.flagDeployDPPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDeployDPFormat)
}

func (s *deployCmds) runDeployDPGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeliveryPipelines.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDeployDPFormat)
}

func (s *deployCmds) runDeployDPSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPName(args[0])
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
	updated, err := svc.Projects.Locations.DeliveryPipelines.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("delivery pipeline [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployDPFormat)
}

func (s *deployCmds) runDeployDPAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeliveryPipelines.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	deployAddBinding(policy, s.flagDeployDPIamRole, s.flagDeployDPIamMember,
		deployBuildCondition(s.flagDeployDPIamCondExpr, s.flagDeployDPIamCondT, s.flagDeployDPIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.DeliveryPipelines.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("delivery pipeline [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployDPFormat)
}

func (s *deployCmds) runDeployDPRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployDPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.DeliveryPipelines.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !deployRemoveBinding(policy, s.flagDeployDPIamRole, s.flagDeployDPIamMember,
		deployBuildCondition(s.flagDeployDPIamCondExpr, s.flagDeployDPIamCondT, s.flagDeployDPIamCondD), s.flagDeployDPIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDeployDPIamRole, s.flagDeployDPIamMember)
	}
	updated, err := svc.Projects.Locations.DeliveryPipelines.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("delivery pipeline [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployDPFormat)
}
