package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	workflowexecutions "google.golang.org/api/workflowexecutions/v1"
)

// --- gcloud workflows (#950) ---

var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Manage Cloud Workflows",
}

var (
	flagWFLocation   string
	flagWFWorkflow   string
	flagWFData       string
	flagWFLabels     map[string]string
	flagWFLogLevel   string
	flagWFHistLevel  string
	flagWFPageSize   int64
	flagWFLimit      int64
	flagWFFilter     string
	flagWFOrderBy    string
	flagWFTimeoutSec int
)

// wfDefaultLocation is the location gcloud falls back to when neither
// --location nor the workflows/location property is set.
const wfDefaultLocation = "us-central1"

// wfRunMaxWait matches gcloud's 24-hour limit on waiting for `workflows run`.
const wfRunMaxWait = 24 * time.Hour

var workflowsRunCmd = &cobra.Command{
	Use:   "run WORKFLOW",
	Short: "Execute a workflow and wait for the execution to complete",
	Args:  cobra.ExactArgs(1),
	RunE:  runWFRun,
}

// --- Executions ---

var workflowsExecutionsCmd = &cobra.Command{
	Use:   "executions",
	Short: "Manage Cloud Workflows executions",
}

var workflowsExecutionsCancelCmd = &cobra.Command{
	Use:   "cancel EXECUTION",
	Short: "Cancel a workflow execution",
	Args:  cobra.ExactArgs(1),
	RunE:  runWFExecutionsCancel,
}

var workflowsExecutionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create (start) a new workflow execution",
	Args:  cobra.NoArgs,
	RunE:  runWFExecutionsCreate,
}

var workflowsExecutionsDeleteCmd = &cobra.Command{
	Use:   "delete EXECUTION",
	Short: "Delete the recorded history of a workflow execution",
	Args:  cobra.ExactArgs(1),
	RunE:  runWFExecutionsDelete,
}

var workflowsExecutionsDescribeCmd = &cobra.Command{
	Use:   "describe EXECUTION",
	Short: "Describe a workflow execution",
	Args:  cobra.ExactArgs(1),
	RunE:  runWFExecutionsDescribe,
}

var workflowsExecutionsListCmd = &cobra.Command{
	Use:   "list [WORKFLOW]",
	Short: "List workflow executions",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWFExecutionsList,
}

var workflowsExecutionsWaitCmd = &cobra.Command{
	Use:   "wait EXECUTION",
	Short: "Poll a workflow execution until it completes",
	Args:  cobra.ExactArgs(1),
	RunE:  runWFExecutionsWait,
}

func init() {
	// All executions subcommands scope to a location and a workflow.
	for _, c := range []*cobra.Command{
		workflowsExecutionsCancelCmd, workflowsExecutionsCreateCmd, workflowsExecutionsDeleteCmd,
		workflowsExecutionsDescribeCmd, workflowsExecutionsListCmd, workflowsExecutionsWaitCmd,
	} {
		c.Flags().StringVar(&flagWFLocation, "location", "", "Location containing the workflow")
		c.Flags().StringVar(&flagWFWorkflow, "workflow", "", "Workflow name (required unless EXECUTION is a fully-qualified resource)")
	}
	workflowsRunCmd.Flags().StringVar(&flagWFLocation, "location", "",
		"Location containing the workflow; alternatively set CLOUDSDK_WORKFLOWS_LOCATION (default "+wfDefaultLocation+")")
	for _, c := range []*cobra.Command{workflowsExecutionsCreateCmd, workflowsRunCmd} {
		c.Flags().StringVar(&flagWFData, "data", "", "JSON-encoded arguments for the execution")
		c.Flags().StringToStringVar(&flagWFLabels, "labels", nil, "Labels (key=value)")
		c.Flags().StringVar(&flagWFLogLevel, "call-log-level", "", "Call log level (LOG_ALL_CALLS, LOG_ERRORS_ONLY, LOG_NONE)")
		c.Flags().StringVar(&flagWFHistLevel, "execution-history-level", "", "Execution history level (EXECUTION_HISTORY_BASIC, EXECUTION_HISTORY_DETAILED)")
	}
	workflowsExecutionsCreateCmd.MarkFlagRequired("workflow")

	workflowsExecutionsListCmd.Flags().StringVar(&flagWFFilter, "filter", "", "Server-side filter expression")
	workflowsExecutionsListCmd.Flags().StringVar(&flagWFOrderBy, "order-by", "", "Server-side ordering expression")
	workflowsExecutionsListCmd.Flags().Int64Var(&flagWFPageSize, "page-size", 0, "Number of results per page")
	workflowsExecutionsListCmd.Flags().Int64Var(&flagWFLimit, "limit", 0, "Maximum number of results to return")

	workflowsExecutionsWaitCmd.Flags().IntVar(&flagWFTimeoutSec, "timeout", 0, "Maximum seconds to wait (0 = no timeout)")

	workflowsExecutionsCmd.AddCommand(
		workflowsExecutionsCancelCmd, workflowsExecutionsCreateCmd, workflowsExecutionsDeleteCmd,
		workflowsExecutionsDescribeCmd, workflowsExecutionsListCmd, workflowsExecutionsWaitCmd,
	)
	workflowsCmd.AddCommand(workflowsExecutionsCmd, workflowsRunCmd)

	// The workflows resource itself (deploy/execute/etc.) is not covered by this
	// task; keep those subcommands as documented stubs so users get a clear
	// "not yet implemented" message rather than a missing-command error.
	for _, name := range []string{"delete", "deploy", "describe", "execute", "list"} {
		registerStubCommand(workflowsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(workflowsCmd)
}

// --- Helpers ---

func wfWorkflowParent(project string) (string, error) {
	if flagWFLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	if flagWFWorkflow == "" {
		return "", fmt.Errorf("--workflow is required")
	}
	return wfWorkflowPath(project, flagWFLocation, flagWFWorkflow), nil
}

// wfWorkflowPath qualifies workflow into a full resource path, passing
// fully-qualified names through unchanged.
func wfWorkflowPath(project, location, workflow string) string {
	if strings.HasPrefix(workflow, "projects/") {
		return workflow
	}
	return fmt.Sprintf("projects/%s/locations/%s/workflows/%s", project, location, workflow)
}

// wfResolveLocation returns the location from --location or
// CLOUDSDK_WORKFLOWS_LOCATION, falling back to gcloud's default location
// with the same warning gcloud prints.
func wfResolveLocation(w io.Writer) string {
	if loc := config.Resolve(flagWFLocation, "CLOUDSDK_WORKFLOWS_LOCATION", ""); loc != "" {
		return loc
	}
	fmt.Fprintf(w, "WARNING: The default location(%s) was used since the location flag was not specified.\n",
		wfDefaultLocation)
	return wfDefaultLocation
}

// wfNewExecution builds an execution from the --data, --labels,
// --call-log-level and --execution-history-level flags.
func wfNewExecution() *workflowexecutions.Execution {
	return &workflowexecutions.Execution{
		Argument:              flagWFData,
		Labels:                flagWFLabels,
		CallLogLevel:          flagWFLogLevel,
		ExecutionHistoryLevel: flagWFHistLevel,
	}
}

// wfExecutionGetter fetches the current state of the execution being waited on.
type wfExecutionGetter func(ctx context.Context) (*workflowexecutions.Execution, error)

// wfWaitForExecution polls get until the execution leaves the ACTIVE and
// QUEUED states. The delay between polls starts at interval and grows by a
// factor of 1.25 up to maxInterval, as gcloud's waiter does.
func wfWaitForExecution(ctx context.Context, get wfExecutionGetter,
	interval, maxInterval time.Duration) (*workflowexecutions.Execution, error) {
	for {
		got, err := get(ctx)
		if err != nil {
			return nil, fmt.Errorf("polling execution: %w", err)
		}
		if got.State != "ACTIVE" && got.State != "QUEUED" {
			return got, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		interval = min(interval*5/4, maxInterval)
	}
}

// wfExecutionName qualifies EXECUTION into a full resource path. It accepts
// either a fully-qualified name (in which case --workflow/--location are
// optional) or a bare id (in which case they are required).
func wfExecutionName(id, project string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := wfWorkflowParent(project)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/executions/%s", parent, id), nil
}

// wfListParent resolves the workflow whose executions are listed, from the
// WORKFLOW argument or --workflow. A bare workflow id is qualified with the
// location from wfResolveLocation, which warns on w when it falls back to the
// default location.
func wfListParent(w io.Writer, args []string, project string) (string, error) {
	workflow := flagWFWorkflow
	if len(args) > 0 {
		workflow = args[0]
	}
	if workflow == "" {
		return "", fmt.Errorf("WORKFLOW argument or --workflow is required")
	}
	if strings.HasPrefix(workflow, "projects/") {
		return workflow, nil
	}
	return wfWorkflowPath(project, wfResolveLocation(w), workflow), nil
}

// wfExecutionsPager fetches one page of executions for the given page token.
type wfExecutionsPager func(ctx context.Context, pageToken string) (*workflowexecutions.ListExecutionsResponse, error)

// wfListExecutions pages through executions until the results run out or
// limit executions have been collected (limit <= 0 means no limit). It never
// returns a nil slice, so an empty result renders as [] in JSON.
func wfListExecutions(ctx context.Context, page wfExecutionsPager, limit int64) ([]*workflowexecutions.Execution, error) {
	all := []*workflowexecutions.Execution{}
	pageToken := ""
	for {
		resp, err := page(ctx, pageToken)
		if err != nil {
			return nil, fmt.Errorf("listing executions: %w", err)
		}
		all = append(all, resp.Executions...)
		if limit > 0 && int64(len(all)) >= limit {
			return all[:limit], nil
		}
		if resp.NextPageToken == "" {
			return all, nil
		}
		pageToken = resp.NextPageToken
	}
}

// --- Executions impl ---

func runWFExecutionsCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name, err := wfExecutionName(args[0], project)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkflowExecutionsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Workflows.Executions.Cancel(name, &workflowexecutions.CancelExecutionRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling execution: %w", err)
	}
	fmt.Printf("Cancelled execution [%s].\n", args[0])
	return nil
}

func runWFExecutionsCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent, err := wfWorkflowParent(project)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkflowExecutionsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	created, err := svc.Projects.Locations.Workflows.Executions.Create(parent, wfNewExecution()).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating execution: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Started execution [%s].\n", created.Name)
	return emitFormatted(created, flagFormat)
}

// runWFExecutionsDelete deletes the *history* of an execution: the workflow
// executions API has no full delete, only DeleteExecutionHistory (which is
// what `gcloud workflows executions delete` maps onto).
func runWFExecutionsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name, err := wfExecutionName(args[0], project)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkflowExecutionsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Workflows.Executions.DeleteExecutionHistory(name, &workflowexecutions.DeleteExecutionHistoryRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting execution history: %w", err)
	}
	fmt.Printf("Deleted history of execution [%s].\n", args[0])
	return nil
}

func runWFExecutionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name, err := wfExecutionName(args[0], project)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkflowExecutionsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Workflows.Executions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing execution: %w", err)
	}
	return emitFormatted(got, flagFormat)
}

func runWFExecutionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent, err := wfListParent(os.Stderr, args, project)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkflowExecutionsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	page := func(ctx context.Context, pageToken string) (*workflowexecutions.ListExecutionsResponse, error) {
		call := svc.Projects.Locations.Workflows.Executions.List(parent).Context(ctx)
		if flagWFPageSize > 0 {
			call = call.PageSize(flagWFPageSize)
		}
		if flagWFFilter != "" {
			call = call.Filter(flagWFFilter)
		}
		if flagWFOrderBy != "" {
			call = call.OrderBy(flagWFOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		return call.Do()
	}
	all, err := wfListExecutions(ctx, page, flagWFLimit)
	if err != nil {
		return err
	}
	return emitFormatted(all, wfListFormat(flagFormat))
}

// wfExecutionsListFormat is gcloud's default output format for
// `workflows executions list`.
const wfExecutionsListFormat = "table(name,state,startTime,endTime)"

// wfListFormat returns format, or gcloud's default table when it is unset.
func wfListFormat(format string) string {
	if format == "" {
		return wfExecutionsListFormat
	}
	return format
}

func runWFExecutionsWait(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name, err := wfExecutionName(args[0], project)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if flagWFTimeoutSec > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(flagWFTimeoutSec)*time.Second)
		defer cancel()
	}
	svc, err := gcp.WorkflowExecutionsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	get := func(ctx context.Context) (*workflowexecutions.Execution, error) {
		return svc.Projects.Locations.Workflows.Executions.Get(name).Context(ctx).Do()
	}
	got, err := wfWaitForExecution(ctx, get, 2*time.Second, 2*time.Second)
	if err != nil {
		return err
	}
	return emitFormatted(got, flagFormat)
}

// runWFRun executes a workflow and waits for the execution to finish,
// printing the final execution like `gcloud workflows run`.
func runWFRun(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := wfWorkflowPath(project, wfResolveLocation(os.Stderr), args[0])
	ctx, cancel := context.WithTimeout(context.Background(), wfRunMaxWait)
	defer cancel()
	svc, err := gcp.WorkflowExecutionsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	created, err := svc.Projects.Locations.Workflows.Executions.Create(parent, wfNewExecution()).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating execution: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Waiting for execution [%s] to complete...", path.Base(created.Name))
	get := func(ctx context.Context) (*workflowexecutions.Execution, error) {
		return svc.Projects.Locations.Workflows.Executions.Get(created.Name).Context(ctx).Do()
	}
	got, err := wfWaitForExecution(ctx, get, time.Second, time.Minute)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed.")
		return fmt.Errorf("waiting for execution %s: %w", created.Name, err)
	}
	fmt.Fprintln(os.Stderr, "done.")
	return emitFormatted(got, flagFormat)
}
