package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai tuning-jobs (#1462) ---

func (s *aiCmds) initAiTuningJobs() {
	all := []*cobra.Command{
		s.aiTJCancelCmd, s.aiTJCreateCmd, s.aiTJDescribeCmd, s.aiTJListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAITJRegion, "region", "", "Region where the tuning job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAITJFormat, "format", "", "Output format")
	}
	s.aiTJCreateCmd.Flags().StringVar(&s.flagAITJConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the TuningJob body (required)")
	_ = s.aiTJCreateCmd.MarkFlagRequired("config-file")
	s.aiTJListCmd.Flags().StringVar(&s.flagAITJFilter, "filter", "", "Server-side filter expression")
	s.aiTJListCmd.Flags().Int64Var(&s.flagAITJPageSize, "page-size", 0, "Maximum results per page")

	s.aiTJCmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiTJCmd)
}

func (s *aiCmds) aiTJParent() (string, error) { return s.aiParent(s.flagAITJRegion) }

func (s *aiCmds) aiTJName(id string) (string, error) {
	parent, err := s.aiTJParent()
	if err != nil {
		return "", err
	}
	return aiChild("tuningJobs", id, parent), nil
}

func (s *aiCmds) runAITJCancel(cmd *cobra.Command, args []string) error {
	name, err := s.aiTJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITJRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.TuningJobs.Cancel(name,
		&aiplatform.GoogleCloudAiplatformV1CancelTuningJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling tuning job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Cancel request issued for tuning job [%s].\n", args[0])
	return nil
}

func (s *aiCmds) runAITJCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiTJParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1TuningJob{}
	if err := loadYAMLOrJSONInto(s.flagAITJConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TuningJobs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tuning job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Created tuning job [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagAITJFormat)
}

func (s *aiCmds) runAITJDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiTJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TuningJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing tuning job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAITJFormat)
}

func (s *aiCmds) runAITJList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiTJParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITJRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1TuningJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.TuningJobs.List(parent).Context(ctx)
		if s.flagAITJFilter != "" {
			call = call.Filter(s.flagAITJFilter)
		}
		if s.flagAITJPageSize > 0 {
			call = call.PageSize(s.flagAITJPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tuning jobs: %w", err)
		}
		all = append(all, resp.TuningJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAITJFormat)
}
