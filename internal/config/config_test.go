package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadINI(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)

	// Create gcloud-style INI config.
	configsDir := filepath.Join(dir, "configurations")
	if err := os.MkdirAll(configsDir, 0700); err != nil {
		t.Fatal(err)
	}
	ini := `[core]
account = user@example.com
project = my-project

[compute]
zone = us-central1-a
region = us-central1
`
	if err := os.WriteFile(filepath.Join(configsDir, "config_default"), []byte(ini), 0600); err != nil {
		t.Fatal(err)
	}

	props, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if props.Core.Account != "user@example.com" {
		t.Errorf("account = %q, want %q", props.Core.Account, "user@example.com")
	}
	if props.Core.Project != "my-project" {
		t.Errorf("project = %q, want %q", props.Core.Project, "my-project")
	}
	if props.Compute.Zone != "us-central1-a" {
		t.Errorf("zone = %q, want %q", props.Compute.Zone, "us-central1-a")
	}
	if props.Compute.Region != "us-central1" {
		t.Errorf("region = %q, want %q", props.Compute.Region, "us-central1")
	}
}

func TestLoadINIWithComments(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)

	configsDir := filepath.Join(dir, "configurations")
	os.MkdirAll(configsDir, 0700)
	ini := `# gcloud config
[core]
; this is the account
account = test@test.com
# project
project = proj
`
	os.WriteFile(filepath.Join(configsDir, "config_default"), []byte(ini), 0600)

	props, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if props.Core.Account != "test@test.com" {
		t.Errorf("account = %q", props.Core.Account)
	}
	if props.Core.Project != "proj" {
		t.Errorf("project = %q", props.Core.Project)
	}
}

func TestLoadNamedConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)

	configsDir := filepath.Join(dir, "configurations")
	os.MkdirAll(configsDir, 0700)

	// Write active_config pointing to "staging".
	os.WriteFile(filepath.Join(dir, "active_config"), []byte("staging\n"), 0600)

	// Write config_staging.
	ini := `[core]
account = staging@example.com
project = staging-proj
`
	os.WriteFile(filepath.Join(configsDir, "config_staging"), []byte(ini), 0600)

	props, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if props.Core.Account != "staging@example.com" {
		t.Errorf("account = %q, want staging@example.com", props.Core.Account)
	}
	if props.Core.Project != "staging-proj" {
		t.Errorf("project = %q, want staging-proj", props.Core.Project)
	}
}

func TestLoadNamedConfigFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "prod")

	configsDir := filepath.Join(dir, "configurations")
	os.MkdirAll(configsDir, 0700)

	ini := `[core]
account = prod@example.com
`
	os.WriteFile(filepath.Join(configsDir, "config_prod"), []byte(ini), 0600)

	props, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if props.Core.Account != "prod@example.com" {
		t.Errorf("account = %q", props.Core.Account)
	}
}

func TestLoadNonexistent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", filepath.Join(dir, "nonexistent"))

	props, err := Load()
	if err != nil {
		t.Fatalf("Load() error for missing file: %v", err)
	}
	if props.Core.Account != "" {
		t.Error("expected empty properties for missing file")
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)

	props := &Properties{
		Core: CoreProperties{
			Account: "test@test.iam.gserviceaccount.com",
			Project: "my-project",
		},
		Compute: ComputeProperties{
			Zone: "us-central1-a",
		},
	}
	if err := props.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	if loaded.Core.Account != "test@test.iam.gserviceaccount.com" {
		t.Errorf("account = %q", loaded.Core.Account)
	}
	if loaded.Core.Project != "my-project" {
		t.Errorf("project = %q", loaded.Core.Project)
	}
	if loaded.Compute.Zone != "us-central1-a" {
		t.Errorf("zone = %q", loaded.Compute.Zone)
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		flag      string
		envKey    string
		envVal    string
		configVal string
		want      string
	}{
		{"flag wins", "flag-val", "", "", "config-val", "flag-val"},
		{"env wins over config", "", "TEST_ENV_VAR", "env-val", "config-val", "env-val"},
		{"config fallback", "", "TEST_ENV_UNSET", "", "config-val", "config-val"},
		{"all empty", "", "TEST_ENV_UNSET", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv(tt.envKey, tt.envVal)
			}
			got := Resolve(tt.flag, tt.envKey, tt.configVal)
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfigDir(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		t.Setenv("CLOUDSDK_CONFIG", "/custom/path")
		dir, err := ConfigDir()
		if err != nil {
			t.Fatal(err)
		}
		if dir != "/custom/path" {
			t.Errorf("ConfigDir() = %q, want /custom/path", dir)
		}
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("CLOUDSDK_CONFIG", "")
		os.Unsetenv("CLOUDSDK_CONFIG")
		dir, err := ConfigDir()
		if err != nil {
			t.Fatal(err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", "gcloud")
		if dir != expected {
			t.Errorf("ConfigDir() = %q, want %q", dir, expected)
		}
	})
}

func TestDeleteActiveConfigurationResetsToDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)

	if err := CreateConfiguration("temp"); err != nil {
		t.Fatalf("CreateConfiguration: %v", err)
	}
	if err := ActivateConfiguration("temp"); err != nil {
		t.Fatalf("ActivateConfiguration: %v", err)
	}
	if got := ActiveConfigName(); got != "temp" {
		t.Fatalf("active config = %q, want temp", got)
	}

	if err := DeleteConfiguration("temp"); err != nil {
		t.Fatalf("DeleteConfiguration returned error: %v", err)
	}

	if got := ActiveConfigName(); got != "default" {
		t.Errorf("active config after delete = %q, want default", got)
	}
}
