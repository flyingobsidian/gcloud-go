package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	sn "google.golang.org/api/servicenetworking/v1"
)

func (g *servicesCmds) initServicesVpcPeerings() {
	scopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&g.flagVpcPeeringService, "service", "servicenetworking.googleapis.com", "Peered service (default: servicenetworking.googleapis.com)")
		c.Flags().StringVar(&g.flagVpcPeeringNetwork, "network", "", "Consumer VPC network name (required)")
		c.MarkFlagRequired("network")
	}
	scopeFlags(g.vpcPeeringConnectCmd)
	g.vpcPeeringConnectCmd.Flags().StringSliceVar(&g.flagVpcPeeringRanges, "ranges", nil, "Reserved IP ranges to use for the peering (required)")
	g.vpcPeeringConnectCmd.Flags().BoolVar(&g.flagVpcPeeringForce, "force", false, "Overwrite existing ranges on the peering")
	g.vpcPeeringConnectCmd.MarkFlagRequired("ranges")

	scopeFlags(g.vpcPeeringListCmd)
	g.vpcPeeringListCmd.Flags().StringVar(&g.flagVpcPeeringListFormat, "format", "", "Output format (json, yaml, or table)")

	scopeFlags(g.vpcPeeringUpdateCmd)
	g.vpcPeeringUpdateCmd.Flags().StringSliceVar(&g.flagVpcPeeringRanges, "ranges", nil, "Reserved IP ranges to use for the peering (required)")
	g.vpcPeeringUpdateCmd.Flags().BoolVar(&g.flagVpcPeeringForce, "force", false, "Overwrite existing ranges on the peering")
	g.vpcPeeringUpdateCmd.MarkFlagRequired("ranges")

	scopeFlags(g.vpcPeeringDeleteCmd)

	for _, c := range []*cobra.Command{g.vpcPeeringEnableVpcSCCmd, g.vpcPeeringDisableVpcSCCmd, g.vpcPeeringGetVpcSCCmd} {
		scopeFlags(c)
	}

	g.servicesVpcPeeringsCmd.AddCommand(
		g.vpcPeeringConnectCmd,
		g.vpcPeeringListCmd,
		g.vpcPeeringUpdateCmd,
		g.vpcPeeringDeleteCmd,
		g.vpcPeeringEnableVpcSCCmd,
		g.vpcPeeringDisableVpcSCCmd,
		g.vpcPeeringGetVpcSCCmd,
	)
	g.servicesCmd.AddCommand(g.servicesVpcPeeringsCmd)
}

// serviceParent returns the service resource path used by the peering API:
// `services/{service}`.
func serviceParent(service string) string {
	if strings.HasPrefix(service, "services/") {
		return service
	}
	return "services/" + service
}

// consumerNetwork returns a fully qualified consumer VPC network name:
// `projects/{project}/global/networks/{network}`.
func consumerNetwork(project, network string) string {
	if strings.HasPrefix(network, "projects/") {
		return network
	}
	return fmt.Sprintf("projects/%s/global/networks/%s", project, network)
}

func (g *servicesCmds) runVpcPeeringConnect(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceNetworking(ctx)
	if err != nil {
		return err
	}
	net := consumerNetwork(project, g.flagVpcPeeringNetwork)
	op, err := svc.Services.Connections.Create(serviceParent(g.flagVpcPeeringService), &sn.Connection{
		Network:               net,
		ReservedPeeringRanges: g.flagVpcPeeringRanges,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("connecting VPC peering: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "VPC peering connect in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) runVpcPeeringList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceNetworking(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Services.Connections.List(serviceParent(g.flagVpcPeeringService)).Network(consumerNetwork(project, g.flagVpcPeeringNetwork)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing VPC peerings: %w", err)
	}
	return g.inv.printListResults(resp.Connections, g.flagVpcPeeringListFormat, func() {
		printTableHeadingTo(g.inv.stdout, len(resp.Connections), "%-30s %-40s %s\n", "PEERING", "NETWORK", "RESERVED_RANGES")
		for _, c := range resp.Connections {
			fmt.Fprintf(g.inv.stdout, "%-30s %-40s %s\n", c.Peering, c.Network, strings.Join(c.ReservedPeeringRanges, ","))
		}
	})
}

func (g *servicesCmds) runVpcPeeringUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceNetworking(ctx)
	if err != nil {
		return err
	}
	// The API refers to connections as
	// `services/{service}/connections/{network}`; use "-" per the docs.
	name := serviceParent(g.flagVpcPeeringService) + "/connections/-"
	op, err := svc.Services.Connections.Patch(name, &sn.Connection{
		Network:               consumerNetwork(project, g.flagVpcPeeringNetwork),
		ReservedPeeringRanges: g.flagVpcPeeringRanges,
	}).Force(g.flagVpcPeeringForce).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating VPC peering: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "VPC peering update in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) runVpcPeeringDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceNetworking(ctx)
	if err != nil {
		return err
	}
	name := serviceParent(g.flagVpcPeeringService) + "/connections/-"
	op, err := svc.Services.Connections.DeleteConnection(name, &sn.DeleteConnectionRequest{
		ConsumerNetwork: consumerNetwork(project, g.flagVpcPeeringNetwork),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting VPC peering: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "VPC peering delete in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) makeVpcSCToggle(enable bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := g.inv.project()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		svc, err := g.inv.clients().ServiceNetworking(ctx)
		if err != nil {
			return err
		}
		parent := serviceParent(g.flagVpcPeeringService)
		net := consumerNetwork(project, g.flagVpcPeeringNetwork)
		if enable {
			op, err := svc.Services.EnableVpcServiceControls(parent, &sn.EnableVpcServiceControlsRequest{
				ConsumerNetwork: net,
			}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("enabling VPC service controls: %w", err)
			}
			fmt.Fprintf(g.inv.stdout, "Enable VPC service controls in progress (operation: %s).\n", op.Name)
			return g.inv.yaml(op)
		}
		op, err := svc.Services.DisableVpcServiceControls(parent, &sn.DisableVpcServiceControlsRequest{
			ConsumerNetwork: net,
		}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("disabling VPC service controls: %w", err)
		}
		fmt.Fprintf(g.inv.stdout, "Disable VPC service controls in progress (operation: %s).\n", op.Name)
		return g.inv.yaml(op)
	}
}

func (g *servicesCmds) runVpcPeeringGetVpcSC(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceNetworking(ctx)
	if err != nil {
		return err
	}
	// The name for the getter is
	// `services/{service}/projects/{project}/networks/{network}/vpcServiceControls`.
	name := fmt.Sprintf("%s/projects/%s/networks/%s/vpcServiceControls",
		serviceParent(g.flagVpcPeeringService), project, g.flagVpcPeeringNetwork)
	resp, err := svc.Services.Projects.Global.Networks.GetVpcServiceControls(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting VPC service controls: %w", err)
	}
	return g.inv.yaml(resp)
}
