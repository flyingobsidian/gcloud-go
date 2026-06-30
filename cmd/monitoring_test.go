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

func TestSnoozeNameAcceptsFullResource(t *testing.T) {
	if got := snoozeName("P", "projects/P/snoozes/1234"); got != "projects/P/snoozes/1234" {
		t.Errorf("full resource name should pass through, got %q", got)
	}
	if got := snoozeName("P", "1234"); got != "projects/P/snoozes/1234" {
		t.Errorf("bare id should be expanded, got %q", got)
	}
}

func TestFormatSnoozeDefaultsToYAML(t *testing.T) {
	s := &monitoring.Snooze{Name: "projects/P/snoozes/1234", DisplayName: "SnoozeName1"}
	out := captureStdout(t, func() {
		if err := formatSnooze(s, ""); err != nil {
			t.Fatalf("formatSnooze: %v", err)
		}
	})
	if strings.HasPrefix(strings.TrimSpace(out), "---") {
		t.Errorf("single describe should not emit a document separator, got:\n%s", out)
	}
	if !strings.Contains(out, "name: projects/P/snoozes/1234") {
		t.Errorf("yaml output missing name, got:\n%s", out)
	}
	if !strings.Contains(out, "displayName: SnoozeName1") {
		t.Errorf("yaml output missing displayName, got:\n%s", out)
	}
}

func TestFormatSnoozeTableAligned(t *testing.T) {
	s := &monitoring.Snooze{Name: "projects/P/snoozes/1234", DisplayName: "SnoozeName1"}
	out := captureStdout(t, func() {
		if err := formatSnooze(s, "table(name,display_name)"); err != nil {
			t.Fatalf("formatSnooze: %v", err)
		}
	})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected heading + 1 row, got %d lines: %q", len(lines), out)
	}
	// Headings are uppercased.
	if got := strings.Fields(lines[0]); len(got) != 2 || got[0] != "NAME" || got[1] != "DISPLAY_NAME" {
		t.Errorf("heading = %q, want NAME DISPLAY_NAME", lines[0])
	}
	if got := strings.Fields(lines[1]); len(got) != 2 || got[0] != "projects/P/snoozes/1234" || got[1] != "SnoozeName1" {
		t.Errorf("row = %q, want projects/P/snoozes/1234 SnoozeName1", lines[1])
	}
	// Columns are aligned: the second column starts at the same offset.
	if strings.Index(lines[0], "DISPLAY_NAME") != strings.Index(lines[1], "SnoozeName1") {
		t.Errorf("columns not aligned:\n%q\n%q", lines[0], lines[1])
	}
}

func TestFormatPoliciesListDefaultsToYAML(t *testing.T) {
	policies := []*monitoring.AlertPolicy{
		{
			Name:        "projects/P/alertPolicies/123",
			DisplayName: "some_policy",
			Enabled:     true,
		},
	}
	out := captureStdout(t, func() {
		if err := formatPoliciesList(policies, ""); err != nil {
			t.Fatalf("formatPoliciesList: %v", err)
		}
	})
	if !strings.HasPrefix(strings.TrimSpace(out), "---") {
		t.Errorf("yaml output should start with a document separator, got:\n%s", out)
	}
	if !strings.Contains(out, "name: projects/P/alertPolicies/123") {
		t.Errorf("yaml output missing policy name, got:\n%s", out)
	}
	if !strings.Contains(out, "displayName: some_policy") {
		t.Errorf("yaml output missing displayName, got:\n%s", out)
	}
	if !strings.Contains(out, "enabled: true") {
		t.Errorf("yaml output missing enabled, got:\n%s", out)
	}
}
