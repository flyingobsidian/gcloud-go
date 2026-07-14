package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func apphubSubgroup(name string) *cobra.Command {
	for _, c := range apphubCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestApphubApplicationsSubcommands(t *testing.T) {
	g := apphubSubgroup("applications")
	if g == nil {
		t.Fatal("apphub applications missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "services", "update", "workloads"})
	svc := findSub(g, "services")
	if svc == nil {
		t.Fatal("applications services missing")
	}
	assertSubcommands(t, svc, []string{"create", "delete", "describe", "list", "update"})
	wl := findSub(g, "workloads")
	if wl == nil {
		t.Fatal("applications workloads missing")
	}
	assertSubcommands(t, wl, []string{"create", "delete", "describe", "list", "update"})
}

func TestApphubBoundarySubcommands(t *testing.T) {
	g := apphubSubgroup("boundary")
	if g == nil {
		t.Fatal("apphub boundary missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestApphubDiscoveredServicesSubcommands(t *testing.T) {
	g := apphubSubgroup("discovered-services")
	if g == nil {
		t.Fatal("apphub discovered-services missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "lookup"})
}

func TestApphubDiscoveredWorkloadsSubcommands(t *testing.T) {
	g := apphubSubgroup("discovered-workloads")
	if g == nil {
		t.Fatal("apphub discovered-workloads missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "lookup"})
}

func TestApphubLocationsSubcommands(t *testing.T) {
	g := apphubSubgroup("locations")
	if g == nil {
		t.Fatal("apphub locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestApphubOperationsSubcommands(t *testing.T) {
	g := apphubSubgroup("operations")
	if g == nil {
		t.Fatal("apphub operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}

func TestApphubServiceProjectsSubcommands(t *testing.T) {
	g := apphubSubgroup("service-projects")
	if g == nil {
		t.Fatal("apphub service-projects missing")
	}
	assertSubcommands(t, g, []string{"add", "describe", "list", "lookup", "remove"})
}
