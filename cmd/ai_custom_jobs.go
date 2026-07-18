package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai custom-jobs (#1451) ---

var aiCustomJobsCmd = &cobra.Command{Use: "custom-jobs", Short: "Manage Vertex AI custom training jobs"}

var (
	flagAICJRegion     string
	flagAICJFormat     string
	flagAICJConfigFile string
	flagAICJFilter     string
	flagAICJOrderBy    string
	flagAICJPageSize   int64
	flagAICJReadMask   string
	flagAICJTimeout    time.Duration
)

var (
	aiCJCancelCmd = &cobra.Command{
		Use: "cancel CUSTOM_JOB", Short: "Cancel a custom training job",
		Args: cobra.ExactArgs(1), RunE: runAICJCancel,
	}
	aiCJCreateCmd = &cobra.Command{
		Use: "create", Short: "Submit a custom training job",
		Args: cobra.NoArgs, RunE: runAICJCreate,
	}
	aiCJDescribeCmd = &cobra.Command{
		Use: "describe CUSTOM_JOB", Short: "Describe a custom training job",
		Args: cobra.ExactArgs(1), RunE: runAICJDescribe,
	}
	aiCJListCmd = &cobra.Command{
		Use: "list", Short: "List custom training jobs",
		Args: cobra.NoArgs, RunE: runAICJList,
	}
	aiCJLocalRunCmd = &cobra.Command{
		Use:   "local-run",
		Short: "Not supported: run a training job locally (requires Docker)",
		Long: "local-run requires Docker and is not supported by gcloud-go; " +
			"use `gcloud ai custom-jobs create` to submit the job to Vertex AI instead.",
		Args: cobra.ArbitraryArgs, RunE: runAICJLocalRun,
	}
	aiCJStreamLogsCmd = &cobra.Command{
		Use: "stream-logs CUSTOM_JOB",
		Short: "Poll a custom job until it reaches a terminal state, " +
			"printing its state each interval",
		Args: cobra.ExactArgs(1), RunE: runAICJStreamLogs,
	}
)

func init() {
	all := []*cobra.Command{
		aiCJCancelCmd, aiCJCreateCmd, aiCJDescribeCmd, aiCJListCmd,
		aiCJStreamLogsCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAICJRegion, "region", "", "Region where the custom job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAICJFormat, "format", "", "Output format")
	}
	aiCJCreateCmd.Flags().StringVar(&flagAICJConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the CustomJob body (required)")
	_ = aiCJCreateCmd.MarkFlagRequired("config-file")
	aiCJListCmd.Flags().StringVar(&flagAICJFilter, "filter", "", "Server-side filter expression")
	aiCJListCmd.Flags().StringVar(&flagAICJOrderBy, "order-by", "", "Order-by expression")
	aiCJListCmd.Flags().Int64Var(&flagAICJPageSize, "page-size", 0, "Maximum results per page")
	aiCJListCmd.Flags().StringVar(&flagAICJReadMask, "read-mask", "", "Field mask for reads")
	aiCJStreamLogsCmd.Flags().DurationVar(&flagAICJTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the job to reach a terminal state")

	aiCustomJobsCmd.AddCommand(all...)
	aiCustomJobsCmd.AddCommand(aiCJLocalRunCmd)
	aiCmd.AddCommand(aiCustomJobsCmd)
}

func aiCJParent() (string, error) { return aiParent(flagAICJRegion) }

func aiCJName(id string) (string, error) {
	parent, err := aiCJParent()
	if err != nil {
		return "", err
	}
	return aiChild("customJobs", id, parent), nil
}

func runAICJCancel(cmd *cobra.Command, args []string) error {
	name, err := aiCJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAICJRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.CustomJobs.Cancel(name, &aiplatform.GoogleCloudAiplatformV1CancelCustomJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling custom job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Cancel request issued for custom job [%s].\n", args[0])
	return nil
}

func runAICJCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiCJParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1CustomJob{}
	if err := loadYAMLOrJSONInto(flagAICJConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAICJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CustomJobs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating custom job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created custom job [%s].\n", got.Name)
	return emitFormatted(got, flagAICJFormat)
}

func runAICJDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiCJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAICJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CustomJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing custom job: %w", err)
	}
	return emitFormatted(got, flagAICJFormat)
}

func runAICJList(cmd *cobra.Command, args []string) error {
	parent, err := aiCJParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAICJRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1CustomJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.CustomJobs.List(parent).Context(ctx)
		if flagAICJFilter != "" {
			call = call.Filter(flagAICJFilter)
		}
		if flagAICJPageSize > 0 {
			call = call.PageSize(flagAICJPageSize)
		}
		if flagAICJReadMask != "" {
			call = call.ReadMask(flagAICJReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing custom jobs: %w", err)
		}
		all = append(all, resp.CustomJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAICJFormat)
}

func runAICJLocalRun(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("local-run requires Docker and is not supported in gcloud-go; use \"create\" to submit to Vertex AI")
}

// runAICJStreamLogs polls the custom job until it reaches a terminal state
// (JOB_STATE_SUCCEEDED / JOB_STATE_FAILED / JOB_STATE_CANCELLED /
// JOB_STATE_EXPIRED / JOB_STATE_PARTIALLY_SUCCEEDED) or --timeout expires.
// The aiplatform REST surface does not expose training-container log entries;
// callers wanting the full log stream should use `gcloud logging tail`.
func runAICJStreamLogs(cmd *cobra.Command, args []string) error {
	name, err := aiCJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAICJRegion)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(flagAICJTimeout)
	for {
		got, err := svc.Projects.Locations.CustomJobs.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling custom job: %w", err)
		}
		fmt.Fprintf(os.Stderr, "[%s] %s state=%s\n",
			time.Now().Format(time.RFC3339), got.Name, got.State)
		if aiJobStateTerminal(got.State) {
			return emitFormatted(got, flagAICJFormat)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for custom job [%s] to reach terminal state (last state=%s)",
				flagAICJTimeout, args[0], got.State)
		}
		time.Sleep(5 * time.Second)
	}
}

// aiJobStateTerminal reports whether the given aiplatform JobState value is a
// terminal state. Shared with hp-tuning-jobs stream-logs.
func aiJobStateTerminal(state string) bool {
	switch strings.ToUpper(state) {
	case "JOB_STATE_SUCCEEDED",
		"JOB_STATE_FAILED",
		"JOB_STATE_CANCELLED",
		"JOB_STATE_EXPIRED",
		"JOB_STATE_PARTIALLY_SUCCEEDED":
		return true
	}
	return false
}
