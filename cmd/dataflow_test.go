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
