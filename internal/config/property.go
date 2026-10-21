package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Property returns the value of section/key from the CLOUDSDK_<SECTION>_<KEY>
// environment variable, falling back to the active configuration file. It
// reaches properties that Properties doesn't model, such as
// workflows/location, and returns "" when the property is unset.
func Property(section, key string) (string, error) {
	if v := os.Getenv(propertyEnvVar(section, key)); v != "" {
		return v, nil
	}
	path, err := activeConfigPath()
	if err != nil {
		return "", err
	}
	f, err := readINIFile(path)
	if err != nil {
		return "", err
	}
	return f.get(section, key), nil
}

// SetProperty sets section/key in the active configuration file, leaving
// every other property in the file untouched.
func SetProperty(section, key, value string) error {
	return editActiveConfig(section, key, func(f *iniFile) { f.set(section, key, value) })
}

// UnsetProperty removes section/key from the active configuration file,
// leaving every other property in the file untouched.
func UnsetProperty(section, key string) error {
	return editActiveConfig(section, key, func(f *iniFile) { f.unset(section, key) })
}

// FileProperties returns every property set in the active configuration
// file, keyed by section and then key.
func FileProperties() (map[string]map[string]string, error) {
	path, err := activeConfigPath()
	if err != nil {
		return nil, err
	}
	f, err := readINIFile(path)
	if err != nil {
		return nil, err
	}
	props := map[string]map[string]string{}
	f.each(func(section, key, val string) {
		if section == "" {
			return
		}
		if props[section] == nil {
			props[section] = map[string]string{}
		}
		props[section][key] = val
	})
	return props, nil
}

// propertyNameRE matches gcloud section and property names, e.g. core,
// quota_project, api_endpoint_overrides.
var propertyNameRE = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// ValidatePropertyName reports whether section/key is a well-formed property
// name. Unlike Python gcloud it doesn't check the name against a schema of
// known properties, so properties gcloud-go doesn't model can still be set.
func ValidatePropertyName(section, key string) error {
	if !propertyNameRE.MatchString(section) {
		return fmt.Errorf("invalid section name [%s]", section)
	}
	if !propertyNameRE.MatchString(key) {
		return fmt.Errorf("invalid property name [%s] in section [%s]", key, section)
	}
	return nil
}

// editActiveConfig validates section/key, applies edit to the active
// configuration file and writes it back.
func editActiveConfig(section, key string, edit func(*iniFile)) error {
	if err := ValidatePropertyName(section, key); err != nil {
		return err
	}
	path, err := activeConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating configurations directory: %w", err)
	}
	f, err := readINIFile(path)
	if err != nil {
		return err
	}
	edit(f)
	return f.write(path)
}

// activeConfigPath returns the path of the active configuration file.
func activeConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "configurations", "config_"+ActiveConfigName()), nil
}

// propertyEnvVar returns the environment variable gcloud reads for
// section/key, e.g. CLOUDSDK_WORKFLOWS_LOCATION for workflows/location.
func propertyEnvVar(section, key string) string {
	return "CLOUDSDK_" + strings.ToUpper(section) + "_" + strings.ToUpper(key)
}
