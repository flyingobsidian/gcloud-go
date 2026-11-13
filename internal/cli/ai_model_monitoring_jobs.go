package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai model-monitoring-jobs (#1457) ---

func (s *aiCmds) initAiModelMonitoringJobs() {
	all := []*cobra.Command{
		s.aiMMJCreateCmd, s.aiMMJDeleteCmd, s.aiMMJDescribeCmd, s.aiMMJListCmd,
		s.aiMMJUpdateCmd, s.aiMMJPauseCmd, s.aiMMJResumeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIMMJRegion, "region", "", "Region where the monitoring job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAIMMJFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.aiMMJCreateCmd, s.aiMMJUpdateCmd} {
		c.Flags().StringVar(&s.flagAIMMJConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ModelDeploymentMonitoringJob body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.aiMMJUpdateCmd.Flags().StringVar(&s.flagAIMMJUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")
	s.aiMMJListCmd.Flags().StringVar(&s.flagAIMMJFilter, "filter", "", "Server-side filter expression")
	s.aiMMJListCmd.Flags().Int64Var(&s.flagAIMMJPageSize, "page-size", 0, "Maximum results per page")
	s.aiMMJListCmd.Flags().StringVar(&s.flagAIMMJReadMask, "read-mask", "", "Field mask for reads")

	s.aiMMJCmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiMMJCmd)
}

func (s *aiCmds) aiMMJParent() (string, error) { return s.aiParent(s.flagAIMMJRegion) }

func (s *aiCmds) aiMMJName(id string) (string, error) {
	parent, err := s.aiMMJParent()
	if err != nil {
		return "", err
	}
	return aiChild("modelDeploymentMonitoringJobs", id, parent), nil
}

func (s *aiCmds) runAIMMJCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiMMJParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1ModelDeploymentMonitoringJob{}
	if err := loadYAMLOrJSONInto(s.flagAIMMJConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMMJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ModelDeploymentMonitoringJobs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating monitoring job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Created model monitoring job [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagAIMMJFormat)
}

func (s *aiCmds) runAIMMJDelete(cmd *cobra.Command, args []string) error {
	name, err := s.aiMMJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMMJRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ModelDeploymentMonitoringJobs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting monitoring job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Delete request issued for monitoring job [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIMMJFormat)
}

func (s *aiCmds) runAIMMJDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiMMJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMMJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ModelDeploymentMonitoringJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing monitoring job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIMMJFormat)
}

func (s *aiCmds) runAIMMJList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiMMJParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMMJRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1ModelDeploymentMonitoringJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.ModelDeploymentMonitoringJobs.List(parent).Context(ctx)
		if s.flagAIMMJFilter != "" {
			call = call.Filter(s.flagAIMMJFilter)
		}
		if s.flagAIMMJPageSize > 0 {
			call = call.PageSize(s.flagAIMMJPageSize)
		}
		if s.flagAIMMJReadMask != "" {
			call = call.ReadMask(s.flagAIMMJReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing monitoring jobs: %w", err)
		}
		all = append(all, resp.ModelDeploymentMonitoringJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIMMJFormat)
}

func (s *aiCmds) runAIMMJUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.aiMMJName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1ModelDeploymentMonitoringJob{}
	if err := loadYAMLOrJSONInto(s.flagAIMMJConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAIMMJUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMMJRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ModelDeploymentMonitoringJobs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating monitoring job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagAIMMJFormat)
}

func (s *aiCmds) runAIMMJPause(cmd *cobra.Command, args []string) error {
	name, err := s.aiMMJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMMJRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.ModelDeploymentMonitoringJobs.Pause(name,
		&aiplatform.GoogleCloudAiplatformV1PauseModelDeploymentMonitoringJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("pausing monitoring job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Pause request issued for monitoring job [%s].\n", args[0])
	return nil
}

func (s *aiCmds) runAIMMJResume(cmd *cobra.Command, args []string) error {
	name, err := s.aiMMJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMMJRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.ModelDeploymentMonitoringJobs.Resume(name,
		&aiplatform.GoogleCloudAiplatformV1ResumeModelDeploymentMonitoringJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming monitoring job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Resume request issued for monitoring job [%s].\n", args[0])
	return nil
}
