package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

func TestWfListParent(t *testing.T) {
	defer func() { flagWFLocation, flagWFWorkflow = "", "" }()
	t.Setenv("CLOUDSDK_WORKFLOWS_LOCATION", "")
	cases := []struct {
		name, location, workflowFlag string
		args                         []string
		want                         string
		wantWarning                  bool
	}{
		{"positional", "europe-west2", "", []string{"wf1"}, "projects/p/locations/europe-west2/workflows/wf1", false},
		{"flag", "europe-west2", "wf2", nil, "projects/p/locations/europe-west2/workflows/wf2", false},
		{"positional beats flag", "europe-west2", "wf2", []string{"wf1"}, "projects/p/locations/europe-west2/workflows/wf1", false},
		{"default location", "", "", []string{"wf1"}, "projects/p/locations/us-central1/workflows/wf1", true},
		{"fully qualified", "", "", []string{"projects/x/locations/y/workflows/z"}, "projects/x/locations/y/workflows/z", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flagWFLocation, flagWFWorkflow = tc.location, tc.workflowFlag
			var buf bytes.Buffer
			got, err := wfListParent(&buf, tc.args, "p")
			if err != nil {
				t.Fatalf("wfListParent: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
			if gotWarning := buf.Len() > 0; gotWarning != tc.wantWarning {
				t.Errorf("warning = %q, want warning %v", buf.String(), tc.wantWarning)
			}
		})
	}
}

func TestWfListParentMissingWorkflow(t *testing.T) {
	flagWFLocation, flagWFWorkflow = "europe-west2", ""
	defer func() { flagWFLocation = "" }()
	if _, err := wfListParent(&bytes.Buffer{}, nil, "p"); err == nil {
		t.Error("want error when no workflow is given")
	}
}

// wfPagedExecutions serves n executions, two per page, and counts calls.
func wfPagedExecutions(n int, calls *int) wfExecutionsPager {
	return func(ctx context.Context, pageToken string) (*workflowexecutions.ListExecutionsResponse, error) {
		*calls++
		start := 0
		if pageToken != "" {
			fmt.Sscanf(pageToken, "%d", &start)
		}
		end := min(start+2, n)
		resp := &workflowexecutions.ListExecutionsResponse{}
		for i := start; i < end; i++ {
			resp.Executions = append(resp.Executions, &workflowexecutions.Execution{Name: fmt.Sprint(i)})
		}
		if end < n {
			resp.NextPageToken = fmt.Sprint(end)
		}
		return resp, nil
	}
}

func TestWfListExecutions(t *testing.T) {
	cases := []struct {
		name               string
		available          int
		limit              int64
		wantLen, wantCalls int
	}{
		{"no limit", 5, 0, 5, 3},
		{"limit within first page", 5, 1, 1, 1},
		{"limit across pages", 5, 3, 3, 2},
		{"limit above available", 3, 10, 3, 2},
		{"empty", 0, 0, 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			got, err := wfListExecutions(context.Background(), wfPagedExecutions(tc.available, &calls), tc.limit)
			if err != nil {
				t.Fatalf("wfListExecutions: %v", err)
			}
			if got == nil {
				t.Fatal("got nil slice, want non-nil")
			}
			if len(got) != tc.wantLen || calls != tc.wantCalls {
				t.Errorf("got %d executions in %d calls, want %d in %d", len(got), calls, tc.wantLen, tc.wantCalls)
			}
			for i, e := range got {
				if e.Name != fmt.Sprint(i) {
					t.Errorf("execution %d name = %q", i, e.Name)
				}
			}
		})
	}
}

func TestWfListExecutionsError(t *testing.T) {
	boom := errors.New("boom")
	page := func(context.Context, string) (*workflowexecutions.ListExecutionsResponse, error) { return nil, boom }
	if _, err := wfListExecutions(context.Background(), page, 0); !errors.Is(err, boom) {
		t.Errorf("got %v, want wrapped boom", err)
	}
}

func TestWorkflowsExecutionsListArgs(t *testing.T) {
	list := findSub(workflowsSubgroup("executions"), "list")
	if list == nil {
		t.Fatal("executions list missing")
	}
	if err := list.Args(list, []string{"wf1"}); err != nil {
		t.Errorf("WORKFLOW arg rejected: %v", err)
	}
	if err := list.Args(list, []string{"a", "b"}); err == nil {
		t.Error("two positional args accepted")
	}
	if f := list.Flags().Lookup("workflow"); f == nil || len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0 {
		t.Error("--workflow should exist and be optional")
	}
}
