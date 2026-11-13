package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud apihub (#298) ---

func (g *apihubCmds) initApihub() {

}

// apihubLocationParent returns projects/PROJECT/locations/LOCATION and errors
// if location is empty. All apihub subgroups are location-scoped.
func (g *apihubCmds) apihubLocationParent(location string) (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// apihubResource returns projects/.../locations/.../COLLECTION/ID. If id is
// already fully qualified it is returned as-is.
func (g *apihubCmds) apihubResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := g.apihubLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// apihubCmds holds the commands and flag values of one invocation's "apihub"
// command group, so invocations don't share state (#2070).
type apihubCmds struct {
	inv *invocation

	apihubCmd                            *cobra.Command
	apihubAddonsCmd                      *cobra.Command
	flagApihubAddonsLocation             string
	flagApihubAddonsFormat               string
	flagApihubAddonsFilter               string
	flagApihubAddonsPageSize             int64
	flagApihubAddonsConfigFile           string
	apihubAddonsDescribeCmd              *cobra.Command
	apihubAddonsListCmd                  *cobra.Command
	apihubAddonsManageConfigCmd          *cobra.Command
	apihubAHICmd                         *cobra.Command
	flagAHILocation                      string
	flagAHIFormat                        string
	flagAHIDescription                   string
	flagAHILabels                        []string
	flagAHICmekKey                       string
	flagAHIDisableSearch                 bool
	flagAHIVertexLoc                     string
	flagAHIEncryption                    string
	apihubAHICreateCmd                   *cobra.Command
	apihubAHIDeleteCmd                   *cobra.Command
	apihubAHIDescribeCmd                 *cobra.Command
	apihubAHILookupCmd                   *cobra.Command
	apihubApisCmd                        *cobra.Command
	flagApihubApisLocation               string
	flagApihubApisFormat                 string
	flagApihubApisDestination            string
	flagApihubApisSource                 string
	flagApihubApisPageSize               int64
	apihubApisDeleteCmd                  *cobra.Command
	apihubApisDescribeCmd                *cobra.Command
	apihubApisExportCmd                  *cobra.Command
	apihubApisImportCmd                  *cobra.Command
	apihubApisListCmd                    *cobra.Command
	apihubAttrsCmd                       *cobra.Command
	flagApihubAttrsLocation              string
	flagApihubAttrsFormat                string
	flagApihubAttrsDestination           string
	flagApihubAttrsSource                string
	flagApihubAttrsPageSize              int64
	apihubAttrsDeleteCmd                 *cobra.Command
	apihubAttrsDescribeCmd               *cobra.Command
	apihubAttrsExportCmd                 *cobra.Command
	apihubAttrsImportCmd                 *cobra.Command
	apihubAttrsListCmd                   *cobra.Command
	apihubCurCmd                         *cobra.Command
	flagApihubCurLocation                string
	flagApihubCurFormat                  string
	flagApihubCurDestination             string
	flagApihubCurSource                  string
	flagApihubCurPageSize                int64
	apihubCurDeleteCmd                   *cobra.Command
	apihubCurDescribeCmd                 *cobra.Command
	apihubCurExportCmd                   *cobra.Command
	apihubCurImportCmd                   *cobra.Command
	apihubCurListCmd                     *cobra.Command
	apihubDepCmd                         *cobra.Command
	flagApihubDepLocation                string
	flagApihubDepFormat                  string
	flagApihubDepDestination             string
	flagApihubDepSource                  string
	flagApihubDepPageSize                int64
	apihubDepDeleteCmd                   *cobra.Command
	apihubDepDescribeCmd                 *cobra.Command
	apihubDepExportCmd                   *cobra.Command
	apihubDepImportCmd                   *cobra.Command
	apihubDepListCmd                     *cobra.Command
	apihubDeplCmd                        *cobra.Command
	flagApihubDeplLocation               string
	flagApihubDeplFormat                 string
	flagApihubDeplDestination            string
	flagApihubDeplSource                 string
	flagApihubDeplPageSize               int64
	apihubDeplDeleteCmd                  *cobra.Command
	apihubDeplDescribeCmd                *cobra.Command
	apihubDeplExportCmd                  *cobra.Command
	apihubDeplImportCmd                  *cobra.Command
	apihubDeplListCmd                    *cobra.Command
	apihubDiscoveredApiObservationsCmd   *cobra.Command
	flagApihubDaoLocation                string
	flagApihubDaoFormat                  string
	flagApihubDaoPageSize                int64
	apihubDaoDescribeCmd                 *cobra.Command
	apihubDaoListCmd                     *cobra.Command
	apihubExtCmd                         *cobra.Command
	flagApihubExtLocation                string
	flagApihubExtFormat                  string
	flagApihubExtDestination             string
	flagApihubExtSource                  string
	flagApihubExtPageSize                int64
	apihubExtDeleteCmd                   *cobra.Command
	apihubExtDescribeCmd                 *cobra.Command
	apihubExtExportCmd                   *cobra.Command
	apihubExtImportCmd                   *cobra.Command
	apihubExtListCmd                     *cobra.Command
	apihubHostProjectRegistrationsCmd    *cobra.Command
	flagApihubHprLocation                string
	flagApihubHprFormat                  string
	flagApihubHprConfigFile              string
	flagApihubHprPageSize                int64
	apihubHprCreateCmd                   *cobra.Command
	apihubHprDescribeCmd                 *cobra.Command
	apihubHprListCmd                     *cobra.Command
	apihubLocationsCmd                   *cobra.Command
	apihubLocationsConfigureAndDeployCmd *cobra.Command
	flagApihubCDFormat                   string
	flagApihubCDMcpTools                 []string
	flagApihubCDMcpToolsFromFile         string
	flagApihubCDApigeeXEnv               string
	flagApihubCDApigeeXProxy             string
	flagApihubCDApigeeXTargetProj        string
	flagApihubCDApigeeXDisplay           string
	flagApihubCDApigeeXDescr             string
	apihubOperationsCmd                  *cobra.Command
	flagApihubOpLocation                 string
	flagApihubOpFormat                   string
	flagApihubOpFilter                   string
	flagApihubOpPageSize                 int64
	apihubOpCancelCmd                    *cobra.Command
	apihubOpDeleteCmd                    *cobra.Command
	apihubOpDescribeCmd                  *cobra.Command
	apihubOpListCmd                      *cobra.Command
	apihubPluginsCmd                     *cobra.Command
	apihubPluginsInstancesCmd            *cobra.Command
	flagAPPluginsLocation                string
	flagAPPluginsFormat                  string
	flagAPPluginsFilter                  string
	flagAPPluginsPageSize                int64
	flagAPPluginsConfigFile              string
	flagAPPluginsPlugin                  string
	flagAPPluginsUpdateMask              string
	flagAPPluginsInstanceCfg             string
	apihubPluginsCreateCmd               *cobra.Command
	apihubPluginsDeleteCmd               *cobra.Command
	apihubPluginsDescribeCmd             *cobra.Command
	apihubPluginsDisableCmd              *cobra.Command
	apihubPluginsEnableCmd               *cobra.Command
	apihubPluginsListCmd                 *cobra.Command
	apihubPluginsInstCreateCmd           *cobra.Command
	apihubPluginsInstDeleteCmd           *cobra.Command
	apihubPluginsInstDescribeCmd         *cobra.Command
	apihubPluginsInstListCmd             *cobra.Command
	apihubPluginsInstUpdateCmd           *cobra.Command
	apihubRPACmd                         *cobra.Command
	flagAPRPALocation                    string
	flagAPRPAFormat                      string
	flagAPRPAConfigFile                  string
	flagAPRPAFilter                      string
	flagAPRPAPageSize                    int64
	flagAPRPARuntimeProject              string
	apihubRPACreateCmd                   *cobra.Command
	apihubRPADeleteCmd                   *cobra.Command
	apihubRPADescribeCmd                 *cobra.Command
	apihubRPAListCmd                     *cobra.Command
	apihubRPALookupCmd                   *cobra.Command
}

// newApihubCmd returns the "apihub" command group for inv.
func newApihubCmd(inv *invocation) *cobra.Command {
	return newApihubCmds(inv).apihubCmd
}

// newApihubCmds builds the "apihub" command group for inv and returns its state.
func newApihubCmds(inv *invocation) *apihubCmds {
	g := &apihubCmds{inv: inv}
	g.apihubCmd = &cobra.Command{Use: "apihub", Short: "Manage API Hub"}
	g.apihubAddonsCmd = &cobra.Command{Use: "addons", Short: "Manage API Hub addons"}
	g.apihubAddonsDescribeCmd = &cobra.Command{
		Use: "describe ADDON", Short: "Describe an API Hub addon",
		Args: cobra.ExactArgs(1), RunE: g.runApihubAddonDescribe,
	}
	g.apihubAddonsListCmd = &cobra.Command{
		Use: "list", Short: "List API Hub addons in a location",
		Args: cobra.NoArgs, RunE: g.runApihubAddonList,
	}
	g.apihubAddonsManageConfigCmd = &cobra.Command{
		Use: "manage-config ADDON", Short: "Manage the config of an API Hub addon",
		Args: cobra.ExactArgs(1), RunE: g.runApihubAddonManageConfig,
	}
	g.apihubAHICmd = &cobra.Command{Use: "api-hub-instances", Short: "Manage API Hub instances"}
	g.apihubAHICreateCmd = &cobra.Command{
		Use: "create API_HUB_INSTANCE", Short: "Create an API Hub instance",
		Args: cobra.ExactArgs(1), RunE: g.runAHICreate,
	}
	g.apihubAHIDeleteCmd = &cobra.Command{
		Use: "delete API_HUB_INSTANCE", Short: "Delete an API Hub instance",
		Args: cobra.ExactArgs(1), RunE: g.runAHIDelete,
	}
	g.apihubAHIDescribeCmd = &cobra.Command{
		Use: "describe API_HUB_INSTANCE", Short: "Describe an API Hub instance",
		Args: cobra.ExactArgs(1), RunE: g.runAHIDescribe,
	}
	g.apihubAHILookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up the API Hub instance for a project location",
		Args: cobra.NoArgs, RunE: g.runAHILookup,
	}
	g.apihubApisCmd = &cobra.Command{Use: "apis", Short: "Manage API Hub APIs"}
	g.apihubApisDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an API",
		Args: cobra.ExactArgs(1), RunE: g.runApihubApisDelete,
	}
	g.apihubApisDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an API",
		Args: cobra.ExactArgs(1), RunE: g.runApihubApisDescribe,
	}
	g.apihubApisExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an API to a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubApisExport,
	}
	g.apihubApisImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an API from a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubApisImport,
	}
	g.apihubApisListCmd = &cobra.Command{
		Use: "list", Short: "List APIs in a location",
		Args: cobra.NoArgs, RunE: g.runApihubApisList,
	}
	g.apihubAttrsCmd = &cobra.Command{Use: "attributes", Short: "Manage API Hub attributes"}
	g.apihubAttrsDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an attribute",
		Args: cobra.ExactArgs(1), RunE: g.runApihubAttrsDelete,
	}
	g.apihubAttrsDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an attribute",
		Args: cobra.ExactArgs(1), RunE: g.runApihubAttrsDescribe,
	}
	g.apihubAttrsExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an attribute to a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubAttrsExport,
	}
	g.apihubAttrsImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an attribute from a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubAttrsImport,
	}
	g.apihubAttrsListCmd = &cobra.Command{
		Use: "list", Short: "List attributes in a location",
		Args: cobra.NoArgs, RunE: g.runApihubAttrsList,
	}
	g.apihubCurCmd = &cobra.Command{Use: "curations", Short: "Manage API Hub curations"}
	g.apihubCurDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete a curation",
		Args: cobra.ExactArgs(1), RunE: g.runApihubCurDelete,
	}
	g.apihubCurDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a curation",
		Args: cobra.ExactArgs(1), RunE: g.runApihubCurDescribe,
	}
	g.apihubCurExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export a curation to a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubCurExport,
	}
	g.apihubCurImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import a curation from a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubCurImport,
	}
	g.apihubCurListCmd = &cobra.Command{
		Use: "list", Short: "List curations in a location",
		Args: cobra.NoArgs, RunE: g.runApihubCurList,
	}
	g.apihubDepCmd = &cobra.Command{Use: "dependencies", Short: "Manage API Hub dependencies"}
	g.apihubDepDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete a dependency",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDepDelete,
	}
	g.apihubDepDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a dependency",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDepDescribe,
	}
	g.apihubDepExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export a dependency to a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDepExport,
	}
	g.apihubDepImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import a dependency from a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDepImport,
	}
	g.apihubDepListCmd = &cobra.Command{
		Use: "list", Short: "List dependencies in a location",
		Args: cobra.NoArgs, RunE: g.runApihubDepList,
	}
	g.apihubDeplCmd = &cobra.Command{Use: "deployments", Short: "Manage API Hub deployments"}
	g.apihubDeplDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete a deployment",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDeplDelete,
	}
	g.apihubDeplDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a deployment",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDeplDescribe,
	}
	g.apihubDeplExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export a deployment to a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDeplExport,
	}
	g.apihubDeplImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import a deployment from a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDeplImport,
	}
	g.apihubDeplListCmd = &cobra.Command{
		Use: "list", Short: "List deployments in a location",
		Args: cobra.NoArgs, RunE: g.runApihubDeplList,
	}
	g.apihubDiscoveredApiObservationsCmd = &cobra.Command{Use: "discovered-api-observations", Short: "Manage API Hub discovered API observations"}
	g.apihubDaoDescribeCmd = &cobra.Command{
		Use: "describe OBSERVATION", Short: "Describe a discovered API observation",
		Args: cobra.ExactArgs(1), RunE: g.runApihubDaoDescribe,
	}
	g.apihubDaoListCmd = &cobra.Command{
		Use: "list", Short: "List discovered API observations in a location",
		Args: cobra.NoArgs, RunE: g.runApihubDaoList,
	}
	g.apihubExtCmd = &cobra.Command{Use: "external-apis", Short: "Manage API Hub external APIs"}
	g.apihubExtDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an external API",
		Args: cobra.ExactArgs(1), RunE: g.runApihubExtDelete,
	}
	g.apihubExtDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an external API",
		Args: cobra.ExactArgs(1), RunE: g.runApihubExtDescribe,
	}
	g.apihubExtExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an external API to a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubExtExport,
	}
	g.apihubExtImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an external API from a YAML file",
		Args: cobra.ExactArgs(1), RunE: g.runApihubExtImport,
	}
	g.apihubExtListCmd = &cobra.Command{
		Use: "list", Short: "List external APIs in a location",
		Args: cobra.NoArgs, RunE: g.runApihubExtList,
	}
	g.apihubHostProjectRegistrationsCmd = &cobra.Command{Use: "host-project-registrations", Short: "Manage API Hub host project registrations"}
	g.apihubHprCreateCmd = &cobra.Command{
		Use: "create REGISTRATION", Short: "Create a host project registration",
		Args: cobra.ExactArgs(1), RunE: g.runApihubHprCreate,
	}
	g.apihubHprDescribeCmd = &cobra.Command{
		Use: "describe REGISTRATION", Short: "Describe a host project registration",
		Args: cobra.ExactArgs(1), RunE: g.runApihubHprDescribe,
	}
	g.apihubHprListCmd = &cobra.Command{
		Use: "list", Short: "List host project registrations in a location",
		Args: cobra.NoArgs, RunE: g.runApihubHprList,
	}
	g.apihubLocationsCmd = &cobra.Command{
		Use:   "locations",
		Short: "Manage API Hub location-level resources",
	}
	g.apihubLocationsConfigureAndDeployCmd = &cobra.Command{
		Use:   "configure-and-deploy-server LOCATION",
		Short: "Configure and deploy an API Hub MCP server (currently on Apigee X)",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runApihubLocationsConfigureAndDeploy,
	}
	g.apihubOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage API Hub long-running operations"}
	g.apihubOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an API Hub operation",
		Args: cobra.ExactArgs(1), RunE: g.runApihubOpCancel,
	}
	g.apihubOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an API Hub operation",
		Args: cobra.ExactArgs(1), RunE: g.runApihubOpDelete,
	}
	g.apihubOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an API Hub operation",
		Args: cobra.ExactArgs(1), RunE: g.runApihubOpDescribe,
	}
	g.apihubOpListCmd = &cobra.Command{
		Use: "list", Short: "List API Hub operations in a location",
		Args: cobra.NoArgs, RunE: g.runApihubOpList,
	}
	g.apihubPluginsCmd = &cobra.Command{Use: "plugins", Short: "Manage API Hub plugins"}
	g.apihubPluginsInstancesCmd = &cobra.Command{
		Use: "instances", Short: "Manage API Hub plugin instances",
	}
	g.apihubPluginsCreateCmd = &cobra.Command{
		Use: "create PLUGIN", Short: "Create an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginCreate,
	}
	g.apihubPluginsDeleteCmd = &cobra.Command{
		Use: "delete PLUGIN", Short: "Delete an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginDelete,
	}
	g.apihubPluginsDescribeCmd = &cobra.Command{
		Use: "describe PLUGIN", Short: "Describe an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginDescribe,
	}
	g.apihubPluginsDisableCmd = &cobra.Command{
		Use: "disable PLUGIN", Short: "Disable an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginDisable,
	}
	g.apihubPluginsEnableCmd = &cobra.Command{
		Use: "enable PLUGIN", Short: "Enable an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginEnable,
	}
	g.apihubPluginsListCmd = &cobra.Command{
		Use: "list", Short: "List API Hub plugins in a location",
		Args: cobra.NoArgs, RunE: g.runAPPluginList,
	}
	g.apihubPluginsInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create an API Hub plugin instance",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginInstCreate,
	}
	g.apihubPluginsInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an API Hub plugin instance",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginInstDelete,
	}
	g.apihubPluginsInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an API Hub plugin instance",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginInstDescribe,
	}
	g.apihubPluginsInstListCmd = &cobra.Command{
		Use: "list", Short: "List API Hub plugin instances under a plugin",
		Args: cobra.NoArgs, RunE: g.runAPPluginInstList,
	}
	g.apihubPluginsInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update an API Hub plugin instance",
		Args: cobra.ExactArgs(1), RunE: g.runAPPluginInstUpdate,
	}
	g.apihubRPACmd = &cobra.Command{
		Use:   "runtime-project-attachments",
		Short: "Manage API Hub runtime project attachments",
	}
	g.apihubRPACreateCmd = &cobra.Command{
		Use: "create ATTACHMENT", Short: "Create a runtime project attachment",
		Args: cobra.ExactArgs(1), RunE: g.runAPRPACreate,
	}
	g.apihubRPADeleteCmd = &cobra.Command{
		Use: "delete ATTACHMENT", Short: "Delete a runtime project attachment",
		Args: cobra.ExactArgs(1), RunE: g.runAPRPADelete,
	}
	g.apihubRPADescribeCmd = &cobra.Command{
		Use: "describe ATTACHMENT", Short: "Describe a runtime project attachment",
		Args: cobra.ExactArgs(1), RunE: g.runAPRPADescribe,
	}
	g.apihubRPAListCmd = &cobra.Command{
		Use: "list", Short: "List runtime project attachments in a location",
		Args: cobra.NoArgs, RunE: g.runAPRPAList,
	}
	g.apihubRPALookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up the runtime project attachment for a runtime project",
		Args: cobra.NoArgs, RunE: g.runAPRPALookup,
	}
	g.initApihub()
	g.initApihubAddons()
	g.initApihubApiHubInstances()
	g.initApihubApis()
	g.initApihubAttributes()
	g.initApihubCurations()
	g.initApihubDependencies()
	g.initApihubDeployments()
	g.initApihubDiscoveredApiObservations()
	g.initApihubExternalApis()
	g.initApihubHostProjectRegistrations()
	g.initApihubLocations()
	g.initApihubOperations()
	g.initApihubPlugins()
	g.initApihubRuntimeProjectAttachments()
	return g
}
