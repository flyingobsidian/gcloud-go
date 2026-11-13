package cli

import "github.com/spf13/cobra"

// --- gcloud edge-cloud (#334) ---

func (s *edgeCloudCmds) initEdgeCloud() {
	container := &cobra.Command{Use: "container", Short: "Manage Edge Container resources"}
	registerStubGroup(container, "clusters", "Manage Edge clusters", "create", "delete", "describe", "list", "update", "upgrade", "get-credentials", "generate-access-token", "generate-offline-credential")
	registerStubGroup(container, "machines", "Manage machines", "describe", "list")
	registerStubGroup(container, "operations", "Manage operations", "describe", "list")
	registerStubGroup(container, "regions", "List Edge Container regions", "list")
	registerStubGroup(container, "zones", "List Edge Container zones", "list")
	// gcloud-python exposes `get-server-config` as a leaf under
	// `edge-cloud container`, not a subgroup with a `describe` sub. Match
	// that shape.
	registerStubCommand(container, "get-server-config", "Describe the Edge Container server config")
	s.edgeCloudCmd.AddCommand(container)

	networking := &cobra.Command{Use: "networking", Short: "Manage Edge Network resources"}
	registerStubGroup(networking, "interconnects", "Manage interconnects", "describe", "list")
	registerStubGroup(networking, "networks", "Manage networks", "create", "delete", "describe", "list", "diagnose")
	registerStubGroup(networking, "operations", "Manage operations", "describe", "list")
	registerStubGroup(networking, "routers", "Manage routers", "create", "delete", "describe", "list", "update", "diagnose", "change-bgp-peering")
	registerStubGroup(networking, "routes", "Manage routes", "create", "delete", "describe", "list")
	registerStubGroup(networking, "subnets", "Manage subnets", "create", "delete", "describe", "list", "update")
	registerStubGroup(networking, "zones", "Manage zones", "describe", "list", "initialize")
	s.edgeCloudCmd.AddCommand(networking)

}

// edgeCloudCmds holds the commands and flag values of one invocation's "edgeCloud"
// command group, so invocations don't share state (#2070).
type edgeCloudCmds struct {
	inv *invocation

	edgeCloudCmd *cobra.Command
}

// newEdgeCloudCmd returns the "edgeCloud" command group for inv.
func newEdgeCloudCmd(inv *invocation) *cobra.Command {
	return newEdgeCloudCmds(inv).edgeCloudCmd
}

// newEdgeCloudCmds builds the "edgeCloud" command group for inv and returns its state.
func newEdgeCloudCmds(inv *invocation) *edgeCloudCmds {
	s := &edgeCloudCmds{inv: inv}
	s.edgeCloudCmd = &cobra.Command{Use: "edge-cloud", Short: "Manage Distributed Cloud Edge"}
	s.initEdgeCloud()
	return s
}
