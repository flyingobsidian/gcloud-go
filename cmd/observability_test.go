package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func observabilitySubgroup(name string) *cobra.Command {
	for _, c := range observabilityCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestObservabilityScopesSubcommands(t *testing.T) {
	g := observabilitySubgroup("scopes")
	if g == nil {
		t.Fatal("observability scopes missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestObservabilityTraceScopesSubcommands(t *testing.T) {
	g := observabilitySubgroup("trace-scopes")
	if g == nil {
		t.Fatal("observability trace-scopes missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
