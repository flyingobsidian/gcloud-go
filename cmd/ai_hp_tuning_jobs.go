package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai hp-tuning-jobs (#1453) ---

var aiHPTCmd = &cobra.Command{Use: "hp-tuning-jobs", Short: "Manage Vertex AI hyperparameter tuning jobs"}

var (
	flagAIHPTRegion     string
	flagAIHPTFormat     string
	flagAIHPTConfigFile string
	flagAIHPTFilter     string
	flagAIHPTPageSize   int64
	flagAIHPTReadMask   string
	flagAIHPTTimeout    time.Duration
)

var (
	aiHPTCancelCmd = &cobra.Command{
		Use: "cancel HP_TUNING_JOB", Short: "Cancel a hyperparameter tuning job",
		Args: cobra.ExactArgs(1), RunE: runAIHPTCancel,
	}
	aiHPTCreateCmd = &cobra.Command{
		Use: "create", Short: "Submit a hyperparameter tuning job",
		Args: cobra.NoArgs, RunE: runAIHPTCreate,
	}
	aiHPTDescribeCmd = &cobra.Command{
		Use: "describe HP_TUNING_JOB", Short: "Describe a hyperparameter tuning job",
		Args: cobra.ExactArgs(1), RunE: runAIHPTDescribe,
	}
	aiHPTListCmd = &cobra.Command{
		Use: "list", Short: "List hyperparameter tuning jobs",
		Args: cobra.NoArgs, RunE: runAIHPTList,
	}
	aiHPTStreamLogsCmd = &cobra.Command{
		Use: "stream-logs HP_TUNING_JOB",
		Short: "Poll a hyperparameter tuning job until it reaches a terminal state, " +
			"printing its state each interval",
		Args: cobra.ExactArgs(1), RunE: runAIHPTStreamLogs,
	}
)

func init() {
	all := []*cobra.Command{
		aiHPTCancelCmd, aiHPTCreateCmd, aiHPTDescribeCmd, aiHPTListCmd, aiHPTStreamLogsCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIHPTRegion, "region", "", "Region where the job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAIHPTFormat, "format", "", "Output format")
	}
	aiHPTCreateCmd.Flags().StringVar(&flagAIHPTConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the HyperparameterTuningJob body (required)")
	_ = aiHPTCreateCmd.MarkFlagRequired("config-file")
	aiHPTListCmd.Flags().StringVar(&flagAIHPTFilter, "filter", "", "Server-side filter expression")
	aiHPTListCmd.Flags().Int64Var(&flagAIHPTPageSize, "page-size", 0, "Maximum results per page")
	aiHPTListCmd.Flags().StringVar(&flagAIHPTReadMask, "read-mask", "", "Field mask for reads")
	aiHPTStreamLogsCmd.Flags().DurationVar(&flagAIHPTTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the job to reach a terminal state")

	aiHPTCmd.AddCommand(all...)
	aiCmd.AddCommand(aiHPTCmd)
}

func aiHPTParent() (string, error) { return aiParent(flagAIHPTRegion) }

func aiHPTName(id string) (string, error) {
	parent, err := aiHPTParent()
	if err != nil {
		return "", err
	}
	return aiChild("hyperparameterTuningJobs", id, parent), nil
}

func runAIHPTCancel(cmd *cobra.Command, args []string) error {
	name, err := aiHPTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIHPTRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.HyperparameterTuningJobs.Cancel(name,
		&aiplatform.GoogleCloudAiplatformV1CancelHyperparameterTuningJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling hyperparameter tuning job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Cancel request issued for hyperparameter tuning job [%s].\n", args[0])
	return nil
}

func runAIHPTCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiHPTParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1HyperparameterTuningJob{}
	if err := loadYAMLOrJSONInto(flagAIHPTConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIHPTRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HyperparameterTuningJobs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating hyperparameter tuning job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created hyperparameter tuning job [%s].\n", got.Name)
	return emitFormatted(got, flagAIHPTFormat)
}

func runAIHPTDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiHPTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIHPTRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HyperparameterTuningJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing hyperparameter tuning job: %w", err)
	}
	return emitFormatted(got, flagAIHPTFormat)
}

func runAIHPTList(cmd *cobra.Command, args []string) error {
	parent, err := aiHPTParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIHPTRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1HyperparameterTuningJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.HyperparameterTuningJobs.List(parent).Context(ctx)
		if flagAIHPTFilter != "" {
			call = call.Filter(flagAIHPTFilter)
		}
		if flagAIHPTPageSize > 0 {
			call = call.PageSize(flagAIHPTPageSize)
		}
		if flagAIHPTReadMask != "" {
			call = call.ReadMask(flagAIHPTReadMask)
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
	return emitFormatted(all, flagAIHPTFormat)
}

func runAIHPTStreamLogs(cmd *cobra.Command, args []string) error {
	name, err := aiHPTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIHPTRegion)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(flagAIHPTTimeout)
	for {
		got, err := svc.Projects.Locations.HyperparameterTuningJobs.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling hyperparameter tuning job: %w", err)
		}
		fmt.Fprintf(os.Stderr, "[%s] %s state=%s\n",
			time.Now().Format(time.RFC3339), got.Name, got.State)
		if aiJobStateTerminal(got.State) {
			return emitFormatted(got, flagAIHPTFormat)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for hp-tuning job [%s] to reach terminal state (last state=%s)",
				flagAIHPTTimeout, args[0], got.State)
		}
		time.Sleep(5 * time.Second)
	}
}
