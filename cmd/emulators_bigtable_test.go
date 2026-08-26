package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// findEmuSubgroup returns the named subgroup under `emulators` (or nil).
func findEmuSubgroup(name string) *cobra.Command {
	for _, c := range emulatorsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// findLeaf returns the named leaf command under parent (or nil).
func findLeaf(parent *cobra.Command, name string) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// TestEmulatorsBigtableRegistered guards #1709: `emulators bigtable` must have
// real `start` and `env-init` leaves (not stubs), and `start` must accept
// --host-port.
func TestEmulatorsBigtableRegistered(t *testing.T) {
	g := findEmuSubgroup("bigtable")
	if g == nil {
		t.Fatal("emulators bigtable subgroup missing")
	}
	start := findLeaf(g, "start")
	if start == nil {
		t.Fatal("emulators bigtable start missing")
	}
	if start.RunE == nil {
		t.Error("emulators bigtable start should have a real RunE (not a stub)")
	}
	if start.Flags().Lookup("host-port") == nil {
		t.Error("emulators bigtable start should expose --host-port")
	}
	envInit := findLeaf(g, "env-init")
	if envInit == nil {
		t.Fatal("emulators bigtable env-init missing")
	}
	if envInit.RunE == nil {
		t.Error("emulators bigtable env-init should have a real RunE (not a stub)")
	}
}

func TestEmuBigtableEnvInitOutput(t *testing.T) {
	prev := flagEmuBigtableHostPort
	defer func() { flagEmuBigtableHostPort = prev }()
	flagEmuBigtableHostPort = "127.0.0.1:9000"

	out := captureStdout(t, func() {
		if err := runEmuBigtableEnvInit(nil, nil); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	if !strings.Contains(out, "export BIGTABLE_EMULATOR_HOST=127.0.0.1:9000") {
		t.Errorf("env-init output missing expected export: %q", out)
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"plain":           "plain",
		"":                "''",
		"has space":       "'has space'",
		"has'quote":       `'has'\''quote'`,
		"127.0.0.1:8086":  "127.0.0.1:8086",
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHostPortSplit(t *testing.T) {
	if got := hostFrom("localhost:8086"); got != "localhost" {
		t.Errorf("hostFrom = %q", got)
	}
	if got := portFrom("localhost:8086"); got != "8086" {
		t.Errorf("portFrom = %q", got)
	}
	if got := hostFrom("bare"); got != "bare" {
		t.Errorf("hostFrom bare = %q", got)
	}
	if got := portFrom("bare"); got != "" {
		t.Errorf("portFrom bare = %q", got)
	}
}
