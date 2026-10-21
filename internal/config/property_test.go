package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProperty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "default")
	if err := os.MkdirAll(filepath.Join(dir, "configurations"), 0o700); err != nil {
		t.Fatal(err)
	}
	ini := "[core]\nproject = p\n\n[workflows]\n# comment\nlocation = europe-west2\n"
	if err := os.WriteFile(filepath.Join(dir, "configurations", "config_default"), []byte(ini), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		section string
		key     string
		env     string
		want    string
	}{
		{"from config", "workflows", "location", "", "europe-west2"},
		{"env wins over config", "workflows", "location", "us-east1", "us-east1"},
		{"unset key", "workflows", "other", "", ""},
		{"unset section", "scheduler", "location", "", ""},
		{"same key in other section", "core", "location", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(propertyEnvVar(tt.section, tt.key), tt.env)
			got, err := Property(tt.section, tt.key)
			if err != nil {
				t.Fatalf("Property() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Property(%q, %q) = %q, want %q", tt.section, tt.key, got, tt.want)
			}
		})
	}
}

func TestPropertyMissingConfig(t *testing.T) {
	t.Setenv("CLOUDSDK_CONFIG", t.TempDir())
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "default")
	got, err := Property("workflows", "location")
	if err != nil || got != "" {
		t.Errorf("Property() = %q, %v; want \"\", nil", got, err)
	}
}
