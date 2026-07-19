package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func seSubgroup(name string) *cobra.Command {
	for _, c := range serviceExtensionsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestServiceExtensionsAuthzExtensionsSubcommands(t *testing.T) {
	g := seSubgroup("authz-extensions")
	if g == nil {
		t.Fatal("service-extensions authz-extensions missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestServiceExtensionsLbEdgeExtensionsSubcommands(t *testing.T) {
	g := seSubgroup("lb-edge-extensions")
	if g == nil {
		t.Fatal("service-extensions lb-edge-extensions missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestServiceExtensionsLbRouteExtensionsSubcommands(t *testing.T) {
	g := seSubgroup("lb-route-extensions")
	if g == nil {
		t.Fatal("service-extensions lb-route-extensions missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestServiceExtensionsLbTrafficExtensionsSubcommands(t *testing.T) {
	g := seSubgroup("lb-traffic-extensions")
	if g == nil {
		t.Fatal("service-extensions lb-traffic-extensions missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestServiceExtensionsWasmPluginsSubcommands(t *testing.T) {
	g := seSubgroup("wasm-plugins")
	if g == nil {
		t.Fatal("service-extensions wasm-plugins missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestServiceExtensionsWasmPluginVersionsSubcommands(t *testing.T) {
	g := seSubgroup("wasm-plugin-versions")
	if g == nil {
		t.Fatal("service-extensions wasm-plugin-versions missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}
