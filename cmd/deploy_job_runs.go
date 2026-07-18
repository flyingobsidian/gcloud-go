package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy job-runs (#1529) ---

var deployJRCmd = &cobra.Command{Use: "job-runs", Short: "Manage Cloud Deploy job runs"}

var (
	flagDeployJRRegion   string
	flagDeployJRPipeline string
	flagDeployJRRelease  string
	flagDeployJRRollout  string
	flagDeployJRFormat   string
	flagDeployJRPageSize int64
)

var (
	deployJRDescribeCmd = &cobra.Command{
		Use: "describe JOB_RUN", Short: "Describe a job run",
		Args: cobra.ExactArgs(1), RunE: runDeployJRDescribe,
	}
	deployJRListCmd = &cobra.Command{
		Use: "list", Short: "List job runs for a rollout",
		Args: cobra.NoArgs, RunE: runDeployJRList,
	}
	deployJRTerminateCmd = &cobra.Command{
		Use: "terminate JOB_RUN", Short: "Terminate a running job run",
		Args: cobra.ExactArgs(1), RunE: runDeployJRTerminate,
	}
)

func init() {
	all := []*cobra.Command{deployJRDescribeCmd, deployJRListCmd, deployJRTerminateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployJRRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployJRPipeline, "delivery-pipeline", "",
			"Delivery pipeline (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&flagDeployJRRelease, "release", "", "Release (required)")
		_ = c.MarkFlagRequired("release")
		c.Flags().StringVar(&flagDeployJRRollout, "rollout", "", "Rollout (required)")
		_ = c.MarkFlagRequired("rollout")
		c.Flags().StringVar(&flagDeployJRFormat, "format", "", "Output format")
	}
	deployJRListCmd.Flags().Int64Var(&flagDeployJRPageSize, "page-size", 0, "Maximum results per page")

	deployJRCmd.AddCommand(all...)
	deployCmd.AddCommand(deployJRCmd)
}

func deployJRParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	pipeline := deployChild("deliveryPipelines", flagDeployJRPipeline, deployLocationParent(project, flagDeployJRRegion))
	release := deployChild("releases", flagDeployJRRelease, pipeline)
	return deployChild("rollouts", flagDeployJRRollout, release), nil
}

func deployJRName(id string) (string, error) {
	parent, err := deployJRParent()
	if err != nil {
		return "", err
	}
	return deployChild("jobRuns", id, parent), nil
}

func runDeployJRDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployJRName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.JobRuns.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job run: %w", err)
	}
	return emitFormatted(got, flagDeployJRFormat)
}

func runDeployJRList(cmd *cobra.Command, args []string) error {
	parent, err := deployJRParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*clouddeploy.JobRun
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.JobRuns.List(parent).Context(ctx)
		if flagDeployJRPageSize > 0 {
			call = call.PageSize(flagDeployJRPageSize)
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
	return emitFormatted(all, flagDeployJRFormat)
}

func runDeployJRTerminate(cmd *cobra.Command, args []string) error {
	name, err := deployJRName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.JobRuns.Terminate(name, &clouddeploy.TerminateJobRunRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("terminating job run: %w", err)
	}
	fmt.Printf("Terminated job run [%s].\n", args[0])
	return nil
}
