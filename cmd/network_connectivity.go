package cmd

import "github.com/spf13/cobra"

// --- gcloud network-connectivity (#361) ---

var networkConnectivityCmd = &cobra.Command{Use: "network-connectivity", Short: "Manage Network Connectivity"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(networkConnectivityCmd, "hubs", "Manage hubs", append(crud, "accept-spoke", "reject-spoke", "list-spokes")...)
	registerStubGroup(networkConnectivityCmd, "internal-ranges", "Manage internal ranges", crud...)
	registerStubGroup(networkConnectivityCmd, "locations", "Get locations", "list", "describe")
	registerStubGroup(networkConnectivityCmd, "multicloud-data-transfer-configs", "Manage multicloud data transfer configs", crud...)
	registerStubGroup(networkConnectivityCmd, "multicloud-data-transfer-supported-services", "Manage supported services", "describe", "list")
	registerStubGroup(networkConnectivityCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(networkConnectivityCmd, "policy-based-routes", "Manage policy-based routes", "create", "delete", "describe", "list")
	registerStubGroup(networkConnectivityCmd, "regional-endpoints", "Manage regional endpoints", "create", "delete", "describe", "list")
	registerStubGroup(networkConnectivityCmd, "service-connection-policies", "Manage service connection policies", crud...)
	registerStubGroup(networkConnectivityCmd, "spokes", "Manage spokes", append(crud, "accept", "reject")...)
	registerStubGroup(networkConnectivityCmd, "transports", "Manage transports", crud...)
	rootCmd.AddCommand(networkConnectivityCmd)
}
