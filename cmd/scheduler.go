package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-golang-cli/internal/gcp"
	"github.com/spf13/cobra"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
)

var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Manage Cloud Scheduler",
}

var schedulerJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage Cloud Scheduler jobs",
}

var schedulerJobsDescribeCmd = &cobra.Command{
	Use:   "describe JOB_ID",
	Short: "Describe a Cloud Scheduler job",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsDescribe,
}

var schedulerJobsPauseCmd = &cobra.Command{
	Use:   "pause JOB_ID",
	Short: "Pause a Cloud Scheduler job",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsPause,
}

var schedulerJobsResumeCmd = &cobra.Command{
	Use:   "resume JOB_ID",
	Short: "Resume a Cloud Scheduler job",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsResume,
}

var schedulerJobsRunCmd = &cobra.Command{
	Use:   "run JOB_ID",
	Short: "Trigger a Cloud Scheduler job to run immediately",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsRun,
}

var flagSchedulerLocation string

func init() {
	for _, c := range []*cobra.Command{
		schedulerJobsDescribeCmd,
		schedulerJobsPauseCmd,
		schedulerJobsResumeCmd,
		schedulerJobsRunCmd,
	} {
		c.Flags().StringVar(&flagSchedulerLocation, "location", "", "Location of the job")
	}

	schedulerJobsCmd.AddCommand(schedulerJobsDescribeCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsPauseCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsResumeCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsRunCmd)
	schedulerCmd.AddCommand(schedulerJobsCmd)
	rootCmd.AddCommand(schedulerCmd)
}

func schedulerJobName(project, location, jobID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project, location, jobID)
}

func resolveSchedulerLocation() (string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", err
	}
	location := flagSchedulerLocation
	if location == "" {
		_, region, err := resolveRegion()
		if err != nil || region == "" {
			return "", "", fmt.Errorf("--location is required; set via --location flag or compute/region config")
		}
		location = region
	}
	return project, location, nil
}

func runSchedulerJobsDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := resolveSchedulerLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	job, err := svc.Projects.Locations.Jobs.Get(schedulerJobName(project, location, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing scheduler job: %w", err)
	}

	return formatOutput(job, "")
}

func runSchedulerJobsPause(cmd *cobra.Command, args []string) error {
	project, location, err := resolveSchedulerLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, args[0])
	_, err = svc.Projects.Locations.Jobs.Pause(name, &cloudscheduler.PauseJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("pausing scheduler job: %w", err)
	}

	fmt.Printf("Paused job [%s].\n", args[0])
	return nil
}

func runSchedulerJobsResume(cmd *cobra.Command, args []string) error {
	project, location, err := resolveSchedulerLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, args[0])
	_, err = svc.Projects.Locations.Jobs.Resume(name, &cloudscheduler.ResumeJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming scheduler job: %w", err)
	}

	fmt.Printf("Resumed job [%s].\n", args[0])
	return nil
}

func runSchedulerJobsRun(cmd *cobra.Command, args []string) error {
	project, location, err := resolveSchedulerLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, args[0])
	_, err = svc.Projects.Locations.Jobs.Run(name, &cloudscheduler.RunJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running scheduler job: %w", err)
	}

	fmt.Printf("Triggered job [%s].\n", args[0])
	return nil
}
