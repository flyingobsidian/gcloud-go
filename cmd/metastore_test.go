package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func metastoreSubgroup(name string) *cobra.Command {
	for _, c := range metastoreCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestMetastoreFederationsSubcommands(t *testing.T) {
	g := metastoreSubgroup("federations")
	if g == nil {
		t.Fatal("metastore federations missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "create", "delete", "describe",
		"get-iam-policy", "list", "remove-iam-policy-binding", "set-iam-policy", "update",
	})
}

func TestMetastoreLocationsSubcommands(t *testing.T) {
	g := metastoreSubgroup("locations")
	if g == nil {
		t.Fatal("metastore locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestMetastoreOperationsSubcommands(t *testing.T) {
	g := metastoreSubgroup("operations")
	if g == nil {
		t.Fatal("metastore operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list", "wait"})
}

func TestMetastoreServicesSubcommands(t *testing.T) {
	g := metastoreSubgroup("services")
	if g == nil {
		t.Fatal("metastore services missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "alter-metadata-resource-location", "alter-table-properties",
		"backups", "create", "delete", "describe", "export", "get-iam-policy", "import",
		"list", "move-table-to-database", "query-metadata", "remove-iam-policy-binding",
		"restore", "set-iam-policy", "update",
	})
}

func TestMetastoreServicesBackupsSubcommands(t *testing.T) {
	svcs := metastoreSubgroup("services")
	if svcs == nil {
		t.Fatal("metastore services missing")
	}
	bk := findSub(svcs, "backups")
	if bk == nil {
		t.Fatal("metastore services backups missing")
	}
	assertSubcommands(t, bk, []string{"create", "delete", "describe", "list"})
}
