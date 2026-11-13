package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud service-extensions (#383) ---

func (s *serviceExtensionsCmds) initServiceExtensions() {

}

// seLocationParent returns projects/PROJECT/locations/LOCATION.
func (s *serviceExtensionsCmds) seLocationParent(location string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// seResource returns projects/.../COLLECTION/ID.
func (s *serviceExtensionsCmds) seResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.seLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// serviceExtensionsCmds holds the commands and flag values of one invocation's "serviceExtensions"
// command group, so invocations don't share state (#2070).
type serviceExtensionsCmds struct {
	inv *invocation

	serviceExtensionsCmd       *cobra.Command
	seAuthzExtensionsCmd       *cobra.Command
	flagSeAuthzLocation        string
	flagSeAuthzFormat          string
	flagSeAuthzDestination     string
	flagSeAuthzSource          string
	flagSeAuthzPageSize        int64
	seAuthzDeleteCmd           *cobra.Command
	seAuthzDescribeCmd         *cobra.Command
	seAuthzExportCmd           *cobra.Command
	seAuthzImportCmd           *cobra.Command
	seAuthzListCmd             *cobra.Command
	seLbEdgeExtensionsCmd      *cobra.Command
	flagSeLbEdgeLocation       string
	flagSeLbEdgeFormat         string
	flagSeLbEdgeDestination    string
	flagSeLbEdgeSource         string
	flagSeLbEdgePageSize       int64
	seLbEdgeDeleteCmd          *cobra.Command
	seLbEdgeDescribeCmd        *cobra.Command
	seLbEdgeExportCmd          *cobra.Command
	seLbEdgeImportCmd          *cobra.Command
	seLbEdgeListCmd            *cobra.Command
	seLbRouteExtensionsCmd     *cobra.Command
	flagSeLbRouteLocation      string
	flagSeLbRouteFormat        string
	flagSeLbRouteDestination   string
	flagSeLbRouteSource        string
	flagSeLbRoutePageSize      int64
	seLbRouteDeleteCmd         *cobra.Command
	seLbRouteDescribeCmd       *cobra.Command
	seLbRouteExportCmd         *cobra.Command
	seLbRouteImportCmd         *cobra.Command
	seLbRouteListCmd           *cobra.Command
	seLbTrafficExtensionsCmd   *cobra.Command
	flagSeLbTrafficLocation    string
	flagSeLbTrafficFormat      string
	flagSeLbTrafficDestination string
	flagSeLbTrafficSource      string
	flagSeLbTrafficPageSize    int64
	seLbTrafficDeleteCmd       *cobra.Command
	seLbTrafficDescribeCmd     *cobra.Command
	seLbTrafficExportCmd       *cobra.Command
	seLbTrafficImportCmd       *cobra.Command
	seLbTrafficListCmd         *cobra.Command
	seWasmPluginVersionsCmd    *cobra.Command
	flagSeWasmVerLocation      string
	flagSeWasmVerPlugin        string
	flagSeWasmVerFormat        string
	flagSeWasmVerConfigFile    string
	flagSeWasmVerPageSize      int64
	seWasmVerCreateCmd         *cobra.Command
	seWasmVerDeleteCmd         *cobra.Command
	seWasmVerDescribeCmd       *cobra.Command
	seWasmVerListCmd           *cobra.Command
	seWasmPluginsCmd           *cobra.Command
	flagSeWasmLocation         string
	flagSeWasmFormat           string
	flagSeWasmDestination      string
	flagSeWasmSource           string
	flagSeWasmPageSize         int64
	seWasmDeleteCmd            *cobra.Command
	seWasmDescribeCmd          *cobra.Command
	seWasmExportCmd            *cobra.Command
	seWasmImportCmd            *cobra.Command
	seWasmListCmd              *cobra.Command
}

// newServiceExtensionsCmd returns the "serviceExtensions" command group for inv.
func newServiceExtensionsCmd(inv *invocation) *cobra.Command {
	return newServiceExtensionsCmds(inv).serviceExtensionsCmd
}

// newServiceExtensionsCmds builds the "serviceExtensions" command group for inv and returns its state.
func newServiceExtensionsCmds(inv *invocation) *serviceExtensionsCmds {
	s := &serviceExtensionsCmds{inv: inv}
	s.serviceExtensionsCmd = &cobra.Command{Use: "service-extensions", Short: "Manage Service Extensions"}
	s.seAuthzExtensionsCmd = &cobra.Command{Use: "authz-extensions", Short: "Manage AuthzExtension resources"}
	s.seAuthzDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an AuthzExtension",
		Args: cobra.ExactArgs(1), RunE: s.runSeAuthzDelete,
	}
	s.seAuthzDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an AuthzExtension",
		Args: cobra.ExactArgs(1), RunE: s.runSeAuthzDescribe,
	}
	s.seAuthzExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an AuthzExtension to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeAuthzExport,
	}
	s.seAuthzImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an AuthzExtension from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeAuthzImport,
	}
	s.seAuthzListCmd = &cobra.Command{
		Use: "list", Short: "List AuthzExtensions in a location",
		Args: cobra.NoArgs, RunE: s.runSeAuthzList,
	}
	s.seLbEdgeExtensionsCmd = &cobra.Command{Use: "lb-edge-extensions", Short: "Manage LbEdgeExtension resources"}
	s.seLbEdgeDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an LbEdgeExtension",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbEdgeDelete,
	}
	s.seLbEdgeDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an LbEdgeExtension",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbEdgeDescribe,
	}
	s.seLbEdgeExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an LbEdgeExtension to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbEdgeExport,
	}
	s.seLbEdgeImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an LbEdgeExtension from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbEdgeImport,
	}
	s.seLbEdgeListCmd = &cobra.Command{
		Use: "list", Short: "List LbEdgeExtensions in a location",
		Args: cobra.NoArgs, RunE: s.runSeLbEdgeList,
	}
	s.seLbRouteExtensionsCmd = &cobra.Command{Use: "lb-route-extensions", Short: "Manage LbRouteExtension resources"}
	s.seLbRouteDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an LbRouteExtension",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbRouteDelete,
	}
	s.seLbRouteDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an LbRouteExtension",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbRouteDescribe,
	}
	s.seLbRouteExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an LbRouteExtension to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbRouteExport,
	}
	s.seLbRouteImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an LbRouteExtension from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbRouteImport,
	}
	s.seLbRouteListCmd = &cobra.Command{
		Use: "list", Short: "List LbRouteExtensions in a location",
		Args: cobra.NoArgs, RunE: s.runSeLbRouteList,
	}
	s.seLbTrafficExtensionsCmd = &cobra.Command{Use: "lb-traffic-extensions", Short: "Manage LbTrafficExtension resources"}
	s.seLbTrafficDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an LbTrafficExtension",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbTrafficDelete,
	}
	s.seLbTrafficDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an LbTrafficExtension",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbTrafficDescribe,
	}
	s.seLbTrafficExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an LbTrafficExtension to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbTrafficExport,
	}
	s.seLbTrafficImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an LbTrafficExtension from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeLbTrafficImport,
	}
	s.seLbTrafficListCmd = &cobra.Command{
		Use: "list", Short: "List LbTrafficExtensions in a location",
		Args: cobra.NoArgs, RunE: s.runSeLbTrafficList,
	}
	s.seWasmPluginVersionsCmd = &cobra.Command{Use: "wasm-plugin-versions", Short: "Manage WasmPluginVersion resources"}
	s.seWasmVerCreateCmd = &cobra.Command{
		Use: "create VERSION", Short: "Create a WasmPluginVersion",
		Args: cobra.ExactArgs(1), RunE: s.runSeWasmVerCreate,
	}
	s.seWasmVerDeleteCmd = &cobra.Command{
		Use: "delete VERSION", Short: "Delete a WasmPluginVersion",
		Args: cobra.ExactArgs(1), RunE: s.runSeWasmVerDelete,
	}
	s.seWasmVerDescribeCmd = &cobra.Command{
		Use: "describe VERSION", Short: "Describe a WasmPluginVersion",
		Args: cobra.ExactArgs(1), RunE: s.runSeWasmVerDescribe,
	}
	s.seWasmVerListCmd = &cobra.Command{
		Use: "list", Short: "List WasmPluginVersions under a WasmPlugin",
		Args: cobra.NoArgs, RunE: s.runSeWasmVerList,
	}
	s.seWasmPluginsCmd = &cobra.Command{Use: "wasm-plugins", Short: "Manage WasmPlugin resources"}
	s.seWasmDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete a WasmPlugin",
		Args: cobra.ExactArgs(1), RunE: s.runSeWasmDelete,
	}
	s.seWasmDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a WasmPlugin",
		Args: cobra.ExactArgs(1), RunE: s.runSeWasmDescribe,
	}
	s.seWasmExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export a WasmPlugin to a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeWasmExport,
	}
	s.seWasmImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import a WasmPlugin from a YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runSeWasmImport,
	}
	s.seWasmListCmd = &cobra.Command{
		Use: "list", Short: "List WasmPlugins in a location",
		Args: cobra.NoArgs, RunE: s.runSeWasmList,
	}
	s.initServiceExtensions()
	s.initServiceExtensionsAuthzExtensions()
	s.initServiceExtensionsLbEdgeExtensions()
	s.initServiceExtensionsLbRouteExtensions()
	s.initServiceExtensionsLbTrafficExtensions()
	s.initServiceExtensionsWasmPluginVersions()
	s.initServiceExtensionsWasmPlugins()
	return s
}
