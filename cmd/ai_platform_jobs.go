package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform jobs (#982) ---

var aiPlatformJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage AI Platform jobs"}

var (
	flagAIPlatformJobsFormat     string
	flagAIPlatformJobsConfigFile string
	flagAIPlatformJobsPageSize   int64
	flagAIPlatformJobsFilter     string
	flagAIPlatformJobsUpdateMask string
	flagAIPlatformJobsJobID      string
	flagAIPlatformJobsPollEvery  time.Duration
	flagAIPlatformJobsTimeout    time.Duration
)

var (
	aiPlatformJobsCancelCmd = &cobra.Command{
		Use: "cancel JOB", Short: "Cancel a running AI Platform job",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformJobsCancel,
	}
	aiPlatformJobsDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe an AI Platform job",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformJobsDescribe,
	}
	aiPlatformJobsListCmd = &cobra.Command{
		Use: "list", Short: "List AI Platform jobs",
		Args: cobra.NoArgs, RunE: runAIPlatformJobsList,
	}
	aiPlatformJobsStreamCmd = &cobra.Command{
		Use: "stream-logs JOB", Short: "Poll an AI Platform job's state until it reaches a terminal state",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformJobsStream,
	}
	aiPlatformJobsSubmitCmd = &cobra.Command{
		Use: "submit", Short: "Submit an AI Platform training or prediction job",
		Args: cobra.NoArgs, RunE: runAIPlatformJobsSubmit,
	}
	aiPlatformJobsUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update an AI Platform job",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformJobsUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		aiPlatformJobsCancelCmd, aiPlatformJobsDescribeCmd, aiPlatformJobsListCmd,
		aiPlatformJobsStreamCmd, aiPlatformJobsSubmitCmd, aiPlatformJobsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIPlatformJobsFormat, "format", "", "Output format")
	}
	aiPlatformJobsListCmd.Flags().Int64Var(&flagAIPlatformJobsPageSize, "page-size", 0, "Maximum results per page")
	aiPlatformJobsListCmd.Flags().StringVar(&flagAIPlatformJobsFilter, "filter", "", "List filter expression")

	aiPlatformJobsStreamCmd.Flags().DurationVar(&flagAIPlatformJobsPollEvery, "poll-interval", 5*time.Second,
		"Interval between job state polls")
	aiPlatformJobsStreamCmd.Flags().DurationVar(&flagAIPlatformJobsTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the job to reach a terminal state")

	aiPlatformJobsSubmitCmd.Flags().StringVar(&flagAIPlatformJobsJobID, "job-id", "", "User-specified job id (required)")
	aiPlatformJobsSubmitCmd.Flags().StringVar(&flagAIPlatformJobsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Job body (required)")
	_ = aiPlatformJobsSubmitCmd.MarkFlagRequired("job-id")
	_ = aiPlatformJobsSubmitCmd.MarkFlagRequired("config-file")

	aiPlatformJobsUpdateCmd.Flags().StringVar(&flagAIPlatformJobsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Job patch body (required)")
	aiPlatformJobsUpdateCmd.Flags().StringVar(&flagAIPlatformJobsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")
	_ = aiPlatformJobsUpdateCmd.MarkFlagRequired("config-file")

	aiPlatformJobsCmd.AddCommand(all...)
	aiPlatformCmd.AddCommand(aiPlatformJobsCmd)
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

func runAIPlatformJobsCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Jobs.Cancel(mlJobName(project, args[0]), &ml.GoogleCloudMlV1__CancelJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling job: %w", err)
	}
	fmt.Printf("Cancelled job [%s].\n", args[0])
	return nil
}

func runAIPlatformJobsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Jobs.Get(mlJobName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormatted(got, flagAIPlatformJobsFormat)
}

func runAIPlatformJobsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*ml.GoogleCloudMlV1__Job
	pageToken := ""
	for {
		call := svc.Projects.Jobs.List(mlProjectPath(project)).Context(ctx)
		if flagAIPlatformJobsPageSize > 0 {
			call = call.PageSize(flagAIPlatformJobsPageSize)
		}
		if flagAIPlatformJobsFilter != "" {
			call = call.Filter(flagAIPlatformJobsFilter)
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
	return emitFormatted(all, flagAIPlatformJobsFormat)
}

func runAIPlatformJobsStream(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := mlJobName(project, args[0])
	deadline := time.Now().Add(flagAIPlatformJobsTimeout)
	lastState := ""
	for {
		job, err := svc.Projects.Jobs.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling job: %w", err)
		}
		if job.State != lastState {
			fmt.Printf("Job [%s] state: %s\n", args[0], job.State)
			lastState = job.State
		}
		if mlJobIsTerminal(job.State) {
			if job.State == "FAILED" || job.State == "CANCELLED" {
				if job.ErrorMessage != "" {
					return fmt.Errorf("job %s terminated in state %s: %s", args[0], job.State, job.ErrorMessage)
				}
				return fmt.Errorf("job %s terminated in state %s", args[0], job.State)
			}
			return emitFormatted(job, flagAIPlatformJobsFormat)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for job [%s] to reach a terminal state (last state: %s)",
				flagAIPlatformJobsTimeout, args[0], job.State)
		}
		time.Sleep(flagAIPlatformJobsPollEvery)
	}
}

func runAIPlatformJobsSubmit(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Job{}
	if err := loadYAMLOrJSONInto(flagAIPlatformJobsConfigFile, body); err != nil {
		return err
	}
	body.JobId = flagAIPlatformJobsJobID
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Jobs.Create(mlProjectPath(project), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting job: %w", err)
	}
	fmt.Printf("Submitted job [%s].\n", flagAIPlatformJobsJobID)
	return emitFormatted(got, flagAIPlatformJobsFormat)
}

func runAIPlatformJobsUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Job{}
	if err := loadYAMLOrJSONInto(flagAIPlatformJobsConfigFile, body); err != nil {
		return err
	}
	mask := flagAIPlatformJobsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
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
	fmt.Printf("Updated job [%s].\n", args[0])
	return emitFormatted(got, flagAIPlatformJobsFormat)
}
