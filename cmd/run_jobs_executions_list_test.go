package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	runv1 "google.golang.org/api/run/v1"
)

func TestRunJobsExecLabelSelector(t *testing.T) {
	if got := runJobsExecLabelSelector("my-job"); got != "run.googleapis.com/job = my-job" {
		t.Errorf("got %q", got)
	}
	if got := runJobsExecLabelSelector(""); got != "" {
		t.Errorf("empty job: got %q, want empty", got)
	}
}

// v1ExecPages serves n executions, two per page, and records continue tokens.
func v1ExecPages(n int, tokens *[]string) runJobsExecPager {
	return func(ctx context.Context, continueToken string) (*runv1.ListExecutionsResponse, error) {
		*tokens = append(*tokens, continueToken)
		start := 0
		if continueToken != "" {
			fmt.Sscanf(continueToken, "%d", &start)
		}
		end := min(start+2, n)
		resp := &runv1.ListExecutionsResponse{Metadata: &runv1.ListMeta{}}
		for i := start; i < end; i++ {
			resp.Items = append(resp.Items, &runv1.Execution{Metadata: &runv1.ObjectMeta{Name: fmt.Sprint(i)}})
		}
		if end < n {
			resp.Metadata.Continue = fmt.Sprint(end)
		}
		return resp, nil
	}
}

func TestRunJobsListV1Executions(t *testing.T) {
	cases := []struct {
		available int
		wantCalls int
	}{{0, 1}, {2, 1}, {5, 3}}
	for _, tc := range cases {
		var tokens []string
		got, err := runJobsListV1Executions(context.Background(), v1ExecPages(tc.available, &tokens))
		if err != nil {
			t.Fatalf("%d available: %v", tc.available, err)
		}
		if got == nil {
			t.Fatalf("%d available: got nil slice, want non-nil", tc.available)
		}
		if len(got) != tc.available || len(tokens) != tc.wantCalls {
			t.Errorf("%d available: got %d in %d calls, want %d in %d", tc.available, len(got), len(tokens), tc.available, tc.wantCalls)
		}
	}
}

func TestRunJobsListV1ExecutionsNoListMeta(t *testing.T) {
	page := func(context.Context, string) (*runv1.ListExecutionsResponse, error) {
		return &runv1.ListExecutionsResponse{Items: []*runv1.Execution{{}}}, nil
	}
	got, err := runJobsListV1Executions(context.Background(), page)
	if err != nil || len(got) != 1 {
		t.Errorf("got %d, %v; want 1 execution", len(got), err)
	}
}

func TestRunJobsListV1ExecutionsError(t *testing.T) {
	boom := errors.New("boom")
	page := func(context.Context, string) (*runv1.ListExecutionsResponse, error) { return nil, boom }
	if _, err := runJobsListV1Executions(context.Background(), page); !errors.Is(err, boom) {
		t.Errorf("got %v, want boom", err)
	}
}

// testV1Exec builds an execution; started is the Started condition status
// ("" for no Started condition).
func testV1Exec(name, created, started, startedAt string) *runv1.Execution {
	e := &runv1.Execution{Metadata: &runv1.ObjectMeta{Name: name, CreationTimestamp: created}, Status: &runv1.ExecutionStatus{}}
	if started != "" {
		e.Status.Conditions = []*runv1.GoogleCloudRunV1Condition{{Type: "Started", Status: started, LastTransitionTime: startedAt}}
	}
	return e
}

func TestRunJobsSortExecutions(t *testing.T) {
	execs := []*runv1.Execution{
		testV1Exec("old", "2026-10-01T00:00:00Z", "True", "2026-10-01T00:00:05Z"),
		testV1Exec("queued-old", "2026-10-02T00:00:00Z", "", ""),
		testV1Exec("new", "2026-10-03T00:00:00Z", "True", "2026-10-03T00:00:05Z"),
		testV1Exec("pending", "2026-10-04T00:00:00Z", "Unknown", "2026-10-04T00:00:01Z"),
		testV1Exec("failed-start", "2026-10-02T12:00:00Z", "False", "2026-10-02T12:00:01Z"),
		testV1Exec("b-tie", "2026-10-01T00:00:00Z", "True", "2026-10-01T00:00:05Z"),
		{Metadata: &runv1.ObjectMeta{Name: "no-status", CreationTimestamp: "2026-10-01T00:00:00Z"}},
	}
	runJobsSortExecutions(execs)
	want := []string{"pending", "queued-old", "no-status", "new", "failed-start", "b-tie", "old"}
	for i, e := range execs {
		if e.Metadata.Name != want[i] {
			got := make([]string, len(execs))
			for j, e := range execs {
				got[j] = e.Metadata.Name
			}
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestRunJobsSortAndLimit(t *testing.T) {
	mk := func() []*runv1.Execution {
		return []*runv1.Execution{
			testV1Exec("a", "2026-10-01T00:00:00Z", "True", "2026-10-01T00:00:01Z"),
			testV1Exec("c", "2026-10-03T00:00:00Z", "True", "2026-10-03T00:00:01Z"),
			testV1Exec("b", "2026-10-02T00:00:00Z", "True", "2026-10-02T00:00:01Z"),
		}
	}
	cases := []struct {
		limit int64
		want  []string
	}{
		{0, []string{"c", "b", "a"}},
		{2, []string{"c", "b"}},
		{5, []string{"c", "b", "a"}},
	}
	for _, tc := range cases {
		got := runJobsSortAndLimit(mk(), tc.limit)
		if len(got) != len(tc.want) {
			t.Fatalf("limit %d: got %d executions, want %d", tc.limit, len(got), len(tc.want))
		}
		for i, e := range got {
			if e.Metadata.Name != tc.want[i] {
				t.Errorf("limit %d: [%d] = %q, want %q", tc.limit, i, e.Metadata.Name, tc.want[i])
			}
		}
	}
}

func TestRunJobsExecListFlags(t *testing.T) {
	for name, def := range map[string]string{"limit": "0", "page-size": "0", "job": "", "region": "", "format": ""} {
		f := runJobsExecListCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("--%s missing", name)
			continue
		}
		if f.DefValue != def {
			t.Errorf("--%s default = %q, want %q", name, f.DefValue, def)
		}
	}
	if len(runJobsExecListCmd.Flags().Lookup("job").Annotations[cobra.BashCompOneRequiredFlag]) > 0 {
		t.Error("--job should be optional for list")
	}
	if len(runJobsExecDescribeCmd.Flags().Lookup("job").Annotations[cobra.BashCompOneRequiredFlag]) == 0 {
		t.Error("--job should stay required for describe")
	}
}

// testTableExec builds an execution with the fields the default table reads.
func testTableExec(name, completed string, running, succeeded, tasks int64, author string) *runv1.Execution {
	e := &runv1.Execution{
		Metadata: &runv1.ObjectMeta{
			Name:              name,
			CreationTimestamp: "2026-10-09T00:00:00.042Z",
			Labels:            map[string]string{"run.googleapis.com/job": "my-job", "cloud.googleapis.com/location": "europe-west2"},
			Annotations:       map[string]string{"serving.knative.dev/creator": author},
		},
		Spec:   &runv1.ExecutionSpec{TaskCount: tasks},
		Status: &runv1.ExecutionStatus{RunningCount: running, SucceededCount: succeeded},
	}
	if completed != "" {
		e.Status.Conditions = []*runv1.GoogleCloudRunV1Condition{{Type: "Completed", Status: completed}}
	}
	return e
}

func TestEmitRunJobsExecTable(t *testing.T) {
	execs := []*runv1.Execution{
		testTableExec("my-job-abc12", "Unknown", 1, 0, 1, "me@example.com"),
		testTableExec("my-job-def34", "True", 0, 2, 2, "me@example.com"),
		testTableExec("my-job-ghi56", "False", 0, 0, 1, ""),
	}
	var out, errOut bytes.Buffer
	if err := emitRunJobsExecTable(&out, &errOut, execs, time.UTC); err != nil {
		t.Fatalf("emitRunJobsExecTable: %v", err)
	}
	want := "   JOB     EXECUTION     REGION        RUNNING  COMPLETE  CREATED                  RUN BY\n" +
		"…  my-job  my-job-abc12  europe-west2  1        0 / 1     2026-10-09 00:00:00 UTC  me@example.com\n" +
		"✔  my-job  my-job-def34  europe-west2  0        2 / 2     2026-10-09 00:00:00 UTC  me@example.com\n" +
		"X  my-job  my-job-ghi56  europe-west2  0        0 / 1     2026-10-09 00:00:00 UTC\n"
	if out.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
	}
	if errOut.Len() != 0 {
		t.Errorf("unexpected stderr %q", errOut.String())
	}
}

func TestEmitRunJobsExecTableEmpty(t *testing.T) {
	var out, errOut bytes.Buffer
	if err := emitRunJobsExecTable(&out, &errOut, nil, time.UTC); err != nil {
		t.Fatalf("emitRunJobsExecTable: %v", err)
	}
	if out.Len() != 0 || errOut.String() != "Listed 0 items.\n" {
		t.Errorf("stdout %q, stderr %q", out.String(), errOut.String())
	}
}

func TestRunJobsExecReadySymbolNoStatus(t *testing.T) {
	if got := runJobsExecReadySymbol(&runv1.Execution{}); got != "…" {
		t.Errorf("got %q, want …", got)
	}
}

func TestRunJobsExecCreated(t *testing.T) {
	bst := time.FixedZone("BST", 3600)
	if got := runJobsExecCreated("2026-10-09T00:00:00Z", bst); got != "2026-10-09 01:00:00 BST" {
		t.Errorf("got %q", got)
	}
	if got := runJobsExecCreated("not-a-time", time.UTC); got != "not-a-time" {
		t.Errorf("unparseable: got %q", got)
	}
}
