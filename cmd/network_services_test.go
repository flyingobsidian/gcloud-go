package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func networkServicesSubgroup(name string) *cobra.Command {
	for _, c := range networkServicesCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestNetworkServicesEndpointPoliciesSubcommands(t *testing.T) {
	g := networkServicesSubgroup("endpoint-policies")
	if g == nil {
		t.Fatal("network-services endpoint-policies missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesGatewaysSubcommands(t *testing.T) {
	g := networkServicesSubgroup("gateways")
	if g == nil {
		t.Fatal("network-services gateways missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesGrpcRoutesSubcommands(t *testing.T) {
	g := networkServicesSubgroup("grpc-routes")
	if g == nil {
		t.Fatal("network-services grpc-routes missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesHttpRoutesSubcommands(t *testing.T) {
	g := networkServicesSubgroup("http-routes")
	if g == nil {
		t.Fatal("network-services http-routes missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesMeshesSubcommands(t *testing.T) {
	g := networkServicesSubgroup("meshes")
	if g == nil {
		t.Fatal("network-services meshes missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesTcpRoutesSubcommands(t *testing.T) {
	g := networkServicesSubgroup("tcp-routes")
	if g == nil {
		t.Fatal("network-services tcp-routes missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesTlsRoutesSubcommands(t *testing.T) {
	g := networkServicesSubgroup("tls-routes")
	if g == nil {
		t.Fatal("network-services tls-routes missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesOperationsSubcommands(t *testing.T) {
	g := networkServicesSubgroup("operations")
	if g == nil {
		t.Fatal("network-services operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "describe", "list", "wait"})
}
