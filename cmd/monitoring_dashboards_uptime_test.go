package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func monitoringSubgroup(name string) *cobra.Command {
	for _, c := range monitoringCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestMonitoringDashboardsSubcommands(t *testing.T) {
	g := monitoringSubgroup("dashboards")
	if g == nil {
		t.Fatal("monitoring dashboards missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestMonitoringUptimeSubcommands(t *testing.T) {
	g := monitoringSubgroup("uptime")
	if g == nil {
		t.Fatal("monitoring uptime missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestDashboardName(t *testing.T) {
	got := dashboardName("p1", "d1")
	want := "projects/p1/dashboards/d1"
	if got != want {
		t.Errorf("dashboardName = %q, want %q", got, want)
	}
	pass := "projects/p/dashboards/full"
	if dashboardName("ignored", pass) != pass {
		t.Errorf("dashboardName should pass through fully-qualified names")
	}
}

func TestUptimeName(t *testing.T) {
	got := uptimeName("p1", "u1")
	want := "projects/p1/uptimeCheckConfigs/u1"
	if got != want {
		t.Errorf("uptimeName = %q, want %q", got, want)
	}
}

func TestMonitoredResourceType(t *testing.T) {
	for in, want := range map[string]string{
		"":                        "uptime_url",
		"uptime-url":              "uptime_url",
		"gce-instance":            "gce_instance",
		"aws-ec2-instance":        "aws_ec2_instance",
		"cloud-run-revision":      "cloud_run_revision",
		"servicedirectory-service": "servicedirectory_service",
	} {
		if got := monitoredResourceType(in); got != want {
			t.Errorf("monitoredResourceType(%q) = %q, want %q", in, got, want)
		}
	}
}
