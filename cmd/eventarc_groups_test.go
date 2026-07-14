package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func eventarcSubgroup(name string) *cobra.Command {
	for _, c := range eventarcCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestEventarcChannelConnectionsSubcommands(t *testing.T) {
	g := eventarcSubgroup("channel-connections")
	if g == nil {
		t.Fatal("channel-connections missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestEventarcChannelsSubcommands(t *testing.T) {
	g := eventarcSubgroup("channels")
	if g == nil {
		t.Fatal("channels missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestEventarcEnrollmentsSubcommands(t *testing.T) {
	g := eventarcSubgroup("enrollments")
	if g == nil {
		t.Fatal("enrollments missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestEventarcGoogleApiSourcesSubcommands(t *testing.T) {
	g := eventarcSubgroup("google-api-sources")
	if g == nil {
		t.Fatal("google-api-sources missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestEventarcGoogleChannelsSubcommands(t *testing.T) {
	g := eventarcSubgroup("google-channels")
	if g == nil {
		t.Fatal("google-channels missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestEventarcLocationsSubcommands(t *testing.T) {
	g := eventarcSubgroup("locations")
	if g == nil {
		t.Fatal("locations missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestEventarcMessageBusesSubcommands(t *testing.T) {
	g := eventarcSubgroup("message-buses")
	if g == nil {
		t.Fatal("message-buses missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "list-enrollments", "publish", "update"})
}

func TestEventarcPipelinesSubcommands(t *testing.T) {
	g := eventarcSubgroup("pipelines")
	if g == nil {
		t.Fatal("pipelines missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestEventarcProvidersSubcommands(t *testing.T) {
	g := eventarcSubgroup("providers")
	if g == nil {
		t.Fatal("providers missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestEventarcTriggersSubcommands(t *testing.T) {
	g := eventarcSubgroup("triggers")
	if g == nil {
		t.Fatal("triggers missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
