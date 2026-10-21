package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfigGetAndGetValueRegistered ensures both the canonical `get` and
// the deprecated `get-value` alias are present, matching gcloud-python (#536).
func TestConfigGetAndGetValueRegistered(t *testing.T) {
	want := map[string]bool{"get": false, "get-value": false}
	for _, c := range configCmd.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("config subcommand %q not registered", name)
		}
	}
	// The deprecated alias must be flagged as such.
	for _, c := range configCmd.Commands() {
		if c.Name() == "get-value" && c.Deprecated == "" {
			t.Errorf("get-value should carry a deprecation notice")
		}
	}
}

// TestConfigSetPreservesOtherSections is the #2012 regression: `config set
// core/project` must not drop [workflows] location, and arbitrary
// section/key properties can be set and unset.
func TestConfigSetPreservesOtherSections(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "")
	path := filepath.Join(dir, "configurations", "config_default")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[workflows]\nlocation = europe-west2\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := runConfigSet(configSetCmd, []string{"core/project", "x"}); err != nil {
		t.Fatal(err)
	}
	if err := runConfigSet(configSetCmd, []string{"builds/region", "us-east1"}); err != nil {
		t.Fatal(err)
	}
	if err := runConfigUnset(configUnsetCmd, []string{"builds/region"}); err != nil {
		t.Fatal(err)
	}
	if err := runConfigSet(configSetCmd, []string{"core/a/b", "x"}); err == nil {
		t.Error("config set core/a/b succeeded, want error")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "[workflows]\nlocation = europe-west2\n\n[core]\nproject = x\n\n[builds]\n"
	if string(data) != want {
		t.Errorf("config file =\n%s\nwant\n%s", data, want)
	}
}
