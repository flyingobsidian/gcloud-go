package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
	"google.golang.org/api/googleapi"
)

// schedulerJobFlags holds the flag values of the `scheduler jobs` commands.
// The create and update commands each have their own, so that a flag's
// default on one doesn't overwrite its default on the other (#2081).
type schedulerJobFlags struct {
	location        string
	listURI         bool
	schedule        string
	timeZone        string
	description     string
	attemptDeadline string
	uri             string
	httpMethod      string
	headers         map[string]string
	body            string
	topic           string
	messageBody     string
	attributes      map[string]string
}

func newSchedulerCmd(inv *invocation) *cobra.Command {
	schedulerCmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Manage Cloud Scheduler",
	}
	schedulerCmd.AddCommand(newSchedulerJobsCmd(inv))
	// cmek-config, locations, operations subgroups are implemented in
	// scheduler_extras.go.
	addSchedulerExtras(inv, schedulerCmd)
	return schedulerCmd
}

func newSchedulerJobsCmd(inv *invocation) *cobra.Command {
	f, cf, uf := &schedulerJobFlags{}, &schedulerJobFlags{}, &schedulerJobFlags{}
	run := func(f *schedulerJobFlags, fn func(context.Context, *invocation, *schedulerJobFlags, string) error) func(*cobra.Command, []string) error {
		return func(c *cobra.Command, args []string) error {
			return fn(c.Context(), inv, f, args[0])
		}
	}

	schedulerJobsCmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage Cloud Scheduler jobs",
	}
	schedulerJobsDescribeCmd := &cobra.Command{
		Use:   "describe JOB_ID",
		Short: "Describe a Cloud Scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE:  run(f, runSchedulerJobsDescribe),
	}
	schedulerJobsPauseCmd := &cobra.Command{
		Use:   "pause JOB_ID",
		Short: "Pause a Cloud Scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE:  run(f, runSchedulerJobsPause),
	}
	schedulerJobsResumeCmd := &cobra.Command{
		Use:   "resume JOB_ID",
		Short: "Resume a Cloud Scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE:  run(f, runSchedulerJobsResume),
	}
	schedulerJobsRunCmd := &cobra.Command{
		Use:   "run JOB_ID",
		Short: "Trigger a Cloud Scheduler job to run immediately",
		Args:  cobra.ExactArgs(1),
		RunE:  run(f, runSchedulerJobsRun),
	}
	schedulerJobsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloud Scheduler jobs",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runSchedulerJobsList(c.Context(), inv, f)
		},
	}
	schedulerJobsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Cloud Scheduler job",
	}
	schedulerJobsCreateHTTPCmd := &cobra.Command{
		Use:   "http JOB_ID",
		Short: "Create an HTTP scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE:  run(cf, runSchedulerJobsCreateHTTP),
	}
	schedulerJobsCreatePubsubCmd := &cobra.Command{
		Use:   "pubsub JOB_ID",
		Short: "Create a Pub/Sub scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE:  run(cf, runSchedulerJobsCreatePubsub),
	}
	schedulerJobsDeleteCmd := &cobra.Command{
		Use:   "delete JOB_ID",
		Short: "Delete a Cloud Scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE:  run(f, runSchedulerJobsDelete),
	}
	schedulerJobsUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Cloud Scheduler job",
	}
	schedulerJobsUpdateHTTPCmd := &cobra.Command{
		Use:   "http JOB_ID",
		Short: "Update an HTTP scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE:  run(uf, runSchedulerJobsUpdateHTTP),
	}
	schedulerJobsUpdatePubsubCmd := &cobra.Command{
		Use:   "pubsub JOB_ID",
		Short: "Update a Pub/Sub scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE:  run(uf, runSchedulerJobsUpdatePubsub),
	}

	for _, c := range []*cobra.Command{
		schedulerJobsDescribeCmd,
		schedulerJobsPauseCmd,
		schedulerJobsResumeCmd,
		schedulerJobsRunCmd,
		schedulerJobsListCmd,
		schedulerJobsDeleteCmd,
	} {
		c.Flags().StringVar(&f.location, "location", "", "Location of the job")
	}

	schedulerJobsListCmd.Flags().BoolVar(&f.listURI, "uri", false, "Print resource names")

	// create common flags
	for _, c := range []*cobra.Command{schedulerJobsCreateHTTPCmd, schedulerJobsCreatePubsubCmd} {
		c.Flags().StringVar(&cf.location, "location", "", "Location")
		c.Flags().StringVar(&cf.schedule, "schedule", "", "Cron schedule expression (required)")
		c.MarkFlagRequired("schedule")
		c.Flags().StringVar(&cf.timeZone, "time-zone", "Etc/UTC", "Time zone")
		c.Flags().StringVar(&cf.description, "description", "", "Description")
		c.Flags().StringVar(&cf.attemptDeadline, "attempt-deadline", "", "Attempt deadline duration")
	}
	schedulerJobsCreateHTTPCmd.Flags().StringVar(&cf.uri, "uri", "", "HTTP target URI (required)")
	schedulerJobsCreateHTTPCmd.MarkFlagRequired("uri")
	schedulerJobsCreateHTTPCmd.Flags().StringVar(&cf.httpMethod, "http-method", "POST", "HTTP method")
	schedulerJobsCreateHTTPCmd.Flags().StringToStringVar(&cf.headers, "headers", nil, "HTTP headers (key=value)")
	schedulerJobsCreateHTTPCmd.Flags().StringVar(&cf.body, "body", "", "HTTP body")
	schedulerJobsCreatePubsubCmd.Flags().StringVar(&cf.topic, "topic", "", "Pub/Sub topic (required)")
	schedulerJobsCreatePubsubCmd.MarkFlagRequired("topic")
	schedulerJobsCreatePubsubCmd.Flags().StringVar(&cf.messageBody, "message-body", "", "Message body")
	schedulerJobsCreatePubsubCmd.Flags().StringToStringVar(&cf.attributes, "attributes", nil, "Message attributes")
	schedulerJobsCreateCmd.AddCommand(schedulerJobsCreateHTTPCmd)
	schedulerJobsCreateCmd.AddCommand(schedulerJobsCreatePubsubCmd)

	// update flags
	for _, c := range []*cobra.Command{schedulerJobsUpdateHTTPCmd, schedulerJobsUpdatePubsubCmd} {
		c.Flags().StringVar(&uf.location, "location", "", "Location")
		c.Flags().StringVar(&uf.schedule, "schedule", "", "Cron schedule expression")
		c.Flags().StringVar(&uf.timeZone, "time-zone", "", "Time zone")
		c.Flags().StringVar(&uf.description, "description", "", "Description")
		c.Flags().StringVar(&uf.attemptDeadline, "attempt-deadline", "", "Attempt deadline duration")
	}
	schedulerJobsUpdateHTTPCmd.Flags().StringVar(&uf.uri, "uri", "", "HTTP target URI")
	schedulerJobsUpdateHTTPCmd.Flags().StringVar(&uf.httpMethod, "http-method", "", "HTTP method")
	schedulerJobsUpdateHTTPCmd.Flags().StringToStringVar(&uf.headers, "headers", nil, "HTTP headers")
	schedulerJobsUpdateHTTPCmd.Flags().StringVar(&uf.body, "body", "", "HTTP body")
	schedulerJobsUpdatePubsubCmd.Flags().StringVar(&uf.topic, "topic", "", "Pub/Sub topic")
	schedulerJobsUpdatePubsubCmd.Flags().StringVar(&uf.messageBody, "message-body", "", "Message body")
	schedulerJobsUpdatePubsubCmd.Flags().StringToStringVar(&uf.attributes, "attributes", nil, "Message attributes")
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
	return schedulerJobsCmd
}

func schedulerJobName(project, location, jobID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project, location, jobID)
}

// resolveSchedulerLocation returns the project and the location from
// --location, falling back like gcloud to the location of the project's App
// Engine app.
func resolveSchedulerLocation(ctx context.Context, inv *invocation, location string) (string, string, error) {
	project, err := inv.project()
	if err != nil {
		return "", "", err
	}
	if location != "" {
		return project, location, nil
	}
	location, err = schedulerAppLocation(ctx, inv.stderr,
		func(ctx context.Context) bool { return appEngineAppExists(ctx, inv, project) },
		func(ctx context.Context) ([]*cloudscheduler.Location, error) {
			svc, err := inv.clients().Scheduler(ctx)
			if err != nil {
				return nil, err
			}
			resp, err := svc.Projects.Locations.List("projects/" + project).PageSize(2).Context(ctx).Do()
			if err != nil {
				return nil, err
			}
			return resp.Locations, nil
		})
	if err != nil {
		return "", "", err
	}
	return project, location, nil
}

// errSchedulerNoLocation is gcloud's error when no location is given and none
// can be inferred from an App Engine app.
var errSchedulerNoLocation = errors.New("Please use the location flag to manually specify a location.")

// schedulerRegionLabel is the Cloud Scheduler location label holding the
// region name.
const schedulerRegionLabel = "cloud.googleapis.com/region"

// schedulerAppLocation mirrors gcloud's AppLocationResolver: when the project
// has an App Engine app, the default location is the first one Cloud
// Scheduler lists (the app's location), with a warning on w when there are
// others to choose from.
func schedulerAppLocation(ctx context.Context, w io.Writer,
	appExists func(context.Context) bool,
	listLocations func(context.Context) ([]*cloudscheduler.Location, error)) (string, error) {
	if !appExists(ctx) {
		return "", errSchedulerNoLocation
	}
	locations, err := listLocations(ctx)
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return "", errSchedulerNoLocation
		}
		return "", fmt.Errorf("listing Cloud Scheduler locations: %w", err)
	}
	if len(locations) == 0 {
		return "", errSchedulerNoLocation
	}
	location := locations[0].Labels[schedulerRegionLabel]
	if location == "" {
		location = locations[0].LocationId
	}
	if len(locations) > 1 {
		fmt.Fprintf(w, "WARNING: We are using the App Engine app location (%s) as the default location. "+
			"Please use the \"--location\" flag if you want to use a different location.\n", location)
	}
	return location, nil
}

// appEngineAppExists reports whether project has an App Engine app. Like
// gcloud, any failure to fetch the app (including missing permission) counts
// as no app.
func appEngineAppExists(ctx context.Context, inv *invocation, project string) bool {
	svc, err := inv.clients().AppEngine(ctx)
	if err != nil {
		return false
	}
	_, err = svc.Apps.Get(project).Context(ctx).Do()
	return err == nil
}

func runSchedulerJobsDescribe(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	job, err := svc.Projects.Locations.Jobs.Get(schedulerJobName(project, location, jobID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing scheduler job: %w", err)
	}

	return emitFormattedTo(inv.stdout, job, "json")
}

func runSchedulerJobsPause(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, jobID)
	_, err = svc.Projects.Locations.Jobs.Pause(name, &cloudscheduler.PauseJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("pausing scheduler job: %w", err)
	}

	fmt.Fprintf(inv.stdout, "Paused job [%s].\n", jobID)
	return nil
}

func runSchedulerJobsResume(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, jobID)
	_, err = svc.Projects.Locations.Jobs.Resume(name, &cloudscheduler.ResumeJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming scheduler job: %w", err)
	}

	fmt.Fprintf(inv.stdout, "Resumed job [%s].\n", jobID)
	return nil
}

func runSchedulerJobsRun(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, jobID)
	_, err = svc.Projects.Locations.Jobs.Run(name, &cloudscheduler.RunJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running scheduler job: %w", err)
	}

	fmt.Fprintf(inv.stdout, "Triggered job [%s].\n", jobID)
	return nil
}

// --- scheduler jobs list (#197) ---

func runSchedulerJobsList(ctx context.Context, inv *invocation, f *schedulerJobFlags) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
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

	if f.listURI {
		for _, job := range allJobs {
			fmt.Fprintln(inv.stdout, job.Name)
		}
		return nil
	}

	if inv.flags.format != "" {
		if allJobs == nil {
			allJobs = []*cloudscheduler.Job{}
		}
		return inv.emit(allJobs)
	}
	return emitSchedulerJobsTable(inv.stdout, inv.stderr, allJobs)
}

// schedulerJobsTableHeaders are the columns of gcloud's default
// `scheduler jobs list` table.
var schedulerJobsTableHeaders = []string{"ID", "LOCATION", "SCHEDULE (TZ)", "TARGET_TYPE", "STATE"}

// schedulerJobsTableRow renders j as a row of gcloud's default table.
func schedulerJobsTableRow(j *cloudscheduler.Job) []string {
	location := ""
	if parts := strings.Split(j.Name, "/"); len(parts) >= 3 {
		location = parts[len(parts)-3]
	}
	targetType := ""
	if j.PubsubTarget != nil {
		targetType += "Pub/Sub"
	}
	if j.AppEngineHttpTarget != nil {
		targetType += "App Engine"
	}
	if j.HttpTarget != nil {
		targetType += "HTTP"
	}
	return []string{path.Base(j.Name), location, fmt.Sprintf("%s (%s)", j.Schedule, j.TimeZone), targetType, j.State}
}

// emitSchedulerJobsTable prints gcloud's default jobs table to w, or
// "Listed 0 items." to errW when there are none, as gcloud does.
func emitSchedulerJobsTable(w, errW io.Writer, jobs []*cloudscheduler.Job) error {
	if len(jobs) == 0 {
		_, err := fmt.Fprintln(errW, "Listed 0 items.")
		return err
	}
	rows := [][]string{schedulerJobsTableHeaders}
	for _, j := range jobs {
		rows = append(rows, schedulerJobsTableRow(j))
	}
	widths := make([]int, len(schedulerJobsTableHeaders))
	for _, row := range rows {
		for i, cell := range row {
			widths[i] = max(widths[i], utf8.RuneCountInString(cell))
		}
	}
	for _, row := range rows {
		if err := writePadded(w, row, widths); err != nil {
			return err
		}
	}
	return nil
}

// --- scheduler jobs create http (#198) ---

func runSchedulerJobsCreateHTTP(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	job := &cloudscheduler.Job{
		Schedule: f.schedule,
		TimeZone: f.timeZone,
		HttpTarget: &cloudscheduler.HttpTarget{
			Uri:        f.uri,
			HttpMethod: strings.ToUpper(f.httpMethod),
		},
	}
	if f.description != "" {
		job.Description = f.description
	}
	if f.attemptDeadline != "" {
		job.AttemptDeadline = f.attemptDeadline
	}
	if len(f.headers) > 0 {
		job.HttpTarget.Headers = f.headers
	}
	if f.body != "" {
		job.HttpTarget.Body = f.body
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	job.Name = fmt.Sprintf("%s/jobs/%s", parent, jobID)
	result, err := svc.Projects.Locations.Jobs.Create(parent, job).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating scheduler job: %w", err)
	}

	fmt.Fprintf(inv.stdout, "Created scheduler job [%s].\n", path.Base(result.Name))
	return nil
}

// --- scheduler jobs create pubsub (#198) ---

func runSchedulerJobsCreatePubsub(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	job := &cloudscheduler.Job{
		Schedule: f.schedule,
		TimeZone: f.timeZone,
		PubsubTarget: &cloudscheduler.PubsubTarget{
			TopicName: f.topic,
		},
	}
	if f.description != "" {
		job.Description = f.description
	}
	if f.attemptDeadline != "" {
		job.AttemptDeadline = f.attemptDeadline
	}
	if f.messageBody != "" {
		job.PubsubTarget.Data = f.messageBody
	}
	if len(f.attributes) > 0 {
		job.PubsubTarget.Attributes = f.attributes
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	job.Name = fmt.Sprintf("%s/jobs/%s", parent, jobID)
	result, err := svc.Projects.Locations.Jobs.Create(parent, job).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating scheduler job: %w", err)
	}

	fmt.Fprintf(inv.stdout, "Created scheduler job [%s].\n", path.Base(result.Name))
	return nil
}

// --- scheduler jobs delete (#199) ---

func runSchedulerJobsDelete(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	if !inv.flags.quiet {
		fmt.Fprintf(inv.stdout, "You are about to delete scheduler job [%s].\n", jobID)
		answer := inv.ask("Do you want to continue (Y/n)? ")
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(inv.stdout, "Aborted.")
			return nil
		}
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, jobID)
	if _, err := svc.Projects.Locations.Jobs.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting scheduler job: %w", err)
	}

	fmt.Fprintf(inv.stdout, "Deleted scheduler job [%s].\n", jobID)
	return nil
}

// --- scheduler jobs update http (#200) ---

func runSchedulerJobsUpdateHTTP(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, jobID)
	job, err := svc.Projects.Locations.Jobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting scheduler job: %w", err)
	}

	var updateMask []string
	if f.schedule != "" {
		job.Schedule = f.schedule
		updateMask = append(updateMask, "schedule")
	}
	if f.timeZone != "" {
		job.TimeZone = f.timeZone
		updateMask = append(updateMask, "time_zone")
	}
	if f.description != "" {
		job.Description = f.description
		updateMask = append(updateMask, "description")
	}
	if f.uri != "" && job.HttpTarget != nil {
		job.HttpTarget.Uri = f.uri
		updateMask = append(updateMask, "http_target.uri")
	}
	if f.httpMethod != "" && job.HttpTarget != nil {
		job.HttpTarget.HttpMethod = strings.ToUpper(f.httpMethod)
		updateMask = append(updateMask, "http_target.http_method")
	}

	if len(updateMask) == 0 {
		return fmt.Errorf("no update flags specified")
	}

	result, err := svc.Projects.Locations.Jobs.Patch(name, job).UpdateMask(strings.Join(updateMask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating scheduler job: %w", err)
	}

	fmt.Fprintf(inv.stdout, "Updated scheduler job [%s].\n", path.Base(result.Name))
	return nil
}

// --- scheduler jobs update pubsub (#200) ---

func runSchedulerJobsUpdatePubsub(ctx context.Context, inv *invocation, f *schedulerJobFlags, jobID string) error {
	project, location, err := resolveSchedulerLocation(ctx, inv, f.location)
	if err != nil {
		return err
	}

	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}

	name := schedulerJobName(project, location, jobID)
	job, err := svc.Projects.Locations.Jobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting scheduler job: %w", err)
	}

	var updateMask []string
	if f.schedule != "" {
		job.Schedule = f.schedule
		updateMask = append(updateMask, "schedule")
	}
	if f.timeZone != "" {
		job.TimeZone = f.timeZone
		updateMask = append(updateMask, "time_zone")
	}
	if f.description != "" {
		job.Description = f.description
		updateMask = append(updateMask, "description")
	}
	if f.topic != "" && job.PubsubTarget != nil {
		job.PubsubTarget.TopicName = f.topic
		updateMask = append(updateMask, "pubsub_target.topic_name")
	}

	if len(updateMask) == 0 {
		return fmt.Errorf("no update flags specified")
	}

	result, err := svc.Projects.Locations.Jobs.Patch(name, job).UpdateMask(strings.Join(updateMask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating scheduler job: %w", err)
	}

	fmt.Fprintf(inv.stdout, "Updated scheduler job [%s].\n", path.Base(result.Name))
	return nil
}
