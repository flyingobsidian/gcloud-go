package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeActiveConfig points CLOUDSDK_CONFIG at a temp dir holding body as the
// default configuration and returns the file's path.
func writeActiveConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "")
	configsDir := filepath.Join(dir, "configurations")
	if err := os.MkdirAll(configsDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configsDir, "config_default")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestSavePreservesUnknownSections is the #2012 regression: `config set
// core/project x` (Load, change, Save) must keep [workflows] location.
func TestSavePreservesUnknownSections(t *testing.T) {
	path := writeActiveConfig(t, "[core]\nproject = a\n\n[workflows]\nlocation = europe-west2\n")

	props, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	props.Core.Project = "x"
	if err := props.Save(); err != nil {
		t.Fatal(err)
	}

	want := "[core]\nproject = x\n\n[workflows]\nlocation = europe-west2\n"
	if got := readFile(t, path); got != want {
		t.Errorf("config file =\n%s\nwant\n%s", got, want)
	}
	if got, err := Property("workflows", "location"); err != nil || got != "europe-west2" {
		t.Errorf("Property(workflows, location) = %q, %v; want europe-west2", got, err)
	}
}

func TestSetPropertyPreservesUnknownSections(t *testing.T) {
	path := writeActiveConfig(t, "[workflows]\nlocation = europe-west2\n")

	if err := SetProperty("core", "project", "x"); err != nil {
		t.Fatal(err)
	}

	want := "[workflows]\nlocation = europe-west2\n\n[core]\nproject = x\n"
	if got := readFile(t, path); got != want {
		t.Errorf("config file =\n%s\nwant\n%s", got, want)
	}
}

func TestSaveUnsetsClearedFields(t *testing.T) {
	path := writeActiveConfig(t, "# keep me\n[core]\naccount = a@example.com\nproject = p\n\n[builds]\nregion = us-east1\n")

	props, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	props.Core.Account = ""
	props.Compute.Zone = "europe-west2-a"
	if err := props.Save(); err != nil {
		t.Fatal(err)
	}

	want := "# keep me\n[core]\nproject = p\n\n[builds]\nregion = us-east1\n\n[compute]\nzone = europe-west2-a\n"
	if got := readFile(t, path); got != want {
		t.Errorf("config file =\n%s\nwant\n%s", got, want)
	}
}

func TestSetUnsetArbitraryProperty(t *testing.T) {
	path := writeActiveConfig(t, "[core]\nproject = p\n\n[composer]\nlocation = us-central1\n")

	if err := SetProperty("composer", "location", "europe-west1"); err != nil {
		t.Fatal(err)
	}
	if err := SetProperty("core", "account", "a@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := SetProperty("workflows", "location", "europe-west2"); err != nil {
		t.Fatal(err)
	}
	want := "[core]\nproject = p\naccount = a@example.com\n\n[composer]\nlocation = europe-west1\n\n[workflows]\nlocation = europe-west2\n"
	if got := readFile(t, path); got != want {
		t.Errorf("after set, config file =\n%s\nwant\n%s", got, want)
	}

	if err := UnsetProperty("composer", "location"); err != nil {
		t.Fatal(err)
	}
	if err := UnsetProperty("builds", "region"); err != nil {
		t.Fatal(err)
	}
	// Python gcloud keeps the now-empty section header.
	want = "[core]\nproject = p\naccount = a@example.com\n\n[composer]\n\n[workflows]\nlocation = europe-west2\n"
	if got := readFile(t, path); got != want {
		t.Errorf("after unset, config file =\n%s\nwant\n%s", got, want)
	}

	got, err := FileProperties()
	if err != nil {
		t.Fatal(err)
	}
	if got["workflows"]["location"] != "europe-west2" || got["core"]["account"] != "a@example.com" || len(got["composer"]) != 0 {
		t.Errorf("FileProperties() = %v", got)
	}
}

func TestSetPropertyCreatesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "")

	if err := SetProperty("workflows", "location", "europe-west2"); err != nil {
		t.Fatal(err)
	}
	want := "[workflows]\nlocation = europe-west2\n"
	if got := readFile(t, filepath.Join(dir, "configurations", "config_default")); got != want {
		t.Errorf("config file = %q, want %q", got, want)
	}
}

func TestSetPropertyReplacesDuplicateKeys(t *testing.T) {
	path := writeActiveConfig(t, "[core]\nproject = a\nproject = b\n")

	if err := SetProperty("core", "project", "c"); err != nil {
		t.Fatal(err)
	}
	if got, want := readFile(t, path), "[core]\nproject = c\n"; got != want {
		t.Errorf("config file = %q, want %q", got, want)
	}
}

func TestValidatePropertyName(t *testing.T) {
	for _, tt := range []struct {
		section, key string
		ok           bool
	}{
		{"core", "project", true},
		{"billing", "quota_project", true},
		{"api_endpoint_overrides", "cloudresourcemanager", true},
		{"core", "", false},
		{"", "project", false},
		{"core", "a/b", false},
		{"core", "Project", false},
		{"co re", "project", false},
		{"core]", "project", false},
		{"core", "project = x", false},
	} {
		err := ValidatePropertyName(tt.section, tt.key)
		if (err == nil) != tt.ok {
			t.Errorf("ValidatePropertyName(%q, %q) = %v, want ok=%v", tt.section, tt.key, err, tt.ok)
		}
	}
	if err := SetProperty("core", "a/b", "x"); err == nil {
		t.Error("SetProperty(core, a/b) succeeded, want error")
	}
}

func TestCreateConfigurationWritesEmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)

	if err := CreateConfiguration("other"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(dir, "configurations", "config_other")); got != "" {
		t.Errorf("new configuration = %q, want empty", got)
	}
}
