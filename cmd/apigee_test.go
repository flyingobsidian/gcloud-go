package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func apigeeSubgroup(name string) *cobra.Command {
	for _, c := range apigeeCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestApigeeApisSubcommands(t *testing.T) {
	g := apigeeSubgroup("apis")
	if g == nil {
		t.Fatal("apigee apis missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "list"})
}

func TestApigeeApplicationsSubcommands(t *testing.T) {
	g := apigeeSubgroup("applications")
	if g == nil {
		t.Fatal("apigee applications missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestApigeeArchivesSubcommands(t *testing.T) {
	g := apigeeSubgroup("archives")
	if g == nil {
		t.Fatal("apigee archives missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestApigeeDeploymentsSubcommands(t *testing.T) {
	g := apigeeSubgroup("deployments")
	if g == nil {
		t.Fatal("apigee deployments missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestApigeeDevelopersSubcommands(t *testing.T) {
	g := apigeeSubgroup("developers")
	if g == nil {
		t.Fatal("apigee developers missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestApigeeEnvironmentsSubcommands(t *testing.T) {
	g := apigeeSubgroup("environments")
	if g == nil {
		t.Fatal("apigee environments missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "update"})
}

func TestApigeeOperationsSubcommands(t *testing.T) {
	g := apigeeSubgroup("operations")
	if g == nil {
		t.Fatal("apigee operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestApigeeOrganizationsSubcommands(t *testing.T) {
	g := apigeeSubgroup("organizations")
	if g == nil {
		t.Fatal("apigee organizations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "provision"})
}

func TestApigeeProductsSubcommands(t *testing.T) {
	g := apigeeSubgroup("products")
	if g == nil {
		t.Fatal("apigee products missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
