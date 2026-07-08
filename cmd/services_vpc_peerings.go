package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	sn "google.golang.org/api/servicenetworking/v1"
)

var servicesVpcPeeringsCmd = &cobra.Command{
	Use:   "vpc-peerings",
	Short: "Manage VPC peerings to services",
}

var vpcPeeringConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Create a peering connection between a consumer network and a service",
	Args:  cobra.NoArgs,
	RunE:  runVpcPeeringConnect,
}

var vpcPeeringListCmd = &cobra.Command{
	Use:   "list",
	Short: "List VPC peering connections for a service",
	Args:  cobra.NoArgs,
	RunE:  runVpcPeeringList,
}

var vpcPeeringUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reserved ranges on a VPC peering connection",
	Args:  cobra.NoArgs,
	RunE:  runVpcPeeringUpdate,
}

var vpcPeeringDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a VPC peering connection",
	Args:  cobra.NoArgs,
	RunE:  runVpcPeeringDelete,
}

var vpcPeeringEnableVpcSCCmd = &cobra.Command{
	Use:   "enable-vpc-service-controls",
	Short: "Enable VPC Service Controls on a peering connection",
	Args:  cobra.NoArgs,
	RunE:  makeVpcSCToggle(true),
}

var vpcPeeringDisableVpcSCCmd = &cobra.Command{
	Use:   "disable-vpc-service-controls",
	Short: "Disable VPC Service Controls on a peering connection",
	Args:  cobra.NoArgs,
	RunE:  makeVpcSCToggle(false),
}

var vpcPeeringGetVpcSCCmd = &cobra.Command{
	Use:   "get-vpc-service-controls",
	Short: "Get VPC Service Controls status on a peering connection",
	Args:  cobra.NoArgs,
	RunE:  runVpcPeeringGetVpcSC,
}

var (
	flagVpcPeeringService    string
	flagVpcPeeringNetwork    string
	flagVpcPeeringRanges     []string
	flagVpcPeeringForce      bool
	flagVpcPeeringListFormat string
)

func init() {
	scopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&flagVpcPeeringService, "service", "servicenetworking.googleapis.com", "Peered service (default: servicenetworking.googleapis.com)")
		c.Flags().StringVar(&flagVpcPeeringNetwork, "network", "", "Consumer VPC network name (required)")
		c.MarkFlagRequired("network")
	}
	scopeFlags(vpcPeeringConnectCmd)
	vpcPeeringConnectCmd.Flags().StringSliceVar(&flagVpcPeeringRanges, "ranges", nil, "Reserved IP ranges to use for the peering (required)")
	vpcPeeringConnectCmd.Flags().BoolVar(&flagVpcPeeringForce, "force", false, "Overwrite existing ranges on the peering")
	vpcPeeringConnectCmd.MarkFlagRequired("ranges")

	scopeFlags(vpcPeeringListCmd)
	vpcPeeringListCmd.Flags().StringVar(&flagVpcPeeringListFormat, "format", "", "Output format (json, yaml, or table)")

	scopeFlags(vpcPeeringUpdateCmd)
	vpcPeeringUpdateCmd.Flags().StringSliceVar(&flagVpcPeeringRanges, "ranges", nil, "Reserved IP ranges to use for the peering (required)")
	vpcPeeringUpdateCmd.Flags().BoolVar(&flagVpcPeeringForce, "force", false, "Overwrite existing ranges on the peering")
	vpcPeeringUpdateCmd.MarkFlagRequired("ranges")

	scopeFlags(vpcPeeringDeleteCmd)

	for _, c := range []*cobra.Command{vpcPeeringEnableVpcSCCmd, vpcPeeringDisableVpcSCCmd, vpcPeeringGetVpcSCCmd} {
		scopeFlags(c)
	}

	servicesVpcPeeringsCmd.AddCommand(
		vpcPeeringConnectCmd,
		vpcPeeringListCmd,
		vpcPeeringUpdateCmd,
		vpcPeeringDeleteCmd,
		vpcPeeringEnableVpcSCCmd,
		vpcPeeringDisableVpcSCCmd,
		vpcPeeringGetVpcSCCmd,
	)
	servicesCmd.AddCommand(servicesVpcPeeringsCmd)
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

func runVpcPeeringConnect(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	net := consumerNetwork(project, flagVpcPeeringNetwork)
	op, err := svc.Services.Connections.Create(serviceParent(flagVpcPeeringService), &sn.Connection{
		Network:               net,
		ReservedPeeringRanges: flagVpcPeeringRanges,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("connecting VPC peering: %w", err)
	}
	fmt.Printf("VPC peering connect in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runVpcPeeringList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Services.Connections.List(serviceParent(flagVpcPeeringService)).Network(consumerNetwork(project, flagVpcPeeringNetwork)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing VPC peerings: %w", err)
	}
	return printListResults(resp.Connections, flagVpcPeeringListFormat, func() {
		fmt.Printf("%-30s %-40s %s\n", "PEERING", "NETWORK", "RESERVED_RANGES")
		for _, c := range resp.Connections {
			fmt.Printf("%-30s %-40s %s\n", c.Peering, c.Network, strings.Join(c.ReservedPeeringRanges, ","))
		}
	})
}

func runVpcPeeringUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// The API refers to connections as
	// `services/{service}/connections/{network}`; use "-" per the docs.
	name := serviceParent(flagVpcPeeringService) + "/connections/-"
	op, err := svc.Services.Connections.Patch(name, &sn.Connection{
		Network:               consumerNetwork(project, flagVpcPeeringNetwork),
		ReservedPeeringRanges: flagVpcPeeringRanges,
	}).Force(flagVpcPeeringForce).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating VPC peering: %w", err)
	}
	fmt.Printf("VPC peering update in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runVpcPeeringDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := serviceParent(flagVpcPeeringService) + "/connections/-"
	op, err := svc.Services.Connections.DeleteConnection(name, &sn.DeleteConnectionRequest{
		ConsumerNetwork: consumerNetwork(project, flagVpcPeeringNetwork),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting VPC peering: %w", err)
	}
	fmt.Printf("VPC peering delete in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func makeVpcSCToggle(enable bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := resolveProject()
		if err != nil {
			return err
		}
		ctx := context.Background()
		svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
		if err != nil {
			return err
		}
		parent := serviceParent(flagVpcPeeringService)
		net := consumerNetwork(project, flagVpcPeeringNetwork)
		if enable {
			op, err := svc.Services.EnableVpcServiceControls(parent, &sn.EnableVpcServiceControlsRequest{
				ConsumerNetwork: net,
			}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("enabling VPC service controls: %w", err)
			}
			fmt.Printf("Enable VPC service controls in progress (operation: %s).\n", op.Name)
			return yamlEncode(op)
		}
		op, err := svc.Services.DisableVpcServiceControls(parent, &sn.DisableVpcServiceControlsRequest{
			ConsumerNetwork: net,
		}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("disabling VPC service controls: %w", err)
		}
		fmt.Printf("Disable VPC service controls in progress (operation: %s).\n", op.Name)
		return yamlEncode(op)
	}
}

func runVpcPeeringGetVpcSC(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceNetworkingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// The name for the getter is
	// `services/{service}/projects/{project}/networks/{network}/vpcServiceControls`.
	name := fmt.Sprintf("%s/projects/%s/networks/%s/vpcServiceControls",
		serviceParent(flagVpcPeeringService), project, flagVpcPeeringNetwork)
	resp, err := svc.Services.Projects.Global.Networks.GetVpcServiceControls(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting VPC service controls: %w", err)
	}
	return yamlEncode(resp)
}
