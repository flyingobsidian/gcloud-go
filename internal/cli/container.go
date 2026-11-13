package cli

import "github.com/spf13/cobra"

// --- gcloud container (#320) ---
//
// Container/GKE is a very large surface. This stub registers the top-level
// subgroups and common commands so callers can discover them; individual
// commands should be wired up to google.golang.org/api/container/v1 in
// follow-up PRs.

func (s *containerCmds) initContainer() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(s.containerCmd, "ai", "Manage AI-related workloads", "profiles")
	registerStubGroup(s.containerCmd, "attached", "Manage attached clusters", append(crud, "generate-install-manifest", "import", "get-credentials")...)
	registerStubGroup(s.containerCmd, "aws", "(DEPRECATED) Manage AWS clusters", crud...)
	registerStubGroup(s.containerCmd, "azure", "(DEPRECATED) Manage Azure clusters", crud...)
	registerStubGroup(s.containerCmd, "images", "Manage container images",
		"delete", "describe", "list", "list-tags", "add-tag", "remove-tag", "untag")
	registerStubGroup(s.containerCmd, "workload", "Manage Workload Optimizer",
		"list-recommendations", "get-recommendation", "apply-recommendation")

}

// containerCmds holds the commands and flag values of one invocation's "container"
// command group, so invocations don't share state (#2070).
type containerCmds struct {
	inv *invocation

	containerCmd                    *cobra.Command
	containerBareMetalCmd           *cobra.Command
	containerBareMetalClustersCmd   *cobra.Command
	flagCbmLocation                 string
	flagCbmFormat                   string
	flagCbmConfigFile               string
	flagCbmUpdateMask               string
	flagCbmPageSize                 int64
	containerBmClCreateCmd          *cobra.Command
	containerBmClDeleteCmd          *cobra.Command
	containerBmClDescribeCmd        *cobra.Command
	containerBmClListCmd            *cobra.Command
	containerBmClUpdateCmd          *cobra.Command
	containerBinauthzCmd            *cobra.Command
	containerBinauthzPolicyCmd      *cobra.Command
	containerBinauthzAttestorsCmd   *cobra.Command
	flagBaFormat                    string
	flagBaConfigFile                string
	flagBaPageSize                  int64
	binauthzPolicyDescribeCmd       *cobra.Command
	binauthzPolicyUpdateCmd         *cobra.Command
	binauthzAttestorsCreateCmd      *cobra.Command
	binauthzAttestorsDeleteCmd      *cobra.Command
	binauthzAttestorsDescribeCmd    *cobra.Command
	binauthzAttestorsListCmd        *cobra.Command
	binauthzAttestorsUpdateCmd      *cobra.Command
	containerClustersCmd            *cobra.Command
	flagCtnClLocation               string
	flagCtnClRegion                 string
	flagCtnClFormat                 string
	flagCtnClConfigFile             string
	containerClCreateCmd            *cobra.Command
	containerClDeleteCmd            *cobra.Command
	containerClDescribeCmd          *cobra.Command
	containerClListCmd              *cobra.Command
	containerClUpdateCmd            *cobra.Command
	containerClCompleteCpUpgradeCmd *cobra.Command
	containerFleetCmd               *cobra.Command
	containerHubCmd                 *cobra.Command
	containerFleetFleetsCmd         *cobra.Command
	containerFleetFeaturesCmd       *cobra.Command
	containerFleetMembershipsCmd    *cobra.Command
	flagCfLocation                  string
	flagCfFormat                    string
	flagCfConfigFile                string
	flagCfUpdateMask                string
	flagCfPageSize                  int64
	// fleets: create/delete/describe/list/update
	containerFleetsCreateCmd   *cobra.Command
	containerFleetsDeleteCmd   *cobra.Command
	containerFleetsDescribeCmd *cobra.Command
	containerFleetsListCmd     *cobra.Command
	containerFleetsUpdateCmd   *cobra.Command
	// features: create/delete/describe/list/update
	containerFeaturesCreateCmd   *cobra.Command
	containerFeaturesDeleteCmd   *cobra.Command
	containerFeaturesDescribeCmd *cobra.Command
	containerFeaturesListCmd     *cobra.Command
	containerFeaturesUpdateCmd   *cobra.Command
	// memberships: create/delete/describe/list/update
	containerMembershipsCreateCmd   *cobra.Command
	containerMembershipsDeleteCmd   *cobra.Command
	containerMembershipsDescribeCmd *cobra.Command
	containerMembershipsListCmd     *cobra.Command
	containerMembershipsUpdateCmd   *cobra.Command
	flagCtnCfgLocation              string
	flagCtnCfgFormat                string
	containerGetServerConfigCmd     *cobra.Command
	containerNodePoolsCmd           *cobra.Command
	flagCtnNpLocation               string
	flagCtnNpCluster                string
	flagCtnNpFormat                 string
	flagCtnNpConfigFile             string
	containerNpCreateCmd            *cobra.Command
	containerNpDeleteCmd            *cobra.Command
	containerNpDescribeCmd          *cobra.Command
	containerNpListCmd              *cobra.Command
	containerNpUpdateCmd            *cobra.Command
	containerNpRollbackCmd          *cobra.Command
	containerOperationsCmd          *cobra.Command
	flagCtnOpLocation               string
	flagCtnOpFormat                 string
	containerOpCancelCmd            *cobra.Command
	containerOpDescribeCmd          *cobra.Command
	containerOpListCmd              *cobra.Command
	containerSubnetsCmd             *cobra.Command
	flagCtnSnFormat                 string
	flagCtnSnFilter                 string
	flagCtnSnPageSize               int64
	containerSubnetsListUsableCmd   *cobra.Command
	containerVmwareCmd              *cobra.Command
	containerVmwareClustersCmd      *cobra.Command
	flagCvmLocation                 string
	flagCvmFormat                   string
	flagCvmConfigFile               string
	flagCvmUpdateMask               string
	flagCvmPageSize                 int64
	containerVmClCreateCmd          *cobra.Command
	containerVmClDeleteCmd          *cobra.Command
	containerVmClDescribeCmd        *cobra.Command
	containerVmClListCmd            *cobra.Command
	containerVmClUpdateCmd          *cobra.Command
}

// newContainerCmd returns the "container" command group for inv.
func newContainerCmd(inv *invocation) *cobra.Command {
	return newContainerCmds(inv).containerCmd
}

// newContainerCmds builds the "container" command group for inv and returns its state.
func newContainerCmds(inv *invocation) *containerCmds {
	s := &containerCmds{inv: inv}
	s.containerCmd = &cobra.Command{Use: "container", Short: "Manage Google Kubernetes Engine"}
	s.containerBareMetalCmd = &cobra.Command{Use: "bare-metal", Short: "Manage Anthos on bare metal"}
	s.containerBareMetalClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage bare-metal user clusters"}
	s.containerBmClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a bare-metal user cluster (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runCbmClCreate,
	}
	s.containerBmClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a bare-metal user cluster",
		Args: cobra.ExactArgs(1), RunE: s.runCbmClDelete,
	}
	s.containerBmClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a bare-metal user cluster",
		Args: cobra.ExactArgs(1), RunE: s.runCbmClDescribe,
	}
	s.containerBmClListCmd = &cobra.Command{
		Use: "list", Short: "List bare-metal user clusters in a location",
		Args: cobra.NoArgs, RunE: s.runCbmClList,
	}
	s.containerBmClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a bare-metal user cluster (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runCbmClUpdate,
	}
	s.containerBinauthzCmd = &cobra.Command{Use: "binauthz", Short: "Manage Binary Authorization"}
	s.containerBinauthzPolicyCmd = &cobra.Command{Use: "policy", Short: "Manage the Binary Authorization policy singleton for the project"}
	s.containerBinauthzAttestorsCmd = &cobra.Command{Use: "attestors", Short: "Manage Binary Authorization attestors"}
	s.binauthzPolicyDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the project's Binary Authorization policy",
		Args: cobra.NoArgs, RunE: s.runBaPolicyDescribe,
	}
	s.binauthzPolicyUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update the project's Binary Authorization policy (loads body from --config-file)",
		Args: cobra.NoArgs, RunE: s.runBaPolicyUpdate,
	}
	s.binauthzAttestorsCreateCmd = &cobra.Command{
		Use: "create ATTESTOR", Short: "Create an attestor (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runBaAttestorsCreate,
	}
	s.binauthzAttestorsDeleteCmd = &cobra.Command{
		Use: "delete ATTESTOR", Short: "Delete an attestor",
		Args: cobra.ExactArgs(1), RunE: s.runBaAttestorsDelete,
	}
	s.binauthzAttestorsDescribeCmd = &cobra.Command{
		Use: "describe ATTESTOR", Short: "Describe an attestor",
		Args: cobra.ExactArgs(1), RunE: s.runBaAttestorsDescribe,
	}
	s.binauthzAttestorsListCmd = &cobra.Command{
		Use: "list", Short: "List attestors in the current project",
		Args: cobra.NoArgs, RunE: s.runBaAttestorsList,
	}
	s.binauthzAttestorsUpdateCmd = &cobra.Command{
		Use: "update ATTESTOR", Short: "Update an attestor (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runBaAttestorsUpdate,
	}
	s.containerClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage GKE clusters"}
	s.containerClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a GKE cluster (loads CreateClusterRequest.cluster from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runCtnClCreate,
	}
	s.containerClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a GKE cluster",
		Args: cobra.ExactArgs(1), RunE: s.runCtnClDelete,
	}
	s.containerClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a GKE cluster",
		Args: cobra.ExactArgs(1), RunE: s.runCtnClDescribe,
	}
	s.containerClListCmd = &cobra.Command{
		Use: "list", Short: "List GKE clusters in a location, or in all locations",
		Args: cobra.NoArgs, RunE: s.runCtnClList,
	}
	s.containerClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a GKE cluster (loads UpdateClusterRequest.update from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runCtnClUpdate,
	}
	s.containerClCompleteCpUpgradeCmd = &cobra.Command{
		Use: "complete-control-plane-upgrade CLUSTER", Short: "Complete a rollbackable (two-step) control plane upgrade",
		Args: cobra.ExactArgs(1), RunE: s.runCtnClCompleteCpUpgrade,
	}
	s.containerFleetCmd = &cobra.Command{Use: "fleet", Short: "Manage GKE Hub fleets, features, and memberships"}
	s.containerHubCmd = &cobra.Command{Use: "hub", Short: "Alias for gcloud container fleet"}
	s.containerFleetFleetsCmd = &cobra.Command{Use: "fleets", Short: "Manage fleets"}
	s.containerFleetFeaturesCmd = &cobra.Command{Use: "features", Short: "Manage fleet features"}
	s.containerFleetMembershipsCmd = &cobra.Command{Use: "memberships", Short: "Manage fleet memberships"}
	s.containerFleetsCreateCmd = &cobra.Command{
		Use: "create FLEET", Short: "Create a fleet",
		Args: cobra.ExactArgs(1), RunE: s.runCfFleetsCreate,
	}
	s.containerFleetsDeleteCmd = &cobra.Command{
		Use: "delete FLEET", Short: "Delete a fleet",
		Args: cobra.ExactArgs(1), RunE: s.runCfFleetsDelete,
	}
	s.containerFleetsDescribeCmd = &cobra.Command{
		Use: "describe FLEET", Short: "Describe a fleet",
		Args: cobra.ExactArgs(1), RunE: s.runCfFleetsDescribe,
	}
	s.containerFleetsListCmd = &cobra.Command{
		Use: "list", Short: "List fleets in the organization",
		Args: cobra.NoArgs, RunE: s.runCfFleetsList,
	}
	s.containerFleetsUpdateCmd = &cobra.Command{
		Use: "update FLEET", Short: "Update a fleet",
		Args: cobra.ExactArgs(1), RunE: s.runCfFleetsUpdate,
	}
	s.containerFeaturesCreateCmd = &cobra.Command{
		Use: "create FEATURE", Short: "Create a fleet feature",
		Args: cobra.ExactArgs(1), RunE: s.runCfFeaturesCreate,
	}
	s.containerFeaturesDeleteCmd = &cobra.Command{
		Use: "delete FEATURE", Short: "Delete a fleet feature",
		Args: cobra.ExactArgs(1), RunE: s.runCfFeaturesDelete,
	}
	s.containerFeaturesDescribeCmd = &cobra.Command{
		Use: "describe FEATURE", Short: "Describe a fleet feature",
		Args: cobra.ExactArgs(1), RunE: s.runCfFeaturesDescribe,
	}
	s.containerFeaturesListCmd = &cobra.Command{
		Use: "list", Short: "List fleet features in a location",
		Args: cobra.NoArgs, RunE: s.runCfFeaturesList,
	}
	s.containerFeaturesUpdateCmd = &cobra.Command{
		Use: "update FEATURE", Short: "Update a fleet feature",
		Args: cobra.ExactArgs(1), RunE: s.runCfFeaturesUpdate,
	}
	s.containerMembershipsCreateCmd = &cobra.Command{
		Use: "create MEMBERSHIP", Short: "Create a fleet membership",
		Args: cobra.ExactArgs(1), RunE: s.runCfMembershipsCreate,
	}
	s.containerMembershipsDeleteCmd = &cobra.Command{
		Use: "delete MEMBERSHIP", Short: "Delete a fleet membership",
		Args: cobra.ExactArgs(1), RunE: s.runCfMembershipsDelete,
	}
	s.containerMembershipsDescribeCmd = &cobra.Command{
		Use: "describe MEMBERSHIP", Short: "Describe a fleet membership",
		Args: cobra.ExactArgs(1), RunE: s.runCfMembershipsDescribe,
	}
	s.containerMembershipsListCmd = &cobra.Command{
		Use: "list", Short: "List fleet memberships in a location",
		Args: cobra.NoArgs, RunE: s.runCfMembershipsList,
	}
	s.containerMembershipsUpdateCmd = &cobra.Command{
		Use: "update MEMBERSHIP", Short: "Update a fleet membership",
		Args: cobra.ExactArgs(1), RunE: s.runCfMembershipsUpdate,
	}
	s.containerGetServerConfigCmd = &cobra.Command{
		Use:   "get-server-config",
		Short: "Get GKE server configuration for a location",
		Args:  cobra.NoArgs,
		RunE:  s.runCtnGetServerConfig,
	}
	s.containerNodePoolsCmd = &cobra.Command{Use: "node-pools", Short: "Manage GKE node pools"}
	s.containerNpCreateCmd = &cobra.Command{
		Use: "create NODE_POOL", Short: "Create a node pool (loads CreateNodePoolRequest.nodePool from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runCtnNpCreate,
	}
	s.containerNpDeleteCmd = &cobra.Command{
		Use: "delete NODE_POOL", Short: "Delete a node pool",
		Args: cobra.ExactArgs(1), RunE: s.runCtnNpDelete,
	}
	s.containerNpDescribeCmd = &cobra.Command{
		Use: "describe NODE_POOL", Short: "Describe a node pool",
		Args: cobra.ExactArgs(1), RunE: s.runCtnNpDescribe,
	}
	s.containerNpListCmd = &cobra.Command{
		Use: "list", Short: "List node pools on a cluster",
		Args: cobra.NoArgs, RunE: s.runCtnNpList,
	}
	s.containerNpUpdateCmd = &cobra.Command{
		Use: "update NODE_POOL", Short: "Update a node pool (loads UpdateNodePoolRequest from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runCtnNpUpdate,
	}
	s.containerNpRollbackCmd = &cobra.Command{
		Use: "rollback NODE_POOL", Short: "Roll back a previously-Aborted or Failed NodePool upgrade",
		Args: cobra.ExactArgs(1), RunE: s.runCtnNpRollback,
	}
	s.containerOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage GKE long-running operations"}
	s.containerOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a GKE operation",
		Args: cobra.ExactArgs(1), RunE: s.runCtnOpCancel,
	}
	s.containerOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a GKE operation",
		Args: cobra.ExactArgs(1), RunE: s.runCtnOpDescribe,
	}
	s.containerOpListCmd = &cobra.Command{
		Use: "list", Short: "List GKE operations in a location",
		Args: cobra.NoArgs, RunE: s.runCtnOpList,
	}
	s.containerSubnetsCmd = &cobra.Command{Use: "subnets", Short: "Inspect subnets usable by GKE"}
	s.containerSubnetsListUsableCmd = &cobra.Command{
		Use:   "list-usable",
		Short: "List subnets usable for creating GKE clusters in the current project",
		Args:  cobra.NoArgs,
		RunE:  s.runCtnSnListUsable,
	}
	s.containerVmwareCmd = &cobra.Command{Use: "vmware", Short: "Manage Anthos on VMware"}
	s.containerVmwareClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage VMware user clusters"}
	s.containerVmClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a VMware user cluster (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runCvmClCreate,
	}
	s.containerVmClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a VMware user cluster",
		Args: cobra.ExactArgs(1), RunE: s.runCvmClDelete,
	}
	s.containerVmClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a VMware user cluster",
		Args: cobra.ExactArgs(1), RunE: s.runCvmClDescribe,
	}
	s.containerVmClListCmd = &cobra.Command{
		Use: "list", Short: "List VMware user clusters in a location",
		Args: cobra.NoArgs, RunE: s.runCvmClList,
	}
	s.containerVmClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a VMware user cluster (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runCvmClUpdate,
	}
	s.initContainer()
	s.initContainerBareMetal()
	s.initContainerBinauthz()
	s.initContainerClusters()
	s.initContainerFleet()
	s.initContainerGetServerConfig()
	s.initContainerNodePools()
	s.initContainerOperations()
	s.initContainerSubnets()
	s.initContainerVmware()
	return s
}
