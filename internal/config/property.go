package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return readINIValue(filepath.Join(dir, "configurations", "config_"+ActiveConfigName()), section, key)
}

// propertyEnvVar returns the environment variable gcloud reads for
// section/key, e.g. CLOUDSDK_WORKFLOWS_LOCATION for workflows/location.
func propertyEnvVar(section, key string) string {
	return "CLOUDSDK_" + strings.ToUpper(section) + "_" + strings.ToUpper(key)
}

// readINIValue returns key from section of the INI file at path, or "" when
// the file, section or key is absent.
func readINIValue(path, section, key string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading config: %w", err)
	}
	defer f.Close()

	current := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = line[1 : len(line)-1]
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok && current == section && strings.TrimSpace(k) == key {
			return strings.TrimSpace(v), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading config: %w", err)
	}
	return "", nil
}
