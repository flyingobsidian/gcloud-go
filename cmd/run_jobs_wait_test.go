package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	runv2 "google.golang.org/api/run/v2"
)

const testRunExecName = "projects/p/locations/europe-west2/jobs/myjob/executions/myjob-abc12"

func testRunCond(condType, state, message string) *runv2.GoogleCloudRunV2Condition {
	return &runv2.GoogleCloudRunV2Condition{Type: condType, State: state, Message: message}
}

// sequenceGetter returns each execution in turn, repeating the last one.
func sequenceGetter(execs ...*runv2.GoogleCloudRunV2Execution) runExecutionGetter {
	i := 0
	return func(ctx context.Context) (*runv2.GoogleCloudRunV2Execution, error) {
		e := execs[i]
		if i < len(execs)-1 {
			i++
		}
		return e, nil
	}
}

func TestRunJobsWaitForExecutionSuccess(t *testing.T) {
	get := sequenceGetter(
		&runv2.GoogleCloudRunV2Execution{Name: testRunExecName, TaskCount: 1},
		&runv2.GoogleCloudRunV2Execution{Name: testRunExecName, TaskCount: 1, Conditions: []*runv2.GoogleCloudRunV2Condition{
			testRunCond("ResourcesAvailable", runCondSucceeded, "Provisioned imported containers."),
			testRunCond("Started", runCondSucceeded, "Started deployed execution in 42.42s."),
			testRunCond("Completed", "CONDITION_RECONCILING", ""),
		}},
		&runv2.GoogleCloudRunV2Execution{Name: testRunExecName, TaskCount: 1, SucceededCount: 1, Conditions: []*runv2.GoogleCloudRunV2Condition{
			testRunCond("ResourcesAvailable", runCondSucceeded, "Provisioned imported containers."),
			testRunCond("Started", runCondSucceeded, "Started deployed execution in 42.42s."),
			testRunCond("Completed", runCondSucceeded, "Execution completed successfully."),
		}},
	)
	var buf bytes.Buffer
	got, err := runJobsWaitForExecution(context.Background(), &buf, get, time.Millisecond)
	if err != nil {
		t.Fatalf("runJobsWaitForExecution: %v", err)
	}
	if got.SucceededCount != 1 {
		t.Errorf("SucceededCount = %d, want 1", got.SucceededCount)
	}
	want := "✓ Creating execution... Done.\n" +
		"  ✓ Provisioning resources... Provisioned imported containers.\n" +
		"  ✓ Starting execution... Started deployed execution in 42.42s.\n" +
		"  ✓ Running execution... 1 / 1 complete\n" +
		"Done.\n"
	if buf.String() != want {
		t.Errorf("output:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestRunJobsWaitForExecutionStageWithoutMessage(t *testing.T) {
	get := sequenceGetter(&runv2.GoogleCloudRunV2Execution{Name: testRunExecName, TaskCount: 2, SucceededCount: 2,
		Conditions: []*runv2.GoogleCloudRunV2Condition{
			testRunCond("ResourcesAvailable", runCondSucceeded, ""),
			testRunCond("Started", runCondSucceeded, ""),
			testRunCond("Completed", runCondSucceeded, ""),
		}})
	var buf bytes.Buffer
	if _, err := runJobsWaitForExecution(context.Background(), &buf, get, time.Millisecond); err != nil {
		t.Fatalf("runJobsWaitForExecution: %v", err)
	}
	for _, line := range []string{"  ✓ Provisioning resources... Done.\n", "  ✓ Running execution... 2 / 2 complete\n"} {
		if !strings.Contains(buf.String(), line) {
			t.Errorf("output missing %q:\n%s", line, buf.String())
		}
	}
}

func TestRunJobsWaitForExecutionFailed(t *testing.T) {
	get := sequenceGetter(&runv2.GoogleCloudRunV2Execution{Name: testRunExecName, TaskCount: 1, FailedCount: 1,
		Conditions: []*runv2.GoogleCloudRunV2Condition{
			testRunCond("ResourcesAvailable", runCondSucceeded, "Provisioned imported containers."),
			testRunCond("Started", runCondSucceeded, "Started deployed execution in 1.00s."),
			testRunCond("Completed", runCondFailed, "Task myjob-abc12-task0 failed with message: exit 1."),
		}})
	var buf bytes.Buffer
	_, err := runJobsWaitForExecution(context.Background(), &buf, get, time.Millisecond)
	if !errors.Is(err, errRunExecutionFailed) {
		t.Fatalf("err = %v, want errRunExecutionFailed", err)
	}
	want := "  X Running execution... Task myjob-abc12-task0 failed with message: exit 1.\nExecuting job failed\n"
	if !strings.HasSuffix(buf.String(), want) {
		t.Errorf("output:\n%s\nwant suffix:\n%s", buf.String(), want)
	}
}

// A failure of the terminal condition must stop polling even when an earlier
// stage never succeeded.
func TestRunJobsWaitForExecutionFailedBeforeStarting(t *testing.T) {
	get := sequenceGetter(&runv2.GoogleCloudRunV2Execution{Name: testRunExecName,
		Conditions: []*runv2.GoogleCloudRunV2Condition{
			testRunCond("Completed", runCondFailed, "Image not found."),
		}})
	var buf bytes.Buffer
	_, err := runJobsWaitForExecution(context.Background(), &buf, get, time.Millisecond)
	if !errors.Is(err, errRunExecutionFailed) {
		t.Fatalf("err = %v, want errRunExecutionFailed", err)
	}
	if !strings.Contains(buf.String(), "  X Provisioning resources... Image not found.\n") {
		t.Errorf("output:\n%s", buf.String())
	}
}

func TestRunJobsWaitForExecutionGetterError(t *testing.T) {
	boom := errors.New("boom")
	get := func(ctx context.Context) (*runv2.GoogleCloudRunV2Execution, error) { return nil, boom }
	_, err := runJobsWaitForExecution(context.Background(), &bytes.Buffer{}, get, time.Millisecond)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrapped boom", err)
	}
}

func TestRunJobsWaitForExecutionCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	get := func(context.Context) (*runv2.GoogleCloudRunV2Execution, error) {
		cancel()
		return &runv2.GoogleCloudRunV2Execution{Name: testRunExecName}, nil
	}
	_, err := runJobsWaitForExecution(ctx, &bytes.Buffer{}, get, time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestRunJobsOpExecution(t *testing.T) {
	op := &runv2.GoogleLongrunningOperation{
		Name:     "projects/p/locations/europe-west2/operations/op1",
		Metadata: []byte(`{"@type":"type.googleapis.com/google.cloud.run.v2.Execution","name":"` + testRunExecName + `"}`),
	}
	got, err := runJobsOpExecution(op)
	if err != nil {
		t.Fatalf("runJobsOpExecution: %v", err)
	}
	if got.Name != testRunExecName {
		t.Errorf("Name = %q, want %q", got.Name, testRunExecName)
	}

	if _, err := runJobsOpExecution(&runv2.GoogleLongrunningOperation{Name: "op2"}); err == nil {
		t.Error("expected an error for an operation without metadata")
	}
}

func TestRunJobsExecDetailsMessage(t *testing.T) {
	exec := &runv2.GoogleCloudRunV2Execution{Name: testRunExecName}
	want := "\nView details about this execution by running:\n" +
		"gcloud run jobs executions describe myjob-abc12\n"
	if got := runJobsExecDetailsMessage("p", "europe-west2", exec); got != want {
		t.Errorf("without log URI:\n%q\nwant:\n%q", got, want)
	}

	exec.LogUri = "https://console.cloud.google.com/logs/viewer?project=p"
	want += "\nOr visit https://console.cloud.google.com/run/jobs/executions/details/europe-west2/myjob-abc12?project=p\n"
	if got := runJobsExecDetailsMessage("p", "europe-west2", exec); got != want {
		t.Errorf("with log URI:\n%q\nwant:\n%q", got, want)
	}
}
