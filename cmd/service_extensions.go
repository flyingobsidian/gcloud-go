package cmd

import "github.com/spf13/cobra"

// --- gcloud service-extensions (#383) ---

var serviceExtensionsCmd = &cobra.Command{Use: "service-extensions", Short: "Manage Service Extensions"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(serviceExtensionsCmd, "authz-extensions", "Manage AuthzExtension resources", crud...)
	registerStubGroup(serviceExtensionsCmd, "lb-edge-extensions", "Manage LbEdgeExtension resources", crud...)
	registerStubGroup(serviceExtensionsCmd, "lb-route-extensions", "Manage LbRouteExtension resources", crud...)
	registerStubGroup(serviceExtensionsCmd, "lb-traffic-extensions", "Manage LbTrafficExtension resources", crud...)
	registerStubGroup(serviceExtensionsCmd, "wasm-plugin-versions", "Manage WasmPluginVersions", "create", "delete", "describe", "list")
	registerStubGroup(serviceExtensionsCmd, "wasm-plugins", "Manage WasmPlugins", crud...)
	rootCmd.AddCommand(serviceExtensionsCmd)
}
