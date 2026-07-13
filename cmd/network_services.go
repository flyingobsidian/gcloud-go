package cmd

import "github.com/spf13/cobra"

// --- gcloud network-services (#364) ---

var networkServicesCmd = &cobra.Command{Use: "network-services", Short: "Manage Network Services"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(networkServicesCmd, "authz-extensions", "Manage authz extensions", crud...)
	registerStubGroup(networkServicesCmd, "edge-cache-keysets", "Manage edge cache keysets", crud...)
	registerStubGroup(networkServicesCmd, "edge-cache-origins", "Manage edge cache origins", crud...)
	registerStubGroup(networkServicesCmd, "edge-cache-services", "Manage edge cache services", crud...)
	registerStubGroup(networkServicesCmd, "endpoint-policies", "Manage endpoint policies", crud...)
	registerStubGroup(networkServicesCmd, "gateways", "Manage gateways", crud...)
	registerStubGroup(networkServicesCmd, "grpc-routes", "Manage gRPC routes", crud...)
	registerStubGroup(networkServicesCmd, "http-routes", "Manage HTTP routes", crud...)
	registerStubGroup(networkServicesCmd, "lb-route-extensions", "Manage LB route extensions", crud...)
	registerStubGroup(networkServicesCmd, "lb-traffic-extensions", "Manage LB traffic extensions", crud...)
	registerStubGroup(networkServicesCmd, "meshes", "Manage service meshes", crud...)
	registerStubGroup(networkServicesCmd, "multicast-consumer-associations", "Manage multicast consumer associations", crud...)
	registerStubGroup(networkServicesCmd, "multicast-domains", "Manage multicast domains", crud...)
	registerStubGroup(networkServicesCmd, "multicast-group-consumer-activations", "Manage multicast group consumer activations", crud...)
	registerStubGroup(networkServicesCmd, "multicast-group-producer-activations", "Manage multicast group producer activations", crud...)
	registerStubGroup(networkServicesCmd, "multicast-group-range-activations", "Manage multicast group range activations", crud...)
	registerStubGroup(networkServicesCmd, "multicast-group-ranges", "Manage multicast group ranges", crud...)
	registerStubGroup(networkServicesCmd, "multicast-producer-associations", "Manage multicast producer associations", crud...)
	registerStubGroup(networkServicesCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(networkServicesCmd, "route-views", "View route views", "describe", "list")
	registerStubGroup(networkServicesCmd, "service-bindings", "Manage service bindings", crud...)
	registerStubGroup(networkServicesCmd, "service-lb-policies", "Manage service LB policies", crud...)
	registerStubGroup(networkServicesCmd, "tcp-routes", "Manage TCP routes", crud...)
	registerStubGroup(networkServicesCmd, "tls-routes", "Manage TLS routes", crud...)
		registerStubGroup(networkServicesCmd, "agent-gateways", "Manage agent-gateways", "list", "describe")
	registerStubGroup(networkServicesCmd, "multicast-domain-activations", "Manage multicast-domain-activations", "list", "describe")
	registerStubGroup(networkServicesCmd, "multicast-domain-groups", "Manage multicast-domain-groups", "list", "describe")
	rootCmd.AddCommand(networkServicesCmd)
}
