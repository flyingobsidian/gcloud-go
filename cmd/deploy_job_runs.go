package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy job-runs (#1529) ---

func (s *deployCmds) initDeployJobRuns() {
	all := []*cobra.Command{s.deployJRDescribeCmd, s.deployJRListCmd, s.deployJRTerminateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployJRRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployJRPipeline, "delivery-pipeline", "",
			"Delivery pipeline (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&s.flagDeployJRRelease, "release", "", "Release (required)")
		_ = c.MarkFlagRequired("release")
		c.Flags().StringVar(&s.flagDeployJRRollout, "rollout", "", "Rollout (required)")
		_ = c.MarkFlagRequired("rollout")
		c.Flags().StringVar(&s.flagDeployJRFormat, "format", "", "Output format")
	}
	s.deployJRListCmd.Flags().Int64Var(&s.flagDeployJRPageSize, "page-size", 0, "Maximum results per page")

	s.deployJRCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployJRCmd)
}

func (s *deployCmds) deployJRParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	pipeline := deployChild("deliveryPipelines", s.flagDeployJRPipeline, deployLocationParent(project, s.flagDeployJRRegion))
	release := deployChild("releases", s.flagDeployJRRelease, pipeline)
	return deployChild("rollouts", s.flagDeployJRRollout, release), nil
}

func (s *deployCmds) deployJRName(id string) (string, error) {
	parent, err := s.deployJRParent()
	if err != nil {
		return "", err
	}
	return deployChild("jobRuns", id, parent), nil
}

func (s *deployCmds) runDeployJRDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployJRName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.JobRuns.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job run: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployJRFormat)
}

func (s *deployCmds) runDeployJRList(cmd *cobra.Command, args []string) error {
	parent, err := s.deployJRParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	var all []*clouddeploy.JobRun
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.JobRuns.List(parent).Context(ctx)
		if s.flagDeployJRPageSize > 0 {
			call = call.PageSize(s.flagDeployJRPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing job runs: %w", err)
		}
		all = append(all, resp.JobRuns...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDeployJRFormat)
}

func (s *deployCmds) runDeployJRTerminate(cmd *cobra.Command, args []string) error {
	name, err := s.deployJRName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.JobRuns.Terminate(name, &clouddeploy.TerminateJobRunRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("terminating job run: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Terminated job run [%s].\n", args[0])
	return nil
}
