package cmd

import "testing"

// TestConfigGetAndGetValueRegistered ensures both the canonical `get` and
// the deprecated `get-value` alias are present, matching gcloud-python (#536).
func TestConfigGetAndGetValueRegistered(t *testing.T) {
	want := map[string]bool{"get": false, "get-value": false}
	for _, c := range configCmd.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("config subcommand %q not registered", name)
		}
	}
	// The deprecated alias must be flagged as such.
	for _, c := range configCmd.Commands() {
		if c.Name() == "get-value" && c.Deprecated == "" {
			t.Errorf("get-value should carry a deprecation notice")
		}
	}
}
