package cmd

import "testing"

// Ensures `monitoring snoozes create` accepts --format (both `--format x` and
// `--format=x` are handled by cobra once the flag is registered).
func TestMonitoringSnoozesCreateHasFormatFlag(t *testing.T) {
	if monitoringSnoozesCreateCmd.Flags().Lookup("format") == nil {
		t.Error("monitoring snoozes create is missing the --format flag")
	}
}
