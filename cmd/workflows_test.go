package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	workflowexecutions "google.golang.org/api/workflowexecutions/v1"
	workflows "google.golang.org/api/workflows/v1"
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

// wfIsolateConfig points gcloud config at an empty directory so tests don't
// read the developer's workflows/location property.
func wfIsolateConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "default")
	t.Setenv("CLOUDSDK_WORKFLOWS_LOCATION", "")
	return dir
}

func TestWfResolveLocation(t *testing.T) {
	defer func() { flagWFLocation = "" }()
	dir := wfIsolateConfig(t)

	resolve := func(name, want string, wantWarning bool) {
		t.Helper()
		var buf bytes.Buffer
		got, err := wfResolveLocation(&buf)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
		if gotWarning := strings.Contains(buf.String(), "WARNING: The default location(us-central1) was used"); gotWarning != wantWarning {
			t.Errorf("%s: output %q, want warning %v", name, buf.String(), wantWarning)
		}
	}

	resolve("default", wfDefaultLocation, true)

	if err := os.MkdirAll(filepath.Join(dir, "configurations"), 0o700); err != nil {
		t.Fatal(err)
	}
	ini := "[workflows]\nlocation = europe-west1\n"
	if err := os.WriteFile(filepath.Join(dir, "configurations", "config_default"), []byte(ini), 0o600); err != nil {
		t.Fatal(err)
	}
	resolve("config", "europe-west1", false)

	t.Setenv("CLOUDSDK_WORKFLOWS_LOCATION", "asia-east1")
	resolve("env", "asia-east1", false)

	flagWFLocation = "europe-west2"
	resolve("flag", "europe-west2", false)
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
	wfIsolateConfig(t)
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

func TestWfListFormat(t *testing.T) {
	if got := wfListFormat(""); got != wfExecutionsListFormat {
		t.Errorf("default = %q, want %q", got, wfExecutionsListFormat)
	}
	if got := wfListFormat("json"); got != "json" {
		t.Errorf("explicit = %q, want json", got)
	}
}

func TestWfExecutionsListDefaultTable(t *testing.T) {
	execs := []*workflowexecutions.Execution{{
		Name:      "projects/p/locations/l/workflows/w/executions/e1",
		State:     "SUCCEEDED",
		StartTime: "2026-10-09T00:00:00.042Z",
		EndTime:   "2026-10-09T00:00:01.042Z",
	}}
	var buf bytes.Buffer
	if err := emitFormattedTo(&buf, execs, wfListFormat("")); err != nil {
		t.Fatalf("emitFormattedTo: %v", err)
	}
	want := "NAME                                              STATE      START_TIME                END_TIME\n" +
		"projects/p/locations/l/workflows/w/executions/e1  SUCCEEDED  2026-10-09T00:00:00.042Z  2026-10-09T00:00:01.042Z\n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestWorkflowsListFlags(t *testing.T) {
	list := workflowsSubgroup("list")
	if list == nil {
		t.Fatal("workflows missing list")
	}
	if list.RunE == nil {
		t.Fatal("workflows list is still a stub")
	}
	for _, name := range []string{"location", "limit", "page-size"} {
		if list.Flags().Lookup(name) == nil {
			t.Errorf("workflows list missing --%s", name)
		}
	}
}

// wfPagedWorkflows serves n workflows, two per page, and counts calls.
func wfPagedWorkflows(n int, calls *int) wfWorkflowsPager {
	return func(ctx context.Context, pageToken string) (*workflows.ListWorkflowsResponse, error) {
		*calls++
		start := 0
		if pageToken != "" {
			fmt.Sscanf(pageToken, "%d", &start)
		}
		resp := &workflows.ListWorkflowsResponse{}
		for i := start; i < n && i < start+2; i++ {
			resp.Workflows = append(resp.Workflows, &workflows.Workflow{Name: fmt.Sprintf("wf%d", i)})
		}
		if start+2 < n {
			resp.NextPageToken = fmt.Sprintf("%d", start+2)
		}
		return resp, nil
	}
}

func TestWfListWorkflows(t *testing.T) {
	cases := []struct {
		name      string
		n         int
		limit     int64
		want      int
		wantCalls int
	}{
		{"all pages", 5, 0, 5, 3},
		{"limit stops paging", 5, 3, 3, 2},
		{"empty", 0, 0, 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			got, err := wfListWorkflows(context.Background(), wfPagedWorkflows(tc.n, &calls), tc.limit)
			if err != nil {
				t.Fatalf("wfListWorkflows: %v", err)
			}
			if got == nil {
				t.Fatal("wfListWorkflows returned nil slice")
			}
			if len(got) != tc.want || calls != tc.wantCalls {
				t.Errorf("got %d workflows in %d calls, want %d in %d", len(got), calls, tc.want, tc.wantCalls)
			}
		})
	}
}

func TestWfListWorkflowsError(t *testing.T) {
	page := func(context.Context, string) (*workflows.ListWorkflowsResponse, error) {
		return nil, errors.New("boom")
	}
	if _, err := wfListWorkflows(context.Background(), page, 0); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("want wrapped error, got %v", err)
	}
}

func TestWfWorkflowsListDefaultTable(t *testing.T) {
	var buf bytes.Buffer
	wfs := []*workflows.Workflow{{
		Name:       "projects/p/locations/us-central1/workflows/wf1",
		State:      "ACTIVE",
		RevisionId: "000001-abc",
		UpdateTime: "2026-10-21T00:00:00Z",
	}}
	if err := emitFormattedTo(&buf, wfs, wfWorkflowsListFormat); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"NAME", "STATE", "REVISION_ID", "UPDATE_TIME", "workflows/wf1", "000001-abc"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}
