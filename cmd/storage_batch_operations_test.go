package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func storageSubgroup(name string) *cobra.Command {
	for _, c := range storageCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestStorageBatchOperationsSubgroups(t *testing.T) {
	g := storageSubgroup("batch-operations")
	if g == nil {
		t.Fatal("storage batch-operations missing")
	}
	assertSubcommands(t, g, []string{"bucket-operations", "jobs"})

	jobs := findSub(g, "jobs")
	if jobs == nil {
		t.Fatal("batch-operations jobs subgroup missing")
	}
	assertSubcommands(t, jobs, []string{"cancel", "create", "delete", "describe", "list"})

	buckOps := findSub(g, "bucket-operations")
	if buckOps == nil {
		t.Fatal("batch-operations bucket-operations subgroup missing")
	}
	assertSubcommands(t, buckOps, []string{"describe", "list"})
}
