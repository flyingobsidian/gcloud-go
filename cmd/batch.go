package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	batchapi "google.golang.org/api/batch/v1"
)

// --- gcloud batch (#304) ---

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Manage Batch jobs and tasks",
}

func batchLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func batchChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

var (
	flagBatchLocation   string
	flagBatchConfigFile string
	flagBatchFormat     string
	flagBatchJob        string
	flagBatchTaskGroup  string
)

// --- jobs ---

var batchJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage Batch jobs"}

var (
	batchJobCancelCmd = &cobra.Command{
		Use: "cancel JOB", Short: "Cancel a Batch job",
		Args: cobra.ExactArgs(1), RunE: runBatchJobCancel,
	}
	batchJobDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Batch job",
		Args: cobra.ExactArgs(1), RunE: runBatchJobDelete,
	}
	batchJobDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Batch job",
		Args: cobra.ExactArgs(1), RunE: runBatchJobDescribe,
	}
	batchJobListCmd = &cobra.Command{
		Use: "list", Short: "List Batch jobs in a location",
		Args: cobra.NoArgs, RunE: runBatchJobList,
	}
	batchJobSubmitCmd = &cobra.Command{
		Use: "submit JOB", Short: "Submit a Batch job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBatchJobSubmit,
	}
)

// --- tasks ---

var batchTasksCmd = &cobra.Command{Use: "tasks", Short: "Manage Batch tasks"}

var (
	batchTaskDescribeCmd = &cobra.Command{
		Use: "describe TASK", Short: "Describe a Batch task",
		Args: cobra.ExactArgs(1), RunE: runBatchTaskDescribe,
	}
	batchTaskListCmd = &cobra.Command{
		Use: "list", Short: "List Batch tasks in a job task group",
		Args: cobra.NoArgs, RunE: runBatchTaskList,
	}
)

func init() {
	// jobs
	for _, c := range []*cobra.Command{batchJobCancelCmd, batchJobDeleteCmd, batchJobDescribeCmd, batchJobListCmd, batchJobSubmitCmd} {
		c.Flags().StringVar(&flagBatchLocation, "location", "", "Location containing the job (required)")
		_ = c.MarkFlagRequired("location")
	}
	batchJobSubmitCmd.Flags().StringVar(&flagBatchConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Job message body (required)")
	_ = batchJobSubmitCmd.MarkFlagRequired("config-file")
	batchJobDescribeCmd.Flags().StringVar(&flagBatchFormat, "format", "", "Output format")
	batchJobListCmd.Flags().StringVar(&flagBatchFormat, "format", "", "Output format")
	batchJobsCmd.AddCommand(batchJobCancelCmd, batchJobDeleteCmd, batchJobDescribeCmd, batchJobListCmd, batchJobSubmitCmd)
	batchCmd.AddCommand(batchJobsCmd)

	// tasks
	for _, c := range []*cobra.Command{batchTaskDescribeCmd, batchTaskListCmd} {
		c.Flags().StringVar(&flagBatchLocation, "location", "", "Location containing the job (required)")
		c.Flags().StringVar(&flagBatchJob, "job", "", "Job containing the task (required)")
		c.Flags().StringVar(&flagBatchTaskGroup, "task-group", "group0", "Task group name (default: group0)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("job")
	}
	batchTaskDescribeCmd.Flags().StringVar(&flagBatchFormat, "format", "", "Output format")
	batchTaskListCmd.Flags().StringVar(&flagBatchFormat, "format", "", "Output format")
	batchTasksCmd.AddCommand(batchTaskDescribeCmd, batchTaskListCmd)
	batchCmd.AddCommand(batchTasksCmd)

	rootCmd.AddCommand(batchCmd)
}

// --- jobs impl ---

func batchJobName(id, project, location string) string {
	return batchChild("jobs", id, batchLocationParent(project, location))
}

func runBatchJobCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BatchService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Cancel(batchJobName(args[0], project, flagBatchLocation), &batchapi.CancelJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling job: %w", err)
	}
	fmt.Printf("Cancelled job [%s].\n", args[0])
	return nil
}

func runBatchJobDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BatchService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Delete(batchJobName(args[0], project, flagBatchLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Printf("Deleted job [%s].\n", args[0])
	return nil
}

func runBatchJobDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BatchService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Get(batchJobName(args[0], project, flagBatchLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormatted(got, flagBatchFormat)
}

func runBatchJobList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BatchService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Jobs.List(batchLocationParent(project, flagBatchLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing jobs: %w", err)
	}
	if flagBatchFormat != "" {
		return emitFormatted(resp.Jobs, flagBatchFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, j := range resp.Jobs {
		state := ""
		if j.Status != nil {
			state = j.Status.State
		}
		fmt.Printf("%-40s %s\n", path.Base(j.Name), state)
	}
	return nil
}

func runBatchJobSubmit(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	job := &batchapi.Job{}
	if err := loadYAMLOrJSONInto(flagBatchConfigFile, job); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BatchService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Create(batchLocationParent(project, flagBatchLocation), job).
		JobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting job: %w", err)
	}
	return emitFormatted(got, "")
}

// --- tasks impl ---

func batchTaskParent(project, location, job, group string) string {
	return fmt.Sprintf("%s/taskGroups/%s", batchJobName(job, project, location), group)
}

func batchTaskName(id, project, location, job, group string) string {
	return batchChild("tasks", id, batchTaskParent(project, location, job, group))
}

func runBatchTaskDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BatchService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.TaskGroups.Tasks.Get(batchTaskName(args[0], project, flagBatchLocation, flagBatchJob, flagBatchTaskGroup)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing task: %w", err)
	}
	return emitFormatted(got, flagBatchFormat)
}

func runBatchTaskList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BatchService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Jobs.TaskGroups.Tasks.List(batchTaskParent(project, flagBatchLocation, flagBatchJob, flagBatchTaskGroup)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing tasks: %w", err)
	}
	if flagBatchFormat != "" {
		return emitFormatted(resp.Tasks, flagBatchFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, t := range resp.Tasks {
		state := ""
		if t.Status != nil {
			state = t.Status.State
		}
		fmt.Printf("%-40s %s\n", path.Base(t.Name), state)
	}
	return nil
}
