package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func workbenchSubgroup(name string) *cobra.Command {
	for _, c := range workbenchCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestWorkbenchInstancesSubcommands(t *testing.T) {
	g := workbenchSubgroup("instances")
	if g == nil {
		t.Fatal("workbench instances missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "check-instance-upgradability", "create",
		"delete", "describe", "diagnose", "get-config", "get-iam-policy",
		"list", "remove-iam-policy-binding", "reset", "resize-disk",
		"restore", "rollback", "set-iam-policy", "start", "stop", "update", "upgrade",
	})
}

func TestWorkbenchExecutionsSubcommands(t *testing.T) {
	g := workbenchSubgroup("executions")
	if g == nil {
		t.Fatal("workbench executions missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestWorkbenchSchedulesSubcommands(t *testing.T) {
	g := workbenchSubgroup("schedules")
	if g == nil {
		t.Fatal("workbench schedules missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "pause", "resume", "update"})
}
