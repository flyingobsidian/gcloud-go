package cmd

import "testing"

// TestRootGlobalFlagsRegistered ensures every persistent flag mirrored from
// gcloud-python is registered on the root command (#530). Without this the
// CLI silently rejects invocations that pass any of these flags, which breaks
// parity with the Python reference.
func TestRootGlobalFlagsRegistered(t *testing.T) {
	want := []string{
		"access-token-file",
		"account",
		"billing-project",
		"configuration",
		"flags-file",
		"flatten",
		"format",
		"impersonate-service-account",
		"log-http",
		"project",
		"quiet",
		"trace-token",
		"user-output-enabled",
		"verbosity",
		"zone",
	}
	for _, name := range want {
		if rootCmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("persistent flag --%s not registered on rootCmd", name)
		}
	}
}
