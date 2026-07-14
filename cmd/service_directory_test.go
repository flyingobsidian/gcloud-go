package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func sdSubgroup(name string) *cobra.Command {
	for _, c := range serviceDirectoryCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestServiceDirectoryLocationsSubcommands(t *testing.T) {
	g := sdSubgroup("locations")
	if g == nil {
		t.Fatal("service-directory locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestServiceDirectoryNamespacesSubcommands(t *testing.T) {
	g := sdSubgroup("namespaces")
	if g == nil {
		t.Fatal("service-directory namespaces missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "create", "delete", "describe", "get-iam-policy",
		"list", "remove-iam-policy-binding", "set-iam-policy", "update",
	})
}

func TestServiceDirectoryServicesSubcommands(t *testing.T) {
	g := sdSubgroup("services")
	if g == nil {
		t.Fatal("service-directory services missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "resolve", "update"})
}

func TestServiceDirectoryEndpointsSubcommands(t *testing.T) {
	g := sdSubgroup("endpoints")
	if g == nil {
		t.Fatal("service-directory endpoints missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
