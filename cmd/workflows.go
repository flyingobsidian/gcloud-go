package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	Use:   "list",
	Short: "List workflow executions",
	Args:  cobra.NoArgs,
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
	workflowsExecutionsCreateCmd.Flags().StringVar(&flagWFData, "data", "", "JSON-encoded arguments for the execution")
	workflowsExecutionsCreateCmd.Flags().StringToStringVar(&flagWFLabels, "labels", nil, "Labels (key=value)")
	workflowsExecutionsCreateCmd.Flags().StringVar(&flagWFLogLevel, "call-log-level", "", "Call log level (LOG_ALL_CALLS, LOG_ERRORS_ONLY, LOG_NONE)")
	workflowsExecutionsCreateCmd.Flags().StringVar(&flagWFHistLevel, "execution-history-level", "", "Execution history level (EXECUTION_HISTORY_BASIC, EXECUTION_HISTORY_DETAILED)")
	workflowsExecutionsCreateCmd.MarkFlagRequired("workflow")

	workflowsExecutionsListCmd.Flags().StringVar(&flagWFFilter, "filter", "", "Server-side filter expression")
	workflowsExecutionsListCmd.Flags().StringVar(&flagWFOrderBy, "order-by", "", "Server-side ordering expression")
	workflowsExecutionsListCmd.Flags().Int64Var(&flagWFPageSize, "page-size", 0, "Number of results per page")
	workflowsExecutionsListCmd.Flags().Int64Var(&flagWFLimit, "limit", 0, "Maximum number of results to return")
	workflowsExecutionsListCmd.MarkFlagRequired("workflow")

	workflowsExecutionsWaitCmd.Flags().IntVar(&flagWFTimeoutSec, "timeout", 0, "Maximum seconds to wait (0 = no timeout)")

	workflowsExecutionsCmd.AddCommand(
		workflowsExecutionsCancelCmd, workflowsExecutionsCreateCmd, workflowsExecutionsDeleteCmd,
		workflowsExecutionsDescribeCmd, workflowsExecutionsListCmd, workflowsExecutionsWaitCmd,
	)
	workflowsCmd.AddCommand(workflowsExecutionsCmd)

	// The workflows resource itself (deploy/execute/etc.) is not covered by this
	// task; keep those subcommands as documented stubs so users get a clear
	// "not yet implemented" message rather than a missing-command error.
	for _, name := range []string{"delete", "deploy", "describe", "execute", "list", "run"} {
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
	return fmt.Sprintf("projects/%s/locations/%s/workflows/%s", project, flagWFLocation, flagWFWorkflow), nil
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
	exec := &workflowexecutions.Execution{
		Argument:              flagWFData,
		Labels:                flagWFLabels,
		CallLogLevel:          flagWFLogLevel,
		ExecutionHistoryLevel: flagWFHistLevel,
	}
	ctx := context.Background()
	svc, err := gcp.WorkflowExecutionsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	created, err := svc.Projects.Locations.Workflows.Executions.Create(parent, exec).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating execution: %w", err)
	}
	fmt.Printf("Started execution [%s].\n", created.Name)
	return emitFormatted(created, "")
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
	return emitFormatted(got, "")
}

func runWFExecutionsList(cmd *cobra.Command, args []string) error {
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
	var all []*workflowexecutions.Execution
	pageToken := ""
	for {
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
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing executions: %w", err)
		}
		all = append(all, resp.Executions...)
		if flagWFLimit > 0 && int64(len(all)) >= flagWFLimit {
			all = all[:flagWFLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
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
	// Workflow executions transition through ACTIVE/QUEUED into SUCCEEDED,
	// FAILED, CANCELLED, or UNAVAILABLE. Anything other than the first two
	// terminates the wait.
	for {
		got, err := svc.Projects.Locations.Workflows.Executions.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling execution: %w", err)
		}
		switch got.State {
		case "SUCCEEDED", "FAILED", "CANCELLED", "UNAVAILABLE":
			return emitFormatted(got, "")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
