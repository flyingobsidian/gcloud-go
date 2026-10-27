package cmd

import (
	sqladmin "google.golang.org/api/sqladmin/v1"
	"testing"

	"github.com/spf13/cobra"
)

func sqlSubgroup(name string) *cobra.Command {
	for _, c := range sqlCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestSQLBackupsSubcommands(t *testing.T) {
	g := sqlSubgroup("backups")
	if g == nil {
		t.Fatal("sql backups missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "restore"})
}

func TestSQLConnectSubcommands(t *testing.T) {
	g := sqlSubgroup("connect")
	if g == nil {
		t.Fatal("sql connect missing")
	}
	assertSubcommands(t, g, []string{"psql", "mysql", "sqlserver"})
}

func TestSQLDatabasesSubcommands(t *testing.T) {
	g := sqlSubgroup("databases")
	if g == nil {
		t.Fatal("sql databases missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "patch"})
}

func TestSQLExportSubcommands(t *testing.T) {
	g := sqlSubgroup("export")
	if g == nil {
		t.Fatal("sql export missing")
	}
	assertSubcommands(t, g, []string{"sql", "csv", "bak"})
}

func TestSQLFlagsSubcommands(t *testing.T) {
	g := sqlSubgroup("flags")
	if g == nil {
		t.Fatal("sql flags missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestSQLImportSubcommands(t *testing.T) {
	g := sqlSubgroup("import")
	if g == nil {
		t.Fatal("sql import missing")
	}
	assertSubcommands(t, g, []string{"sql", "csv", "bak"})
}

func TestSQLInstancesSubcommands(t *testing.T) {
	g := sqlSubgroup("instances")
	if g == nil {
		t.Fatal("sql instances missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "patch",
		"clone", "failover", "restart", "restore-backup",
		"promote-replica", "start-replica", "stop-replica",
		"reencrypt", "reset-ssl-config", "import", "export",
		"list-server-cas", "reschedule-maintenance",
		"acquire-ssrs-lease", "release-ssrs-lease",
		"reset-async-replica-lag", "verify-external-sync-settings",
	})
}

func TestSQLOperationsSubcommands(t *testing.T) {
	g := sqlSubgroup("operations")
	if g == nil {
		t.Fatal("sql operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list", "wait"})
}

func TestSQLSSLServerCACertsSubcommands(t *testing.T) {
	ssl := sqlSubgroup("ssl")
	if ssl == nil {
		t.Fatal("sql ssl missing")
	}
	sca := findSub(ssl, "server-ca-certs")
	if sca == nil {
		t.Fatal("sql ssl server-ca-certs missing")
	}
	assertSubcommands(t, sca, []string{"list", "add", "remove", "rotate"})
}

func TestSQLSSLCertsSubcommands(t *testing.T) {
	g := sqlSubgroup("ssl-certs")
	if g == nil {
		t.Fatal("sql ssl-certs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestSQLTiersSubcommands(t *testing.T) {
	g := sqlSubgroup("tiers")
	if g == nil {
		t.Fatal("sql tiers missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestSQLUsersSubcommands(t *testing.T) {
	g := sqlSubgroup("users")
	if g == nil {
		t.Fatal("sql users missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "set-password", "patch"})
}

func TestSQLInstanceRoutingParams(t *testing.T) {
	for _, tc := range []struct {
		inst *sqladmin.DatabaseInstance
		want string
	}{
		{&sqladmin.DatabaseInstance{}, ""},
		{&sqladmin.DatabaseInstance{Region: "us-central1"}, "region=us-central1"},
		{&sqladmin.DatabaseInstance{Region: "us-central1", Settings: &sqladmin.Settings{Edition: "ENTERPRISE_PLUS"}},
			"edition=ENTERPRISE_PLUS&region=us-central1"},
	} {
		if got := sqlInstanceRoutingParams(tc.inst); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}
