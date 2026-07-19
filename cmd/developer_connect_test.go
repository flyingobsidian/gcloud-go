package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func developerConnectSubgroup(name string) *cobra.Command {
	for _, c := range developerConnectCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestDeveloperConnectConnectionsSubcommands(t *testing.T) {
	g := developerConnectSubgroup("connections")
	if g == nil {
		t.Fatal("developer-connect connections missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestDeveloperConnectInsightsConfigsSubcommands(t *testing.T) {
	g := developerConnectSubgroup("insights-configs")
	if g == nil {
		t.Fatal("developer-connect insights-configs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestDeveloperConnectOperationsSubcommands(t *testing.T) {
	g := developerConnectSubgroup("operations")
	if g == nil {
		t.Fatal("developer-connect operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list", "wait"})
}
