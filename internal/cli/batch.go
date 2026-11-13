package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	batchapi "google.golang.org/api/batch/v1"
)

// --- gcloud batch (#304) ---

func batchLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func batchChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// --- jobs ---

// --- tasks ---

func (s *batchCmds) initBatch() {
	// jobs
	for _, c := range []*cobra.Command{s.batchJobCancelCmd, s.batchJobDeleteCmd, s.batchJobDescribeCmd, s.batchJobListCmd, s.batchJobSubmitCmd} {
		c.Flags().StringVar(&s.flagBatchLocation, "location", "", "Location containing the job (required)")
		_ = c.MarkFlagRequired("location")
	}
	s.batchJobSubmitCmd.Flags().StringVar(&s.flagBatchConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Job message body (required)")
	_ = s.batchJobSubmitCmd.MarkFlagRequired("config-file")
	s.batchJobDescribeCmd.Flags().StringVar(&s.flagBatchFormat, "format", "", "Output format")
	s.batchJobListCmd.Flags().StringVar(&s.flagBatchFormat, "format", "", "Output format")
	s.batchJobsCmd.AddCommand(s.batchJobCancelCmd, s.batchJobDeleteCmd, s.batchJobDescribeCmd, s.batchJobListCmd, s.batchJobSubmitCmd)
	s.batchCmd.AddCommand(s.batchJobsCmd)

	// tasks
	for _, c := range []*cobra.Command{s.batchTaskDescribeCmd, s.batchTaskListCmd} {
		c.Flags().StringVar(&s.flagBatchLocation, "location", "", "Location containing the job (required)")
		c.Flags().StringVar(&s.flagBatchJob, "job", "", "Job containing the task (required)")
		c.Flags().StringVar(&s.flagBatchTaskGroup, "task-group", "group0", "Task group name (default: group0)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("job")
	}
	s.batchTaskDescribeCmd.Flags().StringVar(&s.flagBatchFormat, "format", "", "Output format")
	s.batchTaskListCmd.Flags().StringVar(&s.flagBatchFormat, "format", "", "Output format")
	s.batchTasksCmd.AddCommand(s.batchTaskDescribeCmd, s.batchTaskListCmd)
	s.batchCmd.AddCommand(s.batchTasksCmd)

}

// --- jobs impl ---

func batchJobName(id, project, location string) string {
	return batchChild("jobs", id, batchLocationParent(project, location))
}

func (s *batchCmds) runBatchJobCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Batch(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Cancel(batchJobName(args[0], project, s.flagBatchLocation), &batchapi.CancelJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled job [%s].\n", args[0])
	return nil
}

func (s *batchCmds) runBatchJobDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Batch(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Delete(batchJobName(args[0], project, s.flagBatchLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted job [%s].\n", args[0])
	return nil
}

func (s *batchCmds) runBatchJobDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Batch(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Get(batchJobName(args[0], project, s.flagBatchLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBatchFormat)
}

func (s *batchCmds) runBatchJobList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Batch(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Jobs.List(batchLocationParent(project, s.flagBatchLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing jobs: %w", err)
	}
	if s.flagBatchFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Jobs, s.flagBatchFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Jobs), "%-40s %s\n", "NAME", "STATE")
	for _, j := range resp.Jobs {
		state := ""
		if j.Status != nil {
			state = j.Status.State
		}
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(j.Name), state)
	}
	return nil
}

func (s *batchCmds) runBatchJobSubmit(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	job := &batchapi.Job{}
	if err := loadYAMLOrJSONInto(s.flagBatchConfigFile, job); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Batch(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Create(batchLocationParent(project, s.flagBatchLocation), job).
		JobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- tasks impl ---

func batchTaskParent(project, location, job, group string) string {
	return fmt.Sprintf("%s/taskGroups/%s", batchJobName(job, project, location), group)
}

func batchTaskName(id, project, location, job, group string) string {
	return batchChild("tasks", id, batchTaskParent(project, location, job, group))
}

func (s *batchCmds) runBatchTaskDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Batch(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.TaskGroups.Tasks.Get(batchTaskName(args[0], project, s.flagBatchLocation, s.flagBatchJob, s.flagBatchTaskGroup)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing task: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBatchFormat)
}

func (s *batchCmds) runBatchTaskList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Batch(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Jobs.TaskGroups.Tasks.List(batchTaskParent(project, s.flagBatchLocation, s.flagBatchJob, s.flagBatchTaskGroup)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing tasks: %w", err)
	}
	if s.flagBatchFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Tasks, s.flagBatchFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Tasks), "%-40s %s\n", "NAME", "STATE")
	for _, t := range resp.Tasks {
		state := ""
		if t.Status != nil {
			state = t.Status.State
		}
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(t.Name), state)
	}
	return nil
}

// batchCmds holds the commands and flag values of one invocation's "batch"
// command group, so invocations don't share state (#2070).
type batchCmds struct {
	inv *invocation

	batchCmd             *cobra.Command
	flagBatchLocation    string
	flagBatchConfigFile  string
	flagBatchFormat      string
	flagBatchJob         string
	flagBatchTaskGroup   string
	batchJobsCmd         *cobra.Command
	batchJobCancelCmd    *cobra.Command
	batchJobDeleteCmd    *cobra.Command
	batchJobDescribeCmd  *cobra.Command
	batchJobListCmd      *cobra.Command
	batchJobSubmitCmd    *cobra.Command
	batchTasksCmd        *cobra.Command
	batchTaskDescribeCmd *cobra.Command
	batchTaskListCmd     *cobra.Command
}

// newBatchCmd returns the "batch" command group for inv.
func newBatchCmd(inv *invocation) *cobra.Command {
	return newBatchCmds(inv).batchCmd
}

// newBatchCmds builds the "batch" command group for inv and returns its state.
func newBatchCmds(inv *invocation) *batchCmds {
	s := &batchCmds{inv: inv}
	s.batchCmd = &cobra.Command{
		Use:   "batch",
		Short: "Manage Batch jobs and tasks",
	}
	s.batchJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage Batch jobs"}
	s.batchJobCancelCmd = &cobra.Command{
		Use: "cancel JOB", Short: "Cancel a Batch job",
		Args: cobra.ExactArgs(1), RunE: s.runBatchJobCancel,
	}
	s.batchJobDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Batch job",
		Args: cobra.ExactArgs(1), RunE: s.runBatchJobDelete,
	}
	s.batchJobDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Batch job",
		Args: cobra.ExactArgs(1), RunE: s.runBatchJobDescribe,
	}
	s.batchJobListCmd = &cobra.Command{
		Use: "list", Short: "List Batch jobs in a location",
		Args: cobra.NoArgs, RunE: s.runBatchJobList,
	}
	s.batchJobSubmitCmd = &cobra.Command{
		Use: "submit JOB", Short: "Submit a Batch job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBatchJobSubmit,
	}
	s.batchTasksCmd = &cobra.Command{Use: "tasks", Short: "Manage Batch tasks"}
	s.batchTaskDescribeCmd = &cobra.Command{
		Use: "describe TASK", Short: "Describe a Batch task",
		Args: cobra.ExactArgs(1), RunE: s.runBatchTaskDescribe,
	}
	s.batchTaskListCmd = &cobra.Command{
		Use: "list", Short: "List Batch tasks in a job task group",
		Args: cobra.NoArgs, RunE: s.runBatchTaskList,
	}
	s.initBatch()
	return s
}
