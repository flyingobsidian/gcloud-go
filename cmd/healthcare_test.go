package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func healthcareSubgroup(name string) *cobra.Command {
	for _, c := range healthcareCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestHealthcareDatasetsSubcommands(t *testing.T) {
	g := healthcareSubgroup("datasets")
	if g == nil {
		t.Fatal("healthcare datasets missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy", "deidentify",
	})
}

func TestHealthcareOperationsSubcommands(t *testing.T) {
	g := healthcareSubgroup("operations")
	if g == nil {
		t.Fatal("healthcare operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list"})
}
