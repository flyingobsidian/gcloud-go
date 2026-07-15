package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func sourceManagerSubgroup(name string) *cobra.Command {
	for _, c := range sourceManagerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestSourceManagerHasSubgroups(t *testing.T) {
	for _, want := range []string{"instances", "locations", "operations", "repos"} {
		if sourceManagerSubgroup(want) == nil {
			t.Errorf("source-manager missing subgroup %q", want)
		}
	}
}

func TestSourceManagerInstancesSubcommands(t *testing.T) {
	g := sourceManagerSubgroup("instances")
	if g == nil {
		t.Fatal("instances missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestSourceManagerLocationsSubcommands(t *testing.T) {
	g := sourceManagerSubgroup("locations")
	if g == nil {
		t.Fatal("locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestSourceManagerOperationsSubcommands(t *testing.T) {
	g := sourceManagerSubgroup("operations")
	if g == nil {
		t.Fatal("operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list", "wait"})
}

func TestSourceManagerReposSubcommands(t *testing.T) {
	g := sourceManagerSubgroup("repos")
	if g == nil {
		t.Fatal("repos missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "create", "delete", "describe",
		"get-iam-policy", "list", "remove-iam-policy-binding",
		"set-iam-policy", "update",
	})
}

func TestSMOperationName(t *testing.T) {
	got := smOperationName("op123", "my-project", "us-central1")
	want := "projects/my-project/locations/us-central1/operations/op123"
	if got != want {
		t.Errorf("smOperationName = %q, want %q", got, want)
	}
	pass := "projects/x/locations/y/operations/z"
	if smOperationName(pass, "ignored", "ignored") != pass {
		t.Errorf("smOperationName should pass through fully-qualified names")
	}
}
