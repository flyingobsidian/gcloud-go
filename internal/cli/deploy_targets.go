package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy targets (#1532) ---

func (s *deployCmds) initDeployTargets() {
	all := []*cobra.Command{
		s.deployTgtDeleteCmd, s.deployTgtDescribeCmd, s.deployTgtExportCmd, s.deployTgtListCmd,
		s.deployTgtRedeployCmd, s.deployTgtRollbackCmd, s.deployTgtGetIamCmd, s.deployTgtSetIamCmd,
		s.deployTgtAddIamCmd, s.deployTgtRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployTgtRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployTgtFormat, "format", "", "Output format")
	}
	s.deployTgtListCmd.Flags().Int64Var(&s.flagDeployTgtPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{s.deployTgtRedeployCmd, s.deployTgtRollbackCmd} {
		c.Flags().StringVar(&s.flagDeployTgtPipeline, "delivery-pipeline", "",
			"Delivery pipeline that owns the target (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&s.flagDeployTgtRelease, "release", "",
			"Release to redeploy/roll back to (required for rollback; defaults to the target's latest for redeploy)")
		c.Flags().StringVar(&s.flagDeployTgtRolloutID, "rollout-id", "",
			"Optional rollout ID (auto-generated when omitted)")
		c.Flags().StringVar(&s.flagDeployTgtDescription, "description", "", "Rollout description")
		c.Flags().StringVar(&s.flagDeployTgtStartPhase, "starting-phase-id", "", "Starting phase for the rollout")
		c.Flags().StringSliceVar(&s.flagDeployTgtOverridePol, "override-deploy-policies", nil,
			"Deploy policies to override (repeatable)")
	}
	_ = s.deployTgtRollbackCmd.MarkFlagRequired("release")

	for _, c := range []*cobra.Command{s.deployTgtAddIamCmd, s.deployTgtRemoveIamCmd} {
		deployIamMemberFlags(c, &s.flagDeployTgtIamMember, &s.flagDeployTgtIamRole,
			&s.flagDeployTgtIamCondExpr, &s.flagDeployTgtIamCondT, &s.flagDeployTgtIamCondD)
	}
	s.deployTgtRemoveIamCmd.Flags().BoolVar(&s.flagDeployTgtIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.deployTargetsCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployTargetsCmd)
}

func (s *deployCmds) deployTgtName(id string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return deployChild("targets", id, deployLocationParent(project, s.flagDeployTgtRegion)), nil
}

func (s *deployCmds) runDeployTgtDelete(cmd *cobra.Command, args []string) error {
	name, err := s.deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Targets.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting target: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for target [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDeployTgtFormat)
}

func (s *deployCmds) runDeployTgtDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Targets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing target: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployTgtFormat)
}

func (s *deployCmds) runDeployTgtExport(cmd *cobra.Command, args []string) error {
	name, err := s.deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Targets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting target: %w", err)
	}
	format := s.flagDeployTgtFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, got, format)
}

func (s *deployCmds) runDeployTgtList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	parent := deployLocationParent(project, s.flagDeployTgtRegion)
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	var all []*clouddeploy.Target
	pageToken := ""
	for {
		call := svc.Projects.Locations.Targets.List(parent).Context(ctx)
		if s.flagDeployTgtPageSize > 0 {
			call = call.PageSize(s.flagDeployTgtPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing targets: %w", err)
		}
		all = append(all, resp.Targets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDeployTgtFormat)
}

func (s *deployCmds) runDeployTgtRollback(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	parent := deployLocationParent(project, s.flagDeployTgtRegion)
	pipeline := deployChild("deliveryPipelines", s.flagDeployTgtPipeline, parent)
	req := &clouddeploy.RollbackTargetRequest{
		TargetId:             args[0],
		ReleaseId:            s.flagDeployTgtRelease,
		RolloutId:            s.flagDeployTgtRolloutID,
		OverrideDeployPolicy: s.flagDeployTgtOverridePol,
	}
	if s.flagDeployTgtDescription != "" || s.flagDeployTgtStartPhase != "" {
		req.RollbackConfig = &clouddeploy.RollbackTargetConfig{
			Rollout: &clouddeploy.Rollout{
				Description: s.flagDeployTgtDescription,
			},
			StartingPhaseId: s.flagDeployTgtStartPhase,
		}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DeliveryPipelines.RollbackTarget(pipeline, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back target: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Rollback issued for target [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, resp, s.flagDeployTgtFormat)
}

func (s *deployCmds) runDeployTgtRedeploy(cmd *cobra.Command, args []string) error {
	// Redeploy: create a new Rollout on the target's most-recent (or --release'd) release.
	if s.flagDeployTgtRelease == "" {
		return fmt.Errorf("--release is required (Cloud Deploy has no server-side 'latest' resolver here)")
	}
	if s.flagDeployTgtRolloutID == "" {
		return fmt.Errorf("--rollout-id is required (auto-generation not supported)")
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	parent := deployLocationParent(project, s.flagDeployTgtRegion)
	release := deployChild("releases", s.flagDeployTgtRelease,
		deployChild("deliveryPipelines", s.flagDeployTgtPipeline, parent))
	rollout := &clouddeploy.Rollout{
		TargetId:    args[0],
		Description: s.flagDeployTgtDescription,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Create(release, rollout).RolloutId(s.flagDeployTgtRolloutID).Context(ctx)
	if s.flagDeployTgtStartPhase != "" {
		call = call.StartingPhaseId(s.flagDeployTgtStartPhase)
	}
	_ = s.flagDeployTgtOverridePol
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("redeploying: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Redeploy issued for target [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDeployTgtFormat)
}

func (s *deployCmds) runDeployTgtGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Targets.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDeployTgtFormat)
}

func (s *deployCmds) runDeployTgtSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployTgtName(args[0])
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
	updated, err := svc.Projects.Locations.Targets.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("target [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployTgtFormat)
}

func (s *deployCmds) runDeployTgtAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Targets.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	deployAddBinding(policy, s.flagDeployTgtIamRole, s.flagDeployTgtIamMember,
		deployBuildCondition(s.flagDeployTgtIamCondExpr, s.flagDeployTgtIamCondT, s.flagDeployTgtIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Targets.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("target [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployTgtFormat)
}

func (s *deployCmds) runDeployTgtRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.deployTgtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Targets.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !deployRemoveBinding(policy, s.flagDeployTgtIamRole, s.flagDeployTgtIamMember,
		deployBuildCondition(s.flagDeployTgtIamCondExpr, s.flagDeployTgtIamCondT, s.flagDeployTgtIamCondD), s.flagDeployTgtIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDeployTgtIamRole, s.flagDeployTgtIamMember)
	}
	updated, err := svc.Projects.Locations.Targets.SetIamPolicy(name, &clouddeploy.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.deployUpdatedIam(fmt.Sprintf("target [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDeployTgtFormat)
}
