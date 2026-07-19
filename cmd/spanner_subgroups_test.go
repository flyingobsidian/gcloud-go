package cmd

import "testing"

func TestSpannerDatabasesSubcommands(t *testing.T) {
	g := spannerSubgroup("databases")
	if g == nil {
		t.Fatal("spanner databases missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "change-quorum", "create", "ddl", "delete",
		"describe", "execute-sql", "get-iam-policy", "list",
		"remove-iam-policy-binding", "restore", "roles", "sessions",
		"set-iam-policy", "splits", "update",
	})
}

func TestSpannerInstanceConfigsSubcommands(t *testing.T) {
	g := spannerSubgroup("instance-configs")
	if g == nil {
		t.Fatal("spanner instance-configs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestSpannerInstancePartitionsSubcommands(t *testing.T) {
	g := spannerSubgroup("instance-partitions")
	if g == nil {
		t.Fatal("spanner instance-partitions missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestSpannerInstancesSubcommands(t *testing.T) {
	g := spannerSubgroup("instances")
	if g == nil {
		t.Fatal("spanner instances missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "create", "delete", "describe",
		"get-iam-policy", "get-locations", "list", "move",
		"remove-iam-policy-binding", "set-iam-policy", "update",
	})
}

func TestSpannerOperationsSubcommands(t *testing.T) {
	g := spannerSubgroup("operations")
	if g == nil {
		t.Fatal("spanner operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list"})
}

func TestSpannerRowsSubcommands(t *testing.T) {
	g := spannerSubgroup("rows")
	if g == nil {
		t.Fatal("spanner rows missing")
	}
	assertSubcommands(t, g, []string{"delete", "insert", "update"})
}
