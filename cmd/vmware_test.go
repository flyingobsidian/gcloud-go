package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func vmwareSubgroup(name string) *cobra.Command {
	for _, c := range vmwareCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestVmwareAnnouncementsSubcommands(t *testing.T) {
	g := vmwareSubgroup("announcements")
	if g == nil {
		t.Fatal("vmware announcements missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestVmwareDatastoresSubcommands(t *testing.T) {
	g := vmwareSubgroup("datastores")
	if g == nil {
		t.Fatal("vmware datastores missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestVmwareDnsBindPermissionSubcommands(t *testing.T) {
	g := vmwareSubgroup("dns-bind-permission")
	if g == nil {
		t.Fatal("vmware dns-bind-permission missing")
	}
	assertSubcommands(t, g, []string{"describe", "grant", "revoke"})
}

func TestVmwareLocationsSubcommands(t *testing.T) {
	g := vmwareSubgroup("locations")
	if g == nil {
		t.Fatal("vmware locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestVmwareNetworkPeeringsSubcommands(t *testing.T) {
	g := vmwareSubgroup("network-peerings")
	if g == nil {
		t.Fatal("vmware network-peerings missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestVmwareNetworkPoliciesSubcommands(t *testing.T) {
	g := vmwareSubgroup("network-policies")
	if g == nil {
		t.Fatal("vmware network-policies missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestVmwareNetworksSubcommands(t *testing.T) {
	g := vmwareSubgroup("networks")
	if g == nil {
		t.Fatal("vmware networks missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestVmwareNodeTypesSubcommands(t *testing.T) {
	g := vmwareSubgroup("node-types")
	if g == nil {
		t.Fatal("vmware node-types missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestVmwareOperationsSubcommands(t *testing.T) {
	g := vmwareSubgroup("operations")
	if g == nil {
		t.Fatal("vmware operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "delete"})
}

func TestVmwarePrivateCloudsSubcommands(t *testing.T) {
	g := vmwareSubgroup("private-clouds")
	if g == nil {
		t.Fatal("vmware private-clouds missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestVmwarePrivateConnectionsSubcommands(t *testing.T) {
	g := vmwareSubgroup("private-connections")
	if g == nil {
		t.Fatal("vmware private-connections missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
