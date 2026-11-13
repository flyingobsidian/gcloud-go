package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

func (s *computeCmds) initNetworks() {
	s.networksListCmd.Flags().StringVar(&s.flagNetworksListFormat, "format", "", "Output format (e.g. json)")
	s.networksListCmd.Flags().StringVar(&s.flagNetworksListFilter, "filter", "", "Filter expression")
	s.networksListCmd.Flags().BoolVar(&s.flagNetworksListURI, "uri", false, "Print resource URIs")

	s.networksCmd.AddCommand(s.networksListCmd)
	s.computeCmd.AddCommand(s.networksCmd)
}

func (s *computeCmds) runNetworksList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	var allNetworks []*compute.Network
	pageToken := ""
	for {
		call := svc.Networks.List(project).Context(ctx)
		if s.flagNetworksListFilter != "" {
			call = call.Filter(s.flagNetworksListFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing networks: %w", err)
		}
		allNetworks = append(allNetworks, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if s.flagNetworksListURI {
		for _, n := range allNetworks {
			fmt.Fprintln(s.inv.stdout, n.SelfLink)
		}
		return nil
	}

	if s.flagNetworksListFormat == "json" {
		return s.inv.printJSON(allNetworks)
	}

	// Default table format matching gcloud-python:
	// NAME, SUBNET_MODE, BGP_ROUTING_MODE, IPV4_RANGE, GATEWAY_IPV4
	printTableHeadingTo(s.inv.stdout, len(allNetworks), "%-30s %-15s %-20s %-18s %-15s\n", "NAME", "SUBNET_MODE", "BGP_ROUTING_MODE", "IPV4_RANGE", "GATEWAY_IPV4")
	for _, n := range allNetworks {
		subnetMode := networkSubnetMode(n)
		bgpMode := networkBGPRoutingMode(n)
		fmt.Fprintf(s.inv.stdout, "%-30s %-15s %-20s %-18s %-15s\n", n.Name, subnetMode, bgpMode, n.IPv4Range, n.GatewayIPv4)
	}
	return nil
}

// networkSubnetMode derives the subnet mode from the network resource,
// matching gcloud-python's GetSubnetMode logic.
func networkSubnetMode(n *compute.Network) string {
	if n.IPv4Range != "" {
		return "LEGACY"
	}
	if n.AutoCreateSubnetworks {
		return "AUTO"
	}
	return "CUSTOM"
}

// networkBGPRoutingMode returns the BGP routing mode from the network's
// routing config, matching gcloud-python's GetBgpRoutingMode logic.
func networkBGPRoutingMode(n *compute.Network) string {
	if n.RoutingConfig != nil {
		return n.RoutingConfig.RoutingMode
	}
	return ""
}
