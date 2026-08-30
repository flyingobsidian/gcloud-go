package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	alloydb "google.golang.org/api/alloydb/v1"
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

// TestAlloydbInstanceLabelFlagsMerge exercises the #1751 label helpers:
// --labels on create merges with the config-file body, and
// --update-labels / --remove-labels / --clear-labels apply the expected
// mutations while producing the `labels` field-mask entry.
func TestAlloydbInstanceLabelFlagsMerge(t *testing.T) {
	t.Cleanup(func() {
		flagADBLabels = nil
		flagADBUpdateLabels = nil
		flagADBRemoveLabels = nil
		flagADBClearLabels = false
	})

	// Create: config-file body has one label; --labels adds a second and
	// overrides the first.
	flagADBLabels = map[string]string{"team": "cloud", "env": "prod"}
	inst := &alloydb.Instance{Labels: map[string]string{"team": "legacy"}}
	adbApplyCreateLabels(inst)
	if inst.Labels["team"] != "cloud" || inst.Labels["env"] != "prod" {
		t.Fatalf("create labels merge: got %v", inst.Labels)
	}

	// Update: clear, then re-add, then remove.
	flagADBLabels = nil
	flagADBClearLabels = true
	flagADBUpdateLabels = map[string]string{"team": "cloud", "tier": "web"}
	flagADBRemoveLabels = []string{"tier"}
	inst2 := &alloydb.Instance{Labels: map[string]string{"old": "value"}}
	mask := adbApplyUpdateLabels(inst2)
	if mask != "labels" {
		t.Fatalf("update mask entry: got %q, want %q", mask, "labels")
	}
	if _, ok := inst2.Labels["old"]; ok {
		t.Errorf("clear-labels did not drop pre-existing label")
	}
	if inst2.Labels["team"] != "cloud" {
		t.Errorf("update-labels did not add team=cloud, got %v", inst2.Labels)
	}
	if _, ok := inst2.Labels["tier"]; ok {
		t.Errorf("remove-labels did not drop tier, got %v", inst2.Labels)
	}
}

// TestAlloydbClusterRestoreBackupDR exercises the #1751 BackupDR wiring on
// clusters restore.
func TestAlloydbClusterRestoreBackupDR(t *testing.T) {
	t.Cleanup(func() {
		flagADBBackupDRBackup = ""
		flagADBBackupDRDataSource = ""
	})

	flagADBBackupDRBackup = "projects/p/locations/us/backupVaults/v/dataSources/d/backups/b"
	flagADBBackupDRDataSource = ""
	req := &alloydb.RestoreClusterRequest{}
	if err := adbApplyClusterRestoreBackupDR(req); err != nil {
		t.Fatalf("apply BackupDR backup: %v", err)
	}
	if req.BackupdrBackupSource == nil || req.BackupdrBackupSource.Backup != flagADBBackupDRBackup {
		t.Fatalf("BackupdrBackupSource: got %+v", req.BackupdrBackupSource)
	}

	flagADBBackupDRBackup = ""
	flagADBBackupDRDataSource = "projects/p/locations/us/backupVaults/v/dataSources/d"
	req = &alloydb.RestoreClusterRequest{}
	if err := adbApplyClusterRestoreBackupDR(req); err != nil {
		t.Fatalf("apply BackupDR data-source: %v", err)
	}
	if req.BackupdrPitrSource == nil || req.BackupdrPitrSource.DataSource != flagADBBackupDRDataSource {
		t.Fatalf("BackupdrPitrSource: got %+v", req.BackupdrPitrSource)
	}

	flagADBBackupDRBackup = "projects/p/.../backups/b"
	flagADBBackupDRDataSource = "projects/p/.../dataSources/d"
	err := adbApplyClusterRestoreBackupDR(&alloydb.RestoreClusterRequest{})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("mutually-exclusive flags: got err %v, want mutually-exclusive", err)
	}
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
