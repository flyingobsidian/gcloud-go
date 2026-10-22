package cmd

import (
	"bytes"
	"strings"
	"testing"

	monitoring "google.golang.org/api/monitoring/v3"
)

func TestMonitoringAlertName(t *testing.T) {
	if got := monitoringAlertName("p", "a1"); got != "projects/p/alerts/a1" {
		t.Errorf("short id: %q", got)
	}
	if got := monitoringAlertName("p", "projects/q/alerts/a1"); got != "projects/q/alerts/a1" {
		t.Errorf("full name: %q", got)
	}
}

func TestEmitMonitoringAlerts(t *testing.T) {
	alerts := []*monitoring.Alert{
		{Name: "projects/p/alerts/a1", State: "OPEN"},
		{Name: "projects/p/alerts/a2", State: "CLOSED"},
	}
	var buf bytes.Buffer
	if err := emitMonitoringAlerts(&buf, alerts, "", false); err != nil {
		t.Fatal(err)
	}
	want := "name: projects/p/alerts/a1\nstate: OPEN\n---\nname: projects/p/alerts/a2\nstate: CLOSED\n"
	if buf.String() != want {
		t.Errorf("yaml:\n%s\nwant:\n%s", buf.String(), want)
	}

	buf.Reset()
	if err := emitMonitoringAlerts(&buf, alerts, "", true); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "projects/p/alerts/a1\nprojects/p/alerts/a2\n" {
		t.Errorf("uri: %q", buf.String())
	}

	buf.Reset()
	if err := emitMonitoringAlerts(&buf, nil, "json", false); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Errorf("empty json: %q", buf.String())
	}
}

func TestMonitoringAlertsRegistered(t *testing.T) {
	c, _, err := monitoringCmd.Find([]string{"alerts"})
	if err != nil || c != monitoringAlertsCmd {
		t.Fatalf("monitoring alerts not registered: %v", err)
	}
	assertSubcommands(t, c, []string{"describe", "list"})
	if f := monitoringAlertsListCmd.Flags().Lookup("page-size"); f == nil || f.DefValue != "1000" {
		t.Error("list --page-size should default to 1000")
	}
}
