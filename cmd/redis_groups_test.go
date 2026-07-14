package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func redisSubgroup(name string) *cobra.Command {
	for _, c := range redisCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestRedisAclPoliciesSubcommands(t *testing.T) {
	g := redisSubgroup("acl-policies")
	if g == nil {
		t.Fatal("redis acl-policies missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestRedisClustersSubcommands(t *testing.T) {
	g := redisSubgroup("clusters")
	if g == nil {
		t.Fatal("redis clusters missing")
	}
	assertSubcommands(t, g, []string{
		"add-token-auth-user", "backup-collections", "backups", "create",
		"create-backup", "delete", "describe", "get-cluster-certificate-authority",
		"get-shared-regional-certificate-authority", "list",
		"reschedule-maintenance", "update",
	})
	bc := findSub(g, "backup-collections")
	if bc == nil {
		t.Fatal("clusters backup-collections missing")
	}
	assertSubcommands(t, bc, []string{"describe", "list"})
	bk := findSub(g, "backups")
	if bk == nil {
		t.Fatal("clusters backups missing")
	}
	assertSubcommands(t, bk, []string{"delete", "describe", "export", "list"})
}

func TestRedisOperationsSubcommands(t *testing.T) {
	g := redisSubgroup("operations")
	if g == nil {
		t.Fatal("redis operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list"})
}

func TestRedisRegionsSubcommands(t *testing.T) {
	g := redisSubgroup("regions")
	if g == nil {
		t.Fatal("redis regions missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestRedisZonesSubcommands(t *testing.T) {
	g := redisSubgroup("zones")
	if g == nil {
		t.Fatal("redis zones missing")
	}
	assertSubcommands(t, g, []string{"list"})
}
