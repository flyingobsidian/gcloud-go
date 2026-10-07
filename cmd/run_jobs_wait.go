package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run jobs execute --wait (#1887) ---
//
// Polls the execution created by Jobs.Run, printing each stage as its
// condition succeeds, in the style of gcloud's staged progress tracker.

const (
	runCondSucceeded = "CONDITION_SUCCEEDED"
	runCondFailed    = "CONDITION_FAILED"

	runExecCondCompleted = "Completed"

	runJobsWaitInterval = 2 * time.Second
)

var errRunExecutionFailed = errors.New("the execution failed")

// runExecStage is one line of the progress output, completed when the
// execution condition of the same type succeeds.
type runExecStage struct {
	label     string
	condition string
}

var runExecStages = []runExecStage{
	{label: "Provisioning resources...", condition: "ResourcesAvailable"},
	{label: "Starting execution...", condition: "Started"},
	{label: "Running execution...", condition: runExecCondCompleted},
}

// runExecutionGetter fetches the current state of the execution being waited on.
type runExecutionGetter func(ctx context.Context) (*runv2.GoogleCloudRunV2Execution, error)

// runJobsOpExecution returns the execution recorded in the metadata of the
// operation returned by Jobs.Run.
func runJobsOpExecution(op *runv2.GoogleLongrunningOperation) (*runv2.GoogleCloudRunV2Execution, error) {
	exec := &runv2.GoogleCloudRunV2Execution{}
	if len(op.Metadata) > 0 {
		if err := json.Unmarshal(op.Metadata, exec); err != nil {
			return nil, fmt.Errorf("parsing metadata of operation %s: %w", op.Name, err)
		}
	}
	if exec.Name == "" {
		return nil, fmt.Errorf("operation %s has no execution in its metadata", op.Name)
	}
	return exec, nil
}

// runJobsWaitForExecution polls get every interval until the execution
// completes or fails, writing progress to w as each stage completes.
func runJobsWaitForExecution(ctx context.Context, w io.Writer, get runExecutionGetter,
	interval time.Duration) (*runv2.GoogleCloudRunV2Execution, error) {
	fmt.Fprintln(w, "✓ Creating execution... Done.")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	next := 0
	for {
		exec, err := get(ctx)
		if err != nil {
			return nil, fmt.Errorf("polling execution: %w", err)
		}
		for next < len(runExecStages) {
			stage := runExecStages[next]
			cond := runExecCondition(exec, stage.condition)
			if cond == nil || cond.State != runCondSucceeded {
				break
			}
			fmt.Fprintf(w, "  ✓ %s %s\n", stage.label, runExecStageMessage(exec, stage, cond))
			next++
		}
		if next == len(runExecStages) {
			fmt.Fprintln(w, "Done.")
			return exec, nil
		}
		if cond := runExecFailedCondition(exec); cond != nil {
			fmt.Fprintf(w, "  X %s %s\n", runExecStages[next].label, cond.Message)
			fmt.Fprintln(w, "Executing job failed")
			return exec, errRunExecutionFailed
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func runExecCondition(exec *runv2.GoogleCloudRunV2Execution, condType string) *runv2.GoogleCloudRunV2Condition {
	for _, c := range exec.Conditions {
		if c.Type == condType {
			return c
		}
	}
	return nil
}

// runExecFailedCondition returns the first stage condition that has failed, if any.
func runExecFailedCondition(exec *runv2.GoogleCloudRunV2Execution) *runv2.GoogleCloudRunV2Condition {
	for _, stage := range runExecStages {
		if c := runExecCondition(exec, stage.condition); c != nil && c.State == runCondFailed {
			return c
		}
	}
	return nil
}

func runExecStageMessage(exec *runv2.GoogleCloudRunV2Execution, stage runExecStage,
	cond *runv2.GoogleCloudRunV2Condition) string {
	if stage.condition == runExecCondCompleted {
		return fmt.Sprintf("%d / %d complete", exec.SucceededCount, exec.TaskCount)
	}
	if cond.Message == "" {
		return "Done."
	}
	return cond.Message
}

// runJobsExecDetailsMessage tells the user how to inspect an execution,
// matching gcloud's GetExecutionCreatedMessage.
func runJobsExecDetailsMessage(project, region string, exec *runv2.GoogleCloudRunV2Execution) string {
	name := path.Base(exec.Name)
	msg := fmt.Sprintf("\nView details about this execution by running:\n"+
		"gcloud run jobs executions describe %s\n", name)
	if exec.LogUri != "" {
		msg += fmt.Sprintf("\nOr visit https://console.cloud.google.com/run/jobs/executions/details/%s/%s?project=%s\n",
			region, name, project)
	}
	return msg
}
