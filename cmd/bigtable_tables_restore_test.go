package cmd

import (
	"reflect"
	"testing"

	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// setBTRestoreFlags sets restore flags for one test case and returns a
// function that resets them (values and Changed state).
func setBTRestoreFlags(t *testing.T, vals map[string]string) func() {
	t.Helper()
	fs := bigtableTablesRestoreCmd.Flags()
	for k, v := range vals {
		if err := fs.Set(k, v); err != nil {
			t.Fatalf("set %s: %v", k, err)
		}
	}
	return func() {
		for k := range vals {
			f := fs.Lookup(k)
			f.Changed = false
		}
		flagBTTblRestoreDeletionProtection = false
		flagBTTblRestoreNoDeletionProtection = false
		flagBTTblRestoreEnableAutoBackup = false
		flagBTTblRestoreAutoBackupRetention = ""
		flagBTTblRestoreAutoBackupLocations = nil
	}
}

func TestApplyBTRestoreTableProperties(t *testing.T) {
	cases := []struct {
		name      string
		flags     map[string]string
		wantTable *bigtableadmin.Table
		wantMask  string
	}{
		{"none", nil, nil, ""},
		{
			"deletion protection",
			map[string]string{"deletion-protection": "true"},
			&bigtableadmin.Table{DeletionProtection: true},
			"deletion_protection",
		},
		{
			"no deletion protection",
			map[string]string{"no-deletion-protection": "true"},
			&bigtableadmin.Table{ForceSendFields: []string{"DeletionProtection"}},
			"deletion_protection",
		},
		{
			"enable automated backup",
			map[string]string{"enable-automated-backup": "true"},
			&bigtableadmin.Table{AutomatedBackupPolicy: &bigtableadmin.AutomatedBackupPolicy{
				RetentionPeriod: "604800s", Frequency: "86400s",
			}},
			"automated_backup_policy",
		},
		{
			"retention and locations",
			map[string]string{
				"automated-backup-retention-period": "30d",
				"automated-backup-locations":        "projects/p/locations/us-east1-b",
				"deletion-protection":               "true",
			},
			&bigtableadmin.Table{
				DeletionProtection: true,
				AutomatedBackupPolicy: &bigtableadmin.AutomatedBackupPolicy{
					RetentionPeriod: "2592000s",
					Locations:       []string{"projects/p/locations/us-east1-b"},
				},
			},
			"deletion_protection,automated_backup_policy",
		},
		{
			"locations only defaults retention to 7d",
			map[string]string{"automated-backup-locations": "projects/p/locations/z"},
			&bigtableadmin.Table{AutomatedBackupPolicy: &bigtableadmin.AutomatedBackupPolicy{
				RetentionPeriod: "604800s",
				Locations:       []string{"projects/p/locations/z"},
			}},
			"automated_backup_policy",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer setBTRestoreFlags(t, c.flags)()
			req := &bigtableadmin.RestoreTableRequest{}
			if err := applyBTRestoreTableProperties(bigtableTablesRestoreCmd, req); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(req.Table, c.wantTable) {
				t.Errorf("table = %+v, want %+v", req.Table, c.wantTable)
			}
			if req.UpdateMask != c.wantMask {
				t.Errorf("mask = %q, want %q", req.UpdateMask, c.wantMask)
			}
		})
	}
}

func TestBTRestoreBadRetention(t *testing.T) {
	defer setBTRestoreFlags(t, map[string]string{"automated-backup-retention-period": "soon"})()
	if err := applyBTRestoreTableProperties(bigtableTablesRestoreCmd, &bigtableadmin.RestoreTableRequest{}); err == nil {
		t.Error("expected error for malformed retention period")
	}
}

func TestFormatDurationForJSON(t *testing.T) {
	for in, want := range map[string]string{"7d": "604800s", "1d": "86400s", "90": "90s", "1.5s": "1.5s"} {
		got, err := formatDurationForJSON(in, "--x")
		if err != nil || got != want {
			t.Errorf("%s: got %q, %v; want %q", in, got, err, want)
		}
	}
}
