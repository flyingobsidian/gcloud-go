package cmd

import (
	"bytes"
	"testing"
)

func TestVersionCommandRegistered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("version command not registered on rootCmd")
	}
}

func TestVersionRuns(t *testing.T) {
	// Redirect the command's own writer; runVersion prints via fmt directly,
	// so this is a smoke test that RunE returns without error and doesn't
	// try to hit the network.
	var buf bytes.Buffer
	versionCmd.SetOut(&buf)
	versionCmd.SetErr(&buf)
	if err := runVersion(versionCmd, nil); err != nil {
		t.Fatalf("runVersion returned error: %v", err)
	}
}
