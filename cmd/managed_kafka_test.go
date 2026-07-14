package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func mkSubgroup(name string) *cobra.Command {
	for _, c := range managedKafkaCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestMKAclsSubcommands(t *testing.T) {
	g := mkSubgroup("acls")
	if g == nil {
		t.Fatal("managed-kafka acls missing")
	}
	assertSubcommands(t, g, []string{"add-acl-entry", "create", "delete", "describe", "list", "remove-acl-entry", "update"})
}

func TestMKClustersSubcommands(t *testing.T) {
	g := mkSubgroup("clusters")
	if g == nil {
		t.Fatal("managed-kafka clusters missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestMKConnectClustersSubcommands(t *testing.T) {
	g := mkSubgroup("connect-clusters")
	if g == nil {
		t.Fatal("managed-kafka connect-clusters missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestMKConnectorsSubcommands(t *testing.T) {
	g := mkSubgroup("connectors")
	if g == nil {
		t.Fatal("managed-kafka connectors missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "pause", "restart", "resume", "stop", "update"})
}

func TestMKConsumerGroupsSubcommands(t *testing.T) {
	g := mkSubgroup("consumer-groups")
	if g == nil {
		t.Fatal("managed-kafka consumer-groups missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "list", "update"})
}

func TestMKOperationsSubcommands(t *testing.T) {
	g := mkSubgroup("operations")
	if g == nil {
		t.Fatal("managed-kafka operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestMKTopicsSubcommands(t *testing.T) {
	g := mkSubgroup("topics")
	if g == nil {
		t.Fatal("managed-kafka topics missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
