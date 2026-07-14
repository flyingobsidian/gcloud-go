package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func dpmSubgroup(name string) *cobra.Command {
	for _, c := range deploymentManagerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestDeploymentManagerDeploymentsSubcommands(t *testing.T) {
	g := dpmSubgroup("deployments")
	if g == nil {
		t.Fatal("deployment-manager deployments missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update", "stop", "cancel-preview",
	})
}

func TestDeploymentManagerManifestsSubcommands(t *testing.T) {
	g := dpmSubgroup("manifests")
	if g == nil {
		t.Fatal("deployment-manager manifests missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestDeploymentManagerOperationsSubcommands(t *testing.T) {
	g := dpmSubgroup("operations")
	if g == nil {
		t.Fatal("deployment-manager operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "wait"})
}

func TestDeploymentManagerResourcesSubcommands(t *testing.T) {
	g := dpmSubgroup("resources")
	if g == nil {
		t.Fatal("deployment-manager resources missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestDeploymentManagerTypesSubcommands(t *testing.T) {
	g := dpmSubgroup("types")
	if g == nil {
		t.Fatal("deployment-manager types missing")
	}
	assertSubcommands(t, g, []string{"list", "providers"})
}
