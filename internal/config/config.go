package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Properties holds gcloud configuration properties.
type Properties struct {
	Core      CoreProperties
	Compute   ComputeProperties
	Dataflow  RegionProperty
	Run       RegionProperty
	Redis     RegionProperty
	Functions RegionProperty
}

type CoreProperties struct {
	Account string
	Project string
}

type ComputeProperties struct {
	Zone   string
	Region string
}

// RegionProperty holds a region for service-specific sections.
type RegionProperty struct {
	Region string
}

// ConfigDir returns the gcloud configuration directory.
func ConfigDir() (string, error) {
	if d := os.Getenv("CLOUDSDK_CONFIG"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, ".config", "gcloud"), nil
}

// Load reads the active gcloud configuration (INI format).
// It reads configurations/config_default (or the active named configuration).
func Load() (*Properties, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	configName := ActiveConfigName()
	path := filepath.Join(dir, "configurations", "config_"+configName)
	return loadINI(path)
}

// LoadNamed reads a specific named configuration.
func LoadNamed(name string) (*Properties, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "configurations", "config_"+name)
	return loadINI(path)
}

// Save writes properties in gcloud's INI format to the active configuration.
func (p *Properties) Save() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	configsDir := filepath.Join(dir, "configurations")
	if err := os.MkdirAll(configsDir, 0700); err != nil {
		return fmt.Errorf("creating configurations directory: %w", err)
	}

	configName := ActiveConfigName()
	path := filepath.Join(configsDir, "config_"+configName)
	return saveINI(path, p)
}

// ActiveConfigName returns the name of the active configuration.
func ActiveConfigName() string {
	if name := os.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME"); name != "" {
		return name
	}
	dir, err := ConfigDir()
	if err != nil {
		return "default"
	}
	data, err := os.ReadFile(filepath.Join(dir, "active_config"))
	if err == nil {
		if name := strings.TrimSpace(string(data)); name != "" {
			return name
		}
	}
	return "default"
}

// CreateConfiguration creates a new named configuration.
func CreateConfiguration(name string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	configsDir := filepath.Join(dir, "configurations")
	if err := os.MkdirAll(configsDir, 0700); err != nil {
		return fmt.Errorf("creating configurations directory: %w", err)
	}

	path := filepath.Join(configsDir, "config_"+name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("configuration [%s] already exists", name)
	}

	return saveINI(path, &Properties{})
}

// ActivateConfiguration sets the active configuration by name.
func ActivateConfiguration(name string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "configurations", "config_"+name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("configuration [%s] does not exist", name)
	}

	return os.WriteFile(filepath.Join(dir, "active_config"), []byte(name+"\n"), 0600)
}

// ListConfigurations returns the names of all configurations.
func ListConfigurations() ([]string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	configsDir := filepath.Join(dir, "configurations")
	entries, err := os.ReadDir(configsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"default"}, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "config_") {
			names = append(names, strings.TrimPrefix(e.Name(), "config_"))
		}
	}
	if len(names) == 0 {
		names = []string{"default"}
	}
	return names, nil
}

// DeleteConfiguration removes a named configuration.
func DeleteConfiguration(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete the default configuration")
	}
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "configurations", "config_"+name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("configuration [%s] does not exist", name)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("deleting configuration: %w", err)
	}

	// If this was the active config, reset to default.
	if ActiveConfigName() == name {
		_ = os.WriteFile(filepath.Join(dir, "active_config"), []byte("default\n"), 0600)
	}

	return nil
}

// loadINI parses a gcloud INI configuration file.
func loadINI(path string) (*Properties, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Properties{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	defer f.Close()

	p := &Properties{}
	section := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch section {
		case "core":
			switch key {
			case "account":
				p.Core.Account = val
			case "project":
				p.Core.Project = val
			}
		case "compute":
			switch key {
			case "zone":
				p.Compute.Zone = val
			case "region":
				p.Compute.Region = val
			}
		case "dataflow":
			if key == "region" {
				p.Dataflow.Region = val
			}
		case "run":
			if key == "region" {
				p.Run.Region = val
			}
		case "redis":
			if key == "region" {
				p.Redis.Region = val
			}
		case "functions":
			if key == "region" {
				p.Functions.Region = val
			}
		}
	}
	return p, scanner.Err()
}

// saveINI writes properties in gcloud INI format.
func saveINI(path string, p *Properties) error {
	var b strings.Builder
	b.WriteString("[core]\n")
	if p.Core.Account != "" {
		fmt.Fprintf(&b, "account = %s\n", p.Core.Account)
	}
	if p.Core.Project != "" {
		fmt.Fprintf(&b, "project = %s\n", p.Core.Project)
	}

	if p.Compute.Zone != "" || p.Compute.Region != "" {
		b.WriteString("\n[compute]\n")
		if p.Compute.Zone != "" {
			fmt.Fprintf(&b, "zone = %s\n", p.Compute.Zone)
		}
		if p.Compute.Region != "" {
			fmt.Fprintf(&b, "region = %s\n", p.Compute.Region)
		}
	}

	for _, s := range []struct {
		name   string
		region string
	}{
		{"dataflow", p.Dataflow.Region},
		{"run", p.Run.Region},
		{"redis", p.Redis.Region},
		{"functions", p.Functions.Region},
	} {
		if s.region != "" {
			fmt.Fprintf(&b, "\n[%s]\nregion = %s\n", s.name, s.region)
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0600)
}

// Resolve returns the effective value for a setting, preferring flag > env > config.
func Resolve(flag, envVar, configVal string) string {
	if flag != "" {
		return flag
	}
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	return configVal
}
