package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud vmware (#395) ---

func (s *vmwareCmds) initVmware() {

}

// vmwareLocationParent returns projects/PROJECT/locations/LOCATION.
func (s *vmwareCmds) vmwareLocationParent(location string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// vmwareResource returns projects/.../COLLECTION/ID.
func (s *vmwareCmds) vmwareResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.vmwareLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// vmwareCmds holds the commands and flag values of one invocation's "vmware"
// command group, so invocations don't share state (#2070).
type vmwareCmds struct {
	inv *invocation

	vmwareCmd                   *cobra.Command
	vmwareAnnouncementsCmd      *cobra.Command
	flagVmwareAnnLocation       string
	flagVmwareAnnFormat         string
	flagVmwareAnnPageSize       int64
	vmwareAnnDescribeCmd        *cobra.Command
	vmwareAnnListCmd            *cobra.Command
	vmwareDatastoresCmd         *cobra.Command
	flagVmwareDsLocation        string
	flagVmwareDsFormat          string
	flagVmwareDsConfigFile      string
	flagVmwareDsUpdateMask      string
	flagVmwareDsPageSize        int64
	vmwareDsCreateCmd           *cobra.Command
	vmwareDsDeleteCmd           *cobra.Command
	vmwareDsDescribeCmd         *cobra.Command
	vmwareDsListCmd             *cobra.Command
	vmwareDsUpdateCmd           *cobra.Command
	vmwareDnsBindPermissionCmd  *cobra.Command
	flagVmwareDbpLocation       string
	flagVmwareDbpFormat         string
	flagVmwareDbpConfigFile     string
	vmwareDbpDescribeCmd        *cobra.Command
	vmwareDbpGrantCmd           *cobra.Command
	vmwareDbpRevokeCmd          *cobra.Command
	vmwareLocationsCmd          *cobra.Command
	flagVmwareLocFormat         string
	flagVmwareLocFilter         string
	flagVmwareLocPageSize       int64
	vmwareLocDescribeCmd        *cobra.Command
	vmwareLocListCmd            *cobra.Command
	vmwareNetworkPeeringsCmd    *cobra.Command
	flagVmwareNpLocation        string
	flagVmwareNpFormat          string
	flagVmwareNpConfigFile      string
	flagVmwareNpUpdateMask      string
	flagVmwareNpPageSize        int64
	vmwareNpCreateCmd           *cobra.Command
	vmwareNpDeleteCmd           *cobra.Command
	vmwareNpDescribeCmd         *cobra.Command
	vmwareNpListCmd             *cobra.Command
	vmwareNpUpdateCmd           *cobra.Command
	vmwareNetworkPoliciesCmd    *cobra.Command
	flagVmwareNpolLocation      string
	flagVmwareNpolFormat        string
	flagVmwareNpolConfigFile    string
	flagVmwareNpolUpdateMask    string
	flagVmwareNpolPageSize      int64
	vmwareNpolCreateCmd         *cobra.Command
	vmwareNpolDeleteCmd         *cobra.Command
	vmwareNpolDescribeCmd       *cobra.Command
	vmwareNpolListCmd           *cobra.Command
	vmwareNpolUpdateCmd         *cobra.Command
	vmwareNetworksCmd           *cobra.Command
	flagVmwareNetLocation       string
	flagVmwareNetFormat         string
	flagVmwareNetConfigFile     string
	flagVmwareNetUpdateMask     string
	flagVmwareNetPageSize       int64
	vmwareNetCreateCmd          *cobra.Command
	vmwareNetDeleteCmd          *cobra.Command
	vmwareNetDescribeCmd        *cobra.Command
	vmwareNetListCmd            *cobra.Command
	vmwareNetUpdateCmd          *cobra.Command
	vmwareNodeTypesCmd          *cobra.Command
	flagVmwareNtLocation        string
	flagVmwareNtFormat          string
	flagVmwareNtPageSize        int64
	vmwareNtDescribeCmd         *cobra.Command
	vmwareNtListCmd             *cobra.Command
	vmwareOperationsCmd         *cobra.Command
	flagVmwareOpLocation        string
	flagVmwareOpFormat          string
	flagVmwareOpFilter          string
	flagVmwareOpPageSize        int64
	vmwareOpDeleteCmd           *cobra.Command
	vmwareOpDescribeCmd         *cobra.Command
	vmwareOpListCmd             *cobra.Command
	vmwarePrivateCloudsCmd      *cobra.Command
	flagVmwarePcLocation        string
	flagVmwarePcFormat          string
	flagVmwarePcConfigFile      string
	flagVmwarePcUpdateMask      string
	flagVmwarePcPageSize        int64
	flagVmwarePcClusterID       string
	flagVmwarePcRequestID       string
	flagVmwarePcEtag            string
	vmwarePcCreateCmd           *cobra.Command
	vmwarePcDeleteCmd           *cobra.Command
	vmwarePcDescribeCmd         *cobra.Command
	vmwarePcListCmd             *cobra.Command
	vmwarePcUpdateCmd           *cobra.Command
	vmwarePcMigrateMgmtVmsCmd   *cobra.Command
	vmwarePrivateConnectionsCmd *cobra.Command
	flagVmwarePconLocation      string
	flagVmwarePconFormat        string
	flagVmwarePconConfigFile    string
	flagVmwarePconUpdateMask    string
	flagVmwarePconPageSize      int64
	vmwarePconCreateCmd         *cobra.Command
	vmwarePconDeleteCmd         *cobra.Command
	vmwarePconDescribeCmd       *cobra.Command
	vmwarePconListCmd           *cobra.Command
	vmwarePconUpdateCmd         *cobra.Command
}

// newVmwareCmd returns the "vmware" command group for inv.
func newVmwareCmd(inv *invocation) *cobra.Command {
	return newVmwareCmds(inv).vmwareCmd
}

// newVmwareCmds builds the "vmware" command group for inv and returns its state.
func newVmwareCmds(inv *invocation) *vmwareCmds {
	s := &vmwareCmds{inv: inv}
	s.vmwareCmd = &cobra.Command{Use: "vmware", Short: "Manage VMware Engine"}
	s.vmwareAnnouncementsCmd = &cobra.Command{Use: "announcements", Short: "Manage VMware Engine announcements"}
	s.vmwareAnnDescribeCmd = &cobra.Command{
		Use: "describe ANNOUNCEMENT", Short: "Describe a VMware Engine announcement",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareAnnDescribe,
	}
	s.vmwareAnnListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine announcements in a location",
		Args: cobra.NoArgs, RunE: s.runVmwareAnnList,
	}
	s.vmwareDatastoresCmd = &cobra.Command{Use: "datastores", Short: "Manage VMware Engine datastores"}
	s.vmwareDsCreateCmd = &cobra.Command{
		Use: "create DATASTORE", Short: "Create a VMware Engine datastore",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareDsCreate,
	}
	s.vmwareDsDeleteCmd = &cobra.Command{
		Use: "delete DATASTORE", Short: "Delete a VMware Engine datastore",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareDsDelete,
	}
	s.vmwareDsDescribeCmd = &cobra.Command{
		Use: "describe DATASTORE", Short: "Describe a VMware Engine datastore",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareDsDescribe,
	}
	s.vmwareDsListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine datastores in a location",
		Args: cobra.NoArgs, RunE: s.runVmwareDsList,
	}
	s.vmwareDsUpdateCmd = &cobra.Command{
		Use: "update DATASTORE", Short: "Update a VMware Engine datastore",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareDsUpdate,
	}
	s.vmwareDnsBindPermissionCmd = &cobra.Command{Use: "dns-bind-permission", Short: "Manage the DNS bind permission for the project"}
	s.vmwareDbpDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the DNS bind permission for the project",
		Args: cobra.NoArgs, RunE: s.runVmwareDbpDescribe,
	}
	s.vmwareDbpGrantCmd = &cobra.Command{
		Use: "grant", Short: "Grant DNS bind permission to a principal",
		Args: cobra.NoArgs, RunE: s.runVmwareDbpGrant,
	}
	s.vmwareDbpRevokeCmd = &cobra.Command{
		Use: "revoke", Short: "Revoke DNS bind permission from a principal",
		Args: cobra.NoArgs, RunE: s.runVmwareDbpRevoke,
	}
	s.vmwareLocationsCmd = &cobra.Command{Use: "locations", Short: "List VMware Engine locations"}
	s.vmwareLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a VMware Engine location",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareLocDescribe,
	}
	s.vmwareLocListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine locations",
		Args: cobra.NoArgs, RunE: s.runVmwareLocList,
	}
	s.vmwareNetworkPeeringsCmd = &cobra.Command{Use: "network-peerings", Short: "Manage VMware Engine network peerings"}
	s.vmwareNpCreateCmd = &cobra.Command{
		Use: "create PEERING", Short: "Create a VMware Engine network peering",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNpCreate,
	}
	s.vmwareNpDeleteCmd = &cobra.Command{
		Use: "delete PEERING", Short: "Delete a VMware Engine network peering",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNpDelete,
	}
	s.vmwareNpDescribeCmd = &cobra.Command{
		Use: "describe PEERING", Short: "Describe a VMware Engine network peering",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNpDescribe,
	}
	s.vmwareNpListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine network peerings in a location",
		Args: cobra.NoArgs, RunE: s.runVmwareNpList,
	}
	s.vmwareNpUpdateCmd = &cobra.Command{
		Use: "update PEERING", Short: "Update a VMware Engine network peering",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNpUpdate,
	}
	s.vmwareNetworkPoliciesCmd = &cobra.Command{Use: "network-policies", Short: "Manage VMware Engine network policies"}
	s.vmwareNpolCreateCmd = &cobra.Command{
		Use: "create POLICY", Short: "Create a VMware Engine network policy",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNpolCreate,
	}
	s.vmwareNpolDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete a VMware Engine network policy",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNpolDelete,
	}
	s.vmwareNpolDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe a VMware Engine network policy",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNpolDescribe,
	}
	s.vmwareNpolListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine network policies in a location",
		Args: cobra.NoArgs, RunE: s.runVmwareNpolList,
	}
	s.vmwareNpolUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update a VMware Engine network policy",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNpolUpdate,
	}
	s.vmwareNetworksCmd = &cobra.Command{Use: "networks", Short: "Manage VMware Engine networks"}
	s.vmwareNetCreateCmd = &cobra.Command{
		Use: "create NETWORK", Short: "Create a VMware Engine network",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNetCreate,
	}
	s.vmwareNetDeleteCmd = &cobra.Command{
		Use: "delete NETWORK", Short: "Delete a VMware Engine network",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNetDelete,
	}
	s.vmwareNetDescribeCmd = &cobra.Command{
		Use: "describe NETWORK", Short: "Describe a VMware Engine network",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNetDescribe,
	}
	s.vmwareNetListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine networks in a location",
		Args: cobra.NoArgs, RunE: s.runVmwareNetList,
	}
	s.vmwareNetUpdateCmd = &cobra.Command{
		Use: "update NETWORK", Short: "Update a VMware Engine network",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNetUpdate,
	}
	s.vmwareNodeTypesCmd = &cobra.Command{Use: "node-types", Short: "Show VMware Engine node types"}
	s.vmwareNtDescribeCmd = &cobra.Command{
		Use: "describe NODE_TYPE", Short: "Describe a VMware Engine node type",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareNtDescribe,
	}
	s.vmwareNtListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine node types in a location",
		Args: cobra.NoArgs, RunE: s.runVmwareNtList,
	}
	s.vmwareOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage VMware Engine long-running operations"}
	s.vmwareOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a VMware Engine operation",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareOpDelete,
	}
	s.vmwareOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a VMware Engine operation",
		Args: cobra.ExactArgs(1), RunE: s.runVmwareOpDescribe,
	}
	s.vmwareOpListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine operations in a location",
		Args: cobra.NoArgs, RunE: s.runVmwareOpList,
	}
	s.vmwarePrivateCloudsCmd = &cobra.Command{Use: "private-clouds", Short: "Manage VMware Engine private clouds"}
	s.vmwarePcCreateCmd = &cobra.Command{
		Use: "create PRIVATE_CLOUD", Short: "Create a VMware Engine private cloud",
		Args: cobra.ExactArgs(1), RunE: s.runVmwarePcCreate,
	}
	s.vmwarePcDeleteCmd = &cobra.Command{
		Use: "delete PRIVATE_CLOUD", Short: "Delete a VMware Engine private cloud",
		Args: cobra.ExactArgs(1), RunE: s.runVmwarePcDelete,
	}
	s.vmwarePcDescribeCmd = &cobra.Command{
		Use: "describe PRIVATE_CLOUD", Short: "Describe a VMware Engine private cloud",
		Args: cobra.ExactArgs(1), RunE: s.runVmwarePcDescribe,
	}
	s.vmwarePcListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine private clouds in a location",
		Args: cobra.NoArgs, RunE: s.runVmwarePcList,
	}
	s.vmwarePcUpdateCmd = &cobra.Command{
		Use: "update PRIVATE_CLOUD", Short: "Update a VMware Engine private cloud",
		Args: cobra.ExactArgs(1), RunE: s.runVmwarePcUpdate,
	}
	s.vmwarePcMigrateMgmtVmsCmd = &cobra.Command{
		Use:   "migrate-management-vms PRIVATE_CLOUD",
		Short: "Migrate management VMs from the current management cluster to a workload cluster",
		Args:  cobra.ExactArgs(1), RunE: s.runVmwarePcMigrateMgmtVms,
	}
	s.vmwarePrivateConnectionsCmd = &cobra.Command{Use: "private-connections", Short: "Manage VMware Engine private connections"}
	s.vmwarePconCreateCmd = &cobra.Command{
		Use: "create PRIVATE_CONNECTION", Short: "Create a VMware Engine private connection",
		Args: cobra.ExactArgs(1), RunE: s.runVmwarePconCreate,
	}
	s.vmwarePconDeleteCmd = &cobra.Command{
		Use: "delete PRIVATE_CONNECTION", Short: "Delete a VMware Engine private connection",
		Args: cobra.ExactArgs(1), RunE: s.runVmwarePconDelete,
	}
	s.vmwarePconDescribeCmd = &cobra.Command{
		Use: "describe PRIVATE_CONNECTION", Short: "Describe a VMware Engine private connection",
		Args: cobra.ExactArgs(1), RunE: s.runVmwarePconDescribe,
	}
	s.vmwarePconListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine private connections in a location",
		Args: cobra.NoArgs, RunE: s.runVmwarePconList,
	}
	s.vmwarePconUpdateCmd = &cobra.Command{
		Use: "update PRIVATE_CONNECTION", Short: "Update a VMware Engine private connection",
		Args: cobra.ExactArgs(1), RunE: s.runVmwarePconUpdate,
	}
	s.initVmware()
	s.initVmwareAnnouncements()
	s.initVmwareDatastores()
	s.initVmwareDnsBindPermission()
	s.initVmwareLocations()
	s.initVmwareNetworkPeerings()
	s.initVmwareNetworkPolicies()
	s.initVmwareNetworks()
	s.initVmwareNodeTypes()
	s.initVmwareOperations()
	s.initVmwarePrivateClouds()
	s.initVmwarePrivateConnections()
	return s
}
