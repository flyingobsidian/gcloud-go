package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadYAMLOrJSONInto reads a file at path and decodes it as JSON or YAML into
// dst. YAML is decoded via a generic-map round-trip through JSON so that Google
// API Go types (which only carry JSON tags) unmarshal cleanly.
func loadYAMLOrJSONInto(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	var generic any
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}
	jsonBytes, err := json.Marshal(convertYAMLKeys(generic))
	if err != nil {
		return fmt.Errorf("normalising config file: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, dst); err != nil {
		return fmt.Errorf("decoding config file: %w", err)
	}
	return nil
}

// nonEmptyJSONFields returns the top-level JSON field names present on v.
// It is used to synthesize a default update-mask when the caller does not
// pass --update-mask. Fields tagged with "-" are ignored.
func nonEmptyJSONFields(v any) []string {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	fields := make([]string, 0, len(m))
	for k := range m {
		fields = append(fields, k)
	}
	return fields
}

// joinMask returns fields joined by ",", or "" if empty.
func joinMask(fields []string) string {
	return strings.Join(fields, ",")
}

// saveAsYAML writes v to path as YAML (round-tripped through JSON so that
// Google API Go types, which only carry JSON tags, marshal with the correct
// field names).
func saveAsYAML(path string, v any) error {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshalling resource: %w", err)
	}
	var generic any
	if err := json.Unmarshal(jsonBytes, &generic); err != nil {
		return fmt.Errorf("normalising resource: %w", err)
	}
	out, err := yaml.Marshal(generic)
	if err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	if err := os.WriteFile(path, out, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
