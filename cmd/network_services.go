package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud network-services (#364) ---

var networkServicesCmd = &cobra.Command{Use: "network-services", Short: "Manage Network Services"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	// Groups implemented in dedicated files (this batch: endpoint-policies,
	// gateways, grpc-routes, http-routes, meshes, tcp-routes, tls-routes,
	// operations) unregister themselves from this stub list.
	registerStubGroup(networkServicesCmd, "authz-extensions", "Manage authz extensions", crud...)
	registerStubGroup(networkServicesCmd, "edge-cache-keysets", "Manage edge cache keysets", crud...)
	registerStubGroup(networkServicesCmd, "edge-cache-origins", "Manage edge cache origins", crud...)
	registerStubGroup(networkServicesCmd, "edge-cache-services", "Manage edge cache services", crud...)
	registerStubGroup(networkServicesCmd, "lb-route-extensions", "Manage LB route extensions", crud...)
	registerStubGroup(networkServicesCmd, "lb-traffic-extensions", "Manage LB traffic extensions", crud...)
	registerStubGroup(networkServicesCmd, "multicast-consumer-associations", "Manage multicast consumer associations", crud...)
	registerStubGroup(networkServicesCmd, "multicast-domains", "Manage multicast domains", crud...)
	registerStubGroup(networkServicesCmd, "multicast-group-consumer-activations", "Manage multicast group consumer activations", crud...)
	registerStubGroup(networkServicesCmd, "multicast-group-producer-activations", "Manage multicast group producer activations", crud...)
	registerStubGroup(networkServicesCmd, "multicast-group-range-activations", "Manage multicast group range activations", crud...)
	registerStubGroup(networkServicesCmd, "multicast-group-ranges", "Manage multicast group ranges", crud...)
	registerStubGroup(networkServicesCmd, "multicast-producer-associations", "Manage multicast producer associations", crud...)
	registerStubGroup(networkServicesCmd, "route-views", "View route views", "describe", "list")
	registerStubGroup(networkServicesCmd, "multicast-domain-activations", "Manage multicast-domain-activations", "list", "describe")
	registerStubGroup(networkServicesCmd, "multicast-domain-groups", "Manage multicast-domain-groups", "list", "describe")
	rootCmd.AddCommand(networkServicesCmd)
}

// nsLocationParent returns "projects/PROJECT/locations/LOCATION". Callers pass
// the value of the shared --location flag.
func nsLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// nsResourceName returns a fully qualified resource name for a network-services
// collection under a location. If id is already a full URI it is returned as-is.
func nsResourceName(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := nsLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// nsBindExportImportFlags binds the shared --destination/--source flag pair
// used by every network-services export/import command pair.
func nsBindExportFlags(c *cobra.Command, dest *string) {
	c.Flags().StringVar(dest, "destination", "", "Path to write the exported YAML/JSON (required)")
	_ = c.MarkFlagRequired("destination")
}

func nsBindImportFlags(c *cobra.Command, src *string) {
	c.Flags().StringVar(src, "source", "", "Path to a YAML/JSON file with the resource body (required)")
	_ = c.MarkFlagRequired("source")
}
