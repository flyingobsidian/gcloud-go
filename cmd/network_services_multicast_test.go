package cmd

import "testing"

func TestNetworkServicesMulticastConsumerAssociationsSubcommands(t *testing.T) {
	g := networkServicesSubgroup("multicast-consumer-associations")
	if g == nil {
		t.Fatal("network-services multicast-consumer-associations missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesMulticastGroupConsumerActivationsSubcommands(t *testing.T) {
	g := networkServicesSubgroup("multicast-group-consumer-activations")
	if g == nil {
		t.Fatal("network-services multicast-group-consumer-activations missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestNetworkServicesRouteViewsSubcommands(t *testing.T) {
	g := networkServicesSubgroup("route-views")
	if g == nil {
		t.Fatal("network-services route-views missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}
