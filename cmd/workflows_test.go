package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	workflowexecutions "google.golang.org/api/workflowexecutions/v1"
)

func workflowsSubgroup(name string) *cobra.Command {
	for _, c := range workflowsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestWorkflowsHasExecutions(t *testing.T) {
	if workflowsSubgroup("executions") == nil {
		t.Fatal("workflows missing executions subgroup")
	}
}

func TestWorkflowsExecutionsSubcommands(t *testing.T) {
	g := workflowsSubgroup("executions")
	if g == nil {
		t.Fatal("executions missing")
	}
	assertSubcommands(t, g, []string{"cancel", "create", "delete", "describe", "list", "wait"})
}

func TestWorkflowsExecutionName(t *testing.T) {
	flagWFLocation = "us-central1"
	flagWFWorkflow = "wf1"
	defer func() { flagWFLocation, flagWFWorkflow = "", "" }()
	got, err := wfExecutionName("exec1", "my-proj")
	if err != nil {
		t.Fatalf("wfExecutionName error: %v", err)
	}
	want := "projects/my-proj/locations/us-central1/workflows/wf1/executions/exec1"
	if got != want {
		t.Errorf("wfExecutionName = %q, want %q", got, want)
	}
	pass := "projects/x/locations/y/workflows/z/executions/e"
	got, err = wfExecutionName(pass, "ignored")
	if err != nil {
		t.Fatalf("wfExecutionName pass-through error: %v", err)
	}
	if got != pass {
		t.Errorf("wfExecutionName should pass through fully-qualified names")
	}
}

func TestWorkflowsRunFlags(t *testing.T) {
	run := workflowsSubgroup("run")
	if run == nil {
		t.Fatal("workflows missing run")
	}
	for _, name := range []string{"location", "data", "labels", "call-log-level", "execution-history-level"} {
		if run.Flags().Lookup(name) == nil {
			t.Errorf("workflows run missing --%s", name)
		}
	}
}

func TestWfWorkflowPath(t *testing.T) {
	got := wfWorkflowPath("my-proj", "europe-west2", "wf1")
	if want := "projects/my-proj/locations/europe-west2/workflows/wf1"; got != want {
		t.Errorf("wfWorkflowPath = %q, want %q", got, want)
	}
	full := "projects/x/locations/y/workflows/z"
	if got := wfWorkflowPath("ignored", "ignored", full); got != full {
		t.Errorf("wfWorkflowPath should pass through fully-qualified names, got %q", got)
	}
}

func TestWfResolveLocation(t *testing.T) {
	defer func() { flagWFLocation = "" }()

	t.Setenv("CLOUDSDK_WORKFLOWS_LOCATION", "")
	flagWFLocation = "europe-west2"
	var buf bytes.Buffer
	if got := wfResolveLocation(&buf); got != "europe-west2" || buf.Len() != 0 {
		t.Errorf("flag: got %q with output %q", got, buf.String())
	}

	flagWFLocation = ""
	t.Setenv("CLOUDSDK_WORKFLOWS_LOCATION", "asia-east1")
	buf.Reset()
	if got := wfResolveLocation(&buf); got != "asia-east1" || buf.Len() != 0 {
		t.Errorf("env: got %q with output %q", got, buf.String())
	}

	t.Setenv("CLOUDSDK_WORKFLOWS_LOCATION", "")
	buf.Reset()
	if got := wfResolveLocation(&buf); got != wfDefaultLocation {
		t.Errorf("default: got %q, want %q", got, wfDefaultLocation)
	}
	if !strings.Contains(buf.String(), "WARNING: The default location(us-central1) was used") {
		t.Errorf("default: missing warning, got %q", buf.String())
	}
}

func TestWfWaitForExecution(t *testing.T) {
	states := []string{"QUEUED", "ACTIVE", "ACTIVE", "SUCCEEDED"}
	calls := 0
	get := func(context.Context) (*workflowexecutions.Execution, error) {
		e := &workflowexecutions.Execution{State: states[calls]}
		calls++
		return e, nil
	}
	got, err := wfWaitForExecution(context.Background(), get, time.Millisecond, time.Millisecond)
	if err != nil {
		t.Fatalf("wfWaitForExecution: %v", err)
	}
	if got.State != "SUCCEEDED" || calls != len(states) {
		t.Errorf("got state %q after %d calls, want SUCCEEDED after %d", got.State, calls, len(states))
	}
}

func TestWfWaitForExecutionTerminalStates(t *testing.T) {
	for _, state := range []string{"SUCCEEDED", "FAILED", "CANCELLED", "UNAVAILABLE"} {
		get := func(context.Context) (*workflowexecutions.Execution, error) {
			return &workflowexecutions.Execution{State: state}, nil
		}
		got, err := wfWaitForExecution(context.Background(), get, time.Hour, time.Hour)
		if err != nil || got.State != state {
			t.Errorf("state %s: got %v, %v", state, got, err)
		}
	}
}

func TestWfWaitForExecutionErrors(t *testing.T) {
	boom := errors.New("boom")
	get := func(context.Context) (*workflowexecutions.Execution, error) { return nil, boom }
	if _, err := wfWaitForExecution(context.Background(), get, time.Millisecond, time.Millisecond); !errors.Is(err, boom) {
		t.Errorf("getter error: got %v, want wrapped boom", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	active := func(context.Context) (*workflowexecutions.Execution, error) {
		cancel()
		return &workflowexecutions.Execution{State: "ACTIVE"}, nil
	}
	if _, err := wfWaitForExecution(ctx, active, time.Hour, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("cancelled: got %v, want context.Canceled", err)
	}
}
