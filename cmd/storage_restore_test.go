package cmd

import "testing"

func TestStorageHasRestore(t *testing.T) {
	found := false
	for _, c := range newStorageCmd(newOSInvocation()).Commands() {
		if c.Name() == "restore" {
			found = true
			if c.RunE == nil {
				t.Fatal("storage restore should have a RunE (not just a stub)")
			}
		}
	}
	if !found {
		t.Fatal("storage restore missing")
	}
}
