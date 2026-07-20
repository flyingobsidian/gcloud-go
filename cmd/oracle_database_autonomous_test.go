package cmd

import "testing"

func TestOracleAutonomousDatabaseBackupsSubcommands(t *testing.T) {
	g := oracleSubgroup("autonomous-database-backups")
	if g == nil {
		t.Fatal("oracle-database autonomous-database-backups missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleAutonomousDatabasesSubcommands(t *testing.T) {
	g := oracleSubgroup("autonomous-databases")
	if g == nil {
		t.Fatal("oracle-database autonomous-databases missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"failover", "generate-wallet", "restart", "restore",
		"start", "stop", "switchover",
	})
}
