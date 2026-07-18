package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc jobs (#1513) ---

var dpJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage Dataproc jobs"}

var (
	flagDPJobRegion      string
	flagDPJobFormat      string
	flagDPJobConfigFile  string
	flagDPJobUpdateMask  string
	flagDPJobRequestID   string
	flagDPJobCluster     string
	flagDPJobFilter      string
	flagDPJobStateMatch  string
	flagDPJobPageSize    int64
	flagDPJobWaitTimeout time.Duration
)

var (
	dpJobDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: runDPJobDelete,
	}
	dpJobDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: runDPJobDescribe,
	}
	dpJobKillCmd = &cobra.Command{
		Use: "kill JOB", Short: "Kill (cancel) a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: runDPJobKill,
	}
	dpJobListCmd = &cobra.Command{
		Use: "list", Short: "List Dataproc jobs",
		Args: cobra.NoArgs, RunE: runDPJobList,
	}
	dpJobSubmitCmd = &cobra.Command{
		Use: "submit", Short: "Submit a Dataproc job (body loaded from --config-file)",
		Args: cobra.NoArgs, RunE: runDPJobSubmit,
	}
	dpJobUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: runDPJobUpdate,
	}
	dpJobWaitCmd = &cobra.Command{
		Use: "wait JOB", Short: "Wait for a job to reach a terminal state",
		Args: cobra.ExactArgs(1), RunE: runDPJobWait,
	}
	dpJobGetIamCmd = &cobra.Command{
		Use: "get-iam-policy JOB", Short: "Get the IAM policy for a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: runDPJobGetIam,
	}
	dpJobSetIamCmd = &cobra.Command{
		Use: "set-iam-policy JOB POLICY_FILE", Short: "Set the IAM policy for a Dataproc job",
		Args: cobra.ExactArgs(2), RunE: runDPJobSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		dpJobDeleteCmd, dpJobDescribeCmd, dpJobKillCmd, dpJobListCmd, dpJobSubmitCmd,
		dpJobUpdateCmd, dpJobWaitCmd, dpJobGetIamCmd, dpJobSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDPJobRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDPJobFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dpJobSubmitCmd, dpJobUpdateCmd} {
		c.Flags().StringVar(&flagDPJobConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Job body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dpJobSubmitCmd.Flags().StringVar(&flagDPJobRequestID, "request-id", "", "Optional idempotency ID")
	dpJobUpdateCmd.Flags().StringVar(&flagDPJobUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dpJobListCmd.Flags().StringVar(&flagDPJobCluster, "cluster", "", "Restrict to jobs on the named cluster")
	dpJobListCmd.Flags().StringVar(&flagDPJobFilter, "filter", "", "Server-side filter expression")
	dpJobListCmd.Flags().StringVar(&flagDPJobStateMatch, "state-filter", "",
		"JobStateMatcher (ALL, ACTIVE, NON_ACTIVE)")
	dpJobListCmd.Flags().Int64Var(&flagDPJobPageSize, "page-size", 0, "Maximum results per page")
	dpJobWaitCmd.Flags().DurationVar(&flagDPJobWaitTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the job to reach a terminal state")

	dpJobsCmd.AddCommand(all...)
	dataprocCmd.AddCommand(dpJobsCmd)
}

func dpJobResourceName(project, region, jobID string) string {
	return fmt.Sprintf("projects/%s/regions/%s/jobs/%s", project, region, jobID)
}

func runDPJobDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.Jobs.Delete(project, flagDPJobRegion, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Printf("Deleted job [%s].\n", args[0])
	return nil
}

func runDPJobDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Jobs.Get(project, flagDPJobRegion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormatted(got, flagDPJobFormat)
}

func runDPJobKill(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Jobs.Cancel(project, flagDPJobRegion, args[0], &dataproc.CancelJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("killing job: %w", err)
	}
	fmt.Printf("Cancel request issued for job [%s].\n", args[0])
	return emitFormatted(got, flagDPJobFormat)
}

func runDPJobList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.Job
	pageToken := ""
	for {
		call := svc.Projects.Regions.Jobs.List(project, flagDPJobRegion).Context(ctx)
		if flagDPJobCluster != "" {
			call = call.ClusterName(flagDPJobCluster)
		}
		if flagDPJobFilter != "" {
			call = call.Filter(flagDPJobFilter)
		}
		if flagDPJobStateMatch != "" {
			call = call.JobStateMatcher(flagDPJobStateMatch)
		}
		if flagDPJobPageSize > 0 {
			call = call.PageSize(flagDPJobPageSize)
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
	return emitFormatted(all, flagDPJobFormat)
}

func runDPJobSubmit(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	job := &dataproc.Job{}
	if err := loadYAMLOrJSONInto(flagDPJobConfigFile, job); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Jobs.Submit(project, flagDPJobRegion, &dataproc.SubmitJobRequest{
		Job:       job,
		RequestId: flagDPJobRequestID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting job: %w", err)
	}
	if got.Reference != nil {
		fmt.Printf("Submitted job [%s].\n", got.Reference.JobId)
	} else {
		fmt.Println("Submitted job.")
	}
	return emitFormatted(got, flagDPJobFormat)
}

func runDPJobUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dataproc.Job{}
	if err := loadYAMLOrJSONInto(flagDPJobConfigFile, body); err != nil {
		return err
	}
	mask := flagDPJobUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.Jobs.Patch(project, flagDPJobRegion, args[0], body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	fmt.Printf("Updated job [%s].\n", args[0])
	return emitFormatted(got, flagDPJobFormat)
}

func runDPJobWait(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), flagDPJobWaitTimeout)
	defer cancel()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	backoff := 5 * time.Second
	for {
		got, err := svc.Projects.Regions.Jobs.Get(project, flagDPJobRegion, args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling job: %w", err)
		}
		state := ""
		if got.Status != nil {
			state = got.Status.State
		}
		switch state {
		case "DONE":
			fmt.Printf("Job [%s] succeeded.\n", args[0])
			return emitFormatted(got, flagDPJobFormat)
		case "CANCELLED", "ERROR":
			return fmt.Errorf("job [%s] ended in state %s", args[0], state)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for job %s: %w", args[0], ctx.Err())
		case <-time.After(backoff):
		}
		if backoff < 60*time.Second {
			backoff = time.Duration(float64(backoff) * 1.5)
		}
	}
}

func runDPJobGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := dpJobResourceName(project, flagDPJobRegion, args[0])
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.Jobs.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDPJobFormat)
}

func runDPJobSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := dpJobResourceName(project, flagDPJobRegion, args[0])
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPJobRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.Jobs.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dpUpdatedIam(fmt.Sprintf("job [%s]", args[0]))
	return emitFormatted(updated, flagDPJobFormat)
}
