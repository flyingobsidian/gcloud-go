package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func alloydbSubgroup(name string) *cobra.Command {
	for _, c := range alloydbCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestAlloydbBackupsSubcommands(t *testing.T) {
	g := alloydbSubgroup("backups")
	if g == nil {
		t.Fatal("alloydb backups missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestAlloydbClustersSubcommands(t *testing.T) {
	g := alloydbSubgroup("clusters")
	if g == nil {
		t.Fatal("alloydb clusters missing")
	}
	assertSubcommands(t, g, []string{
		"create", "create-secondary", "delete", "describe", "export", "import",
		"list", "migrate-cloud-sql", "promote", "restore", "switchover",
		"update", "upgrade",
	})
}

func TestAlloydbInstancesSubcommands(t *testing.T) {
	g := alloydbSubgroup("instances")
	if g == nil {
		t.Fatal("alloydb instances missing")
	}
	assertSubcommands(t, g, []string{
		"create", "create-secondary", "delete", "describe", "failover",
		"get-connection-info", "inject-fault", "list", "restart", "update",
	})
}

func TestAlloydbOperationsSubcommands(t *testing.T) {
	g := alloydbSubgroup("operations")
	if g == nil {
		t.Fatal("alloydb operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}

func TestAlloydbUsersSubcommands(t *testing.T) {
	g := alloydbSubgroup("users")
	if g == nil {
		t.Fatal("alloydb users missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
