package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy releases (#1530) ---

func (s *deployCmds) initDeployReleases() {
	all := []*cobra.Command{
		s.deployRelAbandonCmd, s.deployRelCreateCmd, s.deployRelDescribeCmd, s.deployRelListCmd, s.deployRelPromoteCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployRelRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployRelPipeline, "delivery-pipeline", "",
			"Delivery pipeline that owns the release (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&s.flagDeployRelFormat, "format", "", "Output format")
	}
	s.deployRelCreateCmd.Flags().StringVar(&s.flagDeployRelConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Release body (required)")
	_ = s.deployRelCreateCmd.MarkFlagRequired("config-file")
	s.deployRelCreateCmd.Flags().StringSliceVar(&s.flagDeployRelOverride, "override-deploy-policies", nil,
		"Deploy policies to override (repeatable)")
	s.deployRelCreateCmd.Flags().BoolVar(&s.flagDeployRelValidateOnly, "validate-only", false,
		"Only validate the request; do not create the release")
	s.deployRelListCmd.Flags().Int64Var(&s.flagDeployRelPageSize, "page-size", 0, "Maximum results per page")
	s.deployRelPromoteCmd.Flags().StringVar(&s.flagDeployRelPromoteTgt, "to-target", "",
		"Target ID to promote the release to (required)")
	_ = s.deployRelPromoteCmd.MarkFlagRequired("to-target")
	s.deployRelPromoteCmd.Flags().StringVar(&s.flagDeployRelPromoteID, "rollout-id", "",
		"Rollout ID to assign to the new rollout (required)")
	_ = s.deployRelPromoteCmd.MarkFlagRequired("rollout-id")

	s.deployRelCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployRelCmd)
}

func (s *deployCmds) deployRelParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return deployChild("deliveryPipelines", s.flagDeployRelPipeline, deployLocationParent(project, s.flagDeployRelRegion)), nil
}

func (s *deployCmds) deployRelName(id string) (string, error) {
	parent, err := s.deployRelParent()
	if err != nil {
		return "", err
	}
	return deployChild("releases", id, parent), nil
}

func (s *deployCmds) runDeployRelAbandon(cmd *cobra.Command, args []string) error {
	name, err := s.deployRelName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Abandon(name, &clouddeploy.AbandonReleaseRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("abandoning release: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Abandoned release [%s].\n", args[0])
	return nil
}

func (s *deployCmds) runDeployRelCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.deployRelParent()
	if err != nil {
		return err
	}
	body := &clouddeploy.Release{}
	if err := loadYAMLOrJSONInto(s.flagDeployRelConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DeliveryPipelines.Releases.Create(parent, body).ReleaseId(args[0]).Context(ctx)
	if len(s.flagDeployRelOverride) > 0 {
		call = call.OverrideDeployPolicy(s.flagDeployRelOverride...)
	}
	if s.flagDeployRelValidateOnly {
		call = call.ValidateOnly(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating release: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for release [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDeployRelFormat)
}

func (s *deployCmds) runDeployRelDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployRelName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Releases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing release: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployRelFormat)
}

func (s *deployCmds) runDeployRelList(cmd *cobra.Command, args []string) error {
	parent, err := s.deployRelParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	var all []*clouddeploy.Release
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.Releases.List(parent).Context(ctx)
		if s.flagDeployRelPageSize > 0 {
			call = call.PageSize(s.flagDeployRelPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing releases: %w", err)
		}
		all = append(all, resp.Releases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDeployRelFormat)
}

func (s *deployCmds) runDeployRelPromote(cmd *cobra.Command, args []string) error {
	release, err := s.deployRelName(args[0])
	if err != nil {
		return err
	}
	rollout := &clouddeploy.Rollout{
		TargetId: s.flagDeployRelPromoteTgt,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Create(release, rollout).
		RolloutId(s.flagDeployRelPromoteID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("promoting release: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Promote issued for release [%s] to target [%s] (operation: %s).\n",
		args[0], s.flagDeployRelPromoteTgt, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDeployRelFormat)
}
