package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func batchSubgroup(name string) *cobra.Command {
	for _, c := range batchCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestBatchJobsSubcommands(t *testing.T) {
	g := batchSubgroup("jobs")
	if g == nil {
		t.Fatal("batch jobs missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list", "submit"})
}

func TestBatchTasksSubcommands(t *testing.T) {
	g := batchSubgroup("tasks")
	if g == nil {
		t.Fatal("batch tasks missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}
