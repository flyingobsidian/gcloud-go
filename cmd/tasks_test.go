package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func tasksSubgroup(name string) *cobra.Command {
	for _, c := range tasksCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestTasksLocationsSubcommands(t *testing.T) {
	g := tasksSubgroup("locations")
	if g == nil {
		t.Fatal("tasks locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestTasksCmekConfigSubcommands(t *testing.T) {
	g := tasksSubgroup("cmek-config")
	if g == nil {
		t.Fatal("tasks cmek-config missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestTasksQueuesSubcommands(t *testing.T) {
	g := tasksSubgroup("queues")
	if g == nil {
		t.Fatal("tasks queues missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "create", "delete", "describe", "get-iam-policy",
		"list", "pause", "purge", "remove-iam-policy-binding", "resume",
		"set-iam-policy", "update",
	})
}
