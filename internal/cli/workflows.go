package cli

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
	workflowexecutions "google.golang.org/api/workflowexecutions/v1"
	workflows "google.golang.org/api/workflows/v1"
)

// --- gcloud workflows (#950) ---

// workflowsFlags holds the flag values of the workflows commands.
type workflowsFlags struct {
	location   string
	workflow   string
	data       string
	labels     map[string]string
	logLevel   string
	histLevel  string
	pageSize   int64
	limit      int64
	filter     string
	orderBy    string
	timeoutSec int
}

// wfDefaultLocation is the location gcloud falls back to when neither
// --location nor the workflows/location property is set.
const wfDefaultLocation = "us-central1"

// wfRunMaxWait matches gcloud's 24-hour limit on waiting for `workflows run`.
const wfRunMaxWait = 24 * time.Hour

func newWorkflowsCmd(inv *invocation) *cobra.Command {
	f := &workflowsFlags{}
	run := func(fn func(context.Context, *invocation, *workflowsFlags, []string) error) func(*cobra.Command, []string) error {
		return func(c *cobra.Command, args []string) error {
			return fn(c.Context(), inv, f, args)
		}
	}

	workflowsCmd := &cobra.Command{
		Use:   "workflows",
		Short: "Manage Cloud Workflows",
	}
	workflowsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List workflows",
		Args:  cobra.NoArgs,
		RunE:  run(runWFList),
	}
	workflowsRunCmd := &cobra.Command{
		Use:   "run WORKFLOW",
		Short: "Execute a workflow and wait for the execution to complete",
		Args:  cobra.ExactArgs(1),
		RunE:  run(runWFRun),
	}

	// --- Executions ---

	workflowsExecutionsCmd := &cobra.Command{
		Use:   "executions",
		Short: "Manage Cloud Workflows executions",
	}
	workflowsExecutionsCancelCmd := &cobra.Command{
		Use:   "cancel EXECUTION",
		Short: "Cancel a workflow execution",
		Args:  cobra.ExactArgs(1),
		RunE:  run(runWFExecutionsCancel),
	}
	workflowsExecutionsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create (start) a new workflow execution",
		Args:  cobra.NoArgs,
		RunE:  run(runWFExecutionsCreate),
	}
	workflowsExecutionsDeleteCmd := &cobra.Command{
		Use:   "delete EXECUTION",
		Short: "Delete the recorded history of a workflow execution",
		Args:  cobra.ExactArgs(1),
		RunE:  run(runWFExecutionsDelete),
	}
	workflowsExecutionsDescribeCmd := &cobra.Command{
		Use:   "describe EXECUTION",
		Short: "Describe a workflow execution",
		Args:  cobra.ExactArgs(1),
		RunE:  run(runWFExecutionsDescribe),
	}
	workflowsExecutionsListCmd := &cobra.Command{
		Use:   "list [WORKFLOW]",
		Short: "List workflow executions",
		Args:  cobra.MaximumNArgs(1),
		RunE:  run(runWFExecutionsList),
	}
	workflowsExecutionsWaitCmd := &cobra.Command{
		Use:   "wait EXECUTION",
		Short: "Poll a workflow execution until it completes",
		Args:  cobra.ExactArgs(1),
		RunE:  run(runWFExecutionsWait),
	}

	// All executions subcommands scope to a location and a workflow.
	for _, c := range []*cobra.Command{
		workflowsExecutionsCancelCmd, workflowsExecutionsCreateCmd, workflowsExecutionsDeleteCmd,
		workflowsExecutionsDescribeCmd, workflowsExecutionsListCmd, workflowsExecutionsWaitCmd,
	} {
		c.Flags().StringVar(&f.location, "location", "", "Location containing the workflow")
		c.Flags().StringVar(&f.workflow, "workflow", "", "Workflow name (required unless EXECUTION is a fully-qualified resource)")
	}
	workflowsRunCmd.Flags().StringVar(&f.location, "location", "",
		"Location containing the workflow; alternatively set the workflows/location property (default "+wfDefaultLocation+")")
	workflowsListCmd.Flags().StringVar(&f.location, "location", "",
		"Location to list workflows from; alternatively set the workflows/location property (default "+wfDefaultLocation+")")
	for _, c := range []*cobra.Command{workflowsExecutionsCreateCmd, workflowsRunCmd} {
		c.Flags().StringVar(&f.data, "data", "", "JSON-encoded arguments for the execution")
		c.Flags().StringToStringVar(&f.labels, "labels", nil, "Labels (key=value)")
		c.Flags().StringVar(&f.logLevel, "call-log-level", "", "Call log level (LOG_ALL_CALLS, LOG_ERRORS_ONLY, LOG_NONE)")
		c.Flags().StringVar(&f.histLevel, "execution-history-level", "", "Execution history level (EXECUTION_HISTORY_BASIC, EXECUTION_HISTORY_DETAILED)")
	}
	workflowsExecutionsCreateCmd.MarkFlagRequired("workflow")

	workflowsExecutionsListCmd.Flags().StringVar(&f.filter, "filter", "", "Server-side filter expression")
	workflowsExecutionsListCmd.Flags().StringVar(&f.orderBy, "order-by", "", "Server-side ordering expression")
	for _, c := range []*cobra.Command{workflowsExecutionsListCmd, workflowsListCmd} {
		c.Flags().Int64Var(&f.pageSize, "page-size", 0, "Number of results per page")
		c.Flags().Int64Var(&f.limit, "limit", 0, "Maximum number of results to return")
	}

	workflowsExecutionsWaitCmd.Flags().IntVar(&f.timeoutSec, "timeout", 0, "Maximum seconds to wait (0 = no timeout)")

	workflowsExecutionsCmd.AddCommand(
		workflowsExecutionsCancelCmd, workflowsExecutionsCreateCmd, workflowsExecutionsDeleteCmd,
		workflowsExecutionsDescribeCmd, workflowsExecutionsListCmd, workflowsExecutionsWaitCmd,
	)
	workflowsCmd.AddCommand(workflowsExecutionsCmd, workflowsListCmd, workflowsRunCmd)

	// The workflows resource itself (deploy/execute/etc.) is not covered by this
	// task; keep those subcommands as documented stubs so users get a clear
	// "not yet implemented" message rather than a missing-command error.
	for _, name := range []string{"delete", "deploy", "describe", "execute"} {
		registerStubCommand(workflowsCmd, name, "Not yet implemented")
	}
	return workflowsCmd
}

// --- Helpers ---

func wfWorkflowParent(f *workflowsFlags, project string) (string, error) {
	if f.location == "" {
		return "", fmt.Errorf("--location is required")
	}
	if f.workflow == "" {
		return "", fmt.Errorf("--workflow is required")
	}
	return wfWorkflowPath(project, f.location, f.workflow), nil
}

// wfWorkflowPath qualifies workflow into a full resource path, passing
// fully-qualified names through unchanged.
func wfWorkflowPath(project, location, workflow string) string {
	if strings.HasPrefix(workflow, "projects/") {
		return workflow
	}
	return fmt.Sprintf("projects/%s/locations/%s/workflows/%s", project, location, workflow)
}

// wfResolveLocation returns the location from --location or the
// workflows/location property (CLOUDSDK_WORKFLOWS_LOCATION or the active
// configuration), falling back to gcloud's default location with the same
// warning gcloud prints.
func wfResolveLocation(w io.Writer, env config.Env, location string) (string, error) {
	if location != "" {
		return location, nil
	}
	loc, err := env.Property("workflows", "location")
	if err != nil {
		return "", fmt.Errorf("reading workflows/location: %w", err)
	}
	if loc != "" {
		return loc, nil
	}
	fmt.Fprintf(w, "WARNING: The default location(%s) was used since the location flag was not specified.\n",
		wfDefaultLocation)
	return wfDefaultLocation, nil
}

// wfNewExecution builds an execution from the --data, --labels,
// --call-log-level and --execution-history-level flags.
func wfNewExecution(f *workflowsFlags) *workflowexecutions.Execution {
	return &workflowexecutions.Execution{
		Argument:              f.data,
		Labels:                f.labels,
		CallLogLevel:          f.logLevel,
		ExecutionHistoryLevel: f.histLevel,
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
func wfExecutionName(f *workflowsFlags, id, project string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := wfWorkflowParent(f, project)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/executions/%s", parent, id), nil
}

// wfListParent resolves the workflow whose executions are listed, from the
// WORKFLOW argument or --workflow. A bare workflow id is qualified with the
// location from wfResolveLocation, which warns on w when it falls back to the
// default location.
func wfListParent(w io.Writer, env config.Env, f *workflowsFlags, args []string, project string) (string, error) {
	workflow := f.workflow
	if len(args) > 0 {
		workflow = args[0]
	}
	if workflow == "" {
		return "", fmt.Errorf("WORKFLOW argument or --workflow is required")
	}
	if strings.HasPrefix(workflow, "projects/") {
		return workflow, nil
	}
	location, err := wfResolveLocation(w, env, f.location)
	if err != nil {
		return "", err
	}
	return wfWorkflowPath(project, location, workflow), nil
}

// wfWorkflowsPager fetches one page of workflows for the given page token.
type wfWorkflowsPager func(ctx context.Context, pageToken string) (*workflows.ListWorkflowsResponse, error)

// wfListWorkflows pages through workflows until the results run out or limit
// workflows have been collected (limit <= 0 means no limit). It never returns
// a nil slice, so an empty result renders as [] in JSON.
func wfListWorkflows(ctx context.Context, page wfWorkflowsPager, limit int64) ([]*workflows.Workflow, error) {
	all := []*workflows.Workflow{}
	pageToken := ""
	for {
		resp, err := page(ctx, pageToken)
		if err != nil {
			return nil, fmt.Errorf("listing workflows: %w", err)
		}
		all = append(all, resp.Workflows...)
		if limit > 0 && int64(len(all)) >= limit {
			return all[:limit], nil
		}
		if resp.NextPageToken == "" {
			return all, nil
		}
		pageToken = resp.NextPageToken
	}
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

// --- Workflows impl ---

// wfWorkflowsListFormat is gcloud's default output format for `workflows list`.
const wfWorkflowsListFormat = "table(name,state,revisionId,updateTime)"

func runWFList(ctx context.Context, inv *invocation, f *workflowsFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	location, err := wfResolveLocation(inv.stderr, inv.env, f.location)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	svc, err := inv.clients().Workflows(ctx)
	if err != nil {
		return err
	}
	page := func(ctx context.Context, pageToken string) (*workflows.ListWorkflowsResponse, error) {
		call := svc.Projects.Locations.Workflows.List(parent).Context(ctx)
		if f.pageSize > 0 {
			call = call.PageSize(f.pageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		return call.Do()
	}
	all, err := wfListWorkflows(ctx, page, f.limit)
	if err != nil {
		return err
	}
	format := inv.flags.format
	if format == "" {
		format = wfWorkflowsListFormat
	}
	return emitFormattedTo(inv.stdout, all, format)
}

// --- Executions impl ---

func runWFExecutionsCancel(ctx context.Context, inv *invocation, f *workflowsFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	name, err := wfExecutionName(f, args[0], project)
	if err != nil {
		return err
	}
	svc, err := inv.clients().WorkflowExecutions(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Workflows.Executions.Cancel(name, &workflowexecutions.CancelExecutionRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling execution: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Cancelled execution [%s].\n", args[0])
	return nil
}

func runWFExecutionsCreate(ctx context.Context, inv *invocation, f *workflowsFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	parent, err := wfWorkflowParent(f, project)
	if err != nil {
		return err
	}
	svc, err := inv.clients().WorkflowExecutions(ctx)
	if err != nil {
		return err
	}
	created, err := svc.Projects.Locations.Workflows.Executions.Create(parent, wfNewExecution(f)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating execution: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Started execution [%s].\n", created.Name)
	return inv.emit(created)
}

// runWFExecutionsDelete deletes the *history* of an execution: the workflow
// executions API has no full delete, only DeleteExecutionHistory (which is
// what `gcloud workflows executions delete` maps onto).
func runWFExecutionsDelete(ctx context.Context, inv *invocation, f *workflowsFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	name, err := wfExecutionName(f, args[0], project)
	if err != nil {
		return err
	}
	svc, err := inv.clients().WorkflowExecutions(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Workflows.Executions.DeleteExecutionHistory(name, &workflowexecutions.DeleteExecutionHistoryRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting execution history: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Deleted history of execution [%s].\n", args[0])
	return nil
}

func runWFExecutionsDescribe(ctx context.Context, inv *invocation, f *workflowsFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	name, err := wfExecutionName(f, args[0], project)
	if err != nil {
		return err
	}
	svc, err := inv.clients().WorkflowExecutions(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Workflows.Executions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing execution: %w", err)
	}
	return inv.emit(got)
}

func runWFExecutionsList(ctx context.Context, inv *invocation, f *workflowsFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	parent, err := wfListParent(inv.stderr, inv.env, f, args, project)
	if err != nil {
		return err
	}
	svc, err := inv.clients().WorkflowExecutions(ctx)
	if err != nil {
		return err
	}
	page := func(ctx context.Context, pageToken string) (*workflowexecutions.ListExecutionsResponse, error) {
		call := svc.Projects.Locations.Workflows.Executions.List(parent).Context(ctx)
		if f.pageSize > 0 {
			call = call.PageSize(f.pageSize)
		}
		if f.filter != "" {
			call = call.Filter(f.filter)
		}
		if f.orderBy != "" {
			call = call.OrderBy(f.orderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		return call.Do()
	}
	all, err := wfListExecutions(ctx, page, f.limit)
	if err != nil {
		return err
	}
	return emitFormattedTo(inv.stdout, all, wfListFormat(inv.flags.format))
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

func runWFExecutionsWait(ctx context.Context, inv *invocation, f *workflowsFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	name, err := wfExecutionName(f, args[0], project)
	if err != nil {
		return err
	}
	if f.timeoutSec > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(f.timeoutSec)*time.Second)
		defer cancel()
	}
	svc, err := inv.clients().WorkflowExecutions(ctx)
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
	return inv.emit(got)
}

// runWFRun executes a workflow and waits for the execution to finish,
// printing the final execution like `gcloud workflows run`.
func runWFRun(ctx context.Context, inv *invocation, f *workflowsFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	location, err := wfResolveLocation(inv.stderr, inv.env, f.location)
	if err != nil {
		return err
	}
	parent := wfWorkflowPath(project, location, args[0])
	ctx, cancel := context.WithTimeout(ctx, wfRunMaxWait)
	defer cancel()
	svc, err := inv.clients().WorkflowExecutions(ctx)
	if err != nil {
		return err
	}
	created, err := svc.Projects.Locations.Workflows.Executions.Create(parent, wfNewExecution(f)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating execution: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Waiting for execution [%s] to complete...", path.Base(created.Name))
	get := func(ctx context.Context) (*workflowexecutions.Execution, error) {
		return svc.Projects.Locations.Workflows.Executions.Get(created.Name).Context(ctx).Do()
	}
	got, err := wfWaitForExecution(ctx, get, time.Second, time.Minute)
	if err != nil {
		fmt.Fprintln(inv.stderr, "failed.")
		return fmt.Errorf("waiting for execution %s: %w", created.Name, err)
	}
	fmt.Fprintln(inv.stderr, "done.")
	return inv.emit(got)
}
