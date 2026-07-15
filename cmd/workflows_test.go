package cmd

import (
	"testing"

	"github.com/spf13/cobra"
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
