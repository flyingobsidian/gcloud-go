package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Properties holds gcloud configuration properties.
type Properties struct {
	Core      CoreProperties
	Compute   ComputeProperties
	Billing   BillingProperties
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

// BillingProperties mirrors gcloud-python's [billing] config section. Only
// quota_project is used today; it drives the x-goog-user-project header on
// REST calls so per-project quota is applied when the caller's credentials
// aren't tied to a billing project.
type BillingProperties struct {
	QuotaProject string
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
// Sections and keys that Properties doesn't model are preserved.
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

	// Python gcloud creates new configurations as empty files.
	return os.WriteFile(path, nil, 0600)
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
		if err := os.WriteFile(filepath.Join(dir, "active_config"), []byte("default\n"), 0600); err != nil {
			return fmt.Errorf("resetting active config to default: %w", err)
		}
	}

	return nil
}

// loadINI parses a gcloud INI configuration file.
func loadINI(path string) (*Properties, error) {
	f, err := readINIFile(path)
	if err != nil {
		return nil, err
	}
	p := &Properties{}
	for _, field := range p.fields() {
		*field.val = f.get(field.section, field.key)
	}
	return p, nil
}

// saveINI writes the typed properties to the INI file at path, updating only
// their keys so that sections and keys Properties doesn't model survive.
func saveINI(path string, p *Properties) error {
	f, err := readINIFile(path)
	if err != nil {
		return err
	}
	for _, field := range p.fields() {
		if *field.val == "" {
			f.unset(field.section, field.key)
		} else {
			f.set(field.section, field.key, *field.val)
		}
	}
	return f.write(path)
}

// propertyField ties a Properties field to its section/key in the file.
type propertyField struct {
	section, key string
	val          *string
}

// fields lists the properties modelled by Properties.
func (p *Properties) fields() []propertyField {
	return []propertyField{
		{"core", "account", &p.Core.Account},
		{"core", "project", &p.Core.Project},
		{"compute", "zone", &p.Compute.Zone},
		{"compute", "region", &p.Compute.Region},
		{"billing", "quota_project", &p.Billing.QuotaProject},
		{"dataflow", "region", &p.Dataflow.Region},
		{"run", "region", &p.Run.Region},
		{"redis", "region", &p.Redis.Region},
		{"functions", "region", &p.Functions.Region},
	}
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
