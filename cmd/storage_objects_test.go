package cmd

import "testing"

func TestStorageObjectsSubcommands(t *testing.T) {
	g := storageSubgroup("objects")
	if g == nil {
		t.Fatal("storage objects missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "update", "compose"})
}
