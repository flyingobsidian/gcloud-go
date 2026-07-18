package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai model-monitoring-jobs (#1457) ---

var aiMMJCmd = &cobra.Command{Use: "model-monitoring-jobs", Short: "Manage Vertex AI model deployment monitoring jobs"}

var (
	flagAIMMJRegion     string
	flagAIMMJFormat     string
	flagAIMMJConfigFile string
	flagAIMMJUpdateMask string
	flagAIMMJFilter     string
	flagAIMMJPageSize   int64
	flagAIMMJReadMask   string
)

var (
	aiMMJCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a model deployment monitoring job",
		Args: cobra.NoArgs, RunE: runAIMMJCreate,
	}
	aiMMJDeleteCmd = &cobra.Command{
		Use: "delete MONITORING_JOB", Short: "Delete a model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: runAIMMJDelete,
	}
	aiMMJDescribeCmd = &cobra.Command{
		Use: "describe MONITORING_JOB", Short: "Describe a model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: runAIMMJDescribe,
	}
	aiMMJListCmd = &cobra.Command{
		Use: "list", Short: "List model deployment monitoring jobs",
		Args: cobra.NoArgs, RunE: runAIMMJList,
	}
	aiMMJUpdateCmd = &cobra.Command{
		Use: "update MONITORING_JOB", Short: "Update a model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: runAIMMJUpdate,
	}
	aiMMJPauseCmd = &cobra.Command{
		Use: "pause MONITORING_JOB", Short: "Pause a model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: runAIMMJPause,
	}
	aiMMJResumeCmd = &cobra.Command{
		Use: "resume MONITORING_JOB", Short: "Resume a paused model deployment monitoring job",
		Args: cobra.ExactArgs(1), RunE: runAIMMJResume,
	}
)

func init() {
	all := []*cobra.Command{
		aiMMJCreateCmd, aiMMJDeleteCmd, aiMMJDescribeCmd, aiMMJListCmd,
		aiMMJUpdateCmd, aiMMJPauseCmd, aiMMJResumeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIMMJRegion, "region", "", "Region where the monitoring job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAIMMJFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{aiMMJCreateCmd, aiMMJUpdateCmd} {
		c.Flags().StringVar(&flagAIMMJConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ModelDeploymentMonitoringJob body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	aiMMJUpdateCmd.Flags().StringVar(&flagAIMMJUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")
	aiMMJListCmd.Flags().StringVar(&flagAIMMJFilter, "filter", "", "Server-side filter expression")
	aiMMJListCmd.Flags().Int64Var(&flagAIMMJPageSize, "page-size", 0, "Maximum results per page")
	aiMMJListCmd.Flags().StringVar(&flagAIMMJReadMask, "read-mask", "", "Field mask for reads")

	aiMMJCmd.AddCommand(all...)
	aiCmd.AddCommand(aiMMJCmd)
}

func aiMMJParent() (string, error) { return aiParent(flagAIMMJRegion) }

func aiMMJName(id string) (string, error) {
	parent, err := aiMMJParent()
	if err != nil {
		return "", err
	}
	return aiChild("modelDeploymentMonitoringJobs", id, parent), nil
}

func runAIMMJCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiMMJParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1ModelDeploymentMonitoringJob{}
	if err := loadYAMLOrJSONInto(flagAIMMJConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMMJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ModelDeploymentMonitoringJobs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating monitoring job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created model monitoring job [%s].\n", got.Name)
	return emitFormatted(got, flagAIMMJFormat)
}

func runAIMMJDelete(cmd *cobra.Command, args []string) error {
	name, err := aiMMJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMMJRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ModelDeploymentMonitoringJobs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting monitoring job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete request issued for monitoring job [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIMMJFormat)
}

func runAIMMJDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiMMJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMMJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ModelDeploymentMonitoringJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing monitoring job: %w", err)
	}
	return emitFormatted(got, flagAIMMJFormat)
}

func runAIMMJList(cmd *cobra.Command, args []string) error {
	parent, err := aiMMJParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMMJRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1ModelDeploymentMonitoringJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.ModelDeploymentMonitoringJobs.List(parent).Context(ctx)
		if flagAIMMJFilter != "" {
			call = call.Filter(flagAIMMJFilter)
		}
		if flagAIMMJPageSize > 0 {
			call = call.PageSize(flagAIMMJPageSize)
		}
		if flagAIMMJReadMask != "" {
			call = call.ReadMask(flagAIMMJReadMask)
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
	return emitFormatted(all, flagAIMMJFormat)
}

func runAIMMJUpdate(cmd *cobra.Command, args []string) error {
	name, err := aiMMJName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1ModelDeploymentMonitoringJob{}
	if err := loadYAMLOrJSONInto(flagAIMMJConfigFile, body); err != nil {
		return err
	}
	mask := flagAIMMJUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMMJRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ModelDeploymentMonitoringJobs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating monitoring job: %w", err)
	}
	return emitFormatted(op, flagAIMMJFormat)
}

func runAIMMJPause(cmd *cobra.Command, args []string) error {
	name, err := aiMMJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMMJRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.ModelDeploymentMonitoringJobs.Pause(name,
		&aiplatform.GoogleCloudAiplatformV1PauseModelDeploymentMonitoringJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("pausing monitoring job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Pause request issued for monitoring job [%s].\n", args[0])
	return nil
}

func runAIMMJResume(cmd *cobra.Command, args []string) error {
	name, err := aiMMJName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMMJRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.ModelDeploymentMonitoringJobs.Resume(name,
		&aiplatform.GoogleCloudAiplatformV1ResumeModelDeploymentMonitoringJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming monitoring job: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Resume request issued for monitoring job [%s].\n", args[0])
	return nil
}
