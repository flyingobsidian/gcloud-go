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
	Core    CoreProperties
	Compute ComputeProperties
}

type CoreProperties struct {
	Account string
	Project string
}

type ComputeProperties struct {
	Zone   string
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

	// Determine which configuration to read.
	configName := os.Getenv("CLOUDSDK_ACTIVE_CONFIG_NAME")
	if configName == "" {
		// Check the active config file.
		activeFile := filepath.Join(dir, "active_config")
		data, err := os.ReadFile(activeFile)
		if err == nil {
			configName = strings.TrimSpace(string(data))
		}
	}
	if configName == "" {
		configName = "default"
	}

	path := filepath.Join(dir, "configurations", "config_"+configName)
	return loadINI(path)
}

// Save writes properties in gcloud's INI format.
func (p *Properties) Save() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	configsDir := filepath.Join(dir, "configurations")
	if err := os.MkdirAll(configsDir, 0700); err != nil {
		return fmt.Errorf("creating configurations directory: %w", err)
	}

	path := filepath.Join(configsDir, "config_default")
	return saveINI(path, p)
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
