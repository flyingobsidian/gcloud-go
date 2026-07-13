package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
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

// --- scheduler jobs list (#197) ---

var schedulerJobsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Cloud Scheduler jobs",
	Args:  cobra.NoArgs,
	RunE:  runSchedulerJobsList,
}

var (
	flagSchedulerListFormat string
	flagSchedulerListURI    bool
)

// --- scheduler jobs create (#198) ---

var schedulerJobsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Cloud Scheduler job",
}

var schedulerJobsCreateHTTPCmd = &cobra.Command{
	Use:   "http JOB_ID",
	Short: "Create an HTTP scheduler job",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsCreateHTTP,
}

var schedulerJobsCreatePubsubCmd = &cobra.Command{
	Use:   "pubsub JOB_ID",
	Short: "Create a Pub/Sub scheduler job",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsCreatePubsub,
}

var (
	flagSchedSchedule        string
	flagSchedTimeZone        string
	flagSchedDescription     string
	flagSchedAttemptDeadline string
	flagSchedURI             string
	flagSchedHTTPMethod      string
	flagSchedHeaders         map[string]string
	flagSchedBody            string
	flagSchedTopic           string
	flagSchedMessageBody     string
	flagSchedAttributes      map[string]string
)

// --- scheduler jobs delete (#199) ---

var schedulerJobsDeleteCmd = &cobra.Command{
	Use:   "delete JOB_ID",
	Short: "Delete a Cloud Scheduler job",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsDelete,
}


// --- scheduler jobs update (#200) ---

var schedulerJobsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Cloud Scheduler job",
}

var schedulerJobsUpdateHTTPCmd = &cobra.Command{
	Use:   "http JOB_ID",
	Short: "Update an HTTP scheduler job",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsUpdateHTTP,
}

var schedulerJobsUpdatePubsubCmd = &cobra.Command{
	Use:   "pubsub JOB_ID",
	Short: "Update a Pub/Sub scheduler job",
	Args:  cobra.ExactArgs(1),
	RunE:  runSchedulerJobsUpdatePubsub,
}

var flagSchedulerLocation string

func init() {
	for _, c := range []*cobra.Command{
		schedulerJobsDescribeCmd,
		schedulerJobsPauseCmd,
		schedulerJobsResumeCmd,
		schedulerJobsRunCmd,
		schedulerJobsListCmd,
		schedulerJobsDeleteCmd,
	} {
		c.Flags().StringVar(&flagSchedulerLocation, "location", "", "Location of the job")
	}

	schedulerJobsListCmd.Flags().StringVar(&flagSchedulerListFormat, "format", "", "Output format (e.g. json)")
	schedulerJobsListCmd.Flags().BoolVar(&flagSchedulerListURI, "uri", false, "Print resource names")

	// create common flags
	for _, c := range []*cobra.Command{schedulerJobsCreateHTTPCmd, schedulerJobsCreatePubsubCmd} {
		c.Flags().StringVar(&flagSchedulerLocation, "location", "", "Location")
		c.Flags().StringVar(&flagSchedSchedule, "schedule", "", "Cron schedule expression (required)")
		c.MarkFlagRequired("schedule")
		c.Flags().StringVar(&flagSchedTimeZone, "time-zone", "Etc/UTC", "Time zone")
		c.Flags().StringVar(&flagSchedDescription, "description", "", "Description")
		c.Flags().StringVar(&flagSchedAttemptDeadline, "attempt-deadline", "", "Attempt deadline duration")
	}
	schedulerJobsCreateHTTPCmd.Flags().StringVar(&flagSchedURI, "uri", "", "HTTP target URI (required)")
	schedulerJobsCreateHTTPCmd.MarkFlagRequired("uri")
	schedulerJobsCreateHTTPCmd.Flags().StringVar(&flagSchedHTTPMethod, "http-method", "POST", "HTTP method")
	schedulerJobsCreateHTTPCmd.Flags().StringToStringVar(&flagSchedHeaders, "headers", nil, "HTTP headers (key=value)")
	schedulerJobsCreateHTTPCmd.Flags().StringVar(&flagSchedBody, "body", "", "HTTP body")
	schedulerJobsCreatePubsubCmd.Flags().StringVar(&flagSchedTopic, "topic", "", "Pub/Sub topic (required)")
	schedulerJobsCreatePubsubCmd.MarkFlagRequired("topic")
	schedulerJobsCreatePubsubCmd.Flags().StringVar(&flagSchedMessageBody, "message-body", "", "Message body")
	schedulerJobsCreatePubsubCmd.Flags().StringToStringVar(&flagSchedAttributes, "attributes", nil, "Message attributes")
	schedulerJobsCreateCmd.AddCommand(schedulerJobsCreateHTTPCmd)
	schedulerJobsCreateCmd.AddCommand(schedulerJobsCreatePubsubCmd)

	// update flags
	for _, c := range []*cobra.Command{schedulerJobsUpdateHTTPCmd, schedulerJobsUpdatePubsubCmd} {
		c.Flags().StringVar(&flagSchedulerLocation, "location", "", "Location")
		c.Flags().StringVar(&flagSchedSchedule, "schedule", "", "Cron schedule expression")
		c.Flags().StringVar(&flagSchedTimeZone, "time-zone", "", "Time zone")
		c.Flags().StringVar(&flagSchedDescription, "description", "", "Description")
		c.Flags().StringVar(&flagSchedAttemptDeadline, "attempt-deadline", "", "Attempt deadline duration")
	}
	schedulerJobsUpdateHTTPCmd.Flags().StringVar(&flagSchedURI, "uri", "", "HTTP target URI")
	schedulerJobsUpdateHTTPCmd.Flags().StringVar(&flagSchedHTTPMethod, "http-method", "", "HTTP method")
	schedulerJobsUpdateHTTPCmd.Flags().StringToStringVar(&flagSchedHeaders, "headers", nil, "HTTP headers")
	schedulerJobsUpdateHTTPCmd.Flags().StringVar(&flagSchedBody, "body", "", "HTTP body")
	schedulerJobsUpdatePubsubCmd.Flags().StringVar(&flagSchedTopic, "topic", "", "Pub/Sub topic")
	schedulerJobsUpdatePubsubCmd.Flags().StringVar(&flagSchedMessageBody, "message-body", "", "Message body")
	schedulerJobsUpdatePubsubCmd.Flags().StringToStringVar(&flagSchedAttributes, "attributes", nil, "Message attributes")
	schedulerJobsUpdateCmd.AddCommand(schedulerJobsUpdateHTTPCmd)
	schedulerJobsUpdateCmd.AddCommand(schedulerJobsUpdatePubsubCmd)

	schedulerJobsCmd.AddCommand(schedulerJobsDescribeCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsListCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsPauseCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsResumeCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsRunCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsCreateCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsDeleteCmd)
	schedulerJobsCmd.AddCommand(schedulerJobsUpdateCmd)
	schedulerCmd.AddCommand(schedulerJobsCmd)

	// gcloud-python scheduler subgroups not yet implemented (#545).
	registerStubGroup(schedulerCmd, "cmek-config",
		"Manage Customer-Managed Encryption Key configuration",
		"describe", "update", "clear")
	registerStubGroup(schedulerCmd, "locations",
		"Manage / list Cloud Scheduler locations",
		"list", "describe")
	registerStubGroup(schedulerCmd, "operations",
		"Manage long-running scheduler operations",
		"cancel", "delete", "describe", "list")

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

// --- scheduler jobs list (#197) ---

func runSchedulerJobsList(cmd *cobra.Command, args []string) error {
	project, location, err := resolveSchedulerLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	var allJobs []*cloudscheduler.Job
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing scheduler jobs: %w", err)
		}
		allJobs = append(allJobs, resp.Jobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagSchedulerListURI {
		for _, job := range allJobs {
			fmt.Println(job.Name)
		}
		return nil
	}

	if flagSchedulerListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allJobs)
	}

	fmt.Printf("%-30s %-12s %-25s %-15s %-25s\n", "JOB_ID", "STATE", "SCHEDULE", "TARGET_TYPE", "NEXT_RUN")
	for _, j := range allJobs {
		jobID := path.Base(j.Name)
		targetType := ""
		if j.HttpTarget != nil {
			targetType = "HTTP"
		} else if j.PubsubTarget != nil {
			targetType = "PUBSUB"
		} else if j.AppEngineHttpTarget != nil {
			targetType = "APP_ENGINE"
		}
		fmt.Printf("%-30s %-12s %-25s %-15s %-25s\n", jobID, j.State, j.Schedule, targetType, j.ScheduleTime)
	}
	return nil
}

// --- scheduler jobs create http (#198) ---

func runSchedulerJobsCreateHTTP(cmd *cobra.Command, args []string) error {
	project, location, err := resolveSchedulerLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	job := &cloudscheduler.Job{
		Schedule: flagSchedSchedule,
		TimeZone: flagSchedTimeZone,
		HttpTarget: &cloudscheduler.HttpTarget{
			Uri:        flagSchedURI,
			HttpMethod: strings.ToUpper(flagSchedHTTPMethod),
		},
	}
	if flagSchedDescription != "" {
		job.Description = flagSchedDescription
	}
	if flagSchedAttemptDeadline != "" {
		job.AttemptDeadline = flagSchedAttemptDeadline
	}
	if len(flagSchedHeaders) > 0 {
		job.HttpTarget.Headers = flagSchedHeaders
	}
	if flagSchedBody != "" {
		job.HttpTarget.Body = flagSchedBody
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	job.Name = fmt.Sprintf("%s/jobs/%s", parent, args[0])
	result, err := svc.Projects.Locations.Jobs.Create(parent, job).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating scheduler job: %w", err)
	}

	fmt.Printf("Created scheduler job [%s].\n", path.Base(result.Name))
	return nil
}

// --- scheduler jobs create pubsub (#198) ---

func runSchedulerJobsCreatePubsub(cmd *cobra.Command, args []string) error {
	project, location, err := resolveSchedulerLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	job := &cloudscheduler.Job{
		Schedule: flagSchedSchedule,
		TimeZone: flagSchedTimeZone,
		PubsubTarget: &cloudscheduler.PubsubTarget{
			TopicName: flagSchedTopic,
		},
	}
	if flagSchedDescription != "" {
		job.Description = flagSchedDescription
	}
	if flagSchedAttemptDeadline != "" {
		job.AttemptDeadline = flagSchedAttemptDeadline
	}
	if flagSchedMessageBody != "" {
		job.PubsubTarget.Data = flagSchedMessageBody
	}
	if len(flagSchedAttributes) > 0 {
		job.PubsubTarget.Attributes = flagSchedAttributes
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	job.Name = fmt.Sprintf("%s/jobs/%s", parent, args[0])
	result, err := svc.Projects.Locations.Jobs.Create(parent, job).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating scheduler job: %w", err)
	}

	fmt.Printf("Created scheduler job [%s].\n", path.Base(result.Name))
	return nil
}

// --- scheduler jobs delete (#199) ---

func runSchedulerJobsDelete(cmd *cobra.Command, args []string) error {
	project, location, err := resolveSchedulerLocation()
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Printf("You are about to delete scheduler job [%s].\n", args[0])
		fmt.Print("Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, args[0])
	if _, err := svc.Projects.Locations.Jobs.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting scheduler job: %w", err)
	}

	fmt.Printf("Deleted scheduler job [%s].\n", args[0])
	return nil
}

// --- scheduler jobs update http (#200) ---

func runSchedulerJobsUpdateHTTP(cmd *cobra.Command, args []string) error {
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
	job, err := svc.Projects.Locations.Jobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting scheduler job: %w", err)
	}

	var updateMask []string
	if flagSchedSchedule != "" {
		job.Schedule = flagSchedSchedule
		updateMask = append(updateMask, "schedule")
	}
	if flagSchedTimeZone != "" {
		job.TimeZone = flagSchedTimeZone
		updateMask = append(updateMask, "time_zone")
	}
	if flagSchedDescription != "" {
		job.Description = flagSchedDescription
		updateMask = append(updateMask, "description")
	}
	if flagSchedURI != "" && job.HttpTarget != nil {
		job.HttpTarget.Uri = flagSchedURI
		updateMask = append(updateMask, "http_target.uri")
	}
	if flagSchedHTTPMethod != "" && job.HttpTarget != nil {
		job.HttpTarget.HttpMethod = strings.ToUpper(flagSchedHTTPMethod)
		updateMask = append(updateMask, "http_target.http_method")
	}

	if len(updateMask) == 0 {
		return fmt.Errorf("no update flags specified")
	}

	result, err := svc.Projects.Locations.Jobs.Patch(name, job).UpdateMask(strings.Join(updateMask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating scheduler job: %w", err)
	}

	fmt.Printf("Updated scheduler job [%s].\n", path.Base(result.Name))
	return nil
}

// --- scheduler jobs update pubsub (#200) ---

func runSchedulerJobsUpdatePubsub(cmd *cobra.Command, args []string) error {
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
	job, err := svc.Projects.Locations.Jobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting scheduler job: %w", err)
	}

	var updateMask []string
	if flagSchedSchedule != "" {
		job.Schedule = flagSchedSchedule
		updateMask = append(updateMask, "schedule")
	}
	if flagSchedTimeZone != "" {
		job.TimeZone = flagSchedTimeZone
		updateMask = append(updateMask, "time_zone")
	}
	if flagSchedDescription != "" {
		job.Description = flagSchedDescription
		updateMask = append(updateMask, "description")
	}
	if flagSchedTopic != "" && job.PubsubTarget != nil {
		job.PubsubTarget.TopicName = flagSchedTopic
		updateMask = append(updateMask, "pubsub_target.topic_name")
	}

	if len(updateMask) == 0 {
		return fmt.Errorf("no update flags specified")
	}

	result, err := svc.Projects.Locations.Jobs.Patch(name, job).UpdateMask(strings.Join(updateMask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating scheduler job: %w", err)
	}

	fmt.Printf("Updated scheduler job [%s].\n", path.Base(result.Name))
	return nil
}
