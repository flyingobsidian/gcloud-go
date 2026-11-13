package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy rollouts (#1531) ---

func (s *deployCmds) initDeployRollouts() {
	all := []*cobra.Command{
		s.deployRollAdvanceCmd, s.deployRollApproveCmd, s.deployRollCancelCmd, s.deployRollDescribeCmd,
		s.deployRollListCmd, s.deployRollRejectCmd, s.deployRollIgnoreJobCmd, s.deployRollRetryJobCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployRollRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployRollPipeline, "delivery-pipeline", "",
			"Delivery pipeline that owns the release (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&s.flagDeployRollRelease, "release", "",
			"Release that owns the rollout (required)")
		_ = c.MarkFlagRequired("release")
		c.Flags().StringVar(&s.flagDeployRollFormat, "format", "", "Output format")
	}
	s.deployRollListCmd.Flags().Int64Var(&s.flagDeployRollPageSize, "page-size", 0, "Maximum results per page")

	s.deployRollAdvanceCmd.Flags().StringVar(&s.flagDeployRollPhaseID, "phase-id", "",
		"Phase ID to advance the rollout to (required)")
	_ = s.deployRollAdvanceCmd.MarkFlagRequired("phase-id")

	for _, c := range []*cobra.Command{s.deployRollIgnoreJobCmd, s.deployRollRetryJobCmd} {
		c.Flags().StringVar(&s.flagDeployRollPhaseID, "phase-id", "",
			"Phase ID that owns the job (required)")
		_ = c.MarkFlagRequired("phase-id")
		c.Flags().StringVar(&s.flagDeployRollJobID, "job-id", "",
			"Job ID within the phase (required)")
		_ = c.MarkFlagRequired("job-id")
	}
	for _, c := range []*cobra.Command{
		s.deployRollAdvanceCmd, s.deployRollApproveCmd, s.deployRollCancelCmd,
		s.deployRollRejectCmd, s.deployRollIgnoreJobCmd, s.deployRollRetryJobCmd,
	} {
		c.Flags().StringSliceVar(&s.flagDeployRollOverride, "override-deploy-policies", nil,
			"Deploy policies to override (repeatable)")
	}

	s.deployRollCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployRollCmd)
}

func (s *deployCmds) deployRollParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	pipeline := deployChild("deliveryPipelines", s.flagDeployRollPipeline, deployLocationParent(project, s.flagDeployRollRegion))
	return deployChild("releases", s.flagDeployRollRelease, pipeline), nil
}

func (s *deployCmds) deployRollName(id string) (string, error) {
	parent, err := s.deployRollParent()
	if err != nil {
		return "", err
	}
	return deployChild("rollouts", id, parent), nil
}

func (s *deployCmds) runDeployRollAdvance(cmd *cobra.Command, args []string) error {
	name, err := s.deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Advance(name, &clouddeploy.AdvanceRolloutRequest{
		PhaseId:              s.flagDeployRollPhaseID,
		OverrideDeployPolicy: s.flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("advancing rollout: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Advanced rollout [%s] to phase [%s].\n", args[0], s.flagDeployRollPhaseID)
	return nil
}

func (s *deployCmds) runDeployRollApprove(cmd *cobra.Command, args []string) error {
	name, err := s.deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Approve(name, &clouddeploy.ApproveRolloutRequest{
		Approved:             true,
		OverrideDeployPolicy: s.flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("approving rollout: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Approved rollout [%s].\n", args[0])
	return nil
}

func (s *deployCmds) runDeployRollCancel(cmd *cobra.Command, args []string) error {
	name, err := s.deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Cancel(name, &clouddeploy.CancelRolloutRequest{
		OverrideDeployPolicy: s.flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling rollout: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled rollout [%s].\n", args[0])
	return nil
}

func (s *deployCmds) runDeployRollDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing rollout: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployRollFormat)
}

func (s *deployCmds) runDeployRollList(cmd *cobra.Command, args []string) error {
	parent, err := s.deployRollParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	var all []*clouddeploy.Rollout
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.List(parent).Context(ctx)
		if s.flagDeployRollPageSize > 0 {
			call = call.PageSize(s.flagDeployRollPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing rollouts: %w", err)
		}
		all = append(all, resp.Rollouts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDeployRollFormat)
}

func (s *deployCmds) runDeployRollReject(cmd *cobra.Command, args []string) error {
	name, err := s.deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Approve(name, &clouddeploy.ApproveRolloutRequest{
		Approved:             false,
		OverrideDeployPolicy: s.flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("rejecting rollout: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Rejected rollout [%s].\n", args[0])
	return nil
}

func (s *deployCmds) runDeployRollIgnoreJob(cmd *cobra.Command, args []string) error {
	name, err := s.deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.IgnoreJob(name, &clouddeploy.IgnoreJobRequest{
		PhaseId:              s.flagDeployRollPhaseID,
		JobId:                s.flagDeployRollJobID,
		OverrideDeployPolicy: s.flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("ignoring job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Ignored job [%s] in phase [%s] of rollout [%s].\n", s.flagDeployRollJobID, s.flagDeployRollPhaseID, args[0])
	return nil
}

func (s *deployCmds) runDeployRollRetryJob(cmd *cobra.Command, args []string) error {
	name, err := s.deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.RetryJob(name, &clouddeploy.RetryJobRequest{
		PhaseId:              s.flagDeployRollPhaseID,
		JobId:                s.flagDeployRollJobID,
		OverrideDeployPolicy: s.flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("retrying job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Retried job [%s] in phase [%s] of rollout [%s].\n", s.flagDeployRollJobID, s.flagDeployRollPhaseID, args[0])
	return nil
}
