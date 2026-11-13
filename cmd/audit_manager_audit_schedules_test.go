package cmd

import (
	"bytes"
	"strings"
	"testing"

	auditmanager "google.golang.org/api/auditmanager/v1"
)

func TestAuditManagerSubcommands(t *testing.T) {
	grp := newAuditManagerCmds(legacyTestInvocation())
	assertSubcommands(t, grp.auditManagerSchedulesCmd, []string{"create", "list", "update"})
	for _, c := range grp.auditManagerOperationsCmd.Commands() {
		if c.Name() == "describe" && c.RunE == nil {
			t.Error("operations describe has no RunE")
		}
	}
}

func TestAuditManagerBuildSchedule(t *testing.T) {
	grp := newAuditManagerCmds(legacyTestInvocation())
	defer func() {
		grp.flagAMReportFormat, grp.flagAMFrequency, grp.flagAMGcsURI, grp.flagAMFramework = "", "", "", ""
		grp.flagAMStartTime, grp.flagAMTimeZone = "", ""
	}()
	grp.flagAMReportFormat, grp.flagAMFrequency = "ODF", "Weekly"
	grp.flagAMGcsURI, grp.flagAMFramework, grp.flagAMStartTime, grp.flagAMTimeZone = "gs://b", "fedramp_moderate", "2026-03-01T00:00:00Z", "Europe/London"
	got, err := grp.auditManagerBuildSchedule()
	if err != nil {
		t.Fatal(err)
	}
	if got.ReportFormat != "AUDIT_REPORT_FORMAT_ODF" || got.ScheduleConfig.Frequency != "WEEKLY" ||
		got.ScheduleConfig.TimeZone != "Europe/London" || got.GcsUri != "gs://b" {
		t.Errorf("schedule = %+v %+v", got, got.ScheduleConfig)
	}
	grp.flagAMFrequency = "hourly"
	if _, err := grp.auditManagerBuildSchedule(); err == nil ||
		!strings.Contains(err.Error(), "Valid choices are [annually, daily, monthly, quarterly, weekly]") {
		t.Errorf("bad frequency: %v", err)
	}
}

func TestEmitAuditSchedules(t *testing.T) {
	var out, errOut bytes.Buffer
	all := []*auditmanager.AuditSchedule{{
		Name: "projects/p/locations/us/auditSchedules/s1", ComplianceFramework: "fedramp", State: "SCHEDULE_STATE_ACTIVE",
	}}
	if err := emitAuditSchedules(&out, &errOut, all, ""); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out.String(), "\n")
	if !strings.HasPrefix(lines[0], "ID") || !strings.Contains(lines[0], "COMPLIANCE_FRAMEWORK") ||
		!strings.HasPrefix(lines[1], "s1") {
		t.Errorf("table = %q", out.String())
	}
	out.Reset()
	if err := emitAuditSchedules(&out, &errOut, nil, ""); err != nil || errOut.String() != "Listed 0 items.\n" {
		t.Errorf("empty: %v %q", err, errOut.String())
	}
}
