package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai tuning-jobs (#1462) ---

var aiTJCmd = &cobra.Command{Use: "tuning-jobs", Short: "Manage Vertex AI GenAI tuning jobs"}

var (
	flagAITJRegion     string
	flagAITJFormat     string
	flagAITJConfigFile string
	flagAITJFilter     string
	flagAITJPageSize   int64
)

var (
	aiTJCancelCmd = &cobra.Command{
		Use: "cancel TUNING_JOB", Short: "Cancel a tuning job",
		Args: cobra.ExactArgs(1), RunE: runAITJCancel,
	}
	aiTJCreateCmd = &cobra.Command{
		Use: "create", Short: "Submit a tuning job",
		Args: cobra.NoArgs, RunE: runAITJCreate,
	}
	aiTJDescribeCmd = &cobra.Command{
		Use: "describe TUNING_JOB", Short: "Describe a tuning job",
		Args: cobra.ExactArgs(1), RunE: runAITJDescribe,
	}
	aiTJListCmd = &cobra.Command{
		Use: "list", Short: "List tuning jobs",
		Args: cobra.NoArgs, RunE: runAITJList,
	}
)

func init() {
	all := []*cobra.Command{
		aiTJCancelCmd, aiTJCreateCmd, aiTJDescribeCmd, aiTJListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAITJRegion, "region", "", "Region where the tuning job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAITJFormat, "format", "", "Output format")
	}
	aiTJCreateCmd.Flags().StringVar(&flagAITJConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the TuningJob body (required)")
	_ = aiTJCreateCmd.MarkFlagRequired("config-file")
	aiTJListCmd.Flags().StringVar(&flagAITJFilter, "filter", "", "Server-side filter expression")
	aiTJListCmd.Flags().Int64Var(&flagAITJPageSize, "page-size", 0, "Maximum results per page")

	aiTJCmd.AddCommand(all...)
	aiCmd.AddCommand(aiTJCmd)
}

func aiTJParent() (string, error) { return aiParent(flagAITJRegion) }

func aiTJName(id string) (string, error) {
	parent, err := aiTJParent()
	if err != nil {
		return "", err
	}
	return aiChild("tuningJobs", id, parent), nil
}

func runAITJCancel(cmd *cobra.Command, args []string) error {
	name, err := aiTJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITJRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.TuningJobs.Cancel(name,
		&aiplatform.GoogleCloudAiplatformV1CancelTuningJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling tuning job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Cancel request issued for tuning job [%s].\n", args[0])
	return nil
}

func runAITJCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiTJParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1TuningJob{}
	if err := loadYAMLOrJSONInto(flagAITJConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TuningJobs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tuning job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created tuning job [%s].\n", got.Name)
	return emitFormatted(got, flagAITJFormat)
}

func runAITJDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiTJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TuningJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing tuning job: %w", err)
	}
	return emitFormatted(got, flagAITJFormat)
}

func runAITJList(cmd *cobra.Command, args []string) error {
	parent, err := aiTJParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITJRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1TuningJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.TuningJobs.List(parent).Context(ctx)
		if flagAITJFilter != "" {
			call = call.Filter(flagAITJFilter)
		}
		if flagAITJPageSize > 0 {
			call = call.PageSize(flagAITJPageSize)
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
	return emitFormatted(all, flagAITJFormat)
}
