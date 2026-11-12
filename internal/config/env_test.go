package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func writeConfig(t *testing.T, dir, name, ini string) {
	t.Helper()
	configsDir := filepath.Join(dir, "configurations")
	if err := os.MkdirAll(configsDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configsDir, "config_"+name), []byte(ini), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestMapEnvIgnoresProcessEnvironment(t *testing.T) {
	t.Setenv("CLOUDSDK_CONFIG", t.TempDir())
	t.Setenv("CLOUDSDK_CORE_PROJECT", "from-process")

	dir := t.TempDir()
	writeConfig(t, dir, "default", "[core]\nproject = from-file\n")
	env := MapEnv(map[string]string{"CLOUDSDK_CONFIG": dir}, "/nonexistent")

	props, err := env.Load()
	if err != nil {
		t.Fatal(err)
	}
	if props.Core.Project != "from-file" {
		t.Errorf("project = %q, want %q", props.Core.Project, "from-file")
	}
	got, err := env.Property("core", "project")
	if err != nil {
		t.Fatal(err)
	}
	if got != "from-file" {
		t.Errorf("Property(core, project) = %q, want %q", got, "from-file")
	}
	if got := env.Resolve("", "CLOUDSDK_CORE_PROJECT", "cfg"); got != "cfg" {
		t.Errorf("Resolve = %q, want %q", got, "cfg")
	}
}

func TestMapEnvConfigDirFromHome(t *testing.T) {
	home := t.TempDir()
	dir, err := MapEnv(nil, home).ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".config", "gcloud"); dir != want {
		t.Errorf("ConfigDir = %q, want %q", dir, want)
	}
}

func TestEnvsAreIndependentAndConcurrent(t *testing.T) {
	envs := map[string]Env{}
	for _, project := range []string{"alpha-project", "beta-project"} {
		dir := t.TempDir()
		writeConfig(t, dir, "prod", "[core]\nproject = "+project+"\n")
		envs[project] = MapEnv(map[string]string{
			"CLOUDSDK_CONFIG":             dir,
			"CLOUDSDK_ACTIVE_CONFIG_NAME": "prod",
		}, "/nonexistent")
	}

	var wg sync.WaitGroup
	for want, env := range envs {
		for range 20 {
			wg.Go(func() {
				props, err := env.Load()
				if err != nil {
					t.Error(err)
					return
				}
				if props.Core.Project != want {
					t.Errorf("project = %q, want %q", props.Core.Project, want)
				}
			})
		}
	}
	wg.Wait()
}
