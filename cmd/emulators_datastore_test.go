package cmd

import (
	"strings"
	"testing"
)

func TestEmulatorsDatastoreRegistered(t *testing.T) {
	g := findEmuSubgroup("datastore")
	if g == nil {
		t.Fatal("emulators datastore subgroup missing")
	}
	for _, name := range []string{"start", "env-init", "env-unset"} {
		leaf := findLeaf(g, name)
		if leaf == nil {
			t.Fatalf("emulators datastore %s missing", name)
		}
		if leaf.RunE == nil {
			t.Errorf("emulators datastore %s should have a real RunE (not a stub)", name)
		}
	}
	start := findLeaf(g, "start")
	for _, f := range []string{"host-port", "data-dir", "store-on-disk", "consistency", "use-firestore-in-datastore-mode"} {
		if start.Flags().Lookup(f) == nil {
			t.Errorf("emulators datastore start missing --%s", f)
		}
	}
}

func TestEmuDatastoreEnvInitOutput(t *testing.T) {
	prevHost := flagEmuDatastoreHostPort
	defer func() { flagEmuDatastoreHostPort = prevHost }()
	flagEmuDatastoreHostPort = "127.0.0.1:9099"

	out := captureStdout(t, func() {
		if err := runEmuDatastoreEnvInit(nil, nil); err != nil {
			t.Fatalf("run: %v", err)
		}
	})

	for _, want := range []string{
		"export DATASTORE_EMULATOR_HOST=127.0.0.1:9099",
		"export DATASTORE_EMULATOR_HOST_PATH=127.0.0.1:9099/datastore",
		"export DATASTORE_HOST=http://127.0.0.1:9099",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("env-init output missing %q; got:\n%s", want, out)
		}
	}
}

func TestEmuDatastoreEnvUnsetOutput(t *testing.T) {
	out := captureStdout(t, func() {
		if err := runEmuDatastoreEnvUnset(nil, nil); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	for _, want := range []string{
		"unset DATASTORE_EMULATOR_HOST",
		"unset DATASTORE_EMULATOR_HOST_PATH",
		"unset DATASTORE_HOST",
		"unset DATASTORE_PROJECT_ID",
		"unset DATASTORE_DATASET",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("env-unset missing %q; got:\n%s", want, out)
		}
	}
}
