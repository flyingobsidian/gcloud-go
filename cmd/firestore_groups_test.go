package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func firestoreSubgroup(name string) *cobra.Command {
	for _, c := range firestoreCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestFirestoreDatabasesSubcommands(t *testing.T) {
	g := firestoreSubgroup("databases")
	if g == nil {
		t.Fatal("firestore databases missing")
	}
	assertSubcommands(t, g, []string{"clone", "connection-string", "create", "delete", "describe", "list", "restore", "update"})
}

func TestFirestoreBackupsSubcommands(t *testing.T) {
	g := firestoreSubgroup("backups")
	if g == nil {
		t.Fatal("firestore backups missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "list", "schedules"})
	sched := findSub(g, "schedules")
	if sched == nil {
		t.Fatal("backups schedules missing")
	}
	assertSubcommands(t, sched, []string{"create", "delete", "describe", "list", "update"})
}

func TestFirestoreFieldsSubcommands(t *testing.T) {
	g := firestoreSubgroup("fields")
	if g == nil {
		t.Fatal("firestore fields missing")
	}
	assertSubcommands(t, g, []string{"ttls"})
	ttls := findSub(g, "ttls")
	if ttls == nil {
		t.Fatal("fields ttls missing")
	}
	assertSubcommands(t, ttls, []string{"list", "update"})
}

func TestFirestoreIndexesSubcommands(t *testing.T) {
	g := firestoreSubgroup("indexes")
	if g == nil {
		t.Fatal("firestore indexes missing")
	}
	assertSubcommands(t, g, []string{"composite", "fields"})
	comp := findSub(g, "composite")
	if comp == nil {
		t.Fatal("indexes composite missing")
	}
	assertSubcommands(t, comp, []string{"create", "delete", "describe", "list"})
	fields := findSub(g, "fields")
	if fields == nil {
		t.Fatal("indexes fields missing")
	}
	assertSubcommands(t, fields, []string{"describe", "list", "update"})
}

func TestFirestoreLocationsSubcommands(t *testing.T) {
	g := firestoreSubgroup("locations")
	if g == nil {
		t.Fatal("firestore locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestFirestoreOperationsSubcommands(t *testing.T) {
	g := firestoreSubgroup("operations")
	if g == nil {
		t.Fatal("firestore operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}

func TestFirestoreUserCredsSubcommands(t *testing.T) {
	g := firestoreSubgroup("user-creds")
	if g == nil {
		t.Fatal("firestore user-creds missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "disable", "enable", "list", "reset-password"})
}
