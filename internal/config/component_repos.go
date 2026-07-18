package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func componentReposPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "component_manager_repositories"), nil
}

// LoadAdditionalRepositories returns the persisted list of Trusted Tester
// component repository URLs, in the order they were added. Missing file → nil.
func LoadAdditionalRepositories() ([]string, error) {
	path, err := componentReposPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading component repositories: %w", err)
	}
	defer f.Close()

	var repos []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		repos = append(repos, line)
	}
	return repos, scanner.Err()
}

// SaveAdditionalRepositories persists the list of repository URLs.
func SaveAdditionalRepositories(repos []string) error {
	path, err := componentReposPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	var b strings.Builder
	for _, r := range repos {
		b.WriteString(r)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0600)
}
