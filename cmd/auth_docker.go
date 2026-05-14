package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var authConfigureDockerCmd = &cobra.Command{
	Use:   "configure-docker [REGISTRIES]",
	Short: "Configure Docker to use gcloud as a credential helper",
	Long: `Configure Docker to authenticate with Artifact Registry or Container Registry.
Examples:
  gcloud auth configure-docker
  gcloud auth configure-docker europe-west1-docker.pkg.dev
  gcloud auth configure-docker us-docker.pkg.dev,europe-west1-docker.pkg.dev`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAuthConfigureDocker,
}

var flagIncludeArtifactRegistry bool

func init() {
	authConfigureDockerCmd.Flags().BoolVar(&flagIncludeArtifactRegistry, "include-artifact-registry", false, "Also add Artifact Registry domains")
	authCmd.AddCommand(authConfigureDockerCmd)
}

// dockerConfig represents ~/.docker/config.json
type dockerConfig struct {
	CredHelpers map[string]string `json:"credHelpers,omitempty"`
	Auths       map[string]any    `json:"auths,omitempty"`
	// Preserve other fields.
	Extra map[string]json.RawMessage `json:"-"`
}

func runAuthConfigureDocker(cmd *cobra.Command, args []string) error {
	var registries []string
	if len(args) > 0 && args[0] != "" {
		// Split comma-separated registries.
		for _, r := range splitComma(args[0]) {
			registries = append(registries, r)
		}
	} else {
		// Default Container Registry hosts.
		registries = []string{
			"gcr.io",
			"us.gcr.io",
			"eu.gcr.io",
			"asia.gcr.io",
			"staging-k8s.gcr.io",
			"marketplace.gcr.io",
		}
		if flagIncludeArtifactRegistry {
			arRegions := []string{
				"africa-south1", "asia-east1", "asia-east2", "asia-northeast1",
				"asia-northeast2", "asia-northeast3", "asia-south1", "asia-south2",
				"asia-southeast1", "asia-southeast2", "australia-southeast1",
				"australia-southeast2", "europe-central2", "europe-north1",
				"europe-southwest1", "europe-west1", "europe-west2", "europe-west3",
				"europe-west4", "europe-west6", "europe-west8", "europe-west9",
				"europe-west10", "europe-west12", "me-central1", "me-central2",
				"me-west1", "northamerica-northeast1", "northamerica-northeast2",
				"southamerica-east1", "southamerica-west1", "us-central1",
				"us-east1", "us-east4", "us-east5", "us-south1", "us-west1",
				"us-west2", "us-west3", "us-west4",
			}
			for _, r := range arRegions {
				registries = append(registries, r+"-docker.pkg.dev")
			}
		}
	}

	configPath, err := dockerConfigPath()
	if err != nil {
		return err
	}

	// Read existing config.
	cfg, raw, err := readDockerConfig(configPath)
	if err != nil {
		return err
	}

	if cfg.CredHelpers == nil {
		cfg.CredHelpers = make(map[string]string)
	}

	for _, reg := range registries {
		cfg.CredHelpers[reg] = "gcloud"
	}

	// Write back, preserving unknown fields.
	if err := writeDockerConfig(configPath, cfg, raw); err != nil {
		return err
	}

	fmt.Printf("Docker configuration file updated: %s\n", configPath)
	fmt.Printf("Added credential helper for: %v\n", registries)
	return nil
}

func dockerConfigPath() (string, error) {
	if p := os.Getenv("DOCKER_CONFIG"); p != "" {
		return filepath.Join(p, "config.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, ".docker", "config.json"), nil
}

func readDockerConfig(path string) (*dockerConfig, map[string]json.RawMessage, error) {
	cfg := &dockerConfig{}
	raw := make(map[string]json.RawMessage)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, raw, nil
		}
		return nil, nil, fmt.Errorf("reading docker config: %w", err)
	}

	// Parse into raw map to preserve unknown fields.
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("parsing docker config: %w", err)
	}

	// Extract known fields.
	if v, ok := raw["credHelpers"]; ok {
		json.Unmarshal(v, &cfg.CredHelpers)
	}
	if v, ok := raw["auths"]; ok {
		json.Unmarshal(v, &cfg.Auths)
	}

	return cfg, raw, nil
}

func writeDockerConfig(path string, cfg *dockerConfig, raw map[string]json.RawMessage) error {
	// Update known fields in the raw map.
	if cfg.CredHelpers != nil {
		data, _ := json.Marshal(cfg.CredHelpers)
		raw["credHelpers"] = data
	}

	// Ensure directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating docker config directory: %w", err)
	}

	data, err := json.MarshalIndent(raw, "", "\t")
	if err != nil {
		return fmt.Errorf("marshaling docker config: %w", err)
	}

	return os.WriteFile(path, append(data, '\n'), 0600)
}

func splitComma(s string) []string {
	var result []string
	for _, part := range filepath.SplitList(s) {
		result = append(result, part)
	}
	// filepath.SplitList uses OS path separator; use manual split for commas.
	result = nil
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			if part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	return result
}
