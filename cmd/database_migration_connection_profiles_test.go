package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	datamigration "google.golang.org/api/datamigration/v1"
)

func TestDMCPResourceName(t *testing.T) {
	cases := []struct{ in, project, region, want string }{
		{"my-profile", "my-project", "us-central1",
			"projects/my-project/locations/us-central1/connectionProfiles/my-profile"},
		{"projects/other/locations/eu/connectionProfiles/x", "my-project", "us-central1",
			"projects/other/locations/eu/connectionProfiles/x"},
	}
	for _, c := range cases {
		if got := dmCPResourceName(c.in, c.project, c.region); got != c.want {
			t.Errorf("dmCPResourceName(%q,%q,%q) = %q, want %q", c.in, c.project, c.region, got, c.want)
		}
	}
}

func TestDMParent(t *testing.T) {
	if got := dmParent("p", "us"); got != "projects/p/locations/us" {
		t.Errorf("got %q", got)
	}
}

func TestConnectionProfileType(t *testing.T) {
	cases := []struct {
		name    string
		profile *datamigration.ConnectionProfile
		want    string
	}{
		{"alloydb", &datamigration.ConnectionProfile{Alloydb: &datamigration.AlloyDbConnectionProfile{}}, "ALLOYDB"},
		{"cloudsql", &datamigration.ConnectionProfile{Cloudsql: &datamigration.CloudSqlConnectionProfile{}}, "CLOUDSQL"},
		{"mysql", &datamigration.ConnectionProfile{Mysql: &datamigration.MySqlConnectionProfile{}}, "MYSQL"},
		{"oracle", &datamigration.ConnectionProfile{Oracle: &datamigration.OracleConnectionProfile{}}, "ORACLE"},
		{"postgresql", &datamigration.ConnectionProfile{Postgresql: &datamigration.PostgreSqlConnectionProfile{}}, "POSTGRESQL"},
		{"sqlserver", &datamigration.ConnectionProfile{Sqlserver: &datamigration.SqlServerConnectionProfile{}}, "SQLSERVER"},
		{"empty", &datamigration.ConnectionProfile{}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := connectionProfileType(c.profile); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestDeriveConnectionProfileUpdateMask(t *testing.T) {
	p := &datamigration.ConnectionProfile{
		DisplayName: "My Profile",
		Mysql:       &datamigration.MySqlConnectionProfile{},
		Labels:      map[string]string{"env": "dev"},
	}
	got := deriveConnectionProfileUpdateMask(p)
	for _, want := range []string{"displayName", "labels", "mysql"} {
		if !strings.Contains(got, want) {
			t.Errorf("update mask %q missing %q", got, want)
		}
	}
	// Empty profile should still return a non-empty default.
	if got := deriveConnectionProfileUpdateMask(&datamigration.ConnectionProfile{}); got == "" {
		t.Error("empty profile produced empty mask")
	}
}

func TestLoadConnectionProfileFile(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "profile.json")
		content := `{"displayName":"Prod MySQL","mysql":{"host":"10.0.0.5","port":3306}}`
		if err := os.WriteFile(p, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		prof, err := loadConnectionProfileFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if prof.DisplayName != "Prod MySQL" {
			t.Errorf("displayName = %q", prof.DisplayName)
		}
		if prof.Mysql == nil || prof.Mysql.Host != "10.0.0.5" || prof.Mysql.Port != 3306 {
			t.Errorf("mysql = %+v", prof.Mysql)
		}
	})

	t.Run("YAML", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "profile.yaml")
		content := "displayName: Prod PG\npostgresql:\n  host: 10.0.0.6\n  port: 5432\n"
		if err := os.WriteFile(p, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		prof, err := loadConnectionProfileFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if prof.DisplayName != "Prod PG" {
			t.Errorf("displayName = %q", prof.DisplayName)
		}
		if prof.Postgresql == nil || prof.Postgresql.Port != 5432 {
			t.Errorf("postgresql = %+v", prof.Postgresql)
		}
	})
}

func TestDMConnectionProfilesCommandTree(t *testing.T) {
	want := []string{"create", "delete", "describe", "fetch-static-ips", "list", "test", "update"}
	got := map[string]bool{}
	for _, c := range dmConnProfilesCmd.Commands() {
		got[c.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("connection-profiles subcommand %q not registered", name)
		}
	}
	// Ensure it's wired under database-migration.
	found := false
	for _, c := range databaseMigrationCmd.Commands() {
		if c == dmConnProfilesCmd {
			found = true
			break
		}
	}
	if !found {
		t.Error("connection-profiles not attached to database-migration")
	}
}
