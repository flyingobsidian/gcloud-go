package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func netmgmtSubgroup(name string) *cobra.Command {
	for _, c := range networkManagementCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestNetworkManagementHasSubgroups(t *testing.T) {
	for _, want := range []string{
		"connectivity-tests", "network-monitoring-providers",
		"operations", "vpc-flow-logs-configs",
	} {
		if netmgmtSubgroup(want) == nil {
			t.Errorf("network-management missing subgroup %q", want)
		}
	}
}

func TestNetmgmtConnectivityTestsSubcommands(t *testing.T) {
	g := netmgmtSubgroup("connectivity-tests")
	if g == nil {
		t.Fatal("connectivity-tests missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "rerun", "update"})
}

func TestNetmgmtOperationsSubcommands(t *testing.T) {
	g := netmgmtSubgroup("operations")
	if g == nil {
		t.Fatal("operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list", "wait"})
}

func TestNetmgmtVpcFlowLogsConfigsSubcommands(t *testing.T) {
	g := netmgmtSubgroup("vpc-flow-logs-configs")
	if g == nil {
		t.Fatal("vpc-flow-logs-configs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestNMQualify(t *testing.T) {
	flagNMLocation = "global"
	defer func() { flagNMLocation = "" }()
	got := nmQualify("test1", "my-proj", "connectivityTests")
	want := "projects/my-proj/locations/global/connectivityTests/test1"
	if got != want {
		t.Errorf("nmQualify = %q, want %q", got, want)
	}
	pass := "projects/x/locations/global/connectivityTests/full"
	if nmQualify(pass, "ignored", "connectivityTests") != pass {
		t.Errorf("nmQualify should pass through fully-qualified names")
	}
}
