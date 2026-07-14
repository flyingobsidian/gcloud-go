package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func sccSubgroup(name string) *cobra.Command {
	for _, c := range sccCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func sccNestedSubgroup(parent, name string) *cobra.Command {
	p := sccSubgroup(parent)
	if p == nil {
		return nil
	}
	for _, c := range p.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestSccAssetsSubcommands(t *testing.T) {
	g := sccSubgroup("assets")
	if g == nil {
		t.Fatal("scc assets missing")
	}
	assertSubcommands(t, g, []string{"describe", "group", "list", "run-discovery", "update-security-marks"})
}

func TestSccBQExportsSubcommands(t *testing.T) {
	g := sccSubgroup("bqexports")
	if g == nil {
		t.Fatal("scc bqexports missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestSccCustomModulesSubgroups(t *testing.T) {
	g := sccSubgroup("custom-modules")
	if g == nil {
		t.Fatal("scc custom-modules missing")
	}
	assertSubcommands(t, g, []string{"etd", "sha"})
}

func TestSccCustomModulesEtdSubcommands(t *testing.T) {
	g := sccNestedSubgroup("custom-modules", "etd")
	if g == nil {
		t.Fatal("scc custom-modules etd missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "describe-effective",
		"list", "list-descendant", "list-effective", "update",
	})
}

func TestSccCustomModulesShaSubcommands(t *testing.T) {
	g := sccNestedSubgroup("custom-modules", "sha")
	if g == nil {
		t.Fatal("scc custom-modules sha missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "describe-effective",
		"list", "list-descendant", "list-effective", "update",
	})
}

func TestSccFindingsSubcommands(t *testing.T) {
	g := sccSubgroup("findings")
	if g == nil {
		t.Fatal("scc findings missing")
	}
	assertSubcommands(t, g, []string{
		"bulk-mute", "create", "group", "list", "list-marks",
		"set-mute", "set-state", "update", "update-security-marks",
	})
}

func TestSccIaCValidationReportsSubcommands(t *testing.T) {
	g := sccSubgroup("iac-validation-reports")
	if g == nil {
		t.Fatal("scc iac-validation-reports missing")
	}
	assertSubcommands(t, g, []string{"create", "describe", "list"})
}

func TestSccManageSubgroups(t *testing.T) {
	g := sccSubgroup("manage")
	if g == nil {
		t.Fatal("scc manage missing")
	}
	assertSubcommands(t, g, []string{"settings"})
}

func TestSccManageSettingsSubcommands(t *testing.T) {
	g := sccNestedSubgroup("manage", "settings")
	if g == nil {
		t.Fatal("scc manage settings missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestSccMuteConfigsSubcommands(t *testing.T) {
	g := sccSubgroup("muteconfigs")
	if g == nil {
		t.Fatal("scc muteconfigs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestSccNotificationsSubcommands(t *testing.T) {
	g := sccSubgroup("notifications")
	if g == nil {
		t.Fatal("scc notifications missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestSccOperationsSubcommands(t *testing.T) {
	g := sccSubgroup("operations")
	if g == nil {
		t.Fatal("scc operations missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "list"})
}

func TestSccPostureDeploymentsSubcommands(t *testing.T) {
	g := sccSubgroup("posture-deployments")
	if g == nil {
		t.Fatal("scc posture-deployments missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestSccPostureOperationsSubcommands(t *testing.T) {
	g := sccSubgroup("posture-operations")
	if g == nil {
		t.Fatal("scc posture-operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}

func TestSccPostureTemplatesSubcommands(t *testing.T) {
	g := sccSubgroup("posture-templates")
	if g == nil {
		t.Fatal("scc posture-templates missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestSccPosturesSubcommands(t *testing.T) {
	g := sccSubgroup("postures")
	if g == nil {
		t.Fatal("scc postures missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "extract", "list", "list-revisions", "update",
	})
}

func TestSccSourcesSubcommands(t *testing.T) {
	g := sccSubgroup("sources")
	if g == nil {
		t.Fatal("scc sources missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}
