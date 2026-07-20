package cmd

import "testing"

func TestStorageOperationsSubcommands(t *testing.T) {
	g := storageSubgroup("operations")
	if g == nil {
		t.Fatal("storage operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list"})
}

func TestStorageServiceAgentSubcommands(t *testing.T) {
	g := storageSubgroup("service-agent")
	if g == nil {
		t.Fatal("storage service-agent missing")
	}
	assertSubcommands(t, g, []string{"describe"})
}
