package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy rollouts (#1531) ---

var deployRollCmd = &cobra.Command{Use: "rollouts", Short: "Manage Cloud Deploy rollouts"}

var (
	flagDeployRollRegion   string
	flagDeployRollPipeline string
	flagDeployRollRelease  string
	flagDeployRollFormat   string
	flagDeployRollPageSize int64
	flagDeployRollPhaseID  string
	flagDeployRollJobID    string
	flagDeployRollOverride []string
)

var (
	deployRollAdvanceCmd = &cobra.Command{
		Use: "advance ROLLOUT", Short: "Advance a rollout to the specified phase",
		Args: cobra.ExactArgs(1), RunE: runDeployRollAdvance,
	}
	deployRollApproveCmd = &cobra.Command{
		Use: "approve ROLLOUT", Short: "Approve a rollout",
		Args: cobra.ExactArgs(1), RunE: runDeployRollApprove,
	}
	deployRollCancelCmd = &cobra.Command{
		Use: "cancel ROLLOUT", Short: "Cancel a rollout",
		Args: cobra.ExactArgs(1), RunE: runDeployRollCancel,
	}
	deployRollDescribeCmd = &cobra.Command{
		Use: "describe ROLLOUT", Short: "Describe a rollout",
		Args: cobra.ExactArgs(1), RunE: runDeployRollDescribe,
	}
	deployRollListCmd = &cobra.Command{
		Use: "list", Short: "List rollouts for a release",
		Args: cobra.NoArgs, RunE: runDeployRollList,
	}
	deployRollRejectCmd = &cobra.Command{
		Use: "reject ROLLOUT", Short: "Reject a pending rollout",
		Args: cobra.ExactArgs(1), RunE: runDeployRollReject,
	}
	deployRollIgnoreJobCmd = &cobra.Command{
		Use: "ignore-job ROLLOUT", Short: "Ignore a job in a rollout phase",
		Args: cobra.ExactArgs(1), RunE: runDeployRollIgnoreJob,
	}
	deployRollRetryJobCmd = &cobra.Command{
		Use: "retry-job ROLLOUT", Short: "Retry a failed job in a rollout phase",
		Args: cobra.ExactArgs(1), RunE: runDeployRollRetryJob,
	}
)

func init() {
	all := []*cobra.Command{
		deployRollAdvanceCmd, deployRollApproveCmd, deployRollCancelCmd, deployRollDescribeCmd,
		deployRollListCmd, deployRollRejectCmd, deployRollIgnoreJobCmd, deployRollRetryJobCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDeployRollRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDeployRollPipeline, "delivery-pipeline", "",
			"Delivery pipeline that owns the release (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&flagDeployRollRelease, "release", "",
			"Release that owns the rollout (required)")
		_ = c.MarkFlagRequired("release")
		c.Flags().StringVar(&flagDeployRollFormat, "format", "", "Output format")
	}
	deployRollListCmd.Flags().Int64Var(&flagDeployRollPageSize, "page-size", 0, "Maximum results per page")

	deployRollAdvanceCmd.Flags().StringVar(&flagDeployRollPhaseID, "phase-id", "",
		"Phase ID to advance the rollout to (required)")
	_ = deployRollAdvanceCmd.MarkFlagRequired("phase-id")

	for _, c := range []*cobra.Command{deployRollIgnoreJobCmd, deployRollRetryJobCmd} {
		c.Flags().StringVar(&flagDeployRollPhaseID, "phase-id", "",
			"Phase ID that owns the job (required)")
		_ = c.MarkFlagRequired("phase-id")
		c.Flags().StringVar(&flagDeployRollJobID, "job-id", "",
			"Job ID within the phase (required)")
		_ = c.MarkFlagRequired("job-id")
	}
	for _, c := range []*cobra.Command{
		deployRollAdvanceCmd, deployRollApproveCmd, deployRollCancelCmd,
		deployRollRejectCmd, deployRollIgnoreJobCmd, deployRollRetryJobCmd,
	} {
		c.Flags().StringSliceVar(&flagDeployRollOverride, "override-deploy-policies", nil,
			"Deploy policies to override (repeatable)")
	}

	deployRollCmd.AddCommand(all...)
	deployCmd.AddCommand(deployRollCmd)
}

func deployRollParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	pipeline := deployChild("deliveryPipelines", flagDeployRollPipeline, deployLocationParent(project, flagDeployRollRegion))
	return deployChild("releases", flagDeployRollRelease, pipeline), nil
}

func deployRollName(id string) (string, error) {
	parent, err := deployRollParent()
	if err != nil {
		return "", err
	}
	return deployChild("rollouts", id, parent), nil
}

func runDeployRollAdvance(cmd *cobra.Command, args []string) error {
	name, err := deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Advance(name, &clouddeploy.AdvanceRolloutRequest{
		PhaseId:              flagDeployRollPhaseID,
		OverrideDeployPolicy: flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("advancing rollout: %w", err)
	}
	fmt.Printf("Advanced rollout [%s] to phase [%s].\n", args[0], flagDeployRollPhaseID)
	return nil
}

func runDeployRollApprove(cmd *cobra.Command, args []string) error {
	name, err := deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Approve(name, &clouddeploy.ApproveRolloutRequest{
		Approved:             true,
		OverrideDeployPolicy: flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("approving rollout: %w", err)
	}
	fmt.Printf("Approved rollout [%s].\n", args[0])
	return nil
}

func runDeployRollCancel(cmd *cobra.Command, args []string) error {
	name, err := deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Cancel(name, &clouddeploy.CancelRolloutRequest{
		OverrideDeployPolicy: flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling rollout: %w", err)
	}
	fmt.Printf("Cancelled rollout [%s].\n", args[0])
	return nil
}

func runDeployRollDescribe(cmd *cobra.Command, args []string) error {
	name, err := deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing rollout: %w", err)
	}
	return emitFormatted(got, flagDeployRollFormat)
}

func runDeployRollList(cmd *cobra.Command, args []string) error {
	parent, err := deployRollParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*clouddeploy.Rollout
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.List(parent).Context(ctx)
		if flagDeployRollPageSize > 0 {
			call = call.PageSize(flagDeployRollPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing rollouts: %w", err)
		}
		all = append(all, resp.Rollouts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDeployRollFormat)
}

func runDeployRollReject(cmd *cobra.Command, args []string) error {
	name, err := deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.Approve(name, &clouddeploy.ApproveRolloutRequest{
		Approved:             false,
		OverrideDeployPolicy: flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("rejecting rollout: %w", err)
	}
	fmt.Printf("Rejected rollout [%s].\n", args[0])
	return nil
}

func runDeployRollIgnoreJob(cmd *cobra.Command, args []string) error {
	name, err := deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.IgnoreJob(name, &clouddeploy.IgnoreJobRequest{
		PhaseId:              flagDeployRollPhaseID,
		JobId:                flagDeployRollJobID,
		OverrideDeployPolicy: flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("ignoring job: %w", err)
	}
	fmt.Printf("Ignored job [%s] in phase [%s] of rollout [%s].\n", flagDeployRollJobID, flagDeployRollPhaseID, args[0])
	return nil
}

func runDeployRollRetryJob(cmd *cobra.Command, args []string) error {
	name, err := deployRollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudDeployService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.Releases.Rollouts.RetryJob(name, &clouddeploy.RetryJobRequest{
		PhaseId:              flagDeployRollPhaseID,
		JobId:                flagDeployRollJobID,
		OverrideDeployPolicy: flagDeployRollOverride,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("retrying job: %w", err)
	}
	fmt.Printf("Retried job [%s] in phase [%s] of rollout [%s].\n", flagDeployRollJobID, flagDeployRollPhaseID, args[0])
	return nil
}
