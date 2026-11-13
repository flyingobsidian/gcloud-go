package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai hp-tuning-jobs (#1453) ---

func (s *aiCmds) initAiHpTuningJobs() {
	all := []*cobra.Command{
		s.aiHPTCancelCmd, s.aiHPTCreateCmd, s.aiHPTDescribeCmd, s.aiHPTListCmd, s.aiHPTStreamLogsCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIHPTRegion, "region", "", "Region where the job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAIHPTFormat, "format", "", "Output format")
	}
	s.aiHPTCreateCmd.Flags().StringVar(&s.flagAIHPTConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the HyperparameterTuningJob body (required)")
	_ = s.aiHPTCreateCmd.MarkFlagRequired("config-file")
	s.aiHPTListCmd.Flags().StringVar(&s.flagAIHPTFilter, "filter", "", "Server-side filter expression")
	s.aiHPTListCmd.Flags().Int64Var(&s.flagAIHPTPageSize, "page-size", 0, "Maximum results per page")
	s.aiHPTListCmd.Flags().StringVar(&s.flagAIHPTReadMask, "read-mask", "", "Field mask for reads")
	s.aiHPTStreamLogsCmd.Flags().DurationVar(&s.flagAIHPTTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the job to reach a terminal state")

	s.aiHPTCmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiHPTCmd)
}

func (s *aiCmds) aiHPTParent() (string, error) { return s.aiParent(s.flagAIHPTRegion) }

func (s *aiCmds) aiHPTName(id string) (string, error) {
	parent, err := s.aiHPTParent()
	if err != nil {
		return "", err
	}
	return aiChild("hyperparameterTuningJobs", id, parent), nil
}

func (s *aiCmds) runAIHPTCancel(cmd *cobra.Command, args []string) error {
	name, err := s.aiHPTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIHPTRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.HyperparameterTuningJobs.Cancel(name,
		&aiplatform.GoogleCloudAiplatformV1CancelHyperparameterTuningJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling hyperparameter tuning job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Cancel request issued for hyperparameter tuning job [%s].\n", args[0])
	return nil
}

func (s *aiCmds) runAIHPTCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiHPTParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1HyperparameterTuningJob{}
	if err := loadYAMLOrJSONInto(s.flagAIHPTConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIHPTRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HyperparameterTuningJobs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating hyperparameter tuning job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Created hyperparameter tuning job [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagAIHPTFormat)
}

func (s *aiCmds) runAIHPTDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiHPTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIHPTRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HyperparameterTuningJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing hyperparameter tuning job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIHPTFormat)
}

func (s *aiCmds) runAIHPTList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiHPTParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIHPTRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1HyperparameterTuningJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.HyperparameterTuningJobs.List(parent).Context(ctx)
		if s.flagAIHPTFilter != "" {
			call = call.Filter(s.flagAIHPTFilter)
		}
		if s.flagAIHPTPageSize > 0 {
			call = call.PageSize(s.flagAIHPTPageSize)
		}
		if s.flagAIHPTReadMask != "" {
			call = call.ReadMask(s.flagAIHPTReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing hyperparameter tuning jobs: %w", err)
		}
		all = append(all, resp.HyperparameterTuningJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIHPTFormat)
}

func (s *aiCmds) runAIHPTStreamLogs(cmd *cobra.Command, args []string) error {
	name, err := s.aiHPTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIHPTRegion)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(s.flagAIHPTTimeout)
	for {
		got, err := svc.Projects.Locations.HyperparameterTuningJobs.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling hyperparameter tuning job: %w", err)
		}
		fmt.Fprintf(s.inv.stderr, "[%s] %s state=%s\n",
			time.Now().Format(time.RFC3339), got.Name, got.State)
		if aiJobStateTerminal(got.State) {
			return emitFormattedTo(s.inv.stdout, got, s.flagAIHPTFormat)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for hp-tuning job [%s] to reach terminal state (last state=%s)",
				s.flagAIHPTTimeout, args[0], got.State)
		}
		time.Sleep(5 * time.Second)
	}
}
