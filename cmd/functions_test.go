package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func functionsSubgroup(name string) *cobra.Command {
	for _, c := range functionsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestFunctionsIamCommandsRegistered(t *testing.T) {
	want := []string{"add-iam-policy-binding", "get-iam-policy", "remove-iam-policy-binding", "set-iam-policy"}
	got := map[string]bool{}
	for _, c := range functionsCmd.Commands() {
		got[c.Name()] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("functions is missing %q subcommand", w)
		}
	}
}

func TestFunctionsEventTypesSubcommands(t *testing.T) {
	g := functionsSubgroup("event-types")
	if g == nil {
		t.Fatal("functions event-types missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestFunctionsLogsSubcommands(t *testing.T) {
	g := functionsSubgroup("logs")
	if g == nil {
		t.Fatal("functions logs missing")
	}
	assertSubcommands(t, g, []string{"read"})
}

func TestFunctionsRegionsSubcommands(t *testing.T) {
	g := functionsSubgroup("regions")
	if g == nil {
		t.Fatal("functions regions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestFunctionsRuntimesSubcommands(t *testing.T) {
	g := functionsSubgroup("runtimes")
	if g == nil {
		t.Fatal("functions runtimes missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestFunctionsEventTypesListStatic(t *testing.T) {
	if len(gen1TriggerEvents) == 0 {
		t.Fatal("gen1TriggerEvents must not be empty")
	}
	if len(gen2TriggerEvents) == 0 {
		t.Fatal("gen2TriggerEvents must not be empty")
	}
	// Sanity: expected canonical events must be present.
	seen := map[string]bool{}
	for _, e := range gen1TriggerEvents {
		seen[e.EventType] = true
	}
	for _, w := range []string{
		"google.pubsub.topic.publish",
		"google.storage.object.finalize",
	} {
		if !seen[w] {
			t.Errorf("expected gen1 event %q in static list", w)
		}
	}
}
