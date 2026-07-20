package cmd

import "testing"

func TestSpannerBackupSchedulesSubcommands(t *testing.T) {
	g := spannerSubgroup("backup-schedules")
	if g == nil {
		t.Fatal("spanner backup-schedules missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy",
	})
}
