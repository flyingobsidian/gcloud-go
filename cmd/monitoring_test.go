package cmd

import (
	"strings"
	"testing"

	monitoring "google.golang.org/api/monitoring/v3"
)

// Ensures `monitoring snoozes create` accepts --format (both `--format x` and
// `--format=x` are handled by cobra once the flag is registered).
func TestMonitoringSnoozesCreateHasFormatFlag(t *testing.T) {
	if monitoringSnoozesCreateCmd.Flags().Lookup("format") == nil {
		t.Error("monitoring snoozes create is missing the --format flag")
	}
}

func sampleSnoozes() []*monitoring.Snooze {
	return []*monitoring.Snooze{
		{Name: "projects/P/snoozes/1234", DisplayName: "SnoozeName1"},
		{Name: "projects/P/snoozes/1235", DisplayName: "SnoozeName2"},
	}
}

func TestFormatSnoozesListDefaultsToYAML(t *testing.T) {
	out := captureStdout(t, func() {
		if err := formatSnoozesList(sampleSnoozes(), ""); err != nil {
			t.Fatalf("formatSnoozesList: %v", err)
		}
	})
	if !strings.HasPrefix(strings.TrimSpace(out), "---") {
		t.Errorf("yaml output should start with a document separator, got:\n%s", out)
	}
	if !strings.Contains(out, "name: projects/P/snoozes/1234") {
		t.Errorf("yaml output missing snooze name, got:\n%s", out)
	}
	if !strings.Contains(out, "displayName: SnoozeName1") {
		t.Errorf("yaml output missing displayName, got:\n%s", out)
	}
}

func TestFormatSnoozesListCSVWithHeadings(t *testing.T) {
	out := captureStdout(t, func() {
		if err := formatSnoozesList(sampleSnoozes(), "csv(NAME,DISPLAY_NAME)"); err != nil {
			t.Fatalf("formatSnoozesList: %v", err)
		}
	})
	want := "NAME,DISPLAY_NAME\n" +
		"projects/P/snoozes/1234,SnoozeName1\n" +
		"projects/P/snoozes/1235,SnoozeName2\n"
	if out != want {
		t.Errorf("csv output =\n%q\nwant\n%q", out, want)
	}
}

func TestSnoozeField(t *testing.T) {
	s := &monitoring.Snooze{Name: "n", DisplayName: "d"}
	cases := map[string]string{
		"NAME":         "n",
		"name":         "n",
		"DISPLAY_NAME": "d",
		"displayName":  "d",
		"BOGUS":        "",
	}
	for field, want := range cases {
		if got := snoozeField(s, field); got != want {
			t.Errorf("snoozeField(%q) = %q, want %q", field, got, want)
		}
	}
}
