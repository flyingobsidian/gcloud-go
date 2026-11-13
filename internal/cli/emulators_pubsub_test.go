package cli

import (
	"strings"
	"testing"
)

func TestEmulatorsPubsubRegistered(t *testing.T) {
	g := findEmuSubgroup("pubsub")
	if g == nil {
		t.Fatal("emulators pubsub subgroup missing")
	}
	for _, name := range []string{"start", "env-init"} {
		leaf := findLeaf(g, name)
		if leaf == nil {
			t.Fatalf("emulators pubsub %s missing", name)
		}
		if leaf.RunE == nil {
			t.Errorf("emulators pubsub %s should have a real RunE (not a stub)", name)
		}
	}
	start := findLeaf(g, "start")
	for _, f := range []string{"host-port", "data-dir"} {
		if start.Flags().Lookup(f) == nil {
			t.Errorf("emulators pubsub start missing --%s", f)
		}
	}
}

func TestEmuPubsubEnvInitOutput(t *testing.T) {
	grp := newEmulatorsCmds(legacyTestInvocation())
	prev := grp.flagEmuPubsubHostPort
	defer func() { grp.flagEmuPubsubHostPort = prev }()
	grp.flagEmuPubsubHostPort = "127.0.0.1:8085"

	out := captureStdout(t, func() {
		if err := grp.runEmuPubsubEnvInit(nil, nil); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	if !strings.Contains(out, "export PUBSUB_EMULATOR_HOST=127.0.0.1:8085") {
		t.Errorf("env-init missing PUBSUB_EMULATOR_HOST export; got:\n%s", out)
	}
}
