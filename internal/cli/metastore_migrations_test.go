package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	metastore "google.golang.org/api/metastore/v1"
)

func TestMetastoreServicesMigrationsSubcommands(t *testing.T) {
	svcs := metastoreSubgroup("services")
	if svcs == nil {
		t.Fatal("metastore services missing")
	}
	mig := findSub(svcs, "migrations")
	if mig == nil {
		t.Fatal("metastore services migrations missing")
	}
	assertSubcommands(t, mig, []string{"delete", "describe", "list", "start"})
}

// msMigFakeAPI records requests and serves metastore v1 migration responses.
type msMigFakeAPI struct {
	requests []string
	bodies   []string
}

func (f *msMigFakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.requests = append(f.requests, r.Method+" "+r.URL.Path)
	f.bodies = append(f.bodies, string(body))
	w.Header().Set("Content-Type", "application/json")
	p := r.URL.Path
	const svc = "/v1/projects/p/locations/us-central1/services/s1"
	switch {
	case r.Method == http.MethodPost && p == svc+":startMigration":
		fmt.Fprint(w, `{"name":"projects/p/locations/us-central1/operations/op1","done":true,"response":{"name":"projects/p/locations/us-central1/services/s1/migrationExecutions/m1","state":"RUNNING"}}`)
	case r.Method == http.MethodGet && p == svc+"/migrationExecutions":
		if r.URL.Query().Get("pageToken") == "" {
			fmt.Fprint(w, `{"migrationExecutions":[{"name":"projects/p/locations/us-central1/services/s1/migrationExecutions/m1","state":"SUCCEEDED","createTime":"2026-09-10T08:00:00.123456Z","endTime":"2026-09-10T09:30:00Z","stateMessage":"Done"}],"nextPageToken":"n"}`)
			return
		}
		fmt.Fprint(w, `{"migrationExecutions":[{"name":"projects/p/locations/us-central1/services/s1/migrationExecutions/migration-two","state":"RUNNING","createTime":"2026-09-11T10:00:00Z"}]}`)
	case r.Method == http.MethodGet && p == svc+"/migrationExecutions/m1":
		fmt.Fprint(w, `{"name":"projects/p/locations/us-central1/services/s1/migrationExecutions/m1","state":"SUCCEEDED"}`)
	case r.Method == http.MethodDelete && p == svc+"/migrationExecutions/m1":
		fmt.Fprint(w, `{"name":"projects/p/locations/us-central1/operations/op2","done":true}`)
	default:
		http.Error(w, r.Method+" "+p, http.StatusNotFound)
	}
}

func msMigStartBody(t *testing.T, args ...string) metastore.StartMigrationRequest {
	t.Helper()
	api := &msMigFakeAPI{}
	inv, _, _ := newFakeAPIInvocation(t, api)
	base := []string{"metastore", "services", "migrations", "start", "s1", "--location", "us-central1", "--project", "p"}
	if err := runInvocation(inv, append(base, args...)...); err != nil {
		t.Fatal(err)
	}
	if len(api.bodies) != 1 {
		t.Fatalf("requests = %v, want one startMigration", api.requests)
	}
	var req metastore.StartMigrationRequest
	if err := json.Unmarshal([]byte(api.bodies[0]), &req); err != nil {
		t.Fatalf("body %s: %v", api.bodies[0], err)
	}
	return req
}

func TestMetastoreMigrationsStartDefaults(t *testing.T) {
	req := msMigStartBody(t, "--hive-catalog", "projects/p/catalogs/hive")
	cfg := req.MigrationExecution.BiglakeMetastoreMigrationConfig
	if cfg.Mode != "BACKFILL" || cfg.ConflictPolicy != "" || cfg.DryRun || cfg.ReportPath != "" {
		t.Errorf("config = %+v, want mode BACKFILL and nothing else", cfg)
	}
	if cfg.HiveConfig == nil || cfg.HiveConfig.Catalog != "projects/p/catalogs/hive" ||
		strings.Join(cfg.HiveConfig.Databases, ",") != "*" {
		t.Errorf("hiveConfig = %+v, want the catalog with databases [*]", cfg.HiveConfig)
	}
	if cfg.IcebergConfig != nil {
		t.Errorf("icebergConfig = %+v, want unset", cfg.IcebergConfig)
	}
	if req.MigrationExecutionId != "" {
		t.Errorf("migrationExecutionId = %q, want unset", req.MigrationExecutionId)
	}
}

func TestMetastoreMigrationsStartAllFlags(t *testing.T) {
	req := msMigStartBody(t, "--migration-execution-id", "m1", "--mode", "backfill", "--conflict-policy", "overwrite",
		"--dry-run", "--report-path", "gs://b/reports",
		"--hive-catalog", "projects/p/catalogs/hive", "--hive-databases", "db1,db2",
		"--iceberg-catalog", "projects/p/catalogs/ice", "--iceberg-namespaces", "ns1")
	cfg := req.MigrationExecution.BiglakeMetastoreMigrationConfig
	if req.MigrationExecutionId != "m1" || cfg.Mode != "BACKFILL" || cfg.ConflictPolicy != "OVERWRITE" ||
		!cfg.DryRun || cfg.ReportPath != "gs://b/reports" {
		t.Errorf("request = %+v, config = %+v", req, cfg)
	}
	if strings.Join(cfg.HiveConfig.Databases, ",") != "db1,db2" {
		t.Errorf("hive databases = %v", cfg.HiveConfig.Databases)
	}
	if cfg.IcebergConfig == nil || cfg.IcebergConfig.Catalog != "projects/p/catalogs/ice" ||
		strings.Join(cfg.IcebergConfig.Namespaces, ",") != "ns1" {
		t.Errorf("icebergConfig = %+v", cfg.IcebergConfig)
	}
}

func TestMetastoreMigrationsStartOutput(t *testing.T) {
	api := &msMigFakeAPI{}
	inv, stdout, stderr := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "metastore", "services", "migrations", "start", "s1", "--location", "us-central1",
		"--project", "p", "--iceberg-catalog", "projects/p/catalogs/ice")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "migrationExecutions/m1") {
		t.Errorf("stdout = %q, want the migration execution", stdout)
	}
	if want := "Start migration [s1] completed.\n"; stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}
}

// The validation errors match gcloud-python 589.0.0's.
func TestMetastoreMigrationsStartValidation(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"--mode", "incremental-sync"}, "argument --mode: Invalid choice: 'incremental-sync'.\n\nValid choices are [backfill]."},
		{[]string{"--conflict-policy", "foo"}, "argument --conflict-policy: Invalid choice: 'foo'.\n\nValid choices are [overwrite, skip]."},
		{[]string{"--report-path", "x"}, "Invalid value for [--report-path]: Expected URI x to start with `gs://`."},
		{[]string{"--hive-databases", "a"}, "argument --hive-databases: --hive-catalog must be specified."},
		{[]string{"--iceberg-namespaces", "a"}, "argument --iceberg-namespaces: --iceberg-catalog must be specified."},
		{[]string{"--hive-catalog", "bad"}, "Invalid value for [--hive-catalog]: The BigLake catalog resource name should be in the format projects/<project_id>/catalogs/<catalog_id>"},
		{[]string{"--iceberg-catalog", "projects/p/catalogs/c", "--iceberg-namespaces", "*,a"}, `Invalid value for [--iceberg-namespaces]: When "*" is specified to include all items, it must be the only element in the list.`},
		{[]string{"--hive-catalog", "projects/p/catalogs/c", "--hive-databases", "a, "}, "Invalid value for [--hive-databases]: Item in list cannot be empty or whitespace."},
		{[]string{"--hive-catalog", "projects/p/catalogs/c", "--hive-databases", ""}, "Invalid value for [--hive-databases]: The list cannot be empty."},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			api := &msMigFakeAPI{}
			inv, _, _ := newFakeAPIInvocation(t, api)
			base := []string{"metastore", "services", "migrations", "start", "s1", "--location", "us-central1", "--project", "p"}
			err := runInvocation(inv, append(base, tt.args...)...)
			if err == nil || err.Error() != tt.want {
				t.Errorf("err = %v, want %q", err, tt.want)
			}
			if len(api.requests) != 0 {
				t.Errorf("requests = %v, want none", api.requests)
			}
		})
	}
}

func TestMetastoreMigrationsList(t *testing.T) {
	api := &msMigFakeAPI{}
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "metastore", "services", "migrations", "list", "--service", "s1", "--location", "us-central1", "--project", "p")
	if err != nil {
		t.Fatal(err)
	}
	want := "NAME           STATE      CREATE_TIME          END_TIME             MESSAGE\n" +
		"m1             SUCCEEDED  2026-09-10T08:00:00  2026-09-10T09:30:00  Done\n" +
		"migration-two  RUNNING    2026-09-11T10:00:00\n"
	if stdout.String() != want {
		t.Errorf("stdout =\n%s\nwant\n%s", stdout, want)
	}
}

func TestMetastoreMigrationsDescribeAndDelete(t *testing.T) {
	api := &msMigFakeAPI{}
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "metastore", "services", "migrations", "describe", "m1", "--service", "s1",
		"--location", "us-central1", "--project", "p", "--format", "value(state)"); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "SUCCEEDED\n" {
		t.Errorf("describe stdout = %q", stdout)
	}
	inv, _, stderr := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "metastore", "services", "migrations", "delete", "m1", "--service", "s1",
		"--location", "us-central1", "--project", "p"); err != nil {
		t.Fatal(err)
	}
	if want := "Delete migration [m1] completed.\n"; stderr.String() != want {
		t.Errorf("delete stderr = %q, want %q", stderr, want)
	}
	if last := api.requests[len(api.requests)-1]; last != "DELETE /v1/projects/p/locations/us-central1/services/s1/migrationExecutions/m1" {
		t.Errorf("last request = %q", last)
	}
}
