package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	dataflow "google.golang.org/api/dataflow/v1b3"
)

func (s *dataflowCmds) initDataflow() {
	s.dataflowJobsListCmd.Flags().StringVar(&s.flagDataflowRegion, "region", "", "Region of the jobs")
	s.dataflowJobsListCmd.Flags().StringVar(&s.flagDataflowListFormat, "format", "", "Output format (e.g. json)")
	s.dataflowJobsListCmd.Flags().StringVar(&s.flagDataflowListFilter, "filter", "", "Client-side filter by job name (substring match)")
	s.dataflowJobsListCmd.Flags().StringVar(&s.flagDataflowListStatus, "status", "", "Server-side status filter: active, all, or terminated")
	s.dataflowJobsListCmd.Flags().BoolVar(&s.flagDataflowListURI, "uri", false, "Print resource names")
	s.dataflowJobsListCmd.Flags().StringVar(&s.flagDataflowCreatedAfter, "created-after", "", "Filter jobs created after this time (RFC 3339)")
	s.dataflowJobsListCmd.Flags().StringVar(&s.flagDataflowCreatedBefore, "created-before", "", "Filter jobs created before this time (RFC 3339)")

	s.dataflowJobsDescribeCmd.Flags().StringVar(&s.flagDataflowRegion, "region", "", "Region of the job")
	s.dataflowJobsDescribeCmd.Flags().BoolVar(&s.flagDataflowDescribeFull, "full", false, "Show full job details including steps")

	s.dataflowJobsCancelCmd.Flags().StringVar(&s.flagDataflowRegion, "region", "", "Region of the job")
	s.dataflowJobsCancelCmd.Flags().BoolVar(&s.flagDataflowCancelDrain, "drain", false, "Drain the job instead of cancelling")

	s.dataflowJobsDrainCmd.Flags().StringVar(&s.flagDataflowRegion, "region", "", "Region of the job")

	s.dataflowJobsRunCmd.Flags().StringVar(&s.flagDataflowRegion, "region", "", "Region for the job")
	s.dataflowJobsRunCmd.Flags().StringVar(&s.flagDFRunGCSLocation, "gcs-location", "", "GCS path to the template (required)")
	s.dataflowJobsRunCmd.MarkFlagRequired("gcs-location")
	s.dataflowJobsRunCmd.Flags().StringToStringVar(&s.flagDFRunParameters, "parameters", nil, "Template parameters (key=value)")
	s.dataflowJobsRunCmd.Flags().StringSliceVar(&s.flagDFRunPipelineOptions, "additional-pipeline-options", nil,
		"Additional pipeline options to pass to the job for launching a Dataflow template. "+
			"Example: --additional-pipeline-options=option1=value1,option2=value2")
	s.dataflowJobsRunCmd.Flags().StringVar(&s.flagDFRunStagingLocation, "staging-location", "", "GCS staging location")
	s.dataflowJobsRunCmd.Flags().Int64Var(&s.flagDFRunMaxWorkers, "max-workers", 0, "Maximum number of workers")
	s.dataflowJobsRunCmd.Flags().Int64Var(&s.flagDFRunNumWorkers, "num-workers", 0, "Initial number of workers")
	s.dataflowJobsRunCmd.Flags().StringVar(&s.flagDFRunNetwork, "network", "", "Network for workers")
	s.dataflowJobsRunCmd.Flags().StringVar(&s.flagDFRunSubnetwork, "subnetwork", "", "Subnetwork for workers")
	s.dataflowJobsRunCmd.Flags().StringVar(&s.flagDFRunMachineType, "worker-machine-type", "", "Worker machine type")
	s.dataflowJobsRunCmd.Flags().StringVar(&s.flagDFRunServiceAccount, "service-account-email", "", "Service account email")
	s.dataflowJobsRunCmd.Flags().StringVar(&s.flagDFRunJobName, "job-name", "", "Job name")
	s.dataflowJobsRunCmd.Flags().BoolVar(&s.flagDFRunTurnkeyAlerts, "enable-turnkey-alerts", false,
		"Enable turnkey alerts for the launched job")

	s.dataflowJobsPauseCmd.Flags().StringVar(&s.flagDataflowRegion, "region", "", "Region of the job")
	s.dataflowJobsResumeCmd.Flags().StringVar(&s.flagDataflowRegion, "region", "", "Region of the job")

	s.dataflowJobsCmd.AddCommand(s.dataflowJobsListCmd)
	s.dataflowJobsCmd.AddCommand(s.dataflowJobsDescribeCmd)
	s.dataflowJobsCmd.AddCommand(s.dataflowJobsCancelCmd)
	s.dataflowJobsCmd.AddCommand(s.dataflowJobsDrainCmd)
	s.dataflowJobsCmd.AddCommand(s.dataflowJobsRunCmd)
	s.dataflowJobsCmd.AddCommand(s.dataflowJobsPauseCmd)
	s.dataflowJobsCmd.AddCommand(s.dataflowJobsResumeCmd)
	s.dataflowCmd.AddCommand(s.dataflowJobsCmd)

	s.registerDataflowExtras(s.dataflowCmd)

}

func (s *dataflowCmds) resolveDataflowRegion() (string, string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", "", err
	}
	region := s.flagDataflowRegion
	if region == "" {
		_, r, err := s.inv.region()
		if err != nil || r == "" {
			return "", "", fmt.Errorf("--region is required")
		}
		region = r
	}
	return project, region, nil
}

func (s *dataflowCmds) runDataflowJobsList(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveDataflowRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}

	// Map --status to the API's filter enum (UNKNOWN, ALL, TERMINATED, ACTIVE).
	apiFilter := statusToAPIFilter(s.flagDataflowListStatus)

	var allJobs []*dataflow.Job
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.List(project, region).Context(ctx)
		if apiFilter != "" {
			call = call.Filter(apiFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing dataflow jobs: %w", err)
		}
		allJobs = append(allJobs, resp.Jobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	// Client-side filters.
	allJobs = s.filterDataflowJobs(allJobs)

	if s.flagDataflowListURI {
		for _, job := range allJobs {
			fmt.Fprintln(s.inv.stdout, job.Id)
		}
		return nil
	}

	if s.flagDataflowListFormat == "json" {
		return s.inv.printJSON(allJobs)
	}

	printTableHeadingTo(s.inv.stdout, len(allJobs), "%-40s %-20s %-20s %-10s %-15s %-25s\n", "JOB_ID", "NAME", "TYPE", "STATE", "REGION", "CREATION_TIME")
	for _, j := range allJobs {
		fmt.Fprintf(s.inv.stdout, "%-40s %-20s %-20s %-10s %-15s %-25s\n", j.Id, j.Name, j.Type, j.CurrentState, j.Location, j.CreateTime)
	}
	return nil
}

func (s *dataflowCmds) runDataflowJobsDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveDataflowRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}

	call := svc.Projects.Locations.Jobs.Get(project, region, args[0]).Context(ctx)
	if s.flagDataflowDescribeFull {
		call = call.View("JOB_VIEW_ALL")
	}
	job, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing dataflow job: %w", err)
	}

	return s.inv.formatOutput(job, "")
}

func (s *dataflowCmds) runDataflowJobsCancel(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveDataflowRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}

	state := "JOB_STATE_CANCELLED"
	action := "Cancelled"
	if s.flagDataflowCancelDrain {
		state = "JOB_STATE_DRAINED"
		action = "Drained"
	}

	for _, jobID := range args {
		_, err = svc.Projects.Locations.Jobs.Update(project, region, jobID, &dataflow.Job{
			RequestedState: state,
		}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("%s dataflow job %s: %w", strings.ToLower(action), jobID, err)
		}
		fmt.Fprintf(s.inv.stdout, "%s job [%s].\n", action, jobID)
	}
	return nil
}

func (s *dataflowCmds) runDataflowJobsDrain(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveDataflowRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}

	for _, jobID := range args {
		_, err = svc.Projects.Locations.Jobs.Update(project, region, jobID, &dataflow.Job{
			RequestedState: "JOB_STATE_DRAINED",
		}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("draining dataflow job %s: %w", jobID, err)
		}
		fmt.Fprintf(s.inv.stdout, "Drained job [%s].\n", jobID)
	}
	return nil
}

// dataflowTemplateRequestBody returns req as a JSON object with
// environment.additionalPipelineOptions set. The Go client's
// RuntimeEnvironment has no such field yet, so the body is built generically.
func dataflowTemplateRequestBody(req *dataflow.CreateJobFromTemplateRequest, options []string) (map[string]any, error) {
	raw, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	env, _ := body["environment"].(map[string]any)
	if env == nil {
		env = map[string]any{}
		body["environment"] = env
	}
	env["additionalPipelineOptions"] = options
	return body, nil
}

// dataflowCreateJobWithPipelineOptions issues templates.create over REST so
// that environment.additionalPipelineOptions reaches the API.
func (s *dataflowCmds) dataflowCreateJobWithPipelineOptions(ctx context.Context, project, region string, req *dataflow.CreateJobFromTemplateRequest, options []string) (*dataflow.Job, error) {
	body, err := dataflowTemplateRequestBody(req, options)
	if err != nil {
		return nil, err
	}
	var job dataflow.Job
	p := fmt.Sprintf("/projects/%s/locations/%s/templates", url.PathEscape(project), url.PathEscape(region))
	if err := s.inv.restClient("https://dataflow.googleapis.com/v1b3").do(ctx, http.MethodPost, p, nil, body, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *dataflowCmds) runDataflowJobsRun(cmd *cobra.Command, args []string) error {
	project, region, err := s.resolveDataflowRegion()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}

	env := &dataflow.RuntimeEnvironment{}
	if s.flagDFRunMaxWorkers > 0 {
		env.MaxWorkers = s.flagDFRunMaxWorkers
	}
	if s.flagDFRunNumWorkers > 0 {
		env.NumWorkers = s.flagDFRunNumWorkers
	}
	if s.flagDFRunNetwork != "" {
		env.Network = s.flagDFRunNetwork
	}
	if s.flagDFRunSubnetwork != "" {
		env.Subnetwork = s.flagDFRunSubnetwork
	}
	if s.flagDFRunMachineType != "" {
		env.MachineType = s.flagDFRunMachineType
	}
	if s.flagDFRunServiceAccount != "" {
		env.ServiceAccountEmail = s.flagDFRunServiceAccount
	}
	if s.flagDFRunStagingLocation != "" {
		env.TempLocation = s.flagDFRunStagingLocation
	}

	if s.flagDFRunTurnkeyAlerts {
		env.AdditionalExperiments = appendExperiment(env.AdditionalExperiments, turnkeyAlertsExperiment)
	}

	req := &dataflow.CreateJobFromTemplateRequest{
		GcsPath:     s.flagDFRunGCSLocation,
		Environment: env,
	}
	if s.flagDFRunJobName != "" {
		req.JobName = s.flagDFRunJobName
	}
	if len(s.flagDFRunParameters) > 0 {
		req.Parameters = s.flagDFRunParameters
	}

	var job *dataflow.Job
	if len(s.flagDFRunPipelineOptions) > 0 {
		job, err = s.dataflowCreateJobWithPipelineOptions(ctx, project, region, req, s.flagDFRunPipelineOptions)
	} else {
		job, err = svc.Projects.Locations.Templates.Create(project, region, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("launching dataflow job: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Launched job [%s] (ID: %s).\n", job.Name, job.Id)
	return nil
}

// turnkeyAlertsExperiment is the pipeline experiment string that flips turnkey
// alerts on server-side. Kept as a constant so the pause/resume + flex-template
// call sites can share it. When google.golang.org/api eventually exposes a
// dedicated `enableTurnkeyAlerts` field, swap both call sites over.
const turnkeyAlertsExperiment = "enable_turnkey_alerts"

// appendExperiment appends exp to experiments only if not already present.
func appendExperiment(experiments []string, exp string) []string {
	for _, e := range experiments {
		if e == exp {
			return experiments
		}
	}
	return append(experiments, exp)
}

func (s *dataflowCmds) runDataflowJobsPause(cmd *cobra.Command, args []string) error {
	return s.updateDataflowRequestedState(args, "JOB_STATE_PAUSED", "Paused")
}

func (s *dataflowCmds) runDataflowJobsResume(cmd *cobra.Command, args []string) error {
	return s.updateDataflowRequestedState(args, "JOB_STATE_RUNNING", "Resumed")
}

func (s *dataflowCmds) updateDataflowRequestedState(jobIDs []string, state, action string) error {
	project, region, err := s.resolveDataflowRegion()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().Dataflow(ctx)
	if err != nil {
		return err
	}
	for _, jobID := range jobIDs {
		_, err = svc.Projects.Locations.Jobs.Update(project, region, jobID, &dataflow.Job{
			RequestedState: state,
		}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("%s dataflow job %s: %w", strings.ToLower(action), jobID, err)
		}
		fmt.Fprintf(s.inv.stdout, "%s job [%s].\n", action, jobID)
	}
	return nil
}

// statusToAPIFilter maps the user-facing --status flag to the Dataflow API
// filter enum value. The API accepts UNKNOWN, ALL, TERMINATED, or ACTIVE.
func statusToAPIFilter(status string) string {
	switch strings.ToLower(status) {
	case "active":
		return "ACTIVE"
	case "terminated":
		return "TERMINATED"
	case "all":
		return "ALL"
	case "":
		return ""
	default:
		return strings.ToUpper(status)
	}
}

// filterDataflowJobs applies client-side filters (--filter for name matching,
// --created-after/--created-before for time range) to the job list.
func (s *dataflowCmds) filterDataflowJobs(jobs []*dataflow.Job) []*dataflow.Job {
	needsFilter := s.flagDataflowListFilter != "" ||
		s.flagDataflowCreatedAfter != "" ||
		s.flagDataflowCreatedBefore != ""
	if !needsFilter {
		return jobs
	}

	nameFilter := strings.ToLower(s.flagDataflowListFilter)

	var filtered []*dataflow.Job
	for _, j := range jobs {
		if nameFilter != "" && !strings.Contains(strings.ToLower(j.Name), nameFilter) {
			continue
		}
		if s.flagDataflowCreatedAfter != "" || s.flagDataflowCreatedBefore != "" {
			ct, err := time.Parse(time.RFC3339, j.CreateTime)
			if err != nil {
				fmt.Fprintf(s.inv.stderr, "WARNING: could not parse createTime %q for job %s: %v\n", j.CreateTime, j.Id, err)
				continue
			}
			if s.flagDataflowCreatedAfter != "" {
				after, err := time.Parse(time.RFC3339, s.flagDataflowCreatedAfter)
				if err == nil && ct.Before(after) {
					continue
				}
			}
			if s.flagDataflowCreatedBefore != "" {
				before, err := time.Parse(time.RFC3339, s.flagDataflowCreatedBefore)
				if err == nil && ct.After(before) {
					continue
				}
			}
		}
		filtered = append(filtered, j)
	}
	return filtered
}

// dataflowCmds holds the commands and flag values of one invocation's "dataflow"
// command group, so invocations don't share state (#2070).
type dataflowCmds struct {
	inv *invocation

	dataflowCmd               *cobra.Command
	dataflowJobsCmd           *cobra.Command
	dataflowJobsListCmd       *cobra.Command
	dataflowJobsDescribeCmd   *cobra.Command
	dataflowJobsCancelCmd     *cobra.Command
	dataflowJobsDrainCmd      *cobra.Command
	dataflowJobsRunCmd        *cobra.Command
	dataflowJobsPauseCmd      *cobra.Command
	dataflowJobsResumeCmd     *cobra.Command
	flagDataflowRegion        string
	flagDataflowListFormat    string
	flagDataflowListFilter    string
	flagDataflowListStatus    string
	flagDataflowListURI       bool
	flagDataflowCreatedAfter  string
	flagDataflowCreatedBefore string
	flagDataflowDescribeFull  bool
	flagDataflowCancelDrain   bool
	flagDFRunGCSLocation      string
	flagDFRunParameters       map[string]string
	flagDFRunStagingLocation  string
	flagDFRunMaxWorkers       int64
	flagDFRunNumWorkers       int64
	flagDFRunNetwork          string
	flagDFRunSubnetwork       string
	flagDFRunMachineType      string
	flagDFRunServiceAccount   string
	flagDFRunJobName          string
	flagDFRunTurnkeyAlerts    bool
	// flagDFRunPipelineOptions is --additional-pipeline-options on `jobs run`
	// (gcloud-python 585.0.0), sent as environment.additionalPipelineOptions.
	flagDFRunPipelineOptions  []string
	flagDFFTTemplateGCS       string
	flagDFFTStagingLocation   string
	flagDFFTTempLocation      string
	flagDFFTServiceAccount    string
	flagDFFTMaxWorkers        int64
	flagDFFTNumWorkers        int64
	flagDFFTWorkerMachine     string
	flagDFFTLauncherMachine   string
	flagDFFTLauncherTimeout   int64
	flagDFFTSubnetwork        string
	flagDFFTNetwork           string
	flagDFFTKmsKey            string
	flagDFFTWorkerRegion      string
	flagDFFTWorkerZone        string
	flagDFFTEnableStreaming   bool
	flagDFFTDisablePublicIPs  bool
	flagDFFTExperiments       []string
	flagDFFTPipelineOptions   []string
	flagDFFTUserLabels        map[string]string
	flagDFFTParameters        map[string]string
	flagDFFTTransformName     map[string]string
	flagDFFTUpdate            bool
	flagDFFTFlexRSGoal        string
	flagDFFTTurnkeyAlerts     bool
	flagDFFTBuildImage        string
	flagDFFTBuildImageGCRPath string
	flagDFFTBuildSDK          string
	flagDFFTBuildMetadataFile string
	flagDFFTBuildEnv          map[string]string
	flagDFFTBuildFlexOptions  map[string]string
	// snapshots flags
	flagDFSnapRegion      string
	flagDFSnapJobID       string
	flagDFSnapTTL         string
	flagDFSnapDescription string
	flagDFSnapSources     bool
	// yaml run flags
	flagDFYamlPipelineFile string
	flagDFYamlPipeline     string
	flagDFYamlOptions      map[string]string
	flagDFYamlJinjaVars    map[string]string
	flagDFYamlRegion       string
	flagDFYamlGCSLocation  string
}

// newDataflowCmd returns the "dataflow" command group for inv.
func newDataflowCmd(inv *invocation) *cobra.Command {
	return newDataflowCmds(inv).dataflowCmd
}

// newDataflowCmds builds the "dataflow" command group for inv and returns its state.
func newDataflowCmds(inv *invocation) *dataflowCmds {
	s := &dataflowCmds{inv: inv}
	s.dataflowCmd = &cobra.Command{
		Use:   "dataflow",
		Short: "Manage Google Cloud Dataflow",
	}
	s.dataflowJobsCmd = &cobra.Command{
		Use:   "jobs",
		Short: "Manage Dataflow jobs",
	}
	s.dataflowJobsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Dataflow jobs",
		Args:  cobra.NoArgs,
		RunE:  s.runDataflowJobsList,
	}
	s.dataflowJobsDescribeCmd = &cobra.Command{
		Use:   "describe JOB_ID",
		Short: "Describe a Dataflow job",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDataflowJobsDescribe,
	}
	s.dataflowJobsCancelCmd = &cobra.Command{
		Use:   "cancel JOB_ID [JOB_ID...]",
		Short: "Cancel running Dataflow jobs",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runDataflowJobsCancel,
	}
	s.dataflowJobsDrainCmd = &cobra.Command{
		Use:   "drain JOB_ID [JOB_ID...]",
		Short: "Drain running Dataflow jobs",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runDataflowJobsDrain,
	}
	s.dataflowJobsRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run a Dataflow job from a template",
		Args:  cobra.NoArgs,
		RunE:  s.runDataflowJobsRun,
	}
	s.dataflowJobsPauseCmd = &cobra.Command{
		Use:   "pause JOB_ID [JOB_ID...]",
		Short: "Pause running Dataflow jobs",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runDataflowJobsPause,
	}
	s.dataflowJobsResumeCmd = &cobra.Command{
		Use:   "resume JOB_ID [JOB_ID...]",
		Short: "Resume paused Dataflow jobs",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runDataflowJobsResume,
	}
	s.initDataflow()
	return s
}
