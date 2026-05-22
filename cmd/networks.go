package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

var networksCmd = &cobra.Command{
	Use:   "networks",
	Short: "Manage VPC networks",
}

var networksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Compute Engine networks",
	Args:  cobra.NoArgs,
	RunE:  runNetworksList,
}

var (
	flagNetworksListFormat string
	flagNetworksListFilter string
	flagNetworksListURI    bool
)

func init() {
	networksListCmd.Flags().StringVar(&flagNetworksListFormat, "format", "", "Output format (e.g. json)")
	networksListCmd.Flags().StringVar(&flagNetworksListFilter, "filter", "", "Filter expression")
	networksListCmd.Flags().BoolVar(&flagNetworksListURI, "uri", false, "Print resource URIs")

	networksCmd.AddCommand(networksListCmd)
	computeCmd.AddCommand(networksCmd)
}

func runNetworksList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var allNetworks []*compute.Network
	pageToken := ""
	for {
		call := svc.Networks.List(project).Context(ctx)
		if flagNetworksListFilter != "" {
			call = call.Filter(flagNetworksListFilter)
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

	if flagNetworksListURI {
		for _, n := range allNetworks {
			fmt.Println(n.SelfLink)
		}
		return nil
	}

	if flagNetworksListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allNetworks)
	}

	// Default table format matching gcloud-python:
	// NAME, SUBNET_MODE, BGP_ROUTING_MODE, IPV4_RANGE, GATEWAY_IPV4
	fmt.Printf("%-30s %-15s %-20s %-18s %-15s\n", "NAME", "SUBNET_MODE", "BGP_ROUTING_MODE", "IPV4_RANGE", "GATEWAY_IPV4")
	for _, n := range allNetworks {
		subnetMode := networkSubnetMode(n)
		bgpMode := networkBGPRoutingMode(n)
		fmt.Printf("%-30s %-15s %-20s %-18s %-15s\n", n.Name, subnetMode, bgpMode, n.IPv4Range, n.GatewayIPv4)
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
