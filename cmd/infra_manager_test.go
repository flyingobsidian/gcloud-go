package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func imSubgroup(name string) *cobra.Command {
	for _, c := range infraManagerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestInfraManagerAutoMigrationConfigSubcommands(t *testing.T) {
	g := imSubgroup("automigrationconfig")
	if g == nil {
		t.Fatal("infra-manager automigrationconfig missing")
	}
	assertSubcommands(t, g, []string{"describe", "disable-auto-migration", "enable-auto-migration"})
}

func TestInfraManagerDeploymentsSubcommands(t *testing.T) {
	g := imSubgroup("deployments")
	if g == nil {
		t.Fatal("infra-manager deployments missing")
	}
	assertSubcommands(t, g, []string{"apply", "delete", "describe", "export-lock", "export-statefile", "import-statefile", "list", "lock", "unlock"})
}

func TestInfraManagerPreviewsSubcommands(t *testing.T) {
	g := imSubgroup("previews")
	if g == nil {
		t.Fatal("infra-manager previews missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "export", "list"})
}

func TestInfraManagerResourceChangesSubcommands(t *testing.T) {
	g := imSubgroup("resource-changes")
	if g == nil {
		t.Fatal("infra-manager resource-changes missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestInfraManagerResourceDriftsSubcommands(t *testing.T) {
	g := imSubgroup("resource-drifts")
	if g == nil {
		t.Fatal("infra-manager resource-drifts missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestInfraManagerResourcesSubcommands(t *testing.T) {
	g := imSubgroup("resources")
	if g == nil {
		t.Fatal("infra-manager resources missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestInfraManagerRevisionsSubcommands(t *testing.T) {
	g := imSubgroup("revisions")
	if g == nil {
		t.Fatal("infra-manager revisions missing")
	}
	assertSubcommands(t, g, []string{"describe", "export-statefile", "list"})
}

func TestInfraManagerTerraformVersionsSubcommands(t *testing.T) {
	g := imSubgroup("terraform-versions")
	if g == nil {
		t.Fatal("infra-manager terraform-versions missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}
