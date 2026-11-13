package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc jobs (#1513) ---

func (s *dataprocCmds) initDataprocJobs() {
	all := []*cobra.Command{
		s.dpJobDeleteCmd, s.dpJobDescribeCmd, s.dpJobKillCmd, s.dpJobListCmd, s.dpJobSubmitCmd,
		s.dpJobUpdateCmd, s.dpJobWaitCmd, s.dpJobGetIamCmd, s.dpJobSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDPJobRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDPJobFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dpJobSubmitCmd, s.dpJobUpdateCmd} {
		c.Flags().StringVar(&s.flagDPJobConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Job body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dpJobSubmitCmd.Flags().StringVar(&s.flagDPJobRequestID, "request-id", "", "Optional idempotency ID")
	s.dpJobUpdateCmd.Flags().StringVar(&s.flagDPJobUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dpJobListCmd.Flags().StringVar(&s.flagDPJobCluster, "cluster", "", "Restrict to jobs on the named cluster")
	s.dpJobListCmd.Flags().StringVar(&s.flagDPJobFilter, "filter", "", "Server-side filter expression")
	s.dpJobListCmd.Flags().StringVar(&s.flagDPJobStateMatch, "state-filter", "",
		"JobStateMatcher (ALL, ACTIVE, NON_ACTIVE)")
	s.dpJobListCmd.Flags().Int64Var(&s.flagDPJobPageSize, "page-size", 0, "Maximum results per page")
	s.dpJobWaitCmd.Flags().DurationVar(&s.flagDPJobWaitTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the job to reach a terminal state")

	s.dpJobsCmd.AddCommand(all...)
	s.dataprocCmd.AddCommand(s.dpJobsCmd)
}

func dpJobResourceName(project, region, jobID string) string {
	return fmt.Sprintf("projects/%s/regions/%s/jobs/%s", project, region, jobID)
}

func (s *dataprocCmds) runDPJobDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.Jobs.Delete(project, s.flagDPJobRegion, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted job [%s].\n", args[0])
	return nil
}

func (s *dataprocCmds) runDPJobDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Jobs.Get(project, s.flagDPJobRegion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPJobFormat)
}

func (s *dataprocCmds) runDPJobKill(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Jobs.Cancel(project, s.flagDPJobRegion, args[0], &dataproc.CancelJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("killing job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for job [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPJobFormat)
}

func (s *dataprocCmds) runDPJobList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.Job
	pageToken := ""
	for {
		call := svc.Projects.Regions.Jobs.List(project, s.flagDPJobRegion).Context(ctx)
		if s.flagDPJobCluster != "" {
			call = call.ClusterName(s.flagDPJobCluster)
		}
		if s.flagDPJobFilter != "" {
			call = call.Filter(s.flagDPJobFilter)
		}
		if s.flagDPJobStateMatch != "" {
			call = call.JobStateMatcher(s.flagDPJobStateMatch)
		}
		if s.flagDPJobPageSize > 0 {
			call = call.PageSize(s.flagDPJobPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDPJobFormat)
}

func (s *dataprocCmds) runDPJobSubmit(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	job := &dataproc.Job{}
	if err := loadYAMLOrJSONInto(s.flagDPJobConfigFile, job); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Jobs.Submit(project, s.flagDPJobRegion, &dataproc.SubmitJobRequest{
		Job:       job,
		RequestId: s.flagDPJobRequestID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting job: %w", err)
	}
	if got.Reference != nil {
		fmt.Fprintf(s.inv.stdout, "Submitted job [%s].\n", got.Reference.JobId)
	} else {
		fmt.Fprintln(s.inv.stdout, "Submitted job.")
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPJobFormat)
}

func (s *dataprocCmds) runDPJobUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dataproc.Job{}
	if err := loadYAMLOrJSONInto(s.flagDPJobConfigFile, body); err != nil {
		return err
	}
	mask := s.flagDPJobUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.Jobs.Patch(project, s.flagDPJobRegion, args[0], body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated job [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPJobFormat)
}

func (s *dataprocCmds) runDPJobWait(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), s.flagDPJobWaitTimeout)
	defer cancel()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	backoff := 5 * time.Second
	for {
		got, err := svc.Projects.Regions.Jobs.Get(project, s.flagDPJobRegion, args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling job: %w", err)
		}
		state := ""
		if got.Status != nil {
			state = got.Status.State
		}
		switch state {
		case "DONE":
			fmt.Fprintf(s.inv.stdout, "Job [%s] succeeded.\n", args[0])
			return emitFormattedTo(s.inv.stdout, got, s.flagDPJobFormat)
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

func (s *dataprocCmds) runDPJobGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	name := dpJobResourceName(project, s.flagDPJobRegion, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.Jobs.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDPJobFormat)
}

func (s *dataprocCmds) runDPJobSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	name := dpJobResourceName(project, s.flagDPJobRegion, args[0])
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPJobRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.Jobs.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dpUpdatedIam(fmt.Sprintf("job [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDPJobFormat)
}
