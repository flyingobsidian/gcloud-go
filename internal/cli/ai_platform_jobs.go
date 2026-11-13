package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform jobs (#982) ---

func (s *aiPlatformCmds) initAiPlatformJobs() {
	all := []*cobra.Command{
		s.aiPlatformJobsCancelCmd, s.aiPlatformJobsDescribeCmd, s.aiPlatformJobsListCmd,
		s.aiPlatformJobsStreamCmd, s.aiPlatformJobsSubmitCmd, s.aiPlatformJobsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIPlatformJobsFormat, "format", "", "Output format")
	}
	s.aiPlatformJobsListCmd.Flags().Int64Var(&s.flagAIPlatformJobsPageSize, "page-size", 0, "Maximum results per page")
	s.aiPlatformJobsListCmd.Flags().StringVar(&s.flagAIPlatformJobsFilter, "filter", "", "List filter expression")

	s.aiPlatformJobsStreamCmd.Flags().DurationVar(&s.flagAIPlatformJobsPollEvery, "poll-interval", 5*time.Second,
		"Interval between job state polls")
	s.aiPlatformJobsStreamCmd.Flags().DurationVar(&s.flagAIPlatformJobsTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the job to reach a terminal state")

	s.aiPlatformJobsSubmitCmd.Flags().StringVar(&s.flagAIPlatformJobsJobID, "job-id", "", "User-specified job id (required)")
	s.aiPlatformJobsSubmitCmd.Flags().StringVar(&s.flagAIPlatformJobsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Job body (required)")
	_ = s.aiPlatformJobsSubmitCmd.MarkFlagRequired("job-id")
	_ = s.aiPlatformJobsSubmitCmd.MarkFlagRequired("config-file")

	s.aiPlatformJobsUpdateCmd.Flags().StringVar(&s.flagAIPlatformJobsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Job patch body (required)")
	s.aiPlatformJobsUpdateCmd.Flags().StringVar(&s.flagAIPlatformJobsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")
	_ = s.aiPlatformJobsUpdateCmd.MarkFlagRequired("config-file")

	s.aiPlatformJobsCmd.AddCommand(all...)
	s.aiPlatformCmd.AddCommand(s.aiPlatformJobsCmd)
}

// mlJobIsTerminal reports whether the given ml Job state is one that will not
// change (i.e. the job has finished, succeeded, failed, or been cancelled).
func mlJobIsTerminal(state string) bool {
	switch state {
	case "SUCCEEDED", "FAILED", "CANCELLED":
		return true
	}
	return false
}

func (s *aiPlatformCmds) runAIPlatformJobsCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Jobs.Cancel(mlJobName(project, args[0]), &ml.GoogleCloudMlV1__CancelJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled job [%s].\n", args[0])
	return nil
}

func (s *aiPlatformCmds) runAIPlatformJobsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Jobs.Get(mlJobName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPlatformJobsFormat)
}

func (s *aiPlatformCmds) runAIPlatformJobsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	var all []*ml.GoogleCloudMlV1__Job
	pageToken := ""
	for {
		call := svc.Projects.Jobs.List(mlProjectPath(project)).Context(ctx)
		if s.flagAIPlatformJobsPageSize > 0 {
			call = call.PageSize(s.flagAIPlatformJobsPageSize)
		}
		if s.flagAIPlatformJobsFilter != "" {
			call = call.Filter(s.flagAIPlatformJobsFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing jobs: %w", err)
		}
		all = append(all, resp.Jobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIPlatformJobsFormat)
}

func (s *aiPlatformCmds) runAIPlatformJobsStream(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	name := mlJobName(project, args[0])
	deadline := time.Now().Add(s.flagAIPlatformJobsTimeout)
	lastState := ""
	for {
		job, err := svc.Projects.Jobs.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling job: %w", err)
		}
		if job.State != lastState {
			fmt.Fprintf(s.inv.stdout, "Job [%s] state: %s\n", args[0], job.State)
			lastState = job.State
		}
		if mlJobIsTerminal(job.State) {
			if job.State == "FAILED" || job.State == "CANCELLED" {
				if job.ErrorMessage != "" {
					return fmt.Errorf("job %s terminated in state %s: %s", args[0], job.State, job.ErrorMessage)
				}
				return fmt.Errorf("job %s terminated in state %s", args[0], job.State)
			}
			return emitFormattedTo(s.inv.stdout, job, s.flagAIPlatformJobsFormat)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for job [%s] to reach a terminal state (last state: %s)",
				s.flagAIPlatformJobsTimeout, args[0], job.State)
		}
		time.Sleep(s.flagAIPlatformJobsPollEvery)
	}
}

func (s *aiPlatformCmds) runAIPlatformJobsSubmit(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Job{}
	if err := loadYAMLOrJSONInto(s.flagAIPlatformJobsConfigFile, body); err != nil {
		return err
	}
	body.JobId = s.flagAIPlatformJobsJobID
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Jobs.Create(mlProjectPath(project), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Submitted job [%s].\n", s.flagAIPlatformJobsJobID)
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPlatformJobsFormat)
}

func (s *aiPlatformCmds) runAIPlatformJobsUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Job{}
	if err := loadYAMLOrJSONInto(s.flagAIPlatformJobsConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAIPlatformJobsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Jobs.Patch(mlJobName(project, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated job [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPlatformJobsFormat)
}
