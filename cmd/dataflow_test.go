package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func dataflowSubgroup(name string) *cobra.Command {
	for _, c := range dataflowCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestDataflowFlexTemplateSubcommands(t *testing.T) {
	g := dataflowSubgroup("flex-template")
	if g == nil {
		t.Fatal("dataflow flex-template missing")
	}
	assertSubcommands(t, g, []string{"build", "run"})
}

func TestDataflowSnapshotsSubcommands(t *testing.T) {
	g := dataflowSubgroup("snapshots")
	if g == nil {
		t.Fatal("dataflow snapshots missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestDataflowYamlSubcommands(t *testing.T) {
	g := dataflowSubgroup("yaml")
	if g == nil {
		t.Fatal("dataflow yaml missing")
	}
	assertSubcommands(t, g, []string{"run"})
}

// #1773: pause/resume subcommands and --enable-turnkey-alerts flag.
func TestDataflow_1773_JobsSubcommands(t *testing.T) {
	g := dataflowSubgroup("jobs")
	if g == nil {
		t.Fatal("dataflow jobs missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "drain", "list", "pause", "resume", "run"})
}

func TestDataflow_1773_TurnkeyAlertsFlag(t *testing.T) {
	if dataflowJobsRunCmd.Flags().Lookup("enable-turnkey-alerts") == nil {
		t.Error("dataflow jobs run: --enable-turnkey-alerts flag missing")
	}
	ft := dataflowSubgroup("flex-template")
	if ft == nil {
		t.Fatal("dataflow flex-template missing")
	}
	var run *cobra.Command
	for _, c := range ft.Commands() {
		if c.Name() == "run" {
			run = c
			break
		}
	}
	if run == nil {
		t.Fatal("dataflow flex-template run missing")
	}
	if run.Flags().Lookup("enable-turnkey-alerts") == nil {
		t.Error("dataflow flex-template run: --enable-turnkey-alerts flag missing")
	}
}

func TestAppendExperiment(t *testing.T) {
	got := appendExperiment(nil, "e1")
	if len(got) != 1 || got[0] != "e1" {
		t.Errorf("expected [e1], got %v", got)
	}
	got = appendExperiment(got, "e1")
	if len(got) != 1 {
		t.Errorf("expected dedup to keep length 1, got %v", got)
	}
	got = appendExperiment(got, "e2")
	if len(got) != 2 || got[1] != "e2" {
		t.Errorf("expected [e1 e2], got %v", got)
	}
}

func TestStatusToAPIFilter(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"active", "ACTIVE"},
		{"Active", "ACTIVE"},
		{"ACTIVE", "ACTIVE"},
		{"terminated", "TERMINATED"},
		{"all", "ALL"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := statusToAPIFilter(tt.input)
			if got != tt.want {
				t.Errorf("statusToAPIFilter(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
