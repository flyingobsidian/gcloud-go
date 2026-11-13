package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud network-services (#364) ---

func (s *networkServicesCmds) initNetworkServices() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	// Groups implemented in dedicated files (this batch: endpoint-policies,
	// gateways, grpc-routes, http-routes, meshes, tcp-routes, tls-routes,
	// operations) unregister themselves from this stub list.
	// edge-cache-* live under the top-level `edge-cache` command (see internal/cli/edge_cache*.go).
	// authz-extensions / lb-*-extensions live under `service-extensions` (see internal/cli/service_extensions*.go).
	registerStubGroup(s.networkServicesCmd, "multicast-domains", "Manage multicast domains", crud...)
	registerStubGroup(s.networkServicesCmd, "multicast-group-producer-activations", "Manage multicast group producer activations", crud...)
	registerStubGroup(s.networkServicesCmd, "multicast-group-range-activations", "Manage multicast group range activations", crud...)
	registerStubGroup(s.networkServicesCmd, "multicast-group-ranges", "Manage multicast group ranges", crud...)
	registerStubGroup(s.networkServicesCmd, "multicast-producer-associations", "Manage multicast producer associations", crud...)
	registerStubGroup(s.networkServicesCmd, "multicast-domain-activations", "Manage multicast-domain-activations", "list", "describe")
	registerStubGroup(s.networkServicesCmd, "multicast-domain-groups", "Manage multicast-domain-groups", "list", "describe")

}

// nsLocationParent returns "projects/PROJECT/locations/LOCATION". Callers pass
// the value of the shared --location flag.
func (s *networkServicesCmds) nsLocationParent(location string) (string, error) {
	project, err := s.inv.project()
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
func (s *networkServicesCmds) nsResourceName(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.nsLocationParent(location)
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

// networkServicesCmds holds the commands and flag values of one invocation's "networkServices"
// command group, so invocations don't share state (#2070).
type networkServicesCmds struct {
	inv *invocation

	networkServicesCmd                                  *cobra.Command
	networkServicesAgentGatewaysCmd                     *cobra.Command
	flagNsAgLocation                                    string
	flagNsAgFormat                                      string
	flagNsAgDestination                                 string
	flagNsAgSource                                      string
	flagNsAgPageSize                                    int64
	networkServicesAgDeleteCmd                          *cobra.Command
	networkServicesAgDescribeCmd                        *cobra.Command
	networkServicesAgExportCmd                          *cobra.Command
	networkServicesAgImportCmd                          *cobra.Command
	networkServicesAgListCmd                            *cobra.Command
	networkServicesEndpointPoliciesCmd                  *cobra.Command
	flagNsEpLocation                                    string
	flagNsEpFormat                                      string
	flagNsEpDestination                                 string
	flagNsEpSource                                      string
	flagNsEpPageSize                                    int64
	networkServicesEpDeleteCmd                          *cobra.Command
	networkServicesEpDescribeCmd                        *cobra.Command
	networkServicesEpExportCmd                          *cobra.Command
	networkServicesEpImportCmd                          *cobra.Command
	networkServicesEpListCmd                            *cobra.Command
	networkServicesGatewaysCmd                          *cobra.Command
	flagNsGwLocation                                    string
	flagNsGwFormat                                      string
	flagNsGwDestination                                 string
	flagNsGwSource                                      string
	flagNsGwPageSize                                    int64
	networkServicesGwDeleteCmd                          *cobra.Command
	networkServicesGwDescribeCmd                        *cobra.Command
	networkServicesGwExportCmd                          *cobra.Command
	networkServicesGwImportCmd                          *cobra.Command
	networkServicesGwListCmd                            *cobra.Command
	networkServicesGrpcRoutesCmd                        *cobra.Command
	flagNsGrLocation                                    string
	flagNsGrFormat                                      string
	flagNsGrDestination                                 string
	flagNsGrSource                                      string
	flagNsGrPageSize                                    int64
	networkServicesGrDeleteCmd                          *cobra.Command
	networkServicesGrDescribeCmd                        *cobra.Command
	networkServicesGrExportCmd                          *cobra.Command
	networkServicesGrImportCmd                          *cobra.Command
	networkServicesGrListCmd                            *cobra.Command
	networkServicesHttpRoutesCmd                        *cobra.Command
	flagNsHrLocation                                    string
	flagNsHrFormat                                      string
	flagNsHrDestination                                 string
	flagNsHrSource                                      string
	flagNsHrPageSize                                    int64
	networkServicesHrDeleteCmd                          *cobra.Command
	networkServicesHrDescribeCmd                        *cobra.Command
	networkServicesHrExportCmd                          *cobra.Command
	networkServicesHrImportCmd                          *cobra.Command
	networkServicesHrListCmd                            *cobra.Command
	networkServicesMeshesCmd                            *cobra.Command
	flagNsMeshLocation                                  string
	flagNsMeshFormat                                    string
	flagNsMeshDestination                               string
	flagNsMeshSource                                    string
	flagNsMeshPageSize                                  int64
	networkServicesMeshDeleteCmd                        *cobra.Command
	networkServicesMeshDescribeCmd                      *cobra.Command
	networkServicesMeshExportCmd                        *cobra.Command
	networkServicesMeshImportCmd                        *cobra.Command
	networkServicesMeshListCmd                          *cobra.Command
	networkServicesMulticastConsumerAssociationsCmd     *cobra.Command
	flagNsMcaLocation                                   string
	flagNsMcaFormat                                     string
	flagNsMcaDestination                                string
	flagNsMcaSource                                     string
	flagNsMcaPageSize                                   int64
	networkServicesMcaDeleteCmd                         *cobra.Command
	networkServicesMcaDescribeCmd                       *cobra.Command
	networkServicesMcaExportCmd                         *cobra.Command
	networkServicesMcaImportCmd                         *cobra.Command
	networkServicesMcaListCmd                           *cobra.Command
	networkServicesMulticastGroupConsumerActivationsCmd *cobra.Command
	flagNsMgcaLocation                                  string
	flagNsMgcaFormat                                    string
	flagNsMgcaDestination                               string
	flagNsMgcaSource                                    string
	flagNsMgcaPageSize                                  int64
	networkServicesMgcaDeleteCmd                        *cobra.Command
	networkServicesMgcaDescribeCmd                      *cobra.Command
	networkServicesMgcaExportCmd                        *cobra.Command
	networkServicesMgcaImportCmd                        *cobra.Command
	networkServicesMgcaListCmd                          *cobra.Command
	networkServicesOperationsCmd                        *cobra.Command
	flagNsOpLocation                                    string
	flagNsOpFormat                                      string
	flagNsOpFilter                                      string
	flagNsOpPageSize                                    int64
	flagNsOpTimeout                                     time.Duration
	networkServicesOpCancelCmd                          *cobra.Command
	networkServicesOpDescribeCmd                        *cobra.Command
	networkServicesOpListCmd                            *cobra.Command
	networkServicesOpWaitCmd                            *cobra.Command
	networkServicesRouteViewsCmd                        *cobra.Command
	flagNsRvLocation                                    string
	flagNsRvGateway                                     string
	flagNsRvMesh                                        string
	flagNsRvFormat                                      string
	flagNsRvPageSize                                    int64
	networkServicesRvDescribeCmd                        *cobra.Command
	networkServicesRvListCmd                            *cobra.Command
	networkServicesServiceBindingsCmd                   *cobra.Command
	flagNsSbLocation                                    string
	flagNsSbFormat                                      string
	flagNsSbDestination                                 string
	flagNsSbSource                                      string
	flagNsSbConfigFile                                  string
	flagNsSbUpdateMask                                  string
	flagNsSbPageSize                                    int64
	networkServicesSbCreateCmd                          *cobra.Command
	networkServicesSbDeleteCmd                          *cobra.Command
	networkServicesSbDescribeCmd                        *cobra.Command
	networkServicesSbExportCmd                          *cobra.Command
	networkServicesSbImportCmd                          *cobra.Command
	networkServicesSbListCmd                            *cobra.Command
	networkServicesSbUpdateCmd                          *cobra.Command
	networkServicesServiceLbPoliciesCmd                 *cobra.Command
	flagNsSlLocation                                    string
	flagNsSlFormat                                      string
	flagNsSlDestination                                 string
	flagNsSlSource                                      string
	flagNsSlConfigFile                                  string
	flagNsSlUpdateMask                                  string
	flagNsSlPageSize                                    int64
	networkServicesSlCreateCmd                          *cobra.Command
	networkServicesSlDeleteCmd                          *cobra.Command
	networkServicesSlDescribeCmd                        *cobra.Command
	networkServicesSlExportCmd                          *cobra.Command
	networkServicesSlImportCmd                          *cobra.Command
	networkServicesSlListCmd                            *cobra.Command
	networkServicesSlUpdateCmd                          *cobra.Command
	networkServicesTcpRoutesCmd                         *cobra.Command
	flagNsTcLocation                                    string
	flagNsTcFormat                                      string
	flagNsTcDestination                                 string
	flagNsTcSource                                      string
	flagNsTcPageSize                                    int64
	networkServicesTcDeleteCmd                          *cobra.Command
	networkServicesTcDescribeCmd                        *cobra.Command
	networkServicesTcExportCmd                          *cobra.Command
	networkServicesTcImportCmd                          *cobra.Command
	networkServicesTcListCmd                            *cobra.Command
	networkServicesTlsRoutesCmd                         *cobra.Command
	flagNsTlLocation                                    string
	flagNsTlFormat                                      string
	flagNsTlDestination                                 string
	flagNsTlSource                                      string
	flagNsTlPageSize                                    int64
	networkServicesTlDeleteCmd                          *cobra.Command
	networkServicesTlDescribeCmd                        *cobra.Command
	networkServicesTlExportCmd                          *cobra.Command
	networkServicesTlImportCmd                          *cobra.Command
	networkServicesTlListCmd                            *cobra.Command
}

// newNetworkServicesCmd returns the "networkServices" command group for inv.
func newNetworkServicesCmd(inv *invocation) *cobra.Command {
	return newNetworkServicesCmds(inv).networkServicesCmd
}

// newNetworkServicesCmds builds the "networkServices" command group for inv and returns its state.
func newNetworkServicesCmds(inv *invocation) *networkServicesCmds {
	s := &networkServicesCmds{inv: inv}
	s.networkServicesCmd = &cobra.Command{Use: "network-services", Short: "Manage Network Services"}
	s.networkServicesAgentGatewaysCmd = &cobra.Command{Use: "agent-gateways", Short: "Manage agent gateways"}
	s.networkServicesAgDeleteCmd = &cobra.Command{
		Use: "delete GATEWAY", Short: "Delete an agent gateway",
		Args: cobra.ExactArgs(1), RunE: s.runNsAgDelete,
	}
	s.networkServicesAgDescribeCmd = &cobra.Command{
		Use: "describe GATEWAY", Short: "Describe an agent gateway",
		Args: cobra.ExactArgs(1), RunE: s.runNsAgDescribe,
	}
	s.networkServicesAgExportCmd = &cobra.Command{
		Use: "export GATEWAY", Short: "Export an agent gateway to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsAgExport,
	}
	s.networkServicesAgImportCmd = &cobra.Command{
		Use: "import GATEWAY", Short: "Import an agent gateway from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsAgImport,
	}
	s.networkServicesAgListCmd = &cobra.Command{
		Use: "list", Short: "List agent gateways in a location",
		Args: cobra.NoArgs, RunE: s.runNsAgList,
	}
	s.networkServicesEndpointPoliciesCmd = &cobra.Command{Use: "endpoint-policies", Short: "Manage endpoint policies"}
	s.networkServicesEpDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete an endpoint policy",
		Args: cobra.ExactArgs(1), RunE: s.runNsEpDelete,
	}
	s.networkServicesEpDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe an endpoint policy",
		Args: cobra.ExactArgs(1), RunE: s.runNsEpDescribe,
	}
	s.networkServicesEpExportCmd = &cobra.Command{
		Use: "export POLICY", Short: "Export an endpoint policy to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsEpExport,
	}
	s.networkServicesEpImportCmd = &cobra.Command{
		Use: "import POLICY", Short: "Import an endpoint policy from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsEpImport,
	}
	s.networkServicesEpListCmd = &cobra.Command{
		Use: "list", Short: "List endpoint policies in a location",
		Args: cobra.NoArgs, RunE: s.runNsEpList,
	}
	s.networkServicesGatewaysCmd = &cobra.Command{Use: "gateways", Short: "Manage gateways"}
	s.networkServicesGwDeleteCmd = &cobra.Command{
		Use: "delete GATEWAY", Short: "Delete a gateway",
		Args: cobra.ExactArgs(1), RunE: s.runNsGwDelete,
	}
	s.networkServicesGwDescribeCmd = &cobra.Command{
		Use: "describe GATEWAY", Short: "Describe a gateway",
		Args: cobra.ExactArgs(1), RunE: s.runNsGwDescribe,
	}
	s.networkServicesGwExportCmd = &cobra.Command{
		Use: "export GATEWAY", Short: "Export a gateway to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsGwExport,
	}
	s.networkServicesGwImportCmd = &cobra.Command{
		Use: "import GATEWAY", Short: "Import a gateway from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsGwImport,
	}
	s.networkServicesGwListCmd = &cobra.Command{
		Use: "list", Short: "List gateways in a location",
		Args: cobra.NoArgs, RunE: s.runNsGwList,
	}
	s.networkServicesGrpcRoutesCmd = &cobra.Command{Use: "grpc-routes", Short: "Manage gRPC routes"}
	s.networkServicesGrDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete a gRPC route",
		Args: cobra.ExactArgs(1), RunE: s.runNsGrDelete,
	}
	s.networkServicesGrDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe a gRPC route",
		Args: cobra.ExactArgs(1), RunE: s.runNsGrDescribe,
	}
	s.networkServicesGrExportCmd = &cobra.Command{
		Use: "export ROUTE", Short: "Export a gRPC route to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsGrExport,
	}
	s.networkServicesGrImportCmd = &cobra.Command{
		Use: "import ROUTE", Short: "Import a gRPC route from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsGrImport,
	}
	s.networkServicesGrListCmd = &cobra.Command{
		Use: "list", Short: "List gRPC routes in a location",
		Args: cobra.NoArgs, RunE: s.runNsGrList,
	}
	s.networkServicesHttpRoutesCmd = &cobra.Command{Use: "http-routes", Short: "Manage HTTP routes"}
	s.networkServicesHrDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete an HTTP route",
		Args: cobra.ExactArgs(1), RunE: s.runNsHrDelete,
	}
	s.networkServicesHrDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe an HTTP route",
		Args: cobra.ExactArgs(1), RunE: s.runNsHrDescribe,
	}
	s.networkServicesHrExportCmd = &cobra.Command{
		Use: "export ROUTE", Short: "Export an HTTP route to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsHrExport,
	}
	s.networkServicesHrImportCmd = &cobra.Command{
		Use: "import ROUTE", Short: "Import an HTTP route from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsHrImport,
	}
	s.networkServicesHrListCmd = &cobra.Command{
		Use: "list", Short: "List HTTP routes in a location",
		Args: cobra.NoArgs, RunE: s.runNsHrList,
	}
	s.networkServicesMeshesCmd = &cobra.Command{Use: "meshes", Short: "Manage service meshes"}
	s.networkServicesMeshDeleteCmd = &cobra.Command{
		Use: "delete MESH", Short: "Delete a service mesh",
		Args: cobra.ExactArgs(1), RunE: s.runNsMeshDelete,
	}
	s.networkServicesMeshDescribeCmd = &cobra.Command{
		Use: "describe MESH", Short: "Describe a service mesh",
		Args: cobra.ExactArgs(1), RunE: s.runNsMeshDescribe,
	}
	s.networkServicesMeshExportCmd = &cobra.Command{
		Use: "export MESH", Short: "Export a service mesh to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsMeshExport,
	}
	s.networkServicesMeshImportCmd = &cobra.Command{
		Use: "import MESH", Short: "Import a service mesh from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsMeshImport,
	}
	s.networkServicesMeshListCmd = &cobra.Command{
		Use: "list", Short: "List service meshes in a location",
		Args: cobra.NoArgs, RunE: s.runNsMeshList,
	}
	s.networkServicesMulticastConsumerAssociationsCmd = &cobra.Command{Use: "multicast-consumer-associations", Short: "Manage multicast consumer associations"}
	s.networkServicesMcaDeleteCmd = &cobra.Command{
		Use: "delete ASSOCIATION", Short: "Delete a multicast consumer association",
		Args: cobra.ExactArgs(1), RunE: s.runNsMcaDelete,
	}
	s.networkServicesMcaDescribeCmd = &cobra.Command{
		Use: "describe ASSOCIATION", Short: "Describe a multicast consumer association",
		Args: cobra.ExactArgs(1), RunE: s.runNsMcaDescribe,
	}
	s.networkServicesMcaExportCmd = &cobra.Command{
		Use: "export ASSOCIATION", Short: "Export a multicast consumer association to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsMcaExport,
	}
	s.networkServicesMcaImportCmd = &cobra.Command{
		Use: "import ASSOCIATION", Short: "Import a multicast consumer association from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsMcaImport,
	}
	s.networkServicesMcaListCmd = &cobra.Command{
		Use: "list", Short: "List multicast consumer associations in a location",
		Args: cobra.NoArgs, RunE: s.runNsMcaList,
	}
	s.networkServicesMulticastGroupConsumerActivationsCmd = &cobra.Command{Use: "multicast-group-consumer-activations", Short: "Manage multicast group consumer activations"}
	s.networkServicesMgcaDeleteCmd = &cobra.Command{
		Use: "delete ACTIVATION", Short: "Delete a multicast group consumer activation",
		Args: cobra.ExactArgs(1), RunE: s.runNsMgcaDelete,
	}
	s.networkServicesMgcaDescribeCmd = &cobra.Command{
		Use: "describe ACTIVATION", Short: "Describe a multicast group consumer activation",
		Args: cobra.ExactArgs(1), RunE: s.runNsMgcaDescribe,
	}
	s.networkServicesMgcaExportCmd = &cobra.Command{
		Use: "export ACTIVATION", Short: "Export a multicast group consumer activation to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsMgcaExport,
	}
	s.networkServicesMgcaImportCmd = &cobra.Command{
		Use: "import ACTIVATION", Short: "Import a multicast group consumer activation from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsMgcaImport,
	}
	s.networkServicesMgcaListCmd = &cobra.Command{
		Use: "list", Short: "List multicast group consumer activations in a location",
		Args: cobra.NoArgs, RunE: s.runNsMgcaList,
	}
	s.networkServicesOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Network Services long-running operations"}
	s.networkServicesOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Network Services operation",
		Args: cobra.ExactArgs(1), RunE: s.runNsOpCancel,
	}
	s.networkServicesOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Network Services operation",
		Args: cobra.ExactArgs(1), RunE: s.runNsOpDescribe,
	}
	s.networkServicesOpListCmd = &cobra.Command{
		Use: "list", Short: "List Network Services operations in a location",
		Args: cobra.NoArgs, RunE: s.runNsOpList,
	}
	s.networkServicesOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Network Services operation to complete",
		Args: cobra.ExactArgs(1), RunE: s.runNsOpWait,
	}
	s.networkServicesRouteViewsCmd = &cobra.Command{Use: "route-views", Short: "View route views for a Gateway or Mesh"}
	s.networkServicesRvDescribeCmd = &cobra.Command{
		Use: "describe ROUTE_VIEW", Short: "Describe a route view for a Gateway or Mesh",
		Args: cobra.ExactArgs(1), RunE: s.runNsRvDescribe,
	}
	s.networkServicesRvListCmd = &cobra.Command{
		Use: "list", Short: "List route views for a Gateway or Mesh",
		Args: cobra.NoArgs, RunE: s.runNsRvList,
	}
	s.networkServicesServiceBindingsCmd = &cobra.Command{Use: "service-bindings", Short: "Manage service bindings"}
	s.networkServicesSbCreateCmd = &cobra.Command{
		Use: "create BINDING", Short: "Create a service binding",
		Args: cobra.ExactArgs(1), RunE: s.runNsSbCreate,
	}
	s.networkServicesSbDeleteCmd = &cobra.Command{
		Use: "delete BINDING", Short: "Delete a service binding",
		Args: cobra.ExactArgs(1), RunE: s.runNsSbDelete,
	}
	s.networkServicesSbDescribeCmd = &cobra.Command{
		Use: "describe BINDING", Short: "Describe a service binding",
		Args: cobra.ExactArgs(1), RunE: s.runNsSbDescribe,
	}
	s.networkServicesSbExportCmd = &cobra.Command{
		Use: "export BINDING", Short: "Export a service binding to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsSbExport,
	}
	s.networkServicesSbImportCmd = &cobra.Command{
		Use: "import BINDING", Short: "Import a service binding from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsSbImport,
	}
	s.networkServicesSbListCmd = &cobra.Command{
		Use: "list", Short: "List service bindings in a location",
		Args: cobra.NoArgs, RunE: s.runNsSbList,
	}
	s.networkServicesSbUpdateCmd = &cobra.Command{
		Use: "update BINDING", Short: "Update a service binding",
		Args: cobra.ExactArgs(1), RunE: s.runNsSbUpdate,
	}
	s.networkServicesServiceLbPoliciesCmd = &cobra.Command{Use: "service-lb-policies", Short: "Manage service LB policies"}
	s.networkServicesSlCreateCmd = &cobra.Command{
		Use: "create POLICY", Short: "Create a service LB policy",
		Args: cobra.ExactArgs(1), RunE: s.runNsSlCreate,
	}
	s.networkServicesSlDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete a service LB policy",
		Args: cobra.ExactArgs(1), RunE: s.runNsSlDelete,
	}
	s.networkServicesSlDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe a service LB policy",
		Args: cobra.ExactArgs(1), RunE: s.runNsSlDescribe,
	}
	s.networkServicesSlExportCmd = &cobra.Command{
		Use: "export POLICY", Short: "Export a service LB policy to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsSlExport,
	}
	s.networkServicesSlImportCmd = &cobra.Command{
		Use: "import POLICY", Short: "Import a service LB policy from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsSlImport,
	}
	s.networkServicesSlListCmd = &cobra.Command{
		Use: "list", Short: "List service LB policies in a location",
		Args: cobra.NoArgs, RunE: s.runNsSlList,
	}
	s.networkServicesSlUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update a service LB policy",
		Args: cobra.ExactArgs(1), RunE: s.runNsSlUpdate,
	}
	s.networkServicesTcpRoutesCmd = &cobra.Command{Use: "tcp-routes", Short: "Manage TCP routes"}
	s.networkServicesTcDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete a TCP route",
		Args: cobra.ExactArgs(1), RunE: s.runNsTcDelete,
	}
	s.networkServicesTcDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe a TCP route",
		Args: cobra.ExactArgs(1), RunE: s.runNsTcDescribe,
	}
	s.networkServicesTcExportCmd = &cobra.Command{
		Use: "export ROUTE", Short: "Export a TCP route to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsTcExport,
	}
	s.networkServicesTcImportCmd = &cobra.Command{
		Use: "import ROUTE", Short: "Import a TCP route from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsTcImport,
	}
	s.networkServicesTcListCmd = &cobra.Command{
		Use: "list", Short: "List TCP routes in a location",
		Args: cobra.NoArgs, RunE: s.runNsTcList,
	}
	s.networkServicesTlsRoutesCmd = &cobra.Command{Use: "tls-routes", Short: "Manage TLS routes"}
	s.networkServicesTlDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete a TLS route",
		Args: cobra.ExactArgs(1), RunE: s.runNsTlDelete,
	}
	s.networkServicesTlDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe a TLS route",
		Args: cobra.ExactArgs(1), RunE: s.runNsTlDescribe,
	}
	s.networkServicesTlExportCmd = &cobra.Command{
		Use: "export ROUTE", Short: "Export a TLS route to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsTlExport,
	}
	s.networkServicesTlImportCmd = &cobra.Command{
		Use: "import ROUTE", Short: "Import a TLS route from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runNsTlImport,
	}
	s.networkServicesTlListCmd = &cobra.Command{
		Use: "list", Short: "List TLS routes in a location",
		Args: cobra.NoArgs, RunE: s.runNsTlList,
	}
	s.initNetworkServices()
	s.initNetworkServicesAgentGateways()
	s.initNetworkServicesEndpointPolicies()
	s.initNetworkServicesGateways()
	s.initNetworkServicesGrpcRoutes()
	s.initNetworkServicesHttpRoutes()
	s.initNetworkServicesMeshes()
	s.initNetworkServicesMulticastConsumerAssociations()
	s.initNetworkServicesMulticastGroupConsumerActivations()
	s.initNetworkServicesOperations()
	s.initNetworkServicesRouteViews()
	s.initNetworkServicesServiceBindings()
	s.initNetworkServicesServiceLbPolicies()
	s.initNetworkServicesTcpRoutes()
	s.initNetworkServicesTlsRoutes()
	return s
}
